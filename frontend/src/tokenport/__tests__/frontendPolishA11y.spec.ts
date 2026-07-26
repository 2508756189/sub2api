import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import baseDialog from '@/components/common/BaseDialog.vue?raw'
import loginView from '@/views/auth/LoginView.vue?raw'
import zhCommon from '@/i18n/locales/zh/common'
import enCommon from '@/i18n/locales/en/common'

// .css 的 ?raw 导入在 vitest 里会被 CSS 管线截空,改用 fs 直读。
const readCss = (relative: string) =>
  readFileSync(fileURLToPath(new URL(relative, import.meta.url)), 'utf-8')
const styleCss = readCss('../../style.css')

// openspec/changes/add-tokenport-frontend-polish — frontend-a11y 回归断言。
// 锁的是「弹窗焦点圈定 + 键盘焦点可见 + 图标按钮可访问名」,改动前先读 spec。

describe('frontend-a11y R1 模态框焦点圈定', () => {
  it('BaseDialog 监听 Tab 并在首尾元素间循环', () => {
    expect(baseDialog).toContain("event.key === 'Tab'")
    // 反向 Tab 到首元素时跳到尾元素,正向 Tab 到尾元素时跳回首元素。
    expect(baseDialog).toMatch(/shiftKey/)
    expect(baseDialog).toMatch(/last(Focusable)?\??\.focus\(\)/)
    expect(baseDialog).toMatch(/first(Focusable)?\??\.focus\(\)/)
  })

  it('可聚焦元素选择器排除 disabled 与 hidden,避免焦点落进死元素', () => {
    expect(baseDialog).toContain(':not([disabled])')
  })

  it('圈定只在弹窗打开时生效', () => {
    expect(baseDialog).toMatch(/if\s*\(!?\s*props\.show/)
  })
})

describe('frontend-a11y R2 键盘焦点样式', () => {
  it('style.css 提供全局 :focus-visible 规则', () => {
    expect(styleCss).toContain(':focus-visible')
  })

  it('.btn 基类的焦点环改为 focus-visible,鼠标点击不触发', () => {
    const btnBlock = styleCss.slice(styleCss.indexOf('.btn {'), styleCss.indexOf('.btn-primary'))
    expect(btnBlock).toContain('focus-visible:ring-2')
    expect(btnBlock).not.toMatch(/[^-]focus:ring-2/)
  })

  it('.input 保留 focus 边框(点击需可见)但焦点环走 focus-visible', () => {
    const inputBlock = styleCss.slice(styleCss.indexOf('.input {'), styleCss.indexOf('.input-error'))
    expect(inputBlock).toContain('focus:border-primary-500')
    expect(inputBlock).toContain('focus-visible:ring-2')
    expect(inputBlock).not.toMatch(/[^-]focus:ring-2/)
  })
})

describe('frontend-a11y R3 图标按钮可访问名', () => {
  it('BaseDialog 关闭按钮不再硬编码英文', () => {
    expect(baseDialog).not.toContain('aria-label="Close modal"')
    expect(baseDialog).toContain(":aria-label=\"t('common.close')\"")
  })

  it('LoginView 密码可见性切换带随状态变化的 aria-label', () => {
    expect(loginView).toMatch(/:aria-label="showPassword \? t\('common\.hidePassword'\) : t\('common\.showPassword'\)"/)
  })

  it('两个语言包都提供 showPassword / hidePassword', () => {
    for (const pack of [zhCommon, enCommon]) {
      expect(pack.common.showPassword).toBeTruthy()
      expect(pack.common.hidePassword).toBeTruthy()
    }
  })
})
