<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- SECURITY: homeContent is an admin-only setting. -->
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="tokenport-home min-h-screen bg-[#0c0f11] text-[#f4f0e8]">
    <header class="home-header">
      <router-link to="/home" class="brand">
        <span class="brand-mark">
          <img :src="siteLogo || '/logo.png'" alt="Logo" />
        </span>
        <span class="brand-text">
          <strong>{{ displayName }}</strong>
          <small>Connector Control Plane</small>
        </span>
      </router-link>

      <nav class="header-actions" aria-label="Home actions">
        <LocaleSwitcher />
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="icon-button"
          :title="t('home.viewDocs')"
        >
          <Icon name="book" size="md" />
        </a>
        <button
          type="button"
          class="icon-button"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="toggleTheme"
        >
          <Icon v-if="isDark" name="sun" size="md" />
          <Icon v-else name="moon" size="md" />
        </button>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="header-cta">
          <span>{{ isAuthenticated ? t('home.dashboard') : t('home.login') }}</span>
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </nav>
    </header>

    <main>
      <section class="hero">
        <div class="hero-glow" aria-hidden="true"></div>

        <div class="hero-content">
          <p class="eyebrow">Token 经营闭环 · 智能应用接入 · Skill Market</p>
          <h1>{{ displayName }}</h1>
          <p class="hero-tagline">把接入和经营，收进一个控制面</p>
          <p class="hero-copy">
            部门 API Key 池、真实模型路由、Codex / Claude Code 一键接入与技能包分发，
            统一在一个清晰可审计的控制面里完成。
          </p>
          <div class="hero-actions">
            <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="primary-action">
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              <Icon name="arrowRight" size="md" />
            </router-link>
            <router-link to="/available-channels" class="secondary-action">
              查看可用模型
            </router-link>
          </div>
        </div>

        <div class="ops-panel" aria-label="TokenPort operation snapshot">
          <div class="panel-head">
            <span>LIVE ROUTING</span>
            <b class="status-pill">healthy</b>
          </div>
          <div class="route-stack">
            <div class="route-row">
              <span>研发部 App</span>
              <i></i>
              <b>OpenAI pool</b>
            </div>
            <div class="route-row">
              <span>财务部 App</span>
              <i></i>
              <b>Anthropic pool</b>
            </div>
            <div class="route-row">
              <span>技能包安装</span>
              <i></i>
              <b>12 indexed</b>
            </div>
          </div>
          <div class="metric-grid">
            <div>
              <span>模型来源</span>
              <b>真实同步</b>
            </div>
            <div>
              <span>密钥输出</span>
              <b>已脱敏</b>
            </div>
            <div>
              <span>成本归因</span>
              <b>部门级</b>
            </div>
          </div>
        </div>
      </section>

      <section class="workflow">
        <div class="section-title">
          <p>OPERATING SYSTEM</p>
          <h2>从接入到经营，少填配置，多看真实数据</h2>
        </div>

        <div class="workflow-grid">
          <article>
            <span>01</span>
            <h3>真实模型池</h3>
            <p>连接器只读取 `/channels/available` 和已同步渠道模型。没有模型时提示同步或手动填写，不再展示假 GPT / Claude 选项。</p>
          </article>
          <article>
            <span>02</span>
            <h3>一键接入包</h3>
            <p>按 Codex CLI、Claude Code、OpenCode 等客户端生成配置片段、安装命令和诊断路径，接入过程可复制、可审计。</p>
          </article>
          <article>
            <span>03</span>
            <h3>技能市场</h3>
            <p>默认从同源 `/skill-market/index.json` 加载技能包，包含分类、版本、风险等级、安装目标和 checksum。</p>
          </article>
          <article>
            <span>04</span>
            <h3>经营归因</h3>
            <p>把部门、API Key、模型、技能包和 token 成本连起来，为后续预算、定价和效能评估提供基础数据。</p>
          </article>
        </div>
      </section>

      <section class="pricing-strip">
        <div class="section-title">
          <p>UPSTREAM ACCOUNTING</p>
          <h2>上游账号定价，按真实模型扣钱包</h2>
        </div>

        <div class="pricing-grid">
          <article class="pricing-card">
            <span class="plan-kicker">DeepSeek V4 Pro</span>
            <h3>高能力模型</h3>
            <div class="price-line">
              <b>$0.435 / $0.870</b>
              <span>每 1M 输入 / 输出 token</span>
            </div>
            <p>用于复杂推理、代码生成和长链路智能应用。渠道协议可以是 OpenAI 或 Anthropic，但成本按 DeepSeek 模型计算。</p>
            <ul>
              <li>模型别名：deepseek-v4-pro</li>
              <li>Cache miss 按输入价核算</li>
              <li>Cache hit 按 $0.003625 / 1M token</li>
            </ul>
          </article>

          <article class="pricing-card featured">
            <span class="plan-kicker">DeepSeek V4 Flash</span>
            <h3>高性价比模型</h3>
            <div class="price-line">
              <b>$0.140 / $0.280</b>
              <span>每 1M 输入 / 输出 token</span>
            </div>
            <p>用于日常问答、轻量代码辅助和批量工具调用。适合默认分配给部门应用，先把成本跑出真实曲线。</p>
            <ul>
              <li>模型别名：deepseek-v4-flash</li>
              <li>输入输出分项入账</li>
              <li>同样支持 OpenAI / Anthropic 格式池</li>
            </ul>
          </article>

          <article class="pricing-card">
            <span class="plan-kicker">Wallet Billing</span>
            <h3>部门钱包扣费</h3>
            <div class="price-line">
              <b>Token × 单价</b>
              <span>按 API Key / 部门归集</span>
            </div>
            <p>每次请求都会记录 input、output、cache 与实际扣费，后续可沉淀为预算、成本分摊和效能评估。</p>
            <ul>
              <li>usage_logs 保留成本明细</li>
              <li>actual_cost 直接影响钱包余额</li>
              <li>历史零成本记录不自动补扣</li>
            </ul>
          </article>
        </div>
      </section>

      <section class="provider-strip">
        <div>
          <p>PROVIDER MATRIX</p>
          <h2>{{ t('home.providers.title') }}</h2>
          <span>{{ t('home.providers.description') }}</span>
        </div>
        <ul>
          <li><b>OpenAI</b><span>{{ t('home.providers.supported') }}</span></li>
          <li><b>Anthropic</b><span>{{ t('home.providers.supported') }}</span></li>
          <li><b>Gemini</b><span>{{ t('home.providers.supported') }}</span></li>
          <li><b>Antigravity</b><span>{{ t('home.providers.supported') }}</span></li>
          <li><b>{{ t('home.providers.more') }}</b><span>{{ t('home.providers.soon') }}</span></li>
        </ul>
      </section>
    </main>

    <footer class="home-footer">
      <span>&copy; {{ currentYear }} {{ displayName }}. {{ t('home.footer.allRightsReserved') }}</span>
      <span>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.docs') }}</a>
        <a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a>
      </span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const displayName = computed(() => siteName.value === 'Sub2API' ? 'TokenPort' : siteName.value)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const githubUrl = 'https://github.com/2508756189/sub2api'

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const currentYear = computed(() => new Date().getFullYear())

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.tokenport-home {
  --bg: #0c0f11;
  --panel: rgba(18, 24, 25, 0.82);
  --panel-strong: rgba(24, 31, 31, 0.94);
  --line: rgba(232, 226, 214, 0.14);
  --text: #f4f0e8;
  --muted: #a9b0aa;
  --green: #2ee58c;
  --cyan: #55d7ff;
  --amber: #ffc35a;
  --red: #ff6e5d;
  position: relative;
  overflow-x: hidden;
  font-family: "Aptos", "Segoe UI", sans-serif;
}

