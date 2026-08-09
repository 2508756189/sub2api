//go:build unit

package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestEnsembleBodyOverridePathAllowsProviderKnobs(t *testing.T) {
	for _, path := range []string{
		"reasoning_effort",
		"thinking.type",
		"thinking.budget_tokens",
		"extra_body.thinking.budget_tokens",
		"temperature",
		"top_p",
	} {
		require.True(t, EnsembleBodyOverridePathAllowed(path), "expected %q to be allowed", path)
	}
}

func TestEnsembleBodyOverridePathRefusesRuntimeOwnedFields(t *testing.T) {
	for _, path := range []string{
		"model",
		"stream",
		"stream_options",
		"messages",
		"tools",
		"tool_choice",
		"max_tokens",
		"max_completion_tokens",
	} {
		require.False(t, EnsembleBodyOverridePathAllowed(path), "expected %q to be refused", path)
	}
}

// A protected root must stay protected through its children, otherwise the
// denylist is a speed bump: "messages" refused but "messages.0.content" allowed
// would let a config rewrite the conversation the aggregator has to judge.
func TestEnsembleBodyOverridePathRefusesChildrenOfProtectedRoot(t *testing.T) {
	for _, path := range []string{
		"messages.0.content",
		"messages.-1",
		"stream_options.include_usage",
		"tools.0.function.name",
	} {
		require.False(t, EnsembleBodyOverridePathAllowed(path), "expected %q to be refused", path)
	}
}

func TestEnsembleBodyOverridePathRefusesEmptyAfterTrim(t *testing.T) {
	require.False(t, EnsembleBodyOverridePathAllowed(""))
	require.False(t, EnsembleBodyOverridePathAllowed("   "))
}

// Leading/trailing spaces must not smuggle a protected path past the denylist.
func TestEnsembleBodyOverridePathRefusesPaddedProtectedPath(t *testing.T) {
	require.False(t, EnsembleBodyOverridePathAllowed("  model  "))
	require.False(t, EnsembleBodyOverridePathAllowed("\tmessages"))
}

func TestValidateEnsembleAggregatorBodyOverridesAcceptsEmpty(t *testing.T) {
	require.NoError(t, ValidateEnsembleAggregatorBodyOverrides(nil))
	require.NoError(t, ValidateEnsembleAggregatorBodyOverrides(map[string]any{}))
}

func TestValidateEnsembleAggregatorBodyOverridesAcceptsRealKnobs(t *testing.T) {
	require.NoError(t, ValidateEnsembleAggregatorBodyOverrides(map[string]any{
		"reasoning_effort":       "max",
		"thinking.type":          "enabled",
		"thinking.budget_tokens": 4096,
	}))
}

