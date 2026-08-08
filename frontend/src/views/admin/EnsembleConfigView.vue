<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" :disabled="loading" title="刷新" @click="init">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button v-if="!isNew" type="button" class="btn btn-secondary" :disabled="saving || !canSaveAsNew" @click="save(true)">
            <Icon name="plus" size="md" class="mr-1.5" />
            另存为新分组
          </button>
          <button type="button" class="btn btn-primary" :disabled="saving || !canSave" @click="save()">
            <Icon name="check" size="md" class="mr-1.5" />
            {{ saving ? '保存中…' : '保存配置' }}
          </button>
        </div>
      </template>

      <template #table>
        <div class="table-wrapper">
          <div class="space-y-5 p-6">
            <transition name="fade">
              <div v-if="notice" :class="['notice', noticeClass]">
                <Icon :name="noticeType === 'ok' ? 'checkCircle' : noticeType === 'warn' ? 'exclamationTriangle' : 'xCircle'" size="md" />
                <span>{{ notice }}</span>
              </div>
            </transition>

            <div class="intro-panel">
              <div>
                <div class="flex flex-wrap items-center gap-2">
                  <span class="intro-kicker">原生组内聚合</span>
                  <span class="intro-rule">内置组内调度</span>
                </div>
                <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                  先创建一个 Ensemble 分组，再把来源分组的账号复制到这个新分组。请求进入新分组后，只在该分组绑定的账号范围内并行调用候选模型。
                </p>
              </div>
              <div class="intro-flow">
                <span>来源分组账号</span><Icon name="arrowRight" size="sm" /><span>候选模型</span><Icon name="arrowRight" size="sm" /><span>聚合输出</span>
              </div>
            </div>

            <div class="readiness-bar">
              <div class="readiness-item">
                <span class="readiness-label">当前入口</span>
                <strong>{{ isNew ? '尚未保存' : 'ensemble' }}</strong>
                <span class="readiness-meta">{{ isNew ? '保存后生成可调用分组' : `${selectedEnsembleGroup?.account_count ?? 0} 个绑定账号` }}</span>
              </div>
              <div class="readiness-item">
                <span class="readiness-label">候选调度</span>
                <strong>{{ proposers.length }} / {{ MAX_PROPOSERS }}</strong>
                <span class="readiness-meta">至少 {{ options.minProposers }} 个成功</span>
              </div>
              <div class="readiness-item">
                <span class="readiness-label">测试方式</span>
                <strong>分组 API Key</strong>
                <span class="readiness-meta">账号直连测试不适用于 ensemble</span>
              </div>
            </div>

            <div class="grid gap-5 xl:grid-cols-[minmax(0,1.15fr)_minmax(360px,0.85fr)]">
              <section class="section-card">
                <div class="section-header">
                  <div>
                    <h2 class="section-title"><span class="step-no">1</span>创建或编辑 Ensemble 分组</h2>
                    <p class="section-desc">分组是实际的计费和路由边界；来源组只用于创建时复制账号绑定。</p>
                  </div>
                  <span :class="['state-badge', isNew ? 'state-new' : 'state-existing']">
                    {{ isNew ? '新建分组' : '编辑已有分组' }}
                  </span>
                </div>

                <div class="grid gap-4 md:grid-cols-2">
                  <div class="md:col-span-2">
                    <label class="field-label">编辑中的 Ensemble 分组</label>
                    <Select
                      v-model="selectedEnsembleGroupId"
                      :options="ensembleGroupOptions"
                      placeholder="新建 Ensemble 分组"
                      searchable="auto"
                      aria-label="选择 Ensemble 分组"
                      @change="onEnsembleGroupChange"
                    />
                    <p class="field-hint">选择已有分组会加载它当前保存的候选和聚合设置；选择“新建”不会覆盖已有分组。</p>
                  </div>

                  <div>
                    <label class="field-label">分组名称 <span class="required">*</span></label>
                    <input v-model="draft.name" class="field-input" placeholder="例如：研发 Ensemble" maxlength="80" />
                  </div>
                  <div>
                    <label class="field-label">下游计费倍率</label>
                    <input v-model.number="draft.rateMultiplier" class="field-input" type="number" min="0.01" step="0.01" />
                    <p class="field-hint">候选和聚合的实际上游调用会分别产生用量；此倍率沿用普通分组计费。</p>
                  </div>
                  <div class="md:col-span-2">
                    <label class="field-label">描述</label>
                    <textarea v-model="draft.description" class="field-input min-h-20 resize-y" placeholder="说明这个 Ensemble 分组的用途" maxlength="240" />
                  </div>
                </div>

                <div v-if="isNew" class="source-panel mt-5">
                  <div class="mb-3 flex items-start justify-between gap-3">
                    <div>
                      <h3 class="subsection-title">账号来源分组 <span class="required">*</span></h3>
                      <p class="section-desc">可添加多个普通分组；创建后这些分组的账号会绑定到新的 Ensemble 分组。</p>
                    </div>
                    <span class="badge-count">已选 {{ selectedSourceGroupIds.length }}</span>
                  </div>

                  <div v-if="selectedSourceGroups.length" class="flex flex-wrap gap-2">
                    <span v-for="group in selectedSourceGroups" :key="group.id" class="source-chip">
                      <Icon name="server" size="xs" class="text-primary-500" />
                      <span class="truncate">{{ group.name }}</span>
                      <span class="chip-platform">{{ group.platform }}</span>
                      <button type="button" class="chip-remove" :aria-label="`移除 ${group.name}`" @click="removeSourceGroup(group.id)">
                        <Icon name="x" size="xs" />
                      </button>
                    </span>
                  </div>
                  <div v-else class="empty-inline">尚未选择来源分组。没有来源账号时，新的 Ensemble 分组无法调用任何模型。</div>

                  <div class="relative mt-3">
                    <button type="button" class="add-button" :disabled="sourceGroupOptions.length === 0" @click="toggleSourcePicker">
                      <Icon name="plus" size="sm" />
                      {{ sourceGroupOptions.length ? '添加来源分组' : '没有可添加的普通分组' }}
                    </button>
                    <div v-if="sourcePickerOpen" class="picker-panel">
                      <div class="picker-search">
                        <Icon name="search" size="sm" class="text-gray-400" />
                        <input ref="sourcePickerInput" v-model="sourceQuery" class="picker-input" placeholder="搜索分组名称或平台" @keydown.esc="sourcePickerOpen = false" />
                      </div>
                      <button v-for="group in filteredSourceGroupOptions" :key="group.id" type="button" class="picker-option" @click="addSourceGroup(group.id)">
                        <span class="min-w-0 truncate">{{ group.name }}</span>
                        <span class="chip-platform">{{ group.platform }}</span>
                      </button>
                      <div v-if="filteredSourceGroupOptions.length === 0" class="picker-empty">没有匹配的来源分组</div>
                    </div>
                  </div>
                  <div v-if="billingChannel" class="mt-3 rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-xs text-green-700 dark:border-green-900/50 dark:bg-green-900/10 dark:text-green-300">
                    计费渠道：<strong>{{ billingChannel.name }}</strong>。新 Ensemble 分组保存后会自动加入该渠道，并沿用其中的模型定价。
                  </div>
                  <div v-else-if="selectedSourceGroupIds.length" class="mt-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs leading-5 text-amber-700 dark:border-amber-900/50 dark:bg-amber-900/10 dark:text-amber-300">
                    当前来源分组不属于同一个启用渠道。请调整来源分组，保证计费规则唯一后再保存。
                  </div>
                </div>

                <div v-else class="existing-note mt-5">
                  <Icon name="infoCircle" size="md" class="flex-shrink-0 text-primary-500" />
                  <div>
                    <p>这个分组的账号绑定已在分组管理中保存。当前页面只修改 Ensemble 成员和聚合参数，不会偷偷跨组读取账号。</p>
                    <div v-if="selectedSourceGroups.length" class="mt-2 flex flex-wrap gap-1.5">
                      <span v-for="group in selectedSourceGroups" :key="group.id" class="source-chip">
                        <span class="truncate">来源：{{ group.name }}</span>
                        <span class="chip-platform">{{ group.platform }}</span>
                      </span>
                    </div>
                    <p v-if="billingChannel" class="mt-2">计费渠道：{{ billingChannel.name }}</p>
                    <p v-else class="mt-2 text-amber-700 dark:text-amber-300">尚未关联启用渠道，模型定价和实际调用不可用。</p>
                  </div>
                </div>
              </section>

              <section class="section-card">
                <div class="section-header">
                  <div>
                    <h2 class="section-title"><span class="step-no">2</span>候选模型</h2>
                    <p class="section-desc">模型来源于所选来源分组关联渠道的真实定价，不读取平台内置默认清单。</p>
                  </div>
                  <span :class="['badge-count', proposers.length >= 2 ? '' : 'badge-danger']">{{ proposers.length }} / {{ MAX_PROPOSERS }}</span>
                </div>

                <div v-if="proposers.length" class="space-y-2">
                  <div v-for="(model, index) in proposers" :key="model" class="member-row">
                    <span class="member-index">{{ index + 1 }}</span>
                    <span class="min-w-0 flex-1 truncate font-mono text-sm">{{ model }}</span>
                    <span class="member-role">候选</span>
                    <button type="button" class="icon-button" :aria-label="`移除 ${model}`" title="移除" @click="removeProposer(model)">
                      <Icon name="x" size="sm" />
                    </button>
                  </div>
                </div>
                <div v-else class="empty-inline">还没有添加候选模型，请点击下面的加号按钮。</div>

                <div class="relative mt-3">
                  <button type="button" class="add-button" :disabled="overProposerLimit || availableModels.length === 0" @click="toggleModelPicker">
                    <Icon name="plus" size="sm" />
                    {{ availableModels.length === 0 ? '暂无可用模型' : overProposerLimit ? '已达到候选上限' : '添加候选模型' }}
                  </button>
                  <div v-if="modelPickerOpen" class="picker-panel">
                    <div class="picker-search">
                      <Icon name="search" size="sm" class="text-gray-400" />
                      <input ref="modelPickerInput" v-model="modelQuery" class="picker-input" placeholder="搜索模型名称" @keydown.esc="modelPickerOpen = false" />
                    </div>
                    <button v-for="model in filteredModelOptions" :key="model" type="button" class="picker-option" @click="addProposer(model)">
                      <span class="truncate font-mono text-xs">{{ model }}</span>
                      <Icon name="plus" size="sm" class="text-primary-500" />
                    </button>
                    <div v-if="filteredModelOptions.length === 0" class="picker-empty">没有匹配的模型</div>
                  </div>
                </div>
                <p class="field-hint mt-2">建议至少选择 2 个候选，最多 {{ MAX_PROPOSERS }} 个。每增加一个候选，实际上游调用和成本也会增加。</p>
              </section>
            </div>

            <section class="section-card">
              <div class="section-header">
                <div>
                  <h2 class="section-title"><span class="step-no">3</span>聚合模型</h2>
                  <p class="section-desc">聚合模型读取成功的候选答案并输出终稿；不选择时返回最长的候选答案。</p>
                </div>
              </div>
              <div class="max-w-xl">
                <Select
                  v-model="aggregator"
                  :options="aggregatorOptions"
                  placeholder="不使用聚合（返回最长候选）"
                  searchable="auto"
                  clearable
                  aria-label="选择聚合模型"
                />
                <p class="field-hint">聚合模型可以是候选之一，也可以是同一来源范围内的另一个已定价模型。</p>
              </div>
            </section>

            <section class="section-card">
              <div class="section-header">
                <div>
                  <h2 class="section-title"><span class="step-no">4</span>运行参数</h2>
                  <p class="section-desc">这些设置保存到 Ensemble 分组本身，不会修改来源分组。</p>
                </div>
              </div>
              <div class="grid gap-4 sm:grid-cols-3">
                <div>
                  <label class="field-label">最少成功候选数</label>
                  <input v-model.number="options.minProposers" class="field-input" type="number" min="1" :max="Math.max(1, proposers.length)" />
                  <p class="field-hint">低于此数量时请求返回失败。</p>
                </div>
                <div>
                  <label class="field-label">单模型超时（秒）</label>
                  <input v-model.number="options.timeoutSeconds" class="field-input" type="number" min="1" max="600" />
                </div>
                <div>
                  <label class="field-label">单模型最大输出 Token</label>
                  <input v-model.number="options.maxTokens" class="field-input" type="number" min="0" placeholder="0 表示不限制" />
                </div>
              </div>
              <div class="option-toggle mt-4">
                <div>
                  <div class="text-sm font-medium text-gray-800 dark:text-gray-200">返回 Ensemble 执行明细</div>
                  <p class="field-hint">开启后响应附带每个候选/聚合调用的耗时、Token 和成本，执行轨迹里也会显示真实模型名。候选答案原文只在后台测试里返回，不会进入正常响应，因此开启的额外开销很小。关闭后轨迹改用「模型 1 / 模型 2」指代，适合把该分组转售给第三方的场景。</p>
                </div>
                <Toggle v-model="options.exposeMetadata" aria-label="返回 Ensemble 执行明细" />
              </div>
              <div class="option-toggle mt-4">
                <div>
                  <div class="text-sm font-medium text-gray-800 dark:text-gray-200">流式输出执行过程（推理轨迹）</div>
                  <p class="field-hint">开启后流式调用会通过 reasoning_content 实时展示候选模型调用、完成耗时与失败原因，未启用该字段的客户端会自动忽略。该轨迹仅用于展示，不会进入模型上下文。聚合模型自己的思考过程不受此开关控制，始终按单模型调用的方式实时透传。</p>
                </div>
                <Toggle v-model="options.streamTrace" aria-label="流式输出执行过程" />
              </div>
              <div class="estimate-bar mt-4">
                <span>预计上游调用次数</span>
                <strong>{{ proposers.length + (aggregator ? 1 : 0) }} 次</strong>
                <span class="text-gray-500 dark:text-gray-400">候选并行调用{{ aggregator ? '，再调用 1 次聚合模型' : '，不执行聚合' }}</span>
              </div>
            </section>

            <section class="section-card">
              <div class="section-header">
                <div>
                  <h2 class="section-title"><Icon name="play" size="md" class="mr-2 text-primary-500" />保存后测试</h2>
                  <p class="section-desc">测试不会保存 API Key。请输入一个已经绑定到当前 Ensemble 分组的 API Key，页面会直接请求本地网关。</p>
                </div>
              </div>
              <div class="test-key-field">
                <label class="field-label">测试用 API Key</label>
                <div class="test-key-row">
                  <input v-model="testApiKey" class="field-input font-mono" type="password" autocomplete="off" placeholder="粘贴当前 Ensemble 分组的 API Key" />
                  <button type="button" class="btn btn-secondary test-run-button" :disabled="testing || !canTest" @click="runTest">
                    <Icon name="play" size="md" class="mr-1.5" />{{ testing ? '测试中…' : '运行一次测试' }}
                  </button>
                </div>
                <p class="field-hint">这是测试请求的临时凭证，不属于 Ensemble 配置，也不会发送给任何外部 Router。请求固定使用虚拟模型 <code>ensemble</code>。</p>
              </div>
            </section>

            <section v-if="testResult" class="section-card">
              <div class="section-header">
                <div>
                  <h2 class="section-title"><Icon name="chartBar" size="md" class="mr-2 text-primary-500" />最近一次测试结果</h2>
                  <p class="section-desc">每一行对应一次实际上游调用；聚合调用也会单独列出。</p>
                </div>
                <div class="flex flex-wrap gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <span>{{ testResult.metadataPresent ? `成功 ${testResult.successCount} / ${testResult.members.length}` : '未读取到 Ensemble 执行明细' }}</span>
                  <span>耗时 {{ testResult.durationText }}</span>
                  <span>总 Token {{ testResult.totalTokens }}</span>
                </div>
              </div>
              <div v-if="testResult.warning" class="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm leading-6 text-amber-700 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-300">
                {{ testResult.warning }}
              </div>
              <div class="overflow-x-auto">
                <table class="result-table">
                  <thead>
                    <tr><th>角色</th><th>模型</th><th>状态</th><th class="text-right">耗时</th><th class="text-right">输入 Token</th><th class="text-right">输出 Token</th><th class="text-right">成本</th></tr>
                  </thead>
                  <tbody>
                    <tr v-for="(member, index) in testResult.members" :key="`${member.role}-${member.model}-${index}`">
                      <td>{{ member.role === 'aggregator' ? '聚合' : `候选 ${index + 1}` }}</td>
                      <td class="font-mono text-xs">{{ member.model }}</td>
                      <td><span :class="member.status === 'ok' ? 'pill-ok' : 'pill-fail'">{{ member.status === 'ok' ? '成功' : '失败' }}</span><span v-if="member.error" class="ml-2 text-xs text-red-500">{{ member.error }}</span></td>
                      <td class="text-right">{{ formatDuration(member.durationMs) }}</td>
                      <td class="text-right">{{ member.promptTokens ?? '—' }}</td>
                      <td class="text-right">{{ member.completionTokens ?? '—' }}</td>
                      <td class="text-right">
                        <div>{{ formatCost(member.cost) }}</div>
                        <div v-if="member.costSource" class="text-[10px] text-gray-400">{{ formatCostSource(member.costSource) }}</div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="mt-4">
                <div class="mb-1.5 text-sm font-medium text-gray-700 dark:text-gray-300">最终回答</div>
                <div class="result-content">{{ testResult.content || '（没有返回内容）' }}</div>
              </div>
              <details v-if="testResult.proposals.length" class="mt-4">
                <summary class="cursor-pointer text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">查看候选模型原始回答（{{ testResult.proposals.length }} 份）</summary>
                <div class="mt-2 space-y-2">
                  <div v-for="proposal in testResult.proposals" :key="proposal.model" class="proposal-card">
                    <div class="mb-1 font-mono text-xs text-gray-500">{{ proposal.model }}</div>
                    <div class="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">{{ proposal.content }}</div>
                  </div>
                </div>
              </details>
           </section>
          </div>
        </div>
      </template>
    </TablePageLayout>
    <EnsembleTestDialog
      :show="testDialogOpen"
      :testing="testing"
      :events="testEvents"
      :result="testResult"
      :error="testDialogError"
      @close="closeTestDialog"
      @cancel="cancelTest"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import EnsembleTestDialog from '@/components/admin/EnsembleTestDialog.vue'