.tokenport-home::before {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  content: "";
  background:
    linear-gradient(110deg, rgba(46, 229, 140, 0.11), transparent 24%),
    radial-gradient(circle at 72% 12%, rgba(85, 215, 255, 0.16), transparent 27%),
    radial-gradient(circle at 12% 88%, rgba(255, 195, 90, 0.09), transparent 24%),
    linear-gradient(180deg, #0c0f11, #111512 58%, #0c0f11);
}

.home-header,
.hero,
.workflow,
.pricing-strip,
.provider-strip,
.home-footer {
  position: relative;
  z-index: 1;
}

.home-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 22px clamp(18px, 5vw, 72px);
}

.brand,
.header-actions,
.hero-actions,
.home-footer span,
.provider-strip ul,
.route-row {
  display: flex;
  align-items: center;
}

.brand {
  min-width: 0;
  gap: 12px;
  color: var(--text);
}

.brand-mark {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border: 1px solid rgba(85, 215, 255, 0.34);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.06);
}

.brand-mark img {
  width: 30px;
  height: 30px;
  object-fit: contain;
}

.brand-text {
  display: grid;
  gap: 2px;
}

.brand-text strong {
  font-size: 16px;
  line-height: 1;
}

.brand-text small,
.hero-copy,
.workflow article p,
.provider-strip span,
.metric-grid span,
.route-row span,
.home-footer {
  color: var(--muted);
}

