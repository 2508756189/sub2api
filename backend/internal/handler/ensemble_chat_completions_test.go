package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type ensembleHandlerRepoStub struct {
	members []service.EnsembleProposer
}

func (s *ensembleHandlerRepoStub) ListByGroup(_ context.Context, groupID int64, includeDisabled bool) ([]service.EnsembleProposer, error) {
	out := make([]service.EnsembleProposer, 0, len(s.members))
	for _, member := range s.members {
		if member.GroupID != groupID || (!includeDisabled && !member.Enabled) {
			continue
		}
		out = append(out, member)
	}
	return out, nil
}

func (s *ensembleHandlerRepoStub) Create(context.Context, *service.EnsembleProposer) error {
	return nil
}
func (s *ensembleHandlerRepoStub) Update(context.Context, *service.EnsembleProposer) error {
	return nil
}
func (s *ensembleHandlerRepoStub) Delete(context.Context, int64) error        { return nil }
func (s *ensembleHandlerRepoStub) DeleteByGroup(context.Context, int64) error { return nil }

type ensembleDispatchStub struct {
	mu             sync.Mutex
	models         []string
	streams        []bool
	clientIDs      []string
	startGate      chan struct{}
	started        int
	responses      map[string]dispatchResponse
	aggregatorCall bool
}

type dispatchResponse struct {
	status int
	body   any
	delay  time.Duration
}

type ensembleCostEstimatorStub struct {
	cost   float64
	source string
}

func (s ensembleCostEstimatorStub) EstimateEnsembleCost(context.Context, int64, string, string, service.UsageTokens) (*service.EnsembleCostEstimate, error) {
	return &service.EnsembleCostEstimate{Cost: s.cost, Source: s.source}, nil
}

func (s *ensembleDispatchStub) dispatch(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var request struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	_ = json.Unmarshal(body, &request)

	s.mu.Lock()
	s.models = append(s.models, request.Model)
	s.streams = append(s.streams, request.Stream)
	clientID, _ := c.Request.Context().Value(ctxkey.ClientRequestID).(string)
	s.clientIDs = append(s.clientIDs, clientID)
	response, ok := s.responses[request.Model]
	if request.Model == "gpt-5.1" {
		s.aggregatorCall = true
	}
	if s.startGate != nil && request.Model != "gpt-5.1" {
		s.started++
		if s.started == 2 {
			close(s.startGate)
		}
	}
	gate := s.startGate
	s.mu.Unlock()

	if gate != nil && request.Model != "gpt-5.1" {
		select {
		case <-gate:
		case <-time.After(time.Second):
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": gin.H{"message": "proposers did not start concurrently"}})
			return
		}
	}
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "unexpected model"}})
		return
	}
	if response.delay > 0 {
		time.Sleep(response.delay)
	}
	c.JSON(response.status, response.body)
}

func newEnsembleHandlerRequest(t *testing.T, members []service.EnsembleProposer, cfg service.EnsembleConfig, dispatch gin.HandlerFunc, body string, estimators ...service.EnsembleCostEstimator) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		ID:      99,
		GroupID: ensembleInt64Ptr(7),
		Group: &service.Group{
			ID:             7,
			Platform:       service.PlatformEnsemble,
			EnsembleConfig: cfg,
		},
	})
	h := NewEnsembleHandler(service.NewEnsembleRuntimeService(&ensembleHandlerRepoStub{members: members}), estimators...)
	h.SetSubCallDispatcher(dispatch)
	h.ChatCompletions(c)
	return recorder
}

func ensembleInt64Ptr(value int64) *int64 { return &value }

func chatCompletion(content string, promptTokens, completionTokens int) map[string]any {
	return map[string]any{
		"choices": []map[string]any{{"message": map[string]any{"role": "assistant", "content": content}}},
		"usage":   map[string]any{"prompt_tokens": promptTokens, "completion_tokens": completionTokens},
	}
}

