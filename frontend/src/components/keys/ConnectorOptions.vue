<template>
  <section
    v-if="showClaudeOptions || showCodexOptions"
    class="rounded-lg border border-gray-200 dark:border-dark-700 bg-white dark:bg-dark-850"
  >
    <button
      type="button"
      class="flex w-full items-center justify-between px-4 py-3 text-left"
      @click="expanded = !expanded"
    >
      <span>
        <span class="block text-sm font-medium text-gray-900 dark:text-gray-100">连接器参数</span>
        <span class="block text-xs text-gray-500 dark:text-gray-400">
          按客户端生成模型、推理强度、插件和 MCP 配置
        </span>
      </span>
      <span class="text-sm text-gray-500">{{ expanded ? '收起' : '展开' }}</span>
    </button>

    <div v-if="expanded" class="space-y-4 border-t border-gray-100 dark:border-dark-700 px-4 py-4">
      <div v-if="showClaudeOptions" class="space-y-4">
        <div>
          <div class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
            Claude 模型档位
          </div>
          <div class="grid gap-3 md:grid-cols-3">
            <label v-for="tier in CLAUDE_MODEL_TIERS" :key="tier.id" class="space-y-1">
              <span class="text-xs font-medium text-gray-600 dark:text-gray-300">{{ tier.label }}</span>
              <Select
                :model-value="claudeTierSelectValue(tier.id)"
                :options="modelSelectOptions"
                value-key="value"
                label-key="label"
                searchable="auto"
                search-placeholder="搜索模型..."
                empty-text="没有匹配的模型"
                @change="(selected) => handleClaudeTierSelect(tier.id, normalizeSelectValue(selected))"
              >
                <template #selected="{ option }">
                  <span class="truncate">{{ option?.label || '不设置，使用客户端默认值' }}</span>
                </template>
                <template #option="{ option }">
                  <div class="min-w-0 flex-1">
                    <div class="truncate text-sm font-medium">{{ option.label }}</div>
                    <div v-if="option.description" class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">
                      {{ option.description }}
                    </div>
                  </div>
                </template>
              </Select>
              <input
                v-if="modelNames.length === 0 || isClaudeTierManual(tier.id)"
                :value="value.claude?.modelTiers?.[tier.id] || ''"
                :placeholder="tier.placeholder"
                class="input input-sm w-full"
                autocomplete="off"
                autocapitalize="off"
                spellcheck="false"
                @input="updateClaudeTier(tier.id, ($event.target as HTMLInputElement).value)"
              />
            </label>
          </div>
          <p v-if="modelNames.length > 0" class="mt-2 text-xs text-gray-500 dark:text-gray-400">
            模型和定价来自当前 API Key 分组可用渠道；也可以手动填写上游真实模型名。
          </p>
          <p v-if="hasClaudeModelOutsideAvailableList" class="mt-2 text-xs text-amber-600 dark:text-amber-400">
            有 Claude 档位填写的模型不在已同步的上游模型列表中，请确认上游确实支持。
          </p>
        </div>

        <div>
          <div class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
            Claude 官方插件
          </div>
          <div class="grid gap-2 md:grid-cols-2">
            <label
              v-for="plugin in CLAUDE_PLUGINS"
              :key="plugin.id"
              class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-200"
            >
              <input
                type="checkbox"
                class="checkbox"
                :checked="selectedPlugins.includes(plugin.id)"
                @change="togglePlugin(plugin.id, ($event.target as HTMLInputElement).checked)"
              />
              <span>{{ plugin.label }}</span>
            </label>
          </div>
        </div>
      </div>

      <div v-if="showCodexOptions" class="grid gap-4 md:grid-cols-3">
        <label class="space-y-1">
          <span class="text-xs font-medium text-gray-600 dark:text-gray-300">Codex 模型</span>
          <Select
            :model-value="codexModelSelectValue"
            :options="modelSelectOptions"
            value-key="value"
            label-key="label"
            searchable="auto"
            search-placeholder="搜索模型..."
            empty-text="没有匹配的模型"
            @change="(selected) => handleCodexModelSelect(normalizeSelectValue(selected))"
          >
            <template #selected="{ option }">
              <span class="truncate">{{ option?.label || '不设置，使用客户端默认值' }}</span>
            </template>
            <template #option="{ option }">
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-medium">{{ option.label }}</div>
                <div v-if="option.description" class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">
                  {{ option.description }}
                </div>
              </div>
            </template>
          </Select>
          <input
            v-if="modelNames.length === 0 || isCodexModelManual"
            :value="value.codex?.model || ''"
            class="input input-sm w-full"
            autocomplete="off"
            autocapitalize="off"
            spellcheck="false"
            placeholder="从上游模型列表选择，或手动填写真实模型名"
            @input="updateCodexModel(($event.target as HTMLInputElement).value)"
          />
          <p v-if="selectedCodexPricingSummary" class="text-xs text-gray-500 dark:text-gray-400">
            当前模型定价：{{ selectedCodexPricingSummary }}
          </p>
          <p v-if="modelNames.length > 0" class="text-xs text-gray-500 dark:text-gray-400">
            模型和定价来自当前 API Key 分组可用渠道；生成后仍可在 Codex 配置文件中修改 model/review_model。
          </p>
          <p v-if="modelNames.length === 0" class="text-xs text-amber-600 dark:text-amber-400">
            当前分组没有可用模型数据，请先在渠道定价/模型列表中同步，或手动填写上游真实模型名。
          </p>
          <p v-if="selectedModelNotInAvailableList" class="text-xs text-amber-600 dark:text-amber-400">
            当前填写的模型不在已同步的上游模型列表中，请确认上游确实支持。
          </p>
        </label>

        <label class="space-y-1">
          <span class="text-xs font-medium text-gray-600 dark:text-gray-300">推理强度</span>
          <Select
            :model-value="value.codex?.reasoningEffort || 'medium'"
            :options="reasoningSelectOptions"
            value-key="value"
            label-key="label"
            :searchable="false"
            @change="(selected) => updateCodexReasoning(normalizeSelectValue(selected) as CodexReasoningEffort)"
          >
            <template #option="{ option }">
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-medium">{{ option.label }}</div>
                <div v-if="option.description" class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ option.description }}
                </div>
              </div>
            </template>
          </Select>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ selectedReasoningOption.description }}
          </p>
        </label>

        <div class="space-y-2">
          <div class="text-xs font-medium text-gray-600 dark:text-gray-300">MCP Servers</div>
          <label
            v-for="server in CODEX_MCP_SERVERS"
            :key="server.id"
            class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-200"
          >
            <input
              type="checkbox"
              class="checkbox"
              :checked="selectedMcpServers.includes(server.id)"
              @change="toggleMcpServer(server.id, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ server.label }}</span>
          </label>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Select from '@/components/common/Select.vue'
