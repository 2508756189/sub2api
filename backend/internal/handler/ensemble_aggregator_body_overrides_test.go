package handler

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const ensembleOverrideRequest = `{"model":"ensemble-public","messages":[{"role":"user","content":"hello"}],"stream":false}`

func TestApplyEnsembleAggregatorBodyOverrideSetsConfiguredPath(t *testing.T) {
	out, err := applyEnsembleAggregatorBodyOverrides(
		[]byte(ensembleOverrideRequest),
		[]byte(ensembleOverrideRequest),
		map[string]any{"reasoning_effort": "max"},
	)
	require.NoError(t, err)
	require.Equal(t, "max", gjson.GetBytes(out, "reasoning_effort").String())
}

// The group setting is a default, not an override of a stated intent. This is the
// same rule applyEnsembleMaxTokens follows.
func TestApplyEnsembleAggregatorBodyOverrideYieldsToClientValue(t *testing.T) {
	request := `{"model":"ensemble-public","messages":[],"reasoning_effort":"low"}`
	out, err := applyEnsembleAggregatorBodyOverrides(
		[]byte(request),
		[]byte(request),
		map[string]any{"reasoning_effort": "max"},
	)
	require.NoError(t, err)
	require.Equal(t, "low", gjson.GetBytes(out, "reasoning_effort").String())
}

