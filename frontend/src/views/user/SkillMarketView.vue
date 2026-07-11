<template>
  <AppLayout v-if="isAuthenticated">
    <SkillMarketCatalog />
  </AppLayout>

  <div v-else class="min-h-screen bg-[#f4f8f6] text-gray-950">
    <header class="border-b border-emerald-950/10 bg-white/90">
      <div class="mx-auto flex max-w-7xl items-center justify-between px-5 py-4 lg:px-8">
        <router-link to="/home" class="flex items-center gap-3">
          <img :src="brandLogo" alt="天翼云 TokenPort" class="h-9 w-9 rounded-xl object-cover shadow-sm" />
          <span><b class="block text-base">{{ siteName }}</b><small class="text-xs text-gray-500">{{ siteSubtitle }}</small></span>
        </router-link>
        <div class="flex items-center gap-3">
          <router-link to="/home" class="btn btn-secondary">返回首页</router-link>
          <router-link to="/login" class="btn btn-primary">登录控制台</router-link>
        </div>
      </div>
    </header>
    <main class="mx-auto max-w-7xl px-5 py-8 lg:px-8 lg:py-10">
      <SkillMarketCatalog />
    </main>
    <footer class="border-t border-emerald-950/10 bg-white py-6 text-center text-xs text-gray-500">
      © {{ currentYear }} {{ siteName }} · 技能安装需登录后在 API 密钥接入配置中心完成
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore, useAuthStore } from '@/stores'
import { resolveTokenPortName, resolveTokenPortSubtitle } from '@/tokenport/brand/tokenPortBrand'
import SkillMarketCatalog from '@/tokenport/market/SkillMarketCatalog.vue'

const authStore = useAuthStore()
const appStore = useAppStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)
const siteName = computed(() => resolveTokenPortName(appStore.cachedPublicSettings?.site_name || appStore.siteName))
const siteSubtitle = computed(() => resolveTokenPortSubtitle(appStore.cachedPublicSettings?.site_subtitle))
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const brandLogo = computed(() => {
  const logo = siteLogo.value
  return logo && logo !== '/logo.png' ? logo : '/ctyun-logo.svg'
})
const currentYear = new Date().getFullYear()

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>
