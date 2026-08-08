package handler

// Tests for the in-band failure frame and for the reasoning-echo strip.
//
// A fan-out failure is discovered after the SSE 200 is committed, so it can only
// be reported in-band — as a normal assistant turn. That has two consequences a
// client depends on: the stream must still terminate like any other stream, and
// the notice must not come back as model context on the next turn. Helpers used
// here (ensembleDispatchStub, chatCompletion, newEnsembleHandlerRequest,
// bodyCapturingDispatch) live in the sibling ensemble test files.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// ensembleStreamFrames returns every parseable data: frame of an SSE body.
func ensembleStreamFrames(t *testing.T, stream string) []string {
	t.Helper()
	frames := make([]string, 0, 8)
	for _, raw := range strings.Split(stream, "\n") {
		after, found := strings.CutPrefix(strings.TrimSpace(raw), "data:")
		if !found {
			continue
		}
		payload := strings.TrimSpace(after)
		if payload == "" || payload == "[DONE]" || !gjson.Valid(payload) {
			continue
		}
		frames = append(frames, payload)
	}
	return frames
}

// failedEnsembleStream drives a fan-out where every member fails while
// min_proposers demands more, which is the only path that reaches fail().
func failedEnsembleStream(t *testing.T) string {
	t.Helper()
	dispatch := &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusBadGateway, body: gin.H{"error": gin.H{"message": "upstream down"}}},
		"gpt-5.1": {status: http.StatusBadGateway, body: gin.H{"error": gin.H{"message": "upstream down"}}},
	}}

	recorder := newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{
			{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
			{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.1", Enabled: true},
		},
		service.EnsembleConfig{MinProposers: 2, ExposeMetadata: true},
		dispatch.dispatch,
		`{"model":"ensemble","messages":[{"role":"user","content":"hi"}],"stream":true}`,
	)
	require.Equal(t, http.StatusOK, recorder.Code)
	return recorder.Body.String()
}

// A failed stream must terminate exactly like a successful one. Without a
// terminal finish_reason the client is left holding an assistant turn that never
// ended, which is not a state it can cleanly resume from.
func TestEnsembleFailedStreamEmitsTerminalFinishReason(t *testing.T) {
	body := failedEnsembleStream(t)

	reasons := make([]string, 0, 2)
	for _, frame := range ensembleStreamFrames(t, body) {
		if reason := gjson.Get(frame, "choices.0.finish_reason"); reason.Type == gjson.String && reason.String() != "" {
			reasons = append(reasons, reason.String())
		}
	}
	require.Equal(t, []string{"stop"}, reasons,
		"a failed fan-out must close the stream with exactly one terminal finish_reason")

	// The terminal frame has to be the last one before the sentinel, otherwise a
	// client that stops reading at finish_reason would truncate the notice.
	frames := ensembleStreamFrames(t, body)
	last := frames[len(frames)-1]
	require.Equal(t, "stop", gjson.Get(last, "choices.0.finish_reason").String())
	require.True(t, strings.HasSuffix(strings.TrimSpace(body), "data: [DONE]"))
}

// The notice still rides delta.content so every client shows it, and it still
// keeps the standard chunk shape strict SDKs require.
func TestEnsembleFailureNoticeStaysVisibleAndMarked(t *testing.T) {
	body := failedEnsembleStream(t)

	var notice string
	for _, frame := range ensembleStreamFrames(t, body) {
		if content := gjson.Get(frame, "choices.0.delta.content"); content.Type == gjson.String && content.String() != "" {
			notice = frame
		}
	}
	require.NotEmpty(t, notice, "the failure must be delivered in-band")
	require.Equal(t, "chat.completion.chunk", gjson.Get(notice, "object").String())

	text := gjson.Get(notice, "choices.0.delta.content").String()
	require.True(t, strings.HasPrefix(text, ensembleFailureNoticePrefix),
		"the notice must be marked so the next turn can strip it")
	require.Contains(t, text, "minimum 2 required", "the real cause must stay readable")
}

// The notice is stored by the client as a normal assistant turn and echoed back.
// Feeding it to members would make our own error text part of their context.
func TestEnsembleStripsEchoedFailureNotice(t *testing.T) {
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
			`{"role":"assistant","content":"`+ensembleFailureNoticePrefix+`Only 0 proposers succeeded, minimum 2 required"},`+
			`{"role":"user","content":"retry"}],"stream":true}`,
	)
	require.Equal(t, http.StatusOK, recorder.Code)

	require.Len(t, capture.bodies, 1)
	require.NotContains(t, capture.bodies[0], "minimum 2 required",
		"our own failure text must never become member context")
	require.NotContains(t, capture.bodies[0], ensembleFailureNoticePrefix)
	// The surrounding real conversation has to survive intact.
	require.Contains(t, capture.bodies[0], `"content":"first"`)
	require.Contains(t, capture.bodies[0], `"content":"retry"`)
}