func TestEnsembleChatCompletionsFansOutInParallelAndAggregates(t *testing.T) {
	dispatch := &ensembleDispatchStub{
		startGate: make(chan struct{}),
		responses: map[string]dispatchResponse{
			"gpt-5":   {status: http.StatusOK, body: chatCompletion("short", 10, 3), delay: 60 * time.Millisecond},
			"gpt-5.1": {status: http.StatusOK, body: chatCompletion("final", 4, 2)},
			"gpt-5.2": {status: http.StatusOK, body: chatCompletion("long proposer", 11, 5), delay: 20 * time.Millisecond},
		},
	}

	started := time.Now()
	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Priority: 20, Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.2", Priority: 10, Enabled: true},
			{ID: 3, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Priority: 1, Enabled: true},
		},
		service.EnsembleConfig{AggregatorEnabled: true, MinProposers: 2, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[{"role":"user","content":"hello"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Less(t, time.Since(started), 300*time.Millisecond)
	require.Contains(t, recorder.Body.String(), `"content":"final"`)
	require.Contains(t, recorder.Body.String(), `"members_succeeded":3`)
	require.ElementsMatch(t, []string{"gpt-5", "gpt-5.2", "gpt-5.1"}, dispatch.models)
	require.Equal(t, []bool{true, true, true}, dispatch.streams)
	require.Len(t, dispatch.clientIDs, 3)
	require.Len(t, map[string]struct{}{
		dispatch.clientIDs[0]: {}, dispatch.clientIDs[1]: {}, dispatch.clientIDs[2]: {},
	}, 3)
}

func TestEnsembleChatCompletionsUsesLongestProposalWhenAggregatorDisabled(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletion("short", 1, 1)},
		"gpt-5.1": {status: http.StatusOK, body: chatCompletion("the longest answer", 2, 2)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{MinProposers: 2, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"content":"the longest answer"`)
	require.Contains(t, recorder.Body.String(), `"aggregated":false`)
}

func TestEnsembleChatCompletionsUsesConfiguredPlatformForOpaqueModel(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"qwen3.7-max": {status: http.StatusOK, body: chatCompletion("qwen answer", 2, 3)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{
			ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer,
			Model: "qwen3.7-max", Platform: service.PlatformOpenAI, Enabled: true,
		}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"content":"qwen answer"`)
	require.Contains(t, recorder.Body.String(), `"platform":"openai"`)
}

func TestEnsembleChatCompletionsFallsBackWhenAggregatorFails(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletion("the longest proposer", 1, 4)},
		"gpt-5.1": {status: http.StatusBadGateway, body: gin.H{"error": gin.H{"message": "aggregator down"}}},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{AggregatorEnabled: true, MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"content":"the longest proposer"`)
	require.Contains(t, recorder.Body.String(), `"aggregated":false`)
	require.Contains(t, recorder.Body.String(), `"status":"failed"`)
}

func TestEnsembleChatCompletionsReturnsBadGatewayBelowMinimum(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletion("ok", 1, 1)},
		"gpt-5.1": {status: http.StatusBadGateway, body: gin.H{"error": gin.H{"message": "down"}}},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{MinProposers: 2},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[],"stream":false}`,
	)

	require.Equal(t, http.StatusBadGateway, recorder.Code)
	require.Contains(t, recorder.Body.String(), "minimum 2 required")
}

func TestEnsembleChatCompletionsRetriesEmptyMemberCompletion(t *testing.T) {
	var attempts int
	dispatch := func(c *gin.Context) {
		attempts++
		if attempts == 1 {
			c.JSON(http.StatusOK, chatCompletion("", 1, 0))
			return
		}
		c.JSON(http.StatusOK, chatCompletion("recovered", 1, 2))
	}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1},
		dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hello"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 2, attempts)
	require.Contains(t, recorder.Body.String(), `"content":"recovered"`)
}

func TestPrepareEnsembleSubCallBodyKeepsStreamingForTTFT(t *testing.T) {
	body, err := prepareEnsembleSubCallBody(
		[]byte(`{"model":"ensemble","messages":[{"role":"user","content":"hello"}],"stream":false,"stream_options":{"include_usage":true}}`),
		"gpt-5",
		0,
	)
	require.NoError(t, err)
	require.Equal(t, "gpt-5", gjson.GetBytes(body, "model").String())
	require.True(t, gjson.GetBytes(body, "stream").Bool())
	require.True(t, gjson.GetBytes(body, "stream_options.include_usage").Bool())
}

