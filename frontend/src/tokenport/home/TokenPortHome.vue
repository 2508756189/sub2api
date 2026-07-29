<template>
  <div ref="homeRoot" class="home-shell">
    <header class="topbar">
      <router-link to="/home" class="brand-link">
        <img :src="brandLogo" alt="天翼云 TokenPort" />
        <span>
          <b>{{ siteName }}</b>
          <small>{{ siteSubtitle }}</small>
        </span>
      </router-link>

      <button
        type="button"
        class="menu-control"
        :aria-expanded="menuOpen"
        :aria-label="menuOpen ? '关闭导航菜单' : '打开导航菜单'"
        @click="menuOpen = !menuOpen"
      >
        <Icon :name="menuOpen ? 'x' : 'menu'" size="md" />
      </button>

      <nav :class="{ open: menuOpen }" aria-label="首页导航">
        <a href="#platform" class="nav-link nav-link-optional" @click="closeMenu">核心能力</a>
        <a href="#cost" class="nav-link nav-link-optional" @click="closeMenu">Token 经营</a>
        <router-link to="/skill-market" class="nav-link" @click="closeMenu">Skill Market</router-link>
        <a href="#deploy" class="nav-link nav-link-optional" @click="closeMenu">部署交付</a>
        <router-link
          to="/available-channels"
          class="nav-link nav-link-optional"
          :class="{ 'is-gated': !isAuthenticated }"
          :title="isAuthenticated ? undefined : '需登录后查看，登录成功将回到此页'"
          @click="closeMenu"
        >
          模型与渠道
          <Icon v-if="!isAuthenticated" name="lock" size="sm" class="gate-icon" />
        </router-link>
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="nav-link nav-link-optional"
          @click="closeMenu"
        >Docs</a>
        <button
          type="button"
          class="icon-control"
          :title="isDark ? '切换浅色模式' : '切换深色模式'"
          :aria-label="isDark ? '切换浅色模式' : '切换深色模式'"
          @click="toggleTheme"
        >
          <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
        </button>
        <router-link :to="entryPath" class="primary-link compact" @click="closeMenu">
          {{ isAuthenticated ? '进入控制台' : '登录平台' }}
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </nav>
    </header>

    <main>
      <section class="hero-band">
        <div class="hero-grid" aria-hidden="true" />
        <div class="hero-inner reveal-item is-visible">
          <div class="hero-copy">
            <p class="status-label"><i />企业级 AI 运营底座</p>
            <h1>
              <span>让模型资源可管理，</span>
              <span>让每一个 <em>Token</em> 有归属</span>
            </h1>
            <p class="lead">
              TokenPort 将多家模型、部门密钥、Token 成本、智能应用接入与 Skill Market 收敛到同一平台，形成统一接入、统一治理、统一核算和统一交付。
            </p>
            <div class="hero-actions">
              <router-link :to="entryPath" class="primary-link large">
                {{ isAuthenticated ? '进入控制台' : '开始使用' }}
                <Icon name="arrowRight" size="md" />
              </router-link>
              <a href="#platform" class="secondary-link">查看平台能力</a>
            </div>
            <dl class="hero-stats">
              <div>
                <dt>5</dt>
                <dd>核心能力</dd>
              </div>
              <div>
                <dt>{{ marketLoading ? '—' : skillCount }}</dt>
                <dd>收录 Skill</dd>
              </div>
              <div class="wide">
                <dt>OpenAI · Anthropic · Gemini</dt>
                <dd>协议兼容 · 统一入口接入</dd>
              </div>
            </dl>
          </div>
          <ModelFlowPreview />
        </div>
      </section>

      <section class="tool-strip reveal-item" aria-label="支持的模型协议与客户端">
        <div class="tool-strip-inner">
          <p class="tool-strip-label"><i />统一团队正在使用的模型与工具栈</p>
          <div class="tool-list">
            <span v-for="item in TOKENPORT_PRODUCT.clients" :key="item">{{ item }}</span>
            <span v-for="item in TOKENPORT_PRODUCT.protocols" :key="item">{{ item }} 兼容</span>
            <span>企业自有模型</span>
          </div>
        </div>
      </section>

      <section id="platform" class="content-section">
        <div class="section-inner reveal-item">
          <div class="section-heading">
            <p class="section-label"><span>01</span> 核心能力</p>
            <h2>面向 AI 供给到使用的同一控制面</h2>
            <p>模型调用、部门成本、工具接入与技能交付，在一处完成治理。</p>
          </div>
          <div class="capability-grid">
            <article
              v-for="(item, index) in capabilities"
              :key="item.title"
              class="reveal-item"
              :class="{ featured: index === 0 }"
            >
              <div class="capability-icon" v-html="item.icon" />
              <div class="capability-copy">
                <h3>{{ item.title }}</h3>
                <p>{{ item.description }}</p>
              </div>
              <ul v-if="item.points.length">
                <li v-for="point in item.points" :key="point"><i />{{ point }}</li>
              </ul>
            </article>
          </div>
        </div>
      </section>

      <section id="cost" class="content-section section-alt">
        <div class="section-inner split-section reveal-item">
          <div class="section-heading">
            <p class="section-label"><span>02</span> Token 经营</p>
            <h2>把每一次调用变成可核算的经营数据</h2>
            <p>
              按模型、部门、用户与 API Key 记录用量和成本，为预算控制、采购评估和资源优化提供统一口径。
            </p>
            <div class="value-points">
              <div><b>成本归属</b><span>每笔消耗对应部门、Key 与模型</span></div>
              <div><b>预算护栏</b><span>按周期设置额度并提前预警</span></div>
              <div><b>模型优化</b><span>比较价格、质量和真实可用性</span></div>
            </div>
          </div>
          <TokenUsageShowcase />
        </div>
      </section>

      <section class="content-section console-section">
        <div class="section-inner reveal-item">
          <div class="section-heading horizontal">
            <div>
              <p class="section-label"><span>03</span> 平台控制台</p>
              <h2>从模型资源到部门用量，一套界面统一管理</h2>
            </div>
            <p>下方为交互式能力预览，可切换菜单查看用户、渠道、账号、用量、Skill 与 API Key 的管理结构。</p>
          </div>
          <div class="console-stage">
            <ConsolePreview :logo-src="brandLogo" />
          </div>
        </div>
      </section>

      <section id="connectors" class="content-section section-alt">
        <div class="section-inner reveal-item">
          <div class="section-heading connector-heading">
            <p class="section-label"><span>04</span> 智能应用连接器</p>
            <h2>选择客户端，生成可以检查的接入配置</h2>
            <p>
              按当前 API Key 的协议、可用模型和部门策略生成配置；模型可留空，现有文件先备份再合并，Skill 安装脚本独立可审查。
            </p>
          </div>
          <ConnectorShowcase />
        </div>
      </section>

      <section id="skill-market" class="content-section market-section">
        <div class="section-inner reveal-item">
          <div class="section-heading horizontal">
            <div>
              <p class="section-label"><span>05</span> Skill Market</p>
              <h2>能力沉淀一次，在团队与客户项目中持续复用</h2>
            </div>
            <router-link to="/skill-market" class="text-link">
              查看完整市场
              <Icon name="arrowRight" size="sm" />
            </router-link>
          </div>
          <p class="market-intro">
            每个能力包都标注版本、来源、运行时与风险等级，并附带 SHA256 校验。浏览器不会静默安装，需用户确认后再执行可审查的安装脚本。
          </p>
          <div v-if="marketCategories.length" class="market-categories" aria-label="Skill 分类筛选">
            <button
              type="button"
              :class="{ active: activeCategory === 'all' }"
              @click="activeCategory = 'all'"
            >全部</button>
            <button
              v-for="category in marketCategories"
              :key="category.id"
              type="button"
              :class="{ active: activeCategory === category.id }"
              @click="activeCategory = category.id"
            >{{ category.name }}</button>
          </div>
          <div v-if="marketLoading" class="market-grid loading-grid" aria-label="Skill Market 加载中">
            <article v-for="index in 6" :key="index" />
          </div>
          <div v-else-if="visibleSkills.length" class="market-grid">
            <article v-for="skill in visibleSkills" :key="skill.id" class="reveal-item">
              <div class="skill-meta">
                <span>{{ skill.category }}</span>
                <em>{{ skill.risk }}</em>
              </div>
              <h3>{{ skill.name }}</h3>
              <p>{{ skill.description }}</p>
              <footer><span>skill/{{ skill.id }}</span><b>v{{ skill.version }}</b></footer>
            </article>
          </div>
          <div v-else class="market-empty">
            <b>Skill Market 暂时无法加载</b>
            <span>{{ marketError || '请稍后重试，或进入市场页面查看内置目录。' }}</span>
          </div>
        </div>
      </section>

      <section class="content-section architecture-section">
        <div class="section-inner reveal-item">
          <div class="section-heading">
            <p class="section-label"><span>06</span> 系统架构</p>
            <h2>一个平台连接企业用户、AI 工具与模型资源</h2>
          </div>
          <div class="architecture">
            <div class="arch-column">
              <b>使用入口</b>
              <span v-for="item in entryNodes" :key="item">{{ item }}</span>
            </div>
            <div class="flow-arrow" aria-hidden="true"><span /></div>
            <div class="arch-core">
              <em>CONTROL PLANE</em>
              <b>TokenPort</b>
              <span v-for="item in coreNodes" :key="item">{{ item }}</span>
            </div>
            <div class="flow-arrow" aria-hidden="true"><span /></div>
            <div class="arch-column">
              <b>资源供给</b>
              <span v-for="item in supplyNodes" :key="item">{{ item }}</span>
            </div>
          </div>
        </div>
      </section>

      <section id="deploy" class="content-section section-alt">
        <div class="section-inner reveal-item">
          <div class="section-heading">
            <p class="section-label"><span>07</span> 部署与交付</p>
            <h2>统一平台直接使用，也支持企业资源独立部署</h2>
            <p>根据数据边界、模型资源和运维要求，选择托管服务、私有化部署或企业定制。</p>
          </div>
          <div class="deployment-grid">
            <article v-for="item in deployments" :key="item.title" class="reveal-item" :class="{ featured: item.featured }">
              <div><h3>{{ item.title }}</h3><span>{{ item.tag }}</span></div>
              <p>{{ item.description }}</p>
              <ul><li v-for="point in item.points" :key="point"><i />{{ point }}</li></ul>
              <router-link :to="entryPath">{{ item.action }}<Icon name="arrowRight" size="sm" /></router-link>
            </article>
          </div>
        </div>
      </section>

      <section class="final-cta reveal-item">
        <div class="final-cta-copy">
          <p class="section-label light">TOKENPORT</p>
          <h2>让模型资源可管理，让每一个 Token 有归属</h2>
          <p>从一个统一入口开始，建立企业 AI 的用量、成本、权限与能力资产体系。</p>
        </div>
        <router-link :to="entryPath" class="primary-link large final-cta-action">
          {{ isAuthenticated ? '进入控制台' : '登录体验' }}
          <Icon name="arrowRight" size="md" />
        </router-link>
      </section>
    </main>

    <footer class="site-footer">
      <div class="footer-inner">
        <div class="footer-brand">
          <div>
            <img :src="brandLogo" alt="" />
            <span><b>{{ siteName }}</b><small>智能应用与技能接入平台</small></span>
          </div>
          <p>统一接入、统一治理、统一核算、统一交付，让模型和智能体能力成为可管理的企业资源。</p>
          <div class="footer-tags">
            <span v-for="item in TOKENPORT_PRODUCT.protocols" :key="item">{{ item }} 兼容</span>
          </div>
        </div>
        <div class="footer-columns">
          <div v-for="column in footerColumns" :key="column.title" class="footer-column">
            <b>{{ column.title }}</b>
            <a v-for="link in column.links" :key="link.label" :href="link.href">{{ link.label }}</a>
          </div>
        </div>
      </div>
      <div class="footer-bottom">
        <p>© {{ currentYear }} {{ siteName }} · 智能应用与技能接入平台</p>
        <p>
          基于 <a :href="TOKENPORT_BRAND.upstreamUrl" target="_blank" rel="noopener">Sub2API</a>
          持续构建，遵循原项目许可证。
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useAuthStore, useAppStore } from '@/stores'
import { useThemeToggle } from '@/composables/useThemeToggle'
import Icon from '@/components/icons/Icon.vue'
import ConnectorShowcase from '@/tokenport/home/ConnectorShowcase.vue'
import ConsolePreview from '@/tokenport/home/ConsolePreview.vue'
import ModelFlowPreview from '@/tokenport/home/ModelFlowPreview.vue'
import TokenUsageShowcase from '@/tokenport/home/TokenUsageShowcase.vue'
import {
  fetchSkillMarket,
  getSkillCategoryName,
  getSkillDisplayDescription,
  getSkillDisplayName,
  getSkillRiskLabel,
} from '@/api/skillMarket'
import type { SkillMarketEntry } from '@/api/skillMarket'
import {
  TOKENPORT_BRAND,
  TOKENPORT_PRODUCT,
  resolveTokenPortLogo,
  resolveTokenPortName,
  resolveTokenPortSubtitle,
} from '@/tokenport/brand/tokenPortBrand'

