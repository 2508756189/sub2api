<template>
  <div class="tp-auth-shell" :class="{ 'is-dark': isDark }">
    <div class="tp-auth-glow" aria-hidden="true" />

    <header class="tp-auth-topbar">
      <router-link to="/home" class="tp-auth-brand">
        <img
          v-if="settingsLoaded"
          :src="siteLogo || '/ctyun-logo.svg'"
          alt="TokenPort"
        />
        <span>
          <b>{{ siteName }}</b>
          <small>{{ siteSubtitle }}</small>
        </span>
      </router-link>

      <div class="tp-auth-top-actions">
        <button
          type="button"
          class="tp-auth-icon-btn"
          :title="isDark ? '切换浅色模式' : '切换深色模式'"
          @click="toggleTheme"
        >
          <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
        </button>
        <router-link to="/home" class="tp-auth-home-link">
          返回首页
        </router-link>
      </div>
    </header>

    <div class="tp-auth-stage">
      <section class="tp-auth-intro" aria-hidden="false">
        <p class="tp-auth-eyebrow">AI ACCESS · TOKEN · SKILL</p>
        <h1>{{ brandProposition }}</h1>
        <p class="tp-auth-lead">
          把模型账号、API Key、Token 用量和 Skill 放在同一个入口里管理，登录后即可进入控制台。
        </p>
        <ul class="tp-auth-points">
          <li v-for="item in capabilityPoints" :key="item">{{ item }}</li>
        </ul>
        <div class="tp-auth-chips" aria-label="支持协议与工具">
          <span v-for="item in TOKENPORT_PRODUCT.protocols" :key="item" class="chip">{{ item }}</span>
          <span class="chip-divider" />
          <span v-for="item in TOKENPORT_PRODUCT.clients.slice(0, 3)" :key="item" class="chip soft">{{ item }}</span>
        </div>
      </section>

      <section class="tp-auth-panel-wrap">
        <div class="tp-auth-panel">
          <div class="tp-auth-panel-brand">
            <div class="tp-auth-logo">
              <img
                v-if="settingsLoaded"
                :src="siteLogo || '/ctyun-logo.svg'"
                alt="Logo"
              />
            </div>
            <div>
              <h2>{{ siteName }}</h2>
              <p>{{ siteSubtitle }}</p>
            </div>
          </div>

          <div class="tp-auth-panel-body">
            <slot />
          </div>
        </div>

        <div class="tp-auth-footer">
          <slot name="footer" />
        </div>

        <p class="tp-auth-copy">
          &copy; {{ currentYear }} {{ siteName }}
        </p>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import {
  TOKENPORT_BRAND,
  TOKENPORT_PRODUCT,
  resolveTokenPortLogo,
  resolveTokenPortName,
  resolveTokenPortSubtitle,
} from '@/tokenport/brand/tokenPortBrand'

const appStore = useAppStore()

const siteName = computed(() => resolveTokenPortName(appStore.siteName))
const siteLogo = computed(() =>
  resolveTokenPortLogo(
    sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }),
  ),
)
const siteSubtitle = computed(() =>
  resolveTokenPortSubtitle(appStore.cachedPublicSettings?.site_subtitle),
)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const brandProposition = computed(() => TOKENPORT_BRAND.proposition)
const capabilityPoints = computed(() => [
  '统一接入 OpenAI / Anthropic / Gemini 兼容资源',
  '按部门与密钥管理额度、成本与权限',
  '一键配置 Codex、Claude Code 等客户端',
])
const currentYear = computed(() => new Date().getFullYear())
const isDark = ref(document.documentElement.classList.contains('dark'))

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  appStore.fetchPublicSettings()
  isDark.value = document.documentElement.classList.contains('dark')
})
</script>

