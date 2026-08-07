//go:build unit

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsembleConfigPreservesSourceGroupIDsInJSON(t *testing.T) {
	config := EnsembleConfig{SourceGroupIDs: []int64{10, 20, 10}}

	payload, err := json.Marshal(config)
	require.NoError(t, err)
	require.JSONEq(t, `{"aggregator_enabled":false,"expose_metadata":false,"max_tokens":0,"min_proposers":0,"source_group_ids":[10,20,10],"timeout_seconds":0}`, string(payload))

	var decoded EnsembleConfig
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.Equal(t, []int64{10, 20, 10}, decoded.SourceGroupIDs)
}

type ensembleProposerRepoForAdminTest struct {
	created *EnsembleProposer
	updated *EnsembleProposer
	deleted int64
	members []EnsembleProposer
}

func (s *ensembleProposerRepoForAdminTest) ListByGroup(context.Context, int64, bool) ([]EnsembleProposer, error) {
	return append([]EnsembleProposer(nil), s.members...), nil
}

func (s *ensembleProposerRepoForAdminTest) Create(_ context.Context, proposer *EnsembleProposer) error {
	s.created = proposer
	return nil
}

func (s *ensembleProposerRepoForAdminTest) Update(_ context.Context, proposer *EnsembleProposer) error {
	s.updated = proposer
	return nil
}
func (s *ensembleProposerRepoForAdminTest) Delete(_ context.Context, proposerID int64) error {
	s.deleted = proposerID
	return nil
}
func (s *ensembleProposerRepoForAdminTest) DeleteByGroup(context.Context, int64) error { return nil }

func TestAdminServiceRejectsEnsembleModelUnavailableInBoundAccounts(t *testing.T) {
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{
		7: {ID: 7, Platform: PlatformEnsemble},
	}}
	accountRepo := &accountRepoStubForCompositeModelsList{accounts: []Account{
		{ID: 1, Platform: PlatformOpenAI, Credentials: map[string]any{
			"model_mapping": map[string]any{"gpt-allowed": "gpt-5"},
		}},
	}}
	proposerRepo := &ensembleProposerRepoForAdminTest{}
	svc := &adminServiceImpl{groupRepo: groupRepo, accountRepo: accountRepo, ensembleProposerRepo: proposerRepo}

	_, err := svc.CreateEnsembleProposer(context.Background(), 7, EnsembleProposerInput{
		Model:   "gpt-not-bound",
		Enabled: true,
	})

	require.Error(t, err)
	require.Nil(t, proposerRepo.created)
}

func TestAdminServiceAcceptsEnsembleModelMappedByBoundAccount(t *testing.T) {
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{
		7: {ID: 7, Platform: PlatformEnsemble},
	}}
	accountRepo := &accountRepoStubForCompositeModelsList{accounts: []Account{
		{ID: 1, Platform: PlatformOpenAI, Credentials: map[string]any{
			"model_mapping": map[string]any{"gpt-allowed": "gpt-5"},
		}},
	}}
	proposerRepo := &ensembleProposerRepoForAdminTest{}
	svc := &adminServiceImpl{groupRepo: groupRepo, accountRepo: accountRepo, ensembleProposerRepo: proposerRepo}

	created, err := svc.CreateEnsembleProposer(context.Background(), 7, EnsembleProposerInput{
		Model:    "gpt-allowed",
		Platform: PlatformOpenAI,
		Priority: 10,
		Enabled:  true,
	})

	require.NoError(t, err)
	require.Equal(t, "gpt-allowed", created.Model)
	require.Same(t, created, proposerRepo.created)
}