const authStore = useAuthStore()
const appStore = useAppStore()
const props = withDefaults(defineProps<{ siteLogo?: string; docUrl?: string }>(), {
  siteLogo: '',
  docUrl: '',
})

const skillCount = ref(0)
const marketLoading = ref(true)
const marketError = ref('')
const marketCategories = ref<Array<{ id: string; name: string }>>([])
const activeCategory = ref('all')
const menuOpen = ref(false)
const homeRoot = ref<HTMLElement | null>(null)
const featuredSkills = ref<Array<{
  id: string
  name: string
  categoryId: string
  category: string
  description: string
  version: string
  risk: string
}>>([])
let revealObserver: IntersectionObserver | null = null
const { isDark, toggleTheme } = useThemeToggle()

const siteName = computed(() => resolveTokenPortName(appStore.cachedPublicSettings?.site_name || appStore.siteName))
const siteSubtitle = computed(() => resolveTokenPortSubtitle(appStore.cachedPublicSettings?.site_subtitle))
const brandLogo = computed(() => resolveTokenPortLogo(props.siteLogo))
const docUrl = computed(() => props.docUrl)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const entryPath = computed(() => isAuthenticated.value
  ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
  : '/login')
const currentYear = new Date().getFullYear()
const visibleSkills = computed(() => featuredSkills.value
  .filter((skill) => activeCategory.value === 'all' || skill.categoryId === activeCategory.value)
  .slice(0, 6))