import groupsAPI from '@/api/admin/groups'
import channelsAPI, { type Channel } from '@/api/admin/channels'
import ensembleAPI, { type EnsembleMemberStat, type EnsembleProgressEvent, type EnsembleProposer } from '@/api/admin/ensemble'
import type { AdminGroup } from '@/types'
import {
  buildEnsembleGroupPayload,
  deriveEnsembleModelOptions,
  findSharedEnsembleChannel,
  getEnsembleSourceGroups,
  planEnsembleMemberReconciliation,
  validateEnsembleDraft
} from '@/utils/ensemble'

const TARGET_STORAGE_KEY = 'ensemble.native.targetGroupId'
const MAX_PROPOSERS = 6

interface Draft {
  name: string
  description: string
  rateMultiplier: number
}

interface Options {
  minProposers: number
  timeoutSeconds: number
  maxTokens: number
  exposeMetadata: boolean
  streamTrace: boolean
}

interface ResultMember {
  model: string
  role: 'proposer' | 'aggregator' | string
  status: string
  durationMs: number
  promptTokens?: number
  completionTokens?: number
  content?: string
  cost?: number
  costSource?: string
  error?: string
}

interface TestResult {
  members: ResultMember[]
  successCount: number
  durationText: string
  totalTokens: number | string
  content: string
  proposals: Array<{ model: string; content: string }>
  metadataPresent: boolean
  warning?: string
}

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const notice = ref('')
const noticeType = ref<'ok' | 'err' | 'warn'>('ok')
const testApiKey = ref('')
const testResult = ref<TestResult | null>(null)
const testDialogOpen = ref(false)
const testEvents = ref<EnsembleProgressEvent[]>([])
const testDialogError = ref('')
const testAbortController = ref<AbortController | null>(null)

