<template>
  <BaseDialog :show="show" title="TokenPort 接入配置中心" width="full" panel-class="tokenport-access-dialog" @close="emit('close')">
    <div class="flex min-h-0 flex-col gap-4">
      <header class="flex flex-col gap-3 border-b border-gray-200 pb-4 dark:border-dark-700 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ keyName || '当前 API Key' }}</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ platformDescription }}</p>
        </div>
        <div class="inline-flex w-fit rounded-lg bg-gray-100 p-1 dark:bg-dark-800">
          <button v-for="mode in deliveryModes" :key="mode.id" type="button" class="rounded-md px-4 py-2 text-sm font-medium transition-colors" :class="deliveryMode === mode.id ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-700 dark:text-primary-300' : 'text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-white'" @click="deliveryMode = mode.id">
            {{ mode.label }}
          </button>
        </div>
      </header>

      <div v-if="!platform" class="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
        当前密钥尚未关联分组，无法判断客户端协议。请先设置可用分组。
      </div>

      <div v-else class="grid min-h-0 gap-4 lg:grid-cols-12">
        <section class="space-y-4 lg:col-span-5 lg:max-h-[66vh] lg:overflow-y-auto lg:pr-2">
          <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
            <p class="mb-3 text-xs font-semibold uppercase text-gray-500 dark:text-gray-400">目标客户端</p>
            <div class="flex flex-wrap gap-2">
              <button v-for="tab in clientTabs" :key="tab.id" type="button" class="rounded-lg border px-3 py-2 text-sm transition-colors" :class="activeClientTab === tab.id ? 'border-primary-400 bg-primary-50 text-primary-700 dark:border-primary-600 dark:bg-primary-900/20 dark:text-primary-300' : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-dark-600 dark:text-gray-300'" @click="activeClientTab = tab.id">
                {{ tab.label }}
              </button>
            </div>
            <div v-if="activeClientTab === 'codex'" class="mt-3 rounded-md bg-gray-50 px-3 py-2 text-xs leading-5 text-gray-500 dark:bg-dark-900/60 dark:text-gray-400">
              Codex 配置层可用于 ChatGPT 桌面端、Codex CLI 和 IDE 扩展；本页生成的是 API Key 接入配置，ChatGPT 账号登录属于独立认证方式。
            </div>
            <div v-if="showCodexTransport" class="mt-3">
              <div class="mb-2 flex items-center justify-between gap-2">
                <span class="text-xs font-medium text-gray-600 dark:text-gray-300">连接方式</span>
                <span class="text-xs text-gray-400">只影响 Responses API 的传输方式</span>
              </div>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="transport in codexTransports"
                  :key="transport.id"
                  type="button"
                  class="rounded-md px-3 py-1.5 text-xs font-medium transition-colors"
                  :class="codexTransport === transport.id ? 'bg-gray-900 text-white dark:bg-gray-100 dark:text-gray-900' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
                  @click="codexTransport = transport.id"
                >
                  {{ transport.label }}
                </button>
              </div>
            </div>
            <div v-if="showShellTabs" class="mt-4 flex flex-wrap gap-2">
              <button v-for="tab in shellTabs" :key="tab.id" type="button" class="rounded-md px-3 py-1.5 text-xs font-medium" :class="activeShell === tab.id ? 'bg-gray-900 text-white dark:bg-gray-100 dark:text-gray-900' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'" @click="activeShell = tab.id">{{ tab.label }}</button>
            </div>
          </div>

          <ConnectorOptions v-model="connectorOptions" :platform="platform" :client="activeClientTab" :available-model-options="availableModels" />
          <SkillMarketSelector v-if="supportsSkills" v-model="selectedSkills" />
        </section>

        <section class="flex min-h-[520px] flex-col overflow-hidden rounded-lg border border-gray-200 bg-gray-950 lg:col-span-7 lg:max-h-[66vh] dark:border-dark-700">
          <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-800 bg-gray-900 px-4 py-3">
            <div class="flex min-w-0 flex-wrap gap-2">
              <button v-for="(file, index) in currentFiles" :key="`${file.path}-${index}`" type="button" class="max-w-[220px] truncate rounded-md px-3 py-1.5 text-xs font-medium" :class="activeFileIndex === index ? 'bg-primary-500 text-white' : 'bg-gray-800 text-gray-300 hover:bg-gray-700'" @click="activeFileIndex = index">{{ file.path }}</button>
            </div>
            <div class="flex items-center gap-2">
              <button type="button" class="rounded-md bg-gray-800 px-3 py-1.5 text-xs text-gray-200 hover:bg-gray-700" @click="revealSecrets = !revealSecrets">{{ revealSecrets ? '隐藏密钥' : '显示密钥' }}</button>
              <button type="button" class="rounded-md bg-gray-800 px-3 py-1.5 text-xs text-gray-200 hover:bg-gray-700" @click="advancedEditing = !advancedEditing">{{ advancedEditing ? '普通预览' : '高级编辑' }}</button>
            </div>
          </div>

          <div class="min-h-0 flex-1 overflow-auto p-4">
            <textarea v-if="advancedEditing && activeFile" v-model="editableFiles[activeFileIndex]" class="h-full min-h-[430px] w-full resize-none border-0 bg-transparent font-mono text-sm leading-6 text-gray-100 outline-none" spellcheck="false" />
            <pre v-else class="min-h-[430px] whitespace-pre-wrap break-words font-mono text-sm leading-6 text-gray-100"><code>{{ activePreviewContent }}</code></pre>
          </div>

          <div class="flex flex-wrap items-center justify-between gap-3 border-t border-gray-800 bg-gray-900 px-4 py-3">
            <p class="text-xs text-gray-400">{{ deliveryMode === 'ccs' ? 'CCS 只导入客户端配置，Skill 仍通过独立脚本安装。' : '配置写入前会备份原文件，并保留已有项目设置。' }}</p>
            <div class="flex flex-wrap gap-2">
              <button type="button" class="btn btn-secondary" :disabled="!activeFile" @click="downloadActiveFile">下载当前文件</button>
              <button type="button" class="btn btn-secondary" :disabled="!activeFile" @click="copyActiveFile">复制当前文件</button>
              <button v-if="deliveryMode === 'ccs'" type="button" class="btn btn-primary" @click="importToCcs">导入 CCS</button>
            </div>
          </div>
        </section>
      </div>
    </div>

    <template #footer>
      <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex flex-wrap gap-2">
          <button v-if="deliveryMode === 'direct' && clientInstallScript" type="button" class="btn btn-primary" @click="copyClientInstallScript">{{ scriptCopied ? '一键配置脚本已复制' : '复制一键配置脚本' }}</button>
          <button v-if="deliveryMode === 'direct' && skillInstallScript" type="button" class="btn btn-secondary" @click="copySkillInstallScript">复制 Skill 安装脚本</button>
        </div>
        <button type="button" class="btn btn-secondary" @click="emit('close')">关闭</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConnectorOptions from '@/components/keys/ConnectorOptions.vue'
