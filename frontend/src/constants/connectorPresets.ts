export type ClaudeModelTier = 'haiku' | 'sonnet' | 'opus'
export type CodexReasoningEffort = 'minimal' | 'low' | 'medium' | 'high' | 'xhigh'

export interface ConnectorOptions {
  claude?: {
    modelTiers?: Partial<Record<ClaudeModelTier, string>>
    enabledPlugins?: string[]
  }
  codex?: {
    model?: string
    reasoningEffort?: CodexReasoningEffort
    mcpServers?: string[]
  }
}

export interface NormalizedConnectorOptions {
  claude: {
    modelTiers: Partial<Record<ClaudeModelTier, string>>
    enabledPlugins: string[]
  }
  codex: {
    model: string
    reasoningEffort: CodexReasoningEffort
    mcpServers: string[]
  }
}

export interface ClaudeModelTierPreset {
  id: ClaudeModelTier
  label: string
  envName: string
  placeholder: string
}

export interface PluginPreset {
  id: string
  label: string
  default: boolean
}

export interface CodexMcpServerPreset {
  id: string
  label: string
  command: string
  args: string[]
}

export interface CodexReasoningEffortPreset {
  id: CodexReasoningEffort
  label: string
  description: string
}

export const CLAUDE_MODEL_TIERS: ClaudeModelTierPreset[] = [
  {
    id: 'haiku',
    label: 'Haiku',
    envName: 'ANTHROPIC_DEFAULT_HAIKU_MODEL',
    placeholder: 'claude-3-5-haiku-latest',
  },
  {
    id: 'sonnet',
    label: 'Sonnet',
    envName: 'ANTHROPIC_DEFAULT_SONNET_MODEL',
    placeholder: 'claude-sonnet-4-5',
  },
  {
    id: 'opus',
    label: 'Opus',
    envName: 'ANTHROPIC_DEFAULT_OPUS_MODEL',
    placeholder: 'claude-opus-4-5',
  },
]

export const CLAUDE_PLUGINS: PluginPreset[] = [
  { id: 'agent-sdk-dev@claude-plugins-official', label: 'Agent SDK Dev', default: false },
  { id: 'claude-code-setup@claude-plugins-official', label: 'Claude Code Setup', default: false },
  { id: 'code-review@claude-plugins-official', label: 'Code Review', default: false },
  { id: 'context7@claude-plugins-official', label: 'Context7', default: false },
  { id: 'feature-dev@claude-plugins-official', label: 'Feature Dev', default: false },
  { id: 'frontend-design@claude-plugins-official', label: 'Frontend Design', default: false },
  { id: 'github@claude-plugins-official', label: 'GitHub', default: false },
  { id: 'playwright@claude-plugins-official', label: 'Playwright', default: false },
]

export const CODEX_REASONING_EFFORT_OPTIONS: CodexReasoningEffortPreset[] = [
  {
    id: 'low',
    label: '速度优先',
    description: '适合补全、轻量修改和简单问答，响应更快、推理成本更低。',
  },
  {
    id: 'medium',
    label: '日常编码（推荐）',
    description: '适合多数开发、排查和文档任务，速度与质量比较均衡。',
  },
  {
    id: 'high',
    label: '深度分析',
    description: '适合复杂重构、长链路排障、方案评审和关键代码修改。',
  },
  {
    id: 'xhigh',
    label: '最高强度',
    description: '适合高风险决策和难题攻坚，质量优先，耗时和成本更高。',
  },
  {
    id: 'minimal',
    label: '最小推理',
    description: '适合非常确定的机械操作，不建议作为默认值。',
  },
]

export const CODEX_REASONING_EFFORTS: CodexReasoningEffort[] = CODEX_REASONING_EFFORT_OPTIONS.map((item) => item.id)

export const CODEX_MCP_SERVERS: CodexMcpServerPreset[] = [
  {
    id: 'context7',
    label: 'Context7',
    command: 'npx',
    args: ['-y', '@upstash/context7-mcp'],
  },
  {
    id: 'playwright',
    label: 'Playwright',
    command: 'npx',
    args: ['-y', '@playwright/mcp@latest'],
  },
]

export const DEFAULT_CONNECTOR_OPTIONS: ConnectorOptions = {
  claude: {
    modelTiers: {},
    enabledPlugins: [],
  },
  codex: {
    model: '',
    reasoningEffort: 'medium',
    mcpServers: [],
  },
}

export function normalizeConnectorOptions(options?: ConnectorOptions): NormalizedConnectorOptions {
  return {
    claude: {
      modelTiers: {
        ...DEFAULT_CONNECTOR_OPTIONS.claude?.modelTiers,
        ...options?.claude?.modelTiers,
      },
      enabledPlugins: [...(options?.claude?.enabledPlugins ?? DEFAULT_CONNECTOR_OPTIONS.claude?.enabledPlugins ?? [])],
    },
    codex: {
      model: options?.codex?.model || DEFAULT_CONNECTOR_OPTIONS.codex?.model || '',
      reasoningEffort: options?.codex?.reasoningEffort || DEFAULT_CONNECTOR_OPTIONS.codex?.reasoningEffort || 'medium',
      mcpServers: [...(options?.codex?.mcpServers ?? DEFAULT_CONNECTOR_OPTIONS.codex?.mcpServers ?? [])],
    },
  }
}
