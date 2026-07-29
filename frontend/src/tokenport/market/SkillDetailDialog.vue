<template>
  <BaseDialog
    :show="show"
    :title="skill ? getSkillDisplayName(skill) : '技能详情'"
    width="wide"
    @close="emit('close')"
  >
    <div v-if="skill" class="space-y-5">
      <div class="flex flex-wrap items-center gap-2 text-xs">
        <span class="rounded bg-primary-50 px-2 py-1 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">
          {{ getSkillCategoryName(skill.category, registry) }}
        </span>
        <span :class="['rounded px-2 py-1', riskClass(skill.riskLevel)]">
          {{ getSkillRiskLabel(skill.riskLevel) }}
        </span>
        <span class="rounded bg-gray-100 px-2 py-1 text-gray-600 dark:bg-dark-700 dark:text-gray-300">v{{ skill.version }}</span>
        <span v-for="runtime in skill.runtime || []" :key="runtime" class="rounded border border-gray-200 px-2 py-1 text-gray-500 dark:border-dark-600 dark:text-gray-400">
          {{ runtime }}
        </span>
      </div>

      <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
        {{ getSkillDisplayDescription(skill) }}
      </p>

      <div class="grid gap-4 md:grid-cols-2">
        <DetailList title="适用场景" :items="skill.detail?.useCases || []" empty-text="以技能说明中的触发条件为准" />
        <DetailList title="主要能力" :items="skill.detail?.capabilities || []" empty-text="查看完整说明了解能力范围" />
        <DetailList title="运行依赖" :items="skill.detail?.requirements || []" empty-text="无额外依赖说明" />
        <DetailList title="权限说明" :items="skill.detail?.permissions || []" empty-text="按实际任务授权最小权限" />
      </div>

      <div class="grid gap-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-xs dark:border-dark-700 dark:bg-dark-900/50 sm:grid-cols-2 lg:grid-cols-4">
        <div><span class="text-gray-400">安装位置</span><p class="mt-1 break-all text-gray-700 dark:text-gray-200">{{ installTargets }}</p></div>
        <div><span class="text-gray-400">许可证</span><p class="mt-1 text-gray-700 dark:text-gray-200">{{ skill.license || '待确认' }}</p></div>
        <div><span class="text-gray-400">归档大小</span><p class="mt-1 text-gray-700 dark:text-gray-200">{{ archiveSize }}</p></div>
        <div><span class="text-gray-400">SHA256</span><p class="mt-1 break-all font-mono text-gray-700 dark:text-gray-200">{{ skill.archive.sha256 }}</p></div>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3">
        <button type="button" class="btn btn-secondary" @click="expanded = !expanded">
          {{ expanded ? '收起完整说明' : '展开完整说明' }}
        </button>
        <a v-if="skill.source" :href="skill.source" target="_blank" rel="noopener noreferrer" class="text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
          查看来源
        </a>
      </div>

      <section v-if="expanded" class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <div v-if="loading" class="py-8 text-center text-sm text-gray-500">正在加载完整说明...</div>
        <div v-else-if="error" class="py-8 text-center text-sm text-red-600 dark:text-red-400">{{ error }}</div>
        <article v-else-if="detailHtml" class="skill-markdown prose prose-sm max-w-none dark:prose-invert" v-html="detailHtml"></article>
        <div v-else class="py-8 text-center text-sm text-gray-500">该技能暂未提供完整说明。</div>
      </section>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import BaseDialog from '@/components/common/BaseDialog.vue'
import DetailList from './SkillDetailList.vue'
import {
  fetchSkillDetailMarkdown,
  getSkillCategoryName,
  getSkillDisplayDescription,
  getSkillDisplayName,
  getSkillRiskLabel,
  type SkillMarketEntry,
  type SkillMarketRegistry,
} from '@/api/skillMarket'

const props = defineProps<{
  show: boolean
  skill: SkillMarketEntry | null
  registry?: SkillMarketRegistry | null
  registryUrl?: string
}>()

const emit = defineEmits<{ (e: 'close'): void }>()
const expanded = ref(false)
const loading = ref(false)
const error = ref('')
const markdown = ref('')

const detailHtml = computed(() => markdown.value ? DOMPurify.sanitize(marked.parse(markdown.value) as string) : '')
const archiveSize = computed(() => props.skill?.archive.size ? `${Math.round(props.skill.archive.size / 1024)} KB` : '—')
const installTargets = computed(() => {
  const targets = Object.entries(props.skill?.installTargets || {})
  return targets.length ? targets.map(([runtime, target]) => `${runtime}: ${target}`).join(' · ') : '由客户端决定'
})

watch(() => [props.show, props.skill?.id], async ([show]) => {
  if (!show || !props.skill) return
  expanded.value = false
  markdown.value = ''
  error.value = ''
})

watch(expanded, async (open) => {
  if (!open || !props.skill || markdown.value || loading.value) return
  loading.value = true
  error.value = ''
  try {
    markdown.value = await fetchSkillDetailMarkdown(props.skill, props.registryUrl)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '技能详情加载失败'
  } finally {
    loading.value = false
  }
})

function riskClass(risk?: string) {
  if (risk === 'high') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (risk === 'medium') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
}
</script>

<style>
.skill-markdown h1,
.skill-markdown h2,
.skill-markdown h3 { scroll-margin-top: 1rem; }
.skill-markdown pre { overflow-x: auto; border-radius: 8px; padding: 12px; }
.skill-markdown table { display: block; max-width: 100%; overflow-x: auto; }
</style>
