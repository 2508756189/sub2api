import type { GroupPlatform } from '@/types'
import type { ClaudeModelTier } from '@/constants/connectorPresets'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.5'
export const GROK_CC_SWITCH_MODEL = 'grok-4.5'

export type CcSwitchClientType = 'claude' | 'codex' | 'gemini' | 'grokbuild' | 'opencode'
export type CcSwitchConfigFormat = 'json' | 'toml'

export interface CcSwitchImportConfig {
  app: string
  endpoint: string
  model?: string
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

/**
 * Balance query CC Switch runs against the imported provider. CC Switch fills
 * `{{baseUrl}}` with the provider's base URL as stored ? Codex and Grok imports
 * carry a trailing `/v1` (see `withV1Endpoint`), Claude ones do not, and users
 * may edit it either way afterwards ? then evaluates the script, so the URL
 * strips an existing `/v1` instead of blindly appending one (`/v1/v1/usage`
 * is a 404 and CC Switch shows "query failed").
 */
export const CC_SWITCH_USAGE_SCRIPT = `({
    request: {
      url: "{{baseUrl}}".replace(/\\/+$/, "").replace(/\\/v1$/, "") + "/v1/usage",
      method: "GET",
      headers: { "Authorization": "Bearer {{apiKey}}" }
    },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "USD";
      return {
        isValid: response?.is_active ?? response?.isValid ?? true,
        remaining,
        unit
      };
    }
  })`

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
}

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  clientType: CcSwitchClientType,
  baseUrl: string
): CcSwitchImportConfig {
  const cleanBaseUrl = baseUrl.replace(/\/+$/, '')
  if (clientType === 'grokbuild') {
    return { app: 'grokbuild', endpoint: withV1Endpoint(cleanBaseUrl) }
  }
  if (clientType === 'codex') {
    return { app: 'codex', endpoint: withV1Endpoint(cleanBaseUrl) }
  }
  if (clientType === 'opencode') {
    const endpoint = platform === 'gemini'
      ? `${cleanBaseUrl}/v1beta`
      : withV1Endpoint(cleanBaseUrl)
    return { app: 'opencode', endpoint }
  }
  if (clientType === 'gemini') {
    return {
      app: 'gemini',
      endpoint: platform === 'antigravity' ? `${cleanBaseUrl}/antigravity` : cleanBaseUrl
    }
  }
  return {
    app: 'claude',
    endpoint: platform === 'antigravity' ? `${cleanBaseUrl}/antigravity` : cleanBaseUrl
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
  const configFormat = input.config ? input.configFormat || 'json' : 'json'
  const entries: [string, string][] = [
    ['resource', 'provider'],
    ['app', config.app],
    ['name', input.providerName],
    ['homepage', input.baseUrl],
    ['endpoint', config.endpoint],
    ['apiKey', input.apiKey],
    ['configFormat', configFormat],
    ['usageEnabled', 'true'],
    ['usageScript', btoa(input.usageScript)],
    ['usageAutoInterval', '30']
  ]

  const model = input.model?.trim() || config.model
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
    entries.push(['config', encodeBase64Utf8(input.config)])
  }

  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}
