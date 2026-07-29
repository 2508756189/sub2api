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
$detailDir = Join-Path $skillRepo "market\details"
$archiveDir = Join-Path $skillRepo "dist\skills"
$teleagentArchiveDir = Join-Path $skillRepo "dist\teleagent"
$targetRoot = Join-Path $repoRoot "frontend\public\skill-market"
$targetDetailDir = Join-Path $targetRoot "details"
$targetArchiveDir = Join-Path $targetRoot "dist\skills"
$targetTeleAgentArchiveDir = Join-Path $targetRoot "dist\teleagent"

if (-not (Test-Path -LiteralPath $marketIndex)) {
  throw "Missing market index: $marketIndex"
}

if (-not (Test-Path -LiteralPath $archiveDir)) {
  throw "Missing skill archive directory: $archiveDir"
}

if (-not (Test-Path -LiteralPath $teleagentArchiveDir)) {
  throw "Missing TeleAgent skill archive directory: $teleagentArchiveDir"
}

New-Item -ItemType Directory -Force -Path $targetArchiveDir | Out-Null
New-Item -ItemType Directory -Force -Path $targetTeleAgentArchiveDir | Out-Null
New-Item -ItemType Directory -Force -Path $targetDetailDir | Out-Null
Copy-Item -LiteralPath $marketIndex -Destination (Join-Path $targetRoot "index.json") -Force

Get-ChildItem -LiteralPath $targetArchiveDir -Filter *.zip | Remove-Item -Force
Copy-Item -Path (Join-Path $archiveDir "*.zip") -Destination $targetArchiveDir -Force

Get-ChildItem -LiteralPath $targetTeleAgentArchiveDir -Filter *.zip | Remove-Item -Force
Copy-Item -Path (Join-Path $teleagentArchiveDir "*.zip") -Destination $targetTeleAgentArchiveDir -Force

Get-ChildItem -LiteralPath $targetDetailDir -Filter *.md | Remove-Item -Force
if (Test-Path -LiteralPath $detailDir) {
  Copy-Item -Path (Join-Path $detailDir "*.md") -Destination $targetDetailDir -Force
}

$archiveCount = (Get-ChildItem -LiteralPath $targetArchiveDir -Filter *.zip | Measure-Object).Count
$teleagentArchiveCount = (Get-ChildItem -LiteralPath $targetTeleAgentArchiveDir -Filter *.zip | Measure-Object).Count
$detailCount = (Get-ChildItem -LiteralPath $targetDetailDir -Filter *.md | Measure-Object).Count
Write-Output "Synced Skill Market: index.json + $archiveCount standard archives + $teleagentArchiveCount TeleAgent archives + $detailCount details -> $targetRoot"
