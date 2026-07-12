<template>
  <div class="home-shell">
    <div class="bg-glow" aria-hidden="true" />

    <header class="topbar">
      <router-link to="/home" class="brand-link">
        <img :src="brandLogo" alt="天翼云 TokenPort" />
        <span>
          <b>{{ siteName }}</b>
          <small>{{ siteSubtitle }}</small>
        </span>
      </router-link>
      <nav>
        <router-link to="/skill-market">Skill Market</router-link>
        <router-link to="/available-channels">模型与渠道</router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">Docs</a>
        <button
          type="button"
          class="icon-control"
          :title="isDark ? '切换浅色模式' : '切换深色模式'"
          @click="toggleTheme"
        >
          <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
        </button>
        <router-link :to="entryPath" class="primary-link">
          {{ isAuthenticated ? '进入控制台' : '登录平台' }}
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </nav>
    </header>

    <main>
      <section class="hero-band">
        <div class="hero-copy">
          <p class="eyebrow">AI ACCESS · TOKEN OPERATIONS · SKILL DELIVERY</p>
          <h1>统一管理模型、Token 与智能体能力</h1>
          <p class="lead">
            模型越来越多、企业用量越来越大、AI 工具种类繁杂。TokenPort 用一个受控入口连接模型资源、开发工具和可复用技能，让每次调用有归属、有成本、有治理。
          </p>
          <div class="hero-actions">
            <router-link :to="entryPath" class="primary-link large">
              {{ isAuthenticated ? '进入经营控制台' : '开始统一接入' }}
              <Icon name="arrowRight" size="md" />
            </router-link>
            <router-link to="/skill-market" class="secondary-link">浏览能力市场</router-link>
          </div>
          <div class="chip-row" aria-label="支持协议与工具">
            <span v-for="item in TOKENPORT_PRODUCT.protocols" :key="item" class="chip">{{ item }}</span>
            <span class="chip-divider" />
            <span v-for="item in TOKENPORT_PRODUCT.clients" :key="item" class="chip soft">{{ item }}</span>
          </div>
          <dl class="signal-row">
            <div>
              <dt>协议入口</dt>
              <dd>{{ TOKENPORT_PRODUCT.protocols.length }} 类</dd>
            </div>
            <div>
              <dt>工具接入</dt>
              <dd>{{ TOKENPORT_PRODUCT.clients.length }} 类</dd>
            </div>
            <div>
              <dt>能力市场</dt>
              <dd>{{ skillCount }} 个 Skill</dd>
            </div>
          </dl>
        </div>
        <div class="product-stage">
          <ConsolePreview :logo-src="brandLogo" />
          <p class="preview-hint">可点击左侧菜单切换页面，查看控制台能力结构</p>
        </div>
      </section>

      <section class="change-band">
        <div class="section-heading">
          <p class="section-label">市场变化</p>
          <h2>企业不缺模型入口，缺的是统一经营能力</h2>
        </div>
        <div class="change-grid">
          <article v-for="item in changes" :key="item.title">
            <div class="card-index">{{ item.index }}</div>
            <b>{{ item.title }}</b>
            <p>{{ item.description }}</p>
          </article>
        </div>
      </section>

      <section class="capability-band">
        <div class="section-heading wide">
          <div>
            <p class="section-label">产品能力</p>
            <h2>从统一接入到部门经营，形成完整闭环</h2>
          </div>
          <span>模型与价格来自当前可用渠道，首页不维护容易失真的静态价格表。</span>
        </div>
        <div class="capability-grid">
          <article v-for="item in capabilities" :key="item.title">
            <div class="cap-head">
              <span>{{ item.number }}</span>
              <h3>{{ item.title }}</h3>
            </div>
            <p>{{ item.description }}</p>
            <ul>
              <li v-for="point in item.points" :key="point">{{ point }}</li>
            </ul>
          </article>
        </div>
      </section>

      <section class="architecture-band">
        <div class="section-heading">
          <p class="section-label">系统架构</p>
          <h2>一个平台连接企业用户、AI 工具与模型资源</h2>
        </div>
        <div class="architecture">
          <div class="arch-column">
            <b>使用入口</b>
            <span v-for="item in entryNodes" :key="item">{{ item }}</span>
          </div>
          <div class="flow-arrow" aria-hidden="true">
            <span />
          </div>
          <div class="arch-core">
            <div class="core-badge">CONTROL PLANE</div>
            <b>TokenPort</b>
            <span v-for="item in coreNodes" :key="item">{{ item }}</span>
          </div>
          <div class="flow-arrow" aria-hidden="true">
            <span />
          </div>
          <div class="arch-column">
            <b>资源供给</b>
            <span v-for="item in supplyNodes" :key="item">{{ item }}</span>
          </div>
        </div>
      </section>

      <section class="market-band">
        <div class="section-heading wide">
          <div>
            <p class="section-label">能力市场</p>
            <h2>可复用 Skill，直接进入开发与交付流程</h2>
          </div>
          <router-link to="/skill-market" class="text-link">
            查看全部 {{ skillCount }} 个 Skill
            <Icon name="arrowRight" size="sm" />
          </router-link>
        </div>
        <div class="market-meta">
          <span v-for="cat in categoryNames" :key="cat">{{ cat }}</span>
        </div>
        <div class="market-grid">
          <article v-for="skill in featuredSkills" :key="skill.id">
            <div class="skill-top">
              <div class="skill-identity">
                <span class="skill-avatar">{{ skill.name.slice(0, 1) }}</span>
                <div>
                  <b>{{ skill.name }}</b>
                  <em>{{ skill.category }}</em>
                </div>
              </div>
            </div>
            <p>{{ skill.description }}</p>
          </article>
        </div>
      </section>

      <section class="value-band">
        <div class="value-copy">
          <p class="section-label">经营价值</p>
          <h2>把零散的 AI 消耗变成可核算、可优化、可交付的能力</h2>
          <p>
            统一入口减少重复配置，模型路由降低不必要的高价调用，部门报表明确成本归属，Skill 复用减少重复开发。平台既服务公司内部降本增效，也支持客户私有化部署和行业能力包交付。
          </p>
        </div>
        <div class="value-list">
          <div v-for="item in values" :key="item.title">
            <b>{{ item.title }}</b>
            <span>{{ item.description }}</span>
          </div>
        </div>
      </section>

      <section class="deployment-band">
        <div class="section-heading">
          <p class="section-label">部署方式</p>
          <h2>统一平台直接使用，也支持企业资源独立部署</h2>
        </div>
        <div class="deployment-options">
          <article v-for="item in deployments" :key="item.title">
            <div class="deploy-tag">{{ item.tag }}</div>
            <b>{{ item.title }}</b>
            <p>{{ item.description }}</p>
          </article>
        </div>
      </section>

      <section class="final-cta">
        <div>
          <p class="section-label light">TOKENPORT</p>
          <h2>让 Token 成为可经营资产，让 Skill 成为可交付资产</h2>
        </div>
        <router-link :to="entryPath" class="primary-link large">
          {{ isAuthenticated ? '进入控制台' : '登录体验' }}
          <Icon name="arrowRight" size="md" />
        </router-link>
      </section>
    </main>

    <footer>
      <span>© {{ currentYear }} {{ siteName }}</span>
      <span>
        基于
        <a :href="TOKENPORT_BRAND.upstreamUrl" target="_blank" rel="noopener">Sub2API</a>
        持续构建，遵循原项目许可证。
      </span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAuthStore, useAppStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import ConsolePreview from '@/tokenport/home/ConsolePreview.vue'
