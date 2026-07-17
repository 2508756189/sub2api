package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type billingModeRepoStub struct {
	snapshot  BillingModeSnapshot
	saved     BillingModeSnapshot
	converted []BillingModeConversion
}

func (r *billingModeRepoStub) ReadBillingMode(context.Context) (BillingModeSnapshot, error) {
	return r.snapshot, nil
}
func (r *billingModeRepoStub) SaveBillingMode(_ context.Context, currency string, rate float64) error {
	r.saved = BillingModeSnapshot{Currency: currency, USDToCNYRate: rate, Configured: true}
	return nil
}
func (r *billingModeRepoStub) ConvertBillingMode(_ context.Context, conversion BillingModeConversion) (BillingModeConversionResult, error) {
	r.converted = append(r.converted, conversion)
	return BillingModeConversionResult{RowsConverted: 12, Tables: []string{"users"}}, nil
}

func TestBillingModeFactor(t *testing.T) {
	tests := []struct {
		name                       string
		from, to                   string
		fromRate, toRate, expected float64
	}{
		{name: "usd to cny", from: config.BillingCurrencyUSD, to: config.BillingCurrencyCNY, fromRate: 1, toRate: 7.2, expected: 7.2},
		{name: "cny to usd", from: config.BillingCurrencyCNY, to: config.BillingCurrencyUSD, fromRate: 7.2, toRate: 7.2, expected: 1.0 / 7.2},
		{name: "same currency", from: config.BillingCurrencyCNY, to: config.BillingCurrencyCNY, fromRate: 7.2, toRate: 8, expected: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := billingModeFactor(tt.from, tt.fromRate, tt.to, tt.toRate); got != tt.expected {
				t.Fatalf("factor=%v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBillingModeUpdateRequiresConfirmationAndConvertsExistingAmounts(t *testing.T) {
	repo := &billingModeRepoStub{snapshot: BillingModeSnapshot{Currency: config.BillingCurrencyCNY, USDToCNYRate: 7.2, Configured: true}}
	cfg := &config.Config{Billing: config.BillingConfig{Currency: config.BillingCurrencyCNY, USDToCNYRate: 7.2}}
	svc := NewBillingModeService(repo, cfg, nil)
	if err := svc.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Update(context.Background(), config.BillingCurrencyUSD, 7.2, true, ""); err != ErrBillingModeConfirmationRequired {
		t.Fatalf("want confirmation error, got %v", err)
	}
	result, err := svc.Update(context.Background(), config.BillingCurrencyUSD, 7.2, true, BillingModeConfirmationUSD)
	if err != nil {
		t.Fatal(err)
	}
	if result.RowsConverted != 12 || len(repo.converted) != 1 {
		t.Fatalf("unexpected conversion result: %+v %+v", result, repo.converted)
	}
	if got := svc.Current().CurrencyCode(); got != config.BillingCurrencyUSD {
		t.Fatalf("currency=%s, want USD", got)
	}
}

func TestBillingModeRateChangeDoesNotConvertHistory(t *testing.T) {
	repo := &billingModeRepoStub{snapshot: BillingModeSnapshot{Currency: config.BillingCurrencyCNY, USDToCNYRate: 7.2, Configured: true}}
	svc := NewBillingModeService(repo, &config.Config{Billing: config.BillingConfig{Currency: config.BillingCurrencyCNY, USDToCNYRate: 7.2}}, nil)
	if err := svc.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Update(context.Background(), config.BillingCurrencyCNY, 7.3, false, ""); err != nil {
		t.Fatal(err)
	}
	if len(repo.converted) != 0 || repo.saved.USDToCNYRate != 7.3 {
		t.Fatalf("rate update unexpectedly converted history: %+v %+v", repo.saved, repo.converted)
	}
}

func TestBillingModeConversionFlushesInjectedCache(t *testing.T) {
	repo := &billingModeRepoStub{snapshot: BillingModeSnapshot{Currency: config.BillingCurrencyCNY, USDToCNYRate: 7.2, Configured: true}}
	flushes := 0
	svc := NewBillingModeService(
		repo,
		&config.Config{Billing: config.BillingConfig{Currency: config.BillingCurrencyCNY, USDToCNYRate: 7.2}},
		func(context.Context) error {
			flushes++
			return nil
		},
	)
	if err := svc.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Update(context.Background(), config.BillingCurrencyUSD, 7.2, true, BillingModeConfirmationUSD); err != nil {
		t.Fatal(err)
	}
	if flushes != 1 {
		t.Fatalf("cache flushes=%d, want 1", flushes)
	}
}
