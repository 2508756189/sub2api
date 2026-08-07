//go:build unit

package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// TestBuildUsageBillingCommand_SubscriptionAppliesRateMultiplier locks in the fix
// that subscription-mode billing honours the group (and any user-specific) rate
// multiplier — i.e. cmd.SubscriptionCost tracks ActualCost (= TotalCost *
// RateMultiplier), not raw TotalCost.
func TestBuildUsageBillingCommand_SubscriptionAppliesRateMultiplier(t *testing.T) {
	t.Parallel()

	groupID := int64(7)
	subID := int64(42)

	tests := []struct {
		name           string
		totalCost      float64
		actualCost     float64
		isSubscription bool
		wantSub        float64
		wantBalance    float64
	}{
		{
			name:           "subscription with 2x multiplier consumes 2x quota",
			totalCost:      1.0,
			actualCost:     2.0,
			isSubscription: true,
			wantSub:        2.0,
			wantBalance:    0,
		},
		{
			name:           "subscription with 0.5x multiplier consumes 0.5x quota",
			totalCost:      1.0,
			actualCost:     0.5,
			isSubscription: true,
			wantSub:        0.5,
			wantBalance:    0,
		},
		{
			name:           "free subscription (multiplier 0) consumes no quota",
			totalCost:      1.0,
			actualCost:     0,
			isSubscription: true,
			wantSub:        0,
			wantBalance:    0,
		},
		{
			name:           "balance billing keeps using ActualCost (regression)",
			totalCost:      1.0,
			actualCost:     2.0,
			isSubscription: false,
			wantSub:        0,
			wantBalance:    2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &postUsageBillingParams{
				Cost:               &CostBreakdown{TotalCost: tt.totalCost, ActualCost: tt.actualCost},
				User:               &User{ID: 1},
				APIKey:             &APIKey{ID: 2, GroupID: &groupID},
				Account:            &Account{ID: 3},
				Subscription:       &UserSubscription{ID: subID},
				IsSubscriptionBill: tt.isSubscription,
			}

			cmd := buildUsageBillingCommand("req-1", nil, p)
			if cmd == nil {
				t.Fatal("buildUsageBillingCommand returned nil")
			}
			if cmd.SubscriptionCost != tt.wantSub {
				t.Errorf("SubscriptionCost = %v, want %v", cmd.SubscriptionCost, tt.wantSub)
			}
			if cmd.BalanceCost != tt.wantBalance {
				t.Errorf("BalanceCost = %v, want %v", cmd.BalanceCost, tt.wantBalance)
			}
		})
	}
}

// TokenPort may settle usage in CNY while existing installations and plans keep
// the historical USD mode. Quantization belongs after the billing service has
// converted the final settlement amount, so both modes reach NUMERIC(20,8)
// without changing the currency contract.
func TestBuildUsageBillingCommand_QuantizesFinalSettlementCurrency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		billing  config.BillingConfig
		currency string
		exchange float64
	}{
		{
			name:     "legacy USD settlement remains USD",
			billing:  config.BillingConfig{Currency: config.BillingCurrencyUSD, USDToCNYRate: 7.2},
			currency: config.BillingCurrencyUSD,
			exchange: 1,
		},
		{
			name:     "TokenPort CNY settlement quantizes after conversion",
			billing:  config.BillingConfig{Currency: config.BillingCurrencyCNY, USDToCNYRate: 7.2},
			currency: config.BillingCurrencyCNY,
			exchange: 7.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			billingService := NewBillingService(&config.Config{Billing: tt.billing}, nil)
			cost, err := billingService.CalculateCost(
				"claude-sonnet-4",
				UsageTokens{InputTokens: 7, OutputTokens: 3},
				1.234567,
			)
			require.NoError(t, err)
			require.Equal(t, tt.currency, cost.Currency)
			require.InDelta(t, cost.SourceActualCostUSD*tt.exchange, cost.ActualCost, 1e-12)

			subscriptionID := int64(42)
			cmd := buildUsageBillingCommand("req-currency-compat", nil, &postUsageBillingParams{
				Cost:               cost,
				User:               &User{ID: 1},
				APIKey:             &APIKey{ID: 2},
				Account:            &Account{ID: 3},
				Subscription:       &UserSubscription{ID: subscriptionID},
				IsSubscriptionBill: true,
			})

			require.NotNil(t, cmd)
			require.Equal(t, QuantizeUsageBillingAmount(cost.ActualCost), cmd.SubscriptionCost)
			if tt.currency == config.BillingCurrencyCNY {
				require.NotEqual(t, QuantizeUsageBillingAmount(cost.SourceActualCostUSD), cmd.SubscriptionCost,
					"CNY settlement must not quantize the pre-conversion USD amount")
			}
		})
	}
}