func TestValidateEnsembleAggregatorBodyOverridesRefusesProtectedPath(t *testing.T) {
	err := ValidateEnsembleAggregatorBodyOverrides(map[string]any{"model": "gpt-5"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_PROTECTED")
}

func TestValidateEnsembleAggregatorBodyOverridesRefusesEmptyPath(t *testing.T) {
	err := ValidateEnsembleAggregatorBodyOverrides(map[string]any{"   ": "max"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_EMPTY")
}

func TestValidateEnsembleAggregatorBodyOverridesRefusesLongPath(t *testing.T) {
	err := ValidateEnsembleAggregatorBodyOverrides(map[string]any{
		strings.Repeat("a", maxEnsembleBodyOverridePathLen+1): "max",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_TOO_LONG")
}

// gjson/sjson treat these as wildcards, escapes and modifiers. A path carrying
// one would write somewhere the admin never named.
func TestValidateEnsembleAggregatorBodyOverridesRefusesReservedCharacters(t *testing.T) {
	for _, path := range []string{"reasoning*", "thinking?type", "a#b", "a|b", "@reverse", `a\.b`} {
		err := ValidateEnsembleAggregatorBodyOverrides(map[string]any{path: "max"})
		require.Error(t, err, "expected %q to be refused", path)
		require.Contains(t, err.Error(), "ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_INVALID", "path %q", path)
	}
}

func TestValidateEnsembleAggregatorBodyOverridesRefusesTooManyEntries(t *testing.T) {
	overrides := make(map[string]any, MaxEnsembleAggregatorBodyOverrides+1)
	for i := 0; i <= MaxEnsembleAggregatorBodyOverrides; i++ {
		overrides["knob"+string(rune('a'+i))] = 1
	}
	err := ValidateEnsembleAggregatorBodyOverrides(overrides)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ENSEMBLE_AGGREGATOR_OVERRIDES_TOO_MANY")
}

// Entry count alone does not bound the payload: a handful of huge values would
// still inflate every aggregator request.
func TestValidateEnsembleAggregatorBodyOverridesRefusesOversizedPayload(t *testing.T) {
	overrides := make(map[string]any, MaxEnsembleAggregatorBodyOverrides)
	for i := 0; i < MaxEnsembleAggregatorBodyOverrides; i++ {
		overrides["knob"+string(rune('a'+i))] = strings.Repeat("x", 400)
	}
	err := ValidateEnsembleAggregatorBodyOverrides(overrides)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ENSEMBLE_AGGREGATOR_OVERRIDES_TOO_LARGE")
}

func TestValidateEnsembleAggregatorBodyOverridesRefusesUnencodableValue(t *testing.T) {
	err := ValidateEnsembleAggregatorBodyOverrides(map[string]any{"reasoning_effort": make(chan int)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ENSEMBLE_AGGREGATOR_OVERRIDE_VALUE_INVALID")
}

// A refused override has to reach the admin as 400. Bare fmt.Errorf would map to
// UnknownCode and surface as 500, which reads as "the server broke" rather than
// "fix the path you typed".
func TestValidateEnsembleAggregatorBodyOverridesReportsBadRequest(t *testing.T) {
	err := ValidateEnsembleAggregatorBodyOverrides(map[string]any{"messages": []any{}})
	require.Error(t, err)
	status, _ := infraerrors.ToHTTP(err)
	require.Equal(t, 400, status)
}

func TestNormalizeEnsembleAggregatorBodyOverridesTrimsPaths(t *testing.T) {
	normalized := NormalizeEnsembleAggregatorBodyOverrides(map[string]any{
		"  reasoning_effort  ": "max",
	})
	require.Equal(t, map[string]any{"reasoning_effort": "max"}, normalized)
}

func TestNormalizeEnsembleAggregatorBodyOverridesDropsEmptyPaths(t *testing.T) {
	normalized := NormalizeEnsembleAggregatorBodyOverrides(map[string]any{
		"reasoning_effort": "max",
		"   ":              "ignored",
	})
	require.Equal(t, map[string]any{"reasoning_effort": "max"}, normalized)
}

func TestNormalizeEnsembleAggregatorBodyOverridesReturnsNilWhenEmpty(t *testing.T) {
	require.Nil(t, NormalizeEnsembleAggregatorBodyOverrides(nil))
	require.Nil(t, NormalizeEnsembleAggregatorBodyOverrides(map[string]any{}))
	require.Nil(t, NormalizeEnsembleAggregatorBodyOverrides(map[string]any{"  ": 1}))
}

// normalizeEnsembleConfig runs before validation on the admin write path, so a
// padded path must be trimmed by the time it is judged and stored.
func TestNormalizeEnsembleConfigNormalizesAggregatorOverrides(t *testing.T) {
	cfg := normalizeEnsembleConfig(EnsembleConfig{
		AggregatorBodyOverrides: map[string]any{" reasoning_effort ": "max"},
	})
	require.Equal(t, map[string]any{"reasoning_effort": "max"}, cfg.AggregatorBodyOverrides)
}

// The column is shared with every pre-existing group. An always-present key
// would change the stored JSON shape for groups that never set an override.
func TestEnsembleConfigOmitsAggregatorOverridesWhenUnset(t *testing.T) {
	payload, err := json.Marshal(EnsembleConfig{})
	require.NoError(t, err)
	require.NotContains(t, string(payload), "aggregator_body_overrides")
}

// The admin write path has to run the validator, not just have one available.
// Without this the validation call could be deleted and every other test here
// would stay green, leaving a protected path to be discovered at fan-out time.
func TestAdminServiceUpdateEnsembleRefusesProtectedAggregatorOverride(t *testing.T) {
	group := &Group{ID: 7, Platform: PlatformEnsemble}
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{7: group}}
	svc := &adminServiceImpl{
		groupRepo: groupRepo,
		ensembleProposerRepo: &ensembleProposerRepoForAdminTest{members: []EnsembleProposer{{
			ID: 1, GroupID: 7, Role: EnsembleRoleProposer, Model: "gpt-5", Enabled: true,
		}}},
	}

	_, err := svc.UpdateEnsembleConfig(context.Background(), 7, EnsembleConfig{
		MinProposers:            1,
		AggregatorBodyOverrides: map[string]any{"messages": []any{}},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_PROTECTED")
	status, _ := infraerrors.ToHTTP(err)
	require.Equal(t, 400, status)
	// Validation runs before the repository write, so a refused config must not
	// have been persisted.
	require.Nil(t, groupRepo.updated)
}

func TestAdminServiceUpdateEnsemblePersistsAggregatorOverrides(t *testing.T) {
	group := &Group{ID: 7, Platform: PlatformEnsemble}
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{7: group}}
	svc := &adminServiceImpl{
		groupRepo: groupRepo,
		ensembleProposerRepo: &ensembleProposerRepoForAdminTest{members: []EnsembleProposer{{
			ID: 1, GroupID: 7, Role: EnsembleRoleProposer, Model: "gpt-5", Enabled: true,
		}}},
	}

	updated, err := svc.UpdateEnsembleConfig(context.Background(), 7, EnsembleConfig{
		MinProposers:            1,
		AggregatorBodyOverrides: map[string]any{"  reasoning_effort  ": "max"},
	})

	require.NoError(t, err)
	// Stored on the trimmed path: the runtime looks the path up verbatim, so an
	// untrimmed key would be saved and then never match.
	require.Equal(t, map[string]any{"reasoning_effort": "max"}, updated.AggregatorBodyOverrides)
	require.Equal(t, map[string]any{"reasoning_effort": "max"}, group.EnsembleConfig.AggregatorBodyOverrides)
	require.Same(t, group, groupRepo.updated)
}

func TestEnsembleConfigRoundTripsAggregatorOverrides(t *testing.T) {
	payload, err := json.Marshal(EnsembleConfig{
		AggregatorBodyOverrides: map[string]any{"reasoning_effort": "max", "thinking.type": "enabled"},
	})
	require.NoError(t, err)

	var restored EnsembleConfig
	require.NoError(t, json.Unmarshal(payload, &restored))
	require.Equal(t, map[string]any{"reasoning_effort": "max", "thinking.type": "enabled"}, restored.AggregatorBodyOverrides)
}