<style scoped>
.tp-auth-shell {
  --ink: #10211c;
  --muted: #5f7169;
  --line: #d7e5de;
  --surface: #ffffff;
  --soft: #edf8f3;
  --brand: #00a878;
  --brand-deep: #087a58;
  --shadow: 0 18px 50px rgba(16, 52, 38, 0.1);
  position: relative;
  min-height: 100vh;
  overflow-x: clip;
  color: var(--ink);
  background:
    radial-gradient(1200px 500px at 85% -10%, rgba(0, 168, 120, 0.12), transparent 60%),
    radial-gradient(900px 420px at 8% 8%, rgba(0, 119, 230, 0.08), transparent 55%),
    linear-gradient(180deg, #f8fbf9 0%, #f3f8f5 45%, #f7faf8 100%);
  font-family:
    'Noto Sans SC',
    'PingFang SC',
    'HarmonyOS Sans SC',
    'Segoe UI',
    'Microsoft YaHei',
    sans-serif;
  -webkit-font-smoothing: antialiased;
}

.tp-auth-shell.is-dark {
  --ink: #edf5f1;
  --muted: #97aca3;
  --line: #24362e;
  --surface: #101a16;
  --soft: #13241d;
  --shadow: 0 18px 50px rgba(0, 0, 0, 0.35);
  background:
    radial-gradient(1000px 420px at 80% -10%, rgba(0, 168, 120, 0.16), transparent 60%),
    linear-gradient(180deg, #09110e 0%, #0b1410 50%, #0a120e 100%);
}

.tp-auth-glow {
  pointer-events: none;
  position: absolute;
  inset: 0 auto auto 0;
  width: min(48vw, 620px);
  height: 520px;
  background: radial-gradient(circle at 30% 30%, rgba(0, 168, 120, 0.12), transparent 70%);
  filter: blur(10px);
}

.tp-auth-topbar {
  position: relative;
  z-index: 5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px clamp(18px, 4vw, 48px);
  border-bottom: 1px solid color-mix(in srgb, var(--line) 80%, transparent);
  background: color-mix(in srgb, var(--surface) 86%, transparent);
  backdrop-filter: blur(18px);
}

.tp-auth-brand {
  display: flex;
  align-items: center;
  gap: 11px;
  color: var(--ink);
  text-decoration: none;
}

.tp-auth-brand img {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  object-fit: cover;
  box-shadow: 0 8px 18px rgba(0, 102, 204, 0.18);
  background: #fff;
}

.tp-auth-brand span {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.tp-auth-brand b {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.tp-auth-brand small {
  color: var(--muted);
  font-size: 12px;
  font-weight: 500;
}

.tp-auth-top-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tp-auth-icon-btn {
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

.tp-auth-home-link {
  display: inline-flex;
  align-items: center;
  min-height: 36px;
  padding: 0 12px;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--surface);
  color: var(--muted);
  font-size: 13px;
  font-weight: 600;
  text-decoration: none;
}

.tp-auth-home-link:hover {
  color: var(--ink);
}

.tp-auth-stage {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: minmax(300px, 1.15fr) minmax(340px, 420px);
  gap: clamp(32px, 4vw, 64px);
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 70px);
  width: min(1180px, calc(100% - 48px));
  max-width: 1180px;
  margin: 0 auto;
  padding: 32px 0 40px;
  background-image:
    linear-gradient(color-mix(in srgb, var(--line) 55%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--line) 55%, transparent) 1px, transparent 1px);
  background-size: 48px 48px;
  background-position: center top;
}

.tp-auth-intro {
  max-width: none;
  min-width: 0;
}

.tp-auth-eyebrow {
  margin: 0;
  color: var(--brand-deep);
  font: 700 12px/1.4 ui-monospace, 'Cascadia Code', Consolas, monospace;
  letter-spacing: 0.08em;
}

.tp-auth-shell.is-dark .tp-auth-eyebrow {
  color: #41d0a1;
}

.tp-auth-intro h1 {
  margin: 12px 0 0;
  max-width: 18em;
  font-size: clamp(28px, 2.6vw, 40px);
  line-height: 1.28;
  font-weight: 900;
  letter-spacing: -0.03em;
  /* Prefer natural phrase breaks; avoid mid-word splits on CJK/Latin mix */
  overflow-wrap: normal;
  word-break: normal;
  line-break: strict;
  text-wrap: pretty;
}

.tp-auth-lead {
  max-width: 36em;
  margin-top: 14px;
  color: var(--muted);
  font-size: 15px;
  line-height: 1.75;
  overflow-wrap: break-word;
}

.tp-auth-points {
  margin: 16px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 8px;
}

.tp-auth-points li {
  position: relative;
  padding-left: 22px;
  color: var(--ink);
  font-size: 14px;
  line-height: 1.6;
}

.tp-auth-points li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0.55em;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: linear-gradient(180deg, #12b884 0%, var(--brand) 100%);
  box-shadow: 0 0 0 3px rgba(0, 168, 120, 0.12);
}

.tp-auth-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 18px;
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

.tp-auth-shell.is-dark .chip {
  color: #7fe0bc;
}

.chip.soft {
  border-color: var(--line);
  background: var(--surface);
  color: var(--muted);
}

.chip-divider {
  width: 1px;
  align-self: stretch;
  background: var(--line);
  margin: 4px 2px;
}

.tp-auth-panel-wrap {
  width: 100%;
}

.tp-auth-panel {
  border: 1px solid color-mix(in srgb, var(--line) 90%, transparent);
  border-radius: 20px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--surface) 96%, #fff) 0%, var(--surface) 100%);
  box-shadow: var(--shadow);
  backdrop-filter: blur(12px);
  overflow: hidden;
}

.tp-auth-shell.is-dark .tp-auth-panel {
  background: linear-gradient(180deg, #13241d 0%, var(--surface) 100%);
}

.tp-auth-panel-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px 14px;
  border-bottom: 1px solid color-mix(in srgb, var(--line) 80%, transparent);
  background: linear-gradient(180deg, color-mix(in srgb, var(--brand) 6%, transparent), transparent);
}

.tp-auth-logo {
  display: grid;
  place-items: center;
  width: 42px;
  height: 42px;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 8px 18px rgba(0, 168, 120, 0.12);
}

.tp-auth-logo img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.tp-auth-panel-brand h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--ink);
}