function closeMenu() {
  menuOpen.value = false
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
  }, { threshold: 0.08, rootMargin: '0px 0px -48px' })

  homeRoot.value.querySelectorAll<HTMLElement>('.reveal-item:not(.is-visible)')
    .forEach((element) => revealObserver?.observe(element))
}

const capabilities = [
  {
    title: '统一模型网关',
    description: '统一管理上游资源、账号池、模型映射与路由，对外提供稳定的 OpenAI、Anthropic 与 Gemini 兼容入口。',
    points: ['多上游资源聚合', '模型映射与故障切换', '限流与可用性治理', '企业自有模型接入'],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="5" cy="6" r="2"/><circle cx="5" cy="18" r="2"/><circle cx="19" cy="12" r="2"/><path d="M7 6h4a4 4 0 0 1 4 4M7 18h4a4 4 0 0 0 4-4"/></svg>',
  },
  {
    title: '部门密钥与权限',
    description: '为部门、项目和用户签发独立 API Key，绑定模型范围、额度与分组策略，上游密钥不下发终端。',
    points: [],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect x="4" y="10" width="16" height="10" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/><circle cx="12" cy="15" r="1.4"/></svg>',
  },
  {
    title: 'Token 计量与成本',
    description: '记录输入、输出与总 Token，关联模型价格、部门、Key 和时间，形成可以核对的成本明细。',
    points: [],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M3 3v18h18M7 14l3-4 3 3 4-6"/><circle cx="20" cy="7" r="1"/></svg>',
  },
  {
    title: '智能应用连接器',
    description: '为 ChatGPT / Codex、Claude Code、OpenCode 和 Gemini CLI 生成可检查、可合并的接入配置。',
    points: [],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M4 6h10M4 12h16M4 18h7"/><circle cx="17" cy="6" r="2"/><circle cx="14" cy="18" r="2"/></svg>',
  },
  {
    title: 'Skill Market',
    description: '以中文说明、版本、依赖、风险、许可证和 SHA256 管理可复用能力，支持受控安装与私有化交付。',
    points: [],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><path d="M14 17.5h7M17.5 14v7"/></svg>',
  },
]

const entryNodes = ['业务应用', '研发工具', '智能体', '部门用户']
const coreNodes = ['统一密钥与权限', '模型路由与定价', 'Token 计量与预算', '接入配置与 Skill']
const supplyNodes = ['公有模型 API', '企业自有模型', '兼容渠道', '行业 Skill 包']

