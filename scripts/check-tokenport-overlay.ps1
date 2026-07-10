param(
  [string]$Root = (Split-Path -Parent $PSScriptRoot)
)

$ErrorActionPreference = 'Stop'
$manifestPath = Join-Path $Root 'tokenport-overlay.json'
if (-not (Test-Path -LiteralPath $manifestPath)) {
  throw "TokenPort overlay manifest not found: $manifestPath"
}

$manifest = Get-Content -LiteralPath $manifestPath -Raw -Encoding UTF8 | ConvertFrom-Json
$productDirectory = Join-Path $Root $manifest.productDirectory
if (-not (Test-Path -LiteralPath $productDirectory)) {
  throw "TokenPort product directory is missing: $productDirectory"
}

$failures = @()
foreach ($mount in $manifest.mountPoints) {
  $path = Join-Path $Root $mount.file
  if (-not (Test-Path -LiteralPath $path)) {
    $failures += "Missing mount file: $($mount.file)"
    continue
  }
  $content = Get-Content -LiteralPath $path -Raw -Encoding UTF8
  if (-not $content.Contains([string]$mount.contains)) {
    $failures += "Missing marker '$($mount.contains)' in $($mount.file)"
  }
}

if ($failures.Count -gt 0) {
  $failures | ForEach-Object { Write-Error $_ }
  exit 1
}

$moduleCount = (Get-ChildItem -LiteralPath $productDirectory -Recurse -File).Count
Write-Host "TokenPort overlay check passed: $moduleCount product files, $($manifest.mountPoints.Count) mount points."
