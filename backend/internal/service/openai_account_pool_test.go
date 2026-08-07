//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestListSchedulableAccountsMergesPoolGroups(t *testing.T) {
	repo := &groupAwareMockAccountRepo{
		mockAccountRepoForPlatform: &mockAccountRepoForPlatform{},
		allAccounts: []Account{
			{ID: 100, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, AccountGroups: []AccountGroup{{GroupID: 7}}},
			{ID: 200, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, AccountGroups: []AccountGroup{{GroupID: 42}}},
			{ID: 300, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, AccountGroups: []AccountGroup{{GroupID: 42}}},
			// 100 is bound to both group 7 and group 43: the pool merge must dedupe it.
			{ID: 400, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, AccountGroups: []AccountGroup{{GroupID: 43}}},
		},
	}
	svc := &OpenAIGatewayService{accountRepo: repo, cfg: &config.Config{}}

	ctx := WithAccountPoolGroupIDs(context.Background(), []int64{7, 42, 43})
	accounts, err := svc.listSchedulableAccounts(ctx, int64Ptr(7), PlatformOpenAI)
	require.NoError(t, err)

	ids := make(map[int64]bool, len(accounts))
	for _, account := range accounts {
		ids[account.ID] = true
	}
	require.Equal(t, map[int64]bool{100: true, 200: true, 300: true, 400: true}, ids,
		"pool must be the union of all pool groups with duplicates removed")
}

func TestListSchedulableAccountsIgnoresPoolWhenUnset(t *testing.T) {
	repo := &groupAwareMockAccountRepo{
		mockAccountRepoForPlatform: &mockAccountRepoForPlatform{},
		allAccounts: []Account{
			{ID: 100, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, AccountGroups: []AccountGroup{{GroupID: 7}}},
			{ID: 200, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, AccountGroups: []AccountGroup{{GroupID: 42}}},
		},
	}
	svc := &OpenAIGatewayService{accountRepo: repo, cfg: &config.Config{}}

	// No pool set: only the caller's own group is queried.
	accounts, err := svc.listSchedulableAccounts(context.Background(), int64Ptr(7), PlatformOpenAI)
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	require.Equal(t, int64(100), accounts[0].ID)
}

func TestStickyAccountMatchesPoolGroupIDs(t *testing.T) {
	account := &Account{ID: 1, AccountGroups: []AccountGroup{{GroupID: 42}}}
	require.True(t, openAIStickyAccountMatchesPoolGroupIDs(account, []int64{7, 42, 43}),
		"account bound to a pool group must match")
	require.False(t, openAIStickyAccountMatchesPoolGroupIDs(account, []int64{7, 43}),
		"account bound outside the pool must not match")
	require.False(t, openAIStickyAccountMatchesPoolGroupIDs(account, nil),
		"empty pool never matches")
}