const allGroups = ref<AdminGroup[]>([])
const channels = ref<Channel[]>([])
const selectedEnsembleGroupId = ref<number | null>(null)
const selectedSourceGroupIds = ref<number[]>([])
const models = ref<string[]>([])
const modelPlatforms = ref<Record<string, string>>({})
const proposers = ref<string[]>([])
const aggregator = ref<string | null>(null)
const loadedMembers = ref<EnsembleProposer[]>([])
const draft = ref<Draft>({ name: '', description: '', rateMultiplier: 1 })
const options = ref<Options>({ minProposers: 2, timeoutSeconds: 120, maxTokens: 0, exposeMetadata: false, streamTrace: true })

const sourcePickerOpen = ref(false)
const sourceQuery = ref('')
const sourcePickerInput = ref<HTMLInputElement | null>(null)
const modelPickerOpen = ref(false)
const modelQuery = ref('')
const modelPickerInput = ref<HTMLInputElement | null>(null)

const isNew = computed(() => selectedEnsembleGroupId.value === null)
const ensembleGroups = computed(() => allGroups.value.filter(group => group.platform === 'ensemble' && group.status === 'active'))
const selectedEnsembleGroup = computed(() => ensembleGroups.value.find(group => group.id === selectedEnsembleGroupId.value))
const sourceGroups = computed(() => getEnsembleSourceGroups(allGroups.value))
const selectedSourceGroups = computed(() => selectedSourceGroupIds.value
  .map(id => sourceGroups.value.find(group => group.id === id))
  .filter((group): group is AdminGroup => !!group))
