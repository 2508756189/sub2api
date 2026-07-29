<template>
  <div ref="homeRoot" class="tp-home" :class="{ 'dark-mode': isDark }">
    <header class="tp-header" :class="{ scrolled }">
      <div class="nav-inner">
        <router-link to="/home" class="tp-logo" aria-label="TokenPort 首页" @click="closeMenu">
          <TokenPortLogo :src="brandLogo" />
        </router-link>

        <nav class="desktop-nav" aria-label="首页导航">
          <a href="#platform">核心能力</a>
          <a href="#cost">Token 经营</a>
          <a href="#skill-market">Skill Market</a>
          <a href="#deploy">部署交付</a>
        </nav>

        <div class="nav-actions">
          <button
            type="button"
            class="theme-control"
            :title="isDark ? '切换浅色模式' : '切换深色模式'"
            :aria-label="isDark ? '切换浅色模式' : '切换深色模式'"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="sm" :stroke-width="1.8" />
          </button>
          <div class="desktop-actions">
            <router-link :to="entryPath" class="nav-login">登录控制台</router-link>
            <router-link :to="entryPath" class="tp-button primary small">预约演示</router-link>
          </div>
          <button
            type="button"
            class="menu-button"
            :aria-expanded="menuOpen"
            :aria-label="menuOpen ? '关闭导航菜单' : '打开导航菜单'"
            @click="menuOpen = !menuOpen"
          >
            <Icon :name="menuOpen ? 'x' : 'menu'" size="md" :stroke-width="1.8" />
          </button>
        </div>
      </div>

      <div class="mobile-nav" :class="{ open: menuOpen }">
        <nav aria-label="移动端首页导航">
          <a href="#platform" @click="closeMenu">核心能力</a>
          <a href="#cost" @click="closeMenu">Token 经营</a>
          <a href="#skill-market" @click="closeMenu">Skill Market</a>
          <a href="#deploy" @click="closeMenu">部署交付</a>
          <router-link :to="entryPath" class="mobile-outline" @click="closeMenu">登录控制台</router-link>
          <router-link :to="entryPath" class="tp-button primary" @click="closeMenu">预约演示</router-link>
        </nav>
      </div>
    </header>

    <main>
      <section id="top" class="hero-section">
        <div class="hero-ambient" aria-hidden="true" />
        <div class="hero-grid" aria-hidden="true" />
        <div class="hero-layout">
          <div class="hero-copy reveal-item is-visible">
            <p class="hero-status"><i />企业级 AI 运营底座</p>
            <h1>
              <span>让模型资源可管理，</span>
              <span>让每一个 <em>Token</em> 有归属</span>
            </h1>
            <p class="hero-lead">
              TokenPort 智能应用与技能接入平台，把多家模型、部门密钥、Token 成本、Codex / Claude Code /
              OpenCode 等工具接入与 Skill Market 收敛到同一控制面——统一接入、统一治理、统一核算、统一交付。
            </p>
            <div class="hero-actions">
              <router-link :to="entryPath" class="tp-button primary large">预约演示</router-link>
              <a href="#platform" class="tp-button outline large">查看平台能力</a>
            </div>
            <dl class="hero-stats">
              <div><dt>5</dt><dd>核心能力</dd></div>
              <div><dt>{{ marketLoading ? '—' : skillCount }}</dt><dd>收录技能</dd></div>
              <div class="wide"><dt>OpenAI · Anthropic</dt><dd>协议兼容 · 替换地址即接入</dd></div>
            </dl>
          </div>
          <div class="hero-visual reveal-item is-visible">
            <ModelFlowPreview />
          </div>
        </div>
      </section>

      <section class="tool-strip" aria-label="支持的模型协议与客户端">
        <div class="tool-strip-inner">
          <p>统一你团队已在使用的模型与工具栈</p>
          <div>
            <span v-for="tool in tools" :key="tool">{{ tool }}</span>
          </div>
        </div>
      </section>

      <section id="platform" class="section-shell">
        <div class="section-heading reveal-item">
          <p class="section-label"><span>01</span><i />核心能力</p>
          <h2>面向 AI 供给到使用的同一控制面</h2>
          <p>模型调用、部门成本、工具接入与技能交付，统一在一处治理。</p>
        </div>
        <div class="feature-grid">
          <article
            v-for="(feature, index) in capabilities"
            :key="feature.title"
            class="feature-card reveal-item"
            :class="{ featured: index === 0 }"
          >
            <div class="feature-main">
              <div class="feature-icon">
                <Icon :name="feature.icon" size="md" :stroke-width="1.65" />
              </div>
              <h3>{{ feature.title }}</h3>
              <p>{{ feature.body }}</p>
            </div>
            <ul v-if="feature.highlights?.length">
              <li v-for="item in feature.highlights" :key="item"><i />{{ item }}</li>
            </ul>
          </article>
        </div>
      </section>

      <section id="cost" class="section-band">
        <div class="section-shell token-layout">
          <div class="token-copy reveal-item">
            <p class="section-label"><span>02</span><i />Token 经营</p>
            <h2>把每一次调用变成可核算的经营数据</h2>
            <p>
              按模型、部门与 API Key 实时计量 Token 与成本，设置预算、触发告警、执行额度上限——
              让管理者看得清费用由谁产生、为什么产生，为采购、预算与应用评估提供依据。
            </p>
            <div class="token-metrics">
              <div><span>成本 · Cost</span><b>{{ activeTokenData.spend }}</b></div>
              <div><span>Tokens</span><b>{{ activeTokenData.tokens }}</b></div>
            </div>
            <small>* 图表为演示数据，实际以部门试点归集口径为准</small>
          </div>

          <div class="token-panel reveal-item">
            <div class="token-panel-head">
              <div><b>Token 用量 vs 预算</b><span>全渠道汇总 · 实时</span></div>
              <div class="range-tabs" role="tablist" aria-label="Token 用量时间范围">
                <button
                  v-for="range in tokenRanges"
                  :key="range"
                  type="button"
                  role="tab"
                  :aria-selected="tokenRange === range"
                  :class="{ active: tokenRange === range }"
                  @click="tokenRange = range"
                >{{ range }}</button>
              </div>
            </div>
            <div class="chart-wrap">
              <svg viewBox="0 0 640 220" preserveAspectRatio="none" role="img" :aria-label="`${tokenRange} Token 用量趋势`">
                <defs>
                  <linearGradient id="tp-usage-area" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="var(--color-primary)" stop-opacity="0.35" />
                    <stop offset="100%" stop-color="var(--color-primary)" stop-opacity="0" />
                  </linearGradient>
                </defs>
                <line v-for="grid in [0.25, 0.5, 0.75]" :key="grid" x1="0" x2="640" :y1="220 * grid" :y2="220 * grid" class="chart-grid-line" />
                <line x1="0" x2="640" :y1="budgetY" :y2="budgetY" class="budget-line" />
                <path :d="chartPaths.area" fill="url(#tp-usage-area)" />
                <path :d="chartPaths.line" class="usage-line" />
              </svg>
              <span class="budget-label" :style="{ top: `${(1 - activeTokenData.budget / 100) * 100}%` }">预算 {{ activeTokenData.budget }}%</span>
            </div>
            <div class="chart-legend">
              <span><i class="actual" />实际用量</span>
              <span><i class="budget" />预算阈值</span>
            </div>
          </div>
        </div>
      </section>

      <section id="connectors" class="section-shell connector-section">
        <div class="section-heading reveal-item">
          <p class="section-label"><span>03</span><i />智能应用连接器</p>
          <h2>按向导复制配置，工具即刻接入</h2>
          <p>
            针对 Codex、Claude Code、OpenCode 等客户端生成接入参数、环境变量与诊断命令。员工只持有部门 API Key，
            不接触上游真实凭证；模型来自当前分组的真实可用范围，未同步时明确提示，不伪装可用状态。
          </p>
        </div>
        <ConnectorShowcase class="reveal-item" />
      </section>

      <section id="skill-market" class="section-band market-band">
        <div class="section-shell">
          <div class="market-heading reveal-item">
            <div class="section-heading">
              <p class="section-label"><span>04</span><i />Skill Market</p>
              <h2>能力沉淀一次，全组织复用</h2>
              <p>
                把验证有效的 Prompt、Skill、MCP 配置和工作流，沉淀为带版本、运行时、来源、风险等级和
                SHA256 校验的标准能力包。浏览器只生成可检查的安装脚本，不会静默修改本机。
              </p>
            </div>
            <span v-if="!marketLoading && !marketError" class="market-summary">
              {{ skillCount }} 个技能 · {{ marketCategories.length }} 个分类 · registry
            </span>
          </div>

          <div v-if="marketCategories.length" class="market-filters reveal-item" aria-label="Skill 分类筛选">
            <button type="button" :class="{ active: activeCategory === 'all' }" @click="activeCategory = 'all'">全部</button>
            <button
              v-for="category in marketCategories"
              :key="category.id"
              type="button"
              :class="{ active: activeCategory === category.id }"
              @click="activeCategory = category.id"
            >{{ category.name }}</button>
          </div>

          <div v-if="marketLoading" class="skill-grid loading-grid" aria-label="Skill Market 加载中">
            <article v-for="index in 6" :key="index" />
          </div>
          <div v-else-if="visibleSkills.length" class="skill-grid">
            <router-link
              v-for="skill in visibleSkills"
              :key="skill.id"
              to="/skill-market"
              class="skill-card-link reveal-item"
            >
              <SkillMarketCard :skill="skill" :registry="marketRegistry" variant="home">
                <template #action>
                  <span class="skill-card-action">
                    查看详情
                    <Icon name="arrowRight" size="xs" :stroke-width="2" />
                  </span>
                </template>
              </SkillMarketCard>
            </router-link>
          </div>
          <div v-else class="market-error">
            Skill Market registry 暂时无法加载（{{ marketError || '未知错误' }}）。
          </div>
          <p v-if="!marketLoading && filteredSkillCount > visibleSkills.length" class="market-count">
            当前分类共 {{ filteredSkillCount }} 个技能，控制台中查看完整目录
          </p>
        </div>
      </section>

      <section id="deploy" class="section-shell">
        <div class="section-heading reveal-item">
          <p class="section-label"><span>05</span><i />部署与交付</p>
          <h2>同一控制面，按你的方式交付</h2>
          <p>
            同一套控制面与 Skill Market——托管订阅、私有化自建，或面向合规环境的企业定制。兼容公有云模型、
            自建模型与企业内部模型服务。
          </p>
        </div>
        <div class="deployment-grid">
          <article
            v-for="mode in deployments"
            :key="mode.name"
            class="deployment-card reveal-item"
            :class="{ featured: mode.featured }"
          >
            <div class="deployment-head"><h3>{{ mode.name }}</h3><span>{{ mode.tag }}</span></div>
            <p>{{ mode.desc }}</p>
            <ul>
              <li v-for="point in mode.points" :key="point">
                <Icon name="check" size="sm" :stroke-width="2.2" />
                {{ point }}
              </li>
            </ul>
            <router-link :to="entryPath" class="deployment-action" :class="{ primary: mode.featured }">
              {{ mode.featured ? '联系方案团队' : '了解更多' }}
            </router-link>
          </article>
        </div>
      </section>

      <section class="section-shell cta-shell">
        <div class="final-cta reveal-item">
          <div class="cta-glow" aria-hidden="true" />
          <div class="cta-grid" aria-hidden="true" />
          <div class="cta-content">
            <p>开始使用</p>
            <h2>把模型、密钥与预算，收敛到同一个 Port</h2>
            <span>用一次面向你团队与合规需求的演示，看 TokenPort 如何统一治理企业 AI 供给、成本与能力资产。</span>
            <div>
              <router-link :to="entryPath" class="tp-button primary large">预约演示</router-link>
              <a :href="docUrl || '#platform'" class="tp-button outline large">阅读文档</a>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="tp-footer">
      <div class="footer-inner">
        <div class="footer-brand">
          <router-link to="/home" class="tp-logo">
            <TokenPortLogo :src="brandLogo" />
          </router-link>
          <p>TokenPort 智能应用与技能接入平台——统一接入、统一治理、统一核算、统一交付。</p>
        </div>
        <div v-for="column in footerColumns" :key="column.title" class="footer-column">
          <b>{{ column.title }}</b>
          <a v-for="link in column.links" :key="link.label" :href="link.href">{{ link.label }}</a>
        </div>
      </div>
      <div class="footer-bottom">
        <p>© {{ currentYear }} TokenPort · 智能应用与技能接入平台</p>
        <div><a href="#">隐私政策</a><a href="#">服务条款</a><a href="#">安全合规</a></div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import { useThemeToggle } from '@/composables/useThemeToggle'
