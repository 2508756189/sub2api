import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('@/api/channels', () => ({
  default: { getAvailable: vi.fn().mockResolvedValue([]) },
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn().mockResolvedValue(true) }),
}))

import UseKeyModal from '../UseKeyModal.vue'

const global = {
  stubs: {
    BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
    ConnectorOptions: { template: '<div data-test="connector-options" />' },
    SkillMarketSelector: { template: '<div data-test="skill-market" />' },
  },
}

function mountModal(
  platform: 'openai' | 'anthropic' | 'antigravity' | 'grok' = 'openai',
  initialMode: 'direct' | 'ccs' = 'direct',
) {
  return mount(UseKeyModal, {
    props: {
      show: true,
      apiKey: 'sk-test',
      baseUrl: 'https://example.com',
      platform,
      initialMode,
    },
    global,
  })
}

describe('UseKeyModal', () => {
  it('omits a model when the user has not selected one', () => {
    const wrapper = mountModal()
    expect(wrapper.text()).toContain('直接配置')
    expect(wrapper.text()).toContain('model_provider = "OpenAI"')
    expect(wrapper.text()).not.toContain('REPLACE_WITH_AVAILABLE_MODEL')
    expect(wrapper.text()).not.toMatch(/^model\s*=/m)
  })

  it('switches to the Codex WebSocket configuration', async () => {
    const wrapper = mountModal()
    const wsTab = wrapper.findAll('button').find((button) => button.text().includes('Codex WebSocket'))
    expect(wsTab).toBeDefined()
    await wsTab!.trigger('click')
    await nextTick()
    expect(wrapper.text()).toContain('supports_websockets = true')
    expect(wrapper.text()).not.toContain('REPLACE_WITH_AVAILABLE_MODEL')
  })

  it('generates a small OpenCode provider config without invented model catalogs', async () => {
    const wrapper = mountModal()
    const tab = wrapper.findAll('button').find((button) => button.text() === 'OpenCode')
    expect(tab).toBeDefined()
    await tab!.trigger('click')
    await nextTick()
    expect(wrapper.text()).toContain('https://example.com/v1')
    expect(wrapper.text()).toContain('opencode.json')
    expect(wrapper.text()).not.toContain('gpt-5.5')
  })

  it('opens directly in CCS mode without requiring a model', () => {
    const wrapper = mountModal('anthropic', 'ccs')
    expect(wrapper.text()).toContain('CCS 只导入客户端配置')
    expect(wrapper.findAll('button').some((button) => button.text() === '导入 CCS')).toBe(true)
  })

  it('supports Grok without injecting an unverified model', () => {
    const wrapper = mountModal('grok')
    expect(wrapper.text()).toContain('Grok CLI')
    expect(wrapper.text()).toContain('base_url = "https://example.com/v1"')
    expect(wrapper.text()).not.toContain('grok-4.5')
  })
})
