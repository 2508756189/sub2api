import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import authLayout from '@/components/layout/AuthLayout.vue?raw'
import tokenPortHome from '@/tokenport/home/TokenPortHome.vue?raw'
import consolePreview from '@/tokenport/home/ConsolePreview.vue?raw'
import skillMarketView from '@/views/user/SkillMarketView.vue?raw'
import skillMarketCatalog from '@/tokenport/market/SkillMarketCatalog.vue?raw'
import skillMarketCard from '@/tokenport/market/SkillMarketCard.vue?raw'
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

  it('TokenPortHome 提供纯白主题并在 dark 覆盖中锁定 Figma 深色令牌', () => {
    expect(tokenPortHome).toContain('--color-ground: #ffffff')
    expect(tokenPortHome).toContain('--color-panel: #ffffff')
    expect(tokenPortHome).toContain('.tp-home.dark-mode')
    expect(tokenPortHome).toContain('--color-ground: #0a0e0d')
    expect(tokenPortHome).toContain('--color-panel: #121b18')
    expect(tokenPortHome).toContain('--color-primary: #2fd4a0')
    expect(tokenPortHome).toContain('--color-accent: #4fd6e0')
    expect(tokenPortHome).toContain('--radius: 14px')
    expect(tokenPortHome).toContain('useThemeToggle')
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

describe('frontend-theme R3 圆角与层级刻度', () => {
  it('style.css 定义三档圆角与三级层级,且层级有 .dark 变体', () => {
    for (const token of ['--tp-radius-control', '--tp-radius-card', '--tp-radius-panel']) {
      expect(styleCss).toContain(token)
    }
    for (const token of ['--tp-elev-1', '--tp-elev-2', '--tp-elev-3']) {
      expect(styleCss).toContain(token)
    }
    expect(styleCss).toMatch(/\.dark\s*\{[^}]*--tp-elev-1:/s)
  })

  it('层级以组件类提供 —— Tailwind 任意值解析不了带逗号的多层阴影,会静默产出 none', () => {
    expect(styleCss).toContain('.tp-elev-1 {')
    expect(styleCss).toContain('.tp-elev-raise:hover')
    expect(skillMarketCatalog).not.toContain('shadow-[var(')
  })

  it.each([
    ['ConsolePreview', consolePreview],
  ])('%s 的圆角与卡片阴影走令牌,不再各写各的', (_name, source) => {
    // 实测这两个文件曾并存 10/12/14/16/18 五种半径与多组同透明度不同模糊的阴影。
    expect(source).not.toMatch(/border-radius:\s*1[2468]px/)
    expect(source).not.toMatch(/border-radius:\s*10px/)
    expect(source).toContain('var(--tp-radius-')
    expect(source).toContain('var(--tp-elev-')
  })

  it('TokenPortHome 的主面板圆角引用 Figma radius 令牌', () => {
    expect(tokenPortHome).toContain('border-radius: var(--radius)')
    expect(tokenPortHome).not.toContain('var(--tp-elev-')
  })

  it('技能市场卡片有静置层级与悬停抬升(此前是无阴影的纯描边)', () => {
    expect(skillMarketCard).toContain('var(--tp-elev-1')
    expect(skillMarketCard).toContain('var(--tp-elev-2')
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

  it('首页和正式目录共用同一 Skill Market 卡片结构', () => {
    expect(tokenPortHome).toContain("import SkillMarketCard from '@/tokenport/market/SkillMarketCard.vue'")
    expect(skillMarketCatalog).toContain("import SkillMarketCard from './SkillMarketCard.vue'")
    for (const field of ['skill.version', 'skill.runtime', 'skill.archive.size', 'skill.archive.sha256']) {
      expect(skillMarketCard).toContain(field)
    }
  })

  it('首页功能图标统一使用 Icon 组件而不是 v-html SVG', () => {
    expect(tokenPortHome).toContain('<Icon :name="feature.icon"')
    expect(tokenPortHome).not.toContain('v-html="feature.icon"')
  })

  it('首页次按钮避开 Tailwind outline 工具类冲突', () => {
    expect(tokenPortHome).not.toMatch(/class="[^"]*tp-button outline/)
    expect(tokenPortHome).toContain('class="tp-button secondary large"')
    expect(tokenPortHome).toContain('.tp-button.secondary')
  })
})