import TokenPortLogo from '@/tokenport/brand/TokenPortLogo.vue'
import ConnectorShowcase from '@/tokenport/home/ConnectorShowcase.vue'
import ModelFlowPreview from '@/tokenport/home/ModelFlowPreview.vue'
import SkillMarketCard from '@/tokenport/market/SkillMarketCard.vue'
import {
  fetchSkillMarket,
  getSkillCategoryName,
} from '@/api/skillMarket'
import type { SkillMarketEntry, SkillMarketRegistry } from '@/api/skillMarket'

type TokenRange = '7D' | '30D' | '90D'
type HomeCapability = {
  title: string
  body: string
  highlights?: string[]
  icon: 'swap' | 'key' | 'chartBar' | 'terminal' | 'grid'
}

const props = withDefaults(defineProps<{ siteLogo?: string; docUrl?: string }>(), {
  siteLogo: '',
  docUrl: '',
})

const authStore = useAuthStore()
const appStore = useAppStore()
const { isDark, toggleTheme } = useThemeToggle()
const homeRoot = ref<HTMLElement | null>(null)
const menuOpen = ref(false)
const scrolled = ref(false)
const marketLoading = ref(true)
const marketError = ref('')
const skillCount = ref(0)
const marketRegistry = ref<SkillMarketRegistry | null>(null)
const marketCategories = ref<Array<{ id: string; name: string }>>([])
const activeCategory = ref('all')
const tokenRange = ref<TokenRange>('30D')
const featuredSkills = ref<SkillMarketEntry[]>([])
let revealObserver: IntersectionObserver | null = null

