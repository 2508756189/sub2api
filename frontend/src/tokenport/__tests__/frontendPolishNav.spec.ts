import { describe, expect, it } from 'vitest'
import tokenPortHome from '@/tokenport/home/TokenPortHome.vue?raw'
import skillMarketView from '@/views/user/SkillMarketView.vue?raw'
import skillMarketCatalog from '@/tokenport/market/SkillMarketCatalog.vue?raw'

// openspec/changes/add-tokenport-frontend-polish — frontend-navigation 回归断言。
// 锁的是「移动端保底入口 + 公开页头部一致 + 门控预示」,改动前先读 spec。

describe('frontend-navigation R1 移动端公开页导航保底', () => {
  it('≤720px 只隐藏次要入口,不再连主题切换一起隐藏', () => {
    const mobileBlock = tokenPortHome.slice(tokenPortHome.indexOf('@media (max-width: 720px)'))
    expect(mobileBlock).toContain('.topbar nav > .nav-link-optional')
    // 早前的写法把 .icon-control(主题切换)也关掉了,顶栏只剩登录按钮。
    expect(mobileBlock).not.toMatch(/\.topbar nav \.icon-control\s*\{[^}]*display:\s*none/s)
  })

  it('Skill Market 入口不带 optional 标记,移动端保留', () => {
    expect(tokenPortHome).toMatch(/to="\/skill-market"\s+class="nav-link"/)
  })

  it('次要入口(模型与渠道 / Docs)标记为 optional', () => {
    expect(tokenPortHome).toMatch(/to="\/available-channels"[\s\S]{0,120}nav-link-optional/)
    expect(tokenPortHome).toMatch(/:href="docUrl"[\s\S]{0,160}nav-link-optional/)
  })
})

describe('frontend-navigation R2 公开页头部一致', () => {
  it('技能市场未登录壳提供主题切换', () => {
    expect(skillMarketView).toContain('toggleTheme')
    expect(skillMarketView).toMatch(/:name="isDark \? 'sun' : 'moon'"/)
  })

  it('两个公开页共用同一个主题切换实现,不再各自复制', () => {
    for (const source of [tokenPortHome, skillMarketView]) {
      expect(source).toContain("useThemeToggle")
      expect(source).not.toContain("localStorage.setItem('theme'")
    }
  })
})

describe('frontend-navigation R3 门控链接预示', () => {
  it('未登录时「模型与渠道」带锁形标识与说明', () => {
    expect(tokenPortHome).toMatch(/v-if="!isAuthenticated"\s+name="lock"/)
    expect(tokenPortHome).toContain('需登录后查看，登录成功将回到此页')
  })
})

describe('frontend-a11y R4 移动端触控目标', () => {
  it('技能卡「查看详情」移动端不低于 40px', () => {
    expect(skillMarketCatalog).toContain('min-h-[40px]')
  })
})