const deployments = [
  {
    title: '平台服务',
    tag: '快速启用',
    description: '企业和团队直接使用统一平台，快速获得模型接入、用量管理和 Skill Market。',
    points: ['托管网关与持续升级', '模型目录与价格维护', '按团队与用量运营'],
    action: '登录体验',
    featured: false,
  },
  {
    title: '私有化部署',
    tag: '密钥不出域',
    description: '部署到客户自有资源环境，统一纳管内部模型、公有模型接口与安全边界。',
    points: ['客户资源独立部署', '自有模型与专属渠道', '权限、日志与备份恢复'],
    action: '查看方案',
    featured: true,
  },
  {
    title: '企业定制',
    tag: '增值服务',
    description: '围绕组织架构、行业流程和交付标准，扩展专属路由、能力市场与经营报表。',
    points: ['行业 Skill 能力包', '品牌与系统集成', '运营分析与持续运维'],
    action: '联系方案',
    featured: false,
  },
]

const footerColumns = [
  {
    title: '核心能力',
    links: [
      { label: '统一模型网关', href: '#platform' },
      { label: 'Token 经营', href: '#cost' },
      { label: '智能应用连接器', href: '#connectors' },
      { label: 'Skill Market', href: '/skill-market' },
    ],
  },
  {
    title: '适用对象',
    links: [
      { label: '企业管理者', href: '#cost' },
      { label: '研发与产品团队', href: '#connectors' },
      { label: 'AI 运维团队', href: '#platform' },
      { label: '私有化客户', href: '#deploy' },
    ],
  },
  {
    title: '部署交付',
    links: [
      { label: '平台服务', href: '#deploy' },
      { label: '私有化部署', href: '#deploy' },
      { label: '企业定制', href: '#deploy' },
      { label: '模型与渠道', href: '/available-channels' },
    ],
  },
]

onMounted(async () => {
  await nextTick()
  observeRevealItems()
  await Promise.allSettled([authStore.checkAuth(), appStore.fetchPublicSettings()])
  try {
    const registry = await fetchSkillMarket()
    skillCount.value = registry.skills.length
    marketCategories.value = registry.categories
      .map((category) => ({
        id: category.id,
        name: category.name || getSkillCategoryName(category.id, registry),
      }))
      .filter((category) => Boolean(category.name))
    featuredSkills.value = registry.skills.map((skill: SkillMarketEntry) => {
      const description = getSkillDisplayDescription(skill).trim()
      return {
        id: skill.id,
        name: getSkillDisplayName(skill),
        categoryId: skill.category,
        category: getSkillCategoryName(skill.category, registry),
        description: description.length > 92 ? `${description.slice(0, 92)}…` : description,
        version: skill.version,
        risk: getSkillRiskLabel(skill.riskLevel),
      }
    })
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

onBeforeUnmount(() => revealObserver?.disconnect())
</script>

<style scoped>
.home-shell {
  --brand: var(--tp-brand);
  --brand-deep: var(--tp-brand-deep);
  --ink: #11231d;
  --muted: #60746b;
  --line: #d8e5df;
  --surface: #ffffff;
  --soft: #eef7f3;
  --soft-2: #f6faf8;
  min-height: 100vh;
  overflow-x: clip;
  background: #f8fbf9;
  color: var(--ink);
  font-family: "Noto Sans SC", "PingFang SC", "Microsoft YaHei", "Segoe UI", sans-serif;
  -webkit-font-smoothing: antialiased;
  text-rendering: optimizeLegibility;
}

.dark .home-shell {
  --ink: #edf6f2;
  --muted: #91a99f;
  --line: #20362c;
  --surface: #0e1914;
  --soft: #12231b;
  --soft-2: #0a120f;
  background: #08100d;
}

.topbar {
  position: sticky;
  z-index: 30;
  top: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  min-height: 68px;
  padding: 10px max(24px, calc((100vw - 1180px) / 2));
  border-bottom: 1px solid color-mix(in srgb, var(--line) 78%, transparent);
  background: color-mix(in srgb, var(--surface) 84%, transparent);
  backdrop-filter: blur(18px);
}

.brand-link,
.brand-link span,
.topbar nav,
.primary-link,
.text-link,
.footer-brand > div,
.footer-brand > div span {
  display: flex;
  align-items: center;
}

.brand-link {
  gap: 11px;
  color: var(--ink);
  text-decoration: none;
}

.brand-link img,
.site-footer img {
  width: 38px;
  height: 38px;
  border-radius: var(--tp-radius-control);
  object-fit: cover;
  box-shadow: var(--tp-elev-1);
}

.brand-link span,
.footer-brand > div span {
  flex-direction: column;
  align-items: flex-start;
}

.brand-link b { font-size: 17px; font-weight: 800; }
.brand-link small { margin-top: 2px; color: var(--muted); font-size: 11px; }

.topbar nav { gap: 6px; }

.menu-control {
  display: none;
  width: 40px;
  height: 40px;
  place-items: center;
  border: 1px solid var(--line);
  border-radius: var(--tp-radius-control);
  background: var(--surface);
  color: var(--ink);
  cursor: pointer;
}

.nav-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-height: 38px;
  padding: 0 10px;
  color: var(--muted);
  font-size: 13px;
  font-weight: 600;
  text-decoration: none;
  transition: color 0.16s ease;
}

.nav-link:hover { color: var(--ink); }
.nav-link.is-gated .gate-icon { opacity: 0.58; }

.icon-control {
  display: grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border: 1px solid var(--line);
  border-radius: var(--tp-radius-control);
  background: var(--surface);
  color: var(--ink);
  cursor: pointer;
}

.primary-link,
.secondary-link {
  justify-content: center;
  gap: 8px;
  min-height: 46px;
  padding: 0 18px;
  border-radius: var(--tp-radius-card);
  font-size: 14px;
  font-weight: 700;
  white-space: nowrap;
  text-decoration: none;
  transition: transform 0.16s ease, box-shadow 0.16s ease;
}

.primary-link {
  background: var(--brand);
  color: #fff;
  box-shadow: 0 10px 24px color-mix(in srgb, var(--brand) 25%, transparent);
}

.primary-link.compact { min-height: 40px; padding: 0 14px; }
.primary-link.large { min-height: 50px; padding: 0 22px; }
.primary-link:hover,
.secondary-link:hover { transform: translateY(-1px); }

.secondary-link {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--line);
  background: color-mix(in srgb, var(--surface) 88%, transparent);
  color: var(--ink);
}

.hero-band {
  position: relative;
  overflow: hidden;
  border-bottom: 1px solid var(--line);
  background: var(--soft-2);
}

.hero-grid {
  position: absolute;
  inset: 0;
  opacity: 0.72;
  background-image:
    linear-gradient(color-mix(in srgb, var(--line) 54%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--line) 54%, transparent) 1px, transparent 1px);
  background-size: 56px 56px;
  mask-image: linear-gradient(to bottom, #000 0%, #000 82%, transparent 100%);
}

.hero-inner {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(500px, 0.92fr);
  align-items: center;
  gap: 64px;
  width: min(1180px, calc(100% - 48px));
  min-height: min(760px, calc(100vh - 68px));
  margin: 0 auto;
  padding: 70px 0 58px;
}

.status-label,
.section-label {
  color: var(--brand-deep);
  font: 700 12px/1.4 ui-monospace, Consolas, monospace;
  letter-spacing: 0;
}

.status-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 11px;
  border: 1px solid color-mix(in srgb, var(--brand) 30%, var(--line));
  border-radius: 999px;
  background: color-mix(in srgb, var(--brand) 7%, var(--surface));
}