const docUrl = computed(() => props.docUrl)
const brandLogo = computed(() => props.siteLogo || '/tokenport-logo.svg')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const entryPath = computed(() => isAuthenticated.value
  ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
  : '/login')
const currentYear = new Date().getFullYear()

const tools = ['Codex', 'Claude Code', 'OpenCode', 'ChatGPT', 'Gemini CLI', 'OpenAI 兼容', 'Anthropic 兼容', '自建模型']

const capabilities: HomeCapability[] = [
  {
    title: '统一模型网关',
    body: '统一管理多上游 provider、账号池、模型映射与路由，对外只暴露一个 API 地址。兼容 OpenAI / Anthropic，替换地址与密钥即可接入。',
    highlights: ['多 provider 聚合', '模型映射 · 路由', '限流 · 故障切换', 'OpenAI / Anthropic 兼容'],
    icon: 'swap',
  },
  {
    title: '部门密钥与权限',
    body: '为部门 / 项目签发独立 API Key，绑定可用模型与额度；上游真实密钥集中托管，不下发终端。',
    icon: 'key',
  },
  {
    title: 'Token 计量与成本核算',
    body: '记录输入 / 输出 / 总 Token，关联价格、部门、Key 与时间，形成可对账的成本明细。',
    icon: 'chartBar',
  },
  {
    title: '智能应用连接器',
    body: '为 Codex、Claude Code、OpenCode 等客户端，按部门可用模型生成配置、环境变量与安装脚本。',
    icon: 'terminal',
  },
  {
    title: 'Skill Market',
    body: '以标准清单管理技能的版本、运行时、来源与风险等级，把个人经验沉淀为可审查、可升级的组织能力，支持私有化部署。',
    icon: 'grid',
  },
]

const tokenRanges: TokenRange[] = ['7D', '30D', '90D']
const tokenData: Record<TokenRange, { points: number[]; spend: string; tokens: string; budget: number }> = {
  '7D': { points: [42, 55, 48, 63, 58, 71, 66], spend: '¥12.8k', tokens: '312M', budget: 62 },
  '30D': { points: [30, 44, 39, 52, 60, 55, 68, 64, 72, 70, 78, 74], spend: '¥53.6k', tokens: '1.28B', budget: 71 },
  '90D': { points: [22, 34, 30, 41, 38, 50, 47, 58, 55, 66, 62, 70, 68, 76], spend: '¥146k', tokens: '3.6B', budget: 68 },
}
const activeTokenData = computed(() => tokenData[tokenRange.value])
const budgetY = computed(() => 220 * (1 - activeTokenData.value.budget / 100))
const chartPaths = computed(() => buildPath(activeTokenData.value.points, 640, 220))

