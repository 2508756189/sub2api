package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type ensembleRuntimeRepoStub struct {
	requestedGroupID int64
	members          []EnsembleProposer
}

func (s *ensembleRuntimeRepoStub) ListByGroup(_ context.Context, groupID int64, includeDisabled bool) ([]EnsembleProposer, error) {
	s.requestedGroupID = groupID
	out := make([]EnsembleProposer, 0, len(s.members))
	for _, member := range s.members {
		if member.GroupID != groupID {
			continue
		}
		if !includeDisabled && !member.Enabled {
			continue
		}
		out = append(out, member)
	}
	return out, nil
}

func (s *ensembleRuntimeRepoStub) Create(context.Context, *EnsembleProposer) error { return nil }
func (s *ensembleRuntimeRepoStub) Update(context.Context, *EnsembleProposer) error { return nil }
func (s *ensembleRuntimeRepoStub) Delete(context.Context, int64) error             { return nil }

func TestEnsembleRuntimePlanOrdersMembersAndPreservesConfiguredMinimum(t *testing.T) {
	repo := &ensembleRuntimeRepoStub{members: []EnsembleProposer{
		{ID: 2, GroupID: 7, Role: EnsembleRoleProposer, Model: "late", Priority: 20, Enabled: true},
		{ID: 1, GroupID: 7, Role: EnsembleRoleProposer, Model: "early", Priority: 10, Enabled: true},
		{ID: 3, GroupID: 7, Role: EnsembleRoleProposer, Model: "disabled", Priority: 1, Enabled: false},
		{ID: 4, GroupID: 7, Role: EnsembleRoleAggregator, Model: "aggregator-late", Priority: 20, Enabled: true},
		{ID: 5, GroupID: 7, Role: EnsembleRoleAggregator, Model: "aggregator-early", Priority: 10, Enabled: true},
		{ID: 6, GroupID: 99, Role: EnsembleRoleProposer, Model: "other-group", Priority: 1, Enabled: true},
	}}

	plan, err := NewEnsembleRuntimeService(repo).LoadPlan(context.Background(), 7, EnsembleConfig{
		AggregatorEnabled: true,
		MinProposers:      5,
	})

	require.NoError(t, err)
	require.Equal(t, int64(7), repo.requestedGroupID)
	require.Equal(t, []string{"early", "late"}, []string{plan.Proposers[0].Model, plan.Proposers[1].Model})
	require.Equal(t, "aggregator-early", plan.Aggregator.Model)
	require.Equal(t, 5, plan.EffectiveMinProposers(), "configured minimum must not be silently reduced")
}

func TestEnsembleRuntimeCapsConfiguredTimeout(t *testing.T) {
	plan := &EnsemblePlan{Config: EnsembleConfig{TimeoutSeconds: MaxEnsembleTimeoutSeconds + 1}}
	require.Equal(t, MaxEnsembleTimeoutSeconds, plan.EffectiveTimeoutSeconds())
}