.status-label i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--brand);
  box-shadow: 0 0 12px color-mix(in srgb, var(--brand) 62%, transparent);
}

.hero-copy h1 {
  margin: 24px 0 0;
  font-size: 52px;
  line-height: 1.16;
  font-weight: 900;
  letter-spacing: 0;
  word-break: keep-all;
}

.hero-copy h1 span { display: block; }
.hero-copy h1 em { color: var(--brand); font-style: normal; }

.lead {
  max-width: 620px;
  margin: 22px 0 0;
  color: var(--muted);
  font-size: 17px;
  line-height: 1.85;
}

.hero-actions { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 30px; }

.hero-stats {
  display: grid;
  grid-template-columns: 0.7fr 0.8fr 1.5fr;
  gap: 1px;
  max-width: 580px;
  margin: 42px 0 0;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: var(--tp-radius-card);
  background: var(--line);
}

.hero-stats div {
  padding: 15px 18px;
  background: color-mix(in srgb, var(--surface) 90%, var(--soft-2));
}

.hero-stats dt {
  color: var(--ink);
  font: 800 22px/1.2 ui-monospace, Consolas, monospace;
  white-space: nowrap;
}

.hero-stats .wide dt { font-size: 14px; line-height: 1.5; white-space: normal; }
.hero-stats dd { margin: 6px 0 0; color: var(--muted); font-size: 11px; }

.tool-strip {
  padding: 26px 24px;
  border-bottom: 1px solid var(--line);
  background: color-mix(in srgb, var(--surface) 70%, var(--soft-2));
}

.tool-strip-inner {
  display: flex;
  align-items: center;
  gap: 20px 40px;
  width: min(1180px, 100%);
  margin: 0 auto;
}

.tool-strip-label {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  flex: 0 0 auto;
  color: var(--muted);
  font: 700 11px/1.4 ui-monospace, Consolas, monospace;
}

.tool-strip-label i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--brand);
  box-shadow: 0 0 10px color-mix(in srgb, var(--brand) 60%, transparent);
}

.tool-list {
  display: flex;
  flex-wrap: wrap;
  gap: 9px;
}

.tool-list span {
  padding: 7px 13px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--surface);
  color: var(--muted);
  font: 600 12.5px/1 ui-monospace, Consolas, monospace;
  transition: border-color 0.16s ease, color 0.16s ease, transform 0.16s ease;
}

.tool-list span:hover {
  transform: translateY(-1px);
  border-color: color-mix(in srgb, var(--brand) 42%, var(--line));
  color: var(--ink);
}

.content-section { padding: 96px 24px; border-bottom: 1px solid var(--line); }
.section-alt { background: var(--soft-2); }
.section-inner { width: min(1180px, 100%); margin: 0 auto; }
.section-heading { max-width: 740px; }
.section-heading.horizontal { display: flex; align-items: end; justify-content: space-between; gap: 40px; max-width: none; }
.section-heading.horizontal > div { max-width: 760px; }
.section-label span { margin-right: 8px; color: var(--muted); }

.section-heading h2,
.final-cta h2 {
  margin: 16px 0 0;
  font-size: 38px;
  line-height: 1.28;
  font-weight: 850;
  letter-spacing: 0;
  word-break: keep-all;
}

.section-heading > p:last-child,
.section-heading.horizontal > p,
.market-intro {
  margin: 16px 0 0;
  color: var(--muted);
  font-size: 15px;
  line-height: 1.8;
}

.section-heading.horizontal > p { max-width: 390px; }

.capability-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 15px;
  margin-top: 42px;
}

.capability-grid article {
  display: flex;
  flex-direction: column;
  min-height: 250px;
  padding: 24px;
  border: 1px solid var(--line);
  border-radius: var(--tp-radius-panel);
  background: var(--surface);
  box-shadow: var(--tp-elev-1);
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.capability-grid article:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, var(--brand) 42%, var(--line));
  box-shadow: var(--tp-elev-2);
}