// A sub-call replaces the request body, so the body length no longer matches the
// client's Content-Length. Request.Clone copies the header map verbatim, so the
// stale header must be dropped: any consumer that trusts the header over
// Request.ContentLength would read a truncated body and reject the JSON.
func TestEnsembleSubCallDropsStaleContentLengthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var (
		mu           sync.Mutex
		headerCL     string
		structCL     int64
		observedBody int
	)
	dispatch := func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		mu.Lock()
		headerCL = c.Request.Header.Get("Content-Length")
		structCL = c.Request.ContentLength
		observedBody = len(body)
		mu.Unlock()
		c.JSON(http.StatusOK, chatCompletion("ok", 1, 1))
	}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1},
		dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"你好，请介绍一下自己"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	mu.Lock()
	defer mu.Unlock()
	require.Empty(t, headerCL, "stale Content-Length header must be removed from the sub-request")
	require.Equal(t, int64(observedBody), structCL, "Request.ContentLength must match the rewritten body")
}

// gjson scans leniently and will match "choices.0" inside raw SSE text. The
// non-stream fast path must therefore validate the whole document first,
// otherwise a streamed sub-response is returned unparsed, content extraction
// yields "", and the member is misreported as an empty completion.
func TestNormalizeEnsembleChatPayloadParsesSSEDespiteLenientMatch(t *testing.T) {
	sse := "data: {\"choices\":[{\"delta\":{\"content\":\"hel\"},\"finish_reason\":null}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":5,\"completion_tokens\":2}}\n\n" +
		"data: [DONE]\n\n"

	require.True(t, gjson.Get(sse, "choices.0").Exists(), "precondition: gjson matches inside SSE text")

	normalized := normalizeEnsembleChatPayload([]byte(sse))
	require.Equal(t, "hello", extractEnsembleContent(normalized))
	require.Equal(t, "stop", gjson.GetBytes(normalized, "choices.0.finish_reason").String())
	require.Equal(t, 5, int(gjson.GetBytes(normalized, "usage.prompt_tokens").Int()))
	require.Equal(t, 2, int(gjson.GetBytes(normalized, "usage.completion_tokens").Int()))
}

// A single SSE data line can exceed bufio.Scanner's default 64KB token limit
// when an upstream buffers a long answer into few chunks. The parser must not
// silently truncate the answer in that case.
func TestNormalizeEnsembleChatPayloadHandlesOversizedSSELine(t *testing.T) {
	long := strings.Repeat("x", 200_000)
	chunk, err := json.Marshal(map[string]any{
		"choices": []map[string]any{{"delta": map[string]any{"content": long}, "finish_reason": "stop"}},
	})
	require.NoError(t, err)

	normalized := normalizeEnsembleChatPayload([]byte("data: " + string(chunk) + "\n\ndata: [DONE]\n\n"))
	require.Equal(t, long, extractEnsembleContent(normalized))
}

// A non-stream JSON response must pass through untouched.
func TestNormalizeEnsembleChatPayloadPassesThroughPlainJSON(t *testing.T) {
	payload, err := json.Marshal(chatCompletion("plain", 1, 2))
	require.NoError(t, err)
	require.Equal(t, "plain", extractEnsembleContent(normalizeEnsembleChatPayload(payload)))
}

