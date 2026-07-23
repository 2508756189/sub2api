import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn()
  })
}))

vi.mock('@/api/admin/accounts', () => ({
  accountsAPI: {
    syncUpstreamModels: vi.fn(),
    syncUpstreamModelsPreview: vi.fn()
  }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

import ModelWhitelistSelector from '../ModelWhitelistSelector.vue'

const mountSelector = (props: Record<string, unknown>) =>
  mount(ModelWhitelistSelector, {
    props: {
      modelValue: [],
      ...props
    },
    global: {
      stubs: {
        Icon: true,
        ModelIcon: true
      }
    }
  })

describe('ModelWhitelistSelector upstream sync availability', () => {
  it('hides live upstream sync for saved Grok OAuth accounts', () => {
    const wrapper = mountSelector({ platform: 'grok', accountId: 9, accountType: 'oauth' })

    expect(wrapper.find('[data-testid="sync-upstream-models"]').exists()).toBe(false)
  })

  it('shows live upstream sync for saved Grok API key accounts', () => {
    const wrapper = mountSelector({ platform: 'grok', accountId: 10, accountType: 'apikey' })

    expect(wrapper.find('[data-testid="sync-upstream-models"]').exists()).toBe(true)
  })

  it('keeps live upstream sync available for supported non-Grok accounts', () => {
    const wrapper = mountSelector({ platform: 'openai', accountId: 11, accountType: 'oauth' })

    expect(wrapper.find('[data-testid="sync-upstream-models"]').exists()).toBe(true)
  })
})
