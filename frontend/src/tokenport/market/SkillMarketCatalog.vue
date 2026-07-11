<template>
  <div class="space-y-6">
    <section class="overflow-hidden rounded-2xl border border-emerald-950/10 bg-gradient-to-br from-white via-[#f7fbf9] to-[#eef7f3] dark:border-dark-700 dark:from-dark-850 dark:via-dark-850 dark:to-dark-900">
      <div class="grid gap-6 p-5 lg:grid-cols-[1fr_auto] lg:items-end lg:p-7">
        <div>
          <div class="mb-3 inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-white/80 px-3 py-1 text-xs font-semibold text-emerald-700 dark:border-emerald-900/40 dark:bg-dark-900 dark:text-emerald-300">
            <img src="/ctyun-logo.svg" alt="Tianyi Cloud" class="h-4 w-4 rounded" />
            TOKENPORT · SKILL MARKET
          </div>
          <h1 class="text-3xl font-extrabold tracking-tight text-gray-950 dark:text-white">可复用的智能体能力资产</h1>
          <p class="mt-3 max-w-3xl text-[15px] leading-7 text-gray-600 dark:text-gray-300">
            按场景、风险和运行时筛选技能，查看版本、依赖与校验信息。安装操作在 API 密钥的接入配置中心完成。
          </p>
        </div>
        <div class="grid grid-cols-3 gap-3 text-center">
          <div class="rounded-xl border border-emerald-950/8 bg-white px-4 py-3 shadow-sm dark:border-dark-600 dark:bg-dark-900">
            <b class="block text-2xl font-extrabold tracking-tight text-gray-900 dark:text-white">{{ registry?.skills.length || 0 }}</b>
            <span class="text-xs font-medium text-gray-500">技能</span>
          </div>
          <div class="rounded-xl border border-emerald-950/8 bg-white px-4 py-3 shadow-sm dark:border-dark-600 dark:bg-dark-900">
            <b class="block text-2xl font-extrabold tracking-tight text-gray-900 dark:text-white">{{ registry?.categories.length || 0 }}</b>
            <span class="text-xs font-medium text-gray-500">分类</span>
          </div>
          <div class="rounded-xl border border-emerald-950/8 bg-white px-4 py-3 shadow-sm dark:border-dark-600 dark:bg-dark-900">
            <b class="block text-2xl font-extrabold tracking-tight text-emerald-700 dark:text-emerald-300">SHA256</b>
            <span class="text-xs font-medium text-gray-500">校验</span>
          </div>
        </div>
      </div>
    </section>

    <section class="rounded-2xl border border-emerald-950/10 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-850 lg:p-5">
      <div class="grid gap-3 lg:grid-cols-[minmax(240px,1fr)_180px_160px_160px]">
        <div class="relative">
          <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input v-model="query" class="input w-full pl-9" placeholder="搜索技能、场景或标签" />
        </div>
        <Select v-model="runtime" :options="runtimeOptions" :searchable="false" />
        <Select v-model="risk" :options="riskOptions" :searchable="false" />
        <Select v-model="sort" :options="sortOptions" :searchable="false" />
      </div>
      <div class="mt-4 flex flex-wrap gap-2">
        <button
          v-for="category in categoryOptions"
          :key="category.value"
          type="button"
          class="rounded-full border px-3.5 py-1.5 text-sm font-medium transition-all"
          :class="category.value === activeCategory
            ? 'border-emerald-500 bg-emerald-50 text-emerald-700 shadow-sm dark:bg-emerald-900/20 dark:text-emerald-300'
            : 'border-gray-200 text-gray-600 hover:border-emerald-300 hover:bg-emerald-50/50 dark:border-dark-600 dark:text-gray-300'"
          @click="activeCategory = category.value"
        >
          {{ category.label }} <span class="ml-1 text-xs opacity-70">{{ category.count }}</span>
        </button>
      </div>
    </section>

    <div v-if="loading" class="rounded-2xl border border-gray-200 bg-white py-16 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-850">正在加载技能市场...</div>
    <div v-else-if="error" class="rounded-2xl border border-red-200 bg-red-50 py-12 text-center text-sm text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300">
      {{ error }} <button class="ml-2 font-medium underline" @click="load">重试</button>
    </div>
    <div v-else-if="!filteredSkills.length" class="rounded-2xl border border-gray-200 bg-white py-16 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-850">没有符合条件的技能。</div>
    <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <article
        v-for="skill in filteredSkills"
        :key="skill.id"
        class="group flex min-h-[280px] flex-col rounded-2xl border border-emerald-950/10 bg-white p-5 shadow-sm transition duration-200 hover:-translate-y-0.5 hover:border-emerald-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-850 dark:hover:border-emerald-700"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="flex min-w-0 items-start gap-3">
            <div class="flex h-11 w-11 flex-none items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-sky-500 text-sm font-extrabold text-white shadow-sm">
              {{ skillInitial(skill) }}
            </div>
            <div class="min-w-0">
              <h2 class="truncate text-[15px] font-bold tracking-tight text-gray-950 dark:text-white">{{ getSkillDisplayName(skill) }}</h2>
              <p class="mt-1 truncate text-xs text-gray-400">{{ getSkillCategoryName(skill.category, registry) }} · v{{ skill.version }}</p>
            </div>
          </div>
          <span :class="['flex-none rounded-full px-2.5 py-1 text-xs font-semibold', riskClass(skill.riskLevel)]">{{ getSkillRiskLabel(skill.riskLevel) }}</span>
        </div>
        <p class="mt-4 line-clamp-3 text-sm leading-6 text-gray-600 dark:text-gray-300">{{ skill.detail?.summary || getSkillDisplayDescription(skill) }}</p>
        <div class="mt-4 flex flex-wrap gap-1.5">
          <span
            v-for="tag in (skill.tags || []).slice(0, 3)"
            :key="tag"
            class="rounded-full bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300"
          >{{ tag }}</span>
          <span
            v-for="rt in (skill.runtime || []).slice(0, 2)"
            :key="rt"
            class="rounded-full bg-sky-50 px-2.5 py-1 text-xs font-medium text-sky-700 dark:bg-sky-900/20 dark:text-sky-300"
          >{{ rt }}</span>
        </div>
        <div class="mt-auto flex items-end justify-between gap-3 border-t border-gray-100 pt-4 dark:border-dark-700">
          <div class="text-xs text-gray-400">
            <span>{{ formatSize(skill.archive.size) }}</span>
            <span class="mx-2">·</span>
            <span class="font-mono">{{ shortHash(skill.archive.sha256) }}</span>
          </div>
          <button type="button" class="btn btn-secondary btn-sm" @click="selectedSkill = skill">查看详情</button>
        </div>
      </article>
    </div>

    <p v-if="registry" class="text-right text-xs text-gray-400">市场更新：{{ formatGeneratedAt(registry.generatedAt) }}</p>

    <SkillDetailDialog
      :show="Boolean(selectedSkill)"
      :skill="selectedSkill"
      :registry="registry"
      :registry-url="registrySource"
      @close="selectedSkill = null"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import SkillDetailDialog from './SkillDetailDialog.vue'