// Existence is tested against the caller's body, not the accumulating output.
// Testing against the output would let one applied override make a later one look
// client-supplied, so a config naming both a parent and its child would silently
// drop the child.
func TestApplyEnsembleAggregatorBodyOverrideChecksClientBodyNotOutput(t *testing.T) {
	request := `{"model":"ensemble-public","messages":[]}`
	out, err := applyEnsembleAggregatorBodyOverrides(
		[]byte(request),
		[]byte(request),
		map[string]any{
			"thinking":      map[string]any{"type": "enabled"},
			"thinking.type": "disabled",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "disabled", gjson.GetBytes(out, "thinking.type").String())
}

// The admin API refuses protected paths, so one reaching the runtime was written
// straight to the column. Skipping the single key costs the configured depth;
// applying it would rewrite a fan-out invariant.
func TestApplyEnsembleAggregatorBodyOverrideSkipsProtectedPathAtRuntime(t *testing.T) {
	out, err := applyEnsembleAggregatorBodyOverrides(
		[]byte(ensembleOverrideRequest),
		[]byte(ensembleOverrideRequest),
		map[string]any{
			"model":            "smuggled-model",
			"max_tokens":       999999,
			"reasoning_effort": "max",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "ensemble-public", gjson.GetBytes(out, "model").String())
	require.False(t, gjson.GetBytes(out, "max_tokens").Exists())
	require.Equal(t, "max", gjson.GetBytes(out, "reasoning_effort").String())
}

// Go randomizes map iteration. A config naming both "thinking" and "thinking.type"
// would otherwise produce a different body from run to run, so the aggregator's
// depth would depend on which order the runtime happened to pick.
func TestApplyEnsembleAggregatorBodyOverrideIsDeterministic(t *testing.T) {
	request := `{"model":"ensemble-public","messages":[]}`
	overrides := map[string]any{
		"thinking":      map[string]any{"type": "enabled"},
		"thinking.type": "disabled",
	}

	first, err := applyEnsembleAggregatorBodyOverrides([]byte(request), []byte(request), overrides)
	require.NoError(t, err)

	for i := 0; i < 200; i++ {
		out, err := applyEnsembleAggregatorBodyOverrides([]byte(request), []byte(request), overrides)
		require.NoError(t, err)
		require.Equal(t, string(first), string(out), "override application must not depend on map iteration order")
		require.Equal(t, "disabled", gjson.GetBytes(out, "thinking.type").String())
	}
}

func TestApplyEnsembleAggregatorBodyOverrideCreatesNestedPath(t *testing.T) {
	out, err := applyEnsembleAggregatorBodyOverrides(
		[]byte(ensembleOverrideRequest),
		[]byte(ensembleOverrideRequest),
		map[string]any{"thinking.type": "enabled", "thinking.budget_tokens": 4096},
	)
	require.NoError(t, err)
	require.Equal(t, "enabled", gjson.GetBytes(out, "thinking.type").String())
	require.Equal(t, int64(4096), gjson.GetBytes(out, "thinking.budget_tokens").Int())
}

func TestApplyEnsembleAggregatorBodyOverrideYieldsToClientNestedValue(t *testing.T) {
	request := `{"model":"ensemble-public","messages":[],"thinking":{"type":"disabled"}}`
	out, err := applyEnsembleAggregatorBodyOverrides(
		[]byte(request),
		[]byte(request),
		map[string]any{"thinking.type": "enabled"},
	)
	require.NoError(t, err)
	require.Equal(t, "disabled", gjson.GetBytes(out, "thinking.type").String())
}

func TestApplyEnsembleAggregatorBodyOverrideNoopWhenUnconfigured(t *testing.T) {
	out, err := applyEnsembleAggregatorBodyOverrides(
		[]byte(ensembleOverrideRequest),
		[]byte(ensembleOverrideRequest),
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, ensembleOverrideRequest, string(out))
}

// An unconfigured group must get byte-identical behaviour to before the field
// existed: the aggregator body carries the appended instruction and nothing else.
func TestBuildEnsembleAggregatorBodyWithoutOverridesAddsOnlyTheInstruction(t *testing.T) {
	out, err := buildEnsembleAggregatorBody(
		[]byte(ensembleOverrideRequest),
		[]ensembleProposal{{Model: "gpt-5", Content: "candidate"}},
		nil,
	)
	require.NoError(t, err)
	require.False(t, gjson.GetBytes(out, "reasoning_effort").Exists())
	require.Equal(t, 2, len(gjson.GetBytes(out, "messages").Array()))
}

func TestBuildEnsembleAggregatorBodyOverridePreservesAppendedInstruction(t *testing.T) {
	out, err := buildEnsembleAggregatorBody(
		[]byte(ensembleOverrideRequest),
		[]ensembleProposal{{Model: "gpt-5", Content: "candidate"}},
		map[string]any{"reasoning_effort": "max"},
	)
	require.NoError(t, err)

	messages := gjson.GetBytes(out, "messages").Array()
	require.Len(t, messages, 2)
	require.Equal(t, "user", messages[1].Get("role").String())
	require.Contains(t, messages[1].Get("content").String(), "candidate")
	require.Equal(t, "max", gjson.GetBytes(out, "reasoning_effort").String())
}

func ensembleOverrideMembers() []service.EnsembleProposer {
	return []service.EnsembleProposer{
		{ID: 1, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5", Priority: 20, Enabled: true},
		{ID: 2, GroupID: 7, Role: service.EnsembleRoleProposer, Model: "gpt-5.2", Priority: 10, Enabled: true},
		{ID: 3, GroupID: 7, Role: service.EnsembleRoleAggregator, Model: "gpt-5.1", Priority: 1, Enabled: true},
	}
}

func ensembleOverrideDispatch() *ensembleDispatchStub {
	return &ensembleDispatchStub{responses: map[string]dispatchResponse{
		"gpt-5":   {status: http.StatusOK, body: chatCompletion("short", 10, 3)},
		"gpt-5.2": {status: http.StatusOK, body: chatCompletion("long proposer", 11, 5)},
		"gpt-5.1": {status: http.StatusOK, body: chatCompletion("final", 4, 2)},
	}}
}

// The whole point of the feature: proposers buy diversity and stay at their
// upstream defaults, while the single judgement step gets the configured depth.
func TestEnsembleChatCompletionsAppliesOverrideToAggregatorOnly(t *testing.T) {
	dispatch := ensembleOverrideDispatch()

	recorder := newEnsembleHandlerRequest(t,
		ensembleOverrideMembers(),
		service.EnsembleConfig{
			AggregatorEnabled:       true,
			MinProposers:            2,
			AggregatorBodyOverrides: map[string]any{"reasoning_effort": "max"},
		},
		dispatch.dispatch,
		ensembleOverrideRequest,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, dispatch.bodies, 3)
	require.Equal(t, "max", gjson.Get(dispatch.bodies["gpt-5.1"], "reasoning_effort").String())
	require.False(t, gjson.Get(dispatch.bodies["gpt-5"], "reasoning_effort").Exists())
	require.False(t, gjson.Get(dispatch.bodies["gpt-5.2"], "reasoning_effort").Exists())
}

// A client that states a depth keeps it on every member, aggregator included.
func TestEnsembleChatCompletionsClientEffortReachesEveryMemberUnchanged(t *testing.T) {
	dispatch := ensembleOverrideDispatch()

	recorder := newEnsembleHandlerRequest(t,
		ensembleOverrideMembers(),
		service.EnsembleConfig{
			AggregatorEnabled:       true,
			MinProposers:            2,
			AggregatorBodyOverrides: map[string]any{"reasoning_effort": "max"},
		},
		dispatch.dispatch,
		`{"model":"ensemble-public","messages":[{"role":"user","content":"hello"}],"stream":false,"reasoning_effort":"low"}`,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	for _, model := range []string{"gpt-5", "gpt-5.2", "gpt-5.1"} {
		require.Equal(t, "low", gjson.Get(dispatch.bodies[model], "reasoning_effort").String(), "member %s", model)
	}
}

// Without a configured override no member body gains the field, which is the
// behaviour a group that never opts in must keep.
func TestEnsembleChatCompletionsWithoutOverrideLeavesMemberBodiesAlone(t *testing.T) {
	dispatch := ensembleOverrideDispatch()

	recorder := newEnsembleHandlerRequest(t,
		ensembleOverrideMembers(),
		service.EnsembleConfig{AggregatorEnabled: true, MinProposers: 2},
		dispatch.dispatch,
		ensembleOverrideRequest,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, dispatch.bodies, 3)
	for _, model := range []string{"gpt-5", "gpt-5.2", "gpt-5.1"} {
		require.False(t, gjson.Get(dispatch.bodies[model], "reasoning_effort").Exists(), "member %s", model)
	}
}

// A protected path written straight to the column must not reach the aggregator.
//
// max_tokens and tools are the load-bearing cases here because the caller's body
// carries neither: a path the client did send is already stopped by the
// client-wins guard, so asserting on "model" alone would pass even with the
// protected-path guard deleted.
func TestEnsembleChatCompletionsIgnoresProtectedOverrideEndToEnd(t *testing.T) {
	dispatch := ensembleOverrideDispatch()

	recorder := newEnsembleHandlerRequest(t,
		ensembleOverrideMembers(),
		service.EnsembleConfig{
			AggregatorEnabled: true,
			MinProposers:      2,
			AggregatorBodyOverrides: map[string]any{
				"model":            "smuggled-model",
				"max_tokens":       999999,
				"tools":            []any{map[string]any{"type": "function"}},
				"reasoning_effort": "max",
			},
		},
		dispatch.dispatch,
		ensembleOverrideRequest,
	)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotContains(t, dispatch.models, "smuggled-model")

	aggregator := dispatch.bodies["gpt-5.1"]
	require.False(t, gjson.Get(aggregator, "max_tokens").Exists())
	require.False(t, gjson.Get(aggregator, "tools").Exists())
	require.Equal(t, "max", gjson.Get(aggregator, "reasoning_effort").String())
}