import SkillMarketSelector from '@/components/keys/SkillMarketSelector.vue'
import { useClipboard } from '@/composables/useClipboard'
import type { GroupPlatform } from '@/types'
import userChannelsAPI from '@/api/channels'
import type { SkillInstallSelection } from '@/api/skillMarket'
import { extractConnectorModelOptions, type ConnectorModelOption } from '@/utils/connectorModelOptions'
import { buildCcSwitchImportDeeplink, type CcSwitchClientType } from '@/utils/ccswitchImport'
import { DEFAULT_CONNECTOR_OPTIONS, normalizeConnectorOptions, type ConnectorOptions as ConnectorOptionsState } from '@/constants/connectorPresets'
import {
  buildAnthropicFiles,
  buildOpenAIFiles,
  buildOpenAIWsFiles,
  type ConnectorShell,
  type FileConfig,
} from '@/components/keys/connectorTemplates'
import {
  buildClientInstallScript,
  buildGeminiFiles,
  buildGrokFiles,
  buildOpenCodeFiles,
  ensureApiVersion,
  isSkillInstallFile,
} from './accessCenterFiles'

type DeliveryMode = 'direct' | 'ccs'

const props = withDefaults(defineProps<{
  show: boolean
  apiKey: string
  baseUrl: string
  platform: GroupPlatform | null
  groupId?: number | null
  allowMessagesDispatch?: boolean
  initialMode?: DeliveryMode
  keyName?: string
}>(), {
  initialMode: 'direct',
  groupId: null,
  allowMessagesDispatch: false,
  keyName: '',
})

const emit = defineEmits<{ (event: 'close'): void }>()
const { copyToClipboard } = useClipboard()
const deliveryMode = ref<DeliveryMode>(props.initialMode)
const activeClientTab = ref('claude')
const activeShell = ref<ConnectorShell>('unix')
const codexTransport = ref<'responses' | 'websocket'>('responses')
const connectorOptions = ref<ConnectorOptionsState>(structuredClone(DEFAULT_CONNECTOR_OPTIONS))
const selectedSkills = ref<SkillInstallSelection[]>([])
const availableModels = ref<ConnectorModelOption[]>([])
const activeFileIndex = ref(0)
const revealSecrets = ref(false)
const advancedEditing = ref(false)
const editableFiles = ref<string[]>([])
const scriptCopied = ref(false)