const deployments = [
  {
    name: 'SaaS 订阅', tag: '最快起步', featured: false,
    desc: '托管控制面，零基础设施投入，按租户、Key 数、调用量或功能版本计费，适合中小企业和项目团队。',
    points: ['托管网关与弹性伸缩', '模型清单持续更新', '按用量灵活计费'],
  },
  {
    name: '私有化部署', tag: '密钥不出域', featured: true,
    desc: 'Docker Compose 组织网关、PostgreSQL 与 Redis，可在单机、内网或云环境部署，上游密钥服务端集中托管。',
    points: ['内网 / 离线环境部署', '自有模型与专属渠道', '权限隔离 · 日志审计 · 备份恢复'],
  },
  {
    name: '企业定制', tag: '按需交付', featured: false,
    desc: '面向政企与集团客户，提供专属路由、私有 Skill Market、行业能力包与运营分析服务。',
    points: ['架构评审与实施', '私有 Skill Market', '行业能力包 · 运营报表'],
  },
]

const footerColumns = [
  { title: '核心能力', links: [{ label: '统一模型网关', href: '#platform' }, { label: 'Token 计量核算', href: '#cost' }, { label: '智能应用连接器', href: '#connectors' }, { label: 'Skill Market', href: '#skill-market' }] },
  { title: '适用对象', links: [{ label: '企业管理者', href: '#cost' }, { label: '研发团队', href: '#connectors' }, { label: '业务部门', href: '#platform' }, { label: 'AI 运维团队', href: '#platform' }] },
  { title: '部署交付', links: [{ label: 'SaaS 订阅', href: '#deploy' }, { label: '私有化部署', href: '#deploy' }, { label: '企业定制', href: '#deploy' }, { label: '安全合规', href: '#deploy' }] },
  { title: '关于', links: [{ label: '产品介绍', href: '#platform' }, { label: '文档', href: docUrl.value || '#top' }, { label: '更新日志', href: '#top' }, { label: '联系我们', href: '#deploy' }] },
]

const visibleSkills = computed(() => featuredSkills.value
  .filter((skill) => activeCategory.value === 'all' || skill.category === activeCategory.value)
  .slice(0, 9))
const filteredSkillCount = computed(() => featuredSkills.value
  .filter((skill) => activeCategory.value === 'all' || skill.category === activeCategory.value).length)

function buildPath(points: number[], width: number, height: number) {
  const max = Math.max(...points) * 1.15
  const step = width / (points.length - 1)
  const coords = points.map((point, index) => [index * step, height - (point / max) * height])
  const line = coords.map(([x, y], index) => {
    if (index === 0) return `M ${x} ${y}`
    const [previousX, previousY] = coords[index - 1]
    const controlX = (previousX + x) / 2
    return `C ${controlX} ${previousY}, ${controlX} ${y}, ${x} ${y}`
  }).join(' ')
  return { line, area: `${line} L ${width} ${height} L 0 ${height} Z` }
}

function closeMenu() {
  menuOpen.value = false
}

function onScroll() {
  scrolled.value = window.scrollY > 8
}

function observeRevealItems() {
  if (!homeRoot.value || typeof IntersectionObserver === 'undefined') return
  revealObserver?.disconnect()
  revealObserver = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (!entry.isIntersecting) return
      entry.target.classList.add('is-visible')
      revealObserver?.unobserve(entry.target)
    })
  }, { threshold: 0.12, rootMargin: '0px 0px -60px' })
  homeRoot.value.querySelectorAll<HTMLElement>('.reveal-item:not(.is-visible)')
    .forEach((element) => revealObserver?.observe(element))
}

onMounted(async () => {
  window.addEventListener('scroll', onScroll, { passive: true })
  onScroll()
  await nextTick()
  observeRevealItems()
  await Promise.allSettled([authStore.checkAuth(), appStore.fetchPublicSettings()])
  try {
    const registry = await fetchSkillMarket()
    marketRegistry.value = registry
    skillCount.value = registry.skills.length
    marketCategories.value = registry.categories.map((category) => ({
      id: category.id,
      name: category.name || getSkillCategoryName(category.id, registry),
    }))
    featuredSkills.value = registry.skills
  } catch (error) {
    marketError.value = error instanceof Error ? error.message : '市场索引加载失败'
  } finally {
    marketLoading.value = false
    await nextTick()
    observeRevealItems()
  }
})

