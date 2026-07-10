import { describe, expect, it } from 'vitest'
import { buildClientInstallScript, buildGeminiFiles, buildOpenCodeFiles } from '../accessCenterFiles'

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
