package handler

// Tests for merging streamed tool-call fragments back into whole calls.
//
// Members always stream internally, so every tool call reaches the ensemble as a
// series of deltas that must be reassembled before the answer is normalized. The
// fragments are correlated by tool_calls[].index, and reading that key with
// gjson.Int() silently turns "absent" into 0: every index-less fragment landed in
// the same slot, and since names and arguments are concatenated, two distinct
// calls merged into one entry named "BashRead" carrying two argument objects glued
// together — a call no client can execute.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// ensembleSSEBody joins data: frames into the member body shape the normalizer
// reads, terminated by the sentinel a real stream ends with.
func ensembleSSEBody(frames ...string) []byte {
	var sb strings.Builder
	for _, frame := range frames {
		_, _ = sb.WriteString("data: " + frame + "\n\n")
	}
	_, _ = sb.WriteString("data: [DONE]\n\n")
	return []byte(sb.String())
}

// toolCallDelta wraps raw tool_calls array entries in a chunk.
func toolCallDelta(calls string) string {
	return `{"id":"x","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"tool_calls":[` + calls + `]}}]}`
}

func normalizedToolCalls(t *testing.T, frames ...string) []gjson.Result {
	t.Helper()
	normalized := normalizeEnsembleChatPayload(ensembleSSEBody(frames...))
	calls := gjson.GetBytes(normalized, "choices.0.message.tool_calls")
	require.True(t, calls.IsArray(), "the normalized answer must carry a tool_calls array")
	return calls.Array()
}

// The failure that motivated this: two complete calls, neither carrying an index,
// were merged into one unexecutable call.
func TestNormalizeKeepsIndexlessToolCallsSeparate(t *testing.T) {
	calls := normalizedToolCalls(t,
		toolCallDelta(`{"type":"function","function":{"name":"Bash","arguments":"{\"cmd\":\"ls\"}"}}`),
		toolCallDelta(`{"type":"function","function":{"name":"Read","arguments":"{\"path\":\"a\"}"}}`),
		finalFrame(),
	)

	require.Len(t, calls, 2, "an absent index cannot mean these are the same call")
	require.Equal(t, "Bash", calls[0].Get("function.name").String(),
		"merging the slots concatenates the names into something like BashRead")
	require.Equal(t, `{"cmd":"ls"}`, calls[0].Get("function.arguments").String())
	require.Equal(t, "Read", calls[1].Get("function.name").String())
	require.Equal(t, `{"path":"a"}`, calls[1].Get("function.arguments").String())
}

// Two index-less calls in a single delta array are unambiguously distinct, and
// array position must not be mistaken for a correlation key either.
func TestNormalizeKeepsIndexlessToolCallsInOneDeltaSeparate(t *testing.T) {
	calls := normalizedToolCalls(t,
		toolCallDelta(`{"type":"function","function":{"name":"Bash","arguments":"{}"}},`+
			`{"type":"function","function":{"name":"Read","arguments":"{}"}}`),
		finalFrame(),
	)

	require.Len(t, calls, 2)
	require.Equal(t, "Bash", calls[0].Get("function.name").String())
	require.Equal(t, "Read", calls[1].Get("function.name").String())
}

// The reason index exists: a conformant provider opens each call once and then
// sends argument fragments that identify their call only by that key.
func TestNormalizeMergesIndexedToolCallFragments(t *testing.T) {
	calls := normalizedToolCalls(t,
		toolCallDelta(`{"index":0,"id":"call_a","type":"function","function":{"name":"Bash","arguments":""}}`),
		toolCallDelta(`{"index":1,"id":"call_b","type":"function","function":{"name":"Read","arguments":""}}`),
		toolCallDelta(`{"index":0,"function":{"arguments":"{\"cmd\":"}}`),
		toolCallDelta(`{"index":1,"function":{"arguments":"{\"path\":"}}`),
		toolCallDelta(`{"index":0,"function":{"arguments":"\"ls\"}"}}`),
		toolCallDelta(`{"index":1,"function":{"arguments":"\"a\"}"}}`),
		finalFrame(),
	)

	require.Len(t, calls, 2, "interleaved fragments of two calls must not become four")
	require.Equal(t, "call_a", calls[0].Get("id").String())
	require.Equal(t, `{"cmd":"ls"}`, calls[0].Get("function.arguments").String())
	require.Equal(t, "call_b", calls[1].Get("id").String())
	require.Equal(t, `{"path":"a"}`, calls[1].Get("function.arguments").String())
}

