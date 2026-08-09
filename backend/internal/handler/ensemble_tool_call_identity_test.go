package handler

// Tests for tool-call identity across the ensemble boundary.
//
// A tool call is only usable if the caller can pair the result it sends back with
// the call it answers, and that pairing is the id. A member that streams a
// tool_calls delta without one used to be forwarded verbatim, so the caller
// stored a tool part with an empty call id, and every later turn re-serialized
// that unpairable turn — the session was dead from then on, not just the one
// call. Helpers used here (newEnsembleHandlerRequest, ensembleStreamFrames) live
// in the sibling ensemble test files.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// rawSSEDispatch answers every sub-call with a fixed SSE body, which is what a
// member call actually receives: proposers are always forced to stream.
func rawSSEDispatch(sse string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.String(http.StatusOK, sse)
	}
}

// toolCallSSE builds a member stream whose tool_calls delta carries whatever id
// shape the test is about.
func toolCallSSE(idField string) string {
	return "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"tool_calls\":[{\"index\":0," + idField +
		"\"type\":\"function\",\"function\":{\"name\":\"Bash\",\"arguments\":\"{}\"}}]},\"finish_reason\":null}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":1}}\n\n" +
		"data: [DONE]\n\n"
}

func ensembleToolCallRequest(t *testing.T, sse, clientBody string) *httptest.ResponseRecorder {
	t.Helper()
	return newEnsembleHandlerRequest(t,
		[]service.EnsembleProposer{{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Enabled: true}},
		service.EnsembleConfig{MinProposers: 1},
		rawSSEDispatch(sse),
		clientBody,
	)
}

const toolCallClientBody = `{"model":"ensemble","messages":[{"role":"user","content":"run it"}],` +
	`"tools":[{"type":"function","function":{"name":"Bash"}}],"stream":false}`

const toolCallClientBodyStream = `{"model":"ensemble","messages":[{"role":"user","content":"run it"}],` +
	`"tools":[{"type":"function","function":{"name":"Bash"}}],"stream":true}`

// An upstream that omits the id entirely is the case that killed a real session:
// the caller stored callID "" and could not build the tool result that answers
// the call, so its next request failed schema validation before it was even sent
// — and kept failing, because the unpairable turn is permanent history now.
func TestEnsembleSynthesizesMissingToolCallID(t *testing.T) {
	recorder := ensembleToolCallRequest(t, toolCallSSE(""), toolCallClientBody)
	require.Equal(t, http.StatusOK, recorder.Code)

	body := recorder.Body.String()
	require.Equal(t, "tool_calls", gjson.Get(body, "choices.0.finish_reason").String())
	id := gjson.Get(body, "choices.0.message.tool_calls.0.id")
	require.True(t, id.Exists(), "a tool call must never leave the boundary without an id")
	require.NotEmpty(t, id.String())
	require.Equal(t, "Bash", gjson.Get(body, "choices.0.message.tool_calls.0.function.name").String())
}

// An explicit empty string is the same failure as an absent key, and gjson
// reports it as an existing field, so it needs its own guard.
func TestEnsembleSynthesizesBlankToolCallID(t *testing.T) {
	recorder := ensembleToolCallRequest(t, toolCallSSE(`"id":"",`), toolCallClientBody)
	require.Equal(t, http.StatusOK, recorder.Code)

	id := gjson.Get(recorder.Body.String(), "choices.0.message.tool_calls.0.id").String()
	require.NotEmpty(t, id, "a blank id is as unpairable as a missing one")
}

// A synthesized id is a repair of last resort. Whatever the upstream sent is what
// the caller has to echo back, so an id we did not invent must survive byte for
// byte — including shapes no OpenAI model would emit.
func TestEnsemblePreservesUpstreamToolCallID(t *testing.T) {
	recorder := ensembleToolCallRequest(t, toolCallSSE(`"id":"functions.Bash:3",`), toolCallClientBody)
	require.Equal(t, http.StatusOK, recorder.Code)

	require.Equal(t, "functions.Bash:3",
		gjson.Get(recorder.Body.String(), "choices.0.message.tool_calls.0.id").String())
}

// The streaming path emits tool_calls through finish() rather than the response
// builder, so it needs the same guarantee: both paths are reached by real
// clients.
func TestEnsembleStreamedToolCallCarriesID(t *testing.T) {
	recorder := ensembleToolCallRequest(t, toolCallSSE(""), toolCallClientBodyStream)
	require.Equal(t, http.StatusOK, recorder.Code)

	var found bool
	for _, frame := range ensembleStreamFrames(t, recorder.Body.String()) {
		calls := gjson.Get(frame, "choices.0.delta.tool_calls")
		if !calls.IsArray() || len(calls.Array()) == 0 {
			continue
		}
		found = true
		require.NotEmpty(t, calls.Get("0.id").String(),
			"a streamed tool call must carry an id the caller can pair a result with")
		require.Equal(t, "Bash", calls.Get("0.function.name").String())
	}
	require.True(t, found, "the stream must carry the tool call")
}

// A call with no function name tells the caller nothing about what to invoke.
// Forwarding it produces the same unpairable turn as a missing id, so it is
// dropped and the member is reported as having produced nothing.
func TestEnsembleDropsToolCallWithoutFunctionName(t *testing.T) {
	sse := "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"tool_calls\":[{\"index\":0," +
		"\"type\":\"function\",\"function\":{\"arguments\":\"{}\"}}]},\"finish_reason\":null}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"tool_calls\"}]}\n\n" +
		"data: [DONE]\n\n"

	recorder := ensembleToolCallRequest(t, sse, toolCallClientBody)

	// No content and no usable tool call is an empty completion, which is the
	// existing contract for a member that produced nothing.
	require.NotContains(t, recorder.Body.String(), `"tool_calls"`)
}
