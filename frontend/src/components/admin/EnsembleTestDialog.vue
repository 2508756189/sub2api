<template>
  <BaseDialog
    :show="show"
    title="Ensemble 测试进度"
    width="extra-wide"
    :close-on-click-outside="false"
    @close="emit('close')"
  >
    <div class="space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-primary-100 bg-primary-50/70 px-4 py-3 dark:border-primary-900/50 dark:bg-primary-900/10">
        <div class="flex items-center gap-2 text-sm font-medium text-primary-800 dark:text-primary-200">
          <Icon v-if="testing" name="refresh" size="md" class="animate-spin" />
          <Icon v-else-if="hasError" name="xCircle" size="md" class="text-red-500" />
          <Icon v-else name="checkCircle" size="md" class="text-green-500" />
          <span>{{ statusText }}</span>
        </div>
        <div class="flex flex-wrap gap-3 text-xs text-gray-500 dark:text-gray-400">
          <span v-if="proposerTotal">候选 {{ proposerSucceeded }} / {{ proposerTotal }}</span>
          <span v-if="result">总 Token {{ result.totalTokens }}</span>
          <span v-if="result">耗时 {{ result.durationText }}</span>
        </div>
      </div>

      <div v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm leading-6 text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300">
        {{ error }}
      </div>

      <div class="grid gap-4 lg:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
        <section class="overflow-hidden rounded-xl border border-gray-200 dark:border-dark-600">
          <div class="border-b border-gray-100 bg-gray-50 px-4 py-3 text-sm font-semibold text-gray-800 dark:border-dark-700 dark:bg-dark-900/50 dark:text-gray-200">
            调用进度
          </div>
          <div class="max-h-[420px] overflow-y-auto p-3">
            <div v-if="memberRows.length" class="space-y-2">
              <div v-for="member in memberRows" :key="`${member.role}-${member.model}`" class="rounded-lg border border-gray-200 px-3 py-3 dark:border-dark-600">
                <div class="flex items-center justify-between gap-3">
                  <div class="min-w-0">
                    <div class="flex items-center gap-2">
                      <span :class="['h-2 w-2 rounded-full', member.status === 'ok' ? 'bg-green-500' : member.status === 'failed' ? 'bg-red-500' : 'bg-amber-400']" />
                      <span class="truncate font-mono text-sm text-gray-800 dark:text-gray-200">{{ member.model }}</span>
                    </div>
                    <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                      {{ member.role === 'aggregator' ? '聚合模型' : '候选模型' }} · {{ member.platform || '未知协议' }}
                    </div>
                  </div>
                  <span :class="['rounded-full px-2 py-0.5 text-xs', member.status === 'ok' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : member.status === 'failed' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300']">
                    {{ member.status === 'ok' ? '成功' : member.status === 'failed' ? '失败' : '进行中' }}
                  </span>
                </div>
                <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
                  <span>耗时 {{ formatDuration(member.duration_ms) }}</span>
                  <span>输入 {{ member.prompt_tokens ?? '—' }}</span>
                  <span>输出 {{ member.completion_tokens ?? '—' }}</span>
                  <span v-if="member.cost !== undefined">成本 ${{ member.cost.toFixed(6) }}</span>
                </div>
                <div v-if="member.error" class="mt-2 break-words text-xs leading-5 text-red-600 dark:text-red-300">{{ member.error }}</div>
              </div>
            </div>
            <div v-else class="py-12 text-center text-sm text-gray-500 dark:text-gray-400">等待网关返回调用计划...</div>
          </div>
        </section>

        <section class="rounded-xl border border-gray-200 dark:border-dark-600">
          <div class="border-b border-gray-100 bg-gray-50 px-4 py-3 text-sm font-semibold text-gray-800 dark:border-dark-700 dark:bg-dark-900/50 dark:text-gray-200">
            最终输出
          </div>
          <div class="max-h-[420px] overflow-y-auto p-4">
            <div v-if="result?.warning" class="mb-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs leading-5 text-amber-700 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-300">
              {{ result.warning }}
            </div>
            <div class="whitespace-pre-wrap text-sm leading-6 text-gray-800 dark:text-gray-200">{{ result?.content || (testing ? '聚合结果将在候选调用完成后显示。' : '没有返回最终内容。') }}</div>
          </div>
        </section>
      </div>

      <details v-if="proposalRows.length" class="rounded-xl border border-gray-200 px-4 py-3 dark:border-dark-600">
        <summary class="cursor-pointer text-sm font-medium text-gray-700 dark:text-gray-300">查看候选原始回答（{{ proposalRows.length }} 份）</summary>
        <div class="mt-3 space-y-3">
          <div v-for="proposal in proposalRows" :key="proposal.model" class="rounded-lg bg-gray-50 p-3 dark:bg-dark-900/40">
            <div class="mb-1 font-mono text-xs text-gray-500">{{ proposal.model }}</div>
            <div class="whitespace-pre-wrap text-sm leading-6 text-gray-700 dark:text-gray-300">{{ proposal.content }}</div>
          </div>
        </div>
      </details>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">{{ testing ? '取消并关闭' : '关闭' }}</button>
        <button v-if="testing" type="button" class="btn btn-danger" @click="emit('cancel')">
          <Icon name="x" size="sm" class="mr-1.5" />取消测试
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { EnsembleMemberStat, EnsembleProgressEvent } from '@/api/admin/ensemble'

interface DialogResult {
  totalTokens: number | string
  durationText: string
  content: string
  proposals: Array<{ model: string; content: string }>
  warning?: string
}

const props = defineProps<{
  show: boolean
  testing: boolean
  events: EnsembleProgressEvent[]
  result: DialogResult | null
  error: string
}>()

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'cancel'): void
}>()

const memberRows = computed(() => {
  const rows = new Map<string, EnsembleMemberStat>()
  for (const event of props.events) {
    if (!event.model || !event.role) continue
    const key = `${event.role}-${event.model}`
    const current = rows.get(key) ?? {
      model: event.model,
      platform: event.platform,
      role: event.role as 'proposer' | 'aggregator',
      status: 'running',
      duration_ms: 0
    }
    if (event.type === 'member_finished' && event.member) {
      rows.set(key, { ...current, ...event.member, platform: event.member.platform ?? event.platform })
    } else {
      rows.set(key, { ...current, platform: event.platform ?? current.platform })
    }
  }
  return [...rows.values()]
})

const proposalRows = computed(() => props.result?.proposals ?? [])
const proposerTotal = computed(() => props.events.find(event => event.type === 'started')?.proposers_total ?? 0)
const proposerSucceeded = computed(() => [...props.events].reverse().find(event => event.type === 'proposers_finished')?.proposers_succeeded ?? 0)
const hasError = computed(() => !!props.error || props.events.some(event => event.type === 'error') || memberRows.value.some(member => member.status === 'failed'))
const statusText = computed(() => {
  if (props.testing) return '正在调用候选模型和聚合模型...'
  if (hasError.value) return '测试结束，但存在失败或回退'
  return '测试完成'
})

function formatDuration(value?: number) {
  return value && value > 0 ? `${(value / 1000).toFixed(1)}s` : '—'
}
</script>