import {
  fetchSkillMarket,
  getSkillCategoryName,
  getSkillDisplayDescription,
  getSkillDisplayName,
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
const props = withDefaults(
  defineProps<{
    siteLogo?: string
    docUrl?: string
  }>(),
  {
    siteLogo: '',
    docUrl: '',
  },
)

const skillCount = ref(0)
const categoryCount = ref(0)
const categoryNames = ref<string[]>([])
const featuredSkills = ref<Array<{ id: string; name: string; category: string; description: string }>>([])
const isDark = ref(document.documentElement.classList.contains('dark'))

const siteName = computed(() =>
  resolveTokenPortName(appStore.cachedPublicSettings?.site_name || appStore.siteName),
)
const siteSubtitle = computed(() =>
  resolveTokenPortSubtitle(appStore.cachedPublicSettings?.site_subtitle),
)
const siteLogo = computed(() => props.siteLogo)
const brandLogo = computed(() => resolveTokenPortLogo(siteLogo.value))
const docUrl = computed(() => props.docUrl)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const entryPath = computed(() =>
  isAuthenticated.value ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard') : '/login',
)
const currentYear = new Date().getFullYear()

const changes = [
  {
    index: '01',
    title: '模型供给增加',
    description: '国内外模型与兼容渠道持续扩展，单一供应商已无法覆盖所有业务场景。',
  },
  {
    index: '02',
    title: '调用规模扩大',
    description: '研发、产品、运营与客服同时使用 AI，密钥、成本和权限边界快速复杂化。',
  },
  {
    index: '03',
    title: '工具入口分散',
    description: 'Codex、Claude Code、OpenCode、Gemini CLI 等工具需要不同配置和能力目录。',
  },
]

const capabilities = [
  {
    number: '01',
    title: '统一模型接入',
    description: '连接 OpenAI、Anthropic、Gemini 兼容资源与企业自有模型，按可用性和成本组织服务。',
    points: ['多协议兼容', '自有模型接入', '渠道可用性'],
  },
  {
    number: '02',
    title: 'Token 经营核算',
    description: '按部门、用户、API Key 和模型归集调用量、Token、成本、额度与预算预警。',
    points: ['部门成本归属', '预算预警', '用量审计'],
  },
  {
    number: '03',
    title: '智能工具配置',
    description: '为 Codex、Claude Code、OpenCode、Gemini CLI 与 CCS 生成可检查、可合并的配置。',
    points: ['一键配置', '诊断合并', '客户端兼容'],
  },
  {
    number: '04',
    title: 'Skill 能力市场',
    description: '以版本、风险、依赖、许可证和 SHA256 管理可复用智能体能力。',
    points: ['版本管理', '风险标注', '可校验交付'],
  },
]

const entryNodes = ['业务应用', '研发工具', '智能体', '部门用户']
const coreNodes = ['统一密钥与权限', '模型路由与定价', 'Token 计量与预算', '接入配置与 Skill']
const supplyNodes = ['公有模型 API', '企业自有模型', '代理与兼容渠道', '行业 Skill 包']

const values = computed(() => [
  {
    title: '成本透明',
    description: '按部门、Key、模型和时间追踪 Token 与实际成本',
  },
  {
    title: '效率提升',
    description: '一处生成 Codex、Claude、OpenCode 与 CCS 接入配置',
  },
  {
    title: '资产复用',
    description: `${skillCount.value} 个 Skill、${categoryCount.value} 类能力可查看版本、风险与校验信息`,
  },
  {
    title: '客户交付',
    description: '支持品牌配置、自有模型资源接入和私有化部署',
  },
])

const deployments = [
  {
    tag: 'SaaS',
    title: '平台服务',
    description: '企业和团队直接使用统一平台，快速获得模型接入、用量治理与能力市场。',
  },
  {
    tag: 'On-Prem',
    title: '私有化交付',
    description: '客户已有算力、大模型或安全边界时，可在客户资源环境内完成部署与定制。',
  },
  {
    tag: 'Addon',
    title: '增值服务',
    description: '提供行业 Skill 包、模型资源接入、运营报表、运维诊断与持续升级服务。',
  },
]

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(async () => {
  await Promise.allSettled([authStore.checkAuth(), appStore.fetchPublicSettings()])
  try {
    const registry = await fetchSkillMarket()
    skillCount.value = registry.skills.length
    categoryCount.value = registry.categories.length
    categoryNames.value = registry.categories
      .map((cat) => cat.name || getSkillCategoryName(cat.id, registry))
      .filter(Boolean)
      .slice(0, 6)

    featuredSkills.value = registry.skills.slice(0, 4).map((skill: SkillMarketEntry) => {
      const description = getSkillDisplayDescription(skill).trim()
      return {
        id: skill.id,
        name: getSkillDisplayName(skill),
        category: getSkillCategoryName(skill.category, registry),
        description: description.length > 72 ? `${description.slice(0, 72)}…` : description,
      }
    })
  } catch {
    skillCount.value = 0
    categoryCount.value = 0
    categoryNames.value = []
    featuredSkills.value = []
  }
})
</script>

<style scoped>
@import url("https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;700;900&display=swap");

.home-shell {
  --ink: #10211c;
  --muted: #5f7169;
  --line: #d7e5de;
  --surface: #ffffff;
  --soft: #edf8f3;
  --soft-2: #f4faf7;
  --brand: #00a878;
  --brand-deep: #087a58;
  --ctyun: #0077e6;
  --shadow: 0 18px 50px rgba(16, 52, 38, 0.1);
  position: relative;
  min-height: 100vh;
  overflow-x: clip;
  background:
    radial-gradient(1200px 500px at 85% -10%, rgba(0, 168, 120, 0.12), transparent 60%),
    radial-gradient(900px 420px at 8% 8%, rgba(0, 119, 230, 0.08), transparent 55%),
    linear-gradient(180deg, #f8fbf9 0%, #f3f8f5 45%, #f7faf8 100%);
  color: var(--ink);
  font-family: "Noto Sans SC", "PingFang SC", "HarmonyOS Sans SC", "Segoe UI", "Microsoft YaHei", sans-serif;
  -webkit-font-smoothing: antialiased;
  text-rendering: optimizeLegibility;
}

.dark .home-shell {
  --ink: #edf5f1;
  --muted: #97aca3;
  --line: #24362e;
  --surface: #101a16;
  --soft: #13241d;
  --soft-2: #0e1713;
  --shadow: 0 18px 50px rgba(0, 0, 0, 0.35);
  background:
    radial-gradient(1000px 420px at 80% -10%, rgba(0, 168, 120, 0.16), transparent 60%),
    linear-gradient(180deg, #09110e 0%, #0b1410 50%, #0a120e 100%);
}

.bg-glow {
  pointer-events: none;
  position: absolute;
  inset: 0 auto auto 0;
  width: min(48vw, 620px);
  height: 520px;
  background: radial-gradient(circle at 30% 30%, rgba(0, 168, 120, 0.12), transparent 70%);
  filter: blur(10px);
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 14px clamp(20px, 5vw, 72px);
  border-bottom: 1px solid color-mix(in srgb, var(--line) 80%, transparent);
  background: color-mix(in srgb, var(--surface) 86%, transparent);
  backdrop-filter: blur(18px);
}

.brand-link,
.brand-link span,
.topbar nav,
.primary-link,
.text-link {
  display: flex;
  align-items: center;
}

.brand-link {
  gap: 11px;
  color: var(--ink);
  text-decoration: none;
}

.brand-link img {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  object-fit: cover;
  box-shadow: 0 8px 18px rgba(0, 102, 204, 0.18);
}

.brand-link span {
  flex-direction: column;
  align-items: flex-start;
}

.brand-link b {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.brand-link small {
  color: var(--muted);
  font-size: 12px;
  font-weight: 500;
}

.topbar nav {
  gap: 16px;
  font-size: 13px;
}

.topbar nav > a:not(.primary-link) {
  color: var(--muted);
  text-decoration: none;
  transition: color 0.15s ease;
}

.topbar nav > a:not(.primary-link):hover {
  color: var(--ink);
}

.icon-control {
  display: grid;
  width: 36px;
  height: 36px;
  place-items: center;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--surface);
  color: var(--ink);
  cursor: pointer;
}

.primary-link,
.secondary-link {
  min-height: 42px;
  justify-content: center;
  gap: 8px;
  border-radius: 12px;
  padding: 0 16px;
  font-size: 13px;
  font-weight: 700;
  text-decoration: none;
  transition: transform 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
}

.primary-link {
  background: linear-gradient(180deg, #12b884 0%, var(--brand) 100%);
  color: #fff;
  box-shadow: 0 10px 24px rgba(0, 168, 120, 0.28);
}

.primary-link:hover,
.secondary-link:hover {
  transform: translateY(-1px);
}

.primary-link.large,
.secondary-link {
  min-height: 48px;
  padding: 0 20px;
}

.secondary-link {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--line);
  background: color-mix(in srgb, var(--surface) 90%, transparent);
  color: var(--ink);
}

.hero-band {
  display: grid;
  grid-template-columns: minmax(320px, 0.88fr) minmax(520px, 1.12fr);
  align-items: center;
  gap: clamp(24px, 3.5vw, 48px);
  min-height: calc(100vh - 72px);
  padding: 48px clamp(20px, 5vw, 72px) 64px;
  background-image:
    linear-gradient(color-mix(in srgb, var(--line) 55%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--line) 55%, transparent) 1px, transparent 1px);
  background-size: 48px 48px;
  background-position: center top;
}

.eyebrow,
.section-label {
  color: var(--brand-deep);
  font: 700 12px/1.4 ui-monospace, "Cascadia Code", Consolas, monospace;
  letter-spacing: 0.08em;
}

.dark .eyebrow,
.dark .section-label {
  color: #41d0a1;
}

.hero-copy h1 {
  max-width: none;
  margin: 18px 0 0;
  font-size: clamp(34px, 3.6vw, 52px);
  line-height: 1.18;
  font-weight: 900;
  letter-spacing: -0.03em;
  text-wrap: pretty;
  word-break: keep-all;
}

.lead {
  max-width: 620px;
  margin-top: 20px;
  color: var(--muted);
  font-size: 16.5px;
  font-weight: 400;
  line-height: 1.85;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 28px;
}

.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 26px;
}

.chip {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 10px;
  border: 1px solid color-mix(in srgb, var(--brand) 28%, var(--line));
  border-radius: 999px;
  background: color-mix(in srgb, var(--brand) 8%, var(--surface));
  color: var(--brand-deep);
  font-size: 12px;
  font-weight: 700;
}

.dark .chip {
  color: #7fe0bc;
}

.chip.soft {
  border-color: var(--line);
  background: var(--surface);
  color: var(--muted);
  font-weight: 600;
}

.chip-divider {
  width: 1px;
  align-self: stretch;
  margin: 4px 2px;
  background: var(--line);
}

.signal-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1px;
  margin-top: 28px;
  border: 1px solid var(--line);
  border-radius: 14px;
  overflow: hidden;
  background: var(--line);
  box-shadow: var(--shadow);
}

