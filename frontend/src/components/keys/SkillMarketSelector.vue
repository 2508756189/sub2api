<template>
  <section class="rounded-lg border border-gray-200 dark:border-dark-700 bg-white dark:bg-dark-800">
    <button
      type="button"
      class="flex w-full items-center justify-between px-4 py-3 text-left"
      @click="expanded = !expanded"
    >
      <span>
        <span class="block text-sm font-medium text-gray-900 dark:text-gray-100">Skill Market</span>
        <span class="block text-xs text-gray-500 dark:text-gray-400">
          {{ selectedIds.length ? `已选择 ${selectedIds.length} 个技能包，将生成安装脚本` : '从已同步的 Skill Market 选择能力包' }}
        </span>
      </span>
      <span class="text-sm text-gray-500">{{ expanded ? '收起' : '展开' }}</span>
    </button>

    <div v-if="expanded" class="space-y-3 border-t border-gray-100 dark:border-dark-700 px-4 py-4">
      <div class="rounded-md border border-blue-100 bg-blue-50 px-3 py-2 text-xs leading-5 text-blue-700 dark:border-blue-900/50 dark:bg-blue-900/20 dark:text-blue-300">
        技能包不会由浏览器自动写入本机。勾选后会在下方生成 Bash/PowerShell 安装脚本，复制到终端执行后，
        才会下载 zip、校验 SHA256，并安装到 Codex 或 Claude Code 的 skills 目录。
        <router-link to="/skill-market" class="ml-1 font-medium underline underline-offset-2">
          查看完整市场
        </router-link>
      </div>

      <div class="flex flex-col gap-2 md:flex-row">
        <input
          v-model="query"
          class="input input-sm flex-1"
          placeholder="搜索技能名称、场景、标签"
        />
        <select v-model="category" class="input input-sm md:w-48">
          <option value="">全部分类</option>
          <option v-for="item in registry?.categories || []" :key="item.id" :value="item.id">
            {{ item.name }}
          </option>
        </select>
      </div>

      <div v-if="loading" class="text-sm text-gray-500 dark:text-gray-400">正在加载市场...</div>
      <div v-else-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</div>
      <div v-else class="max-h-72 space-y-2 overflow-y-auto pr-1">
        <label
          v-for="skill in filteredSkills"
          :key="skill.id"
          class="flex gap-3 rounded-md border border-gray-100 p-3 text-sm hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800"
        >
          <input
            type="checkbox"
            class="checkbox mt-1"
            :checked="selectedIds.includes(skill.id)"
            @change="toggleSkill(skill, ($event.target as HTMLInputElement).checked)"
          />
          <span class="min-w-0 flex-1">
            <span class="flex flex-wrap items-center gap-2">
              <span class="font-medium text-gray-900 dark:text-gray-100">{{ getSkillDisplayName(skill) }}</span>
              <span class="text-xs text-gray-400">{{ skill.id }}</span>
              <span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300">v{{ skill.version }}</span>
              <span
                class="rounded px-1.5 py-0.5 text-xs"
                :class="riskClass(skill.riskLevel)"
              >
                {{ getSkillRiskLabel(skill.riskLevel) }}
              </span>
            </span>
            <span class="mt-1 block line-clamp-2 text-xs text-gray-500 dark:text-gray-400">
              {{ getSkillDisplayDescription(skill) }}
            </span>
            <button
              type="button"
              class="mt-2 text-xs font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400"
              @click.prevent.stop="selectedSkill = skill"
            >
              查看详情
            </button>
          </span>
        </label>
      </div>
    </div>

    <SkillDetailDialog
      :show="Boolean(selectedSkill)"
      :skill="selectedSkill"
      :registry="registry"
      :registry-url="loadedRegistryUrl || registryUrl"
      @close="selectedSkill = null"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import SkillDetailDialog from '@/tokenport/market/SkillDetailDialog.vue'
import {
  DEFAULT_SKILL_MARKET_REGISTRY_URL,
  fetchSkillMarketWithSource,
  getSkillDisplayDescription,
  getSkillDisplayName,
  getSkillRiskLabel,
  toSkillInstallSelection,
  type SkillInstallSelection,
  type SkillMarketEntry,
  type SkillMarketRegistry,
} from '@/api/skillMarket'

const props = defineProps<{
  modelValue?: SkillInstallSelection[]
  registryUrl?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: SkillInstallSelection[]): void
}>()

const expanded = ref(true)
const loading = ref(false)
const error = ref('')
const query = ref('')
const category = ref('')
const registry = ref<SkillMarketRegistry | null>(null)
const loadedRegistryUrl = ref('')
const selectedSkill = ref<SkillMarketEntry | null>(null)

const selectedIds = computed(() => (props.modelValue || []).map((skill) => skill.id))
const registryUrl = computed(() => props.registryUrl || DEFAULT_SKILL_MARKET_REGISTRY_URL)

const filteredSkills = computed(() => {
  const needle = query.value.trim().toLowerCase()
  return (registry.value?.skills || []).filter((skill) => {
    if (category.value && skill.category !== category.value) return false
    if (!needle) return true
    return [
      skill.name,
      getSkillDisplayName(skill),
      getSkillDisplayDescription(skill),
      skill.description,
      skill.category,
      ...(skill.tags || []),
    ].some((value) => value.toLowerCase().includes(needle))
  })
})

onMounted(async () => {
  loading.value = true
  error.value = ''
  try {
    const result = await fetchSkillMarketWithSource(registryUrl.value)
    registry.value = result.registry
    loadedRegistryUrl.value = result.registryUrl
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Skill Market 加载失败'
  } finally {
    loading.value = false
  }
})

function toggleSkill(skill: SkillMarketEntry, checked: boolean) {
  const current = props.modelValue || []
  if (!checked) {
    emit('update:modelValue', current.filter((item) => item.id !== skill.id))
    return
  }
  if (current.some((item) => item.id === skill.id)) return
  emit('update:modelValue', [
    ...current,
    toSkillInstallSelection(skill, loadedRegistryUrl.value || registryUrl.value),
  ])
}

function riskClass(risk?: string) {
  switch (risk) {
    case 'high':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
    case 'medium':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
    default:
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  }
}
</script>
