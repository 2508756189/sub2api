import { beforeEach, describe, expect, it } from 'vitest'

import {
  formatBillingAmount,
  formatBillingNumber,
  getBillingCurrency,
  getBillingCurrencyLabel,
  getBillingCurrencySymbol
} from '../currency'

describe('TokenPort billing currency', () => {
  beforeEach(() => {
    window.__APP_CONFIG__ = {} as typeof window.__APP_CONFIG__
  })

  it('defaults to CNY for the TokenPort deployment', () => {
    expect(getBillingCurrency()).toBe('CNY')
    expect(getBillingCurrencySymbol()).toBe('\u00a5')
    expect(formatBillingNumber(12.3)).toBe('\u00a512.30')
  })

  it('formats the dedicated CNY deployment in RMB', () => {
    window.__APP_CONFIG__ = { billing_currency: 'CNY' } as typeof window.__APP_CONFIG__

    expect(getBillingCurrency()).toBe('CNY')
    expect(getBillingCurrencySymbol()).toBe('\u00a5')
    expect(getBillingCurrencyLabel()).toBe('\u4eba\u6c11\u5e01')
    expect(formatBillingNumber(12.3)).toBe('\u00a512.30')
    expect(formatBillingAmount(12.3)).toContain('12.30')
  })
})
