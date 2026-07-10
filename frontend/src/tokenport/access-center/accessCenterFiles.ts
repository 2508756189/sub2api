import type { FileConfig } from '@/components/keys/connectorTemplates'
import type { ConnectorShell } from '@/components/keys/connectorTemplates'

export function ensureApiVersion(baseUrl: string, version = 'v1'): string {
  const clean = baseUrl.replace(/\/+$/, '').replace(/\/v1(?:beta)?\/?$/, '')
  return `${clean}/${version}`
}

export function buildGeminiFiles(baseUrl: string, apiKey: string, shell: ConnectorShell): FileConfig[] {
  const endpoint = ensureApiVersion(baseUrl, 'v1beta')
  if (shell === 'cmd') {
    return [{ path: 'Command Prompt', content: `set GOOGLE_GEMINI_BASE_URL=${endpoint}\nset GEMINI_API_KEY=${apiKey}` }]
  }
  if (shell === 'powershell' || shell === 'windows') {
    return [{ path: 'PowerShell', content: `$env:GOOGLE_GEMINI_BASE_URL="${endpoint}"\n$env:GEMINI_API_KEY="${apiKey}"` }]
  }
  return [{ path: 'Terminal', content: `export GOOGLE_GEMINI_BASE_URL="${endpoint}"\nexport GEMINI_API_KEY="${apiKey}"` }]
}

export function buildOpenCodeFiles(baseUrl: string, apiKey: string, provider: string): FileConfig[] {
  return [{
    path: 'opencode.json',
    content: JSON.stringify({
      $schema: 'https://opencode.ai/config.json',
      provider: {
        [provider]: {
          options: { baseURL: baseUrl, apiKey },
        },
      },
    }, null, 2),
  }]
}

export function isSkillInstallFile(file: FileConfig): boolean {
  return file.path.startsWith('Install ') && file.path.includes(' skills ')
}

export function isWritableClientFile(file: FileConfig): boolean {
  const path = file.path.toLowerCase().replace(/\\/g, '/')
  return (path.includes('/.codex/') || path.includes('/.claude/')) &&
    (path.endsWith('.json') || path.endsWith('.toml'))
}

function encodeBase64Utf8(value: string): string {
  const bytes = new TextEncoder().encode(value)
  let binary = ''
  bytes.forEach((byte) => { binary += String.fromCharCode(byte) })
  return btoa(binary)
}

function bashQuote(value: string): string {
  return `'${value.replace(/'/g, `'\\''`)}'`
}

function psQuote(value: string): string {
  return `'${value.replace(/'/g, "''")}'`
}

function bashPath(path: string): string {
  return path.replace(/^~[\\/]/, '$HOME/').replace(/^%userprofile%[\\/]/i, '$HOME/').replace(/\\/g, '/')
}

