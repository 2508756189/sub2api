export type BillingCurrency = 'CNY' | 'USD'

export function normalizeBillingCurrency(value?: string | null): BillingCurrency {
  return value?.trim().toUpperCase() === 'CNY' ? 'CNY' : 'USD'
}

export function getBillingCurrency(): BillingCurrency {
  if (typeof window === 'undefined') return 'CNY'
  return window.__APP_CONFIG__?.billing_currency
    ? normalizeBillingCurrency(window.__APP_CONFIG__.billing_currency)
    : 'CNY'
}

export function getBillingCurrencySymbol(currency = getBillingCurrency()): string {
  return currency === 'CNY' ? '¥' : '$'
}

export function getBillingCurrencyLabel(currency = getBillingCurrency()): string {
  return currency === 'CNY' ? '人民币' : '美元'
}

export function formatBillingAmount(
  value: number | string | null | undefined,
  fractionDigits = 2
): string {
  const amount = Number(value ?? 0)
  const currency = getBillingCurrency()
  return new Intl.NumberFormat(currency === 'CNY' ? 'zh-CN' : 'en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits
  }).format(Number.isFinite(amount) ? amount : 0)
}

export function formatBillingNumber(
  value: number | string | null | undefined,
  fractionDigits = 2
): string {
  const amount = Number(value ?? 0)
  return `${getBillingCurrencySymbol()}${(Number.isFinite(amount) ? amount : 0).toFixed(fractionDigits)}`
}
