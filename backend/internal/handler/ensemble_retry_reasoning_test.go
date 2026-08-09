package handler

// Tests for the interaction between the empty-completion retry and the live
// reasoning stream.
//
// An empty completion is retried once. The aggregator is also the one member
// whose reasoning is teed to the caller as it arrives, and both attempts used to
// receive the same sink — so a model that thinks out loud and then returns an
// empty completion had its thinking replayed to the caller on the retry. Streamed
// bytes cannot be retracted, which is what makes this a correctness bug rather
// than a cosmetic one: the caller sees one answer preceded by two thought
// streams and cannot tell that a retry happened.
//
// Helpers used here (newEnsembleHandlerRequest, ensembleReasoningTrace,
// reasoningFrame, contentFrame, finalFrame) live in the sibling ensemble test
// files.

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// ensembleAttemptDispatch answers each sub-call of the same model with the next
// script in line, which is how a retry is made observable: attempt 1 returns an
// empty completion, attempt 2 returns a real one.
type ensembleAttemptDispatch struct {
	mu       sync.Mutex
	attempts map[string][][]string
	calls    map[string]int
}

func newEnsembleAttemptDispatch(attempts map[string][][]string) *ensembleAttemptDispatch {
	return &ensembleAttemptDispatch{attempts: attempts, calls: make(map[string]int)}
}

func (d *ensembleAttemptDispatch) dispatch(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	model := gjson.GetBytes(body, "model").String()

	d.mu.Lock()
	scripts, ok := d.attempts[model]
	attempt := d.calls[model]
	d.calls[model]++
	d.mu.Unlock()

	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "unexpected model " + model}})
		return
	}
	// A model called more times than scripted repeats its last script, so the
	// test fails on the assertion rather than on a stub lookup.
	if attempt >= len(scripts) {
		attempt = len(scripts) - 1
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.WriteHeader(http.StatusOK)
	for _, frame := range scripts[attempt] {
		_, _ = c.Writer.WriteString("data: " + frame + "\n\n")
		c.Writer.Flush()
	}
	_, _ = c.Writer.WriteString("data: [DONE]\n\n")
	c.Writer.Flush()
}

func (d *ensembleAttemptDispatch) callCount(model string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.calls[model]
}

func ensembleRetryRequest(t *testing.T, dispatch gin.HandlerFunc) string {
	t.Helper()
	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{AggregatorEnabled: true, MinProposers: 1, ExposeMetadata: true},
		dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	)
	require.Equal(t, http.StatusOK, recorder.Code)
	return recorder.Body.String()
}

// The retried attempt must not replay thinking the caller already received. This
// is the common shape of an empty completion: the model spends its budget
// reasoning and never starts the answer.
func TestEnsembleRetryDoesNotReplayForwardedReasoning(t *testing.T) {
	dispatch := newEnsembleAttemptDispatch(map[string][][]string{
		"gpt-5": {{contentFrame("candidate"), finalFrame()}},
		"gpt-5.1": {
			// Attempt 1: thinks, then returns nothing — an empty completion.
			{reasoningFrame("AGG_THINKING"), finalFrame()},
			// Attempt 2: thinks the same way, and this time answers.
			{reasoningFrame("AGG_THINKING"), contentFrame("final answer"), finalFrame()},
		},
	})

	body := ensembleRetryRequest(t, dispatch.dispatch)

	require.Equal(t, 2, dispatch.callCount("gpt-5.1"), "an empty completion must still be retried")
	require.Contains(t, body, `"content":"final answer"`, "the retry's answer must reach the caller")

	deltas := ensembleReasoningTrace(t, body)
	seen := 0
	for _, delta := range deltas {
		if strings.Contains(delta, "AGG_THINKING") {
			seen++
		}
	}
	require.Equal(t, 1, seen,
		"the caller keeps the thinking that already reached them; the retry must not stream it again")
}

// When the failed attempt forwarded nothing at all there is no earlier fragment
// for the retry to contradict, so the live sink must stay attached — otherwise
// suppressing the replay would cost the caller its thinking on every retry.
func TestEnsembleRetryKeepsReasoningWhenNothingWasForwarded(t *testing.T) {
	dispatch := newEnsembleAttemptDispatch(map[string][][]string{
		"gpt-5": {{contentFrame("candidate"), finalFrame()}},
		"gpt-5.1": {
			// Attempt 1: an entirely empty body, no reasoning at all.
			{},
			{reasoningFrame("RETRY_THINKING"), contentFrame("final answer"), finalFrame()},
		},
	})

	body := ensembleRetryRequest(t, dispatch.dispatch)

	require.Equal(t, 2, dispatch.callCount("gpt-5.1"))
	require.Contains(t, body, `"content":"final answer"`)
	require.Contains(t, body, "RETRY_THINKING",
		"nothing was forwarded before, so the retry's thinking is the caller's first and only")
}

// The ordinary path must be untouched: a member that answers on its first attempt
// is called once and its reasoning is forwarded once.
func TestEnsembleSuccessfulCallForwardsReasoningOnce(t *testing.T) {
	dispatch := newEnsembleAttemptDispatch(map[string][][]string{
		"gpt-5":   {{contentFrame("candidate"), finalFrame()}},
		"gpt-5.1": {{reasoningFrame("ONLY_THINKING"), contentFrame("final answer"), finalFrame()}},
	})

	body := ensembleRetryRequest(t, dispatch.dispatch)

	require.Equal(t, 1, dispatch.callCount("gpt-5.1"), "a successful call must not be retried")
	seen := 0
	for _, delta := range ensembleReasoningTrace(t, body) {
		if strings.Contains(delta, "ONLY_THINKING") {
			seen++
		}
	}
	require.Equal(t, 1, seen)
}