func TestEnsembleChatCompletionsParsesInternalSSEAndPreservesHistory(t *testing.T) {
	var bodies [][]byte
	var mu sync.Mutex
	dispatch := func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		mu.Lock()
		bodies = append(bodies, append([]byte(nil), body...))
		mu.Unlock()
		model := string(gjson.GetBytes(body, "model").String())
		if model == "gpt-5.1" {
			c.Data(http.StatusOK, "application/json", []byte("data: {\"choices\":[{\"delta\":{\"content\":\"final\"},\"finish_reason\":null}]}\n\ndata: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":2,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n"))
			return
		}
		c.Data(http.StatusOK, "text/event-stream", []byte("data: {\"choices\":[{\"delta\":{\"content\":\"candidate\"},\"finish_reason\":null}]}\n\ndata: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":2}}\n\ndata: [DONE]\n\n"))
	}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{AggregatorEnabled: true, MinProposers: 1, ExposeMetadata: true},
		dispatch,
		`{"model":"ensemble","messages":[{"role":"system","content":"keep this"},{"role":"user","content":"first"},{"role":"assistant","content":"prior answer"},{"role":"user","content":"follow up"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"content":"final"`)
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, bodies, 2)

	// Multi-turn is client-driven: the client resends its whole history every
	// turn, so the only server-side requirement is that the history reaches each
	// member unmodified. The proposer therefore sees exactly the client's
	// messages; the aggregator sees those plus one appended candidates message.
	var proposerBody, aggregatorBody []byte
	for _, body := range bodies {
		if gjson.GetBytes(body, "model").String() == "gpt-5.1" {
			aggregatorBody = body
			continue
		}
		proposerBody = body
	}
	require.NotNil(t, proposerBody)
	require.NotNil(t, aggregatorBody)

	require.Equal(t, 4, int(gjson.GetBytes(proposerBody, "messages.#").Int()))
	require.Equal(t, "keep this", gjson.GetBytes(proposerBody, "messages.0.content").String())
	require.Equal(t, "prior answer", gjson.GetBytes(proposerBody, "messages.2.content").String())
	require.Equal(t, "follow up", gjson.GetBytes(proposerBody, "messages.3.content").String())

	require.Equal(t, 5, int(gjson.GetBytes(aggregatorBody, "messages.#").Int()))
	require.Equal(t, "prior answer", gjson.GetBytes(aggregatorBody, "messages.2.content").String())
	require.Contains(t, gjson.GetBytes(aggregatorBody, "messages.4.content").String(), "candidate")
}

func TestEnsembleChatCompletionsPreservesToolCallsForContinuation(t *testing.T) {
	dispatch := func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{
					"role": "assistant",
					"tool_calls": []map[string]any{{
						"id": "call_1", "type": "function",
						"function": map[string]any{"name": "inspect", "arguments": `{"path":"."}`},
					}},
				},
				"finish_reason": "tool_calls",
			}},
			"usage": map[string]any{"prompt_tokens": 2, "completion_tokens": 1},
		})
	}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1},
		dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"inspect"}],"tools":[{"type":"function","function":{"name":"inspect"}}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"tool_calls"`)
	require.Contains(t, recorder.Body.String(), `"finish_reason":"tool_calls"`)
}

func TestEnsembleCompactCallsOnlyConfiguredAggregator(t *testing.T) {
	var calls []string
	dispatch := func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		calls = append(calls, gjson.GetBytes(body, "model").String())
		c.JSON(http.StatusOK, map[string]any{
			"id":     "resp_compact_1",
			"object": "response",
			"model":  gjson.GetBytes(body, "model").String(),
			"output": []any{},
		})
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses/compact", strings.NewReader(`{"model":"ensemble","input":[{"role":"user","content":"compress this"}],"stream":false}`))
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		ID: 99, GroupID: ensembleInt64Ptr(7), Group: &service.Group{
			ID: 7, Platform: service.PlatformEnsemble,
			EnsembleConfig: service.EnsembleConfig{AggregatorEnabled: true, MinProposers: 1},
		},
	})
	h := NewEnsembleHandler(service.NewEnsembleRuntimeService(&ensembleHandlerRepoStub{members: []service.EnsembleProposer{
		{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
		{ID: 2, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Platform: service.PlatformOpenAI, Enabled: true},
	}}))
	h.SetResponsesSubCallDispatcher(dispatch)
	h.Compact(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, []string{"gpt-5.1"}, calls)
}

func TestEnsembleChatCompletionsStreamsFinalAnswerAndMetadata(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("streamed", 3, 2)},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[],"stream":true,"stream_options":{"include_usage":true}}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Header().Get("Content-Type"), "text/event-stream")
	require.Contains(t, recorder.Body.String(), `"content":"streamed"`)
	require.Contains(t, recorder.Body.String(), `"ensemble_metadata"`)
	require.Contains(t, recorder.Body.String(), "[DONE]")
}

// The ops log reads TTFT off the outer request context. Member sub-calls write
// it to their own isolated context, so without an explicit record here the
// ensemble row always shows a null first-token time even though the client did
// receive a first token.
func TestEnsembleStreamingRecordsOuterTimeToFirstToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("streamed", 3, 2), delay: 15 * time.Millisecond},
	}}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		ID: 99, GroupID: ensembleInt64Ptr(7),
		Group: &service.Group{ID: 7, Platform: service.PlatformEnsemble,
			EnsembleConfig: service.EnsembleConfig{MinProposers: 1}},
	})
	h := NewEnsembleHandler(service.NewEnsembleRuntimeService(&ensembleHandlerRepoStub{members: []service.EnsembleProposer{
		{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
	}}))
	h.SetSubCallDispatcher(dispatch.dispatch)

	h.ChatCompletions(c)

	raw, ok := c.Get(service.OpsTimeToFirstTokenMsKey)
	require.True(t, ok, "streaming ensemble request must record time-to-first-token")
	ttft, ok := raw.(int64)
	require.True(t, ok, "TTFT must be int64 for the ops logger")
	require.Greater(t, ttft, int64(0))
}

// A non-stream request has no first token by definition, and the ops
// aggregation weights TTFT by streaming sample count. Recording it here would
// pollute the percentile math.
func TestEnsembleNonStreamingDoesNotRecordTimeToFirstToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("plain", 3, 2)},
	}}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":false}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		ID: 99, GroupID: ensembleInt64Ptr(7),
		Group: &service.Group{ID: 7, Platform: service.PlatformEnsemble,
			EnsembleConfig: service.EnsembleConfig{MinProposers: 1}},
	})
	h := NewEnsembleHandler(service.NewEnsembleRuntimeService(&ensembleHandlerRepoStub{members: []service.EnsembleProposer{
		{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
	}}))
	h.SetSubCallDispatcher(dispatch.dispatch)

	h.ChatCompletions(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	_, ok := c.Get(service.OpsTimeToFirstTokenMsKey)
	require.False(t, ok, "non-stream ensemble request must not record TTFT")
}

func TestEnsembleMetadataContainsCandidateContentAndReportedCost(t *testing.T) {
	response := chatCompletion("candidate answer", 3, 2)
	usage, ok := response["usage"].(map[string]any)
	require.True(t, ok)
	usage["cost"] = 0.0123
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: response},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"content":"candidate answer"`)
	require.Contains(t, recorder.Body.String(), `"cost":0.0123`)
}

