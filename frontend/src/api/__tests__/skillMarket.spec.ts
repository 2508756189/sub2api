import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  fetchSkillMarketWithSource,
  fetchSkillDetailMarkdown,
  getSkillDisplayDescription,
  getSkillDisplayName,
  resolveSkillArchiveUrl,
  skillSupportsRuntime,
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
      detail: {
        summary: 'Convert documents',
        useCases: ['Knowledge ingestion'],
        capabilities: ['Conversion'],
        requirements: [],
        permissions: ['Read files'],
        markdownPath: 'details/markitdown.md',
      },
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

  it('loads the bundled same-origin registry by default', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: async () => registry })

    vi.stubGlobal('fetch', fetchMock)

    const result = await fetchSkillMarketWithSource()

    expect(result.registryUrl).toBe('/skill-market/index.json')
    expect(result.registry.skills).toHaveLength(1)
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('falls back to remote registries when the bundled registry is unavailable', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: false, status: 404 })
      .mockResolvedValueOnce({ ok: true, json: async () => registry })

    vi.stubGlobal('fetch', fetchMock)

    const result = await fetchSkillMarketWithSource()

    expect(result.registryUrl).toBe('https://cdn.jsdelivr.net/gh/2508756189/state-of-art-skills@main/market/index.json')
    expect(result.registry.skills).toHaveLength(1)
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })

  it('resolves archive paths relative to the loaded registry source', () => {
    const archiveUrl = resolveSkillArchiveUrl('/skill-market/index.json', 'dist/skills/markitdown.zip')
    const selection = toSkillInstallSelection(registry.skills[0], '/skill-market/index.json')

    expect(archiveUrl).toContain('/skill-market/dist/skills/markitdown.zip')
    expect(selection.archiveUrl).toContain('/skill-market/dist/skills/markitdown.zip')
    expect(selection.sha256).toBe(registry.skills[0].archive.sha256)
  })

  it('loads detail markdown relative to the active registry source', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, text: async () => '# MarkItDown' })
    vi.stubGlobal('fetch', fetchMock)

    const markdown = await fetchSkillDetailMarkdown(registry.skills[0], '/skill-market/index.json')

    expect(markdown).toBe('# MarkItDown')
    expect(fetchMock.mock.calls[0][0]).toContain('/skill-market/details/markitdown.md')
  })

  it('matches Codex and Claude filters through install targets', () => {
    const portableSkill = {
      ...registry.skills[0],
      runtime: ['portable'],
    }

    expect(skillSupportsRuntime(portableSkill, 'codex')).toBe(true)
    expect(skillSupportsRuntime(portableSkill, 'claude')).toBe(true)
    expect(skillSupportsRuntime(portableSkill, 'portable')).toBe(true)
  })

  it('does not claim support for a runtime without metadata or an install target', () => {
    const portableOnlySkill = {
      ...registry.skills[0],
      runtime: ['portable'],
      installTargets: { portable: '~/.agents/skills/markitdown' },
    }

    expect(skillSupportsRuntime(portableOnlySkill, 'codex')).toBe(false)
    expect(skillSupportsRuntime(portableOnlySkill, 'claude')).toBe(false)
  })

  it('prefers registry-localized display text and keeps legacy fallback support', () => {
    const localized = {
      ...registry.skills[0],
      id: 'localized-diagram-skill',
      name: 'dashmotion',
      description: 'Create animated technical diagrams',
      detail: {
        ...registry.skills[0].detail!,
        displayName: '动态技术图生成器',
        summary: '将流程和架构生成可离线打开的动态图。',
      },
    }

    expect(getSkillDisplayName(localized)).toBe('动态技术图生成器')
    expect(getSkillDisplayDescription(localized)).toBe('将流程和架构生成可离线打开的动态图。')
    expect(getSkillDisplayName(registry.skills[0])).toBe('文档转 Markdown')
    expect(getSkillDisplayDescription(registry.skills[0])).toBe('将 Office、PDF、图片、音频、网页和压缩包等资料转换为干净 Markdown，便于 AI 读取和加工。')
  })
})
