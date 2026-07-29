<template>
  <AppLayout v-if="isAuthenticated">
    <SkillMarketCatalog />
  </AppLayout>

  <div v-else class="min-h-screen bg-[#f4f8f6] text-gray-950 dark:bg-dark-950 dark:text-gray-100">
    <header class="border-b border-primary-950/10 bg-white/90 dark:border-dark-700 dark:bg-dark-900/90">
      <div class="mx-auto flex max-w-7xl items-center justify-between px-5 py-4 lg:px-8">
        <router-link to="/home" aria-label="返回 TokenPort 首页">
          <TokenPortLogo :src="brandLogo" :name="siteName" :subtitle="siteSubtitle" />
        </router-link>
        <div class="flex items-center gap-3">
          <button
            type="button"
            class="grid h-9 w-9 place-items-center rounded-xl border border-gray-200 bg-white text-gray-600 transition-colors hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700"
            :title="isDark ? '切换浅色模式' : '切换深色模式'"
            :aria-label="isDark ? '切换浅色模式' : '切换深色模式'"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
          </button>
          <router-link to="/home" class="btn btn-secondary px-2.5 sm:px-4" aria-label="返回首页">
            <Icon name="home" size="sm" />
            <span class="hidden sm:inline">返回首页</span>
          </router-link>
          <router-link to="/login" class="btn btn-primary px-2.5 sm:px-4" aria-label="登录控制台">
            <Icon name="login" size="sm" />
            <span class="hidden sm:inline">登录控制台</span>
          </router-link>
        </div>
      </div>
    </header>
    <main class="mx-auto max-w-7xl px-5 py-8 lg:px-8 lg:py-10">
      <SkillMarketCatalog />
    </main>
    <footer class="border-t border-primary-950/10 bg-white py-6 text-center text-xs text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-400">
      © {{ currentYear }} {{ siteName }} · 技能安装需登录后在 API 密钥接入配置中心完成
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import TokenPortLogo from '@/tokenport/brand/TokenPortLogo.vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import { useThemeToggle } from '@/composables/useThemeToggle'
import { resolveTokenPortLogo, resolveTokenPortName, resolveTokenPortSubtitle } from '@/tokenport/brand/tokenPortBrand'
import SkillMarketCatalog from '@/tokenport/market/SkillMarketCatalog.vue'

const { isDark, toggleTheme } = useThemeToggle()
const authStore = useAuthStore()
const appStore = useAppStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)
const siteName = computed(() => resolveTokenPortName(appStore.cachedPublicSettings?.site_name || appStore.siteName))
const siteSubtitle = computed(() => resolveTokenPortSubtitle(appStore.cachedPublicSettings?.site_subtitle))
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const brandLogo = computed(() => resolveTokenPortLogo(sanitizeUrl(siteLogo.value, { allowRelative: true, allowDataUrl: true })))
const currentYear = new Date().getFullYear()

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>