import {
  DEFAULT_SKILL_MARKET_REGISTRY_URL,
  fetchSkillMarketWithSource,
  getSkillCategoryName,
  getSkillDisplayDescription,
  getSkillDisplayName,
  getSkillRiskLabel,
  type SkillMarketEntry,
  type SkillMarketRegistry,
} from '@/api/skillMarket'

const loading = ref(false)
const error = ref('')
const registry = ref<SkillMarketRegistry | null>(null)
const registrySource = ref(DEFAULT_SKILL_MARKET_REGISTRY_URL)
const selectedSkill = ref<SkillMarketEntry | null>(null)
const query = ref('')
const activeCategory = ref('all')
const runtime = ref<string | number | boolean | null>('all')
const risk = ref<string | number | boolean | null>('all')
const sort = ref<string | number | boolean | null>('name')

const runtimeOptions = [
  { value: 'all', label: '全部运行时' },
  { value: 'codex', label: 'Codex' },
  { value: 'claude', label: 'Claude Code' },
  { value: 'portable', label: '通用运行时' },
]
const riskOptions = [
  { value: 'all', label: '全部风险等级' },
  { value: 'low', label: '低风险' },
  { value: 'medium', label: '中风险' },
  { value: 'high', label: '高风险' },
]
const sortOptions = [
  { value: 'name', label: '按名称排序' },
  { value: 'category', label: '按分类排序' },
  { value: 'risk', label: '按风险排序' },
]

const categoryOptions = computed(() => [
  { value: 'all', label: '全部', count: registry.value?.skills.length || 0 },
  ...(registry.value?.categories || []).map((category) => ({
    value: category.id,
    label: getSkillCategoryName(category.id, registry.value),
    count: registry.value?.skills.filter((skill) => skill.category === category.id).length || 0,
  })),
])

const filteredSkills = computed(() => {
  const needle = query.value.trim().toLowerCase()
  const riskWeight: Record<string, number> = { high: 0, medium: 1, low: 2 }
  return [...(registry.value?.skills || [])]
    .filter((skill) => activeCategory.value === 'all' || skill.category === activeCategory.value)
    .filter((skill) => runtime.value === 'all' || skill.runtime?.includes(String(runtime.value)))
    .filter((skill) => risk.value === 'all' || skill.riskLevel === risk.value)
    .filter((skill) => !needle || [skill.id, skill.name, skill.description, skill.detail?.summary, ...(skill.tags || [])].filter(Boolean).join(' ').toLowerCase().includes(needle))
    .sort((a, b) => {
      if (sort.value === 'category') return a.category.localeCompare(b.category) || a.name.localeCompare(b.name)
      if (sort.value === 'risk') return (riskWeight[a.riskLevel || 'low'] ?? 9) - (riskWeight[b.riskLevel || 'low'] ?? 9)
      return getSkillDisplayName(a).localeCompare(getSkillDisplayName(b), 'zh-CN')
    })
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await fetchSkillMarketWithSource()
    registry.value = result.registry
    registrySource.value = result.registryUrl
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Skill Market 加载失败'
  } finally {
    loading.value = false
  }
}

function riskClass(level?: string) {
  if (level === 'high') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (level === 'medium') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
}

function formatSize(size?: number) {
  return size ? `${Math.round(size / 1024)} KB` : '—'
}

function formatGeneratedAt(value: string) {
  return value ? new Date(value).toLocaleString('zh-CN') : '—'
}

function skillInitial(skill: SkillMarketEntry) {
  const name = getSkillDisplayName(skill).trim()
  return name ? name.slice(0, 1).toUpperCase() : 'S'
}

function shortHash(hash?: string) {
  if (!hash) return '—'
  return `${hash.slice(0, 6)}…${hash.slice(-4)}`
}

onMounted(load)
</script>