watch(activeCategory, async () => {
  await nextTick()
  observeRevealItems()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
  revealObserver?.disconnect()
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700;800&family=Noto+Sans+SC:wght@400;500;600;700;900&display=swap');

.tp-home {
  --color-ground: #ffffff;
  --color-ground-2: #f7faf8;
  --color-panel: #ffffff;
  --color-panel-2: #eaf7f1;
  --color-line: rgb(12 82 58 / 12%);
  --color-line-strong: rgb(12 82 58 / 24%);
  --color-fg: #10231c;
  --color-muted: #526d63;
  --color-faint: #71877e;
  --color-primary: #00a878;
  --color-primary-2: #008f66;
  --color-primary-fg: #ffffff;
  --color-primary-soft: rgb(0 168 120 / 10%);
  --color-accent: #168aad;
  --color-header: rgb(255 255 255 / 88%);
  --color-ambient: rgb(0 168 120 / 8%);
  --color-band: rgb(247 250 248 / 88%);
  --color-card: rgb(255 255 255 / 94%);
  --color-chip: #f1f7f4;
  --color-grid-dot: rgb(12 82 58 / 10%);
  --color-cta-dot: rgb(12 82 58 / 12%);
  --color-shadow: rgb(16 52 38 / 10%);
  --radius: 14px;
  min-height: 100vh;
  overflow-x: clip;
  background: var(--color-ground);
  color: var(--color-fg);
  font-family: 'Inter', 'Noto Sans SC', ui-sans-serif, system-ui, sans-serif;
  -webkit-font-smoothing: antialiased;
  text-rendering: optimizeLegibility;
}

.tp-home.dark-mode {
  --color-ground: #0a0e0d;
  --color-ground-2: #0e1513;
  --color-panel: #121b18;
  --color-panel-2: #16211d;
  --color-line: rgb(120 190 165 / 14%);
  --color-line-strong: rgb(120 190 165 / 26%);
  --color-fg: #e8f2ee;
  --color-muted: #8ca69c;
  --color-faint: #5f7a70;
  --color-primary: #2fd4a0;
  --color-primary-2: #23b98a;
  --color-primary-fg: #04140e;
  --color-primary-soft: rgb(47 212 160 / 11%);
  --color-accent: #4fd6e0;
  --color-header: rgb(10 14 13 / 80%);
  --color-ambient: rgb(47 212 160 / 10%);
  --color-band: rgb(14 21 19 / 40%);
  --color-card: rgb(18 27 24 / 50%);
  --color-chip: rgb(14 21 19 / 60%);
  --color-grid-dot: rgb(120 190 165 / 10%);
  --color-cta-dot: rgb(120 190 165 / 12%);
  --color-shadow: rgb(0 0 0 / 46%);
}

.tp-home *, .tp-home *::before, .tp-home *::after { box-sizing: border-box; }
.tp-home a { color: inherit; text-decoration: none; }
.tp-home button { font: inherit; }

.tp-header {
  position: fixed;
  inset: 0 0 auto;
  z-index: 50;
  border-bottom: 1px solid transparent;
  transition: background 0.25s ease, border-color 0.25s ease;
}
.tp-header.scrolled { border-color: var(--color-line); background: var(--color-header); box-shadow: 0 10px 30px -24px var(--color-shadow); backdrop-filter: blur(24px); }
.nav-inner { display: flex; align-items: center; justify-content: space-between; width: min(1152px, 100%); height: 64px; margin: 0 auto; padding: 0 32px; }
.tp-logo { display: inline-flex; align-items: center; }
.desktop-nav { display: flex; align-items: center; gap: 32px; }
.desktop-nav a, .nav-login { color: var(--color-muted); font-size: 14px; line-height: 20px; transition: color 0.18s ease; }
.desktop-nav a:hover, .nav-login:hover { color: var(--color-fg); }
.nav-actions, .desktop-actions { display: flex; align-items: center; gap: 8px; }
.nav-login { padding: 10px 12px; font-weight: 500; letter-spacing: 0; }
.theme-control, .menu-button { display: grid; width: 40px; height: 40px; flex: 0 0 40px; place-items: center; border: 1px solid var(--color-line-strong); border-radius: 10px; background: var(--color-card); color: var(--color-muted); cursor: pointer; transition: color 0.18s ease, border-color 0.18s ease, background 0.18s ease; }
.theme-control:hover, .menu-button:hover { border-color: var(--color-primary); background: var(--color-primary-soft); color: var(--color-primary); }
.tp-button { display: inline-flex; align-items: center; justify-content: center; gap: 8px; padding: 10px 20px; border: 0; border-radius: 10px; font-size: 14px; line-height: 20px; font-weight: 500; letter-spacing: 0; transition: transform 0.2s ease, background 0.2s ease, border-color 0.2s ease, color 0.2s ease, box-shadow 0.2s ease; }
.tp-button.primary { background: var(--color-primary); color: var(--color-primary-fg); box-shadow: 0 0 0 1px rgb(47 212 160 / 40%); }
.tp-button.primary:hover { transform: translateY(-1px); background: var(--color-primary-2); box-shadow: 0 8px 30px -8px rgb(47 212 160 / 55%); }
.tp-button.outline { border: 1px solid var(--color-line-strong); background: transparent; color: var(--color-fg); }
.tp-button.outline:hover { border-color: var(--color-primary); color: var(--color-primary); }
.tp-button.small { padding: 10px 20px; }
.tp-button.large { padding: 12px 24px; }
.menu-button { display: none; }
.mobile-nav { display: none; }

.hero-section { position: relative; overflow: hidden; padding: 144px 0 80px; }
.hero-ambient { position: absolute; top: 0; left: 50%; width: 820px; height: 520px; border-radius: 50%; background: var(--color-ambient); filter: blur(140px); transform: translateX(-50%); pointer-events: none; }
.hero-grid { position: absolute; inset: 0; background-image: linear-gradient(var(--color-line) 1px, transparent 1px), linear-gradient(90deg, var(--color-line) 1px, transparent 1px); background-size: 64px 64px; opacity: 0.4; mask-image: radial-gradient(ellipse 80% 60% at 50% 0%, #000 20%, transparent 75%); pointer-events: none; }
.hero-layout { position: relative; display: grid; grid-template-columns: 1.05fr 1fr; align-items: center; gap: 56px; width: min(1152px, 100%); margin: 0 auto; padding: 0 32px; }
.hero-copy { min-width: 0; }
.hero-status { display: flex; align-items: center; width: max-content; gap: 8px; margin: 0; padding: 6px 14px; border: 1px solid var(--color-line-strong); border-radius: 999px; background: var(--color-card); color: var(--color-muted); font: 500 12px/16px 'JetBrains Mono', 'Noto Sans SC', monospace; }
.hero-status i { width: 6px; height: 6px; border-radius: 50%; background: var(--color-primary); animation: tp-pulse 2s ease-in-out infinite; }
.hero-copy h1 { margin: 24px 0 0; font-size: 48px; line-height: 1.15; font-weight: 600; letter-spacing: 0; }
.hero-copy h1 span { display: block; white-space: nowrap; }
.hero-copy h1 em { color: var(--color-primary); font-style: normal; }
.hero-lead { max-width: 576px; margin: 24px 0 0; color: var(--color-muted); font-size: 18px; line-height: 1.625; }
.hero-actions { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 32px; }
.hero-stats { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 24px; max-width: 512px; margin: 48px 0 0; padding-top: 24px; border-top: 1px solid var(--color-line); }
.hero-stats div { min-width: 0; }
.hero-stats .wide { grid-column: span 2; }
.hero-stats dt { overflow: hidden; color: var(--color-fg); font: 700 24px/32px 'JetBrains Mono', monospace; text-overflow: ellipsis; white-space: nowrap; }
.hero-stats .wide dt { font-size: 18px; line-height: 28px; }
.hero-stats dd { margin: 4px 0 0; color: var(--color-faint); font-size: 12px; line-height: 1.375; white-space: nowrap; }
.hero-visual { min-width: 0; transition-delay: 120ms; }

.tool-strip { border-block: 1px solid var(--color-line); background: var(--color-band); }
.tool-strip-inner { width: min(1152px, 100%); margin: 0 auto; padding: 32px; }
.tool-strip p { margin: 0; text-align: center; color: var(--color-faint); font: 500 12px/16px 'JetBrains Mono', 'Noto Sans SC', monospace; letter-spacing: 0; text-transform: uppercase; }
.tool-strip-inner > div { display: flex; flex-wrap: wrap; align-items: center; justify-content: center; gap: 16px 32px; margin-top: 24px; }
.tool-strip span { color: color-mix(in srgb, var(--color-muted) 76%, transparent); font: 400 14px/20px 'JetBrains Mono', 'Noto Sans SC', monospace; transition: color 0.18s ease; }
.tool-strip span:hover { color: var(--color-fg); }

.section-shell { width: min(1152px, 100%); margin: 0 auto; padding: 96px 32px; }
.section-band { border-block: 1px solid var(--color-line); background: var(--color-band); }
.section-heading { max-width: 680px; }
.section-label { display: flex; align-items: center; gap: 12px; margin: 0; color: var(--color-muted); font-size: 12px; font-weight: 500; letter-spacing: 0; text-transform: uppercase; }
.section-label span { color: var(--color-faint); font-family: 'JetBrains Mono', monospace; letter-spacing: 0; }
.section-label i { width: 24px; height: 1px; background: var(--color-line-strong); }
.section-heading h2, .token-copy h2, .cta-content h2 { margin: 20px 0 0; font-size: 36px; line-height: 1.28; font-weight: 600; letter-spacing: 0; }
.section-heading > p:last-child, .token-copy > p { margin: 16px 0 0; color: var(--color-muted); font-size: 15px; line-height: 1.75; }

.feature-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); grid-auto-rows: 1fr; gap: 16px; margin-top: 48px; }
.feature-card { display: flex; flex-direction: column; min-height: 250px; padding: 24px; border: 1px solid var(--color-line); border-radius: var(--radius); background: var(--color-card); box-shadow: 0 16px 40px -32px var(--color-shadow); transition: transform 0.25s ease, background 0.25s ease, border-color 0.25s ease; }
.feature-card:hover { transform: translateY(-4px); border-color: var(--color-line-strong); background: var(--color-panel); }
.feature-card.featured { grid-column: span 2; }
.feature-card.featured { flex-direction: row; gap: 32px; }
.feature-card.featured .feature-main { flex: 1; }
.feature-card.featured ul { display: grid; width: 46%; align-self: center; gap: 10px; margin: 0; padding: 0; list-style: none; }
.feature-icon { display: grid; width: 44px; height: 44px; place-items: center; border: 1px solid var(--color-line-strong); border-radius: 11px; background: var(--color-ground-2); color: var(--color-primary); }
.feature-icon :deep(svg) { width: 20px; height: 20px; stroke-width: 1.6; stroke-linecap: round; stroke-linejoin: round; }
.feature-card h3 { margin: 20px 0 0; font-size: 18px; font-weight: 500; }
.feature-card p { margin: 10px 0 0; color: var(--color-muted); font-size: 14px; line-height: 1.7; }
.feature-card li { display: flex; align-items: center; gap: 10px; padding: 10px 12px; border: 1px solid var(--color-line); border-radius: 10px; background: var(--color-chip); color: var(--color-muted); font-size: 14px; }
.feature-card li i { width: 6px; height: 6px; flex: 0 0 6px; border-radius: 50%; background: var(--color-primary); }

.token-layout { display: grid; grid-template-columns: 1fr 1.4fr; align-items: center; gap: 48px; }
.token-copy small { display: block; margin-top: 12px; color: var(--color-faint); font: 400 11px/1.5 'JetBrains Mono', 'Noto Sans SC', monospace; }
.token-metrics { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; margin-top: 30px; }
.token-metrics div { padding: 16px; border: 1px solid var(--color-line); border-radius: 12px; background: var(--color-card); }
.token-metrics span { color: var(--color-faint); font: 400 12px/1.4 'JetBrains Mono', 'Noto Sans SC', monospace; letter-spacing: 0; text-transform: uppercase; }
.token-metrics b { display: block; margin-top: 8px; font: 700 24px/1.2 'JetBrains Mono', monospace; }
.token-panel { padding: 24px; border: 1px solid var(--color-line); border-radius: var(--radius); background: var(--color-card); box-shadow: 0 18px 44px -36px var(--color-shadow); }
.token-panel-head { display: flex; align-items: center; justify-content: space-between; gap: 18px; }
.token-panel-head > div:first-child { display: flex; flex-direction: column; gap: 4px; }
.token-panel-head b { font-size: 14px; font-weight: 500; }
.token-panel-head span { color: var(--color-faint); font: 400 12px/1.4 'JetBrains Mono', 'Noto Sans SC', monospace; }
.range-tabs { display: flex; gap: 2px; padding: 4px; border: 1px solid var(--color-line); border-radius: 10px; background: var(--color-ground); }
.range-tabs button { min-width: 42px; min-height: 30px; padding: 0 10px; border: 0; border-radius: 7px; background: transparent; color: var(--color-muted); font: 500 12px/1 'JetBrains Mono', monospace; cursor: pointer; }
.range-tabs button:hover { color: var(--color-fg); }
.range-tabs button.active { background: var(--color-primary); color: var(--color-primary-fg); }
.chart-wrap { position: relative; height: 220px; margin-top: 24px; }
.chart-wrap svg { width: 100%; height: 100%; }
.chart-grid-line { stroke: var(--color-line); stroke-width: 1; }
.budget-line { stroke: var(--color-accent); stroke-width: 1.2; stroke-dasharray: 5 5; opacity: 0.7; }
.usage-line { fill: none; stroke: var(--color-primary); stroke-width: 2.4; }
.budget-label { position: absolute; right: 0; color: var(--color-accent); font: 400 10px/1 'JetBrains Mono', 'Noto Sans SC', monospace; transform: translateY(-120%); }
.chart-legend { display: flex; flex-wrap: wrap; gap: 18px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--color-line); color: var(--color-faint); font: 400 12px/1.4 'JetBrains Mono', 'Noto Sans SC', monospace; }
.chart-legend span { display: flex; align-items: center; gap: 8px; }
.chart-legend i.actual { width: 8px; height: 8px; border-radius: 50%; background: var(--color-primary); }
.chart-legend i.budget { width: 12px; height: 8px; border: 1px dashed var(--color-accent); border-radius: 999px; }

.connector-section .section-heading { margin-bottom: 48px; }
.market-band { background: var(--color-band); }
.market-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 32px; }
.market-summary { flex: 0 0 auto; color: var(--color-faint); font: 400 12px/1.5 'JetBrains Mono', 'Noto Sans SC', monospace; }
.market-filters { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 40px; }
.market-filters button { min-height: 32px; padding: 6px 14px; border: 1px solid var(--color-line); border-radius: 999px; background: transparent; color: var(--color-muted); font: 400 12px/1 'JetBrains Mono', 'Noto Sans SC', monospace; cursor: pointer; transition: border-color 0.2s ease, color 0.2s ease, background 0.2s ease; }
.market-filters button:hover { border-color: var(--color-line-strong); color: var(--color-fg); }
.market-filters button.active { border-color: color-mix(in srgb, var(--color-primary) 50%, transparent); background: var(--color-primary-soft); color: var(--color-primary); }
.skill-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; margin-top: 32px; }
.skill-card-link { display: block; min-width: 0; }
.loading-grid article { height: 290px; border: 1px solid var(--color-line); border-radius: var(--radius); background: linear-gradient(90deg, var(--color-card), var(--color-panel), var(--color-card)); background-size: 200% 100%; animation: tp-loading 1.5s ease infinite; }
.market-error { margin-top: 32px; padding: 20px; border: 1px solid var(--color-line); border-radius: var(--radius); background: var(--color-card); color: var(--color-muted); font-size: 14px; }
.market-count { margin: 30px 0 0; text-align: center; color: var(--color-faint); font: 400 12px/1.5 'JetBrains Mono', 'Noto Sans SC', monospace; }