import type { GroupPlatform } from '@/types'
import {
  CLAUDE_MODEL_TIERS,
  CLAUDE_PLUGINS,
  CODEX_MCP_SERVERS,
  CODEX_REASONING_EFFORT_OPTIONS,
  type ClaudeModelTier,
  type CodexReasoningEffort,
  type ConnectorOptions,
  normalizeConnectorOptions,
} from '@/constants/connectorPresets'
import type { ConnectorModelOption } from '@/utils/connectorModelOptions'

const props = defineProps<{
  modelValue?: ConnectorOptions
  platform: GroupPlatform | null
  client: string
  availableModels?: string[]
  availableModelOptions?: ConnectorModelOption[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ConnectorOptions): void
}>()

const expanded = ref(true)
const CUSTOM_MODEL_SELECT_VALUE = '__tokenport_custom_model__'
const manualClaudeTierIds = ref<ClaudeModelTier[]>([])
const codexModelManual = ref(false)

const value = computed(() => normalizeConnectorOptions(props.modelValue))
const modelOptions = computed<ConnectorModelOption[]>(() => {
  if (props.availableModelOptions?.length) return props.availableModelOptions
  return (props.availableModels ?? []).map((name) => ({
    name,
    platform: props.platform || '',
    pricing: null,
    pricingSummary: '暂无定价',
    label: name,
    compactLabel: name,
  }))
})
const modelNames = computed(() => modelOptions.value.map((model) => model.name))
const modelSelectOptions = computed(() => [
  {
    value: '',
    label: '不设置，使用客户端默认值',
    description: '不写入模型配置，客户端按自身默认策略选择',
  },
  ...modelOptions.value.map((model) => ({
    value: model.name,
    label: model.name,
    description: model.pricingSummary,
  })),
  {
    value: CUSTOM_MODEL_SELECT_VALUE,
    label: '手动填写模型名',
    description: modelNames.value.length > 0 ? '填写未同步但上游实际支持的模型' : '当前暂无同步模型，可手动填写',
  },
])
const reasoningSelectOptions = computed(() =>
  CODEX_REASONING_EFFORT_OPTIONS.map((effort) => ({
    value: effort.id,
    label: `${effort.label} · ${effort.id}`,
    description: effort.description,
  })),
)
const showClaudeOptions = computed(() => props.client === 'claude')
const showCodexOptions = computed(() => props.client === 'codex' || props.client === 'codex-ws')
const selectedPlugins = computed(() => value.value.claude.enabledPlugins)
const selectedMcpServers = computed(() => value.value.codex.mcpServers)
const selectedClaudeModels = computed(() =>
  CLAUDE_MODEL_TIERS
    .map((tier) => value.value.claude.modelTiers?.[tier.id]?.trim() || '')
    .filter(Boolean),
)
const hasClaudeModelOutsideAvailableList = computed(() =>
  selectedClaudeModels.value.length > 0 &&
  modelNames.value.length > 0 &&
  selectedClaudeModels.value.some((model) => !modelNames.value.includes(model)),
)
const selectedCodexModel = computed(() => value.value.codex.model.trim())
const selectedModelNotInAvailableList = computed(() =>
  selectedCodexModel.value.length > 0 &&
  modelNames.value.length > 0 &&
  !modelNames.value.includes(selectedCodexModel.value),
)
const selectedCodexModelOption = computed(() =>
  modelOptions.value.find((model) => model.name === selectedCodexModel.value),
)
const selectedCodexPricingSummary = computed(() =>
  selectedCodexModelOption.value?.pricingSummary || '',
)
const isCodexModelManual = computed(() =>
  codexModelManual.value ||
  (selectedCodexModel.value.length > 0 && modelNames.value.length > 0 && !modelNames.value.includes(selectedCodexModel.value)),
)
const codexModelSelectValue = computed(() =>
  isCodexModelManual.value ? CUSTOM_MODEL_SELECT_VALUE : selectedCodexModel.value,
)
const selectedReasoningOption = computed(() =>
  CODEX_REASONING_EFFORT_OPTIONS.find((item) => item.id === value.value.codex.reasoningEffort) ||
  CODEX_REASONING_EFFORT_OPTIONS.find((item) => item.id === 'medium') ||
  CODEX_REASONING_EFFORT_OPTIONS[0],
)

