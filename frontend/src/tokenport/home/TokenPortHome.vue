<template>
  <div class="home-shell">
    <header class="topbar">
      <router-link to="/home" class="brand-link">
        <img :src="siteLogo || '/logo.png'" alt="TokenPort" />
        <span><b>{{ siteName }}</b><small>{{ siteSubtitle }}</small></span>
      </router-link>
      <nav>
        <router-link to="/skill-market">Skill Market</router-link>
        <router-link to="/available-channels">模型与渠道</router-link>
        <button type="button" class="icon-control" :title="isDark ? '切换浅色模式' : '切换深色模式'" @click="toggleTheme">
          <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
        </button>
        <router-link :to="entryPath" class="primary-link">{{ isAuthenticated ? '进入控制台' : '登录平台' }}<Icon name="arrowRight" size="sm" /></router-link>
      </nav>
    </header>

    <main>
      <section class="hero-band">
        <div class="hero-copy">
          <p class="eyebrow">AI ACCESS · TOKEN OPERATIONS · SKILL DELIVERY</p>
          <h1>统一管理模型、Token 与智能体能力</h1>
          <p class="lead">模型越来越多、企业用量越来越大、AI 工具种类繁杂。TokenPort 用一个受控入口连接模型资源、开发工具和可复用技能，让每次调用有归属、有成本、有治理。</p>
          <div class="hero-actions">
            <router-link :to="entryPath" class="primary-link large">{{ isAuthenticated ? '进入经营控制台' : '开始统一接入' }}<Icon name="arrowRight" size="md" /></router-link>
            <router-link to="/skill-market" class="secondary-link">浏览能力市场</router-link>
          </div>
          <dl class="signal-row">
            <div><dt>协议入口</dt><dd>{{ TOKENPORT_PRODUCT.protocols.length }} 类</dd></div>
            <div><dt>工具接入</dt><dd>{{ TOKENPORT_PRODUCT.clients.length }} 类</dd></div>
            <div><dt>能力市场</dt><dd>{{ skillCount }} 个 Skill</dd></div>
          </dl>
        </div>
        <div class="product-frame">
          <div class="frame-bar"><span></span><span></span><span></span><b>TokenPort 经营控制台</b></div>
          <img src="/tokenport-console.png" alt="TokenPort 管理控制台仪表盘" />
        </div>
      </section>

      <section class="change-band">
        <div><p class="section-label">市场变化</p><h2>企业不缺模型入口，缺的是统一经营能力</h2></div>
        <div class="change-grid">
          <article><b>模型供给增加</b><p>国内外模型与兼容渠道持续扩展，单一供应商已无法覆盖所有业务场景。</p></article>
          <article><b>调用规模扩大</b><p>研发、产品、运营与客服同时使用 AI，密钥、成本和权限边界快速复杂化。</p></article>
          <article><b>工具入口分散</b><p>Codex、Claude Code、OpenCode、Gemini CLI 等工具需要不同配置和能力目录。</p></article>
        </div>
      </section>

      <section class="capability-band">
        <div class="section-heading"><p class="section-label">产品能力</p><h2>从统一接入到部门经营，形成完整闭环</h2><span>模型与价格来自当前可用渠道，首页不维护容易失真的静态价格表。</span></div>
        <div class="capability-grid">
          <article v-for="item in capabilities" :key="item.title"><span>{{ item.number }}</span><div><h3>{{ item.title }}</h3><p>{{ item.description }}</p></div></article>
        </div>
      </section>

      <section class="architecture-band">
        <div class="section-heading"><p class="section-label">系统架构</p><h2>一个平台连接企业用户、AI 工具与模型资源</h2></div>
        <div class="architecture">
          <div class="arch-column"><b>使用入口</b><span v-for="item in ['业务应用', '研发工具', '智能体', '部门用户']" :key="item">{{ item }}</span></div>
          <div class="flow-arrow">→</div>
          <div class="arch-core"><b>TokenPort</b><span>统一密钥与权限</span><span>模型路由与定价</span><span>Token 计量与预算</span><span>接入配置与 Skill</span></div>
          <div class="flow-arrow">→</div>
          <div class="arch-column"><b>资源供给</b><span v-for="item in ['公有模型 API', '企业自有模型', '代理与兼容渠道', '行业 Skill 包']" :key="item">{{ item }}</span></div>
        </div>
      </section>

      <section class="value-band">
        <div class="value-copy"><p class="section-label">经营价值</p><h2>把零散的 AI 消耗变成可核算、可优化、可交付的能力</h2><p>统一入口减少重复配置，模型路由降低不必要的高价调用，部门报表明确成本归属，Skill 复用减少重复开发。平台既服务公司内部降本增效，也支持客户私有化部署和行业能力包交付。</p></div>
        <div class="value-list">
          <div><b>成本透明</b><span>按部门、Key、模型和时间追踪 Token 与实际成本</span></div>
          <div><b>效率提升</b><span>一处生成 Codex、Claude、OpenCode 与 CCS 接入配置</span></div>
          <div><b>资产复用</b><span>{{ skillCount }} 个 Skill、{{ categoryCount }} 类能力可查看版本、风险与校验信息</span></div>
          <div><b>客户交付</b><span>支持品牌配置、自有模型资源接入和私有化部署</span></div>
        </div>
      </section>

      <section class="deployment-band">
        <div><p class="section-label">部署方式</p><h2>统一平台直接使用，也支持企业资源独立部署</h2></div>
        <div class="deployment-options"><article><b>平台服务</b><p>企业和团队直接使用统一平台，快速获得模型接入、用量治理与能力市场。</p></article><article><b>私有化交付</b><p>客户已有算力、大模型或安全边界时，可在客户资源环境内完成部署与定制。</p></article><article><b>增值服务</b><p>提供行业 Skill 包、模型资源接入、运营报表、运维诊断与持续升级服务。</p></article></div>
      </section>

      <section class="final-cta"><div><p class="section-label">TOKENPORT</p><h2>让 Token 成为可经营资产，让 Skill 成为可交付资产</h2></div><router-link :to="entryPath" class="primary-link large">{{ isAuthenticated ? '进入控制台' : '登录体验' }}<Icon name="arrowRight" size="md" /></router-link></section>
    </main>

    <footer><span>© {{ currentYear }} {{ siteName }}</span><span>基于 <a :href="TOKENPORT_BRAND.upstreamUrl" target="_blank" rel="noopener">Sub2API</a> 持续构建，遵循原项目许可证。</span></footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAuthStore, useAppStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import { fetchSkillMarket } from '@/api/skillMarket'
