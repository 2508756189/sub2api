package config

import (
	"fmt"
	"strings"
)

const (
	BillingCurrencyUSD = "USD"
	BillingCurrencyCNY = "CNY"
)

func NormalizeBillingCurrency(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case BillingCurrencyCNY:
		return BillingCurrencyCNY
	case "", BillingCurrencyUSD:
		return BillingCurrencyUSD
	default:
		return strings.ToUpper(strings.TrimSpace(value))
	}
}

func (c BillingConfig) CurrencyCode() string {
	return NormalizeBillingCurrency(c.Currency)
}

func (c BillingConfig) USDExchangeRate() float64 {
	if c.CurrencyCode() == BillingCurrencyCNY && c.USDToCNYRate > 0 {
		return c.USDToCNYRate
	}
	return 1
}

func (c BillingConfig) ValidateCurrency() error {
	currency := c.CurrencyCode()
	if currency != BillingCurrencyUSD && currency != BillingCurrencyCNY {
		return fmt.Errorf("billing.currency must be USD or CNY")
	}
	if currency == BillingCurrencyCNY && c.USDToCNYRate <= 0 {
		return fmt.Errorf("billing.usd_to_cny_rate must be positive when billing.currency is CNY")
	}
	return nil
}