const deliveryModes: Array<{ id: DeliveryMode; label: string }> = [
  { id: 'direct', label: '直接配置' },
  { id: 'ccs', label: '导入 CCS' },
]
const shellTabs: Array<{ id: ConnectorShell; label: string }> = [
  { id: 'unix', label: 'macOS / Linux' },
  { id: 'cmd', label: 'Windows CMD' },
  { id: 'powershell', label: 'PowerShell' },
]
const codexTransports: Array<{ id: 'responses' | 'websocket'; label: string }> = [
  { id: 'responses', label: '标准连接（推荐）' },
  { id: 'websocket', label: 'WebSocket' },
]

const defaultClient = computed(() => {
  if (props.platform === 'openai') return 'codex'
  if (props.platform === 'grok') return 'codex'
  if (props.platform === 'gemini') return 'gemini'
  return 'claude'
})

const clientTabs = computed(() => {
  if (props.platform === 'openai') {
    const tabs = [
      { id: 'codex', label: 'Codex' },
      { id: 'opencode', label: 'OpenCode' },
    ]
    if (props.allowMessagesDispatch) tabs.splice(2, 0, { id: 'claude', label: 'Claude Code' })
    return tabs
  }
  if (props.platform === 'grok') {
    return [
      { id: 'codex', label: 'Codex' },
      { id: 'claude', label: 'Claude Code' },
      { id: 'grok', label: 'Grok CLI' },
      { id: 'opencode', label: 'OpenCode' },
    ]
  }
  if (props.platform === 'gemini') return [{ id: 'gemini', label: 'Gemini CLI' }, { id: 'opencode', label: 'OpenCode' }]
  if (props.platform === 'antigravity') return [{ id: 'claude', label: 'Claude Code' }, { id: 'gemini', label: 'Gemini CLI' }, { id: 'opencode', label: 'OpenCode' }]
  return [{ id: 'claude', label: 'Claude Code' }, { id: 'opencode', label: 'OpenCode' }]
})

const platformDescription = computed(() => {
  if (!props.platform) return '请先为 API Key 配置可用分组。'
  const labels: Record<string, string> = { openai: 'OpenAI 兼容协议', anthropic: 'Anthropic 兼容协议', gemini: 'Gemini 兼容协议', antigravity: 'Gemini / Anthropic 双协议', grok: 'Grok · Responses / Messages 双协议' }
  return `${labels[props.platform] || props.platform} · 选择客户端、模型与能力后生成可检查的配置`
})

const showShellTabs = computed(() => activeClientTab.value !== 'opencode')
const showCodexTransport = computed(() => props.platform === 'openai' && activeClientTab.value === 'codex')
const supportsSkills = computed(() => activeClientTab.value === 'claude' || activeClientTab.value.startsWith('codex'))

const generatedFiles = computed<FileConfig[]>(() => {
  const baseUrl = (props.baseUrl || window.location.origin).replace(/\/+$/, '')
  const common = { baseUrl, apiKey: props.apiKey, shell: activeShell.value, options: connectorOptions.value, selectedSkills: selectedSkills.value }
  if (activeClientTab.value === 'opencode') {
    const provider = props.platform === 'anthropic' ? 'anthropic' : props.platform === 'gemini' ? 'google' : 'openai'
    const endpoint = props.platform === 'gemini' ? ensureApiVersion(baseUrl, 'v1beta') : ensureApiVersion(baseUrl)
    return buildOpenCodeFiles(endpoint, props.apiKey, provider)
  }
  if (activeClientTab.value === 'gemini') {
    const endpoint = props.platform === 'antigravity' ? `${baseUrl}/antigravity` : baseUrl
    return buildGeminiFiles(endpoint, props.apiKey, activeShell.value)
  }
  if (activeClientTab.value === 'grok') return buildGrokFiles(baseUrl, props.apiKey, activeShell.value, connectorOptions.value)
  if (activeClientTab.value === 'codex') {
    return codexTransport.value === 'websocket' ? buildOpenAIWsFiles(common) : buildOpenAIFiles(common)
  }
  return buildAnthropicFiles({ ...common, baseUrl: props.platform === 'antigravity' ? `${baseUrl}/antigravity` : baseUrl })
})