const sourceGroupOptions = computed(() => sourceGroups.value.filter(group => !selectedSourceGroupIds.value.includes(group.id)))
const billingChannel = computed(() => findSharedEnsembleChannel(
  selectedSourceGroupIds.value.length > 0 ? selectedSourceGroupIds.value : (isNew.value ? [] : [selectedEnsembleGroupId.value as number]),
  channels.value
))
const filteredSourceGroupOptions = computed(() => {
  const query = sourceQuery.value.trim().toLowerCase()
  if (!query) return sourceGroupOptions.value
  return sourceGroupOptions.value.filter(group => `${group.name} ${group.platform}`.toLowerCase().includes(query))
})
const ensembleGroupOptions = computed<SelectOption[]>(() => [
  { value: null, label: '新建 Ensemble 分组' },
  ...ensembleGroups.value.map(group => ({ value: group.id, label: `${group.name}（${group.account_count ?? 0} 个账号）` }))
])
const availableModels = computed(() => models.value.filter(model => !proposers.value.includes(model)))
const filteredModelOptions = computed(() => {
  const query = modelQuery.value.trim().toLowerCase()
  return query ? availableModels.value.filter(model => model.toLowerCase().includes(query)) : availableModels.value
})
const aggregatorOptions = computed<SelectOption[]>(() => [
  { value: null, label: '不使用聚合（返回最长候选）' },
  ...models.value.map(model => ({ value: model, label: model }))
])
const overProposerLimit = computed(() => proposers.value.length >= MAX_PROPOSERS)
const canSave = computed(() => {
  const hasSource = !isNew.value || selectedSourceGroupIds.value.length > 0
  return !loading.value && !saving.value && draft.value.name.trim().length > 0 && hasSource &&
    !!billingChannel.value &&
    proposers.value.length >= 2 && options.value.minProposers >= 1 && options.value.minProposers <= proposers.value.length
})
const canSaveAsNew = computed(() => {
  return !loading.value && !saving.value && draft.value.name.trim().length > 0 &&
    selectedSourceGroupIds.value.length > 0 && !!findSharedEnsembleChannel(selectedSourceGroupIds.value, channels.value) &&
    proposers.value.length >= 2 && options.value.minProposers >= 1 && options.value.minProposers <= proposers.value.length
})
const canTest = computed(() => !isNew.value && !!testApiKey.value.trim() && proposers.value.length >= 2)
const noticeClass = computed(() => ({
  ok: 'border-green-200 bg-green-50 text-green-800 dark:border-green-800 dark:bg-green-900/20 dark:text-green-300',
  warn: 'border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-300',
  err: 'border-red-200 bg-red-50 text-red-800 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300'
}[noticeType.value]))

