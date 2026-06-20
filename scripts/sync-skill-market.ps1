param(
  [string]$SkillRepoPath = "..\state-of-art-skills"
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
if ([System.IO.Path]::IsPathRooted($SkillRepoPath)) {
  $skillRepo = Resolve-Path $SkillRepoPath
} else {
  $skillRepo = Resolve-Path (Join-Path $repoRoot $SkillRepoPath)
}
$marketIndex = Join-Path $skillRepo "market\index.json"
$archiveDir = Join-Path $skillRepo "dist\skills"
$targetRoot = Join-Path $repoRoot "frontend\public\skill-market"
$targetArchiveDir = Join-Path $targetRoot "dist\skills"

if (-not (Test-Path -LiteralPath $marketIndex)) {
  throw "Missing market index: $marketIndex"
}

if (-not (Test-Path -LiteralPath $archiveDir)) {
  throw "Missing skill archive directory: $archiveDir"
}

New-Item -ItemType Directory -Force -Path $targetArchiveDir | Out-Null
Copy-Item -LiteralPath $marketIndex -Destination (Join-Path $targetRoot "index.json") -Force

Get-ChildItem -LiteralPath $targetArchiveDir -Filter *.zip | Remove-Item -Force
Copy-Item -Path (Join-Path $archiveDir "*.zip") -Destination $targetArchiveDir -Force

$archiveCount = (Get-ChildItem -LiteralPath $targetArchiveDir -Filter *.zip | Measure-Object).Count
Write-Output "Synced Skill Market: index.json + $archiveCount archives -> $targetRoot"