.deployment-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; margin-top: 48px; }
.deployment-card { display: flex; flex-direction: column; min-height: 380px; padding: 24px; border: 1px solid var(--color-line); border-radius: var(--radius); background: var(--color-card); transition: transform 0.25s ease, border-color 0.25s ease; }
.deployment-card:hover { transform: translateY(-4px); border-color: var(--color-line-strong); }
.deployment-card.featured { border-color: rgb(47 212 160 / 50%); background: var(--color-panel); box-shadow: 0 0 50px -20px rgb(47 212 160 / 50%); }
.deployment-head { display: flex; align-items: center; justify-content: space-between; gap: 10px; }
.deployment-head h3 { margin: 0; font-size: 18px; font-weight: 500; }
.deployment-head span { padding: 6px 10px; border: 1px solid var(--color-line-strong); border-radius: 999px; color: var(--color-faint); font: 400 10px/1 'JetBrains Mono', 'Noto Sans SC', monospace; letter-spacing: 0; }
.deployment-card.featured .deployment-head span { border-color: transparent; background: rgb(47 212 160 / 15%); color: var(--color-primary); }
.deployment-card > p { margin: 14px 0 0; color: var(--color-muted); font-size: 14px; line-height: 1.7; }
.deployment-card ul { display: grid; gap: 12px; margin: 24px 0; padding: 0; list-style: none; }
.deployment-card li { display: flex; align-items: flex-start; gap: 10px; color: var(--color-muted); font-size: 14px; }
.deployment-card li :deep(svg) { flex: 0 0 16px; color: var(--color-primary); }
.deployment-action { display: inline-flex; align-items: center; justify-content: center; min-height: 42px; margin-top: auto; border: 1px solid var(--color-line-strong); border-radius: 10px; font-size: 14px; font-weight: 500; }
.deployment-action:hover { border-color: var(--color-primary); color: var(--color-primary); }
.deployment-action.primary { border-color: var(--color-primary); background: var(--color-primary); color: var(--color-primary-fg); }