.capability-grid article.featured {
  display: grid;
  grid-template-columns: 50px 1fr 0.9fr;
  grid-column: span 2;
  gap: 18px;
  align-items: start;
  border-color: color-mix(in srgb, var(--brand) 26%, var(--line));
  background: color-mix(in srgb, var(--brand) 6%, var(--surface));
}

.capability-grid .featured .capability-icon {
  border-color: color-mix(in srgb, var(--brand) 46%, var(--line));
  background: color-mix(in srgb, var(--brand) 15%, var(--soft));
}

.capability-grid .featured ul { align-self: center; }

.capability-icon {
  display: grid;
  width: 46px;
  height: 46px;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--brand) 32%, var(--line));
  border-radius: var(--tp-radius-control);
  background: color-mix(in srgb, var(--brand) 8%, var(--soft));
  color: var(--brand-deep);
}

.capability-icon :deep(svg) { width: 21px; height: 21px; stroke-width: 1.7; stroke-linecap: round; stroke-linejoin: round; }
.capability-copy h3 { margin: 18px 0 0; font-size: 19px; }
.capability-grid .featured .capability-copy h3 { margin-top: 0; }
.capability-copy p { margin: 10px 0 0; color: var(--muted); font-size: 14px; line-height: 1.75; }
.capability-grid ul { display: grid; gap: 10px; margin: 4px 0 0; padding: 0; list-style: none; }
.capability-grid li { display: flex; align-items: center; gap: 9px; color: var(--muted); font-size: 13px; }
.capability-grid li i,
.deployment-grid li i { width: 6px; height: 6px; flex: 0 0 6px; border-radius: 50%; background: var(--brand); }

.split-section { display: grid; grid-template-columns: minmax(0, 0.82fr) minmax(0, 1.18fr); align-items: center; gap: 72px; }
.split-section > * { min-width: 0; }
.split-section .section-heading { max-width: none; }
.split-section .section-heading h2 { font-size: 34px; word-break: normal; overflow-wrap: anywhere; }
.value-points { display: grid; gap: 1px; margin-top: 30px; overflow: hidden; border: 1px solid var(--line); border-radius: var(--tp-radius-panel); background: var(--line); }
.value-points div { display: grid; grid-template-columns: 104px 1fr; gap: 16px; align-items: center; padding: 16px 18px; background: var(--surface); transition: background 0.16s ease; }
.value-points div:hover { background: color-mix(in srgb, var(--brand) 5%, var(--surface)); }
.value-points b { position: relative; padding-left: 14px; font-size: 14px; }
.value-points b::before { content: ""; position: absolute; left: 0; top: 50%; width: 5px; height: 5px; border-radius: 50%; background: var(--brand); transform: translateY(-50%); }
.value-points span { color: var(--muted); font-size: 13px; line-height: 1.6; }

.console-section { overflow: hidden; }
.console-stage { position: relative; margin-top: 42px; }
.console-stage::before {
  content: "";
  position: absolute;
  inset: 24px -24px -24px;
  border: 1px solid color-mix(in srgb, var(--brand) 16%, transparent);
  border-radius: var(--tp-radius-panel);
  background-image: radial-gradient(circle at center, color-mix(in srgb, var(--brand) 15%, transparent) 1px, transparent 1.4px);
  background-size: 22px 22px;
  opacity: 0.5;
}
.console-stage :deep(.console-preview) { position: relative; z-index: 1; }
.connector-heading { margin-bottom: 38px; }

.text-link {
  flex: 0 0 auto;
  gap: 7px;
  min-height: 42px;
  padding: 0 14px;
  border: 1px solid var(--line);
  border-radius: var(--tp-radius-control);
  color: var(--brand-deep);
  font-size: 13px;
  font-weight: 700;
  text-decoration: none;
}

