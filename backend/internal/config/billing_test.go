package config

import "testing"

func TestBillingConfigCurrency(t *testing.T) {
	tests := []struct {
		name    string
		cfg     BillingConfig
		wantCur string
		wantFX  float64
		wantErr bool
	}{
		{name: "empty defaults to usd", cfg: BillingConfig{}, wantCur: BillingCurrencyUSD, wantFX: 1},
		{name: "usd ignores cny rate", cfg: BillingConfig{Currency: "usd", USDToCNYRate: 7.2}, wantCur: BillingCurrencyUSD, wantFX: 1},
		{name: "cny uses configured rate", cfg: BillingConfig{Currency: "cny", USDToCNYRate: 7.2}, wantCur: BillingCurrencyCNY, wantFX: 7.2},
		{name: "cny requires positive rate", cfg: BillingConfig{Currency: "CNY"}, wantCur: BillingCurrencyCNY, wantFX: 1, wantErr: true},
		{name: "unsupported currency", cfg: BillingConfig{Currency: "EUR", USDToCNYRate: 7.2}, wantCur: "EUR", wantFX: 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.CurrencyCode(); got != tt.wantCur {
				t.Fatalf("CurrencyCode()=%q, want %q", got, tt.wantCur)
			}
			if got := tt.cfg.USDExchangeRate(); got != tt.wantFX {
				t.Fatalf("USDExchangeRate()=%v, want %v", got, tt.wantFX)
			}
			if err := tt.cfg.ValidateCurrency(); (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCurrency() error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
