package handler

// Tests for forwarding a member model's own reasoning to the caller.
//
// A direct single-model call streams delta.reasoning_content straight through.
// The ensemble buffers each member's SSE body so it can normalize a complete
// answer before aggregating, and that buffering used to drop reasoning entirely:
// the caller saw only our scheduling trace and none of the model's thinking,
// which made the ensemble a capability regression against a direct call.
//
// Only the aggregator is forwarded. It writes the answer the caller receives,
// whereas proposers run in parallel and interleaving several models' thinking
// into one field produces an unreadable stream.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// ensembleSSEDispatch answers sub-calls with a real SSE body, which is what the
// members actually receive from the gateway. The plain-JSON stub used elsewhere
// cannot exercise the tee because there are no frames to scan.
type ensembleSSEDispatch struct {
	mu sync.Mutex
	// frames maps a member model to the data: payloads it streams, in order.
	frames map[string][]string
	// splitEvery > 0 writes each frame in chunks of that size, reproducing a
	// provider that splits one SSE line across several Write calls.
	splitEvery int
}

func (d *ensembleSSEDispatch) dispatch(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	model := gjson.GetBytes(body, "model").String()

	d.mu.Lock()
	frames, ok := d.frames[model]
	d.mu.Unlock()
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "unexpected model " + model}})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.WriteHeader(http.StatusOK)
	for _, frame := range frames {
		d.writeLine(c, "data: "+frame+"\n\n")
	}
	d.writeLine(c, "data: [DONE]\n\n")
}

func (d *ensembleSSEDispatch) writeLine(c *gin.Context, line string) {
	if d.splitEvery <= 0 {
		_, _ = c.Writer.WriteString(line)
		c.Writer.Flush()
		return
	}
	for start := 0; start < len(line); start += d.splitEvery {
		end := start + d.splitEvery
		if end > len(line) {
			end = len(line)
		}
		_, _ = c.Writer.WriteString(line[start:end])
		c.Writer.Flush()
	}
}

func reasoningFrame(text string) string {
	return `{"id":"x","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"reasoning_content":"` + text + `"}}]}`
}

func contentFrame(text string) string {
	return `{"id":"x","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"` + text + `"}}]}`
}

func finalFrame() string {
	return `{"id":"x","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":4,"completion_tokens":2,"total_tokens":6}}`
}