.market-intro { max-width: 860px; }
.market-categories { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 24px; }
.market-categories button {
  min-height: 32px;
  padding: 0 12px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--surface);
  color: var(--muted);
  font: inherit;
  font-size: 12px;
  cursor: pointer;
  transition: border-color 0.16s ease, background 0.16s ease, color 0.16s ease;
}
.market-categories button:hover { border-color: color-mix(in srgb, var(--brand) 36%, var(--line)); color: var(--ink); }
.market-categories button.active { border-color: color-mix(in srgb, var(--brand) 48%, var(--line)); background: color-mix(in srgb, var(--brand) 12%, var(--surface)); color: var(--brand-deep); font-weight: 700; }
.market-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 15px; margin-top: 24px; }
.market-grid article {
  display: flex;
  flex-direction: column;
  min-height: 230px;
  padding: 22px;
  border: 1px solid var(--line);
  border-radius: var(--tp-radius-panel);
  background: var(--surface);
  box-shadow: var(--tp-elev-1);
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}
.market-grid article:hover { transform: translateY(-3px); border-color: color-mix(in srgb, var(--brand) 42%, var(--line)); box-shadow: var(--tp-elev-2); }
.skill-meta { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.skill-meta span,
.skill-meta em { padding: 5px 8px; border: 1px solid var(--line); border-radius: 999px; color: var(--muted); font: 700 10px/1 ui-monospace, Consolas, monospace; font-style: normal; }
.skill-meta em { color: var(--brand-deep); }
.market-grid h3 { margin: 20px 0 0; font-size: 17px; line-height: 1.45; }
.market-grid p { margin: 10px 0 0; color: var(--muted); font-size: 13px; line-height: 1.7; }
.market-grid footer { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: auto; padding-top: 16px; border-top: 1px solid var(--line); }
.market-grid footer span { overflow: hidden; color: var(--muted); font: 500 11px/1.4 ui-monospace, Consolas, monospace; letter-spacing: -0.01em; text-overflow: ellipsis; white-space: nowrap; }
.market-grid footer b { flex: 0 0 auto; padding: 3px 8px; border-radius: 999px; background: color-mix(in srgb, var(--brand) 12%, transparent); color: var(--brand-deep); font: 700 11px/1 ui-monospace, Consolas, monospace; }
.loading-grid article { min-height: 230px; background: linear-gradient(90deg, var(--surface), var(--soft), var(--surface)); background-size: 200% 100%; animation: tp-loading 1.5s ease infinite; }
.market-empty { display: grid; gap: 7px; margin-top: 24px; padding: 24px; border: 1px solid var(--line); border-radius: var(--tp-radius-panel); background: var(--surface); }
.market-empty span { color: var(--muted); font-size: 13px; }

.architecture { display: grid; grid-template-columns: 1fr auto 1.3fr auto 1fr; align-items: stretch; gap: 16px; margin-top: 42px; }
.arch-column,
.arch-core { display: grid; gap: 10px; padding: 22px; border: 1px solid var(--line); border-radius: var(--tp-radius-panel); background: var(--surface); box-shadow: var(--tp-elev-1); }
.arch-column > b,
.arch-core > b { margin-bottom: 3px; font-size: 18px; }
.arch-column span,
.arch-core span { padding: 12px 13px; border: 1px solid var(--line); border-radius: var(--tp-radius-card); background: var(--soft-2); color: var(--muted); font-size: 13px; }
.arch-core { border-color: color-mix(in srgb, var(--brand) 44%, var(--line)); background: color-mix(in srgb, var(--brand) 7%, var(--surface)); box-shadow: var(--tp-elev-2); }
.arch-core em { width: fit-content; padding: 5px 8px; border-radius: 999px; background: color-mix(in srgb, var(--brand) 13%, transparent); color: var(--brand-deep); font: 700 10px/1 ui-monospace, Consolas, monospace; font-style: normal; }
.arch-core > b { color: var(--brand-deep); font-size: 24px; }
.flow-arrow { display: grid; place-items: center; }
.flow-arrow span { position: relative; width: 42px; height: 2px; background: linear-gradient(90deg, transparent, var(--brand), transparent); }
.flow-arrow span::after { content: ""; position: absolute; right: -1px; top: 50%; width: 8px; height: 8px; border-top: 2px solid var(--brand); border-right: 2px solid var(--brand); transform: translateY(-50%) rotate(45deg); }

.deployment-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 15px; margin-top: 42px; }
.deployment-grid article { display: flex; flex-direction: column; min-height: 360px; padding: 24px; border: 1px solid var(--line); border-radius: var(--tp-radius-panel); background: var(--surface); box-shadow: var(--tp-elev-1); }
.deployment-grid article.featured { border-color: color-mix(in srgb, var(--brand) 48%, var(--line)); box-shadow: var(--tp-elev-2); }
.deployment-grid article > div { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.deployment-grid h3 { font-size: 19px; }
.deployment-grid article > div span { padding: 6px 9px; border-radius: 999px; background: color-mix(in srgb, var(--brand) 12%, transparent); color: var(--brand-deep); font: 700 10px/1 ui-monospace, Consolas, monospace; }
.deployment-grid p { margin: 16px 0 0; color: var(--muted); font-size: 14px; line-height: 1.75; }
.deployment-grid ul { display: grid; gap: 12px; margin: 24px 0; padding: 0; list-style: none; }
.deployment-grid li { display: flex; align-items: center; gap: 9px; color: var(--muted); font-size: 13px; }
.deployment-grid a { display: inline-flex; align-items: center; justify-content: center; gap: 7px; min-height: 42px; margin-top: auto; border: 1px solid var(--line); border-radius: var(--tp-radius-control); color: var(--ink); font-size: 13px; font-weight: 700; text-decoration: none; }
.deployment-grid .featured a { border-color: var(--brand); background: var(--brand); color: #fff; }

.final-cta {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 48px;
  width: min(1180px, 100% - 48px);
  margin: 96px auto;
  padding: 64px 60px;
  border: 1px solid color-mix(in srgb, var(--brand) 38%, transparent);
  border-radius: var(--tp-radius-panel);
  background:
    radial-gradient(130% 150% at 100% 0%, color-mix(in srgb, var(--brand) 34%, transparent) 0%, transparent 56%),
    linear-gradient(150deg, #0f2a20 0%, #0a1c15 100%);
  color: #fff;
  box-shadow: var(--tp-elev-3);
}
/* dot-grid texture, faded toward the copy so text stays crisp */
.final-cta::before {
  content: "";
  position: absolute;
  inset: 0;
  z-index: -1;
  background-image: radial-gradient(color-mix(in srgb, var(--brand) 46%, transparent) 1px, transparent 1.6px);
  background-size: 22px 22px;
  opacity: 0.18;
  mask-image: radial-gradient(120% 130% at 100% 0%, #000 0%, transparent 72%);
}
.final-cta-copy { position: relative; max-width: 620px; }
.final-cta .section-label.light { color: color-mix(in srgb, var(--brand) 52%, #ffffff); letter-spacing: 0.14em; }
.final-cta h2 { max-width: 620px; font-size: 36px; color: #fff; }
.final-cta p:last-child { margin: 16px 0 0; max-width: 560px; color: rgb(255 255 255 / 74%); font-size: 15px; line-height: 1.72; }
/* high-contrast solid button so the CTA reads clearly on the deep band in both themes */
.final-cta-action { flex-shrink: 0; background: #ffffff; color: #0a1c15; box-shadow: 0 12px 28px rgb(6 20 15 / 34%); }
.final-cta-action:hover { transform: translateY(-1px); box-shadow: 0 16px 34px rgb(6 20 15 / 42%); }

.site-footer { padding: 76px 24px 30px; border-top: 1px solid var(--line); background: var(--soft-2); color: var(--muted); }
.footer-inner { display: grid; grid-template-columns: minmax(0, 1.5fr) minmax(0, 2fr); gap: 72px; width: min(1180px, 100%); margin: 0 auto; }
.footer-brand > div { gap: 12px; }
.footer-brand b { color: var(--ink); font-size: 16px; font-weight: 800; }
.footer-brand small { margin-top: 3px; color: var(--muted); font-size: 12px; }
.footer-brand > p { max-width: 380px; margin: 20px 0 0; color: var(--muted); font-size: 13.5px; line-height: 1.85; }
.footer-tags { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 24px; }
.footer-tags span { padding: 6px 12px; border: 1px solid var(--line); border-radius: 999px; background: var(--surface); color: var(--muted); font: 600 11px/1 ui-monospace, Consolas, monospace; }
.footer-columns { display: grid; grid-template-columns: repeat(3, 1fr); gap: 44px; }
.footer-column { display: flex; flex-direction: column; align-items: flex-start; gap: 14px; }
.footer-column > b { margin-bottom: 4px; color: var(--brand-deep); font: 700 11px/1.4 ui-monospace, Consolas, monospace; letter-spacing: 0.09em; text-transform: uppercase; }
.footer-column > a { color: var(--muted); font-size: 13px; text-decoration: none; transition: color 0.16s ease, transform 0.16s ease; }
.footer-column > a:hover { color: var(--brand-deep); transform: translateX(3px); }
.footer-bottom { display: flex; align-items: center; justify-content: space-between; gap: 20px; width: min(1180px, 100%); margin: 52px auto 0; padding-top: 24px; border-top: 1px solid var(--line); }
.footer-bottom p { color: var(--muted); font-size: 12.5px; }
.site-footer p { margin: 0; }
.site-footer a { color: var(--brand-deep); text-decoration: none; }

@keyframes tp-loading { to { background-position: -200% 0; } }

.reveal-item {
  opacity: 0;
  transform: translateY(22px);
  transition: opacity 0.66s cubic-bezier(0.22, 1, 0.36, 1), transform 0.66s cubic-bezier(0.22, 1, 0.36, 1);
}

.reveal-item.is-visible {
  opacity: 1;
  transform: none;
}

@media (max-width: 1120px) {
  .topbar .nav-link-optional { display: none; }
  .hero-inner { grid-template-columns: 1fr; gap: 44px; min-height: auto; }
  .hero-copy { max-width: 760px; }
  .hero-inner :deep(.flow-preview) { max-width: 760px; margin: 0 auto; }
  .split-section { grid-template-columns: minmax(0, 1fr); }
  .architecture { grid-template-columns: 1fr; }
  .flow-arrow span { width: 2px; height: 36px; background: linear-gradient(180deg, transparent, var(--brand), transparent); }
  .flow-arrow span::after { right: 50%; top: auto; bottom: -1px; transform: translateX(50%) rotate(135deg); }
}

@media (max-width: 820px) {
  .capability-grid,
  .market-grid,
  .deployment-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .capability-grid article.featured { grid-template-columns: 46px 1fr; grid-column: 1 / -1; }
  .capability-grid article.featured ul { grid-column: 2; }
  .section-heading.horizontal { align-items: flex-start; flex-direction: column; gap: 20px; }
  .tool-strip-inner { flex-direction: column; align-items: flex-start; gap: 14px; }
  .footer-inner { grid-template-columns: 1fr; gap: 44px; }
}

@media (max-width: 720px) {
  .topbar nav > .nav-link-optional { display: none; }
  .topbar { padding-inline: 16px; }
  .brand-link small { display: none; }
  .menu-control { display: grid; margin-left: auto; }
  .topbar nav {
    position: absolute;
    top: calc(100% + 8px);
    right: 16px;
    left: 16px;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 3px;
    max-height: 0;
    padding: 0 12px;
    overflow: hidden;
    border: 1px solid transparent;
    border-radius: var(--tp-radius-panel);
    background: var(--surface);
    box-shadow: var(--tp-elev-3);
    opacity: 0;
    pointer-events: none;
    transition: max-height 0.26s ease, padding 0.26s ease, opacity 0.2s ease;
  }
  .topbar nav.open { max-height: 520px; padding: 12px; border-color: var(--line); opacity: 1; pointer-events: auto; }
  .topbar nav.open > .nav-link-optional { display: flex; }
  .nav-link { width: 100%; justify-content: flex-start; min-height: 42px; padding-inline: 12px; font-size: 13px; }
  .icon-control,
  .primary-link.compact { width: 100%; min-height: 42px; }
  .hero-inner { width: calc(100% - 32px); padding: 54px 0 44px; }
  .hero-copy h1 { font-size: 38px; }
  .lead { font-size: 15px; }
  .hero-stats { grid-template-columns: 0.8fr 0.9fr 1.4fr; }
  .content-section { padding: 72px 16px; }
  .section-heading h2,
  .final-cta h2 { font-size: 30px; }
  .final-cta { align-items: flex-start; flex-direction: column; gap: 28px; width: calc(100% - 32px); margin: 64px auto; padding: 44px 28px; }
  .final-cta-action { width: 100%; }
  .footer-bottom { align-items: flex-start; flex-direction: column; }
}

@media (max-width: 560px) {
  .hero-copy h1 { font-size: 34px; }
  .hero-stats { grid-template-columns: repeat(2, 1fr); }
  .hero-stats .wide { grid-column: 1 / -1; }
  .tool-list { gap: 7px; }
  .capability-grid,
  .market-grid,
  .deployment-grid { grid-template-columns: 1fr; }
  .capability-grid article.featured { display: flex; }
  .value-points div { grid-template-columns: 1fr; gap: 5px; }
  .market-grid article { min-height: 210px; }
  .footer-columns { grid-template-columns: repeat(2, 1fr); gap: 28px 24px; }
}

@media (prefers-reduced-motion: reduce) {
  .reveal-item { opacity: 1; transform: none; transition: none; }
}
</style>