func TestEnsembleMetadataMarksUnavailableCostAsNull(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("candidate answer", 3, 4)},
	}}
	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"cost":null`)
}

func TestEnsembleMetadataEstimatesCostFromConfiguredPricing(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"qwen3.7-max": {status: http.StatusOK, body: chatCompletion("candidate answer", 3, 4)},
	}}
	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{
			ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer,
			Model: "qwen3.7-max", Platform: service.PlatformOpenAI, Enabled: true,
		}},
		service.EnsembleConfig{MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[],"stream":false}`,
		ensembleCostEstimatorStub{cost: 0.0042, source: service.PricingSourceChannel},
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"cost":0.0042`)
	require.Contains(t, recorder.Body.String(), `"cost_source":"channel"`)
}

func TestEnsembleChatCompletionsRejectsMalformedRequest(t *testing.T) {
	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1},
		func(c *gin.Context) { t.Fatal("dispatcher must not be called") },
		`{"model":`,
	)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestEnsembleChatCompletionsRequiresRequestedModel(t *testing.T) {
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("unexpected", 1, 1)},
	}}
	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1},
		dispatch.dispatch,
		`{"messages":[],"stream":false}`,
	)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestEnsembleSubCallReturnsWhenDispatcherIgnoresContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	parent := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(parent)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"ensemble-public","messages":[]}`))

	started := make(chan struct{})
	release := make(chan struct{})
	h := NewEnsembleHandler(nil)
	h.SetSubCallDispatcher(func(*gin.Context) {
		close(started)
		<-release
	})

	resultCh := make(chan ensembleSubResult, 1)
	go func() {
		resultCh <- h.runSubCall(c, 0, "gpt-5", service.PlatformOpenAI, service.EnsembleRoleProposer,
			[]byte(`{"model":"ensemble-public","messages":[]}`), 0, 20*time.Millisecond)
	}()
	<-started

	select {
	case result := <-resultCh:
		require.Error(t, result.err)
		require.Contains(t, result.err.Error(), "timed out")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("sub-call waited for a dispatcher that ignored context cancellation")
	}
	close(release)
}

func TestEnsembleAggregatorBodyRejectsOversizedCandidateContext(t *testing.T) {
	_, err := buildEnsembleAggregatorBody(
		[]byte(`{"model":"ensemble-public","messages":[]}`),
		[]ensembleProposal{{Model: "gpt-5", Content: strings.Repeat("x", maxEnsembleAggregatorBodyBytes)}},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "aggregator request exceeds")
}

