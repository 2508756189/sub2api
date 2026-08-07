package handler

// Execution-trace tests for the ensemble streaming path.
//
// The fan-out is silent for as long as the slowest member takes. The trace
// (progress events rendered as reasoning_content deltas) is what lets a caller
// watch the run and see why a request failed, instead of staring at a blank
// stream and then getting an unparseable error frame. The helpers used here
// (ensembleDispatchStub, chatCompletion, newEnsembleHandlerRequest) live in
// ensemble_chat_completions_test.go.

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func ensembleBoolPtr(value bool) *bool { return &value }

// ensembleReasoningTrace collects the reasoning_content deltas of an SSE body in
// arrival order — exactly what a trace-rendering client shows.
func ensembleReasoningTrace(t *testing.T, stream string) []string {
	t.Helper()
	lines := make([]string, 0, 8)
	for _, raw := range strings.Split(stream, "\n") {
		after, found := strings.CutPrefix(strings.TrimSpace(raw), "data:")
		if !found {
			continue
		}
		payload := strings.TrimSpace(after)
		if payload == "" || payload == "[DONE]" || !gjson.Valid(payload) {
			continue
		}
		if reasoning := gjson.Get(payload, "choices.0.delta.reasoning_content"); reasoning.Exists() {
			lines = append(lines, reasoning.String())
		}
	}
	return lines
}

// Every trace line carries an elapsed-time stamp, which is what turns the
// events into a readable execution log.
func TestEnsembleStreamTraceShowsExecutionStepsToClient(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletion("candidate", 3, 2)},
		"gpt-5.1": {status: http.StatusOK, body: chatCompletion("final", 4, 3)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{AggregatorEnabled: true, MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	trace := strings.Join(ensembleReasoningTrace(t, recorder.Body.String()), "")
	require.NotEmpty(t, trace, "stream_trace defaults to on, so a streaming client must receive the trace")

	// Order matters: it is what makes the trace readable as an execution log.
	fanOut := strings.Index(trace, "并行调用 1 个模型")
	memberStart := strings.Index(trace, "→ gpt-5 开始")
	memberDone := strings.Index(trace, "✓ gpt-5 完成")
	proposersDone := strings.Index(trace, "候选回答 1/1 可用")
	aggregatorStart := strings.Index(trace, "→ gpt-5.1 开始")
	require.Greater(t, fanOut, -1, "trace must announce the fan-out")
	require.Greater(t, memberStart, fanOut)
	require.Greater(t, memberDone, memberStart)
	require.Greater(t, proposersDone, memberDone)
	require.Greater(t, aggregatorStart, proposersDone, "aggregation must be reported after the candidates")
	require.Contains(t, trace, "再由 gpt-5.1 聚合")

	// The trace rides the reasoning channel; the answer stays in content so the
	// caller's stored history is unaffected.
	body := recorder.Body.String()
	require.Contains(t, body, `"content":"final"`)
	require.NotContains(t, trace, "final")
}

// The trace has to be readable without leaking member model ids, because
// expose_metadata is the group's control over what identities leave the group.
func TestEnsembleStreamTraceMasksModelNamesWithoutExposeMetadata(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletion("candidate", 3, 2)},
		"gpt-5.1": {status: http.StatusBadGateway, body: gin.H{"error": gin.H{"message": "upstream exploded"}}},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: false},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	trace := strings.Join(ensembleReasoningTrace(t, recorder.Body.String()), "")
	require.NotEmpty(t, trace)
	require.NotContains(t, trace, "gpt-5", "expose_metadata=false must keep member model ids out of the trace")
	require.Contains(t, trace, "模型 1")
	require.Contains(t, trace, "模型 2")
	// A masked label must still not hide why a member failed.
	require.Contains(t, trace, "失败")
	require.Contains(t, trace, "upstream exploded")
}

func TestEnsembleStreamTraceCanBeDisabledPerGroup(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("quiet", 3, 2)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true, StreamTrace: ensembleBoolPtr(false)},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Empty(t, ensembleReasoningTrace(t, recorder.Body.String()),
		"stream_trace=false must produce no reasoning_content deltas")
	require.Contains(t, recorder.Body.String(), `"content":"quiet"`)
}

// A non-stream response has no channel for progress, so the trace must not leak
// into the JSON body.
func TestEnsembleNonStreamingResponseCarriesNoTrace(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("plain", 3, 2)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "reasoning_content")
	require.NotContains(t, recorder.Body.String(), "并行调用")
}

