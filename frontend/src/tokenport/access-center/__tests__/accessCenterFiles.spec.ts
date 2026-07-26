import { describe, expect, it } from 'vitest'

import type { SkillInstallSelection } from '@/api/skillMarket'
import {
  buildClientInstallScript,
  buildGeminiFiles,
  buildGrokFiles,
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
    // 数组必须做并集而不是整体覆盖，否则用户已有的 enabledPlugins 会被删掉，
    // 与「保留已有项目设置」的文案相反。
    expect(script).toContain('def merge_arrays(cur, inc):')
    expect(script).not.toContain('current[key].update(value)')
    expect(script).toContain('tokenport-backup')
  })

  it('expands $HOME in the bash install script instead of quoting it literally', () => {
    const files = [
      { path: '~/.codex/config.toml', content: 'model_provider = "OpenAI"' },
      { path: '~/.claude/settings.json', content: '{"env":{}}' },
    ]
    const script = buildClientInstallScript(files, 'unix')

    expect(script).toContain(`merge_tokenport_file "$HOME"/'.codex/config.toml'`)
    expect(script).toContain(`merge_tokenport_file "$HOME"/'.claude/settings.json'`)
    // 单引号内 POSIX shell 不做参数展开，配置会落到当前目录下一个名为 $HOME 的文件夹。
    expect(script).not.toMatch(/'\$HOME/)
  })

  it('keeps the PowerShell install script compatible with Windows PowerShell 5.1', () => {
    const files = [
      { path: '%userprofile%\\.codex\\config.toml', content: 'model_provider = "OpenAI"' },
      { path: '%userprofile%\\.codex\\auth.json', content: '{"OPENAI_API_KEY":"sk-test"}' },
    ]
    const script = buildClientInstallScript(files, 'powershell')

    // ConvertFrom-Json -AsHashtable 只有 PowerShell 6+ 才有，5.1 会直接终止脚本。
    expect(script).not.toContain('-AsHashtable')
    expect(script).toContain('function ConvertTo-TokenPortMap($Value) {')
    expect(script).toContain('$Value.PSObject.Properties')
    // 目标文件存在但为空时不能崩在 ConvertFrom-Json 上。
    expect(script).toContain('if ([string]::IsNullOrWhiteSpace($raw)) { return @{} }')
    expect(script).toContain(`Merge-TokenPortJson -Target (Join-Path $HOME '.codex\\auth.json')`)
    // PS 5.1 的 Set-Content -Encoding UTF8 会写 BOM，Grok/Codex 的 TOML/JSON 解析会失败。
    expect(script).not.toContain('Set-Content')
    expect(script).toContain('New-Object System.Text.UTF8Encoding($false)')
    // 0 字节文件 Get-Content -Raw 返回 $null，后续 .Trim()/IsMatch 会崩。
    expect(script).toContain('if ($null -eq $existing) { $existing = "" }')
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
    // psQuote 自带外层单引号，模板再包一层会得到 ''markitdown''，
    // PowerShell 里三个参数全部绑定为空串，下载目录里只剩一个空文件夹。
    expect(preparation!.content).toContain("-SkillId 'markitdown'")
    expect(preparation!.content).toContain("-ArchiveUrl 'https://example.com/markitdown.zip'")
    expect(preparation!.content).toContain(
      "-ExpectedSha '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'",
    )
    expect(preparation!.content).not.toContain("''markitdown''")
  })
})

describe('buildGrokFiles', () => {
  it('builds the current Grok Build model profile schema', () => {
    const [file] = buildGrokFiles('https://example.com/', 'sk-test', 'powershell', {
      codex: { model: 'grok-4.5', reasoningEffort: 'medium', mcpServers: [] },
    })

    expect(file.path).toBe('%userprofile%\\.grok/config.toml')
    expect(file.content).toContain('[models]\ndefault = "grok-4.5"')
    expect(file.content).toContain('[model."grok-4.5"]')
    expect(file.content).toContain('model = "grok-4.5"')
    expect(file.content).toContain('base_url = "https://example.com/v1"')
    expect(file.content).toContain('api_key = "sk-test"')
    expect(file.content).toContain('api_backend = "responses"')
    expect(file.content).toContain('context_window = 500000')
  })

  it('does not create an invalid Grok Build profile when no model is selected', () => {
    const [file] = buildGrokFiles('https://example.com', 'sk-test', 'unix', {
      codex: { model: '', reasoningEffort: 'medium', mcpServers: [] },
    })

    expect(file.path).toBe('Grok Build 配置说明.txt')
    expect(file.content).toContain('必须指定一个模型')
    expect(file.content).not.toContain('grok-4.5')
  })
})
