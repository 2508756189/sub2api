export const DEFAULT_SKILL_MARKET_REGISTRY_URL =
  'https://cdn.jsdelivr.net/gh/2508756189/state-of-art-skills@main/market/index.json'

export const SKILL_MARKET_REGISTRY_FALLBACK_URLS = [
  'https://raw.githubusercontent.com/2508756189/state-of-art-skills/main/market/index.json',
  '/skill-market/index.json',
]

export interface SkillMarketCategory {
  id: string
  name: string
  description?: string
}

export interface SkillMarketEntry {
  id: string
  name: string
  description: string
  category: string
  tags?: string[]
  runtime?: string[]
  installTargets: {
    codex?: string
    claude?: string
    portable?: string
  }
  version: string
  license?: string
  source?: string
  riskLevel?: 'low' | 'medium' | 'high' | string
  archive: {
    path: string
    sha256: string
    size?: number
  }
}

export interface SkillMarketRegistry {
  schemaVersion: string
  generatedAt: string
  repository?: string
  categories: SkillMarketCategory[]
  skills: SkillMarketEntry[]
}

export interface SkillInstallSelection {
  id: string
  name: string
  archiveUrl: string
  sha256: string
  installTargets: {
    codex?: string
    claude?: string
    portable?: string
  }
}

export function resolveSkillArchiveUrl(registryUrl: string, archivePath: string): string {
  if (/^https?:\/\//i.test(archivePath)) {
    return archivePath
  }
  const baseUrl = /^https?:\/\//i.test(registryUrl)
    ? registryUrl
    : new URL(registryUrl, globalThis.location?.origin || 'http://localhost').toString()
  return new URL(archivePath.replace(/^\/+/, ''), baseUrl).toString()
}

export function toSkillInstallSelection(
  skill: SkillMarketEntry,
  registryUrl = DEFAULT_SKILL_MARKET_REGISTRY_URL,
): SkillInstallSelection {
  return {
    id: skill.id,
    name: skill.name,
    archiveUrl: resolveSkillArchiveUrl(registryUrl, skill.archive.path),
    sha256: skill.archive.sha256,
    installTargets: skill.installTargets,
  }
}

export async function fetchSkillMarket(
  registryUrl = DEFAULT_SKILL_MARKET_REGISTRY_URL,
): Promise<SkillMarketRegistry> {
  const result = await fetchSkillMarketWithSource(registryUrl)
  return result.registry
}

export async function fetchSkillMarketWithSource(
  registryUrl = DEFAULT_SKILL_MARKET_REGISTRY_URL,
): Promise<{ registry: SkillMarketRegistry; registryUrl: string }> {
  const urls = [registryUrl, ...SKILL_MARKET_REGISTRY_FALLBACK_URLS]
    .filter((url, index, all) => all.indexOf(url) === index)
  const errors: string[] = []

  for (const url of urls) {
    try {
      const registry = await fetchOneSkillMarket(url)
      return { registry, registryUrl: url }
    } catch (error) {
      errors.push(`${url}: ${error instanceof Error ? error.message : String(error)}`)
    }
  }

  throw new Error(`Failed to load skill market registry. ${errors.join(' | ')}`)
}

async function fetchOneSkillMarket(registryUrl: string): Promise<SkillMarketRegistry> {
  const response = await fetch(registryUrl, { cache: 'no-store' })
  if (!response.ok) {
    throw new Error(`Failed to load skill market registry: ${response.status}`)
  }
  return response.json() as Promise<SkillMarketRegistry>
}
