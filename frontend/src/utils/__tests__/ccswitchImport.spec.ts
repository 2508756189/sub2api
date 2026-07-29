import { describe, expect, it } from 'vitest'
import {
  buildCcSwitchImportDeeplink
} from '@/utils/ccswitchImport'
import type { GroupPlatform } from '@/types'

function paramsFromDeeplink(deeplink: string): URLSearchParams {
  const query = deeplink.split('?')[1] || ''
  return new URLSearchParams(query)
}

describe('ccswitchImport utils', () => {
  const baseInput = {
    baseUrl: 'https://api.example.com',
    providerName: 'Sub2API',
    apiKey: 'sk-test',
    usageScript: 'return true'
  }

  it('does not invent a model for OpenAI CC Switch imports', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        clientType: 'codex'
      })
    )

    expect(params.get('resource')).toBe('provider')
    expect(params.get('app')).toBe('codex')
    expect(params.get('endpoint')).toBe(`${baseInput.baseUrl}/v1`)
    expect(params.has('model')).toBe(false)
    expect(atob(params.get('usageScript') || '')).toBe(baseInput.usageScript)
  })

  it('adds the selected Codex model and config payload when provided', () => {
    const config = JSON.stringify({
      auth: { OPENAI_API_KEY: 'sk-test' },
      config: 'model = "deepseek-v4-pro"\nmodel_reasoning_effort = "high"'
    })
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        clientType: 'codex',
        model: 'deepseek-v4-pro',
        configFormat: 'json',
        config
      })
    )

    expect(params.get('model')).toBe('deepseek-v4-pro')
    expect(params.get('configFormat')).toBe('json')
    expect(params.getAll('configFormat')).toEqual(['json'])
    expect(atob(params.get('config') || '')).toBe(config)
  })

  it.each([
    { platform: 'anthropic' as GroupPlatform, clientType: 'claude' as const, app: 'claude' },
    { platform: 'gemini' as GroupPlatform, clientType: 'gemini' as const, app: 'gemini' }
  ])('does not add a model parameter for $platform imports', ({ platform, clientType, app }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform,
        clientType
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.has('model')).toBe(false)
  })

  it.each([
    'https://api.example.com',
    'https://api.example.com/',
    'https://api.example.com/v1',
    'https://api.example.com/v1/'
  ])('imports Grok Build with one /v1 suffix for base URL %s', (baseUrl) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        baseUrl,
        platform: 'grok',
        clientType: 'grokbuild'
      })
    )

    expect(params.get('app')).toBe('grokbuild')
    expect(params.get('endpoint')).toBe('https://api.example.com/v1')
    expect(params.has('model')).toBe(false)
  })

  it.each([
    { clientType: 'codex' as const, app: 'codex', endpoint: 'https://api.example.com/v1' },
    { clientType: 'claude' as const, app: 'claude', endpoint: 'https://api.example.com' },
    { clientType: 'grokbuild' as const, app: 'grokbuild', endpoint: 'https://api.example.com/v1' },
    { clientType: 'opencode' as const, app: 'opencode', endpoint: 'https://api.example.com/v1' }
  ])('uses target client $clientType instead of the Grok upstream platform', ({ clientType, app, endpoint }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'grok',
        clientType
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(endpoint)
  })

  it('keeps Antigravity imports on the selected client endpoint without a model parameter', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'antigravity',
        clientType: 'gemini'
      })
    )

    expect(params.get('app')).toBe('gemini')
    expect(params.get('endpoint')).toBe(`${baseInput.baseUrl}/antigravity`)
    expect(params.has('model')).toBe(false)
  })

  it('adds Claude tier model parameters when provided', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'anthropic',
        clientType: 'claude',
        claudeModelTiers: {
          haiku: 'claude-haiku-custom',
          sonnet: 'claude-sonnet-custom'
        }
      })
    )

    expect(params.get('haikuModel')).toBe('claude-haiku-custom')
    expect(params.get('sonnetModel')).toBe('claude-sonnet-custom')
    expect(params.has('opusModel')).toBe(false)
  })
})
