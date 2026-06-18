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

export interface CodexModelPreset {
  id: string
  label: string
}

export interface CodexMcpServerPreset {
  id: string
  label: string
  command: string
  args: string[]
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

export const CODEX_MODELS: CodexModelPreset[] = [
  { id: 'gpt-5.5', label: 'GPT-5.5' },
  { id: 'gpt-5.2', label: 'GPT-5.2' },
  { id: 'gpt-5.4', label: 'GPT-5.4' },
  { id: 'gpt-5.4-mini', label: 'GPT-5.4 Mini' },
  { id: 'gpt-5.3-codex', label: 'GPT-5.3 Codex' },
  { id: 'codex-mini-latest', label: 'Codex Mini' },
]

export const CODEX_REASONING_EFFORTS: CodexReasoningEffort[] = [
  'minimal',
  'low',
  'medium',
  'high',
  'xhigh',
]

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
    model: 'gpt-5.5',
    reasoningEffort: 'xhigh',
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
      model: options?.codex?.model || DEFAULT_CONNECTOR_OPTIONS.codex?.model || 'gpt-5.5',
      reasoningEffort: options?.codex?.reasoningEffort || DEFAULT_CONNECTOR_OPTIONS.codex?.reasoningEffort || 'xhigh',
      mcpServers: [...(options?.codex?.mcpServers ?? DEFAULT_CONNECTOR_OPTIONS.codex?.mcpServers ?? [])],
    },
  }
}
