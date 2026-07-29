import { describe, expect, it } from 'vitest'
import tokenPortHome from '@/tokenport/home/TokenPortHome.vue?raw'
import skillMarketView from '@/views/user/SkillMarketView.vue?raw'
import skillMarketCatalog from '@/tokenport/market/SkillMarketCatalog.vue?raw'

// TokenPort Figma Make 最终稿 — 首页导航与公开入口回归断言。

describe('frontend-navigation R1 移动端公开页导航保底', () => {
  it('桌面导航锚点与 Figma 信息架构一致', () => {
    for (const href of ['#platform', '#cost', '#skill-market', '#deploy']) {
      expect(tokenPortHome).toContain(`href="${href}"`)
    }
  })

  it('移动端提供可访问的折叠菜单并保留四个核心入口', () => {
    expect(tokenPortHome).toContain(':aria-expanded="menuOpen"')
    expect(tokenPortHome).toContain('class="mobile-nav"')
    expect(tokenPortHome).toContain('@click="menuOpen = !menuOpen"')
  })

  it('Skill Market 卡片进入真实公开目录', () => {
    expect(tokenPortHome).toMatch(/to="\/skill-market"\s+class="skill-card/)
  })
})

describe('frontend-navigation R2 公开页头部一致', () => {
  it('技能市场未登录壳提供主题切换', () => {
    expect(skillMarketView).toContain('toggleTheme')
    expect(skillMarketView).toMatch(/:name="isDark \? 'sun' : 'moon'"/)
  })

  it('Figma 首页固定深色视觉，不受本地主题状态污染', () => {
    expect(tokenPortHome).not.toContain('useThemeToggle')
    expect(tokenPortHome).not.toContain("localStorage.setItem('theme'")
  })
})

describe('frontend-a11y R4 移动端触控目标', () => {
  it('技能卡「查看详情」移动端不低于 40px', () => {
    expect(skillMarketCatalog).toContain('min-h-[40px]')
  })
})
