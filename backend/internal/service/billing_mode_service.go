package service

import (
	"context"
	"math"
	"strings"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SettingKeyBillingCurrency        = "billing_currency"
	SettingKeyBillingUSDToCNYRate    = "billing_usd_to_cny_rate"
	SettingKeyBillingCurrencyHistory = "billing_currency_history"
	BillingModeConfirmationUSD       = "SWITCH TO USD"
	BillingModeConfirmationCNY       = "SWITCH TO CNY"
)

var (
	ErrBillingModeInvalidCurrency      = infraerrors.BadRequest("BILLING_MODE_INVALID_CURRENCY", "billing currency must be USD or CNY")
	ErrBillingModeInvalidRate          = infraerrors.BadRequest("BILLING_MODE_INVALID_RATE", "USD to CNY rate must be positive")
	ErrBillingModeConfirmationRequired = infraerrors.BadRequest("BILLING_MODE_CONFIRMATION_REQUIRED", "explicit confirmation is required before converting existing platform amounts")
	ErrBillingModeActiveJobs           = infraerrors.Conflict("BILLING_MODE_ACTIVE_JOBS", "wait for active batch image jobs to finish before changing billing currency")
)

type BillingModeSnapshot struct {
	Currency     string
	USDToCNYRate float64
	Configured   bool
}

type BillingModeConversion struct {
	FromCurrency string
	FromRate     float64
	ToCurrency   string
	ToRate       float64
	Factor       float64
	HistoryEntry map[string]any
}

type BillingModeConversionResult struct {
	RowsConverted int64
	Tables        []string
}

// BillingModeRepository owns the database transaction that changes the
// platform settlement currency. Keeping it behind a service interface makes
// the conversion testable without coupling billing code to Ent internals.
type BillingModeRepository interface {
	ReadBillingMode(ctx context.Context) (BillingModeSnapshot, error)
	SaveBillingMode(ctx context.Context, currency string, usdToCNYRate float64) error
	ConvertBillingMode(ctx context.Context, conversion BillingModeConversion) (BillingModeConversionResult, error)
}

type BillingModeService struct {
	repo       BillingModeRepository
	cacheFlush func(context.Context) error
	fallback   config.BillingConfig

	mu      sync.RWMutex
	current config.BillingConfig
}

func NewBillingModeService(repo BillingModeRepository, cfg *config.Config, cacheFlush func(context.Context) error) *BillingModeService {
	var fallback config.BillingConfig
	if cfg != nil {
		fallback = cfg.Billing
	}
	if fallback.CurrencyCode() != config.BillingCurrencyCNY && fallback.CurrencyCode() != config.BillingCurrencyUSD {
		fallback.Currency = config.BillingCurrencyUSD
	}
	if fallback.USDToCNYRate <= 0 {
		fallback.USDToCNYRate = 7.2
	}
	return &BillingModeService{repo: repo, cacheFlush: cacheFlush, fallback: fallback, current: fallback}
}

func (s *BillingModeService) Load(ctx context.Context) error {
	if s == nil || s.repo == nil {
		return nil
	}
	snapshot, err := s.repo.ReadBillingMode(ctx)
	if err != nil {
		return err
	}
	if !snapshot.Configured {
		return nil
	}
	if snapshot.Currency != config.BillingCurrencyUSD && snapshot.Currency != config.BillingCurrencyCNY {
		return ErrBillingModeInvalidCurrency
	}
	if snapshot.USDToCNYRate <= 0 {
		snapshot.USDToCNYRate = s.fallback.USDToCNYRate
	}
	s.mu.Lock()
	s.current = s.withCurrency(snapshot.Currency, snapshot.USDToCNYRate)
	s.mu.Unlock()
	return nil
}

func (s *BillingModeService) withCurrency(currency string, rate float64) config.BillingConfig {
	billing := s.fallback
	billing.Currency = currency
	billing.USDToCNYRate = rate
	return billing
}