// A fan-out failure is discovered after the SSE 200 is already committed, so it
// can only be reported in-band. This asserts the two properties that make that
// report usable: a client can key on it, and ops can see it.
func TestEnsembleStreamFailureIsParseableAndRecordedForOps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusBadGateway, body: gin.H{"error": gin.H{"message": "upstream 503"}}},
		"gpt-5.1": {status: http.StatusBadGateway, body: gin.H{"error": gin.H{"message": "upstream 503"}}},
	}}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
		ID: 99, GroupID: ensembleInt64Ptr(7),
		Group: &service.Group{ID: 7, Platform: service.PlatformEnsemble,
			EnsembleConfig: service.EnsembleConfig{MinProposers: 2, ExposeMetadata: true}},
	})
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7, Concurrency: 2})
	h := NewEnsembleHandler(service.NewEnsembleRuntimeService(&ensembleHandlerRepoStub{members: []service.EnsembleProposer{
		{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
		{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.1", Enabled: true},
	}}))
	h.SetSubCallDispatcher(dispatch.dispatch)

	h.ChatCompletions(c)

	// The SSE headers were committed before the fan-out, so the wire status stays 200.
	require.Equal(t, http.StatusOK, recorder.Code)

	var errorChunk string
	for _, raw := range strings.Split(recorder.Body.String(), "\n") {
		after, found := strings.CutPrefix(strings.TrimSpace(raw), "data:")
		if !found {
			continue
		}
		payload := strings.TrimSpace(after)
		if gjson.Valid(payload) && gjson.Get(payload, "choices.0.delta.content").Exists() {
			errorChunk = payload
		}
	}
	require.NotEmpty(t, errorChunk, "a failed fan-out must emit an in-band error chunk")
	// Strict agent SDKs (Codex CLI, ZCode) only accept the standard
	// chat.completion.chunk shape; a bare {"type":"error"} frame makes them
	// report an unknown failure reason instead of the real cause.
	require.Equal(t, "chat.completion.chunk", gjson.Get(errorChunk, "object").String())
	require.Contains(t, gjson.Get(errorChunk, "choices.0.delta.content").String(), "minimum 2 required")
	require.Contains(t, recorder.Body.String(), "[DONE]")

	// ops_error_logger only collects rows with status >= 400, so a failure riding
	// on a committed 200 is invisible unless it is marked explicitly.
	streamErr, ok := service.GetOpsStreamError(c)
	require.True(t, ok, "in-band ensemble failure must be recorded for the ops error log")
	require.Equal(t, http.StatusBadGateway, streamErr.IntendedStatus)
	require.True(t, streamErr.CountTowardsSLA, "a failed request must count towards the error rate")
	require.Contains(t, streamErr.Message, "minimum 2 required")

	// The human-readable cause belongs in the same place the caller watched the run.
	trace := strings.Join(ensembleReasoningTrace(t, recorder.Body.String()), "")
	require.Contains(t, trace, "本次请求失败")
}

// The sub-call billing identity must stay on the sub-request even when a trace
// sink is attached; the trace must not perturb the per-member usage rows.
func TestEnsembleStreamTraceDoesNotReplaceSubCallBillingIdentity(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("ok", 3, 2)},
	}}

	newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	)

	require.Len(t, dispatch.clientIDs, 1)
	require.Contains(t, dispatch.clientIDs[0], "ensemble-", "sub-call billing identity must keep its own prefix")
	require.NotEqual(t, "", dispatch.clientIDs[0])
}

// bodyCapturingDispatch wraps a dispatch stub and records every sub-call body,
// so tests can assert on what the members actually receive.
type bodyCapturingDispatch struct {
	inner  gin.HandlerFunc
	mu     sync.Mutex
	bodies []string
}

func (d *bodyCapturingDispatch) dispatch(c *gin.Context) {
	raw, _ := io.ReadAll(c.Request.Body)
	d.mu.Lock()
	d.bodies = append(d.bodies, string(raw))
	d.mu.Unlock()
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	d.inner(c)
}

func TestStripEnsembleReasoningFieldsRemovesEchoedTrace(t *testing.T) {
	body := `{"model":"ensemble","messages":[` +
		`{"role":"system","content":"sys"},` +
		`{"role":"user","content":"hi"},` +
		`{"role":"assistant","content":"answer","reasoning_content":"[0.0s] 并行调用 2 个模型\n","thinking":"previous thinking"},` +
		`{"role":"user","content":"next"}]}`

	out, err := stripEnsembleReasoningFields([]byte(body))
	require.NoError(t, err)
	require.NotContains(t, string(out), "reasoning_content", "echoed trace must never reach members")
	require.NotContains(t, string(out), "thinking", "echoed thinking must never reach members")
	require.Contains(t, string(out), `"content":"sys"`)
	require.Contains(t, string(out), `"content":"answer"`)
	require.Contains(t, string(out), `"content":"next"`)
}

// A reasoning-capable client stores the trace delta on the assistant message and
// echoes it back next turn. That echo must be stripped before fan-out: feeding
// it to members would pollute their context with our own trace lines.
func TestEnsembleDropsClientEchoedReasoningFromSubCalls(t *testing.T) {
	inner := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("fresh", 3, 2)},
	}}
	capture := &bodyCapturingDispatch{inner: inner.dispatch}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		capture.dispatch,
		`{"model":"ensemble","messages":[`+
			`{"role":"user","content":"first"},`+
			`{"role":"assistant","content":"ok","reasoning_content":"[0.0s] 并行调用 1 个模型\n"},`+
			`{"role":"user","content":"second"}],"stream":true}`,
	)
	require.Equal(t, http.StatusOK, recorder.Code)

	require.Len(t, capture.bodies, 1, "one proposer sub-call")
	require.NotContains(t, capture.bodies[0], "reasoning_content",
		"member body must not carry the client's echoed trace")
	require.Contains(t, capture.bodies[0], `"content":"ok"`,
		"real assistant content survives")
}