function powershellPath(path: string): string {
  const normalized = path.replace(/\//g, '\\')
  const relative = normalized.match(/^(?:%userprofile%|~)\\(.+)$/i)?.[1]
  return relative ? `Join-Path $HOME ${psQuote(relative)}` : psQuote(normalized)
}

export function buildClientInstallScript(files: FileConfig[], shell: ConnectorShell): string {
  const writable = files.filter(isWritableClientFile)
  if (!writable.length) return ''
  return shell === 'unix' ? buildBashInstallScript(writable) : buildPowerShellInstallScript(writable)
}

function buildPowerShellInstallScript(files: FileConfig[]): string {
  const lines = [
    '$ErrorActionPreference = "Stop"',
    '',
    'function Backup-TokenPortFile([string]$Target) {',
    '  if (Test-Path -LiteralPath $Target) {',
    '    $backup = "$Target.tokenport-backup-$(Get-Date -Format yyyyMMddHHmmss)"',
    '    Copy-Item -LiteralPath $Target -Destination $backup -Force',
    '    Write-Host "TokenPort backup written: $backup"',
    '  }',
    '}',
    '',
    'function Merge-TokenPortJson([string]$Target, [string]$Payload) {',
    '  $incoming = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($Payload)) | ConvertFrom-Json -AsHashtable',
    '  $current = @{}',
    '  if (Test-Path -LiteralPath $Target) { $current = Get-Content -LiteralPath $Target -Raw | ConvertFrom-Json -AsHashtable }',
    '  foreach ($key in $incoming.Keys) {',
    '    if ($incoming[$key] -is [hashtable] -and $current[$key] -is [hashtable]) {',
    '      foreach ($child in $incoming[$key].Keys) { $current[$key][$child] = $incoming[$key][$child] }',
    '    } else { $current[$key] = $incoming[$key] }',
    '  }',
    '  $parent = Split-Path -Parent $Target; if ($parent) { New-Item -ItemType Directory -Force -Path $parent | Out-Null }',
    '  Backup-TokenPortFile $Target',
    '  $current | ConvertTo-Json -Depth 20 | Set-Content -LiteralPath $Target -Encoding UTF8',
    '}',
    '',
    'function Merge-TokenPortToml([string]$Target, [string]$Payload) {',
    '  $content = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($Payload)).Trim()',
    '  $begin = "# >>> TokenPort managed config >>>"; $end = "# <<< TokenPort managed config <<<"',
    '  $block = "$begin`n$content`n$end"',
    '  $existing = if (Test-Path -LiteralPath $Target) { Get-Content -LiteralPath $Target -Raw } else { "" }',
    '  $pattern = [regex]::Escape($begin) + ".*?" + [regex]::Escape($end)',
    '  $next = if ([regex]::IsMatch($existing, $pattern, "Singleline")) { [regex]::Replace($existing, $pattern, $block, "Singleline") } elseif ($existing.Trim()) { $existing.TrimEnd() + "`n`n" + $block } else { $block }',
    '  $parent = Split-Path -Parent $Target; if ($parent) { New-Item -ItemType Directory -Force -Path $parent | Out-Null }',
    '  Backup-TokenPortFile $Target',
    '  Set-Content -LiteralPath $Target -Value $next -Encoding UTF8',
    '}',
    '',
  ]
  files.forEach((file) => {
    const fn = file.path.toLowerCase().endsWith('.toml') ? 'Merge-TokenPortToml' : 'Merge-TokenPortJson'
    lines.push(`${fn} -Target (${powershellPath(file.path)}) -Payload ${psQuote(encodeBase64Utf8(file.content))}`)
  })
  return lines.join('\n')
}

function buildBashInstallScript(files: FileConfig[]): string {
  const lines = [
    'set -euo pipefail',
    '',
    'merge_tokenport_file() {',
    '  target="$1"; payload="$2"; kind="$3"',
    '  mkdir -p "$(dirname "$target")"',
    '  [ ! -f "$target" ] || cp "$target" "${target}.tokenport-backup-$(date +%Y%m%d%H%M%S)"',
    '  python3 - "$target" "$payload" "$kind" <<\'PY\'',
    'import base64, json, pathlib, re, sys',
    'target, payload, kind = pathlib.Path(sys.argv[1]), sys.argv[2], sys.argv[3]',
    'incoming = base64.b64decode(payload).decode("utf-8")',
    'if kind == "json":',
    '    current = json.loads(target.read_text(encoding="utf-8-sig")) if target.exists() and target.read_text(encoding="utf-8-sig").strip() else {}',
    '    patch = json.loads(incoming)',
    '    for key, value in patch.items():',
    '        if isinstance(value, dict) and isinstance(current.get(key), dict): current[key].update(value)',
    '        else: current[key] = value',
    '    target.write_text(json.dumps(current, ensure_ascii=False, indent=2) + "\\n", encoding="utf-8")',
    'else:',
    '    begin, end = "# >>> TokenPort managed config >>>", "# <<< TokenPort managed config <<<"',
    '    block = f"{begin}\\n{incoming.strip()}\\n{end}"',
    '    current = target.read_text(encoding="utf-8") if target.exists() else ""',
    '    pattern = re.compile(re.escape(begin) + r".*?" + re.escape(end), re.S)',
    '    next_text = pattern.sub(block, current) if pattern.search(current) else (current.rstrip() + "\\n\\n" + block if current.strip() else block)',
    '    target.write_text(next_text + "\\n", encoding="utf-8")',
    'PY',
    '}',
    '',
  ]
  files.forEach((file) => {
    const kind = file.path.toLowerCase().endsWith('.toml') ? 'toml' : 'json'
    lines.push(`merge_tokenport_file ${bashQuote(bashPath(file.path))} ${bashQuote(encodeBase64Utf8(file.content))} ${kind}`)
  })
  return lines.join('\n')
}