func TestAdminServiceLimitsProposersAndAggregators(t *testing.T) {
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{
		7: {ID: 7, Platform: PlatformEnsemble},
	}}
	accountRepo := &accountRepoStubForCompositeModelsList{accounts: []Account{{
		ID: 1, Platform: PlatformOpenAI, Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-1": "gpt-1", "gpt-2": "gpt-2", "gpt-3": "gpt-3",
				"gpt-4": "gpt-4", "gpt-5": "gpt-5", "gpt-6": "gpt-6",
			},
		},
	}}}
	proposers := make([]EnsembleProposer, 0, MaxEnsembleProposers)
	for i := 0; i < MaxEnsembleProposers; i++ {
		proposers = append(proposers, EnsembleProposer{ID: int64(i + 1), GroupID: 7, Role: EnsembleRoleProposer, Model: fmt.Sprintf("gpt-%d", i+1), Enabled: true})
	}
	proposerRepo := &ensembleProposerRepoForAdminTest{members: proposers}
	svc := &adminServiceImpl{groupRepo: groupRepo, accountRepo: accountRepo, ensembleProposerRepo: proposerRepo}

	_, err := svc.CreateEnsembleProposer(context.Background(), 7, EnsembleProposerInput{Model: "gpt-1", Enabled: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot exceed")

	proposerRepo.members = append(proposerRepo.members, EnsembleProposer{ID: 99, GroupID: 7, Role: EnsembleRoleAggregator, Model: "gpt-1", Enabled: true})
	_, err = svc.CreateEnsembleProposer(context.Background(), 7, EnsembleProposerInput{Role: EnsembleRoleAggregator, Model: "gpt-2", Enabled: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "only one aggregator")
}

func TestAdminServiceEnsembleCandidatesComeFromBoundConcreteAccounts(t *testing.T) {
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{
		7: {ID: 7, Platform: PlatformEnsemble},
	}}
	accountRepo := &accountRepoStubForCompositeModelsList{accounts: []Account{
		{ID: 1, Platform: PlatformOpenAI, Credentials: map[string]any{
			"model_mapping": map[string]any{"gpt-allowed": "gpt-5"},
		}},
	}}
	svc := &adminServiceImpl{groupRepo: groupRepo, accountRepo: accountRepo}

	candidates, err := svc.GetGroupModelsListCandidates(context.Background(), 7, PlatformEnsemble)

	require.NoError(t, err)
	require.Contains(t, candidates, "gpt-allowed")
	require.NotContains(t, candidates, "ensemble")
}

func TestAdminServiceRejectsMinimumAboveEnabledProposers(t *testing.T) {
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{
		7: {ID: 7, Platform: PlatformEnsemble},
	}}
	proposerRepo := &ensembleProposerRepoForAdminTest{members: []EnsembleProposer{
		{ID: 1, GroupID: 7, Role: EnsembleRoleProposer, Model: "gpt-5", Enabled: true},
		{ID: 2, GroupID: 7, Role: EnsembleRoleProposer, Model: "gpt-5.1", Enabled: true},
	}}
	svc := &adminServiceImpl{groupRepo: groupRepo, ensembleProposerRepo: proposerRepo}

	_, err := svc.UpdateEnsembleConfig(context.Background(), 7, EnsembleConfig{MinProposers: 3})

	require.Error(t, err)
}

func TestAdminServiceAllowsConfigBeforeFreshGroupHasMembers(t *testing.T) {
	group := &Group{ID: 7, Platform: PlatformEnsemble}
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{7: group}}
	svc := &adminServiceImpl{
		groupRepo:            groupRepo,
		ensembleProposerRepo: &ensembleProposerRepoForAdminTest{},
	}

	updated, err := svc.UpdateEnsembleConfig(context.Background(), 7, EnsembleConfig{MinProposers: 1})

	require.NoError(t, err)
	require.Equal(t, 1, updated.MinProposers)
	require.Same(t, group, groupRepo.updated)
}

func TestAdminServiceClampsMinimumBeforeDisablingOrDeletingProposer(t *testing.T) {
	for _, operation := range []string{"disable", "delete"} {
		t.Run(operation, func(t *testing.T) {
			group := &Group{ID: 7, Platform: PlatformEnsemble, EnsembleConfig: EnsembleConfig{MinProposers: 2}}
			groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{7: group}}
			proposerRepo := &ensembleProposerRepoForAdminTest{members: []EnsembleProposer{
				{ID: 1, GroupID: 7, Role: EnsembleRoleProposer, Model: "gpt-5", Platform: PlatformOpenAI, Enabled: true},
				{ID: 2, GroupID: 7, Role: EnsembleRoleProposer, Model: "gpt-5.1", Platform: PlatformOpenAI, Enabled: true},
			}}
			svc := &adminServiceImpl{groupRepo: groupRepo, accountRepo: &accountRepoStubForCompositeModelsList{accounts: []Account{{
				ID: 1, Platform: PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-5": "gpt-5", "gpt-5.1": "gpt-5.1"}},
			}}}, ensembleProposerRepo: proposerRepo}

			if operation == "disable" {
				_, err := svc.UpdateEnsembleProposer(context.Background(), 7, 2, EnsembleProposerInput{
					Role: EnsembleRoleProposer, Model: "gpt-5.1", Platform: PlatformOpenAI, Enabled: false,
				})
				require.NoError(t, err)
			} else {
				require.NoError(t, svc.DeleteEnsembleProposer(context.Background(), 7, 2))
			}

			require.Equal(t, 1, group.EnsembleConfig.MinProposers)
			require.Same(t, group, groupRepo.updated)
		})
	}
}

