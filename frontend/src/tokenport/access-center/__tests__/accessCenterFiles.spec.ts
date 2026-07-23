import { describe, expect, it } from 'vitest'

import type { SkillInstallSelection } from '@/api/skillMarket'
import {
  buildClientInstallScript,
  buildGeminiFiles,
  buildOpenCodeFiles,
  buildTeleAgentFiles,
} from '../accessCenterFiles'

const skill: SkillInstallSelection = {
  id: 'markitdown',
  name: 'markitdown',
  archiveUrl: 'https://example.com/markitdown.zip',
  sha256: '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef',
  installTargets: {},
}

describe('accessCenterFiles', () => {
  it('generates Gemini environment variables without inventing a model', () => {
    const [file] = buildGeminiFiles('https://gateway.example.com', 'sk-test', 'unix')
    expect(file.content).toContain('GOOGLE_GEMINI_BASE_URL="https://gateway.example.com/v1beta"')
    expect(file.content).not.toContain('GEMINI_MODEL')
  })

  it('generates a minimal OpenCode provider configuration', () => {
    const [file] = buildOpenCodeFiles('https://gateway.example.com/v1', 'sk-test', 'openai')
    expect(JSON.parse(file.content)).toEqual({
      $schema: 'https://opencode.ai/config.json',
      provider: { openai: { options: { baseURL: 'https://gateway.example.com/v1', apiKey: 'sk-test' } } },
    })
  })

  it('uses managed TOML blocks and JSON merge with backups', () => {
    const files = [
      { path: '~/.codex/config.toml', content: 'model_provider = "OpenAI"' },
      { path: '~/.codex/auth.json', content: '{"OPENAI_API_KEY":"sk-test"}' },
    ]
    const script = buildClientInstallScript(files, 'unix')
    expect(script).toContain('TokenPort managed config')
    expect(script).toContain('current[key].update(value)')
    expect(script).toContain('tokenport-backup')
  })
})

describe('buildTeleAgentFiles', () => {
  it('creates OpenAI-compatible provider fields without inventing a model', () => {
    const files = buildTeleAgentFiles('https://example.com/', 'sk-test', { codex: { model: '' } })

    expect(files).toHaveLength(1)
    expect(files[0].path).toBe('teleagent-provider-fields.json')
    expect(files[0].content).toContain('"protocol": "OpenAI Compatible"')
    expect(files[0].content).toContain('"baseUrl": "https://example.com/v1"')
    expect(files[0].content).not.toContain('"model"')
  })

  it('adds selected skills as a SHA256-verified import manifest', () => {
    const files = buildTeleAgentFiles('https://example.com', 'sk-test', { codex: { model: 'deepseek-v4-flash' } }, [skill])
    const provider = files.find((file) => file.path === 'teleagent-provider-fields.json')
    const manifest = files.find((file) => file.path === 'teleagent-skill-import-manifest.json')
    const preparation = files.find((file) => file.path === 'Prepare TeleAgent skills (PowerShell)')

    expect(provider).toBeDefined()
    expect(provider!.content).toContain('deepseek-v4-flash')
    expect(manifest).toBeDefined()
    expect(manifest!.content).toContain('"verification": "SHA256"')
    expect(manifest!.content).toContain(skill.archiveUrl)
    expect(manifest!.content).toContain(skill.sha256)
    expect(manifest!.content).toContain('teleagent-root')
    expect(preparation).toBeDefined()
    expect(preparation!.content).toContain('Invoke-WebRequest')
    expect(preparation!.content).toContain('Get-FileHash')
    expect(preparation!.content).toContain(skill.archiveUrl)
  })
})