.signal-row div {
  padding: 14px 16px;
  background: var(--surface);
}

.signal-row dt {
  color: var(--muted);
  font-size: 11px;
}

.signal-row dd {
  margin-top: 6px;
  font-size: 16px;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.product-stage {
  position: relative;
  transform: none;
  filter: none;
}

.product-stage :deep(.console-preview) {
  width: 100%;
}

.preview-hint {
  margin-top: 12px;
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.change-band,
.capability-band,
.architecture-band,
.market-band,
.value-band,
.deployment-band {
  padding: 84px clamp(20px, 5vw, 72px);
  border-top: 1px solid var(--line);
}

.section-heading {
  max-width: 760px;
}

.section-heading.wide {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  max-width: none;
}

.section-heading.wide > div {
  max-width: 760px;
}

.change-band h2,
.section-heading h2,
.value-copy h2,
.deployment-band h2,
.final-cta h2 {
  margin-top: 10px;
  max-width: none;
  font-size: clamp(28px, 3vw, 40px);
  line-height: 1.28;
  font-weight: 800;
  letter-spacing: -0.02em;
  text-wrap: pretty;
  word-break: keep-all;
}

.section-heading > span {
  max-width: 280px;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.7;
}

.change-grid,
.deployment-options {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-top: 34px;
}

.change-grid article,
.deployment-options article {
  position: relative;
  padding: 24px;
  border: 1px solid var(--line);
  border-radius: 16px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--surface) 92%, var(--soft)), var(--surface));
  box-shadow: 0 10px 28px rgba(16, 52, 38, 0.04);
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.change-grid article:hover,
.deployment-options article:hover,
.capability-grid article:hover,
.market-grid article:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, var(--brand) 40%, var(--line));
  box-shadow: 0 16px 36px rgba(0, 120, 86, 0.1);
}