// Splitting on "absent index means a new call" must not break the provider that
// omits index and still fragments: only the opening frame of a call carries an id
// or a name, so an arguments-only fragment continues the call in flight. Without
// this the arguments would be scattered across slots and dropped for having no
// name.
func TestNormalizeMergesIndexlessArgumentFragments(t *testing.T) {
	calls := normalizedToolCalls(t,
		toolCallDelta(`{"id":"call_a","type":"function","function":{"name":"Bash","arguments":""}}`),
		toolCallDelta(`{"function":{"arguments":"{\"cmd\":"}}`),
		toolCallDelta(`{"function":{"arguments":"\"ls\"}"}}`),
		finalFrame(),
	)

	require.Len(t, calls, 1, "an arguments-only fragment extends the open call, it does not start one")
	require.Equal(t, "Bash", calls[0].Get("function.name").String())
	require.Equal(t, `{"cmd":"ls"}`, calls[0].Get("function.arguments").String())
}

// A name may itself arrive in pieces under one index; that concatenation is
// correct and must survive.
func TestNormalizeMergesIndexedNameFragments(t *testing.T) {
	calls := normalizedToolCalls(t,
		toolCallDelta(`{"index":0,"id":"call_a","type":"function","function":{"name":"Ba"}}`),
		toolCallDelta(`{"index":0,"function":{"name":"sh","arguments":"{}"}}`),
		finalFrame(),
	)

	require.Len(t, calls, 1)
	require.Equal(t, "Bash", calls[0].Get("function.name").String())
}

// The emitted array is accumulated by index on the client side, so a sparse
// upstream numbering has to be closed up: a hole is where a strict SDK
// materializes an empty, unexecutable call.
func TestNormalizeRenumbersSparseToolCallIndexes(t *testing.T) {
	calls := normalizedToolCalls(t,
		toolCallDelta(`{"index":3,"id":"call_a","type":"function","function":{"name":"Bash","arguments":"{}"}}`),
		toolCallDelta(`{"index":9,"id":"call_b","type":"function","function":{"name":"Read","arguments":"{}"}}`),
		finalFrame(),
	)

	require.Len(t, calls, 2)
	require.Equal(t, int64(0), calls[0].Get("index").Int())
	require.Equal(t, int64(1), calls[1].Get("index").Int())
}

// Some providers render the index as a string. That is still a correlation key,
// and treating it as absent would split one call into several.
func TestNormalizeMergesFragmentsWithStringIndex(t *testing.T) {
	calls := normalizedToolCalls(t,
		toolCallDelta(`{"index":"0","id":"call_a","type":"function","function":{"name":"Bash","arguments":"{\"cmd\":"}}`),
		toolCallDelta(`{"index":"0","function":{"arguments":"\"ls\"}"}}`),
		finalFrame(),
	)

	require.Len(t, calls, 1)
	require.Equal(t, `{"cmd":"ls"}`, calls[0].Get("function.arguments").String())
}

// A non-numeric index is not a key at all. gjson renders it as 0, which is
// exactly the collapse this guards against.
func TestNormalizeTreatsUnparseableIndexAsAbsent(t *testing.T) {
	calls := normalizedToolCalls(t,
		toolCallDelta(`{"index":"first","type":"function","function":{"name":"Bash","arguments":"{}"}}`),
		toolCallDelta(`{"index":"second","type":"function","function":{"name":"Read","arguments":"{}"}}`),
		finalFrame(),
	)

	require.Len(t, calls, 2)
	require.Equal(t, "Bash", calls[0].Get("function.name").String())
	require.Equal(t, "Read", calls[1].Get("function.name").String())
}

// Dropping an unusable call must not leave a hole behind either: the surviving
// calls are what the caller accumulates by index.
func TestRepairRenumbersAfterDroppingUnusableCall(t *testing.T) {
	repaired := repairEnsembleToolCallIdentity([]map[string]any{
		{"index": 0, "type": "function", "function": map[string]any{"arguments": "{}"}},
		{"index": 1, "id": "call_b", "type": "function", "function": map[string]any{"name": "Read", "arguments": "{}"}},
	})

	require.Len(t, repaired, 1, "a call with no function name tells the caller nothing to invoke")
	require.Equal(t, 0, repaired[0]["index"], "the surviving call must not sit at index 1 with a hole below it")
}

// A non-stream answer from a real provider has no index field on its tool calls.
// Renumbering must not invent one, because that is a wire-shape change on a path
// this repair only passes through.
func TestRepairDoesNotAddIndexWhereUpstreamHadNone(t *testing.T) {
	repaired := repairEnsembleToolCallIdentity([]map[string]any{
		{"id": "call_a", "type": "function", "function": map[string]any{"name": "Bash", "arguments": "{}"}},
	})

	require.Len(t, repaired, 1)
	_, hasIndex := repaired[0]["index"]
	require.False(t, hasIndex, "the repair fixes identity; it does not restructure a conformant call")
}