import { TOKENPORT_BRAND, TOKENPORT_PRODUCT, resolveTokenPortName, resolveTokenPortSubtitle } from '@/tokenport/brand/tokenPortBrand'

const authStore = useAuthStore()
const appStore = useAppStore()
const skillCount = ref(0)
const categoryCount = ref(0)
const isDark = ref(document.documentElement.classList.contains('dark'))
const siteName = computed(() => resolveTokenPortName(appStore.cachedPublicSettings?.site_name || appStore.siteName))
const siteSubtitle = computed(() => resolveTokenPortSubtitle(appStore.cachedPublicSettings?.site_subtitle))
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const entryPath = computed(() => isAuthenticated.value ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard') : '/login')
const currentYear = new Date().getFullYear()
const capabilities = [
  { number: '01', title: '统一模型接入', description: '连接 OpenAI、Anthropic、Gemini 兼容资源与企业自有模型，按可用性和成本组织服务。' },
  { number: '02', title: 'Token 经营核算', description: '按部门、用户、API Key 和模型归集调用量、Token、成本、额度与预算预警。' },
  { number: '03', title: '智能工具配置', description: '为 Codex、Claude Code、OpenCode、Gemini CLI 与 CCS 生成可检查、可合并的配置。' },
  { number: '04', title: 'Skill 能力市场', description: '以版本、风险、依赖、许可证和 SHA256 管理可复用智能体能力。' },
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
  } catch {
    skillCount.value = 0
    categoryCount.value = 0
  }
})
</script>