onMounted(init)

async function init() {
  loading.value = true
  notice.value = ''
  testResult.value = null
  try {
    const [groups, channelResult] = await Promise.all([
      groupsAPI.getAll(),
      channelsAPI.list(1, 200, { status: 'active' })
    ])
    allGroups.value = groups
    channels.value = channelResult.items
    const savedTarget = Number(localStorage.getItem(TARGET_STORAGE_KEY))
    const target = Number.isSafeInteger(savedTarget) && ensembleGroups.value.some(group => group.id === savedTarget)
      ? savedTarget
      : null
    await loadTarget(target)
  } catch (error) {
    show(errorMessage(error, '加载 Ensemble 配置失败'), 'err')
  } finally {
    loading.value = false
  }
}

async function loadTarget(groupId: number | null) {
  selectedEnsembleGroupId.value = groupId
  sourcePickerOpen.value = false
  modelPickerOpen.value = false
  sourceQuery.value = ''
  modelQuery.value = ''
  testResult.value = null

  if (groupId === null) {
    loadedMembers.value = []
    selectedSourceGroupIds.value = []
    proposers.value = []
    aggregator.value = null
    draft.value = { name: '', description: '', rateMultiplier: 1 }
    options.value = { minProposers: 2, timeoutSeconds: 120, maxTokens: 0, exposeMetadata: false, streamTrace: true }
    await refreshModels()
    return
  }

  const group = allGroups.value.find(item => item.id === groupId)
  if (!group) {
    await loadTarget(null)
    return
  }

  try {
    const [members, config] = await Promise.all([
      ensembleAPI.listMembers(groupId),
      ensembleAPI.getConfig(groupId)
    ])
    loadedMembers.value = members
    draft.value = { name: group.name, description: group.description ?? '', rateMultiplier: group.rate_multiplier || 1 }
    options.value = {
      minProposers: config.min_proposers || 1,
      timeoutSeconds: config.timeout_seconds || 120,
      maxTokens: config.max_tokens || 0,
      exposeMetadata: config.expose_metadata === true,
      streamTrace: config.stream_trace !== false
    }
    selectedSourceGroupIds.value = [...(config.source_group_ids ?? [])]
    proposers.value = members
      .filter(member => member.role === 'proposer' && member.enabled)
      .sort((a, b) => a.priority - b.priority)
      .map(member => member.model)
    aggregator.value = config.aggregator_enabled
      ? members.find(member => member.role === 'aggregator' && member.enabled)?.model ?? null
      : null
    localStorage.setItem(TARGET_STORAGE_KEY, String(groupId))
    await refreshModels()
  } catch (error) {
    show(errorMessage(error, '加载 Ensemble 分组配置失败'), 'err')
  }
}

async function refreshModels() {
  const pricingGroupIds = selectedSourceGroupIds.value.length > 0
    ? selectedSourceGroupIds.value
    : (isNew.value ? [] : [selectedEnsembleGroupId.value as number])
  const derived = billingChannel.value
    ? deriveEnsembleModelOptions(pricingGroupIds, [billingChannel.value], selectedSourceGroups.value)
    : []
  const memberModels = loadedMembers.value.map(member => member.model.trim()).filter(Boolean)
  const platforms: Record<string, string> = {}
  for (const option of derived) {
    if (!platforms[option.model]) platforms[option.model] = option.platform
  }
  for (const member of loadedMembers.value) {
    if (member.model.trim() && member.platform) platforms[member.model.trim()] = member.platform
  }
  modelPlatforms.value = platforms
  models.value = [...new Set([...derived.map(option => option.model), ...memberModels])].sort((a, b) => a.localeCompare(b))
  proposers.value = proposers.value.filter(model => models.value.includes(model))
  if (aggregator.value && !models.value.includes(aggregator.value)) aggregator.value = null
  if (options.value.minProposers > proposers.value.length && proposers.value.length > 0) {
    options.value.minProposers = proposers.value.length
  }
}

function platformForModel(model: string): string {
  return modelPlatforms.value[model] ?? loadedMembers.value.find(member => member.model === model)?.platform ?? ''
}

async function onEnsembleGroupChange(value: string | number | boolean | null) {
  const groupId = typeof value === 'number' ? value : null
  await loadTarget(groupId)
}

function toggleSourcePicker() {
  sourcePickerOpen.value = !sourcePickerOpen.value
  modelPickerOpen.value = false
  if (sourcePickerOpen.value) nextTick(() => sourcePickerInput.value?.focus())
}