.cta-shell { padding-top: 0; }
.final-cta { position: relative; overflow: hidden; padding: 64px 40px; border: 1px solid var(--color-line-strong); border-radius: 24px; background: var(--color-ground-2); text-align: center; }
.cta-glow { position: absolute; top: 50%; left: 50%; width: 600px; height: 400px; border-radius: 50%; background: var(--color-ambient); filter: blur(120px); transform: translate(-50%, -50%); }
.cta-grid { position: absolute; inset: 0; background-image: radial-gradient(circle, var(--color-cta-dot) 1px, transparent 1.4px); background-size: 22px 22px; opacity: 0.35; }
.cta-content { position: relative; max-width: 720px; margin: 0 auto; }
.cta-content > p { margin: 0; color: var(--color-primary); font-size: 14px; font-weight: 500; letter-spacing: 0; }
.cta-content h2 { font-size: 36px; }
.cta-content > span { display: block; max-width: 600px; margin: 16px auto 0; color: var(--color-muted); font-size: 15px; line-height: 1.7; }
.cta-content > div { display: flex; flex-wrap: wrap; justify-content: center; gap: 12px; margin-top: 32px; }

.tp-footer { border-top: 1px solid var(--color-line); background: var(--color-band); }
.footer-inner { display: grid; grid-template-columns: 1.4fr repeat(4, 1fr); gap: 40px; width: min(1152px, 100%); margin: 0 auto; padding: 64px 32px; }
.footer-brand p { max-width: 300px; margin: 16px 0 0; color: var(--color-faint); font-size: 14px; line-height: 1.7; }
.footer-column { display: flex; flex-direction: column; align-items: flex-start; gap: 11px; }
.footer-column b { margin-bottom: 5px; color: var(--color-muted); font: 400 12px/1.4 'JetBrains Mono', 'Noto Sans SC', monospace; letter-spacing: 0; text-transform: uppercase; }
.footer-column a { color: var(--color-faint); font-size: 14px; transition: color 0.18s ease; }
.footer-column a:hover { color: var(--color-fg); }
.footer-bottom { display: flex; align-items: center; justify-content: space-between; gap: 20px; width: min(1152px, 100%); margin: 0 auto; padding: 24px 32px 32px; border-top: 1px solid var(--color-line); }
.footer-bottom p, .footer-bottom a { color: var(--color-faint); font: 400 12px/1.4 'JetBrains Mono', 'Noto Sans SC', monospace; }
.footer-bottom p { margin: 0; }
.footer-bottom div { display: flex; gap: 24px; }
.footer-bottom a:hover { color: var(--color-fg); }

