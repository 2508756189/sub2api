package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

// EnsembleHandler fans one client request out to every enabled proposer model of
// an ensemble group, then optionally synthesizes the answers with an aggregator
// model. All member models are served by the ensemble group's own bound accounts
// (in-group aggregation only — there is no cross-group or cross-upstream routing).
//
// Each sub-call re-enters the normal per-platform gateway handler, so account
// selection, model mapping, failover, quota and usage recording behave exactly as
// they do for a direct call. That also means an N-proposer request bills N
// sub-calls (plus one for the aggregator), which is the intended cost model.
type EnsembleHandler struct {
	runtime       *service.EnsembleRuntimeService
	dispatch      gin.HandlerFunc
	costEstimator service.EnsembleCostEstimator
}

func NewEnsembleHandler(runtime *service.EnsembleRuntimeService, estimators ...service.EnsembleCostEstimator) *EnsembleHandler {
	var costEstimator service.EnsembleCostEstimator
	if len(estimators) > 0 {
		costEstimator = estimators[0]
	}
	return &EnsembleHandler{runtime: runtime, costEstimator: costEstimator}
}

// SetSubCallDispatcher injects the per-platform routing closure used for member
// sub-calls. It is wired at route-registration time because the routing decision
// (which gateway handler serves a given platform) lives with the routes.
func (h *EnsembleHandler) SetSubCallDispatcher(dispatch gin.HandlerFunc) {
	h.dispatch = dispatch
}

// ensembleProposal is one successful member answer.
type ensembleProposal struct {
	Model   string
	Content string
}

// ensembleMemberStat is the per-member record exposed via ensemble_metadata.
type ensembleMemberStat struct {
	Model            string   `json:"model"`
	Platform         string   `json:"platform,omitempty"`
	Role             string   `json:"role"`
	Status           string   `json:"status"`
	DurationMs       int64    `json:"duration_ms"`
	PromptTokens     int      `json:"prompt_tokens,omitempty"`
	CompletionTokens int      `json:"completion_tokens,omitempty"`
	Content          string   `json:"content,omitempty"`
	Cost             *float64 `json:"cost"`
	CostSource       string   `json:"cost_source,omitempty"`
	Error            string   `json:"error,omitempty"`
}

// ensembleSubResult is the internal outcome of one member sub-call.
type ensembleSubResult struct {
	index   int
	model   string
	role    string
	content string
	stat    ensembleMemberStat
	err     error
}

const (
	ensembleStatusOK               = "ok"
	ensembleStatusFailed           = "failed"
	maxEnsembleAggregatorBodyBytes = 2 << 20
	ensembleProgressSinkKey        = "ensemble.progress.sink"
)

// EnsembleProgressEvent is the small, stable event contract used by the
// authenticated diagnostic stream. The normal Chat Completions response stays
// unchanged; these events only make the execution observable to the admin UI.
type EnsembleProgressEvent struct {
	Type               string              `json:"type"`
	Model              string              `json:"model,omitempty"`
	Platform           string              `json:"platform,omitempty"`
	Role               string              `json:"role,omitempty"`
	Status             string              `json:"status,omitempty"`
	Error              string              `json:"error,omitempty"`
	Member             *ensembleMemberStat `json:"member,omitempty"`
	ProposersTotal     int                 `json:"proposers_total,omitempty"`
	ProposersSucceeded int                 `json:"proposers_succeeded,omitempty"`
	Aggregator         string              `json:"aggregator,omitempty"`
	Aggregated         bool                `json:"aggregated,omitempty"`
	DurationMs         int64               `json:"duration_ms,omitempty"`
	StatusCode         int                 `json:"status_code,omitempty"`
	Response           json.RawMessage     `json:"response,omitempty"`
}

func emitEnsembleProgress(c *gin.Context, event EnsembleProgressEvent) {
	if c == nil {
		return
	}
	sinkValue, ok := c.Get(ensembleProgressSinkKey)
	if !ok {
		return
	}
	sink, ok := sinkValue.(func(EnsembleProgressEvent))
	if ok && sink != nil {
		sink(event)
	}
}