function normalizeSelectValue(selected: string | number | boolean | null) {
  return typeof selected === 'string' ? selected : ''
}

function emitValue(next: ConnectorOptions) {
  emit('update:modelValue', normalizeConnectorOptions(next))
}

function updateClaudeTier(tier: ClaudeModelTier, model: string) {
  const next = normalizeConnectorOptions(props.modelValue)
  next.claude.modelTiers = { ...next.claude.modelTiers, [tier]: model.trim() }
  emitValue(next)
}

function isClaudeTierManual(tier: ClaudeModelTier) {
  const current = value.value.claude.modelTiers?.[tier]?.trim() || ''
  return (
    manualClaudeTierIds.value.includes(tier) ||
    (current.length > 0 && modelNames.value.length > 0 && !modelNames.value.includes(current))
  )
}

function claudeTierSelectValue(tier: ClaudeModelTier) {
  const current = value.value.claude.modelTiers?.[tier]?.trim() || ''
  return isClaudeTierManual(tier) ? CUSTOM_MODEL_SELECT_VALUE : current
}

function handleClaudeTierSelect(tier: ClaudeModelTier, selected: string) {
  if (selected === CUSTOM_MODEL_SELECT_VALUE) {
    if (!manualClaudeTierIds.value.includes(tier)) {
      manualClaudeTierIds.value = [...manualClaudeTierIds.value, tier]
    }
    return
  }

  manualClaudeTierIds.value = manualClaudeTierIds.value.filter((item) => item !== tier)
  updateClaudeTier(tier, selected)
}

function togglePlugin(pluginId: string, checked: boolean) {
  const next = normalizeConnectorOptions(props.modelValue)
  const plugins = new Set(next.claude.enabledPlugins)
  checked ? plugins.add(pluginId) : plugins.delete(pluginId)
  next.claude.enabledPlugins = Array.from(plugins)
  emitValue(next)
}

function updateCodexModel(model: string) {
  const next = normalizeConnectorOptions(props.modelValue)
  next.codex.model = model.trim()
  emitValue(next)
}

function handleCodexModelSelect(selected: string) {
  if (selected === CUSTOM_MODEL_SELECT_VALUE) {
    codexModelManual.value = true
    return
  }

  codexModelManual.value = false
  updateCodexModel(selected)
}

function updateCodexReasoning(reasoningEffort: CodexReasoningEffort) {
  const next = normalizeConnectorOptions(props.modelValue)
  next.codex.reasoningEffort = reasoningEffort
  emitValue(next)
}

function toggleMcpServer(serverId: string, checked: boolean) {
  const next = normalizeConnectorOptions(props.modelValue)
  const servers = new Set(next.codex.mcpServers)
  checked ? servers.add(serverId) : servers.delete(serverId)
  next.codex.mcpServers = Array.from(servers)
  emitValue(next)
}
</script>
