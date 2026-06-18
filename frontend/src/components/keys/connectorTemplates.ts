import type { SkillInstallSelection } from '@/api/skillMarket'
import {
  CLAUDE_MODEL_TIERS,
  CODEX_MCP_SERVERS,
  type ConnectorOptions,
  normalizeConnectorOptions,
} from '@/constants/connectorPresets'

export type ConnectorShell = 'unix' | 'cmd' | 'powershell' | 'windows'

export interface FileConfig {
  path: string
  content: string
  hint?: string
  highlighted?: string
}

export interface BuildConnectorFilesInput {
  baseUrl: string
  apiKey: string
  shell: ConnectorShell
  options?: ConnectorOptions
  selectedSkills?: SkillInstallSelection[]
  configTomlHint?: string
}

function hasClaudeOptions(options?: ConnectorOptions): boolean {
  const normalized = normalizeConnectorOptions(options)
  return CLAUDE_MODEL_TIERS.some((tier) => normalized.claude.modelTiers?.[tier.id]) ||
    normalized.claude.enabledPlugins.length > 0
}

function hasCodexOptions(options?: ConnectorOptions): boolean {
  const normalized = normalizeConnectorOptions(options)
  return normalized.codex.model !== 'gpt-5.5' ||
    normalized.codex.reasoningEffort !== 'xhigh' ||
    normalized.codex.mcpServers.length > 0
}

function unixEnvLine(name: string, value: string): string {
  return `export ${name}="${value}"`
}

function cmdEnvLine(name: string, value: string): string {
  return `set ${name}=${value}`
}

function powershellEnvLine(name: string, value: string): string {
  return `$env:${name}="${value}"`
}