.card-index,
.deploy-tag {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  margin-bottom: 14px;
  padding: 0 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--brand) 12%, transparent);
  color: var(--brand-deep);
  font: 700 11px/1 ui-monospace, Consolas, monospace;
}

.dark .card-index,
.dark .deploy-tag {
  color: #7fe0bc;
}

.change-grid b,
.deployment-options b {
  display: block;
  font-size: 18px;
  letter-spacing: -0.02em;
}

.change-grid p,
.deployment-options p,
.value-copy > p:last-child {
  margin-top: 10px;
  color: var(--muted);
  line-height: 1.7;
}

.capability-band,
.market-band,
.value-band {
  background: linear-gradient(180deg, var(--soft-2), var(--soft));
}

.capability-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-top: 34px;
}

.capability-grid article {
  display: grid;
  gap: 12px;
  padding: 26px;
  border: 1px solid var(--line);
  border-radius: 18px;
  background: var(--surface);
  box-shadow: 0 10px 28px rgba(16, 52, 38, 0.04);
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.cap-head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cap-head span {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  border-radius: 10px;
  background: color-mix(in srgb, var(--brand) 14%, transparent);
  color: var(--brand-deep);
  font: 800 12px/1 ui-monospace, Consolas, monospace;
}

.dark .cap-head span {
  color: #7fe0bc;
}

.capability-grid h3 {
  font-size: 19px;
  letter-spacing: -0.02em;
}

.capability-grid p {
  color: var(--muted);
  line-height: 1.7;
}

.capability-grid ul {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 4px 0 0;
  padding: 0;
  list-style: none;
}

.capability-grid li {
  min-height: 28px;
  padding: 0 10px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--soft-2);
  color: var(--muted);
  font-size: 12px;
  line-height: 28px;
}