func billingConfigAfterConversion(current config.BillingConfig, currency string, rate, factor float64) config.BillingConfig {
	next := current
	next.Currency = currency
	next.USDToCNYRate = rate
	if next.MinimumBalanceReserve > 0 {
		next.MinimumBalanceReserve *= factor
	}
	return next
}

func (s *BillingModeService) Current() config.BillingConfig {
	if s == nil {
		return config.BillingConfig{Currency: config.BillingCurrencyUSD, USDToCNYRate: 1}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

func (s *BillingModeService) SetCurrentResolver() func() config.BillingConfig {
	return func() config.BillingConfig {
		if s == nil {
			return config.BillingConfig{Currency: config.BillingCurrencyUSD, USDToCNYRate: 1}
		}
		return s.Current()
	}
}

func billingModeFactor(fromCurrency string, fromRate float64, toCurrency string, toRate float64) float64 {
	if fromCurrency == toCurrency {
		return 1
	}
	if toCurrency == config.BillingCurrencyCNY {
		return toRate
	}
	return 1 / fromRate
}

func (s *BillingModeService) Update(ctx context.Context, currency string, usdToCNYRate float64, convertExisting bool, confirmation string) (BillingModeConversionResult, error) {
	if s == nil || s.repo == nil {
		return BillingModeConversionResult{}, infraerrors.InternalServer("BILLING_MODE_UNAVAILABLE", "billing mode service is not configured")
	}
	target := config.NormalizeBillingCurrency(currency)
	if target != config.BillingCurrencyUSD && target != config.BillingCurrencyCNY {
		return BillingModeConversionResult{}, ErrBillingModeInvalidCurrency
	}
	current := s.Current()
	if usdToCNYRate <= 0 {
		usdToCNYRate = current.USDToCNYRate
		if usdToCNYRate <= 0 {
			usdToCNYRate = s.fallback.USDToCNYRate
		}
	}
	if usdToCNYRate <= 0 || math.IsNaN(usdToCNYRate) || math.IsInf(usdToCNYRate, 0) {
		return BillingModeConversionResult{}, ErrBillingModeInvalidRate
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	current = s.current
	if target != current.CurrencyCode() {
		if !convertExisting {
			return BillingModeConversionResult{}, ErrBillingModeConfirmationRequired
		}
		expected := BillingModeConfirmationUSD
		if target == config.BillingCurrencyCNY {
			expected = BillingModeConfirmationCNY
		}
		if strings.TrimSpace(confirmation) != expected {
			return BillingModeConversionResult{}, ErrBillingModeConfirmationRequired
		}
		factor := billingModeFactor(current.CurrencyCode(), current.USDExchangeRate(), target, usdToCNYRate)
		result, err := s.repo.ConvertBillingMode(ctx, BillingModeConversion{
			FromCurrency: current.CurrencyCode(), FromRate: current.USDExchangeRate(),
			ToCurrency: target, ToRate: usdToCNYRate, Factor: factor,
			HistoryEntry: map[string]any{"from": current.CurrencyCode(), "to": target, "from_rate": current.USDExchangeRate(), "to_rate": usdToCNYRate, "factor": factor},
		})
		if err != nil {
			if err == ErrBillingModeActiveJobs {
				return BillingModeConversionResult{}, err
			}
			return BillingModeConversionResult{}, infraerrors.InternalServer("BILLING_MODE_CONVERSION_FAILED", "billing currency conversion failed").WithCause(err)
		}
		s.current = billingConfigAfterConversion(current, target, usdToCNYRate, factor)
		if s.cacheFlush != nil {
			// Balances and quota snapshots are cached in Redis. A full cache clear is
			// intentional here: a partial clear could mix USD and CNY values.
			_ = s.cacheFlush(ctx)
		}
		return result, nil
	}

	if err := s.repo.SaveBillingMode(ctx, target, usdToCNYRate); err != nil {
		return BillingModeConversionResult{}, infraerrors.InternalServer("BILLING_MODE_SAVE_FAILED", "billing mode settings could not be saved").WithCause(err)
	}
	s.current = s.withCurrency(target, usdToCNYRate)
	return BillingModeConversionResult{}, nil
}
