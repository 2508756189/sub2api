<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="init">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button v-if="!isNew" type="button" class="btn btn-secondary" :disabled="saving || !canSaveAsNew" @click="save(true)">
            <Icon name="plus" size="md" class="mr-1.5" />
            {{ t('admin.ensemble.actions.saveAsNew') }}
          </button>
          <button type="button" class="btn btn-primary" :disabled="saving || !canSave" @click="save()">
            <Icon name="check" size="md" class="mr-1.5" />
            {{ saving ? t('admin.ensemble.actions.saving') : t('admin.ensemble.actions.save') }}
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
                  <span class="intro-kicker">{{ t('admin.ensemble.intro.kicker') }}</span>
                  <span class="intro-rule">{{ t('admin.ensemble.intro.rule') }}</span>
                </div>
                <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                  {{ t('admin.ensemble.intro.body') }}
                </p>
              </div>
              <div class="intro-flow">
                <span>{{ t('admin.ensemble.intro.flowSource') }}</span><Icon name="arrowRight" size="sm" /><span>{{ t('admin.ensemble.intro.flowProposers') }}</span><Icon name="arrowRight" size="sm" /><span>{{ t('admin.ensemble.intro.flowAggregate') }}</span>
              </div>
            </div>

            <div class="readiness-bar">
              <div class="readiness-item">
                <span class="readiness-label">{{ t('admin.ensemble.readiness.entry') }}</span>
                <strong>{{ isNew ? t('admin.ensemble.readiness.entryUnsaved') : 'ensemble' }}</strong>
                <span class="readiness-meta">{{ isNew ? t('admin.ensemble.readiness.entryUnsavedHint') : t('admin.ensemble.readiness.entryAccounts', { count: selectedEnsembleGroup?.account_count ?? 0 }) }}</span>
              </div>
              <div class="readiness-item">
                <span class="readiness-label">{{ t('admin.ensemble.readiness.dispatch') }}</span>
                <strong>{{ proposers.length }} / {{ MAX_PROPOSERS }}</strong>
                <span class="readiness-meta">{{ t('admin.ensemble.readiness.dispatchHint', { count: options.minProposers }) }}</span>
              </div>
              <div class="readiness-item">
                <span class="readiness-label">{{ t('admin.ensemble.readiness.testMethod') }}</span>
                <strong>{{ t('admin.ensemble.readiness.testMethodValue') }}</strong>
                <span class="readiness-meta">{{ t('admin.ensemble.readiness.testMethodHint') }}</span>
              </div>
            </div>

            <div class="grid gap-5 xl:grid-cols-[minmax(0,1.15fr)_minmax(360px,0.85fr)]">
              <section class="section-card">
                <div class="section-header">
                  <div>
                    <h2 class="section-title"><span class="step-no">1</span>{{ t('admin.ensemble.group.title') }}</h2>
                    <p class="section-desc">{{ t('admin.ensemble.group.description') }}</p>
                  </div>
                  <span :class="['state-badge', isNew ? 'state-new' : 'state-existing']">
                    {{ isNew ? t('admin.ensemble.group.stateNew') : t('admin.ensemble.group.stateExisting') }}
                  </span>
                </div>

                <div class="grid gap-4 md:grid-cols-2">
                  <div class="md:col-span-2">
                    <label class="field-label">{{ t('admin.ensemble.group.selectLabel') }}</label>
                    <Select
                      v-model="selectedEnsembleGroupId"
                      :options="ensembleGroupOptions"
                      :placeholder="t('admin.ensemble.group.newOption')"
                      searchable="auto"
                      :aria-label="t('admin.ensemble.group.selectAria')"
                      @change="onEnsembleGroupChange"
                    />
                    <p class="field-hint">{{ t('admin.ensemble.group.selectHint') }}</p>
                  </div>

                  <div>
                    <label class="field-label">{{ t('admin.ensemble.group.nameLabel') }} <span class="required">*</span></label>
                    <input v-model="draft.name" class="field-input" :placeholder="t('admin.ensemble.group.namePlaceholder')" maxlength="80" />
                  </div>
                  <div>
                    <label class="field-label">{{ t('admin.ensemble.group.rateLabel') }}</label>
                    <input v-model.number="draft.rateMultiplier" class="field-input" type="number" min="0.01" step="0.01" />
                    <p class="field-hint">{{ t('admin.ensemble.group.rateHint') }}</p>
                  </div>
                  <div class="md:col-span-2">
                    <label class="field-label">{{ t('admin.ensemble.group.descriptionLabel') }}</label>
                    <textarea v-model="draft.description" class="field-input min-h-20 resize-y" :placeholder="t('admin.ensemble.group.descriptionPlaceholder')" maxlength="240" />
                  </div>
                </div>

                <div v-if="isNew" class="source-panel mt-5">
                  <div class="mb-3 flex items-start justify-between gap-3">
                    <div>
                      <h3 class="subsection-title">{{ t('admin.ensemble.source.title') }} <span class="required">*</span></h3>
                      <p class="section-desc">{{ t('admin.ensemble.source.description') }}</p>
                    </div>
                    <span class="badge-count">{{ t('admin.ensemble.source.selectedCount', { count: selectedSourceGroupIds.length }) }}</span>
                  </div>

                  <div v-if="selectedSourceGroups.length" class="flex flex-wrap gap-2">
                    <span v-for="group in selectedSourceGroups" :key="group.id" class="source-chip">
                      <Icon name="server" size="xs" class="text-primary-500" />
                      <span class="truncate">{{ group.name }}</span>
                      <span class="chip-platform">{{ group.platform }}</span>
                      <button type="button" class="chip-remove" :aria-label="t('admin.ensemble.source.removeAria', { name: group.name })" @click="removeSourceGroup(group.id)">
                        <Icon name="x" size="xs" />
                      </button>
                    </span>
                  </div>
                  <div v-else class="empty-inline">{{ t('admin.ensemble.source.empty') }}</div>

                  <div class="relative mt-3">
                    <button type="button" class="add-button" :disabled="sourceGroupOptions.length === 0" @click="toggleSourcePicker">
                      <Icon name="plus" size="sm" />
                      {{ sourceGroupOptions.length ? t('admin.ensemble.source.add') : t('admin.ensemble.source.addEmpty') }}
                    </button>
                    <div v-if="sourcePickerOpen" class="picker-panel">
                      <div class="picker-search">
                        <Icon name="search" size="sm" class="text-gray-400" />
                        <input ref="sourcePickerInput" v-model="sourceQuery" class="picker-input" :placeholder="t('admin.ensemble.source.searchPlaceholder')" @keydown.esc="sourcePickerOpen = false" />
                      </div>
                      <button v-for="group in filteredSourceGroupOptions" :key="group.id" type="button" class="picker-option" @click="addSourceGroup(group.id)">
                        <span class="min-w-0 truncate">{{ group.name }}</span>
                        <span class="chip-platform">{{ group.platform }}</span>
                      </button>
                      <div v-if="filteredSourceGroupOptions.length === 0" class="picker-empty">{{ t('admin.ensemble.source.noMatch') }}</div>
                    </div>
                  </div>
                  <div v-if="billingChannel" class="mt-3 rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-xs text-green-700 dark:border-green-900/50 dark:bg-green-900/10 dark:text-green-300">
                    <i18n-t keypath="admin.ensemble.source.billingChannel" tag="span" scope="global">
                      <template #name>
                        <strong>{{ billingChannel.name }}</strong>
                      </template>
                    </i18n-t>
                  </div>
                  <div v-else-if="selectedSourceGroupIds.length" class="mt-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs leading-5 text-amber-700 dark:border-amber-900/50 dark:bg-amber-900/10 dark:text-amber-300">
                    {{ t('admin.ensemble.source.billingChannelConflict') }}
                  </div>
                </div>

                <div v-else class="existing-note mt-5">
                  <Icon name="infoCircle" size="md" class="flex-shrink-0 text-primary-500" />
                  <div>
                    <p>{{ t('admin.ensemble.source.existingNote') }}</p>
                    <div v-if="selectedSourceGroups.length" class="mt-2 flex flex-wrap gap-1.5">
                      <span v-for="group in selectedSourceGroups" :key="group.id" class="source-chip">
                        <span class="truncate">{{ t('admin.ensemble.source.fromLabel', { name: group.name }) }}</span>
                        <span class="chip-platform">{{ group.platform }}</span>
                      </span>
                    </div>
                    <p v-if="billingChannel" class="mt-2">{{ t('admin.ensemble.source.billingChannelPlain', { name: billingChannel.name }) }}</p>
                    <p v-else class="mt-2 text-amber-700 dark:text-amber-300">{{ t('admin.ensemble.source.billingChannelMissing') }}</p>
                  </div>
                </div>
              </section>

              <section class="section-card">
                <div class="section-header">
                  <div>
                    <h2 class="section-title"><span class="step-no">2</span>{{ t('admin.ensemble.proposers.title') }}</h2>
                    <p class="section-desc">{{ t('admin.ensemble.proposers.description') }}</p>
                  </div>
                  <span :class="['badge-count', proposers.length >= 2 ? '' : 'badge-danger']">{{ proposers.length }} / {{ MAX_PROPOSERS }}</span>
                </div>

                <div v-if="proposers.length" class="space-y-2">
                  <div v-for="(model, index) in proposers" :key="model" class="member-row">
                    <span class="member-index">{{ index + 1 }}</span>
                    <span class="min-w-0 flex-1 truncate font-mono text-sm">{{ model }}</span>
                    <span class="member-role">{{ t('admin.ensemble.proposers.roleLabel') }}</span>
                    <button type="button" class="icon-button" :aria-label="t('admin.ensemble.proposers.removeAria', { model })" :title="t('admin.ensemble.proposers.remove')" @click="removeProposer(model)">
                      <Icon name="x" size="sm" />
                    </button>
                  </div>
                </div>
                <div v-else class="empty-inline">{{ t('admin.ensemble.proposers.empty') }}</div>

                <div class="relative mt-3">
                  <button type="button" class="add-button" :disabled="overProposerLimit || availableModels.length === 0" @click="toggleModelPicker">
                    <Icon name="plus" size="sm" />
                    {{ availableModels.length === 0 ? t('admin.ensemble.proposers.noModels') : overProposerLimit ? t('admin.ensemble.proposers.limitReached') : t('admin.ensemble.proposers.add') }}
                  </button>
                  <div v-if="modelPickerOpen" class="picker-panel">
                    <div class="picker-search">
                      <Icon name="search" size="sm" class="text-gray-400" />
                      <input ref="modelPickerInput" v-model="modelQuery" class="picker-input" :placeholder="t('admin.ensemble.proposers.searchPlaceholder')" @keydown.esc="modelPickerOpen = false" />
                    </div>
                    <button v-for="model in filteredModelOptions" :key="model" type="button" class="picker-option" @click="addProposer(model)">
                      <span class="truncate font-mono text-xs">{{ model }}</span>
                      <Icon name="plus" size="sm" class="text-primary-500" />
                    </button>
                    <div v-if="filteredModelOptions.length === 0" class="picker-empty">{{ t('admin.ensemble.proposers.noMatch') }}</div>
                  </div>
                </div>
                <p class="field-hint mt-2">{{ t('admin.ensemble.proposers.hint', { max: MAX_PROPOSERS }) }}</p>
              </section>
            </div>

            <section class="section-card">
              <div class="section-header">
                <div>
                  <h2 class="section-title"><span class="step-no">3</span>{{ t('admin.ensemble.aggregator.title') }}</h2>
                  <p class="section-desc">{{ t('admin.ensemble.aggregator.description') }}</p>
                </div>
              </div>
              <div class="max-w-xl">
                <Select
                  v-model="aggregator"
                  :options="aggregatorOptions"
                  :placeholder="t('admin.ensemble.aggregator.none')"
                  searchable="auto"
                  clearable
                  :aria-label="t('admin.ensemble.aggregator.selectAria')"
                />
                <p class="field-hint">{{ t('admin.ensemble.aggregator.hint') }}</p>
              </div>
            </section>

            <section class="section-card">
              <div class="section-header">
                <div>
                  <h2 class="section-title"><span class="step-no">4</span>{{ t('admin.ensemble.options.title') }}</h2>
                  <p class="section-desc">{{ t('admin.ensemble.options.description') }}</p>
                </div>
              </div>
              <div class="grid gap-4 sm:grid-cols-3">
                <div>
                  <label class="field-label">{{ t('admin.ensemble.options.minProposers') }}</label>
                  <input v-model.number="options.minProposers" class="field-input" type="number" min="1" :max="Math.max(1, proposers.length)" />
                  <p class="field-hint">{{ t('admin.ensemble.options.minProposersHint') }}</p>
                </div>
                <div>
                  <label class="field-label">{{ t('admin.ensemble.options.timeout') }}</label>
                  <input v-model.number="options.timeoutSeconds" class="field-input" type="number" min="1" max="600" />
                </div>
                <div>
                  <label class="field-label">{{ t('admin.ensemble.options.maxTokens') }}</label>
                  <input v-model.number="options.maxTokens" class="field-input" type="number" min="0" :placeholder="t('admin.ensemble.options.maxTokensPlaceholder')" />
                </div>
              </div>
              <div class="option-toggle mt-4">
                <div>
                  <div class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ t('admin.ensemble.options.exposeMetadata') }}</div>
                  <p class="field-hint">{{ t('admin.ensemble.options.exposeMetadataHint') }}</p>
                </div>
                <Toggle v-model="options.exposeMetadata" :aria-label="t('admin.ensemble.options.exposeMetadata')" />
              </div>
              <div class="option-toggle mt-4">
                <div>
                  <div class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ t('admin.ensemble.options.streamTrace') }}</div>
                  <p class="field-hint">{{ t('admin.ensemble.options.streamTraceHint') }}</p>
                </div>
                <Toggle v-model="options.streamTrace" :aria-label="t('admin.ensemble.options.streamTraceAria')" />
              </div>
              <div class="mt-4">
                <div class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ t('admin.ensemble.options.aggregatorOverrides') }}</div>
                <p class="field-hint">{{ t('admin.ensemble.options.aggregatorOverridesHint') }}</p>
                <div v-if="options.aggregatorOverrides.length" class="mt-3 space-y-2">
                  <div v-for="row in options.aggregatorOverrides" :key="row.uid" class="member-row">
                    <input
                      v-model="row.path"
                      class="field-input min-w-0 flex-1 font-mono"
                      type="text"
                      :placeholder="t('admin.ensemble.options.aggregatorOverridePathPlaceholder')"
                      :aria-label="t('admin.ensemble.options.aggregatorOverridePath')"
                    />
                    <input
                      v-model="row.value"
                      class="field-input min-w-0 flex-1 font-mono"
                      type="text"
                      :placeholder="t('admin.ensemble.options.aggregatorOverrideValuePlaceholder')"
                      :aria-label="t('admin.ensemble.options.aggregatorOverrideValue')"
                    />
                    <button
                      type="button"
                      class="icon-button"
                      :aria-label="t('admin.ensemble.options.aggregatorOverrideRemoveAria', { path: row.path || t('admin.ensemble.options.aggregatorOverridePath') })"
                      :title="t('admin.ensemble.options.aggregatorOverrideRemove')"
                      @click="removeAggregatorOverride(row.uid)"
                    >
                      <Icon name="x" size="sm" />
                    </button>
                  </div>
                </div>
                <div v-else class="empty-inline mt-3">{{ t('admin.ensemble.options.aggregatorOverridesEmpty') }}</div>
                <button
                  type="button"
                  class="add-button mt-3"
                  :disabled="options.aggregatorOverrides.length >= MAX_AGGREGATOR_OVERRIDES"
                  @click="addAggregatorOverride"
                >
                  <Icon name="plus" size="sm" />
                  {{ options.aggregatorOverrides.length >= MAX_AGGREGATOR_OVERRIDES
                    ? t('admin.ensemble.options.aggregatorOverrideLimitReached', { count: MAX_AGGREGATOR_OVERRIDES })
                    : t('admin.ensemble.options.aggregatorOverrideAdd') }}
                </button>
              </div>
              <div class="estimate-bar mt-4">
                <span>{{ t('admin.ensemble.options.estimate') }}</span>
                <strong>{{ t('admin.ensemble.options.estimateCalls', { count: proposers.length + (aggregator ? 1 : 0) }) }}</strong>
                <span class="text-gray-500 dark:text-gray-400">{{ aggregator ? t('admin.ensemble.options.estimateWithAggregator') : t('admin.ensemble.options.estimateWithoutAggregator') }}</span>
              </div>
            </section>

            <section class="section-card">
              <div class="section-header">
                <div>
                  <h2 class="section-title"><Icon name="play" size="md" class="mr-2 text-primary-500" />{{ t('admin.ensemble.test.title') }}</h2>
                  <p class="section-desc">{{ t('admin.ensemble.test.description') }}</p>
                </div>
              </div>
              <div class="test-key-field">
                <label class="field-label">{{ t('admin.ensemble.test.keyLabel') }}</label>
                <div class="test-key-row">
                  <input v-model="testApiKey" class="field-input font-mono" type="password" autocomplete="off" :placeholder="t('admin.ensemble.test.keyPlaceholder')" />
                  <button type="button" class="btn btn-secondary test-run-button" :disabled="testing || !canTest" @click="runTest">
                    <Icon name="play" size="md" class="mr-1.5" />{{ testing ? t('admin.ensemble.test.running') : t('admin.ensemble.test.run') }}
                  </button>
                </div>
                <i18n-t keypath="admin.ensemble.test.keyHint" tag="p" class="field-hint" scope="global">
                  <template #model>
                    <code>ensemble</code>
                  </template>
                </i18n-t>
              </div>
            </section>

            <section v-if="testResult" class="section-card">
              <div class="section-header">
                <div>
                  <h2 class="section-title"><Icon name="chartBar" size="md" class="mr-2 text-primary-500" />{{ t('admin.ensemble.result.title') }}</h2>
                  <p class="section-desc">{{ t('admin.ensemble.result.description') }}</p>
                </div>
                <div class="flex flex-wrap gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <span>{{ testResult.metadataPresent ? t('admin.ensemble.result.successCount', { succeeded: testResult.successCount, total: testResult.members.length }) : t('admin.ensemble.result.noMetadata') }}</span>
                  <span>{{ t('admin.ensemble.stats.duration', { value: testResult.durationText }) }}</span>
                  <span>{{ t('admin.ensemble.stats.totalTokens', { value: testResult.totalTokens }) }}</span>
                </div>
              </div>
              <div v-if="testResult.warning" class="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm leading-6 text-amber-700 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-300">
                {{ testResult.warning }}
              </div>
              <div class="overflow-x-auto">
                <table class="result-table">
                  <thead>
                    <tr><th>{{ t('admin.ensemble.result.columns.role') }}</th><th>{{ t('admin.ensemble.result.columns.model') }}</th><th>{{ t('admin.ensemble.result.columns.status') }}</th><th class="text-right">{{ t('admin.ensemble.result.columns.duration') }}</th><th class="text-right">{{ t('admin.ensemble.result.columns.promptTokens') }}</th><th class="text-right">{{ t('admin.ensemble.result.columns.completionTokens') }}</th><th class="text-right">{{ t('admin.ensemble.result.columns.cost') }}</th></tr>
                  </thead>
                  <tbody>
                    <tr v-for="(member, index) in testResult.members" :key="`${member.role}-${member.model}-${index}`">
                      <td>{{ member.role === 'aggregator' ? t('admin.ensemble.result.roleAggregator') : t('admin.ensemble.result.roleProposer', { index: index + 1 }) }}</td>
                      <td class="font-mono text-xs">{{ member.model }}</td>
                      <td><span :class="member.status === 'ok' ? 'pill-ok' : 'pill-fail'">{{ member.status === 'ok' ? t('admin.ensemble.stats.statusOk') : t('admin.ensemble.stats.statusFailed') }}</span><span v-if="member.error" class="ml-2 text-xs text-red-500">{{ member.error }}</span></td>
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
                <div class="mb-1.5 text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ensemble.result.finalAnswer') }}</div>
                <div class="result-content">{{ testResult.content || t('admin.ensemble.result.emptyContent') }}</div>
              </div>
              <details v-if="testResult.proposals.length" class="mt-4">
                <summary class="cursor-pointer text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">{{ t('admin.ensemble.result.viewProposals', { count: testResult.proposals.length }) }}</summary>
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
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()

