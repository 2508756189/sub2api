import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import authLayout from '@/components/layout/AuthLayout.vue?raw'
import tokenPortHome from '@/tokenport/home/TokenPortHome.vue?raw'
import consolePreview from '@/tokenport/home/ConsolePreview.vue?raw'
import skillMarketView from '@/views/user/SkillMarketView.vue?raw'
import skillMarketCatalog from '@/tokenport/market/SkillMarketCatalog.vue?raw'
import settingsView from '@/views/admin/SettingsView.vue?raw'

// .css 的 ?raw 导入在 vitest 里会被 CSS 管线截空,改用 fs 直读。
const readCss = (relative: string) =>
  readFileSync(fileURLToPath(new URL(relative, import.meta.url)), 'utf-8')
const styleCss = readCss('../../style.css')
const consoleCss = readCss('../brand/tokenport-console.css')

// openspec/changes/add-tokenport-frontend-polish — frontend-theme 回归断言。
// 这些断言锁的是「单一品牌令牌 + 主题完整性 + 刻度回归」,改动样式时若触发失败,
// 先读 spec 再决定是否修订断言。

describe('frontend-theme R2 品牌令牌单一来源', () => {
  it('style.css 是品牌色唯一定义点,且带 .dark 变体', () => {
    expect(styleCss).toContain('--tp-brand: #00a878')
    expect(styleCss).toMatch(/\.dark\s*\{[^}]*--tp-brand:/s)
  })

  it.each([
    ['AuthLayout', authLayout],
    ['TokenPortHome', tokenPortHome],
    ['ConsolePreview', consolePreview],
  ])('%s 不再持有局部 --brand 字面量,改为引用全局令牌', (_name, source) => {
    expect(source).not.toMatch(/--brand:\s*#[0-9a-fA-F]/)
    expect(source).toContain('--brand: var(--tp-brand)')
  })

  it('tokenport-console.css 的 mint 系映射到全局令牌,不再持有 #2ad4b8 家族字面量', () => {
    expect(consoleCss).toContain('--tp-mint: var(--tp-brand)')
    expect(consoleCss).not.toContain('#2ad4b8')
    expect(consoleCss).not.toContain('#36dcc0')
  })

  it('TokenPortHome 的 .dark 强调色走令牌(final-cta 恒暗横幅保留唯一字面量)', () => {
    expect(tokenPortHome.split('#7fe0bc').length - 1).toBe(1)
    expect(tokenPortHome).not.toContain('#41d0a1')
  })
})

describe('frontend-theme R1 主题完整性', () => {
  it('SkillMarketView 未登录壳三处容器均有 dark: 变体', () => {
    expect(skillMarketView).toMatch(/bg-\[#f4f8f6\][^"]*dark:bg-dark-950/)
    expect(skillMarketView).toMatch(/bg-white\/90[^"]*dark:bg-dark-900\/90/)
    expect(skillMarketView).toMatch(/<footer[^>]*dark:bg-dark-900/)
  })

  it('ConsolePreview 提供 .dark 覆盖块', () => {
    expect(consolePreview).toContain('.dark .console-preview')
  })

  it('SettingsView settings-tabs-shell 有 dark: 变体与深色阴影覆盖', () => {
    expect(settingsView).toMatch(/settings-tabs-shell\s*\{[^}]*dark:bg-dark-900\/90/s)
    expect(settingsView).toContain('.dark .settings-tabs-shell')
  })
})

describe('frontend-theme R3 定制面回归全站刻度', () => {
  it('SkillMarketCatalog 去除任意字号与超标圆角', () => {
    expect(skillMarketCatalog).not.toContain('text-[26px]')
    expect(skillMarketCatalog).not.toContain('text-[15px]')
    expect(skillMarketCatalog).not.toContain('rounded-2xl')
  })

  it('SkillMarketCatalog 品牌强调收敛到 primary(风险徽章的语义色 emerald-100/700 除外)', () => {
    expect(skillMarketCatalog).not.toContain('border-emerald')
    expect(skillMarketCatalog).not.toContain('bg-emerald-50 ')
    expect(skillMarketCatalog).not.toContain('from-emerald-500')
    expect(skillMarketCatalog).not.toContain('hover:border-emerald')
  })

  it('SkillMarketView 未登录壳不再使用 emerald 边框', () => {
    expect(skillMarketView).not.toContain('emerald')
  })
})