.reveal-item { opacity: 0; transform: translateY(24px); transition: opacity 0.7s cubic-bezier(0.22, 1, 0.36, 1), transform 0.7s cubic-bezier(0.22, 1, 0.36, 1); }
.reveal-item.is-visible { opacity: 1; transform: none; }

@keyframes tp-pulse { 0%, 100% { opacity: 0.35; } 50% { opacity: 1; } }
@keyframes tp-loading { to { background-position: -200% 0; } }

@media (max-width: 1020px) {
  .desktop-nav { gap: 20px; }
  .hero-layout { grid-template-columns: 1fr; }
  .hero-copy { max-width: 760px; }
  .hero-visual { width: min(760px, 100%); margin: 0 auto; }
  .token-layout { grid-template-columns: 1fr; }
  .feature-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .feature-card.featured { grid-column: 1 / -1; }
  .skill-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 767px) {
  .desktop-nav, .desktop-actions { display: none; }
  .menu-button { position: relative; display: grid; }
  .mobile-nav { display: block; max-height: 0; overflow: hidden; border-bottom: 1px solid transparent; background: var(--color-header); backdrop-filter: blur(24px); opacity: 0; pointer-events: none; visibility: hidden; transition: max-height 0.3s ease, opacity 0.2s ease, border-color 0.2s ease, visibility 0s linear 0.3s; }
  .mobile-nav.open { max-height: 420px; border-color: var(--color-line); opacity: 1; pointer-events: auto; visibility: visible; transition-delay: 0s; }
  .mobile-nav nav { display: flex; flex-direction: column; gap: 4px; width: min(1152px, 100%); margin: 0 auto; padding: 16px 20px 20px; }
  .mobile-nav nav > a { min-height: 44px; padding: 13px 12px; border-radius: 9px; color: var(--color-muted); font-size: 14px; }
  .mobile-nav nav > a:hover { background: var(--color-panel); color: var(--color-fg); }
  .mobile-nav .mobile-outline { margin-top: 8px; border: 1px solid var(--color-line-strong); text-align: center; color: var(--color-fg); }
  .mobile-nav .tp-button { color: var(--color-primary-fg); }
  .market-heading { align-items: flex-start; flex-direction: column; }
  .market-summary { white-space: normal; }
  .feature-grid, .deployment-grid { grid-template-columns: 1fr; }
  .feature-card.featured { grid-column: auto; flex-direction: column; }
  .feature-card.featured ul { width: 100%; margin-top: 22px; }
  .footer-inner { grid-template-columns: repeat(2, 1fr); }
  .footer-brand { grid-column: 1 / -1; }
}

@media (max-width: 639px) {
  .nav-inner { padding-inline: 20px; }
  .hero-section { padding: 112px 0 80px; }
  .hero-layout, .section-shell { padding-inline: 20px; }
  .hero-copy h1 { font-size: 36px; }
  .hero-copy h1 span { white-space: normal; }
  .hero-lead { font-size: 16px; }
  .hero-stats { grid-template-columns: repeat(2, 1fr); }
  .hero-stats .wide { grid-column: 1 / -1; }
  .tool-strip-inner { padding-inline: 20px; }
  .section-heading h2, .token-copy h2, .cta-content h2 { font-size: 30px; }
  .skill-grid { grid-template-columns: 1fr; }
  .token-panel-head { align-items: flex-start; flex-direction: column; }
  .chart-wrap { height: 190px; }
  .final-cta { padding: 48px 24px; }
  .cta-shell { padding-top: 0; }
  .footer-inner { padding: 48px 20px; }
  .footer-bottom { align-items: flex-start; flex-direction: column; padding-inline: 20px; }
}

@media (max-width: 440px) {
  .hero-copy h1 { font-size: 33px; }
  .hero-copy h1 span { white-space: nowrap; }
  .tool-strip-inner > div { gap: 12px 18px; }
  .token-metrics { grid-template-columns: 1fr; }
  .range-tabs button { min-width: 38px; padding-inline: 8px; }
  .skill-card-link { min-height: 0; }
  .deployment-card { min-height: 350px; }
  .footer-inner { gap: 32px 24px; }
  .footer-bottom div { flex-wrap: wrap; gap: 12px 20px; }
}

@media (max-width: 360px) {
  .hero-copy h1 { font-size: 29px; }
}

@media (prefers-reduced-motion: reduce) {
  .tp-home { scroll-behavior: auto; }
  .reveal-item { opacity: 1; transform: none; transition: none; }
  .hero-status i, .loading-grid article { animation: none; }
}
</style>
