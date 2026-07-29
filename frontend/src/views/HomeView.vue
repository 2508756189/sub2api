<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe v-if="isHomeContentUrl" :src="safeHomeContentUrl" class="h-screen w-full border-0" allowfullscreen />
    <div v-else v-html="homeContent" />
  </div>
  <TokenPortHome v-else :site-logo="siteLogo" :doc-url="docUrl" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores'
import TokenPortHome from '@/tokenport/home/TokenPortHome.vue'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const safeHomeContentUrl = computed(() => sanitizeUrl(homeContent.value.trim()))
const isHomeContentUrl = computed(() => Boolean(safeHomeContentUrl.value))
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
</script>
