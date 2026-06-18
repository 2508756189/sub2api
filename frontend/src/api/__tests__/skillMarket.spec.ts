import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  fetchSkillMarketWithSource,
  resolveSkillArchiveUrl,
  toSkillInstallSelection,
  type SkillMarketRegistry,
} from '../skillMarket'

const registry: SkillMarketRegistry = {
  schemaVersion: '1.0',
  generatedAt: '2026-06-18T00:00:00Z',
  categories: [{ id: 'engineering', name: 'Engineering' }],
  skills: [
    {
      id: 'markitdown',
      name: 'markitdown',
      description: 'Convert files to Markdown',
      category: 'engineering',
      installTargets: {
        claude: '~/.claude/skills/markitdown',
        codex: '~/.codex/skills/markitdown',
      },
      version: '0.1.0',
      riskLevel: 'medium',
      archive: {
        path: 'dist/skills/markitdown.zip',
        sha256: '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef',
      },
    },
  ],
}

describe('skillMarket', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('falls back to the bundled same-origin registry when remote registries fail', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: false, status: 404 })
      .mockResolvedValueOnce({ ok: false, status: 404 })
      .mockResolvedValueOnce({ ok: true, json: async () => registry })

    vi.stubGlobal('fetch', fetchMock)

    const result = await fetchSkillMarketWithSource()

    expect(result.registryUrl).toBe('/skill-market/index.json')
    expect(result.registry.skills).toHaveLength(1)
    expect(fetchMock).toHaveBeenCalledTimes(3)
  })

  it('resolves archive paths relative to the loaded registry source', () => {
    const archiveUrl = resolveSkillArchiveUrl('/skill-market/index.json', 'dist/skills/markitdown.zip')
    const selection = toSkillInstallSelection(registry.skills[0], '/skill-market/index.json')

    expect(archiveUrl).toContain('/skill-market/dist/skills/markitdown.zip')
    expect(selection.archiveUrl).toContain('/skill-market/dist/skills/markitdown.zip')
    expect(selection.sha256).toBe(registry.skills[0].archive.sha256)
  })
})