function normalizeHomeTarget(target: string | undefined, runtime: 'claude' | 'codex', shell: ConnectorShell): string {
  const fallback = runtime === 'claude' ? `~/.claude/skills` : `~/.codex/skills`
  const source = target || fallback
  if (shell === 'powershell' || shell === 'windows') {
    return source.replace(/^~\//, '$HOME\\').replace(/\//g, '\\')
  }
  return source.replace(/^~\//, '$HOME/')
}

function shellQuote(value: string): string {
  return value.replace(/"/g, '\\"')
}

function psQuote(value: string): string {
  return value.replace(/'/g, "''")
}

function buildUnixSkillInstallScript(runtime: 'claude' | 'codex', selectedSkills: SkillInstallSelection[]): string {
  const lines = [
    'set -euo pipefail',
    '',
    'install_skill() {',
    '  skill_id="$1"',
    '  archive_url="$2"',
    '  expected_sha="$3"',
    '  target="$4"',
    '  zip_path="$(mktemp -t "${skill_id}.XXXXXX.zip")"',
    '  echo "Installing ${skill_id} -> ${target}"',
    '  mkdir -p "$(dirname "$target")"',
    '  curl -L "$archive_url" -o "$zip_path"',
    '  if command -v sha256sum >/dev/null 2>&1; then',
    '    actual_sha="$(sha256sum "$zip_path" | awk \'{print $1}\')"',
    '  else',
    '    actual_sha="$(shasum -a 256 "$zip_path" | awk \'{print $1}\')"',
    '  fi',
    '  if [ "$actual_sha" != "$expected_sha" ]; then',
    '    echo "SHA256 mismatch for ${skill_id}: ${actual_sha}" >&2',
    '    exit 1',
    '  fi',
    '  rm -rf "$target"',
    '  unzip -o -q "$zip_path" -d "$(dirname "$target")"',
    '  rm -f "$zip_path"',
    '}',
    '',
  ]

  selectedSkills.forEach((skill) => {
    const target = normalizeHomeTarget(skill.installTargets[runtime], runtime, 'unix')
    lines.push(
      `install_skill "${shellQuote(skill.id)}" "${shellQuote(skill.archiveUrl)}" "${shellQuote(skill.sha256)}" "${shellQuote(target)}"`,
    )
  })

  return lines.join('\n')
}

function buildPowerShellSkillInstallScript(runtime: 'claude' | 'codex', selectedSkills: SkillInstallSelection[]): string {
  const lines = [
    '$ErrorActionPreference = "Stop"',
    '',
    'function Install-Skill {',
    '  param(',
    '    [string]$SkillId,',
    '    [string]$ArchiveUrl,',
    '    [string]$ExpectedSha,',
    '    [string]$Target',
    '  )',
    '  $zipPath = Join-Path $env:TEMP "$SkillId.zip"',
    '  Write-Host "Installing $SkillId -> $Target"',
    '  New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Target) | Out-Null',
    '  Invoke-WebRequest -Uri $ArchiveUrl -OutFile $zipPath',
    '  $actualSha = (Get-FileHash -Algorithm SHA256 -LiteralPath $zipPath).Hash.ToLowerInvariant()',
    '  if ($actualSha -ne $ExpectedSha.ToLowerInvariant()) {',
    '    throw "SHA256 mismatch for $SkillId: $actualSha"',
    '  }',
    '  if (Test-Path -LiteralPath $Target) { Remove-Item -LiteralPath $Target -Recurse -Force }',
    '  Expand-Archive -LiteralPath $zipPath -DestinationPath (Split-Path -Parent $Target) -Force',
    '  Remove-Item -LiteralPath $zipPath -Force',
    '}',
    '',
  ]

  selectedSkills.forEach((skill) => {
    const target = normalizeHomeTarget(skill.installTargets[runtime], runtime, 'powershell')
    lines.push(
      `Install-Skill -SkillId '${psQuote(skill.id)}' -ArchiveUrl '${psQuote(skill.archiveUrl)}' -ExpectedSha '${psQuote(skill.sha256)}' -Target '${psQuote(target)}'`,
    )
  })

  return lines.join('\n')
}

function buildSkillInstallFile(
  runtime: 'claude' | 'codex',
  shell: ConnectorShell,
  selectedSkills: SkillInstallSelection[] | undefined,
): FileConfig | null {
  if (!selectedSkills?.length) {
    return null
  }

  const isPowerShell = shell === 'powershell' || shell === 'windows'
  return {
    path: isPowerShell
      ? `Install ${runtime === 'claude' ? 'Claude Code' : 'Codex'} skills (PowerShell)`
      : `Install ${runtime === 'claude' ? 'Claude Code' : 'Codex'} skills (Bash)`,
    content: isPowerShell
      ? buildPowerShellSkillInstallScript(runtime, selectedSkills)
      : buildUnixSkillInstallScript(runtime, selectedSkills),
  }
}

export function buildAnthropicFiles(input: BuildConnectorFilesInput): FileConfig[] {
  const options = normalizeConnectorOptions(input.options)
  const selectedSkills = input.selectedSkills ?? []
  const extraEnv = CLAUDE_MODEL_TIERS
    .map((tier) => [tier.envName, options.claude.modelTiers?.[tier.id]] as const)
    .filter(([, value]) => Boolean(value))

  let path: string
  let content: string

  switch (input.shell) {
    case 'cmd':
      path = 'Command Prompt'
      content = [
        cmdEnvLine('ANTHROPIC_BASE_URL', input.baseUrl),
        cmdEnvLine('ANTHROPIC_AUTH_TOKEN', input.apiKey),
        cmdEnvLine('CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC', '1'),
        ...extraEnv.map(([name, value]) => cmdEnvLine(name, value || '')),
      ].join('\n')
      break
    case 'powershell':
      path = 'PowerShell'
      content = [
        powershellEnvLine('ANTHROPIC_BASE_URL', input.baseUrl),
        powershellEnvLine('ANTHROPIC_AUTH_TOKEN', input.apiKey),
        '$env:CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1',
        ...extraEnv.map(([name, value]) => powershellEnvLine(name, value || '')),
      ].join('\n')
      break
    default:
      path = 'Terminal'
      content = [
        unixEnvLine('ANTHROPIC_BASE_URL', input.baseUrl),
        unixEnvLine('ANTHROPIC_AUTH_TOKEN', input.apiKey),
        'export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1',
        ...extraEnv.map(([name, value]) => unixEnvLine(name, value || '')),
      ].join('\n')
  }

  const envSettings: Record<string, string> = {
    ANTHROPIC_BASE_URL: input.baseUrl,
    ANTHROPIC_AUTH_TOKEN: input.apiKey,
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: '1',
    CLAUDE_CODE_ATTRIBUTION_HEADER: '0',
  }

  extraEnv.forEach(([name, value]) => {
    if (value) {
      envSettings[name] = value
    }
  })

  const settings: Record<string, unknown> = { env: envSettings }
  if (options.claude.enabledPlugins.length > 0) {
    settings.enabledPlugins = options.claude.enabledPlugins
  }

  const files: FileConfig[] = [
    { path, content },
    {
      path: input.shell === 'unix' ? '~/.claude/settings.json' : '%userprofile%\\.claude\\settings.json',
      content: hasClaudeOptions(input.options)
        ? JSON.stringify(settings, null, 2)
        : `{
  "env": {
    "ANTHROPIC_BASE_URL": "${input.baseUrl}",
    "ANTHROPIC_AUTH_TOKEN": "${input.apiKey}",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0"
  }
}`,
      hint: 'VSCode Claude Code',
    },
  ]

  const installFile = buildSkillInstallFile('claude', input.shell, selectedSkills)
  if (installFile) {
    files.push(installFile)
  }

  return files
}

function codexConfigDir(shell: ConnectorShell): string {
  return shell === 'windows' || shell === 'powershell' ? '%userprofile%\\.codex' : '~/.codex'
}

function buildCodexConfig(input: BuildConnectorFilesInput, webSocket: boolean): string {
  const options = normalizeConnectorOptions(input.options)
  const model = hasCodexOptions(input.options) ? options.codex.model : 'gpt-5.5'
  const reasoningEffort = hasCodexOptions(input.options) ? options.codex.reasoningEffort : 'xhigh'
  const lines = [
    'model_provider = "OpenAI"',
    `model = "${model}"`,
    `review_model = "${model}"`,
    `model_reasoning_effort = "${reasoningEffort}"`,
    'disable_response_storage = true',
    'network_access = "enabled"',
    'windows_wsl_setup_acknowledged = true',
    '',
    '[model_providers.OpenAI]',
    'name = "OpenAI"',
    `base_url = "${input.baseUrl}"`,
    'wire_api = "responses"',
  ]

  if (webSocket) {
    lines.push('supports_websockets = true')
  }

  lines.push('requires_openai_auth = true', '', '[features]')

  if (webSocket) {
    lines.push('responses_websockets_v2 = true')
  }

  lines.push('goals = true')

  options.codex.mcpServers.forEach((serverId) => {
    const preset = CODEX_MCP_SERVERS.find((server) => server.id === serverId)
    if (!preset) return
    lines.push(
      '',
      `[mcp_servers.${preset.id}]`,
      `command = "${preset.command}"`,
      `args = ${JSON.stringify(preset.args)}`,
    )
  })

  return lines.join('\n')
}

function buildCodexFiles(input: BuildConnectorFilesInput, webSocket: boolean): FileConfig[] {
  const configDir = codexConfigDir(input.shell)
  const files: FileConfig[] = [
    {
      path: `${configDir}/config.toml`,
      content: buildCodexConfig(input, webSocket),
      hint: input.configTomlHint,
    },
    {
      path: `${configDir}/auth.json`,
      content: `{
  "OPENAI_API_KEY": "${input.apiKey}"
}`,
    },
  ]

  const installFile = buildSkillInstallFile('codex', input.shell, input.selectedSkills)
  if (installFile) {
    files.push(installFile)
  }

  return files
}

export function buildOpenAIFiles(input: BuildConnectorFilesInput): FileConfig[] {
  return buildCodexFiles(input, false)
}

export function buildOpenAIWsFiles(input: BuildConnectorFilesInput): FileConfig[] {
  return buildCodexFiles(input, true)
}