func TestEnsembleTestStreamReportsMembersAndFinalResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletion("candidate", 7, 3)},
		"gpt-5.1": {status: http.StatusOK, body: chatCompletion("final answer", 11, 4)},
	}}
	members := []service.EnsembleProposer{
		{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Platform: service.PlatformOpenAI, Enabled: true},
		{ID: 2, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Platform: service.PlatformOpenAI, Enabled: true},
	}
	h := NewEnsembleHandler(service.NewEnsembleRuntimeService(&ensembleHandlerRepoStub{members: members}))
	h.SetSubCallDispatcher(dispatch.dispatch)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/ensemble/test", strings.NewReader(
		`{"model":"ensemble-public","messages":[{"role":"user","content":"hello"}],"stream":false}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		ID:      99,
		GroupID: ensembleInt64Ptr(7),
		Group: &service.Group{
			ID:       7,
			Platform: service.PlatformEnsemble,
			EnsembleConfig: service.EnsembleConfig{
				AggregatorEnabled: true,
				MinProposers:      1,
				ExposeMetadata:    true,
			},
		},
	})

	h.TestStream(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "text/event-stream", recorder.Header().Get("Content-Type"))
	stream := recorder.Body.String()
	require.Contains(t, stream, "event:started")
	require.Contains(t, stream, "event:member_started")
	require.Contains(t, stream, "event:member_finished")
	require.Contains(t, stream, "event:completed")
	require.Contains(t, stream, `"model":"gpt-5"`)
	require.Contains(t, stream, `"model":"gpt-5.1"`)
	require.Contains(t, stream, `"content":"final answer"`)
}

// A diagnostic request with stream:true must still produce a parseable terminal
// event. The inner gateway call is forced non-streaming so the recorder holds
// JSON; if it held SSE text the terminal event would fail to marshal and the
// admin UI would see a truncated stream instead of the real result.
func TestEnsembleTestStreamForcesInnerRequestNonStreaming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5": {status: http.StatusOK, body: chatCompletion("candidate", 7, 3)},
	}}
	members := []service.EnsembleProposer{
		{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Platform: service.PlatformOpenAI, Enabled: true},
	}
	h := NewEnsembleHandler(service.NewEnsembleRuntimeService(&ensembleHandlerRepoStub{members: members}))
	h.SetSubCallDispatcher(dispatch.dispatch)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/ensemble/test", strings.NewReader(
		`{"model":"ensemble","messages":[{"role":"user","content":"hello"}],"stream":true,"stream_options":{"include_usage":true}}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		ID: 99, GroupID: ensembleInt64Ptr(7),
		Group: &service.Group{ID: 7, Platform: service.PlatformEnsemble, EnsembleConfig: service.EnsembleConfig{MinProposers: 1}},
	})

	h.TestStream(c)

	// Member sub-calls stream internally so the normal gateway records TTFT.
	// That is independent of the client-facing diagnostic stream.
	require.Equal(t, []bool{true}, dispatch.streams)

	stream := recorder.Body.String()
	require.Contains(t, stream, "event:completed")
	require.NotContains(t, stream, "event:error")

	terminal := ensembleTerminalEvent(t, stream)
	require.Equal(t, http.StatusOK, int(gjson.Get(terminal, "status_code").Int()))
	response := gjson.Get(terminal, "response")
	require.True(t, response.Exists(), "terminal event must carry the inner response")
	require.True(t, gjson.Valid(response.Raw), "inner response must be JSON, not SSE text")
	require.Equal(t, "candidate", gjson.Get(response.Raw, "choices.0.message.content").String())
}

// ensembleTerminalEvent returns the data payload of the last SSE event, which is
// the terminal completed/error event emitted by TestStream.
func ensembleTerminalEvent(t *testing.T, stream string) string {
	t.Helper()
	last := ""
	for _, line := range strings.Split(stream, "\n") {
		line = strings.TrimSpace(line)
		if after, found := strings.CutPrefix(line, "data:"); found {
			if payload := strings.TrimSpace(after); payload != "" {
				last = payload
			}
		}
	}
	require.NotEmpty(t, last, "no SSE data payload found in diagnostic stream")
	return last
}
