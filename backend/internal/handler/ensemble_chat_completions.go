package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/securityaudit"
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
	runtime                  *service.EnsembleRuntimeService
	dispatch                 gin.HandlerFunc
	responsesDispatch        gin.HandlerFunc
	costEstimator            service.EnsembleCostEstimator
	securityAuditCoordinator *securityaudit.Coordinator
	contentModerationService *service.ContentModerationService
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

// SetResponsesSubCallDispatcher injects the Responses API routing closure used
// by the aggregator-only compact path. It is separate from the Chat
// Completions dispatcher because the two protocols have different wire
// formats, response parsers, and usage paths.
func (h *EnsembleHandler) SetResponsesSubCallDispatcher(dispatch gin.HandlerFunc) {
	h.responsesDispatch = dispatch
}

// ensembleProposal is one successful member answer.
type ensembleProposal struct {
	Model   string
	Content string
}

// ensembleMemberStat is the per-member record exposed via ensemble_metadata.
type ensembleMemberStat struct {
	Model            string `json:"model"`
	Platform         string `json:"platform,omitempty"`
	Role             string `json:"role"`
	Status           string `json:"status"`
	DurationMs       int64  `json:"duration_ms"`
	PromptTokens     int    `json:"prompt_tokens,omitempty"`
	CompletionTokens int    `json:"completion_tokens,omitempty"`
	// CachedTokens is the member's prompt-cache hit count as reported by its
	// upstream usage. The aggregate usage object sums it across members so a
	// caller (e.g. the ZCode client's cache-hit display) sees the same
	// prompt_tokens_details.cached_tokens shape a direct call would return.
	CachedTokens int      `json:"cached_tokens,omitempty"`
	Content      string   `json:"content,omitempty"`
	Cost         *float64 `json:"cost"`
	CostSource   string   `json:"cost_source,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// ensembleSubResult is the internal outcome of one member sub-call.
type ensembleSubResult struct {
	index     int
	model     string
	role      string
	content   string
	toolCalls []map[string]any
	stat      ensembleMemberStat
	err       error
}

type ensembleResponsesSubResult struct {
	status int
	header http.Header
	body   []byte
	err    error
}

const (
	ensembleStatusOK               = "ok"
	ensembleStatusFailed           = "failed"
	maxEnsembleAggregatorBodyBytes = 2 << 20
	ensembleProgressSinkKey        = "ensemble.progress.sink"
	// ensembleDiagnosticKey marks the admin diagnostic run. Only that path needs
	// every member's full answer text echoed back in ensemble_metadata; putting it
	// on production responses multiplies the payload by the member count for data
	// no caller reads.
	ensembleDiagnosticKey = "ensemble.diagnostic"
	// ensembleFailureNoticePrefix marks the in-band failure text emitted by fail().
	//
	// That text has to ride delta.content to stay visible in every client, which
	// means the client stores it as a normal assistant turn and echoes it back on
	// every later turn. Without a marker the fan-out cannot tell its own failure
	// notice apart from a real answer, and "Only 0 proposers succeeded…" becomes
	// permanent conversation history fed to every member. The prefix is visible on
	// purpose: a caller reading the stream should be able to see that the line came
	// from the gateway and not from a model.
	ensembleFailureNoticePrefix = "[ensemble] "
	// ensembleKeepAliveInterval bounds how long a streaming client waits without
	// receiving any bytes. Fan-out plus aggregation can legitimately take minutes,
	// which is longer than the idle-read timeout of most SDKs and proxies.
	ensembleKeepAliveInterval = 15 * time.Second
	// A provider can occasionally return HTTP 200 without a visible assistant
	// message. Retry that malformed-but-transient response once; other failures
	// are already handled by the normal gateway failover path.
	ensembleEmptyCompletionRetryCount = 1
)

var errEnsembleEmptyCompletion = errors.New("empty ensemble completion")

// EnsembleProgressEvent is the small, stable event contract for observing one
// fan-out. It feeds both the authenticated admin diagnostic stream and, when the
// group enables stream_trace, the execution trace sent to a streaming caller.
type EnsembleProgressEvent struct {
	Type string `json:"type"`
	// Index is the member's position in the fan-out. It is what lets the
	// client-facing trace name a member ("模型 2") when expose_metadata forbids
	// revealing the real model id.
	Index              int                 `json:"index,omitempty"`
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

// ensembleAccountPoolContext extends a sub-call context with the account-pool
// group set: the caller's own group plus every configured source group. The
// scheduler then picks each member's account from that union exactly as it
// would for a direct call to the source group, so ensemble requests never
// depend on hand-bound accounts and inherit the normal load balancing and
// failover behaviour.
func ensembleAccountPoolContext(ctx context.Context, parent *gin.Context) context.Context {
	if ctx == nil || parent == nil {
		return ctx
	}
	apiKey, ok := middleware2.GetAPIKeyFromContext(parent)
	if !ok || apiKey == nil || apiKey.Group == nil {
		return ctx
	}
	pool := make([]int64, 0, 1+len(apiKey.Group.EnsembleConfig.SourceGroupIDs))
	pool = append(pool, apiKey.Group.ID)
	pool = append(pool, apiKey.Group.EnsembleConfig.SourceGroupIDs...)
	return service.WithAccountPoolGroupIDs(ctx, pool)
}

// addEnsembleProgressSink attaches an observer without displacing an existing
// one, so the admin diagnostic stream and a caller's execution trace can watch
// the same run. It must be called before the fan-out goroutines start: gin's
// context guards Keys, but the sink has to be in place before the first event or
// that event is silently dropped.
func addEnsembleProgressSink(c *gin.Context, sink func(EnsembleProgressEvent)) {
	if c == nil || sink == nil {
		return
	}
	if existingValue, ok := c.Get(ensembleProgressSinkKey); ok {
		if existing, ok := existingValue.(func(EnsembleProgressEvent)); ok && existing != nil {
			c.Set(ensembleProgressSinkKey, func(event EnsembleProgressEvent) {
				existing(event)
				sink(event)
			})
			return
		}
	}
	c.Set(ensembleProgressSinkKey, sink)
}

// ChatCompletions handles POST /v1/chat/completions for platform=ensemble groups.
func (h *EnsembleHandler) ChatCompletions(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.Group == nil {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
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
	if decision := h.checkSecurityAudit(c, reqLog, apiKey, subject, service.ContentModerationProtocolOpenAIChat, requestedModel, body); decision != nil && !decision.AllowNextStage {
		h.openAISecurityAuditError(c, decision)
		return
	}

	// The execution trace rides delta.reasoning_content, and reasoning-capable
	// clients persist that delta on the assistant message and echo it back on
	// the next turn. Members must never see it: the trace is presentation for
	// the caller, not model output, and feeding it back would pollute every
	// member's context. Strip it once here so proposers and the aggregator both
	// receive a clean conversation.
	body, err = stripEnsembleReasoningFields(body)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}

	clientWantsStream := gjson.GetBytes(body, "stream").Bool()
	// Only the admin diagnostic run echoes every member's full answer back in
	// ensemble_metadata; see ensembleDiagnosticKey.
	_, includeMemberContent := c.Get(ensembleDiagnosticKey)

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
	// Open the SSE response before the fan-out so a streaming client starts
	// receiving bytes immediately. Past this point the status code is fixed at 200
	// and every failure has to be reported through the stream.
	var stream *ensembleStreamWriter
	if clientWantsStream {
		stream = h.beginStreamResponse(c, requestedModel, plan.EffectiveStreamTrace(), plan.Config.ExposeMetadata)
		defer stream.stopKeepAlive()
		// Progress events are emitted from the fan-out goroutines below. Attaching
		// the stream as a sink here — before the first event — is what turns those
		// previously admin-only events into the caller's live execution trace.
		// An existing sink (the admin diagnostic stream) is preserved: both need to
		// observe the same run.
		addEnsembleProgressSink(c, func(event EnsembleProgressEvent) {
			stream.traceLine(stream.traceProgress(event))
		})
	}

	// Emitted after the sink is attached so the opening line reaches the client.
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
				Index:    idx,
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
					Index:  idx,
					Model:  stat.Model,
					Role:   stat.Role,
					Status: stat.Status,
					Member: &stat,
				})
			}()
			// Proposers run in parallel, so interleaving several models' reasoning on
			// one stream would be unreadable. Only the aggregator's reasoning is
			// forwarded; proposers stay buffered.
			result = h.runSubCall(c, idx, member.Model, member.Platform, service.EnsembleRoleProposer, body, plan.Config.MaxTokens, timeout, nil)
		}(i, plan.Proposers[i])
	}
	wg.Wait()

	proposals := make([]ensembleProposal, 0, len(results))
	stats := make([]ensembleMemberStat, 0, len(results)+1)
	structured := []map[string]any(nil)
	// structuredPreface is the prose the *same* member emitted alongside its tool
	// calls. Models routinely explain what they are about to do before calling a
	// tool, and on a tool-call turn that sentence is the only thing the caller can
	// read. It is taken from the member that won the tool-call race rather than
	// from the proposal pool: pairing one model's tool_calls with another model's
	// prose would narrate an action that member never requested.
	structuredPreface := ""
	successfulMembers := 0
	for _, res := range results {
		stats = append(stats, res.stat)
		if res.err == nil && (strings.TrimSpace(res.content) != "" || len(res.toolCalls) > 0) {
			successfulMembers++
			if len(res.toolCalls) > 0 && len(structured) == 0 {
				structured = res.toolCalls
				structuredPreface = res.content
			}
			if strings.TrimSpace(res.content) != "" {
				proposals = append(proposals, ensembleProposal{Model: res.model, Content: res.content})
			}
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
	if len(proposals) < minProposers && successfulMembers < minProposers {
		reqLog.Error("ensemble did not reach min_proposers",
			zap.Int("succeeded", len(proposals)),
			zap.Int("required", minProposers))
		message := fmt.Sprintf("Only %d proposers succeeded, minimum %d required", len(proposals), minProposers)
		emitEnsembleProgress(c, EnsembleProgressEvent{
			Type:               "error",
			Error:              message,
			ProposersTotal:     len(plan.Proposers),
			ProposersSucceeded: len(proposals),
		})
		if stream != nil {
			stream.fail("api_error", message, http.StatusBadGateway)
			return
		}
		h.errorResponse(c, http.StatusBadGateway, "api_error", message)
		return
	}
	// Tool calls are a continuation boundary for agent clients. They must stay
	// structured so the client can execute the requested tool and send the next
	// turn back with its history; converting them to candidate text would end
	// the agent loop after the first answer.
	//
	// The member's own preface travels with them. Aggregation is skipped on this
	// turn (there is no answer to synthesize yet), so discarding that prose left
	// the caller with a silent tool invocation and no statement of intent.
	if len(structured) > 0 {
		payload := buildEnsembleChatResponse(requestedModel, structuredPreface, structured, stats, plan.Config.ExposeMetadata, includeMemberContent, false, time.Since(started))
		if stream != nil {
			stream.finish(structuredPreface, payload)
			return
		}
		c.JSON(http.StatusOK, payload)
		return
	}

	// Aggregate, or fall back to the longest proposal.
	finalContent := longestEnsembleProposal(proposals)
	finalToolCalls := []map[string]any(nil)
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
				Index:    len(results),
				Model:    plan.Aggregator.Model,
				Platform: plan.Aggregator.Platform,
				Role:     service.EnsembleRoleAggregator,
			})
			// The aggregator writes the answer the caller actually receives, so its
			// reasoning is the one piece of real model thinking that belongs on the
			// caller's stream. A direct single-model call forwards reasoning_content
			// verbatim; without this the ensemble silently drops it and the caller
			// only ever sees our scheduling trace.
			var aggReasoning ensembleReasoningSink
			if stream != nil {
				aggReasoning = stream.modelReasoning
			}
			aggRes := h.runSubCall(c, len(results), plan.Aggregator.Model, plan.Aggregator.Platform, service.EnsembleRoleAggregator,
				aggBody, plan.Config.MaxTokens, timeout, aggReasoning)
			stats = append(stats, aggRes.stat)
			stat := aggRes.stat
			emitEnsembleProgress(c, EnsembleProgressEvent{
				Type:   "member_finished",
				Index:  len(results),
				Model:  stat.Model,
				Role:   stat.Role,
				Status: stat.Status,
				Member: &stat,
			})
			if aggRes.err == nil && (strings.TrimSpace(aggRes.content) != "" || len(aggRes.toolCalls) > 0) {
				finalContent = aggRes.content
				finalToolCalls = aggRes.toolCalls
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

	payload := buildEnsembleChatResponse(requestedModel, finalContent, finalToolCalls, stats, plan.Config.ExposeMetadata, includeMemberContent, aggregated, time.Since(started))
	if stream != nil {
		stream.finish(finalContent, payload)
		return
	}
	c.JSON(http.StatusOK, payload)
}

// Compact handles the OpenAI Responses compaction endpoint for Ensemble
// groups. Compaction is context-state maintenance, not a user question: it
// must run once on the configured aggregator and must never fan out to
// proposers or append candidate answers to the compaction input.
func (h *EnsembleHandler) Compact(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.Group == nil {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	if apiKey.Group.Platform != service.PlatformEnsemble {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "The API key is not bound to an Ensemble group")
		return
	}
	if h.runtime == nil || h.responsesDispatch == nil {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "Ensemble Responses runtime is not available")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 || !gjson.ValidBytes(body) {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}
	plan, err := h.runtime.LoadPlan(c.Request.Context(), apiKey.Group.ID, apiKey.Group.EnsembleConfig)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "Failed to load ensemble configuration")
		return
	}
	if !plan.ShouldAggregate() || strings.TrimSpace(plan.Aggregator.Model) == "" {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Ensemble compact requires an enabled aggregator model")
		return
	}

	// The compact body is client conversation plus echoed reasoning fields from
	// previous ensemble turns; strip them the same way the fan-out path does so
	// the compaction input is clean model context.
	body, err = stripEnsembleReasoningFields(body)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid compact request body")
		return
	}

	subBody, err := sjson.SetBytes(body, "model", plan.Aggregator.Model)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid compact request body")
		return
	}
	result := h.runResponsesSubCall(c, plan.Aggregator.Model, plan.Aggregator.Platform, subBody, time.Duration(plan.EffectiveTimeoutSeconds())*time.Second)
	if result.err != nil {
		h.errorResponse(c, http.StatusBadGateway, "api_error", result.err.Error())
		return
	}
	for key, values := range result.header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	c.Status(result.status)
	_, _ = c.Writer.Write(result.body)
}

func (h *EnsembleHandler) runResponsesSubCall(
	parent *gin.Context,
	model string,
	platform string,
	body []byte,
	timeout time.Duration,
) ensembleResponsesSubResult {
	result := ensembleResponsesSubResult{}
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform == "" {
		platform, _ = service.DetectModelPlatform(model)
	}
	if platform == "" {
		result.err = fmt.Errorf("cannot determine upstream platform for model %q", model)
		return result
	}

	recorder := httptest.NewRecorder()
	subCtx, _ := gin.CreateTestContext(recorder)
	ctx, cancel := context.WithTimeout(parent.Request.Context(), timeout)
	defer cancel()
	ctx = service.WithResolvedTargetPlatform(ctx, platform)
	ctx = context.WithValue(ctx, ctxkey.ClientRequestID, fmt.Sprintf("ensemble-compact-%s", uuid.NewString()))
	// Compact runs on the aggregator; it draws the same account pool as a
	// regular member call.
	ctx = ensembleAccountPoolContext(ctx, parent)
	req := parent.Request.Clone(ctx)
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Accept-Encoding")
	// The compact body is rewritten with the aggregator model, so the cloned
	// Content-Length no longer matches. See runSubCallOnce for the rationale.
	req.Header.Del("Content-Length")
	subCtx.Request = req
	for key, value := range parent.Keys {
		subCtx.Set(key, value)
	}

	dispatchDone := make(chan struct{})
	var dispatchPanic error
	go func() {
		defer close(dispatchDone)
		defer func() {
			if recovered := recover(); recovered != nil {
				dispatchPanic = fmt.Errorf("downstream Responses dispatcher panic: %v", recovered)
			}
		}()
		h.responsesDispatch(subCtx)
	}()
	select {
	case <-dispatchDone:
		if dispatchPanic != nil {
			result.err = dispatchPanic
			return result
		}
	case <-ctx.Done():
		result.err = fmt.Errorf("aggregator %s timed out after %s", model, timeout)
		return result
	}
	result.status = recorder.Code
	result.header = recorder.Header().Clone()
	result.body = append([]byte(nil), recorder.Body.Bytes()...)
	if result.status < http.StatusOK || result.status >= http.StatusMultipleChoices {
		result.err = fmt.Errorf("aggregator %s returned status %d: %s", model, result.status, ensembleTruncate(recorder.Body.String(), 300))
	}
	return result
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
	// The diagnostic endpoint emits its own SSE progress events. The inner
	// gateway call must therefore be non-streaming; otherwise its SSE bytes are
	// not valid JSON and cannot be attached to the terminal response event.
	body, err = sjson.SetBytes(body, "stream", false)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid request body")
		return
	}
	body, err = sjson.DeleteBytes(body, "stream_options")
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid request body")
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
	// Diagnostics always need member details for the admin view, while normal
	// production responses follow the group's expose_metadata setting.
	diagnosticGroup := *apiKey.Group
	diagnosticGroup.EnsembleConfig.ExposeMetadata = true
	diagnosticAPIKey := *apiKey
	diagnosticAPIKey.Group = &diagnosticGroup
	diagnosticContext.Set(string(middleware2.ContextKeyAPIKey), &diagnosticAPIKey)
	// The admin view renders every member's raw answer, so this run also opts into
	// the member content that production responses no longer carry.
	diagnosticContext.Set(ensembleDiagnosticKey, true)
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
// reasoningSink, when non-nil, receives the member's own delta.reasoning_content
// as it streams. Only the member whose thinking the caller asked to see passes
// one; the rest run with nil and are buffered exactly as before.
func (h *EnsembleHandler) runSubCall(
	parent *gin.Context,
	index int,
	model string,
	platform string,
	role string,
	body []byte,
	maxTokens int,
	timeout time.Duration,
	reasoningSink ensembleReasoningSink,
) ensembleSubResult {
	for attempt := 0; attempt <= ensembleEmptyCompletionRetryCount; attempt++ {
		result := h.runSubCallOnce(parent, index, model, platform, role, body, maxTokens, timeout, reasoningSink)
		if result.err == nil || !errors.Is(result.err, errEnsembleEmptyCompletion) {
			return result
		}
		if attempt == ensembleEmptyCompletionRetryCount {
			return result
		}
	}
	panic("unreachable")
}

func (h *EnsembleHandler) runSubCallOnce(
	parent *gin.Context,
	index int,
	model string,
	platform string,
	role string,
	body []byte,
	maxTokens int,
	timeout time.Duration,
	reasoningSink ensembleReasoningSink,
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
	// The member streams internally so the normal gateway records TTFT and usage,
	// and the Ensemble boundary needs the whole body to normalize. The tee adds a
	// live read of delta.reasoning_content on top of that buffering, which is how a
	// caller sees the model actually thinking instead of only our scheduling trace.
	// Proposers pass a nil sink and stay buffered exactly as before.
	tee := newEnsembleReasoningTee(recorder, reasoningSink)
	// The dispatcher runs in its own goroutine, and a provider that ignores context
	// cancellation can still be writing after this function returns on timeout. The
	// recorder is per-sub-call so a late write there is harmless, but the sink
	// writes to the shared client stream, so it must be closed on every return path.
	defer tee.disable()
	subCtx, _ := gin.CreateTestContext(tee)

	ctx, cancel := context.WithTimeout(parent.Request.Context(), timeout)
	defer cancel()
	ctx = service.WithResolvedTargetPlatform(ctx, platform)
	// Usage billing deduplicates by (client request id, API key). Every
	// Ensemble member is a real upstream call and must therefore get its own
	// billing identity; otherwise the parallel calls collapse into one usage row.
	ctx = context.WithValue(ctx, ctxkey.ClientRequestID, fmt.Sprintf("ensemble-%s-%d", uuid.NewString(), index))
	// Members draw their account pool from the caller's group plus the
	// configured source groups, so each model is scheduled exactly like a direct
	// call to its source group: load balancing, failover and capacity all apply,
	// and the caller never depends on a hand-bound account.
	ctx = ensembleAccountPoolContext(ctx, parent)

	req := parent.Request.Clone(ctx)
	req.Body = io.NopCloser(bytes.NewReader(subBody))
	req.ContentLength = int64(len(subBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Accept-Encoding")
	// Clone copies the client's headers verbatim, but the sub-call body is a
	// different length (model swap, stream flags, max_tokens cap). Leaving the
	// original Content-Length behind means any consumer that trusts the header
	// over req.ContentLength reads a truncated body. The struct field is the
	// single source of truth, so drop the stale header.
	req.Header.Del("Content-Length")
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

	payload := normalizeEnsembleChatPayload(recorder.Body.Bytes())
	content := extractEnsembleContent(payload)
	toolCalls := extractEnsembleToolCalls(payload)
	if strings.TrimSpace(content) == "" && len(toolCalls) == 0 {
		res.err = fmt.Errorf("model %s returned an empty completion: %w", model, errEnsembleEmptyCompletion)
		res.stat.Error = res.err.Error()
		return res
	}

	res.content = content
	res.toolCalls = toolCalls
	res.stat.Status = ensembleStatusOK
	res.stat.PromptTokens = int(gjson.GetBytes(payload, "usage.prompt_tokens").Int())
	res.stat.CompletionTokens = int(gjson.GetBytes(payload, "usage.completion_tokens").Int())
	// OpenAI-compatible upstreams report prompt-cache hits as
	// usage.prompt_tokens_details.cached_tokens; some Responses-style bodies use
	// input_tokens_details.cached_tokens. Both are accepted so the aggregated
	// usage can surface the same cache-hit figure a direct call shows.
	res.stat.CachedTokens = int(firstEnsembleCacheHit(gjson.GetBytes(payload, "usage")))
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

// stripEnsembleReasoningFields removes client-echoed reasoning fields from the
// conversation before it fans out to members.
//
// The execution trace is delivered to the caller as delta.reasoning_content. A
// reasoning-capable client (Cherry Studio, NextChat, …) stores that delta on
// the assistant message and sends it back verbatim on the next turn. Without
// this pass the trace lines — "[0.0s] 并行调用 2 个模型…" — would be fed to
// every member model as if they were the model's own thinking, polluting the
// context the members aggregate over. The trace is presentation for the caller,
// never model input, so it is dropped at the ensemble boundary.
//
// Three shapes carry echoed reasoning, and all three have to go:
//   - a top-level field on the assistant message (reasoning_content, the OpenAI
//     reasoning spelling, or Anthropic's thinking);
//   - a typed block inside a content array, which is how Anthropic-shaped clients
//     store thinking — deleting the sibling field would leave the block untouched;
//   - our own in-band failure notice, which rides delta.content to stay visible
//     and is therefore stored by the client as a normal assistant turn.
//
// Real content and tool_calls are conversation facts and must survive.
func stripEnsembleReasoningFields(body []byte) ([]byte, error) {
	// Cheap reject for the common case. The probes are deliberately loose (no
	// quotes, so "redacted_thinking" and "reasoning_content" both match): a false
	// positive only costs one full scan, while a false negative would silently let
	// echoed reasoning through.
	if !bytes.Contains(body, []byte("reasoning")) &&
		!bytes.Contains(body, []byte("thinking")) &&
		!bytes.Contains(body, []byte(ensembleFailureNoticePrefix)) {
		return body, nil
	}
	messages := gjson.GetBytes(body, "messages")
	if !messages.Exists() || !messages.IsArray() {
		return body, nil
	}

	out := body
	// Whole-message deletions are collected and applied last, in descending order:
	// removing an array element shifts every later index, which would corrupt the
	// per-field edits below if the two were interleaved.
	drop := make([]int, 0, 4)
	index := 0
	messages.ForEach(func(_, message gjson.Result) bool {
		current := index
		index++
		if message.Get("role").String() != "assistant" {
			return true
		}
		if ensembleIsFailureNotice(message) {
			drop = append(drop, current)
			return true
		}
		prefix := fmt.Sprintf("messages.%d.", current)
		for _, field := range []string{"reasoning_content", "reasoning", "thinking"} {
			if message.Get(field).Exists() {
				out, _ = sjson.DeleteBytes(out, prefix+field)
			}
		}
		var emptied bool
		out, emptied = stripEnsembleReasoningBlocks(out, message, current)
		// An assistant turn whose content was nothing but thinking carries no
		// conversation fact once the thinking is gone, and an empty content array is
		// rejected by some upstreams. Drop the turn instead of sending a hollow one.
		if emptied && !message.Get("tool_calls").Exists() {
			drop = append(drop, current)
		}
		return true
	})

	for i := len(drop) - 1; i >= 0; i-- {
		out, _ = sjson.DeleteBytes(out, fmt.Sprintf("messages.%d", drop[i]))
	}
	return out, nil
}

// stripEnsembleReasoningBlocks removes typed reasoning blocks from an assistant
// message's content array, reporting whether the array ended up empty.
//
// Anthropic-shaped clients store thinking as content blocks rather than as a
// sibling field, so a pass that only deletes messages[N].thinking leaves the
// echoed reasoning sitting inside messages[N].content.
func stripEnsembleReasoningBlocks(out []byte, message gjson.Result, messageIndex int) ([]byte, bool) {
	content := message.Get("content")
	if !content.IsArray() {
		return out, false
	}
	total := 0
	drop := make([]int, 0, 2)
	content.ForEach(func(_, block gjson.Result) bool {
		current := total
		total++
		switch block.Get("type").String() {
		case "thinking", "redacted_thinking", "reasoning":
			drop = append(drop, current)
		}
		return true
	})
	if len(drop) == 0 {
		return out, false
	}
	for i := len(drop) - 1; i >= 0; i-- {
		out, _ = sjson.DeleteBytes(out, fmt.Sprintf("messages.%d.content.%d", messageIndex, drop[i]))
	}
	return out, len(drop) == total
}

// ensembleIsFailureNotice reports whether an assistant message is the in-band
// failure text this handler emitted on an earlier turn.
//
// A turn carrying tool_calls is never a failure notice, and dropping one would
// break the agent loop, so it is excluded regardless of its text.
func ensembleIsFailureNotice(message gjson.Result) bool {
	if message.Get("tool_calls").Exists() {
		return false
	}
	content := message.Get("content")
	if content.Type == gjson.String {
		return strings.HasPrefix(content.String(), ensembleFailureNoticePrefix)
	}
	if content.IsArray() {
		text := content.Get("0.text")
		return text.Type == gjson.String && strings.HasPrefix(text.String(), ensembleFailureNoticePrefix)
	}
	return false
}

// prepareEnsembleSubCallBody rewrites the client body for one member call.
// Members stream internally so the normal gateway records TTFT and usage; the
// Ensemble boundary consumes the SSE and waits for the complete answer before
// building the aggregator request.
func prepareEnsembleSubCallBody(body []byte, model string, maxTokens int) ([]byte, error) {
	out, err := sjson.SetBytes(body, "model", model)
	if err != nil {
		return nil, err
	}
	out, err = sjson.SetBytes(out, "stream", true)
	if err != nil {
		return nil, err
	}
	// Usage-only terminal chunks are needed by the normal billing path when the
	// upstream supports them. Preserve all client stream options and enable the
	// usage flag without requiring clients to opt into the internal protocol.
	if out, err = sjson.SetBytes(out, "stream_options.include_usage", true); err != nil {
		return nil, err
	}
	return applyEnsembleMaxTokens(out, maxTokens)
}

// applyEnsembleMaxTokens applies the group's per-member output limit as a ceiling.
//
// Three properties matter, and unconditionally writing max_tokens broke all
// three:
//
//   - It is a ceiling, not an override. A client that asked for fewer tokens than
//     the group limit meant it, and raising the limit spends the caller's quota on
//     output nobody asked for.
//   - It has to land on the field the client actually used. Reasoning models
//     reject max_tokens in favour of max_completion_tokens, and some upstreams
//     reject a body carrying both, so introducing the other spelling can turn a
//     working request into a 400.
//   - It must never cut below a declared thinking budget. Anthropic requires
//     max_tokens to exceed thinking.budget_tokens, so a limit that violates that
//     makes every member fail with invalid_request_error instead of merely
//     answering more briefly. When the limit cannot be honoured, the client's own
//     value stands.
func applyEnsembleMaxTokens(body []byte, maxTokens int) ([]byte, error) {
	if maxTokens <= 0 {
		return body, nil
	}
	field := "max_tokens"
	if !gjson.GetBytes(body, field).Exists() && gjson.GetBytes(body, "max_completion_tokens").Exists() {
		field = "max_completion_tokens"
	}
	if current := gjson.GetBytes(body, field); current.Exists() && current.Int() > 0 && current.Int() <= int64(maxTokens) {
		return body, nil
	}
	// Reasoning budgets are spelled differently per protocol; the limit has to
	// clear whichever one this request carries.
	for _, path := range []string{
		"thinking.budget_tokens",
		"reasoning.max_tokens",
		"extra_body.thinking.budget_tokens",
		"output_config.thinking.budget_tokens",
	} {
		if budget := gjson.GetBytes(body, path); budget.Exists() && budget.Int() >= int64(maxTokens) {
			return body, nil
		}
	}
	return sjson.SetBytes(body, field, maxTokens)
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

// normalizeEnsembleChatPayload converts an internally streamed Chat
// Completions response into the same JSON shape used by a non-stream response.
// The downstream gateway has already normalized all provider-specific streams
// to this wire format, so the parser only needs to merge delta content,
// tool-call fragments, finish_reason, and the terminal usage object.
func normalizeEnsembleChatPayload(payload []byte) []byte {
	// The fast path requires the payload to be a single valid JSON document.
	// gjson scans leniently, so on an SSE body it happily matches "choices.0"
	// inside the first data: line and would return the raw stream unparsed.
	if gjson.ValidBytes(payload) && gjson.GetBytes(payload, "choices.0").Exists() {
		return payload
	}
	if !bytes.Contains(payload, []byte("data:")) {
		return payload
	}

	var content strings.Builder
	toolCalls := make(map[int]map[string]any)
	finishReason := "stop"
	var usage json.RawMessage
	responseID := ""
	model := ""
	scanner := bufio.NewScanner(bytes.NewReader(payload))
	// A provider can emit one very large SSE line (a big content delta, or a long
	// tool-call argument fragment). At the default 64KB token limit Scan would stop
	// early and this function would silently return a truncated answer, so allow
	// lines up to the same order as the aggregator body cap.
	scanner.Buffer(make([]byte, 0, 64*1024), maxEnsembleAggregatorBodyBytes)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" || !gjson.Valid(data) {
			continue
		}
		if responseID == "" {
			responseID = gjson.Get(data, "id").String()
		}
		if model == "" {
			model = gjson.Get(data, "model").String()
		}
		if usageResult := gjson.Get(data, "usage"); usageResult.Exists() && usageResult.IsObject() {
			usage = json.RawMessage(usageResult.Raw)
		}
		choice := gjson.Get(data, "choices.0")
		if reason := choice.Get("finish_reason"); reason.Exists() && reason.Type == gjson.String && reason.String() != "" {
			finishReason = reason.String()
		}
		if delta := choice.Get("delta"); delta.Exists() {
			if deltaContent := delta.Get("content"); deltaContent.Exists() && deltaContent.Type == gjson.String {
				_, _ = content.WriteString(deltaContent.String())
			}
			if calls := delta.Get("tool_calls"); calls.IsArray() {
				calls.ForEach(func(_, call gjson.Result) bool {
					index := int(call.Get("index").Int())
					current, exists := toolCalls[index]
					if !exists {
						current = map[string]any{"index": index, "type": "function", "function": map[string]any{}}
						toolCalls[index] = current
					}
					if id := call.Get("id"); id.Exists() && id.Type == gjson.String {
						current["id"] = id.String()
					}
					if kind := call.Get("type"); kind.Exists() && kind.Type == gjson.String {
						current["type"] = kind.String()
					}
					fn, _ := current["function"].(map[string]any)
					if name := call.Get("function.name"); name.Exists() && name.Type == gjson.String {
						previous, _ := fn["name"].(string)
						fn["name"] = previous + name.String()
					}
					if arguments := call.Get("function.arguments"); arguments.Exists() && arguments.Type == gjson.String {
						previous, _ := fn["arguments"].(string)
						fn["arguments"] = previous + arguments.String()
					}
					return true
				})
			}
		}
	}

	if scanner.Err() != nil {
		// Returning the raw payload makes the caller report an empty completion and
		// retry, which is correct: a partially scanned stream must never be passed
		// off as a complete member answer.
		return payload
	}

	callList := make([]map[string]any, 0, len(toolCalls))
	indexes := make([]int, 0, len(toolCalls))
	for index := range toolCalls {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		callList = append(callList, toolCalls[index])
	}
	message := map[string]any{"role": "assistant"}
	if content.Len() > 0 {
		message["content"] = content.String()
	}
	if len(callList) > 0 {
		message["tool_calls"] = callList
		finishReason = "tool_calls"
	}
	choice := map[string]any{"index": 0, "message": message, "finish_reason": finishReason}
	result := map[string]any{
		"id":      responseID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]any{choice},
	}
	if len(usage) > 0 {
		var usageValue map[string]any
		if json.Unmarshal(usage, &usageValue) == nil {
			result["usage"] = usageValue
		}
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return payload
	}
	return encoded
}

// extractEnsembleContent pulls assistant text out of a normalized Chat
// Completions payload, handling both plain string content and content-parts.
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

func extractEnsembleToolCalls(payload []byte) []map[string]any {
	raw := gjson.GetBytes(payload, "choices.0.message.tool_calls")
	if !raw.Exists() || !raw.IsArray() {
		return nil
	}
	var calls []map[string]any
	if err := json.Unmarshal([]byte(raw.Raw), &calls); err != nil {
		return nil
	}
	return calls
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

// firstEnsembleCacheHit returns the prompt-cache hit count from an upstream
// usage object, preferring the OpenAI Chat Completions spelling
// (prompt_tokens_details.cached_tokens) over the Responses spelling
// (input_tokens_details.cached_tokens), matching openAIUsageFromGJSON.
func firstEnsembleCacheHit(usage gjson.Result) int64 {
	if !usage.Exists() || !usage.IsObject() {
		return 0
	}
	for _, path := range []string{"prompt_tokens_details.cached_tokens", "input_tokens_details.cached_tokens"} {
		if v := usage.Get(path); v.Exists() && v.Type == gjson.Number {
			return v.Int()
		}
	}
	return 0
}

// buildEnsembleChatResponse assembles the client-facing Chat Completions payload.
// Usage is the sum across member calls so the client sees the true token cost.
func buildEnsembleChatResponse(
	model string,
	content string,
	toolCalls []map[string]any,
	stats []ensembleMemberStat,
	exposeMetadata bool,
	includeMemberContent bool,
	aggregated bool,
	elapsed time.Duration,
) map[string]any {
	promptTokens, completionTokens, cachedTokens := 0, 0, 0
	succeeded := 0
	for _, stat := range stats {
		promptTokens += stat.PromptTokens
		completionTokens += stat.CompletionTokens
		cachedTokens += stat.CachedTokens
		if stat.Status == ensembleStatusOK {
			succeeded++
		}
	}

	message := map[string]any{"role": "assistant", "content": content}
	finishReason := "stop"
	if len(toolCalls) > 0 {
		// OpenAI's own tool-call shape carries content alongside tool_calls, so a
		// model's preface ("我先查一下配置文件…") is kept. Only genuinely empty prose
		// is dropped, which is the null-content turn a client expects when the model
		// requested a tool without saying anything.
		if strings.TrimSpace(content) == "" {
			delete(message, "content")
		}
		message["tool_calls"] = toolCalls
		finishReason = "tool_calls"
	}
	usage := map[string]any{
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"total_tokens":      promptTokens + completionTokens,
	}
	// A direct call surfaces the upstream cache-hit figure as
	// usage.prompt_tokens_details.cached_tokens, which is what the client's
	// cache-hit display reads. The aggregate usage must keep that shape so the
	// ensemble response shows the same cache-hit rate instead of silently
	// dropping it. Zero is omitted, matching single-model responses where the
	// upstream never reports cache hits.
	if cachedTokens > 0 {
		usage["prompt_tokens_details"] = map[string]any{"cached_tokens": cachedTokens}
	}
	payload := map[string]any{
		"id":      fmt.Sprintf("chatcmpl-ensemble-%d", time.Now().UnixNano()),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]any{
			{
				"index":         0,
				"message":       message,
				"finish_reason": finishReason,
			},
		},
		"usage": usage,
	}

	if exposeMetadata {
		payload["ensemble_metadata"] = map[string]any{
			"members":           ensembleMetadataMembers(stats, includeMemberContent),
			"members_total":     len(stats),
			"members_succeeded": succeeded,
			"aggregated":        aggregated,
			"duration_ms":       elapsed.Milliseconds(),
		}
	}
	return payload
}

// ensembleMetadataMembers prepares the per-member records for the response.
//
// Each stat carries the member's full answer text, which the admin diagnostic
// view renders as "候选模型原始回答". On the production path nothing reads it: a
// caller already has the final answer, so echoing every member's complete answer
// back multiplies the response size by the member count for no one's benefit.
// The text is therefore kept for diagnostics and dropped everywhere else, which
// is what makes expose_metadata cheap enough to leave on.
func ensembleMetadataMembers(stats []ensembleMemberStat, includeContent bool) []ensembleMemberStat {
	if includeContent {
		return stats
	}
	trimmed := make([]ensembleMemberStat, len(stats))
	for i, stat := range stats {
		stat.Content = ""
		trimmed[i] = stat
	}
	return trimmed
}

// ensembleStreamWriter owns the client SSE connection for a stream=true request.
//
// The whole point is that headers and a first chunk go out *before* the fan-out
// starts: members always run non-streaming (aggregation needs complete text), so
// without this the client would receive nothing until every sub-call finished and
// would hit its own idle-read timeout on a slow ensemble. A background ticker
// keeps the connection warm with SSE comments, which every compliant client
// ignores, until the real content is ready.
type ensembleStreamWriter struct {
	ctx     *gin.Context
	id      string
	model   string
	created int64
	start   time.Time

	// traceEnabled mirrors the group's stream_trace setting; exposeModels mirrors
	// expose_metadata, which decides whether a trace line may name the member
	// model or has to fall back to its ordinal. Both settings answer the same
	// question — how much of the group's composition may leave it — so they stay
	// coupled; what used to make expose_metadata expensive (every member's full
	// answer echoed back in the response) is no longer on the production path.
	traceEnabled bool
	exposeModels bool

	mu      sync.Mutex
	done    chan struct{}
	stopped chan struct{}
	once    sync.Once
	// ttftOnce records time-to-first-token exactly once, for whichever real model
	// output reaches the client first — a member's reasoning delta or the answer.
	ttftOnce sync.Once
}

// beginStreamResponse commits the response as SSE and starts the keepalive.
// Every later write for this request must go through the returned writer;
// switching back to c.JSON is no longer possible once headers are flushed.
func (h *EnsembleHandler) beginStreamResponse(c *gin.Context, model string, traceEnabled, exposeModels bool) *ensembleStreamWriter {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	w := &ensembleStreamWriter{
		ctx:          c,
		id:           fmt.Sprintf("chatcmpl-ensemble-%d", time.Now().UnixNano()),
		model:        model,
		created:      time.Now().Unix(),
		start:        time.Now(),
		traceEnabled: traceEnabled,
		exposeModels: exposeModels,
		done:         make(chan struct{}),
		stopped:      make(chan struct{}),
	}

	// An empty role delta is the conventional stream opener and tells the client
	// the request was accepted.
	w.writeChunk(map[string]any{
		"id": w.id, "object": "chat.completion.chunk", "created": w.created, "model": w.model,
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{"role": "assistant", "content": ""}, "finish_reason": nil},
		},
	})

	go w.keepAlive()
	return w
}

func (w *ensembleStreamWriter) keepAlive() {
	ticker := time.NewTicker(ensembleKeepAliveInterval)
	defer ticker.Stop()
	defer close(w.stopped)
	for {
		select {
		case <-w.done:
			return
		case <-w.ctx.Request.Context().Done():
			return
		case <-ticker.C:
			w.writeRaw(": ensemble-keepalive\n\n")
		}
	}
}

// stopKeepAlive halts the ticker and waits for it to observe the stop, so the
// terminal chunks below can never interleave with a heartbeat.
func (w *ensembleStreamWriter) stopKeepAlive() {
	w.once.Do(func() { close(w.done) })
	<-w.stopped
}

func (w *ensembleStreamWriter) writeRaw(payload string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, _ = fmt.Fprint(w.ctx.Writer, payload)
	w.ctx.Writer.Flush()
}

func (w *ensembleStreamWriter) writeChunk(chunk map[string]any) {
	encoded, err := json.Marshal(chunk)
	if err != nil {
		return
	}
	w.writeRaw(fmt.Sprintf("data: %s\n\n", encoded))
}

// traceLine sends one execution-progress line to the client as a
// reasoning_content delta.
//
// reasoning_content is the field the platform already uses for provider thinking
// output, so it needs no client change: a client that renders it shows the
// fan-out as it happens, and one that does not simply ignores an unknown delta
// field. The answer itself is never written here, which keeps the trace out of
// the assistant message the caller stores in its history.
// Every line is stamped with the elapsed time since the fan-out opened. A
// wall-clock timestamp would be noise here; what a caller needs to see is that
// members really do run in parallel and which one is holding the request up.
func (w *ensembleStreamWriter) traceLine(text string) {
	if w == nil || !w.traceEnabled || text == "" {
		return
	}
	stamped := fmt.Sprintf("[%.1fs] %s", time.Since(w.start).Seconds(), text)
	w.writeChunk(map[string]any{
		"id": w.id, "object": "chat.completion.chunk", "created": w.created, "model": w.model,
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{"reasoning_content": stamped}, "finish_reason": nil},
		},
	})
}

// modelReasoning forwards a member model's own reasoning delta to the caller
// verbatim, as it arrives.
//
// This is the model's output, not our scheduling log, so it carries no elapsed
// stamp and is not gated on stream_trace: a direct call to the same model streams
// its thinking, and the ensemble must not silently drop that. Only the aggregator
// is wired to this — running proposers in parallel would interleave several
// models' thinking into one unreadable field.
func (w *ensembleStreamWriter) modelReasoning(text string) {
	if w == nil || text == "" {
		return
	}
	w.markTTFT()
	w.writeChunk(map[string]any{
		"id": w.id, "object": "chat.completion.chunk", "created": w.created, "model": w.model,
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{"reasoning_content": text}, "finish_reason": nil},
		},
	})
}

// markTTFT records the ensemble row's TTFT on the first real model
// output. Members run in isolated sub-contexts and record their own TTFT against
// their own usage rows; without this the ensemble row itself would report null.
// Reasoning tokens are model output, so whichever of reasoning or answer lands
// first is the honest first token. Following the platform convention, TTFT is
// recorded for streaming requests only.
func (w *ensembleStreamWriter) markTTFT() {
	w.ttftOnce.Do(func() {
		service.SetOpsLatencyMs(w.ctx, service.OpsTimeToFirstTokenMsKey, time.Since(w.start).Milliseconds())
	})
}

// traceProgress renders one progress event as a human-readable trace line.
// Returning "" means the event carries nothing worth showing a caller.
func (w *ensembleStreamWriter) traceProgress(event EnsembleProgressEvent) string {
	if w == nil || !w.traceEnabled {
		return ""
	}
	switch event.Type {
	case "started":
		if event.Aggregator != "" {
			return fmt.Sprintf("并行调用 %d 个模型，再由 %s 聚合\n",
				event.ProposersTotal, w.memberLabel(0, service.EnsembleRoleAggregator, event.Aggregator))
		}
		return fmt.Sprintf("并行调用 %d 个模型\n", event.ProposersTotal)
	case "member_started":
		return fmt.Sprintf("→ %s 开始\n", w.memberLabel(event.Index, event.Role, event.Model))
	case "member_finished":
		label := w.memberLabel(event.Index, event.Role, event.Model)
		if event.Status != ensembleStatusOK {
			// The member error is the whole reason a caller looks at the trace, so
			// it is shown even when expose_metadata hides the model id.
			reason := "未知原因"
			if event.Member != nil && event.Member.Error != "" {
				reason = ensembleTruncate(w.redactMemberError(event, event.Member.Error), 200)
			}
			return fmt.Sprintf("✗ %s 失败：%s\n", label, reason)
		}
		if event.Member != nil {
			return fmt.Sprintf("✓ %s 完成（%dms，%d tokens）\n",
				label, event.Member.DurationMs, event.Member.CompletionTokens)
		}
		return fmt.Sprintf("✓ %s 完成\n", label)
	case "proposers_finished":
		return fmt.Sprintf("候选回答 %d/%d 可用\n", event.ProposersSucceeded, event.ProposersTotal)
	case "fallback":
		return fmt.Sprintf("聚合未完成，改用最长候选回答（%s）\n",
			ensembleTruncate(w.redactMemberError(event, event.Error), 200))
	case "error":
		return fmt.Sprintf("✗ 本次请求失败：%s\n", ensembleTruncate(event.Error, 200))
	default:
		return ""
	}
}

// memberLabel names a member in the client-facing trace. expose_metadata governs
// whether real model ids may leave the group, so without it the trace still
// shows the execution shape but refers to members by ordinal.
func (w *ensembleStreamWriter) memberLabel(index int, role, model string) string {
	if w.exposeModels && strings.TrimSpace(model) != "" {
		return model
	}
	if role == service.EnsembleRoleAggregator {
		return "聚合模型"
	}
	return fmt.Sprintf("模型 %d", index+1)
}

// redactMemberError strips the member's real model id out of a sub-call error
// string when expose_metadata is off. The error text is built by the dispatcher
// as "model <name> returned status ...", so masking the label alone would leak
// the id through the message.
func (w *ensembleStreamWriter) redactMemberError(event EnsembleProgressEvent, message string) string {
	if w.exposeModels || strings.TrimSpace(event.Model) == "" {
		return message
	}
	label := w.memberLabel(event.Index, event.Role, event.Model)
	return strings.ReplaceAll(message, event.Model, label)
}

// finish emits the aggregated answer, the terminal chunk and [DONE].
func (w *ensembleStreamWriter) finish(content string, payload map[string]any) {
	w.stopKeepAlive()
	delta := map[string]any{}
	if content != "" {
		delta["content"] = content
	}
	finishReason := "stop"
	if choices, ok := payload["choices"].([]map[string]any); ok && len(choices) > 0 {
		if message, ok := choices[0]["message"].(map[string]any); ok {
			if toolCalls, ok := message["tool_calls"]; ok {
				delta["tool_calls"] = toolCalls
				finishReason = "tool_calls"
			}
		}
		if reason, ok := choices[0]["finish_reason"].(string); ok && reason != "" {
			finishReason = reason
		}
	}

	// Members run in isolated sub-contexts and record their own TTFT against their
	// own usage rows; without this the ensemble row itself would always report a
	// null TTFT. Following the platform convention, TTFT is recorded for streaming
	// requests only. The guard is shared with modelReasoning: when the aggregator
	// streams its thinking first, that delta is the real first token and this
	// chunk must not overwrite it.
	if len(delta) > 0 {
		w.markTTFT()
	}

	w.writeChunk(map[string]any{
		"id": w.id, "object": "chat.completion.chunk", "created": w.created, "model": w.model,
		"choices": []map[string]any{
			{"index": 0, "delta": delta, "finish_reason": nil},
		},
	})

	terminal := map[string]any{
		"id": w.id, "object": "chat.completion.chunk", "created": w.created, "model": w.model,
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{}, "finish_reason": finishReason},
		},
		"usage": payload["usage"],
	}
	if metadata, ok := payload["ensemble_metadata"]; ok {
		terminal["ensemble_metadata"] = metadata
	}
	w.writeChunk(terminal)
	w.writeRaw("data: [DONE]\n\n")
}

// fail reports a mid-stream failure. Headers are already sent, so the status code
// is fixed at 200 and the error can only be delivered in-band.
//
// The frame is a full chat.completion.chunk with the error text as
// delta.content, not a bare {"type":"error"} object. Strict agent SDKs (Codex
// CLI, ZCode) only accept the standard chunk shape; a non-standard frame makes
// them surface "reason=unknown" and hide the actual cause. Delivering the text
// as content keeps the message visible in every client. MarkOpsStreamFailure
// still records the failure for the ops error logger, which otherwise only
// collects rows with status >= 400 and would never see a failure riding on a
// committed 200.
//
// Two properties beyond the chunk shape matter:
//
// The text is prefixed with ensembleFailureNoticePrefix. Because it rides
// delta.content, a client stores it as a normal assistant turn and echoes it back
// forever after; the marker is what lets the fan-out boundary recognize its own
// failure notice next turn and drop it instead of feeding "Only 0 proposers
// succeeded…" to every member as conversation history.
//
// A terminal chunk carrying a non-nil finish_reason is emitted before [DONE],
// mirroring finish(). A stream that ends on finish_reason: null leaves the turn
// open as far as a strict client is concerned, which is not a state any caller
// can cleanly resume from.
func (w *ensembleStreamWriter) fail(errType, message string, intendedStatus int) {
	w.stopKeepAlive()
	service.MarkOpsStreamFailure(w.ctx, errType, "", message, intendedStatus)
	w.writeChunk(map[string]any{
		"id": w.id, "object": "chat.completion.chunk", "created": w.created, "model": w.model,
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{
				"role":    "assistant",
				"content": ensembleFailureNoticePrefix + message,
			}, "finish_reason": nil},
		},
	})
	w.writeChunk(map[string]any{
		"id": w.id, "object": "chat.completion.chunk", "created": w.created, "model": w.model,
		"choices": []map[string]any{
			{"index": 0, "delta": map[string]any{}, "finish_reason": "stop"},
		},
	})
	w.writeRaw("data: [DONE]\n\n")
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
