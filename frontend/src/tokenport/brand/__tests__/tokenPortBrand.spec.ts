import { describe, expect, it } from 'vitest'

import { resolveTokenPortLogo, TOKENPORT_BRAND } from '../tokenPortBrand'

describe('resolveTokenPortLogo', () => {
  it('uses the TokenPort default only when no logo is configured', () => {
    expect(resolveTokenPortLogo()).toBe(TOKENPORT_BRAND.logo)
    expect(resolveTokenPortLogo('   ')).toBe(TOKENPORT_BRAND.logo)
  })

  it('preserves legacy and custom logo paths', () => {
    expect(resolveTokenPortLogo('/logo.png')).toBe('/logo.png')
    expect(resolveTokenPortLogo('/branding/company.svg')).toBe('/branding/company.svg')
  })
})