// ChatCompletions handles POST /v1/chat/completions for platform=ensemble groups.
func (h *EnsembleHandler) ChatCompletions(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.Group == nil {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	if h.runtime == nil || h.dispatch == nil {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "Ensemble runtime is not available")
		return
	}

	reqLog := requestLogger(
		c,
		"handler.ensemble.chat_completions",
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 || !gjson.ValidBytes(body) {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}

	requestedModel := gjson.GetBytes(body, "model").String()
	if strings.TrimSpace(requestedModel) == "" {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "The model field is required")
		return
	}
	clientWantsStream := gjson.GetBytes(body, "stream").Bool()

	plan, err := h.runtime.LoadPlan(c.Request.Context(), apiKey.Group.ID, apiKey.Group.EnsembleConfig)
	if err != nil {
		reqLog.Error("failed to load ensemble plan", zap.Error(err))
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "Failed to load ensemble configuration")
		return
	}
	if len(plan.Proposers) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error",
			"This ensemble group has no enabled proposer models. Add members in the admin Ensemble configuration.")
		return
	}
	aggregatorModel := ""
	if plan.ShouldAggregate() {
		aggregatorModel = plan.Aggregator.Model
	}
	emitEnsembleProgress(c, EnsembleProgressEvent{
		Type:           "started",
		ProposersTotal: len(plan.Proposers),
		Aggregator:     aggregatorModel,
	})

	timeout := time.Duration(plan.EffectiveTimeoutSeconds()) * time.Second
	started := time.Now()

	// Fan out to every proposer in parallel.
	results := make([]ensembleSubResult, len(plan.Proposers))
	var wg sync.WaitGroup
	for i := range plan.Proposers {
		wg.Add(1)
		go func(idx int, member service.EnsembleProposer) {
			defer wg.Done()
			emitEnsembleProgress(c, EnsembleProgressEvent{
				Type:     "member_started",
				Model:    member.Model,
				Platform: member.Platform,
				Role:     service.EnsembleRoleProposer,
			})
			var result ensembleSubResult
			defer func() {
				if r := recover(); r != nil {
					result = ensembleSubResult{
						index: idx,
						model: member.Model,
						role:  service.EnsembleRoleProposer,
						err:   fmt.Errorf("panic in proposer sub-call: %v", r),
						stat: ensembleMemberStat{
							Model:    member.Model,
							Platform: member.Platform,
							Role:     service.EnsembleRoleProposer,
							Status:   ensembleStatusFailed,
							Error:    fmt.Sprintf("panic: %v", r),
						},
					}
				}
				results[idx] = result
				stat := result.stat
				emitEnsembleProgress(c, EnsembleProgressEvent{
					Type:   "member_finished",
					Model:  stat.Model,
					Role:   stat.Role,
					Status: stat.Status,
					Member: &stat,
				})
			}()
			result = h.runSubCall(c, idx, member.Model, member.Platform, service.EnsembleRoleProposer, body, plan.Config.MaxTokens, timeout)
		}(i, plan.Proposers[i])
	}
	wg.Wait()

	proposals := make([]ensembleProposal, 0, len(results))
	stats := make([]ensembleMemberStat, 0, len(results)+1)
	for _, res := range results {
		stats = append(stats, res.stat)
		if res.err == nil && strings.TrimSpace(res.content) != "" {
			proposals = append(proposals, ensembleProposal{Model: res.model, Content: res.content})
		} else if res.err != nil {
			reqLog.Warn("ensemble proposer failed",
				zap.String("model", res.model),
				zap.Error(res.err))
		}
	}
	emitEnsembleProgress(c, EnsembleProgressEvent{
		Type:               "proposers_finished",
		ProposersTotal:     len(plan.Proposers),
		ProposersSucceeded: len(proposals),
	})

	minProposers := plan.EffectiveMinProposers()
	if len(proposals) < minProposers {
		reqLog.Error("ensemble did not reach min_proposers",
			zap.Int("succeeded", len(proposals)),
			zap.Int("required", minProposers))
		emitEnsembleProgress(c, EnsembleProgressEvent{
			Type:               "error",
			Error:              fmt.Sprintf("Only %d proposers succeeded, minimum %d required", len(proposals), minProposers),
			ProposersTotal:     len(plan.Proposers),
			ProposersSucceeded: len(proposals),
		})
		h.errorResponse(c, http.StatusBadGateway, "api_error",
			fmt.Sprintf("Only %d proposers succeeded, minimum %d required", len(proposals), minProposers))
		return
	}

	// Aggregate, or fall back to the longest proposal.
	finalContent := longestEnsembleProposal(proposals)
	aggregated := false
	if plan.ShouldAggregate() && len(proposals) > 0 {
		aggBody, buildErr := buildEnsembleAggregatorBody(body, proposals)
		if buildErr != nil {
			reqLog.Warn("failed to build aggregator request, falling back to longest proposal", zap.Error(buildErr))
			emitEnsembleProgress(c, EnsembleProgressEvent{
				Type:       "fallback",
				Model:      plan.Aggregator.Model,
				Role:       service.EnsembleRoleAggregator,
				Status:     ensembleStatusFailed,
				Error:      buildErr.Error(),
				Aggregated: false,
			})
		} else {
			emitEnsembleProgress(c, EnsembleProgressEvent{
				Type:     "member_started",
				Model:    plan.Aggregator.Model,
				Platform: plan.Aggregator.Platform,
				Role:     service.EnsembleRoleAggregator,
			})
			aggRes := h.runSubCall(c, len(results), plan.Aggregator.Model, plan.Aggregator.Platform, service.EnsembleRoleAggregator,
				aggBody, plan.Config.MaxTokens, timeout)
			stats = append(stats, aggRes.stat)
			stat := aggRes.stat
			emitEnsembleProgress(c, EnsembleProgressEvent{
				Type:   "member_finished",
				Model:  stat.Model,
				Role:   stat.Role,
				Status: stat.Status,
				Member: &stat,
			})
			if aggRes.err == nil && strings.TrimSpace(aggRes.content) != "" {
				finalContent = aggRes.content
				aggregated = true
			} else {
				emitEnsembleProgress(c, EnsembleProgressEvent{
					Type:       "fallback",
					Model:      plan.Aggregator.Model,
					Role:       service.EnsembleRoleAggregator,
					Status:     ensembleStatusFailed,
					Error:      aggRes.stat.Error,
					Aggregated: false,
				})
				reqLog.Warn("ensemble aggregator failed, falling back to longest proposal",
					zap.String("model", plan.Aggregator.Model),
					zap.Error(aggRes.err))
			}
		}
	}

	reqLog.Info("ensemble request complete",
		zap.Int("proposers_total", len(plan.Proposers)),
		zap.Int("proposers_succeeded", len(proposals)),
		zap.Bool("aggregated", aggregated),
		zap.Duration("duration", time.Since(started)))

	payload := buildEnsembleChatResponse(requestedModel, finalContent, stats, plan.Config.ExposeMetadata, aggregated, time.Since(started))
	if clientWantsStream {
		h.writeStreamResponse(c, requestedModel, finalContent, payload)
		return
	}
	c.JSON(http.StatusOK, payload)
}

