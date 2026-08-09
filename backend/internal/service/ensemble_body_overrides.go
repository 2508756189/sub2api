package service

import (
	"encoding/json"
	"fmt"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	// MaxEnsembleAggregatorBodyOverrides bounds how many paths a group may set.
	// The field exists to carry a handful of provider knobs (a reasoning effort,
	// a thinking toggle), not to become a second request body.
	MaxEnsembleAggregatorBodyOverrides = 16
	// maxEnsembleBodyOverridePathLen bounds one dotted path.
	maxEnsembleBodyOverridePathLen = 128
	// maxEnsembleBodyOverridesBytes bounds the encoded values in total, so a
	// stored config cannot inflate every aggregator request.
	maxEnsembleBodyOverridesBytes = 4096
)

// ensembleProtectedBodyPaths are request fields an aggregator body override must
// never touch.
//
// The first six carry fan-out invariants: model and stream are rewritten per
// sub-call, stream_options.include_usage is what lets billing see usage at all,
// messages is the conversation the aggregator has to judge, and tools/tool_choice
// drive the tool-call path whose id repair runs after normalization.
//
// The two token limits are excluded for a different reason: applyEnsembleMaxTokens
// already owns them under three constraints (a ceiling rather than an override,
// written on the spelling the client actually used, never cutting below a declared
// thinking budget). A raw override would bypass all three silently, and the third
// one turns every member into invalid_request_error when violated.
var ensembleProtectedBodyPaths = map[string]struct{}{
	"model":                 {},
	"stream":                {},
	"stream_options":        {},
	"messages":              {},
	"tools":                 {},
	"tool_choice":           {},
	"max_tokens":            {},
	"max_completion_tokens": {},
}

// ensembleBodyOverridePathReserved lists characters with special meaning to
// gjson/sjson paths. A legitimate knob ("reasoning_effort", "thinking.type",
// "extra_body.thinking.budget_tokens") contains none of them, while a path that
// does would write somewhere the admin did not name.
const ensembleBodyOverridePathReserved = "*?#|@\\"

// EnsembleBodyOverridePathAllowed reports whether a dotted override path may be
// written into a sub-call body. The decision is made on the first segment, so
// "messages.0.content" is refused along with "messages".
func EnsembleBodyOverridePathAllowed(path string) bool {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return false
	}
	root := trimmed
	if index := strings.IndexByte(root, '.'); index >= 0 {
		root = root[:index]
	}
	_, protected := ensembleProtectedBodyPaths[root]
	return !protected
}

// ValidateEnsembleAggregatorBodyOverrides checks a group's aggregator overrides
// at write time.
//
// Rejecting here rather than at request time is deliberate: an admin who typed a
// protected path should be told, whereas a fan-out that fails hours later over a
// saved typo gives them nothing to act on. Validation covers the shape of the
// path only. It deliberately does not check the value against a per-model effort
// vocabulary — that table would need an entry per model and would silently go
// stale, which is the same failure mode that made per-model capability marking a
// bad trade here.
func ValidateEnsembleAggregatorBodyOverrides(overrides map[string]any) error {
	if len(overrides) == 0 {
		return nil
	}
	if len(overrides) > MaxEnsembleAggregatorBodyOverrides {
		return infraerrors.BadRequest(
			"ENSEMBLE_AGGREGATOR_OVERRIDES_TOO_MANY",
			fmt.Sprintf("aggregator body overrides cannot exceed %d entries", MaxEnsembleAggregatorBodyOverrides),
		)
	}

	encodedBytes := 0
	for path, value := range overrides {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			return infraerrors.BadRequest(
				"ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_EMPTY",
				"aggregator body override path cannot be empty",
			)
		}
		if len(trimmed) > maxEnsembleBodyOverridePathLen {
			return infraerrors.BadRequest(
				"ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_TOO_LONG",
				fmt.Sprintf("aggregator body override path %q is too long", trimmed),
			)
		}
		if strings.ContainsAny(trimmed, ensembleBodyOverridePathReserved) {
			return infraerrors.BadRequest(
				"ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_INVALID",
				fmt.Sprintf("aggregator body override path %q contains a reserved character", trimmed),
			)
		}
		if !EnsembleBodyOverridePathAllowed(trimmed) {
			return infraerrors.BadRequest(
				"ENSEMBLE_AGGREGATOR_OVERRIDE_PATH_PROTECTED",
				fmt.Sprintf("aggregator body override path %q is managed by the ensemble runtime and cannot be overridden", trimmed),
			)
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return infraerrors.BadRequest(
				"ENSEMBLE_AGGREGATOR_OVERRIDE_VALUE_INVALID",
				fmt.Sprintf("aggregator body override %q has a value that cannot be encoded as JSON", trimmed),
			)
		}
		encodedBytes += len(encoded)
	}
	if encodedBytes > maxEnsembleBodyOverridesBytes {
		return infraerrors.BadRequest(
			"ENSEMBLE_AGGREGATOR_OVERRIDES_TOO_LARGE",
			fmt.Sprintf("aggregator body overrides exceed %d encoded bytes", maxEnsembleBodyOverridesBytes),
		)
	}
	return nil
}

// NormalizeEnsembleAggregatorBodyOverrides trims paths and drops empty ones so a
// stored config does not carry keys that could never be applied. It is shape-only
// cleanup; validation is what refuses a bad path.
func NormalizeEnsembleAggregatorBodyOverrides(overrides map[string]any) map[string]any {
	if len(overrides) == 0 {
		return nil
	}
	normalized := make(map[string]any, len(overrides))
	for path, value := range overrides {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		normalized[trimmed] = value
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}