// A direct call surfaces usage.prompt_tokens_details.cached_tokens, which is
// what the client's cache-hit display reads. The aggregate usage must keep that
// shape instead of silently dropping the members' cache hits.
func TestEnsembleAggregateUsageSurfacesCachedTokens(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletionWithUsage("cached", 10, 5, 7)},
		"gpt-5.1": {status: http.StatusOK, body: chatCompletionWithUsage("plain", 20, 8, 0)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	usage := gjson.GetBytes(recorder.Body.Bytes(), "usage")
	require.Equal(t, int64(30), usage.Get("prompt_tokens").Int())
	require.Equal(t, int64(13), usage.Get("completion_tokens").Int())
	require.Equal(t, int64(7), usage.Get("prompt_tokens_details.cached_tokens").Int(),
		"aggregate usage must carry the summed cache hits for the client's cache-hit display")
}

// When no member reports cache hits the details object is omitted, matching a
// direct single-model response whose upstream never reports cache usage.
func TestEnsembleAggregateUsageOmitsCachedTokensWhenNone(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("plain", 10, 5)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.False(t, gjson.GetBytes(recorder.Body.Bytes(), "usage.prompt_tokens_details").Exists(),
		"no cache hits -> no details object, like a direct call")
}

// chatCompletionWithUsage builds a member response whose upstream usage carries
// a prompt-token cache hit, exercising the OpenAI spelling of the field.
func chatCompletionWithUsage(content string, promptTokens, completionTokens, cachedTokens int) map[string]any {
	payload := chatCompletion(content, promptTokens, completionTokens)
	usage, ok := payload["usage"].(map[string]any)
	if !ok {
		return payload
	}
	if cachedTokens > 0 {
		usage["prompt_tokens_details"] = map[string]any{"cached_tokens": cachedTokens}
	}
	return payload
}

// poolCapturingDispatch records the account-pool group set on each sub-call
// context, proving the scheduler receives the union of the caller group and the
// configured source groups instead of a hand-bound single group.
type poolCapturingDispatch struct {
	inner gin.HandlerFunc
	mu    sync.Mutex
	pools [][]int64
}

func (d *poolCapturingDispatch) dispatch(c *gin.Context) {
	d.mu.Lock()
	d.pools = append(d.pools, service.AccountPoolGroupIDsFromContext(c.Request.Context()))
	d.mu.Unlock()
	d.inner(c)
}

// The member sub-call must draw its account pool from the caller's group plus
// the configured source groups. That is what makes ensemble scheduling reuse
// the normal load-balanced, failover-capable account pool of the source group
// instead of whatever accounts happen to be hand-bound to the ensemble group.
func TestEnsembleSubCallUsesSourceGroupAccountPool(t *testing.T) {
	inner := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("pooled", 3, 2)},
	}}
	capture := &poolCapturingDispatch{inner: inner.dispatch}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, SourceGroupIDs: []int64{42, 43}, ExposeMetadata: true},
		capture.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	)
	require.Equal(t, http.StatusOK, recorder.Code)

	require.Len(t, capture.pools, 1, "one proposer sub-call")
	require.ElementsMatch(t, []int64{7, 42, 43}, capture.pools[0],
		"sub-call pool must be the caller group plus the source groups")
}

// Without source groups the pool is still the caller's own group, preserving
// the pre-existing behaviour for groups configured before the field existed.
func TestEnsembleSubCallPoolFallsBackToCallerGroup(t *testing.T) {
	inner := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("own", 3, 2)},
	}}
	capture := &poolCapturingDispatch{inner: inner.dispatch}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		capture.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	)
	require.Equal(t, http.StatusOK, recorder.Code)

	require.Len(t, capture.pools, 1)
	require.Equal(t, []int64{7}, capture.pools[0],
		"without source groups the pool must stay the caller's own group")
}
func TestEnsemblePartialFailureStillReturnsWhenMinimumMet(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletion("first", 2, 1)},
		"gpt-5.1": {status: http.StatusBadGateway, body: gin.H{"error": gin.H{"message": "down"}}},
		"gpt-5.2": {status: http.StatusOK, body: chatCompletion("second", 3, 1)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.1", Enabled: true},
			{ID: 3, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.2", Enabled: true},
		},
		service.EnsembleConfig{MinProposers: 2, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "minimum")
	// The two successful answers must be present.
	require.Contains(t, recorder.Body.String(), "first")
	require.Contains(t, recorder.Body.String(), "second")
}