.brand-text small {
  font-size: 11px;
}

.header-actions {
  gap: 10px;
}

.icon-button,
.header-cta,
.primary-action,
.secondary-action {
  border: 1px solid var(--line);
  border-radius: 8px;
  transition: border-color 160ms ease, background 160ms ease, transform 160ms ease;
}

.icon-button {
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text);
}

.header-cta,
.primary-action,
.secondary-action {
  gap: 8px;
  min-height: 40px;
  padding: 0 14px;
  font-size: 13px;
  font-weight: 700;
}

.header-cta,
.primary-action {
  background: var(--green);
  border-color: rgba(46, 229, 140, 0.9);
  color: #06110b;
}

.secondary-action {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text);
}

.icon-button:hover,
.header-cta:hover,
.primary-action:hover,
.secondary-action:hover {
  transform: translateY(-1px);
  border-color: rgba(85, 215, 255, 0.78);
}

.hero {
  display: grid;
  min-height: calc(100vh - 86px);
  grid-template-columns: minmax(0, 1fr) minmax(360px, 520px);
  gap: clamp(28px, 5vw, 68px);
  align-items: center;
  padding: 54px clamp(18px, 5vw, 72px) 84px;
}

.hero-glow {
  position: absolute;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  background:
    radial-gradient(ellipse 60% 50% at 78% 38%, rgba(85, 215, 255, 0.14), transparent 70%),
    radial-gradient(ellipse 50% 60% at 88% 70%, rgba(46, 229, 140, 0.1), transparent 72%);
}

.eyebrow,
.section-title p,
.provider-strip p,
.panel-head span {
  color: var(--amber);
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.hero-content h1 {
  margin-top: 16px;
  color: var(--text);
  font-size: clamp(48px, 7vw, 86px);
  font-weight: 900;
  line-height: 0.94;
  letter-spacing: -0.02em;
}

.hero-tagline {
  margin-top: 14px;
  color: var(--text);
  font-size: clamp(22px, 3vw, 34px);
  font-weight: 800;
  line-height: 1.2;
}

.hero-copy {
  max-width: 560px;
  margin-top: 20px;
  font-size: clamp(15px, 1.4vw, 18px);
  line-height: 1.7;
}

.hero-actions {
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 34px;
}

.primary-action,
.secondary-action {
  display: inline-flex;
  min-height: 46px;
  padding: 0 18px;
}

.ops-panel {
  border: 1px solid var(--line);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.06), transparent),
    var(--panel-strong);
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.34);
  backdrop-filter: blur(20px);
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 18px 14px;
  border-bottom: 1px solid var(--line);
}

.panel-head b {
  border-radius: 999px;
  padding: 4px 9px;
  background: rgba(46, 229, 140, 0.1);
  color: var(--green);
  font-size: 12px;
}

.route-stack {
  display: grid;
  gap: 12px;
  padding: 18px;
}

.route-row {
  min-height: 48px;
  gap: 12px;
}

.route-row i {
  height: 1px;
  flex: 1;
  background: linear-gradient(90deg, rgba(85, 215, 255, 0.15), rgba(46, 229, 140, 0.75), rgba(85, 215, 255, 0.15));
}

.route-row b,
.metric-grid b {
  color: var(--text);
  font-size: 13px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-top: 1px solid var(--line);
}

.metric-grid div {
  display: grid;
  gap: 7px;
  padding: 16px;
  border-right: 1px solid var(--line);
}

.metric-grid div:last-child {
  border-right: 0;
}

.workflow,
.pricing-strip,
.provider-strip {
  padding: 74px clamp(18px, 5vw, 72px);
}

.section-title {
  display: grid;
  max-width: 780px;
  gap: 10px;
}

.section-title h2,
.provider-strip h2 {
  color: var(--text);
  font-size: clamp(28px, 4vw, 42px);
  font-weight: 800;
  line-height: 1.16;
  letter-spacing: -0.01em;
}

.workflow-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  margin-top: 34px;
}

.workflow article {
  min-height: 230px;
  padding: 22px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--panel);
}

.workflow article span {
  color: var(--cyan);
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 12px;
  font-weight: 800;
}

.workflow article h3 {
  margin-top: 34px;
  color: var(--text);
  font-size: 20px;
  font-weight: 800;
}