const currentFiles = computed<FileConfig[]>(() => generatedFiles.value.map((file, index) => ({ ...file, content: editableFiles.value[index] ?? file.content })))
const activeFile = computed(() => currentFiles.value[activeFileIndex.value] || null)
const activePreviewContent = computed(() => {
  const content = activeFile.value?.content || ''
  return revealSecrets.value || !props.apiKey ? content : content.split(props.apiKey).join(maskedApiKey.value)
})
const maskedApiKey = computed(() => props.apiKey.length > 8 ? `${props.apiKey.slice(0, 5)}${'•'.repeat(12)}${props.apiKey.slice(-4)}` : '••••••••••••')
const clientFiles = computed(() => currentFiles.value.filter((file) => !isSkillInstallFile(file)))
const skillInstallFile = computed(() => currentFiles.value.find(isSkillInstallFile) || null)
const clientInstallScript = computed(() => buildClientInstallScript(clientFiles.value, activeShell.value))
const skillInstallScript = computed(() => skillInstallFile.value?.content || '')

watch(() => props.show, (visible) => {
  if (!visible) return
  deliveryMode.value = props.initialMode
  activeClientTab.value = defaultClient.value
  activeShell.value = 'unix'
  codexTransport.value = 'responses'
  connectorOptions.value = structuredClone(DEFAULT_CONNECTOR_OPTIONS)
  selectedSkills.value = []
  revealSecrets.value = false
  advancedEditing.value = false
  void loadModels()
}, { immediate: true })

watch(generatedFiles, (files) => {
  editableFiles.value = files.map((file) => file.content)
  activeFileIndex.value = Math.min(activeFileIndex.value, Math.max(files.length - 1, 0))
}, { immediate: true })

async function loadModels() {
  availableModels.value = []
  if (!props.platform) return
  try {
    const channels = await userChannelsAPI.getAvailable()
    availableModels.value = extractConnectorModelOptions({ channels, platform: props.platform, groupId: props.groupId })
  } catch {
    availableModels.value = []
  }
}

async function copyActiveFile() {
  if (activeFile.value) await copyToClipboard(activeFile.value.content, '配置文件已复制')
}

function downloadActiveFile() {
  if (!activeFile.value) return
  const blob = new Blob([activeFile.value.content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = activeFile.value.path.split(/[\\/]/).pop() || 'tokenport-config.txt'
  anchor.click()
  URL.revokeObjectURL(url)
}

async function copyClientInstallScript() {
  if (!clientInstallScript.value) return
  if (await copyToClipboard(clientInstallScript.value, '一键配置脚本已复制')) {
    scriptCopied.value = true
    setTimeout(() => { scriptCopied.value = false }, 2000)
  }
}

async function copySkillInstallScript() {
  if (skillInstallScript.value) await copyToClipboard(skillInstallScript.value, 'Skill 安装脚本已复制')
}

function ccsClientType(): CcSwitchClientType {
  if (activeClientTab.value === 'gemini') return 'gemini'
  if (activeClientTab.value === 'codex' || activeClientTab.value === 'codex-ws' || activeClientTab.value === 'grok') return 'codex'
  return 'claude'
}

function ccsConfig(): { config?: string; configFormat?: 'json' | 'toml' } {
  if (activeClientTab.value.startsWith('codex')) {
    const auth = clientFiles.value.find((file) => file.path.endsWith('auth.json'))?.content || '{}'
    const config = clientFiles.value.find((file) => file.path.endsWith('config.toml'))?.content || ''
    return { config: JSON.stringify({ auth: JSON.parse(auth), config }, null, 2), configFormat: 'json' }
  }
  if (activeClientTab.value === 'grok') {
    const config = clientFiles.value.find((file) => file.path.endsWith('config.toml'))?.content
    return config ? { config, configFormat: 'toml' } : {}
  }
  const settings = clientFiles.value.find((file) => file.path.endsWith('settings.json'))?.content
  return settings ? { config: settings, configFormat: 'json' } : {}
}

function importToCcs() {
  const normalized = normalizeConnectorOptions(connectorOptions.value)
  const payload = ccsConfig()
  const clientType = ccsClientType()
  const usageScript = `({ endpoint: ${JSON.stringify(props.baseUrl || window.location.origin)}, key: ${JSON.stringify(props.apiKey)} })`
  const deeplink = buildCcSwitchImportDeeplink({
    baseUrl: props.baseUrl || window.location.origin,
    platform: props.platform,
    clientType,
    providerName: 'TokenPort',
    apiKey: props.apiKey,
    usageScript,
    model: clientType === 'codex' ? normalized.codex.model || undefined : undefined,
    claudeModelTiers: normalized.claude.modelTiers,
    ...payload,
  })
  window.location.href = deeplink
}
</script>

<style>
@media (max-width: 640px) {
  .modal-overlay:has(.tokenport-access-dialog) {
    padding: 0;
  }

  .tokenport-access-dialog {
    height: 100dvh;
    max-height: 100dvh;
    border-radius: 0;
    border-left: 0;
    border-right: 0;
  }

  .tokenport-access-dialog .modal-body {
    padding-top: 14px;
    padding-bottom: 14px;
  }
}
</style>
