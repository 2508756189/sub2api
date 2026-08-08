package service

import (
	"context"
	"sort"
)

// Ensemble runtime defaults applied when the group config leaves a field zero.
const (
	defaultEnsembleTimeoutSeconds = 120
	defaultEnsembleMinProposers   = 1
)

// EnsemblePlan is the resolved execution plan for one ensemble request.
// Proposers are called in parallel; Aggregator (when non-nil and enabled)
// synthesizes the final answer from their outputs.
type EnsemblePlan struct {
	Proposers  []EnsembleProposer
	Aggregator *EnsembleProposer
	Config     EnsembleConfig
}

// EffectiveMinProposers is the number of proposer answers required for success.
func (p *EnsemblePlan) EffectiveMinProposers() int {
	minProposers := p.Config.MinProposers
	if minProposers <= 0 {
		minProposers = defaultEnsembleMinProposers
	}
	return minProposers
}

// EffectiveTimeoutSeconds bounds each individual sub-call.
func (p *EnsemblePlan) EffectiveTimeoutSeconds() int {
	if p.Config.TimeoutSeconds > 0 {
		if p.Config.TimeoutSeconds > MaxEnsembleTimeoutSeconds {
			return MaxEnsembleTimeoutSeconds
		}
		return p.Config.TimeoutSeconds
	}
	return defaultEnsembleTimeoutSeconds
}

// ShouldAggregate reports whether the aggregation step runs.
func (p *EnsemblePlan) ShouldAggregate() bool {
	return p.Config.AggregatorEnabled && p.Aggregator != nil
}

// EffectiveStreamTrace reports whether streaming clients receive the fan-out
// execution trace. It defaults to on: a group configured before the field
// existed stores nil, and an ensemble request is otherwise silent for as long as
// the slowest member takes.
func (p *EnsemblePlan) EffectiveStreamTrace() bool {
	if p.Config.StreamTrace == nil {
		return true
	}
	return *p.Config.StreamTrace
}

// EnsembleRuntimeService resolves the ensemble plan for a group on the request path.
type EnsembleRuntimeService struct {
	proposerRepo EnsembleProposerRepository
}

func NewEnsembleRuntimeService(proposerRepo EnsembleProposerRepository) *EnsembleRuntimeService {
	return &EnsembleRuntimeService{proposerRepo: proposerRepo}
}

// LoadPlan reads the enabled members of an ensemble group and pairs them with
// the group-level config. Members are ordered by priority then id so the
// proposal order shown to the aggregator is stable across requests.
func (s *EnsembleRuntimeService) LoadPlan(ctx context.Context, groupID int64, config EnsembleConfig) (*EnsemblePlan, error) {
	if s == nil || s.proposerRepo == nil {
		return nil, ErrEnsembleRuntimeUnavailable
	}

	members, err := s.proposerRepo.ListByGroup(ctx, groupID, false)
	if err != nil {
		return nil, err
	}

	plan := &EnsemblePlan{Config: config}
	for i := range members {
		member := members[i]
		if !member.Enabled {
			continue
		}
		if member.Role == EnsembleRoleAggregator {
			// Only one aggregator is meaningful; keep the highest-priority one
			// (lowest priority value) and ignore any extras.
			if plan.Aggregator == nil || member.Priority < plan.Aggregator.Priority {
				aggregator := member
				plan.Aggregator = &aggregator
			}
			continue
		}
		plan.Proposers = append(plan.Proposers, member)
	}

	sort.SliceStable(plan.Proposers, func(i, j int) bool {
		if plan.Proposers[i].Priority != plan.Proposers[j].Priority {
			return plan.Proposers[i].Priority < plan.Proposers[j].Priority
		}
		return plan.Proposers[i].ID < plan.Proposers[j].ID
	})

	return plan, nil
}