.workflow article p {
  margin-top: 12px;
  font-size: 14px;
  line-height: 1.72;
}

.pricing-strip {
  border-top: 1px solid var(--line);
  background:
    linear-gradient(90deg, rgba(85, 215, 255, 0.055), transparent 34%),
    rgba(255, 255, 255, 0.018);
}

.pricing-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-top: 34px;
}

.pricing-card {
  display: flex;
  min-height: 360px;
  flex-direction: column;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--panel);
  padding: 24px;
}

.pricing-card.featured {
  border-color: rgba(46, 229, 140, 0.44);
  background:
    linear-gradient(180deg, rgba(46, 229, 140, 0.12), transparent 44%),
    var(--panel-strong);
  box-shadow: 0 22px 70px rgba(46, 229, 140, 0.09);
}

.plan-kicker {
  width: fit-content;
  border: 1px solid rgba(85, 215, 255, 0.26);
  border-radius: 999px;
  padding: 4px 9px;
  color: var(--cyan);
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 11px;
  font-weight: 800;
}

.pricing-card h3 {
  margin-top: 20px;
  color: var(--text);
  font-size: 22px;
  font-weight: 800;
}

.price-line {
  display: grid;
  gap: 6px;
  margin-top: 18px;
  padding-bottom: 22px;
  border-bottom: 1px solid var(--line);
}

.price-line b {
  color: var(--text);
  font-size: clamp(28px, 4vw, 38px);
  font-weight: 900;
  line-height: 1;
}

.price-line span,
.pricing-card p,
.pricing-card li {
  color: var(--muted);
}

.price-line span {
  font-size: 13px;
}

.pricing-card p {
  margin-top: 20px;
  font-size: 14px;
  line-height: 1.72;
}

.pricing-card ul {
  display: grid;
  gap: 10px;
  margin: auto 0 0;
  padding: 22px 0 0;
  list-style: none;
}

.pricing-card li {
  position: relative;
  padding-left: 18px;
  font-size: 13px;
  line-height: 1.5;
}

.pricing-card li::before {
  position: absolute;
  top: 0.62em;
  left: 0;
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: var(--green);
  content: "";
  transform: translateY(-50%);
}

.provider-strip {
  display: grid;
  grid-template-columns: minmax(0, 0.82fr) minmax(0, 1.18fr);
  gap: 34px;
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.025);
}

.provider-strip > div {
  display: grid;
  align-content: center;
  gap: 10px;
}

.provider-strip ul {
  flex-wrap: wrap;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.provider-strip li {
  display: grid;
  min-width: 132px;
  flex: 1;
  gap: 10px;
  padding: 16px;
  border: 1px solid rgba(46, 229, 140, 0.24);
  border-radius: 8px;
  background: rgba(46, 229, 140, 0.06);
}

.provider-strip li:last-child {
  border-color: var(--line);
  background: rgba(255, 255, 255, 0.04);
}

.provider-strip b {
  color: var(--text);
  font-size: 15px;
}

.provider-strip li span {
  color: var(--green);
  font-size: 12px;
}

.provider-strip li:last-child span {
  color: var(--muted);
}

.home-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 28px clamp(18px, 5vw, 72px);
  font-size: 13px;
}

.home-footer span {
  gap: 16px;
}

.home-footer a {
  color: var(--muted);
}

.home-footer a:hover {
  color: var(--text);
}

@media (max-width: 1080px) {
  .hero {
    grid-template-columns: 1fr;
  }

  .ops-panel {
    max-width: 720px;
  }

  .workflow-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .pricing-grid {
    grid-template-columns: 1fr;
  }

  .pricing-card {
    min-height: auto;
  }

  .provider-strip {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .home-header {
    align-items: flex-start;
    padding: 16px;
  }

  .brand-text small,
  .icon-button {
    display: none;
  }

  .header-actions {
    gap: 8px;
  }

  .hero {
    min-height: auto;
    padding: 54px 18px 64px;
  }

  .hero-content h1 {
    font-size: clamp(44px, 13vw, 60px);
  }

  .metric-grid,
  .workflow-grid {
    grid-template-columns: 1fr;
  }

  .metric-grid div {
    border-right: 0;
    border-bottom: 1px solid var(--line);
  }

  .metric-grid div:last-child {
    border-bottom: 0;
  }

  .workflow,
  .pricing-strip,
  .provider-strip {
    padding: 56px 18px;
  }

  .home-footer {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
