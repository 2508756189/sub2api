import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import zhAdminEnsemble from '@/i18n/locales/zh/admin/ensemble'
import zhCommon from '@/i18n/locales/zh/common'
import EnsembleTestDialog from '../EnsembleTestDialog.vue'

// vitest.config.ts 把 vue-i18n 指向 runtime-only 构建且没有开 JIT，
// createI18n 装真实 messages 也只会回落成 key 本身（见 SubscriptionPlanCard.spec
// 故意断言 "/ payment.perMonth" 这种未解析形态）。所以这里沿用仓库里
// 104 个 spec 的 vi.mock 约定，但让 t 去查真实的 zh 文案模块而不是返回 key：
// key 写错或文案漏了，下面的断言照样会红。
const messages: Record<string, unknown> = { ...zhCommon, admin: { ...zhAdminEnsemble } }

function resolve(key: string): string {
  const value = key.split('.').reduce<unknown>(
    (node, part) => (node && typeof node === 'object' ? (node as Record<string, unknown>)[part] : undefined),
    messages
  )
  return typeof value === 'string' ? value : key
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, string | number>) =>
      resolve(key).replace(/\{(\w+)\}/g, (whole, name) =>
        params && name in params ? String(params[name]) : whole
      )
  })
}))

describe('EnsembleTestDialog', () => {
  it('成员成本为 null 时显示未返回且不会调用 toFixed', () => {
    const wrapper = mount(EnsembleTestDialog, {
      props: {
        show: true,
        testing: false,
        error: '',
        result: null,
        events: [{
          type: 'member_finished',
          model: 'gpt-5',
          platform: 'openai',
          role: 'proposer',
          member: {
            model: 'gpt-5',
            platform: 'openai',
            role: 'proposer',
            status: 'failed',
            duration_ms: 20,
            cost: null,
            error: 'upstream failed'
          }
        } as any]
      },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('未返回')
    expect(wrapper.text()).toContain('upstream failed')
    // 任何 key 没解析都会把点分路径本身渲染出来，这条断言把"文案接线断了"
    // 和"断言恰好还能过"区分开。
    expect(wrapper.text()).not.toContain('admin.ensemble')
  })
})