.architecture {
  display: grid;
  grid-template-columns: 1fr auto 1.3fr auto 1fr;
  align-items: stretch;
  gap: 16px;
  margin-top: 36px;
}

.arch-column,
.arch-core {
  display: grid;
  gap: 10px;
  padding: 22px;
  border: 1px solid var(--line);
  border-radius: 18px;
  background: var(--surface);
}

.arch-column b,
.arch-core b {
  letter-spacing: -0.02em;
}

.arch-column span,
.arch-core span {
  padding: 12px 14px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--soft-2);
  color: var(--muted);
  font-size: 13px;
}

.dark .arch-column span,
.dark .arch-core span {
  background: #0c1511;
}

.arch-core {
  border-color: color-mix(in srgb, var(--brand) 45%, var(--line));
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--brand) 10%, var(--surface)), var(--soft));
  box-shadow: 0 16px 40px rgba(0, 120, 86, 0.1);
}

.core-badge {
  width: fit-content;
  padding: 4px 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--brand) 16%, transparent);
  color: var(--brand-deep);
  font: 700 10px/1 ui-monospace, Consolas, monospace;
  letter-spacing: 0.08em;
}

.dark .core-badge {
  color: #7fe0bc;
}

.arch-core > b {
  color: var(--brand-deep);
  font-size: 24px;
}