// The aggregator's thinking must reach the caller as it arrives, the same way a
// direct call to that model would stream it.
func TestEnsembleForwardsAggregatorReasoningToClient(t *testing.T) {
	dispatch := &ensembleSSEDispatch{frames: map[string][]string{
		"gpt-5": {
			reasoningFrame("PROPOSER_PRIVATE_THINKING"),
			contentFrame("candidate"),
			finalFrame(),
		},
		"gpt-5.1": {
			reasoningFrame("AGGREGATOR_THINKING_ONE "),
			reasoningFrame("AGGREGATOR_THINKING_TWO"),
			contentFrame("final answer"),
			finalFrame(),
		},
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
	body := recorder.Body.String()

	require.Contains(t, body, "AGGREGATOR_THINKING_ONE",
		"the aggregator's own reasoning must reach the caller, like a direct call")
	require.Contains(t, body, "AGGREGATOR_THINKING_TWO")
	require.Contains(t, body, `"content":"final answer"`)

	// Parallel proposers would interleave into one unreadable reasoning field, so
	// only the aggregator is wired to the caller's stream.
	require.NotContains(t, body, "PROPOSER_PRIVATE_THINKING",
		"proposer reasoning must stay internal")
}

// The forwarded reasoning is model output, not our scheduling log, so it carries
// no elapsed stamp and must not be confused with a trace line.
func TestEnsembleAggregatorReasoningIsNotStampedLikeTrace(t *testing.T) {
	dispatch := &ensembleSSEDispatch{frames: map[string][]string{
		"gpt-5":   {contentFrame("candidate"), finalFrame()},
		"gpt-5.1": {reasoningFrame("RAW_MODEL_THINKING"), contentFrame("done"), finalFrame()},
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
	deltas := ensembleReasoningTrace(t, recorder.Body.String())

	var forwarded string
	for _, delta := range deltas {
		if strings.Contains(delta, "RAW_MODEL_THINKING") {
			forwarded = delta
		}
	}
	require.NotEmpty(t, forwarded, "the model's reasoning must be present")
	require.Equal(t, "RAW_MODEL_THINKING", forwarded,
		"model reasoning is forwarded verbatim; only our own trace lines get an elapsed stamp")
}

// A member's reasoning must not be forwarded when the caller did not ask for a
// stream: there is no channel for it, and it must never contaminate the JSON body.
func TestEnsembleNonStreamingCallDropsAggregatorReasoning(t *testing.T) {
	dispatch := &ensembleSSEDispatch{frames: map[string][]string{
		"gpt-5":   {contentFrame("candidate"), finalFrame()},
		"gpt-5.1": {reasoningFrame("HIDDEN_THINKING"), contentFrame("done"), finalFrame()},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{AggregatorEnabled: true, MinProposers: 1, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":false}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "HIDDEN_THINKING")
	require.NotContains(t, recorder.Body.String(), "reasoning_content")
}

// A provider is free to split one SSE line across several writes. The tee carries
// the partial line over, otherwise reasoning would be silently lost whenever a
// frame straddled a write boundary.
func TestEnsembleForwardsReasoningSplitAcrossWrites(t *testing.T) {
	dispatch := &ensembleSSEDispatch{
		splitEvery: 7,
		frames: map[string][]string{
			"gpt-5":   {contentFrame("candidate"), finalFrame()},
			"gpt-5.1": {reasoningFrame("SPLIT_ACROSS_MANY_WRITES"), contentFrame("done"), finalFrame()},
		},
	}

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
	require.Contains(t, recorder.Body.String(), "SPLIT_ACROSS_MANY_WRITES",
		"a frame split across writes must still be forwarded")
	require.Contains(t, recorder.Body.String(), `"content":"done"`,
		"splitting writes must not corrupt the normalized answer")
}

// The tee is a pass-through: whatever the sub-call pipeline reads afterwards must
// be byte-identical to what the provider wrote, or normalization breaks.
func TestEnsembleReasoningTeePreservesRecordedBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	tee := newEnsembleReasoningTee(recorder, nil)

	payload := "data: " + reasoningFrame("x") + "\n\ndata: " + contentFrame("y") + "\n\ndata: [DONE]\n\n"
	n, err := tee.Write([]byte(payload))
	require.NoError(t, err)
	require.Equal(t, len(payload), n)
	require.Equal(t, payload, recorder.Body.String(),
		"the tee must not alter the buffered body the normalizer reads")
}

// Once a sub-call returns, a provider that ignores cancellation may still be
// writing. The sink is shared with the client stream, so a late write must not be
// able to appear after the response has been finished.
func TestEnsembleReasoningTeeStopsForwardingAfterDisable(t *testing.T) {
	var mu sync.Mutex
	seen := make([]string, 0, 2)
	tee := newEnsembleReasoningTee(httptest.NewRecorder(), func(text string) {
		mu.Lock()
		seen = append(seen, text)
		mu.Unlock()
	})

	_, _ = tee.Write([]byte("data: " + reasoningFrame("before") + "\n\n"))
	tee.disable()
	_, _ = tee.Write([]byte("data: " + reasoningFrame("after") + "\n\n"))

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{"before"}, seen,
		"a write arriving after the sub-call returned must not reach the client stream")
}

// ensembleResponseMembers digs ensemble_metadata.members out of a built response.
// Each hop is asserted separately: errcheck runs with check-type-assertions, so a
// bare chained assertion fails lint, and a checked one names the hop that broke if
// the response shape ever changes.
func ensembleResponseMembers(t *testing.T, response map[string]any) []ensembleMemberStat {
	t.Helper()

	metadata, ok := response["ensemble_metadata"].(map[string]any)
	require.True(t, ok, "response must carry ensemble_metadata when expose_metadata is on")
	members, ok := metadata["members"].([]ensembleMemberStat)
	require.True(t, ok, "ensemble_metadata.members must be a member-stat slice")
	return members
}

// members[].content repeats every member's full answer. Only the admin diagnostic
// view renders it; on the production path it multiplies the response size by the
// member count for data no caller reads.
func TestEnsembleMetadataOmitsMemberContentOnProductionPath(t *testing.T) {
	stats := []ensembleMemberStat{
		{Model: "gpt-5", Role: service.EnsembleRoleProposer, Status: ensembleStatusOK, Content: "FULL_MEMBER_ANSWER"},
	}

	production := buildEnsembleChatResponse("ensemble", "final", nil, stats, true, false, false, time.Second)
	members := ensembleResponseMembers(t, production)
	require.Len(t, members, 1)
	require.Equal(t, "", members[0].Content,
		"a caller already has the final answer; echoing every member's answer back is pure payload")
	require.Equal(t, "gpt-5", members[0].Model, "the useful per-member fields must survive")

	diagnostic := buildEnsembleChatResponse("ensemble", "final", nil, stats, true, true, false, time.Second)
	diagnosticMembers := ensembleResponseMembers(t, diagnostic)
	require.Equal(t, "FULL_MEMBER_ANSWER", diagnosticMembers[0].Content,
		"the admin diagnostic view still needs the raw member answers")

	// Trimming must not mutate the caller's slice: the same stats feed the logs.
	require.Equal(t, "FULL_MEMBER_ANSWER", stats[0].Content)
}