.tp-auth-panel-brand p {
  margin: 2px 0 0;
  color: var(--muted);
  font-size: 12px;
}

.tp-auth-panel-body {
  padding: 18px 20px 22px;
}

/* Harmonize common form controls inside auth cards */
.tp-auth-panel-body :deep(.input),
.tp-auth-panel-body :deep(input),
.tp-auth-panel-body :deep(textarea) {
  border-radius: 12px;
  border-color: var(--line);
  background: #ffffff;
  color: #10211c;
  -webkit-text-fill-color: #10211c;
  caret-color: #087a58;
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.55) inset;
}

.tp-auth-panel-body :deep(.input::placeholder),
.tp-auth-panel-body :deep(input::placeholder) {
  color: #8a9a93;
  -webkit-text-fill-color: #8a9a93;
  opacity: 1;
}

.tp-auth-shell.is-dark .tp-auth-panel-body :deep(.input),
.tp-auth-shell.is-dark .tp-auth-panel-body :deep(input),
.tp-auth-shell.is-dark .tp-auth-panel-body :deep(textarea) {
  background: #0e1713;
  color: #edf5f1;
  -webkit-text-fill-color: #edf5f1;
  caret-color: #41d0a1;
  box-shadow: none;
}

.tp-auth-shell.is-dark .tp-auth-panel-body :deep(.input::placeholder),
.tp-auth-shell.is-dark .tp-auth-panel-body :deep(input::placeholder) {
  color: #7a9087;
  -webkit-text-fill-color: #7a9087;
}

/* Chrome autofill often washes out typed text on mint cards */
.tp-auth-panel-body :deep(input:-webkit-autofill),
.tp-auth-panel-body :deep(input:-webkit-autofill:hover),
.tp-auth-panel-body :deep(input:-webkit-autofill:focus) {
  -webkit-text-fill-color: #10211c !important;
  caret-color: #087a58;
  transition: background-color 99999s ease-out;
  box-shadow: 0 0 0 1000px #f3fbf7 inset !important;
}

.tp-auth-shell.is-dark .tp-auth-panel-body :deep(input:-webkit-autofill),
.tp-auth-shell.is-dark .tp-auth-panel-body :deep(input:-webkit-autofill:hover),
.tp-auth-shell.is-dark .tp-auth-panel-body :deep(input:-webkit-autofill:focus) {
  -webkit-text-fill-color: #edf5f1 !important;
  caret-color: #41d0a1;
  box-shadow: 0 0 0 1000px #0e1713 inset !important;
}

.tp-auth-panel-body :deep(.input:focus),
.tp-auth-panel-body :deep(input:focus) {
  border-color: var(--brand);
  box-shadow: 0 0 0 3px rgba(0, 168, 120, 0.14);
}

.tp-auth-panel-body :deep(.btn-primary) {
  border-radius: 12px;
  background: linear-gradient(180deg, #12b884 0%, var(--brand) 100%);
  color: #fff;
  box-shadow: 0 10px 24px rgba(0, 168, 120, 0.28);
  font-weight: 700;
}

.tp-auth-panel-body :deep(.btn-primary:hover) {
  background: linear-gradient(180deg, #17c48d 0%, #0b9a6d 100%);
}

.tp-auth-panel-body :deep(h2) {
  letter-spacing: -0.02em;
}

.tp-auth-footer {
  margin-top: 12px;
  text-align: center;
  font-size: 14px;
  color: var(--muted);
}

.tp-auth-footer :deep(a) {
  color: var(--brand-deep);
  font-weight: 700;
  text-decoration: none;
}

.tp-auth-shell.is-dark .tp-auth-footer :deep(a) {
  color: #7fe0bc;
}

.tp-auth-copy {
  margin-top: 14px;
  text-align: center;
  color: color-mix(in srgb, var(--muted) 85%, transparent);
  font-size: 12px;
}

@media (min-width: 1400px) {
  .tp-auth-stage {
    width: min(1280px, calc(100% - 64px));
    max-width: 1280px;
    gap: 72px;
  }
}

@media (max-width: 960px) {
  .tp-auth-stage {
    grid-template-columns: minmax(0, 440px);
    justify-content: center;
    width: min(100% - 32px, 440px);
    min-height: auto;
    padding-top: 22px;
  }

  .tp-auth-intro h1 {
    max-width: none;
    font-size: clamp(26px, 7vw, 34px);
  }
}

@media (max-width: 640px) {
  .tp-auth-brand small {
    display: none;
  }

  .tp-auth-home-link {
    display: none;
  }

  .tp-auth-panel-body {
    padding: 18px 16px 22px;
  }

  .tp-auth-panel-brand {
    padding: 16px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .tp-auth-panel-body :deep(.btn-primary) {
    transition: none;
  }
}
</style>
