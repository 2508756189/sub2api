export const DEFAULT_SKILL_MARKET_REGISTRY_URL = '/skill-market/index.json'

export const SKILL_MARKET_REGISTRY_FALLBACK_URLS = [
  'https://cdn.jsdelivr.net/gh/2508756189/state-of-art-skills@main/market/index.json',
  'https://raw.githubusercontent.com/2508756189/state-of-art-skills/main/market/index.json',
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

const SKILL_DISPLAY_TEXT: Record<string, { name: string; description: string }> = {
  'anbeime-agent-team': {
    name: '多智能体团队框架',
    description: '用于组织多个角色型智能体协作，适合方案评审、系统构建、任务分派和团队协同工作流。',
  },
  'anbeime-content-research-writer': {
    name: '内容研究写作助手',
    description: '面向调研写作、资料引用、提纲迭代和章节润色，适合报告、文章和方案材料生产。',
  },
  'anbeime-frontend-design': {
    name: '前端体验设计',
    description: '用于构建更有质感的前端页面和交互界面，强调布局、视觉层级、组件细节和产品气质。',
  },
  'anbeime-multi-agent-meeting': {
    name: '多智能体会议',
    description: '模拟多个专业角色开会讨论，适合项目决策、技术方案、商业策略和风险评估场景。',
  },
  'anbeime-pptx-generator': {
    name: 'PPT 生成器',
    description: '将结构化内容生成可编辑 PowerPoint，适合路演汇报、项目计划、方案展示和演示材料。',
  },
  'anbeime-product-manager-toolkit': {
    name: '产品经理工具包',
    description: '覆盖 PRD、RICE 优先级、用户访谈、需求分析和产品路线规划，适合产品与运营团队。',
  },
  'anbeime-web-design-analyzer': {
    name: '网页设计分析器',
    description: '从网页截图中提取设计系统、配色、排版和组件风格，并生成可用于前端实现的提示词。',
  },
  compound: {
    name: '复合工程工作流',
    description: '用于多步骤工程任务的计划、实现、复查和验证，适合较复杂的代码交付流程。',
  },
  ecc: {
    name: '跨 Agent 兼容套件',
    description: '维护 Codex、Claude Code、Gemini CLI 等不同智能体运行时之间的技能和提示兼容。',
  },
  headroom: {
    name: '上下文余量管理',
    description: '用于长任务中的上下文压力控制、资料取舍、记忆压缩和执行节奏管理。',
  },
  markitdown: {
    name: '文档转 Markdown',
    description: '将 Office、PDF、图片、音频、网页和压缩包等资料转换为干净 Markdown，便于 AI 读取和加工。',
  },
  supermemory: {
    name: '长期记忆与知识检索',
    description: '用于沉淀项目事实、决策和知识片段，并在后续工作中检索复用，同时避免泄露敏感信息。',
  },
  'taste-skill': {
    name: '产品品味优化',
    description: '用于审视和提升界面质感、信息密度、文案表达和交互细节，让产品更像真实业务工具。',
  },
}

const CATEGORY_DISPLAY_TEXT: Record<string, string> = {
  engineering: '工程研发',
  product: '产品与研究',
  design: '设计体验',
  knowledge: '知识与记忆',
  workflow: '组织协作',
}

const RISK_LEVEL_TEXT: Record<string, string> = {
  low: '低风险',
  medium: '中风险',
  high: '高风险',
}

export function getSkillDisplayName(skill: SkillMarketEntry): string {
  return SKILL_DISPLAY_TEXT[skill.id]?.name || skill.name
}

export function getSkillDisplayDescription(skill: SkillMarketEntry): string {
  return SKILL_DISPLAY_TEXT[skill.id]?.description || skill.description
}

export function getSkillCategoryName(categoryId: string, registry?: SkillMarketRegistry | null): string {
  return registry?.categories.find((category) => category.id === categoryId)?.name ||
    CATEGORY_DISPLAY_TEXT[categoryId] ||
    categoryId
}

export function getSkillRiskLabel(riskLevel?: string): string {
  return RISK_LEVEL_TEXT[riskLevel || 'low'] || riskLevel || '低风险'
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
    name: getSkillDisplayName(skill),
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
