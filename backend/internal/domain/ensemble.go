package domain

// EnsembleConfig holds group-level options for an ensemble platform group.
// It is stored as a JSON column on the groups table and only meaningful when
// the group platform is "ensemble".
type EnsembleConfig struct {
	// SourceGroupIDs records the concrete groups whose accounts were copied
	// into this ensemble group. It is audit metadata only and is not used for
	// runtime cross-group routing.
	SourceGroupIDs []int64 `json:"source_group_ids"`
	// AggregatorEnabled controls whether proposer answers are synthesized by an
	// aggregator model. When false, the longest successful proposer answer is
	// returned as-is.
	AggregatorEnabled bool `json:"aggregator_enabled"`
	// MinProposers is the minimum number of proposers that must succeed for the
	// request to be considered successful. Zero is treated as 1.
	MinProposers int `json:"min_proposers"`
	// TimeoutSeconds bounds each individual sub-call. Zero uses the handler default.
	TimeoutSeconds int `json:"timeout_seconds"`
	// MaxTokens optionally caps each sub-call's completion length. Zero means no cap.
	MaxTokens int `json:"max_tokens"`
	// ExposeMetadata adds an ensemble_metadata object to the response describing
	// per-model timing, tokens and cost.
	ExposeMetadata bool `json:"expose_metadata"`
	// StreamTrace emits the fan-out execution steps to streaming clients as
	// reasoning_content deltas, so a caller can watch which members ran and why a
	// request failed instead of staring at a silent connection. Nil means enabled:
	// rows written before this field existed should still get the trace.
	StreamTrace *bool `json:"stream_trace,omitempty"`
	// AggregatorBodyOverrides sets request fields on the aggregator sub-call only,
	// as dotted paths (for example "reasoning_effort" or "thinking.type"). Empty
	// means the aggregator body is left alone and every member runs at whatever
	// depth its upstream defaults to.
	//
	// It exists because the ensemble has no reasoning-effort control of its own:
	// the group-level effort policy is keyed on an openai/composite group platform
	// and never fires for an ensemble key, and no client can express a depth for a
	// virtual model that appears in no model catalog. The aggregator is where the
	// setting earns its keep — proposers buy diversity, the aggregator is the
	// single judgement step, and it reads the largest prompt of the whole request.
	//
	// Values are stored in the member model's own spelling rather than a portable
	// level name. Providers disagree on both the field and the vocabulary (one
	// takes reasoning_effort with off/high/max, another only understands
	// thinking:{type:enabled}), so translating a shared level would need a
	// per-model table that silently goes stale as models are added.
	AggregatorBodyOverrides map[string]any `json:"aggregator_body_overrides,omitempty"`
}