<style scoped>
.home-shell{--ink:#10211c;--muted:#61716a;--line:#d9e5df;--surface:#fff;--soft:#edf8f4;--brand:#00a878;min-height:100vh;background:#f7faf8;color:var(--ink);font-family:"Microsoft YaHei","Segoe UI",sans-serif}.dark .home-shell{--ink:#f3f7f5;--muted:#9bb0a7;--line:#263c34;--surface:#111b17;--soft:#142a22;background:#09110e}.topbar{position:sticky;top:0;z-index:20;display:flex;align-items:center;justify-content:space-between;padding:14px clamp(20px,5vw,72px);border-bottom:1px solid var(--line);background:color-mix(in srgb,var(--surface) 92%,transparent);backdrop-filter:blur(18px)}.brand-link,.brand-link span,.topbar nav,.primary-link{display:flex;align-items:center}.brand-link{gap:11px;color:var(--ink)}.brand-link img{width:38px;height:38px;border-radius:8px}.brand-link span{align-items:flex-start;flex-direction:column}.brand-link b{font-size:17px}.brand-link small{color:var(--muted);font-size:11px}.topbar nav{gap:18px;font-size:13px}.topbar nav>a:not(.primary-link){color:var(--muted)}.icon-control{display:grid;width:36px;height:36px;place-items:center;border:1px solid var(--line);border-radius:8px;background:var(--surface)}.primary-link,.secondary-link{min-height:42px;justify-content:center;gap:8px;border-radius:8px;padding:0 16px;font-size:13px;font-weight:700}.primary-link{background:var(--brand);color:#fff}.primary-link.large,.secondary-link{min-height:48px;padding:0 20px}.secondary-link{display:inline-flex;align-items:center;border:1px solid var(--line);background:var(--surface);color:var(--ink)}.hero-band{display:grid;grid-template-columns:minmax(360px,.8fr) minmax(520px,1.2fr);align-items:center;gap:clamp(36px,5vw,84px);min-height:690px;padding:64px clamp(20px,5vw,72px);background-image:linear-gradient(#dce8e2 1px,transparent 1px),linear-gradient(90deg,#dce8e2 1px,transparent 1px);background-size:44px 44px}.dark .hero-band{background-image:linear-gradient(#16241e 1px,transparent 1px),linear-gradient(90deg,#16241e 1px,transparent 1px)}.eyebrow,.section-label{color:#008d68;font:700 12px/1.4 Consolas,monospace}.hero-copy h1{max-width:680px;margin:18px 0 0;font-size:clamp(46px,5vw,76px);line-height:1.04;font-weight:900}.lead{max-width:650px;margin-top:22px;color:var(--muted);font-size:17px;line-height:1.8}.hero-actions{display:flex;flex-wrap:wrap;gap:12px;margin-top:28px}.signal-row{display:grid;grid-template-columns:repeat(3,1fr);gap:1px;margin-top:34px;border:1px solid var(--line);background:var(--line)}.signal-row div{padding:14px;background:var(--surface)}.signal-row dt{color:var(--muted);font-size:11px}.signal-row dd{margin-top:5px;font-size:15px;font-weight:800}.product-frame{overflow:hidden;border:1px solid #b9ccc3;border-radius:8px;background:#fff;box-shadow:0 32px 80px rgba(22,66,49,.2);transform:perspective(1400px) rotateY(-3deg)}.frame-bar{display:flex;align-items:center;gap:6px;height:38px;padding:0 12px;border-bottom:1px solid #dbe6e1;background:#f4f7f6}.frame-bar span{width:8px;height:8px;border-radius:50%;background:#c5d4cd}.frame-bar span:first-child{background:#ff796f}.frame-bar span:nth-child(2){background:#ffc85c}.frame-bar span:nth-child(3){background:#45c98c}.frame-bar b{margin-left:8px;color:#52645c;font-size:11px}.product-frame img{display:block;width:100%;aspect-ratio:16/10;object-fit:cover;object-position:top left}.change-band,.capability-band,.architecture-band,.value-band,.deployment-band{padding:78px clamp(20px,5vw,72px);border-top:1px solid var(--line)}.change-band>div:first-child,.section-heading{max-width:760px}.change-band h2,.section-heading h2,.value-copy h2,.deployment-band h2,.final-cta h2{margin-top:10px;font-size:clamp(30px,3.4vw,46px);line-height:1.18}.change-grid,.deployment-options{display:grid;grid-template-columns:repeat(3,1fr);gap:16px;margin-top:32px}.change-grid article,.deployment-options article{padding:24px;border-left:3px solid var(--brand);background:var(--surface)}.change-grid b,.deployment-options b{font-size:18px}.change-grid p,.deployment-options p,.section-heading span,.value-copy>p:last-child{margin-top:10px;color:var(--muted);line-height:1.7}.capability-band,.value-band{background:var(--soft)}.capability-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:16px;margin-top:34px}.capability-grid article{display:flex;gap:18px;padding:26px;border:1px solid var(--line);background:var(--surface)}.capability-grid article>span{color:var(--brand);font:800 13px Consolas}.capability-grid h3{font-size:19px}.capability-grid p{margin-top:8px;color:var(--muted);line-height:1.7}.architecture{display:grid;grid-template-columns:1fr auto 1.25fr auto 1fr;align-items:stretch;gap:18px;margin-top:36px}.arch-column,.arch-core{display:grid;gap:10px;padding:22px;border:1px solid var(--line);background:var(--surface)}.arch-column span,.arch-core span{padding:11px;border:1px solid var(--line);background:#f8fbf9;color:var(--muted);font-size:13px}.dark .arch-column span,.dark .arch-core span{background:#0c1511}.arch-core{border-color:#59bf9b;background:var(--soft)}.arch-core>b{color:var(--brand);font-size:22px}.flow-arrow{display:grid;place-items:center;color:var(--brand);font-size:26px}.value-band{display:grid;grid-template-columns:.9fr 1.1fr;gap:60px;align-items:start}.value-list{display:grid;gap:1px;border:1px solid var(--line);background:var(--line)}.value-list div{display:grid;grid-template-columns:130px 1fr;gap:20px;padding:20px;background:var(--surface)}.value-list span{color:var(--muted);line-height:1.6}.final-cta{display:flex;align-items:center;justify-content:space-between;gap:30px;padding:64px clamp(20px,5vw,72px);background:#102a21;color:#fff}.final-cta h2{max-width:800px}.home-shell footer{display:flex;justify-content:space-between;gap:20px;padding:22px clamp(20px,5vw,72px);color:var(--muted);font-size:12px}.home-shell footer a{color:var(--brand)}
@media(max-width:1024px){.hero-band{grid-template-columns:1fr;min-height:auto}.product-frame{transform:none}.architecture{grid-template-columns:1fr}.flow-arrow{transform:rotate(90deg)}.value-band{grid-template-columns:1fr}}
@media(max-width:720px){.topbar nav>a:not(.primary-link),.topbar nav .icon-control{display:none}.hero-band{padding-top:48px}.hero-copy h1{font-size:42px}.signal-row,.change-grid,.capability-grid,.deployment-options{grid-template-columns:1fr}.value-list div{grid-template-columns:1fr;gap:6px}.final-cta,.home-shell footer{align-items:flex-start;flex-direction:column}.product-frame img{aspect-ratio:4/3}.topbar{padding-inline:16px}}
</style>