async function addSourceGroup(groupId: number) {
  if (!selectedSourceGroupIds.value.includes(groupId)) selectedSourceGroupIds.value.push(groupId)
  sourcePickerOpen.value = false
  sourceQuery.value = ''
  await refreshModels()
}

async function removeSourceGroup(groupId: number) {
  selectedSourceGroupIds.value = selectedSourceGroupIds.value.filter(id => id !== groupId)
  await refreshModels()
}

function toggleModelPicker() {
  modelPickerOpen.value = !modelPickerOpen.value
  sourcePickerOpen.value = false
  if (modelPickerOpen.value) nextTick(() => modelPickerInput.value?.focus())
}

function addProposer(model: string) {
  if (proposers.value.includes(model) || overProposerLimit.value) return
  proposers.value.push(model)
  if (options.value.minProposers < 1) options.value.minProposers = 1
  modelQuery.value = ''
  modelPickerOpen.value = false
}

function removeProposer(model: string) {
  proposers.value = proposers.value.filter(item => item !== model)
  if (options.value.minProposers > proposers.value.length) options.value.minProposers = Math.max(1, proposers.value.length)
  if (aggregator.value === model && !models.value.includes(model)) aggregator.value = null
}

async function save(forceCreate = false) {
  const creating = forceCreate || isNew.value
  const validation = validateEnsembleDraft({
    name: draft.value.name,
    sourceGroupIds: creating ? selectedSourceGroupIds.value : [selectedEnsembleGroupId.value as number],
    proposers: proposers.value,
    minProposers: options.value.minProposers
  })
  if (validation) {
    show(validationMessage(validation), 'err')
    return
  }
  const effectiveBillingChannel = creating
    ? findSharedEnsembleChannel(selectedSourceGroupIds.value, channels.value)
    : billingChannel.value
  if (!effectiveBillingChannel) {
    show('所选来源组必须属于同一个启用中的渠道，才能保证模型定价和计费规则唯一。', 'err')
    return
  }
  const duplicate = allGroups.value.find(group =>
    group.name.trim().toLowerCase() === draft.value.name.trim().toLowerCase() &&
    (creating || group.id !== selectedEnsembleGroupId.value)
  )
  if (duplicate) {
    show(`分组名称“${draft.value.name.trim()}”已存在，请换一个名称。`, 'err')
    return
  }

  saving.value = true
  let createdGroupId: number | null = null
  let creationConfigured = false
  try {
    let groupId = creating ? null : selectedEnsembleGroupId.value
    if (groupId === null) {
      const channel = effectiveBillingChannel
      if (!channel) throw new Error('来源分组没有共同的启用渠道，无法创建 Ensemble 分组。')
      const created = await groupsAPI.create(buildEnsembleGroupPayload({
        name: draft.value.name,
        description: draft.value.description,
        sourceGroupIds: selectedSourceGroupIds.value,
        rateMultiplier: draft.value.rateMultiplier
      }))
      groupId = created.id
      createdGroupId = groupId
      selectedEnsembleGroupId.value = groupId
      allGroups.value = [...allGroups.value, created]
      await channelsAPI.update(channel.id, {
        group_ids: [...new Set([...(channel.group_ids ?? []), groupId])]
      })
    } else {
      await groupsAPI.update(groupId, {
        name: draft.value.name.trim(),
        description: draft.value.description.trim() || null,
        rate_multiplier: draft.value.rateMultiplier
      })
    }

    await reconcileMembers(groupId)
    await ensembleAPI.updateConfig(groupId, {
      aggregator_enabled: !!aggregator.value,
      min_proposers: options.value.minProposers,
      timeout_seconds: options.value.timeoutSeconds,
      max_tokens: options.value.maxTokens || 0,
      expose_metadata: options.value.exposeMetadata,
      stream_trace: options.value.streamTrace,
      source_group_ids: [...selectedSourceGroupIds.value]
    })
    creationConfigured = true
    localStorage.setItem(TARGET_STORAGE_KEY, String(groupId))
    await reloadGroupsAndTarget(groupId)
    show(`Ensemble 分组“${draft.value.name.trim()}”已保存，共 ${proposers.value.length} 个候选${aggregator.value ? `，聚合模型为 ${aggregator.value}` : '，不使用聚合模型'}`, 'ok')
  } catch (error) {
    if (createdGroupId !== null && !creationConfigured) {
      try {
        await groupsAPI.delete(createdGroupId)
      } catch {
        show(`保存失败，且自动清理新建分组失败，请检查分组“${draft.value.name.trim()}”。`, 'err')
        return
      }
    }
    show(`保存失败：${errorMessage(error, '服务器返回未知错误')}`, 'err')
  } finally {
    saving.value = false
  }
}

async function reconcileMembers(groupId: number) {
  const desired: Array<{ role: 'proposer' | 'aggregator'; model: string; platform: string; priority: number }> = proposers.value.map((model, index) => ({
    role: 'proposer', model, platform: platformForModel(model), priority: 100 + index
  }))
  if (aggregator.value) desired.push({ role: 'aggregator', model: aggregator.value, platform: platformForModel(aggregator.value), priority: 10 })

  const plan = planEnsembleMemberReconciliation(loadedMembers.value, desired)
  for (const update of plan.updates) {
    await ensembleAPI.updateMember(groupId, update.id, { ...update.member, enabled: true })
  }
  for (const create of plan.creates) {
    await ensembleAPI.createMember(groupId, { ...create, enabled: true })
  }
  for (const memberId of plan.deletes) {
    await ensembleAPI.deleteMember(groupId, memberId)
  }
}

async function reloadGroupsAndTarget(groupId: number) {
  allGroups.value = await groupsAPI.getAll()
  channels.value = (await channelsAPI.list(1, 200, { status: 'active' })).items
  await loadTarget(groupId)
}