// TestStream runs the normal Ensemble Chat Completions handler in an isolated
// recorder and forwards progress over SSE. API-key authentication and the
// downstream dispatch path are inherited from the gateway route, so this is a
// diagnostic view of the same billing behavior rather than a second executor.
func (h *EnsembleHandler) TestStream(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.Group == nil {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	if apiKey.Group.Platform != service.PlatformEnsemble {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "The API key is not bound to an Ensemble group")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	events := make(chan EnsembleProgressEvent, 32)
	started := time.Now()
	recorder := httptest.NewRecorder()
	diagnosticContext, _ := gin.CreateTestContext(recorder)
	diagnosticContext.Request = c.Request.Clone(c.Request.Context())
	diagnosticContext.Request.Body = io.NopCloser(bytes.NewReader(body))
	diagnosticContext.Request.ContentLength = int64(len(body))
	for key, value := range c.Keys {
		diagnosticContext.Set(key, value)
	}
	diagnosticContext.Set(ensembleProgressSinkKey, func(event EnsembleProgressEvent) {
		select {
		case events <- event:
		case <-c.Request.Context().Done():
		}
	})

	go func() {
		h.ChatCompletions(diagnosticContext)
		terminal := EnsembleProgressEvent{
			Type:       "completed",
			StatusCode: recorder.Code,
			DurationMs: time.Since(started).Milliseconds(),
			Response:   json.RawMessage(append([]byte(nil), recorder.Body.Bytes()...)),
		}
		if recorder.Code >= http.StatusBadRequest {
			terminal.Type = "error"
			terminal.Error = gjson.GetBytes(recorder.Body.Bytes(), "error.message").String()
			if terminal.Error == "" {
				terminal.Error = ensembleTruncate(recorder.Body.String(), 500)
			}
		}
		select {
		case events <- terminal:
		case <-c.Request.Context().Done():
		}
		close(events)
	}()

	for {
		select {
		case event, open := <-events:
			if !open {
				return
			}
			c.SSEvent(event.Type, event)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

// runSubCall executes one member call by re-entering the normal gateway pipeline
// with the member's model substituted in. The resolved target platform on the
// sub-request context both routes the call to the right upstream handler and acts
// as the recursion guard (the outer ensemble gate only fires when it is absent).
func (h *EnsembleHandler) runSubCall(
	parent *gin.Context,
	index int,
	model string,
	platform string,
	role string,
	body []byte,
	maxTokens int,
	timeout time.Duration,
) ensembleSubResult {
	res := ensembleSubResult{index: index, model: model, role: role}
	callStart := time.Now()
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform == "" {
		platform, _ = service.DetectModelPlatform(model)
	}
	res.stat = ensembleMemberStat{Model: model, Platform: platform, Role: role, Status: ensembleStatusFailed}

	if platform == "" {
		res.err = fmt.Errorf("cannot determine upstream platform for model %q", model)
		res.stat.Error = res.err.Error()
		res.stat.DurationMs = time.Since(callStart).Milliseconds()
		return res
	}

	subBody, err := prepareEnsembleSubCallBody(body, model, maxTokens)
	if err != nil {
		res.err = err
		res.stat.Error = err.Error()
		res.stat.DurationMs = time.Since(callStart).Milliseconds()
		return res
	}

	recorder := httptest.NewRecorder()
	subCtx, _ := gin.CreateTestContext(recorder)

	ctx, cancel := context.WithTimeout(parent.Request.Context(), timeout)
	defer cancel()
	ctx = service.WithResolvedTargetPlatform(ctx, platform)
	// Usage billing deduplicates by (client request id, API key). Every
	// Ensemble member is a real upstream call and must therefore get its own
	// billing identity; otherwise the parallel calls collapse into one usage row.
	ctx = context.WithValue(ctx, ctxkey.ClientRequestID, fmt.Sprintf("ensemble-%s-%d", uuid.NewString(), index))

	req := parent.Request.Clone(ctx)
	req.Body = io.NopCloser(bytes.NewReader(subBody))
	req.ContentLength = int64(len(subBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Accept-Encoding")
	subCtx.Request = req

	// Carry auth/subscription/ops context established by upstream middleware.
	for key, value := range parent.Keys {
		subCtx.Set(key, value)
	}

	// Run the downstream handler separately so a provider implementation that
	// ignores context cancellation cannot hold the outer Ensemble request open
	// forever. The recorder is isolated per sub-call, so a late write cannot
	// corrupt the client response.
	dispatchDone := make(chan struct{})
	dispatchPanic := make(chan error, 1)
	go func() {
		defer close(dispatchDone)
		defer func() {
			if recovered := recover(); recovered != nil {
				dispatchPanic <- fmt.Errorf("downstream dispatcher panic: %v", recovered)
			}
		}()
		h.dispatch(subCtx)
	}()
	select {
	case <-dispatchDone:
		select {
		case dispatchErr := <-dispatchPanic:
			res.err = dispatchErr
			res.stat.Error = dispatchErr.Error()
			res.stat.DurationMs = time.Since(callStart).Milliseconds()
			return res
		default:
		}
	case <-ctx.Done():
		res.err = fmt.Errorf("model %s timed out after %s", model, timeout)
		res.stat.Error = res.err.Error()
		res.stat.DurationMs = time.Since(callStart).Milliseconds()
		return res
	}

	res.stat.DurationMs = time.Since(callStart).Milliseconds()

	if recorder.Code != http.StatusOK {
		res.err = fmt.Errorf("model %s returned status %d: %s", model, recorder.Code, ensembleTruncate(recorder.Body.String(), 300))
		res.stat.Error = res.err.Error()
		return res
	}

	payload := recorder.Body.Bytes()
	content := extractEnsembleContent(payload)
	if strings.TrimSpace(content) == "" {
		res.err = fmt.Errorf("model %s returned an empty completion", model)
		res.stat.Error = res.err.Error()
		return res
	}

	res.content = content
	res.stat.Status = ensembleStatusOK
	res.stat.PromptTokens = int(gjson.GetBytes(payload, "usage.prompt_tokens").Int())
	res.stat.CompletionTokens = int(gjson.GetBytes(payload, "usage.completion_tokens").Int())
	res.stat.Content = content
	res.stat.Cost = extractEnsembleReportedCost(payload)
	if res.stat.Cost != nil {
		res.stat.CostSource = "upstream"
	} else if h.costEstimator != nil {
		apiKey, _ := middleware2.GetAPIKeyFromContext(parent)
		if apiKey != nil && apiKey.Group != nil {
			estimate, estimateErr := h.costEstimator.EstimateEnsembleCost(
				parent.Request.Context(),
				apiKey.Group.ID,
				model,
				platform,
				service.UsageTokens{
					InputTokens:  res.stat.PromptTokens,
					OutputTokens: res.stat.CompletionTokens,
				},
			)
			if estimateErr == nil && estimate != nil {
				res.stat.Cost = &estimate.Cost
				res.stat.CostSource = estimate.Source
			}
		}
	}
	return res
}

// prepareEnsembleSubCallBody rewrites the client body for one member call:
// the member model replaces the requested model and streaming is disabled,
// because the full text is required before aggregation can run.
func prepareEnsembleSubCallBody(body []byte, model string, maxTokens int) ([]byte, error) {
	out, err := sjson.SetBytes(body, "model", model)
	if err != nil {
		return nil, err
	}
	out, err = sjson.SetBytes(out, "stream", false)
	if err != nil {
		return nil, err
	}
	// stream_options is only valid alongside stream=true.
	if out, err = sjson.DeleteBytes(out, "stream_options"); err != nil {
		return nil, err
	}
	if maxTokens > 0 {
		if out, err = sjson.SetBytes(out, "max_tokens", maxTokens); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// buildEnsembleAggregatorBody appends the aggregation instruction plus every
// proposal as a trailing user message, preserving the original conversation.
func buildEnsembleAggregatorBody(body []byte, proposals []ensembleProposal) ([]byte, error) {
	var sb strings.Builder
	_, _ = sb.WriteString("以下是多个模型对同一问题给出的候选回答。请综合这些回答，给出一个更完整、更准确的最终答案。\n")
	_, _ = sb.WriteString("要求：\n")
	_, _ = sb.WriteString("1. 直接给出最终答案，不要提及这些候选方案，也不要评价它们；\n")
	_, _ = sb.WriteString("2. 保留其中正确且有价值的细节，纠正明显的错误；\n")
	_, _ = sb.WriteString("3. 若候选回答互相矛盾，选择更可靠的一种，并保持答案内部一致。\n\n")
	for i, proposal := range proposals {
		_, _ = sb.WriteString(fmt.Sprintf("【方案%d - %s】\n%s\n\n", i+1, proposal.Model, proposal.Content))
	}
	if len(body)+sb.Len() > maxEnsembleAggregatorBodyBytes {
		return nil, fmt.Errorf("aggregator request exceeds %d bytes", maxEnsembleAggregatorBodyBytes)
	}

	message := map[string]string{"role": "user", "content": sb.String()}
	return sjson.SetBytes(body, "messages.-1", message)
}

// extractEnsembleContent pulls assistant text out of a Chat Completions payload,
// handling both plain string content and the content-parts array form.
func extractEnsembleContent(payload []byte) string {
	content := gjson.GetBytes(payload, "choices.0.message.content")
	if !content.Exists() {
		return ""
	}
	if content.Type == gjson.String {
		return content.String()
	}
	if content.IsArray() {
		var sb strings.Builder
		content.ForEach(func(_, part gjson.Result) bool {
			if text := part.Get("text"); text.Exists() {
				_, _ = sb.WriteString(text.String())
			}
			return true
		})
		return sb.String()
	}
	return ""
}

// extractEnsembleReportedCost only forwards a cost explicitly present in the
// sub-response. When it is absent, the caller may use the configured pricing
// estimator while the normal gateway remains responsible for actual billing.
func extractEnsembleReportedCost(payload []byte) *float64 {
	for _, path := range []string{"usage.cost", "usage.cost_usd"} {
		value := gjson.GetBytes(payload, path)
		if !value.Exists() || value.Type != gjson.Number {
			continue
		}
		cost := value.Float()
		return &cost
	}
	return nil
}

func longestEnsembleProposal(proposals []ensembleProposal) string {
	best := ""
	for _, proposal := range proposals {
		if len(proposal.Content) > len(best) {
			best = proposal.Content
		}
	}
	return best
}

// buildEnsembleChatResponse assembles the client-facing Chat Completions payload.
// Usage is the sum across member calls so the client sees the true token cost.
func buildEnsembleChatResponse(
	model string,
	content string,
	stats []ensembleMemberStat,
	exposeMetadata bool,
	aggregated bool,
	elapsed time.Duration,
) map[string]any {
	promptTokens, completionTokens := 0, 0
	succeeded := 0
	for _, stat := range stats {
		promptTokens += stat.PromptTokens
		completionTokens += stat.CompletionTokens
		if stat.Status == ensembleStatusOK {
			succeeded++
		}
	}

	payload := map[string]any{
		"id":      fmt.Sprintf("chatcmpl-ensemble-%d", time.Now().UnixNano()),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]any{
			{
				"index":         0,
				"message":       map[string]any{"role": "assistant", "content": content},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      promptTokens + completionTokens,
		},
	}

	if exposeMetadata {
		payload["ensemble_metadata"] = map[string]any{
			"members":           stats,
			"members_total":     len(stats),
			"members_succeeded": succeeded,
			"aggregated":        aggregated,
			"duration_ms":       elapsed.Milliseconds(),
		}
	}
	return payload
}

// writeStreamResponse emits the aggregated answer as a minimal SSE stream for
// clients that requested stream=true. Member calls always run non-streaming, so
// the text is delivered as a single content chunk followed by a terminal chunk.
func (h *EnsembleHandler) writeStreamResponse(c *gin.Context, model, content string, payload map[string]any) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	id := fmt.Sprintf("chatcmpl-ensemble-%d", time.Now().UnixNano())
	created := time.Now().Unix()

	writeChunk := func(chunk map[string]any) {
		encoded, err := json.Marshal(chunk)
		if err != nil {
			return
		}
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", encoded)
		c.Writer.Flush()
	}

	writeChunk(map[string]any{
		"id": id, "object": "chat.completion.chunk", "created": created, "model": model,
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{"role": "assistant", "content": content}, "finish_reason": nil},
		},
	})

	terminal := map[string]any{
		"id": id, "object": "chat.completion.chunk", "created": created, "model": model,
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{}, "finish_reason": "stop"},
		},
		"usage": payload["usage"],
	}
	if metadata, ok := payload["ensemble_metadata"]; ok {
		terminal["ensemble_metadata"] = metadata
	}
	writeChunk(terminal)

	_, _ = fmt.Fprint(c.Writer, "data: [DONE]\n\n")
	c.Writer.Flush()
}

func (h *EnsembleHandler) errorResponse(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

func ensembleTruncate(s string, limit int) string {
	s = strings.TrimSpace(s)
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}
