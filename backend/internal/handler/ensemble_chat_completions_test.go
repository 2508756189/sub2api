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
		GroupID: ptrInt64(7),
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

func ptrInt64(value int64) *int64 { return &value }

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
	require.Equal(t, []bool{false, false, false}, dispatch.streams)
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

func TestEnsembleMetadataContainsCandidateContentAndReportedCost(t *testing.T) {
	response := chatCompletion("candidate answer", 3, 2)
	response["usage"].(map[string]any)["cost"] = 0.0123
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