async function runTest() {
  if (isNew.value) {
    show('请先保存配置，创建 Ensemble 分组后再测试。', 'warn')
    return
  }
  if (!testApiKey.value.trim()) {
    show('请输入当前 Ensemble 分组的测试 API Key。', 'warn')
    return
  }
  testing.value = true
  testResult.value = null
  testDialogOpen.value = true
  testEvents.value = []
  testDialogError.value = ''
  testAbortController.value?.abort()
  testAbortController.value = new AbortController()
  try {
    const data = await ensembleAPI.testStream(testApiKey.value.trim(), [
      { role: 'user', content: '请用两三句话说明什么是多模型聚合。' }
    ], event => {
      testEvents.value.push(event)
      if (event.type === 'error' || event.type === 'fallback') {
        testDialogError.value = event.error ?? (event.type === 'fallback' ? '聚合模型失败，已回退到候选回答。' : '')
      }
      if (event.type === 'completed' && event.response) {
        testResult.value = normalizeTestResult(event.response)
      }
    }, testAbortController.value.signal)
    if (data && !testResult.value) testResult.value = normalizeTestResult(data)
    if (testDialogError.value) {
      show(`测试结束：${testDialogError.value}`, 'warn')
    } else if (testResult.value) {
      show(`测试完成：${testResult.value.successCount}/${testResult.value.members.length} 个调用成功，耗时 ${testResult.value.durationText}`, testResult.value.successCount > 0 ? 'ok' : 'warn')
    }
  } catch (error) {
    if ((error as { name?: string })?.name === 'AbortError') {
      testDialogError.value = '测试已取消。'
      show('测试已取消。', 'warn')
    } else {
      testDialogError.value = errorMessage(error, '网关请求失败')
      show(`测试失败：${testDialogError.value}`, 'err')
    }
  } finally {
    testing.value = false
    testAbortController.value = null
  }
}

function cancelTest() {
  testAbortController.value?.abort()
}

function closeTestDialog() {
  if (testing.value) {
    cancelTest()
  }
  testDialogOpen.value = false
}

onUnmounted(() => testAbortController.value?.abort())

function normalizeTestResult(data: any): TestResult {
  const metadata = data?.ensemble_metadata
  const metadataPresent = !!metadata && Array.isArray(metadata.members)
  const members: EnsembleMemberStat[] = metadataPresent ? metadata.members : [
    ...(metadata?.proposer_results ?? []).map((member: EnsembleMemberStat) => ({ ...member, role: 'proposer' })),
    ...(metadata?.aggregator_result ? [{ ...metadata.aggregator_result, role: 'aggregator' }] : [])
  ]
  const normalized = members.map(member => ({
    model: member.model,
    role: member.role,
    status: member.status ?? 'failed',
    durationMs: member.duration_ms ?? 0,
    promptTokens: member.prompt_tokens,
    completionTokens: member.completion_tokens,
    content: member.content,
    cost: member.cost,
    costSource: member.cost_source,
    error: member.error
  }))
  const proposals = normalized
    .filter(member => member.role === 'proposer' && member.status === 'ok')
    .map(member => ({ model: member.model, content: member.content ?? '' }))
  const totalTokens = data?.usage?.total_tokens ?? '—'
  return {
    members: normalized,
    successCount: normalized.filter(member => member.status === 'ok').length,
    durationText: formatDuration(metadata?.duration_ms ?? 0),
    totalTokens,
    content: data?.choices?.[0]?.message?.content ?? '',
    proposals,
    metadataPresent,
    warning: metadataPresent ? undefined : '网关没有返回 Ensemble 执行元数据，可能没有进入 Ensemble 路由；请检查测试 Key 是否绑定到 Ensemble 分组。'
  }
}

function formatDuration(value: number): string {
  return value > 0 ? `${(value / 1000).toFixed(1)}s` : '—'
}

function formatCost(value?: number): string {
  return typeof value === 'number' ? `$${value.toFixed(6)}` : '未返回'
}

function formatCostSource(source?: string): string {
  if (source === 'upstream') return '上游返回'
  if (source === 'channel') return '渠道估算'
  if (source === 'litellm') return '全局估算'
  if (source === 'fallback') return '默认估算'
  return source ?? ''
}

function validationMessage(key: string): string {
  return {
    'name-required': '请先填写 Ensemble 分组名称。',
    'source-group-required': '请至少添加一个账号来源分组。',
    'at-least-two-proposers': '请至少添加 2 个候选模型。',
    'invalid-min-proposers': '最少成功候选数必须在 1 和已添加候选数之间。'
  }[key] ?? '请检查 Ensemble 配置。'
}

function errorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'string') return error
  if (error && typeof error === 'object') {
    const value = error as { message?: unknown; response?: { data?: { message?: unknown } } }
    if (typeof value.message === 'string') return value.message
    if (typeof value.response?.data?.message === 'string') return value.response.data.message
  }
  return fallback
}

function show(message: string, type: 'ok' | 'err' | 'warn') {
  notice.value = message
  noticeType.value = type
  if (type === 'ok') window.setTimeout(() => { notice.value = '' }, 8000)
}
</script>

