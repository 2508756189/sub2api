import { describe, expect, it } from 'vitest'

import type { ConnectorOptions } from '@/constants/connectorPresets'
import type { SkillInstallSelection } from '@/api/skillMarket'
import {
  buildAnthropicFiles,
  buildOpenAIFiles,
  buildOpenAIWsFiles,
} from '../connectorTemplates'

const skill: SkillInstallSelection = {
  id: 'markitdown',
  name: 'markitdown',
  archiveUrl: 'https://cdn.jsdelivr.net/gh/2508756189/state-of-art-skills@main/dist/skills/markitdown.zip',
  sha256: '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef',
  installTargets: {
    claude: '~/.claude/skills/markitdown',
    codex: '~/.codex/skills/markitdown',
    portable: '~/.agents/skills/markitdown',
  },
}

describe('connectorTemplates', () => {
  it('keeps Claude Code output at the upstream baseline when no options are selected', () => {
    const files = buildAnthropicFiles({
      baseUrl: 'http://127.0.0.1:8080',
      apiKey: 'sk-test',
      shell: 'unix',
      options: undefined,
      selectedSkills: [],
    })

    expect(files).toHaveLength(2)
    expect(files[0].content).toBe(`export ANTHROPIC_BASE_URL="http://127.0.0.1:8080"
export ANTHROPIC_AUTH_TOKEN="sk-test"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`)
    expect(files[1].content).not.toContain('enabledPlugins')
    expect(files.map((file) => file.path).join('\n')).not.toContain('Install')
  })

  it('adds Claude Code model tier env and official plugins when selected', () => {
    const options: ConnectorOptions = {
      claude: {
        modelTiers: {
          haiku: 'claude-3-5-haiku-latest',
          sonnet: 'claude-sonnet-4-5',
          opus: '',
        },
        enabledPlugins: ['github@claude-plugins-official'],
      },
      codex: {
        model: 'gpt-5.5',
        reasoningEffort: 'xhigh',
        mcpServers: [],
      },
    }

    const files = buildAnthropicFiles({
      baseUrl: 'http://127.0.0.1:8080',
      apiKey: 'sk-test',
      shell: 'powershell',
      options,
      selectedSkills: [],
    })

    expect(files[0].content).toContain('$env:ANTHROPIC_DEFAULT_HAIKU_MODEL="claude-3-5-haiku-latest"')
    expect(files[0].content).toContain('$env:ANTHROPIC_DEFAULT_SONNET_MODEL="claude-sonnet-4-5"')
    expect(files[0].content).not.toContain('ANTHROPIC_DEFAULT_OPUS_MODEL')
    expect(files[1].content).toContain('"enabledPlugins": [')
    expect(files[1].content).toContain('"github@claude-plugins-official"')
  })

  it('adds Codex model, reasoning, and MCP settings when selected', () => {
    const files = buildOpenAIFiles({
      baseUrl: 'http://127.0.0.1:8080',
      apiKey: 'sk-test',
      shell: 'unix',
      options: {
        codex: {
          model: 'gpt-5.2',
          reasoningEffort: 'high',
          mcpServers: ['context7'],
        },
      },
      selectedSkills: [],
    })

    expect(files[0].content).toContain('model = "gpt-5.2"')
    expect(files[0].content).toContain('review_model = "gpt-5.2"')
    expect(files[0].content).toContain('model_reasoning_effort = "high"')
    expect(files[0].content).toContain('[mcp_servers.context7]')
    expect(files[0].content).toContain('command = "npx"')
  })

  it('generates Claude Code skill install scripts with checksum verification', () => {
    const files = buildAnthropicFiles({
      baseUrl: 'http://127.0.0.1:8080',
      apiKey: 'sk-test',
      shell: 'unix',
      selectedSkills: [skill],
    })

    const installScript = files.find((file) => file.path.includes('Install Claude Code skills'))

    expect(installScript).toBeDefined()
    expect(installScript!.content).toContain('curl -L')
    expect(installScript!.content).toContain(skill.archiveUrl)
    expect(installScript!.content).toContain(skill.sha256)
    expect(installScript!.content).toContain('sha256sum')
    expect(installScript!.content).toContain('shasum -a 256')
    expect(installScript!.content).toContain('unzip -o -q')
    expect(installScript!.content).toContain('$HOME/.claude/skills/markitdown')
  })

  it('generates PowerShell skill install scripts with checksum verification', () => {
    const files = buildAnthropicFiles({
      baseUrl: 'http://127.0.0.1:8080',
      apiKey: 'sk-test',
      shell: 'powershell',
      selectedSkills: [skill],
    })

    const installScript = files.find((file) => file.path.includes('Install Claude Code skills'))

    expect(installScript).toBeDefined()
    expect(installScript!.content).toContain('Invoke-WebRequest')
    expect(installScript!.content).toContain('Get-FileHash')
    expect(installScript!.content).toContain('Expand-Archive')
    expect(installScript!.content).toContain('$HOME\\.claude\\skills\\markitdown')
  })

  it('adds selected skills to Codex install script and keeps WebSocket settings', () => {
    const files = buildOpenAIWsFiles({
      baseUrl: 'http://127.0.0.1:8080',
      apiKey: 'sk-test',
      shell: 'unix',
      selectedSkills: [skill],
    })

    const config = files.find((file) => file.path.endsWith('config.toml'))
    const installScript = files.find((file) => file.path.includes('Install Codex skills'))

    expect(config!.content).toContain('supports_websockets = true')
    expect(config!.content).toContain('responses_websockets_v2 = true')
    expect(installScript).toBeDefined()
    expect(installScript!.content).toContain('$HOME/.codex/skills/markitdown')
    expect(installScript!.content).toContain('unzip -o -q')
  })
})