// Anthropic-shaped clients store thinking as a content block, not as a sibling
// field, so deleting messages[N].thinking alone leaves the echo in place.
func TestStripEnsembleReasoningRemovesThinkingContentBlocks(t *testing.T) {
	body := `{"model":"ensemble","messages":[` +
		`{"role":"user","content":"hi"},` +
		`{"role":"assistant","content":[` +
		`{"type":"thinking","thinking":"[0.0s] 并行调用 2 个模型"},` +
		`{"type":"text","text":"the real answer"}]},` +
		`{"role":"user","content":"next"}]}`

	out, err := stripEnsembleReasoningFields([]byte(body))
	require.NoError(t, err)
	require.NotContains(t, string(out), "并行调用", "echoed thinking block must be removed")
	require.NotContains(t, string(out), `"type":"thinking"`)
	require.Contains(t, string(out), "the real answer", "the answer block must survive")

	// The assistant turn keeps its place in the conversation.
	messages := gjson.GetBytes(out, "messages")
	require.Equal(t, 3, int(messages.Get("#").Int()))
	require.Equal(t, "assistant", messages.Get("1.role").String())
}

// The OpenAI "reasoning" spelling used to slip through the cheap-reject probe
// entirely, so a client echoing that field polluted every member.
func TestStripEnsembleReasoningRemovesReasoningSpelling(t *testing.T) {
	body := `{"model":"ensemble","messages":[` +
		`{"role":"user","content":"hi"},` +
		`{"role":"assistant","content":"answer","reasoning":"[0.0s] 候选回答 2/2 可用"},` +
		`{"role":"user","content":"next"}]}`

	out, err := stripEnsembleReasoningFields([]byte(body))
	require.NoError(t, err)
	require.NotContains(t, string(out), "候选回答")
	require.NotContains(t, string(out), `"reasoning"`)
	require.Contains(t, string(out), `"content":"answer"`)
}

// An assistant turn that was nothing but thinking carries no conversation fact
// once stripped, and an empty content array is rejected by some upstreams.
func TestStripEnsembleReasoningDropsTurnLeftEmpty(t *testing.T) {
	body := `{"model":"ensemble","messages":[` +
		`{"role":"user","content":"hi"},` +
		`{"role":"assistant","content":[{"type":"thinking","thinking":"only thinking"}]},` +
		`{"role":"user","content":"next"}]}`

	out, err := stripEnsembleReasoningFields([]byte(body))
	require.NoError(t, err)
	messages := gjson.GetBytes(out, "messages")
	require.Equal(t, 2, int(messages.Get("#").Int()),
		"a turn with nothing left but thinking must be dropped, not sent hollow")
	require.Equal(t, "user", messages.Get("0.role").String())
	require.Equal(t, "user", messages.Get("1.role").String())
}

// Tool calls are the agent loop's continuation boundary. A turn carrying them is
// never a failure notice and must survive even if its text resembles one.
func TestStripEnsembleReasoningKeepsToolCallTurns(t *testing.T) {
	body := `{"model":"ensemble","messages":[` +
		`{"role":"user","content":"hi"},` +
		`{"role":"assistant","content":"` + ensembleFailureNoticePrefix + `looks like a notice",` +
		`"tool_calls":[{"id":"call_1","type":"function","function":{"name":"f","arguments":"{}"}}]},` +
		`{"role":"user","content":"next"}]}`

	out, err := stripEnsembleReasoningFields([]byte(body))
	require.NoError(t, err)
	require.Contains(t, string(out), "call_1", "a tool-call turn must never be dropped")
	require.Equal(t, 3, int(gjson.GetBytes(out, "messages.#").Int()))
}

// Several echoed shapes in one conversation must all be handled in a single
// pass, including the index shifts caused by dropping whole messages.
func TestStripEnsembleReasoningHandlesMixedShapes(t *testing.T) {
	body := `{"model":"ensemble","messages":[` +
		`{"role":"user","content":"q1"},` +
		`{"role":"assistant","content":"a1","reasoning_content":"trace one"},` +
		`{"role":"user","content":"q2"},` +
		`{"role":"assistant","content":"` + ensembleFailureNoticePrefix + `it broke"},` +
		`{"role":"user","content":"q3"},` +
		`{"role":"assistant","content":[{"type":"thinking","thinking":"trace two"},{"type":"text","text":"a3"}]},` +
		`{"role":"user","content":"q4"}]}`

	out, err := stripEnsembleReasoningFields([]byte(body))
	require.NoError(t, err)
	require.NotContains(t, string(out), "trace one")
	require.NotContains(t, string(out), "trace two")
	require.NotContains(t, string(out), "it broke")
	// Every real turn survives; only the failure notice is dropped.
	require.Equal(t, 6, int(gjson.GetBytes(out, "messages.#").Int()))
	for _, want := range []string{"q1", "a1", "q2", "q3", "a3", "q4"} {
		require.Contains(t, string(out), want)
	}
}

// The cheap-reject probe must not fire on a clean conversation, and a clean body
// has to come back byte-identical so the fan-out sees exactly what the client sent.
func TestStripEnsembleReasoningLeavesCleanBodyUntouched(t *testing.T) {
	body := `{"model":"ensemble","messages":[{"role":"user","content":"hi"},{"role":"assistant","content":"answer"}]}`
	out, err := stripEnsembleReasoningFields([]byte(body))
	require.NoError(t, err)
	require.Equal(t, body, string(out))
}
