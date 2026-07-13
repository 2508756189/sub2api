import type { GroupPlatform } from '@/types'
import type { ClaudeModelTier } from '@/constants/connectorPresets'

export type CcSwitchClientType = 'claude' | 'gemini'
export type CcSwitchConfigFormat = 'json' | 'toml'

export interface CcSwitchImportConfig {
  app: string
  endpoint: string
}

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
  platform?: GroupPlatform | null
  clientType: CcSwitchClientType
  providerName: string
  apiKey: string
  usageScript: string
  model?: string
  claudeModelTiers?: Partial<Record<ClaudeModelTier, string>>
  config?: string
  configFormat?: CcSwitchConfigFormat
}

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  clientType: CcSwitchClientType,
  baseUrl: string
): CcSwitchImportConfig {
  switch (platform || 'anthropic') {
    case 'antigravity':
      return {
        app: clientType === 'gemini' ? 'gemini' : 'claude',
        endpoint: `${baseUrl}/antigravity`
      }
    case 'openai':
      return {
        app: 'codex',
        endpoint: baseUrl
      }
    case 'gemini':
      return {
        app: 'gemini',
        endpoint: baseUrl
      }
    case 'grok':
      return {
        app: 'codex',
        endpoint: baseUrl
      }
    default:
      return {
        app: 'claude',
        endpoint: baseUrl
      }
  }
}

function encodeBase64Utf8(value: string): string {
  if (typeof TextEncoder !== 'undefined') {
    const bytes = new TextEncoder().encode(value)
    let binary = ''
    bytes.forEach((byte) => {
      binary += String.fromCharCode(byte)
    })
    return btoa(binary)
  }

  return btoa(unescape(encodeURIComponent(value)))
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const config = resolveCcSwitchImportConfig(input.platform, input.clientType, input.baseUrl)
  const entries: [string, string][] = [
    ['resource', 'provider'],
    ['app', config.app],
    ['name', input.providerName],
    ['homepage', input.baseUrl],
    ['endpoint', config.endpoint],
    ['apiKey', input.apiKey],
    ['configFormat', 'json'],
    ['usageEnabled', 'true'],
    ['usageScript', btoa(input.usageScript)],
    ['usageAutoInterval', '30']
  ]

  const model = input.model?.trim()
  if (model) {
    entries.splice(2, 0, ['model', model])
  }

  const tiers = input.claudeModelTiers ?? {}
  const tierParams: Partial<Record<ClaudeModelTier, string>> = {
    haiku: tiers.haiku?.trim(),
    sonnet: tiers.sonnet?.trim(),
    opus: tiers.opus?.trim()
  }
  if (tierParams.haiku) entries.push(['haikuModel', tierParams.haiku])
  if (tierParams.sonnet) entries.push(['sonnetModel', tierParams.sonnet])
  if (tierParams.opus) entries.push(['opusModel', tierParams.opus])

  if (input.config) {
    entries.push(['configFormat', input.configFormat || 'json'])
    entries.push(['config', encodeBase64Utf8(input.config)])
  }

  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}
