export const TOKENPORT_BRAND = {
  name: 'TokenPort',
  subtitle: '天翼云智能应用接入与经营平台',
  proposition: '统一管理模型、Token 与智能体能力',
  upstreamName: 'Sub2API',
  upstreamUrl: 'https://github.com/Wei-Shaw/sub2api',
  logo: '/ctyun-logo.svg',
} as const

export const TOKENPORT_PRODUCT = {
  protocols: ['OpenAI', 'Anthropic', 'Gemini'],
  clients: ['Codex CLI', 'Claude Code', 'OpenCode', 'Gemini CLI', 'CCS'],
  capabilities: ['统一接入', '模型路由', 'Token 核算', '部门预算', 'Skill Market', '私有化部署'],
} as const

export function resolveTokenPortName(configuredName?: string | null): string {
  const name = configuredName?.trim()
  return !name || name === TOKENPORT_BRAND.upstreamName ? TOKENPORT_BRAND.name : name
}

export function resolveTokenPortSubtitle(configuredSubtitle?: string | null): string {
  const subtitle = configuredSubtitle?.trim()
  if (!subtitle || subtitle === 'Subscription to API Conversion Platform') {
    return TOKENPORT_BRAND.subtitle
  }
  return subtitle
}

export function resolveTokenPortLogo(configuredLogo?: string | null): string {
  return configuredLogo?.trim() || TOKENPORT_BRAND.logo
}