<style scoped>
.intro-panel { @apply flex flex-wrap items-center justify-between gap-5 rounded-xl border border-primary-100 bg-primary-50/70 p-5 dark:border-primary-900/50 dark:bg-primary-900/10; }
.intro-kicker { @apply text-sm font-semibold text-primary-700 dark:text-primary-300; }
.intro-rule { @apply rounded-full bg-white/80 px-2.5 py-1 text-xs text-primary-600 dark:bg-dark-800/70 dark:text-primary-300; }
.intro-flow { @apply flex items-center gap-2 text-xs font-medium text-primary-700 dark:text-primary-300; }
.readiness-bar { @apply grid gap-px overflow-hidden rounded-xl border border-gray-200 bg-gray-200 sm:grid-cols-3 dark:border-dark-600 dark:bg-dark-600; }
.readiness-item { @apply flex min-h-[78px] flex-col justify-center bg-white px-4 py-3 dark:bg-dark-800; }
.readiness-label { @apply text-[11px] font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500; }
.readiness-item strong { @apply mt-1 text-sm font-semibold text-gray-900 dark:text-white; }
.readiness-meta { @apply mt-0.5 text-xs text-gray-500 dark:text-gray-400; }
.section-card { @apply rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800; }
.section-header { @apply mb-4 flex flex-wrap items-start justify-between gap-3; }
.section-title { @apply flex items-center text-base font-semibold text-gray-900 dark:text-white; }
.section-desc { @apply mt-1 max-w-3xl text-xs leading-relaxed text-gray-500 dark:text-gray-400; }
.step-no { @apply mr-2 inline-flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary-600 text-xs font-bold text-white; }
.state-badge, .badge-count { @apply rounded-full px-3 py-1 text-xs font-medium; }
.state-new { @apply bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300; }
.state-existing { @apply bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-300; }
.badge-count { @apply bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300; }
.badge-danger { @apply bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300; }
.subsection-title { @apply text-sm font-semibold text-gray-800 dark:text-gray-200; }
.required { @apply text-red-500; }
.field-label { @apply mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300; }
.field-input { @apply w-full rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-white; }
.field-hint { @apply mt-1 text-xs leading-5 text-gray-400 dark:text-gray-500; }
.source-panel { @apply rounded-lg border border-dashed border-gray-300 bg-gray-50/70 p-4 dark:border-dark-500 dark:bg-dark-900/30; }
.source-chip { @apply inline-flex max-w-full items-center gap-1.5 rounded-full border border-primary-100 bg-white px-2.5 py-1.5 text-xs text-gray-700 shadow-sm dark:border-primary-900/50 dark:bg-dark-800 dark:text-gray-200; }
.chip-platform { @apply rounded bg-gray-100 px-1.5 py-0.5 text-[10px] uppercase text-gray-500 dark:bg-dark-700 dark:text-gray-400; }
.chip-remove { @apply ml-0.5 rounded p-0.5 text-gray-400 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20; }
.empty-inline { @apply rounded-lg border border-dashed border-gray-200 px-4 py-4 text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400; }
.existing-note { @apply flex items-start gap-2 rounded-lg border border-blue-100 bg-blue-50/70 p-3 text-xs leading-5 text-blue-700 dark:border-blue-900/50 dark:bg-blue-900/10 dark:text-blue-300; }
.add-button { @apply inline-flex items-center gap-1.5 rounded-xl border border-dashed border-primary-300 px-3 py-2 text-sm font-medium text-primary-700 transition hover:border-primary-500 hover:bg-primary-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-primary-700 dark:text-primary-300 dark:hover:bg-primary-900/20; }
.picker-panel { @apply absolute left-0 top-full z-20 mt-2 w-full max-w-lg overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xl dark:border-dark-600 dark:bg-dark-800; }
.picker-search { @apply flex items-center gap-2 border-b border-gray-100 px-3 py-2 dark:border-dark-700; }
.picker-input { @apply min-w-0 flex-1 bg-transparent text-sm text-gray-900 outline-none placeholder:text-gray-400 dark:text-white; }
.picker-option { @apply flex w-full items-center justify-between gap-3 px-4 py-3 text-left text-sm text-gray-700 hover:bg-primary-50 dark:text-gray-200 dark:hover:bg-primary-900/20; }
.picker-empty { @apply px-4 py-6 text-center text-sm text-gray-500 dark:text-gray-400; }
.member-row { @apply flex items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2.5 dark:border-dark-600 dark:bg-dark-900/30; }
.member-index { @apply flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-semibold text-primary-700 dark:bg-primary-900/30 dark:text-primary-300; }
.member-role { @apply text-[11px] text-gray-400; }
.icon-button { @apply rounded-lg p-1.5 text-gray-400 transition hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20; }
.estimate-bar { @apply flex flex-wrap items-center gap-2 rounded-lg bg-gray-50 px-4 py-3 text-sm text-gray-600 dark:bg-dark-900/40 dark:text-gray-300; }
.estimate-bar strong { @apply text-primary-700 dark:text-primary-300; }
.option-toggle { @apply flex items-center justify-between gap-4 rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 dark:border-dark-600 dark:bg-dark-900/30; }
.test-key-row { @apply flex items-center gap-3; }
.test-key-row .field-input { @apply min-w-0 flex-1; }
.test-run-button { @apply min-h-[44px] flex-shrink-0 whitespace-nowrap; }
.notice { @apply flex items-start gap-2.5 rounded-lg border p-4 text-sm; }
.result-table { @apply min-w-full text-sm; }
.result-table th { @apply bg-gray-50 px-3 py-2.5 text-left text-xs font-semibold text-gray-500 dark:bg-dark-900/50 dark:text-gray-400; }
.result-table td { @apply border-t border-gray-100 px-3 py-3 text-gray-700 dark:border-dark-700 dark:text-gray-300; }
.pill-ok, .pill-fail { @apply inline-flex rounded-full px-2 py-0.5 text-xs font-medium; }
.pill-ok { @apply bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300; }
.pill-fail { @apply bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300; }
.result-content { @apply max-h-72 overflow-y-auto whitespace-pre-wrap rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm leading-6 text-gray-800 dark:border-dark-600 dark:bg-dark-900/40 dark:text-gray-200; }
.proposal-card { @apply rounded-lg border border-gray-200 p-3 dark:border-dark-600; }
@media (max-width: 639px) {
  .test-key-row { @apply items-stretch flex-col; }
  .test-run-button { @apply w-full justify-center; }
}
</style>