.dark .arch-core > b {
  color: #7fe0bc;
}

.flow-arrow {
  display: grid;
  place-items: center;
}

.flow-arrow span {
  width: 42px;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--brand), transparent);
  position: relative;
}

.flow-arrow span::after {
  content: "";
  position: absolute;
  right: -1px;
  top: 50%;
  width: 8px;
  height: 8px;
  border-right: 2px solid var(--brand);
  border-top: 2px solid var(--brand);
  transform: translateY(-50%) rotate(45deg);
}

.market-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 22px;
}

.market-meta span {
  min-height: 28px;
  padding: 0 12px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--surface);
  color: var(--muted);
  font-size: 12px;
  line-height: 28px;
}

.market-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-top: 18px;
}

.market-grid article {
  display: grid;
  gap: 10px;
  min-height: 160px;
  padding: 18px;
  border: 1px solid var(--line);
  border-radius: 16px;
  background: var(--surface);
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.skill-top {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 10px;
}

.skill-identity {
  display: flex;
  align-items: start;
  gap: 10px;
  min-width: 0;
}

.skill-avatar {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  place-items: center;
  border-radius: 10px;
  background: linear-gradient(135deg, #12b884, #0077e6);
  color: #fff;
  font-size: 14px;
  font-weight: 800;
}

.skill-top b {
  display: block;
  font-size: 14px;
  letter-spacing: -0.02em;
}

.skill-top em {
  display: block;
  margin-top: 3px;
  color: var(--brand-deep);
  font-style: normal;
  font-size: 11px;
  font-weight: 700;
}

.dark .skill-top em {
  color: #7fe0bc;
}

.market-grid p {
  color: var(--muted);
  font-size: 13px;
  line-height: 1.65;
}

.text-link {
  gap: 6px;
  color: var(--brand-deep);
  font-size: 13px;
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
}

.dark .text-link {
  color: #7fe0bc;
}

.value-band {
  display: grid;
  grid-template-columns: 0.95fr 1.05fr;
  gap: 48px;
  align-items: start;
}

.value-list {
  display: grid;
  gap: 1px;
  border: 1px solid var(--line);
  border-radius: 18px;
  overflow: hidden;
  background: var(--line);
  box-shadow: var(--shadow);
}

.value-list div {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 18px;
  padding: 20px 22px;
  background: var(--surface);
}

.value-list b {
  letter-spacing: -0.02em;
}

.value-list span {
  color: var(--muted);
  line-height: 1.6;
}

.final-cta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 30px;
  padding: 72px clamp(20px, 5vw, 72px);
  background:
    radial-gradient(700px 240px at 15% 20%, rgba(18, 184, 132, 0.22), transparent 60%),
    linear-gradient(135deg, #0d241c 0%, #12352a 55%, #0f2a21 100%);
  color: #fff;
}

.final-cta .section-label.light {
  color: #7fe0bc;
}

.final-cta h2 {
  max-width: 820px;
}

.home-shell footer {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 22px clamp(20px, 5vw, 72px);
  color: var(--muted);
  font-size: 12px;
}

.home-shell footer a {
  color: var(--brand);
  text-decoration: none;
}

@media (max-width: 1180px) {
  .market-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 1024px) {
  .hero-band {
    grid-template-columns: 1fr;
    min-height: auto;
    padding-top: 40px;
  }

  .product-stage {
    transform: none;
  }

  .architecture {
    grid-template-columns: 1fr;
  }

  .flow-arrow span {
    width: 2px;
    height: 36px;
    background: linear-gradient(180deg, transparent, var(--brand), transparent);
  }

  .flow-arrow span::after {
    right: 50%;
    top: auto;
    bottom: -1px;
    transform: translateX(50%) rotate(135deg);
  }

  .value-band,
  .section-heading.wide {
    grid-template-columns: 1fr;
    display: grid;
    align-items: start;
  }
}

@media (max-width: 720px) {
  .topbar nav > a:not(.primary-link),
  .topbar nav .icon-control {
    display: none;
  }

  .topbar {
    padding-inline: 16px;
  }

  .hero-copy h1 {
    font-size: 34px;
  }

  .signal-row,
  .change-grid,
  .capability-grid,
  .deployment-options,
  .market-grid {
    grid-template-columns: 1fr;
  }

  .value-list div {
    grid-template-columns: 1fr;
    gap: 6px;
  }

  .final-cta,
  .home-shell footer {
    align-items: flex-start;
    flex-direction: column;
  }

  .chip-divider {
    display: none;
  }
}
</style>