const TARGET_STORAGE_KEY = 'ensemble.native.targetGroupId'
const MAX_PROPOSERS = 6

interface Draft {
  name: string
  description: string
  rateMultiplier: number
}

interface AggregatorOverrideRow {
  uid: number
  path: string
  value: string
}

interface Options {
  minProposers: number
  timeoutSeconds: number
  maxTokens: number
  exposeMetadata: boolean
  streamTrace: boolean
  aggregatorOverrides: AggregatorOverrideRow[]
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
const options = ref<Options>({ minProposers: 2, timeoutSeconds: 120, maxTokens: 0, exposeMetadata: false, streamTrace: true, aggregatorOverrides: [] })

// Matches MaxEnsembleAggregatorBodyOverrides in the backend. Capping here only
// saves a round-trip; the backend is what actually refuses an over-long set.
const MAX_AGGREGATOR_OVERRIDES = 16
let aggregatorOverrideUid = 0

// Values are edited as text but stored as JSON, so `4096` saves as a number and
// `max` saves as a string. A string that would itself parse as JSON is shown
// quoted, otherwise "true" would come back as a boolean on the next load.
function formatAggregatorOverrideValue(value: unknown): string {
  if (typeof value !== 'string') return JSON.stringify(value) ?? ''
  try {
    if (typeof JSON.parse(value) !== 'string') return JSON.stringify(value)
  } catch {
    return value
  }
  return value
}

function parseAggregatorOverrideValue(raw: string): unknown {
  const trimmed = raw.trim()
  if (trimmed === '') return ''
  try {
    return JSON.parse(trimmed)
  } catch {
    return trimmed
  }
}

// Sorted by path so the editor shows a stable order: the backend stores the set
// as a JSON object, and Go marshals object keys sorted anyway.
function aggregatorOverridesToRows(overrides: Record<string, unknown> | undefined): AggregatorOverrideRow[] {
  if (!overrides) return []
  return Object.keys(overrides)
    .sort((a, b) => a.localeCompare(b))
    .map(path => ({ uid: aggregatorOverrideUid++, path, value: formatAggregatorOverrideValue(overrides[path]) }))
}

// Returns undefined rather than {} when nothing is configured, so the field is
// omitted from the payload and the stored config keeps its previous shape.
function rowsToAggregatorOverrides(rows: AggregatorOverrideRow[]): Record<string, unknown> | undefined {
  const overrides: Record<string, unknown> = {}
  for (const row of rows) {
    const path = row.path.trim()
    if (path === '') continue
    overrides[path] = parseAggregatorOverrideValue(row.value)
  }
  return Object.keys(overrides).length > 0 ? overrides : undefined
}

function addAggregatorOverride() {
  if (options.value.aggregatorOverrides.length >= MAX_AGGREGATOR_OVERRIDES) return
  options.value.aggregatorOverrides.push({ uid: aggregatorOverrideUid++, path: '', value: '' })
}

function removeAggregatorOverride(uid: number) {
  options.value.aggregatorOverrides = options.value.aggregatorOverrides.filter(row => row.uid !== uid)
}

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
  { value: null, label: t('admin.ensemble.group.newOption') },
  ...ensembleGroups.value.map(group => ({
    value: group.id,
    label: t('admin.ensemble.group.optionLabel', { name: group.name, count: group.account_count ?? 0 })
  }))
])
const availableModels = computed(() => models.value.filter(model => !proposers.value.includes(model)))
const filteredModelOptions = computed(() => {
  const query = modelQuery.value.trim().toLowerCase()
  return query ? availableModels.value.filter(model => model.toLowerCase().includes(query)) : availableModels.value
})
const aggregatorOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.ensemble.aggregator.none') },
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
    show(errorMessage(error, t('admin.ensemble.errors.loadConfig')), 'err')
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
    options.value = { minProposers: 2, timeoutSeconds: 120, maxTokens: 0, exposeMetadata: false, streamTrace: true, aggregatorOverrides: [] }
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
      streamTrace: config.stream_trace !== false,
      aggregatorOverrides: aggregatorOverridesToRows(config.aggregator_body_overrides)
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
    show(errorMessage(error, t('admin.ensemble.errors.loadGroupConfig')), 'err')
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
    show(t('admin.ensemble.errors.sharedChannelRequired'), 'err')
    return
  }
  const duplicate = allGroups.value.find(group =>
    group.name.trim().toLowerCase() === draft.value.name.trim().toLowerCase() &&
    (creating || group.id !== selectedEnsembleGroupId.value)
  )
  if (duplicate) {
    show(t('admin.ensemble.errors.duplicateName', { name: draft.value.name.trim() }), 'err')
    return
  }

  saving.value = true
  let createdGroupId: number | null = null
  let creationConfigured = false
  try {
    let groupId = creating ? null : selectedEnsembleGroupId.value
    if (groupId === null) {
      const channel = effectiveBillingChannel
      if (!channel) throw new Error(t('admin.ensemble.errors.noSharedChannel'))
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
      source_group_ids: [...selectedSourceGroupIds.value],
      aggregator_body_overrides: rowsToAggregatorOverrides(options.value.aggregatorOverrides)
    })
    creationConfigured = true
    localStorage.setItem(TARGET_STORAGE_KEY, String(groupId))
    await reloadGroupsAndTarget(groupId)
    show(aggregator.value
      ? t('admin.ensemble.messages.savedWithAggregator', {
        name: draft.value.name.trim(),
        count: proposers.value.length,
        aggregator: aggregator.value
      })
      : t('admin.ensemble.messages.savedWithoutAggregator', {
        name: draft.value.name.trim(),
        count: proposers.value.length
      }), 'ok')
  } catch (error) {
    if (createdGroupId !== null && !creationConfigured) {
      try {
        await groupsAPI.delete(createdGroupId)
      } catch {
        show(t('admin.ensemble.errors.rollbackFailed', { name: draft.value.name.trim() }), 'err')
        return
      }
    }
    show(t('admin.ensemble.errors.saveFailed', {
      message: errorMessage(error, t('admin.ensemble.errors.unknownServerError'))
    }), 'err')
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
    show(t('admin.ensemble.errors.saveBeforeTest'), 'warn')
    return
  }
  if (!testApiKey.value.trim()) {
    show(t('admin.ensemble.errors.testKeyRequired'), 'warn')
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
      { role: 'user', content: t('admin.ensemble.test.prompt') }
    ], event => {
      testEvents.value.push(event)
      if (event.type === 'error' || event.type === 'fallback') {
        testDialogError.value = event.error ?? (event.type === 'fallback' ? t('admin.ensemble.messages.fallbackNotice') : '')
      }
      if (event.type === 'completed' && event.response) {
        testResult.value = normalizeTestResult(event.response)
      }
    }, testAbortController.value.signal)
    if (data && !testResult.value) testResult.value = normalizeTestResult(data)
    if (testDialogError.value) {
      show(t('admin.ensemble.messages.testFinished', { message: testDialogError.value }), 'warn')
    } else if (testResult.value) {
      show(t('admin.ensemble.messages.testCompleted', {
        succeeded: testResult.value.successCount,
        total: testResult.value.members.length,
        duration: testResult.value.durationText
      }), testResult.value.successCount > 0 ? 'ok' : 'warn')
    }
  } catch (error) {
    if ((error as { name?: string })?.name === 'AbortError') {
      testDialogError.value = t('admin.ensemble.messages.testCancelled')
      show(t('admin.ensemble.messages.testCancelled'), 'warn')
    } else {
      testDialogError.value = errorMessage(error, t('admin.ensemble.errors.gatewayFailed'))
      show(t('admin.ensemble.errors.testFailed', { message: testDialogError.value }), 'err')
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
    warning: metadataPresent ? undefined : t('admin.ensemble.result.metadataWarning')
  }
}

function formatDuration(value: number): string {
  return value > 0 ? `${(value / 1000).toFixed(1)}s` : '—'
}

function formatCost(value?: number): string {
  return typeof value === 'number' ? `$${value.toFixed(6)}` : t('admin.ensemble.stats.costUnavailable')
}

function formatCostSource(source?: string): string {
  if (source === 'upstream') return t('admin.ensemble.costSource.upstream')
  if (source === 'channel') return t('admin.ensemble.costSource.channel')
  if (source === 'litellm') return t('admin.ensemble.costSource.litellm')
  if (source === 'fallback') return t('admin.ensemble.costSource.fallback')
  return source ?? ''
}

function validationMessage(key: string): string {
  // validateEnsembleDraft returns stable kebab-case reason codes; they stay the
  // lookup keys so the util has no reason to know about i18n.
  const messages: Record<string, string> = {
    'name-required': t('admin.ensemble.validation.nameRequired'),
    'source-group-required': t('admin.ensemble.validation.sourceGroupRequired'),
    'at-least-two-proposers': t('admin.ensemble.validation.atLeastTwoProposers'),
    'invalid-min-proposers': t('admin.ensemble.validation.invalidMinProposers')
  }
  return messages[key] ?? t('admin.ensemble.validation.generic')
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
