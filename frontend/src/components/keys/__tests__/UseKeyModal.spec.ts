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

  it('switches the Codex client to WebSocket transport', async () => {
    const wrapper = mountModal()
    expect(wrapper.text()).toContain('Codex')
    expect(wrapper.text()).toContain('ChatGPT 桌面端')
    expect(wrapper.findAll('button').some((button) => button.text() === 'Codex CLI')).toBe(false)
    const wsButton = wrapper.findAll('button').find((button) => button.text() === 'WebSocket')
    expect(wsButton).toBeDefined()
    await wsButton!.trigger('click')
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

  it('offers Codex and Claude Code for Grok without injecting an unverified model', async () => {
    const wrapper = mountModal('grok')
    expect(wrapper.text()).toContain('Codex')
    expect(wrapper.text()).toContain('Claude Code')
    expect(wrapper.text()).toContain('Grok CLI')
    expect(wrapper.text()).toContain('model_provider = "OpenAI"')
    expect(wrapper.text()).toContain('~/.codex/config.toml')
    expect(wrapper.text()).not.toContain('grok-4.5')

    const claudeTab = wrapper.findAll('button').find((button) => button.text() === 'Claude Code')
    expect(claudeTab).toBeDefined()
    await claudeTab!.trigger('click')
    await nextTick()

    expect(wrapper.text()).toContain('ANTHROPIC_BASE_URL')
    expect(wrapper.text()).toContain('~/.claude/settings.json')
    expect(wrapper.text()).not.toContain('grok-4.5')

    const grokTab = wrapper.findAll('button').find((button) => button.text() === 'Grok CLI')
    expect(grokTab).toBeDefined()
    await grokTab!.trigger('click')
    await nextTick()

    expect(wrapper.text()).toContain('base_url = "https://example.com/v1"')
    expect(wrapper.text()).not.toContain('grok-4.5')
  })
})
