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
}
