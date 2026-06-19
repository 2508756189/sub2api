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
        <span class="block text-xs text-gray-500 dark:text-gray-400">模型档位、官方插件、Codex MCP 按需生成到配置中</span>
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
              <input
                :value="value.claude?.modelTiers?.[tier.id] || ''"
                :placeholder="tier.placeholder"
                class="input input-sm w-full"
                @input="updateClaudeTier(tier.id, ($event.target as HTMLInputElement).value)"
              />
            </label>
          </div>
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
          <select
            v-if="availableModels.length > 0"
            class="input input-sm w-full"
            :value="value.codex?.model || availableModels[0]"
            @change="updateCodexModel(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="model in availableModels" :key="model" :value="model">
              {{ model }}
            </option>
          </select>
          <input
            v-else
            :value="value.codex?.model || ''"
            class="input input-sm w-full"
            placeholder="需同步渠道模型，或手动填写真实模型名"
            @input="updateCodexModel(($event.target as HTMLInputElement).value)"
          />
          <p v-if="availableModels.length === 0" class="text-xs text-amber-600 dark:text-amber-400">
            当前分组没有可用模型数据，请先在渠道定价/模型列表中同步，或手动填写上游真实模型名。
          </p>
        </label>

        <label class="space-y-1">
          <span class="text-xs font-medium text-gray-600 dark:text-gray-300">推理强度</span>
          <select class="input input-sm w-full" :value="value.codex?.reasoningEffort || 'xhigh'" @change="updateCodexReasoning(($event.target as HTMLSelectElement).value as CodexReasoningEffort)">
            <option v-for="effort in CODEX_REASONING_EFFORTS" :key="effort" :value="effort">
              {{ effort }}
            </option>
          </select>
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
import { computed, ref, watch } from 'vue'
import type { GroupPlatform } from '@/types'
import {
  CLAUDE_MODEL_TIERS,
  CLAUDE_PLUGINS,
  CODEX_MCP_SERVERS,
  CODEX_REASONING_EFFORTS,
  type ClaudeModelTier,
  type CodexReasoningEffort,
  type ConnectorOptions,
  normalizeConnectorOptions,
} from '@/constants/connectorPresets'

const props = defineProps<{
  modelValue?: ConnectorOptions
  platform: GroupPlatform | null
  client: string
  availableModels?: string[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ConnectorOptions): void
}>()

const expanded = ref(true)

const value = computed(() => normalizeConnectorOptions(props.modelValue))
const availableModels = computed(() => props.availableModels ?? [])
const showClaudeOptions = computed(() => props.client === 'claude')
const showCodexOptions = computed(() => props.client === 'codex' || props.client === 'codex-ws')
const selectedPlugins = computed(() => value.value.claude.enabledPlugins)
const selectedMcpServers = computed(() => value.value.codex.mcpServers)

watch([availableModels, showCodexOptions], ([models, enabled]) => {
  if (!enabled || models.length === 0 || value.value.codex.model) return
  updateCodexModel(models[0])
})

function emitValue(next: ConnectorOptions) {
  emit('update:modelValue', normalizeConnectorOptions(next))
}

function updateClaudeTier(tier: ClaudeModelTier, model: string) {
  const next = normalizeConnectorOptions(props.modelValue)
  next.claude.modelTiers = { ...next.claude.modelTiers, [tier]: model.trim() }
  emitValue(next)
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
  next.codex.model = model
  emitValue(next)
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