func TestAdminServicePersistsEnsembleSourceGroupIDs(t *testing.T) {
	group := &Group{ID: 7, Platform: PlatformEnsemble}
	groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{7: group}}
	svc := &adminServiceImpl{
		groupRepo: groupRepo,
		ensembleProposerRepo: &ensembleProposerRepoForAdminTest{members: []EnsembleProposer{{
			ID: 1, GroupID: 7, Role: EnsembleRoleProposer, Model: "gpt-5", Enabled: true,
		}}},
	}

	updated, err := svc.UpdateEnsembleConfig(context.Background(), 7, EnsembleConfig{
		SourceGroupIDs: []int64{10, 20},
		MinProposers:   1,
	})

	require.NoError(t, err)
	require.Equal(t, []int64{10, 20}, updated.SourceGroupIDs)
	require.Equal(t, []int64{10, 20}, group.EnsembleConfig.SourceGroupIDs)
}

func TestCanCopyAccountsFromGroupPlatformAllowsConcreteSourcesForEnsemble(t *testing.T) {
	for _, source := range []string{
		PlatformAnthropic,
		PlatformOpenAI,
		PlatformGemini,
		PlatformAntigravity,
		PlatformGrok,
	} {
		require.True(t, canCopyAccountsFromGroupPlatform(PlatformEnsemble, source), "source=%s", source)
	}
	require.False(t, canCopyAccountsFromGroupPlatform(PlatformEnsemble, PlatformComposite))
	require.False(t, canCopyAccountsFromGroupPlatform(PlatformEnsemble, PlatformEnsemble))
}

func TestAdminServiceCreateEnsembleCopiesConcreteSourceGroups(t *testing.T) {
	var copiedFrom []int64
	var boundAccounts []int64
	groupRepo := &groupRepoStubForAdmin{
		createID: 99,
		getByIDByID: map[int64]*Group{
			10: {ID: 10, Platform: PlatformOpenAI},
			20: {ID: 20, Platform: PlatformAnthropic},
		},
		getAccountIDsByGroupIDsFn: func(groupIDs []int64) ([]int64, error) {
			copiedFrom = append([]int64(nil), groupIDs...)
			return []int64{101, 202}, nil
		},
		bindAccountsToGroupFn: func(_ int64, accountIDs []int64) error {
			boundAccounts = append([]int64(nil), accountIDs...)
			return nil
		},
	}
	svc := &adminServiceImpl{groupRepo: groupRepo}

	group, err := svc.CreateGroup(context.Background(), &CreateGroupInput{
		Name:                     "Ensemble",
		Platform:                 PlatformEnsemble,
		RateMultiplier:           1,
		CopyAccountsFromGroupIDs: []int64{10, 20, 10},
	})

	require.NoError(t, err)
	require.Equal(t, PlatformEnsemble, group.Platform)
	require.ElementsMatch(t, []int64{10, 20}, copiedFrom)
	require.ElementsMatch(t, []int64{101, 202}, boundAccounts)
}

func TestAdminServiceUpdateEnsembleRejectsCompositeOrEnsembleSourceGroups(t *testing.T) {
	for _, sourcePlatform := range []string{PlatformComposite, PlatformEnsemble} {
		t.Run(sourcePlatform, func(t *testing.T) {
			groupRepo := &groupRepoStubForAdmin{getByIDByID: map[int64]*Group{
				7:  {ID: 7, Platform: PlatformEnsemble, RateMultiplier: 1, SubscriptionType: SubscriptionTypeStandard},
				10: {ID: 10, Platform: sourcePlatform},
			}}
			svc := &adminServiceImpl{groupRepo: groupRepo}

			_, err := svc.UpdateGroup(context.Background(), 7, &UpdateGroupInput{
				CopyAccountsFromGroupIDs: []int64{10},
			})

			require.Error(t, err)
		})
	}
}
