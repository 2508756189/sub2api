<template>
  <div class="connector-showcase">
    <div class="client-tabs" role="tablist" aria-label="客户端接入示例">
      <button
        v-for="client in clients"
        :key="client"
        type="button"
        role="tab"
        :aria-selected="activeClient === client"
        :class="{ active: activeClient === client }"
        @click="activeClient = client"
      >{{ client }}</button>
    </div>
    <div class="workspace">
      <aside>
        <span v-for="client in clients" :key="client" :class="{ active: activeClient === client }">
          {{ activeClient === client ? '›' : '' }} {{ configs[client].file }}
        </span>
      </aside>
      <div class="editor">
        <div class="editor-head">
          <span>{{ configs[activeClient].path }}</span>
          <em>配置可检查 · 写入前自动备份</em>
        </div>
        <code>
          <span v-for="line in configs[activeClient].lines" :key="line.key">
            <i>{{ line.key }}</i><b> = </b><strong>{{ line.value }}</strong>
          </span>
        </code>
        <div class="editor-status"><i />已按当前 API Key 权限生成 {{ activeClient }} 配置</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const clients = ['Codex', 'Claude Code', 'OpenCode', 'ChatGPT', 'Gemini CLI'] as const
type Client = typeof clients[number]

const activeClient = ref<Client>('Codex')
const configs: Record<Client, { file: string; path: string; lines: Array<{ key: string; value: string }> }> = {
  Codex: {
    file: 'config.toml',
    path: '~/.codex/config.toml',
    lines: [
      { key: 'model_provider', value: '"TokenPort"' },
      { key: 'base_url', value: '"https://tokenport.example/v1"' },
      { key: 'model', value: '"来自当前分组的可用模型"' },
      { key: 'skills', value: '["已选择的能力包"]' },
    ],
  },
  'Claude Code': {
    file: 'settings.json',
    path: '~/.claude/settings.json',
    lines: [
      { key: 'ANTHROPIC_BASE_URL', value: '"https://tokenport.example"' },
      { key: 'ANTHROPIC_AUTH_TOKEN', value: '"tp-dept-••••"' },
      { key: 'model', value: '"按当前分组授权选择"' },
      { key: 'enabledPlugins', value: '["用户勾选的插件"]' },
    ],
  },
  OpenCode: {
    file: 'opencode.json',
    path: '~/.config/opencode/opencode.json',
    lines: [
      { key: 'provider.baseURL', value: '"https://tokenport.example/v1"' },
      { key: 'provider.apiKey', value: '"tp-dept-••••"' },
      { key: 'route', value: '"部门分组 · 可用模型"' },
      { key: 'skills', value: '["已选择的能力包"]' },
    ],
  },
  ChatGPT: {
    file: 'provider.json',
    path: 'ChatGPT / Codex 兼容配置',
    lines: [
      { key: 'OPENAI_BASE_URL', value: '"https://tokenport.example/v1"' },
      { key: 'OPENAI_API_KEY', value: '"tp-dept-••••"' },
      { key: 'protocol', value: '"OpenAI 兼容"' },
      { key: 'metering', value: '"按 Key 与模型归集"' },
    ],
  },
  'Gemini CLI': {
    file: '.env',
    path: 'Gemini CLI 环境配置',
    lines: [
      { key: 'GOOGLE_GEMINI_BASE_URL', value: '"https://tokenport.example"' },
      { key: 'GEMINI_API_KEY', value: '"tp-dept-••••"' },
      { key: 'model', value: '"来自当前分组的可用模型"' },
      { key: 'budget_alert', value: '"按部门策略"' },
    ],
  },
}
</script>

<style scoped>
.connector-showcase {
  --brand: var(--tp-brand);
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--brand) 22%, var(--line));
  border-radius: var(--tp-radius-panel);
  background: color-mix(in srgb, var(--surface) 88%, transparent);
  box-shadow: var(--tp-elev-2);
}

.client-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 8px;
  border-bottom: 1px solid var(--line);
  background: var(--soft-2);
}

.client-tabs button {
  min-height: 36px;
  padding: 0 14px;
  border: 1px solid transparent;
  border-radius: var(--tp-radius-control);
  background: transparent;
  color: var(--muted);
  font: 600 12px/1 ui-monospace, Consolas, monospace;
  cursor: pointer;
}

.client-tabs button:hover {
  border-color: var(--line);
  color: var(--ink);
}

.client-tabs button.active {
  border-color: color-mix(in srgb, var(--brand) 38%, transparent);
  background: color-mix(in srgb, var(--brand) 13%, transparent);
  color: var(--brand-deep);
}

.workspace {
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr);
  min-height: 340px;
}

.workspace aside {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 22px 18px;
  border-right: 1px solid var(--line);
  background: color-mix(in srgb, var(--soft-2) 72%, transparent);
}

.workspace aside span {
  color: var(--muted);
  font: 600 11px/1.5 ui-monospace, Consolas, monospace;
}

.workspace aside span.active { color: var(--brand-deep); }

.editor {
  min-width: 0;
  padding: 22px;
  background: color-mix(in srgb, var(--surface) 92%, transparent);
}

.editor-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--line);
}

.editor-head span,
.editor-head em,
.editor code,
.editor-status {
  font-family: ui-monospace, Consolas, monospace;
}

.editor-head span { color: var(--ink); font-size: 12px; font-weight: 700; }
.editor-head em { color: var(--muted); font-size: 10px; font-style: normal; }

.editor code {
  display: grid;
  gap: 10px;
  padding: 24px 0;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.5;
}

.editor code span { white-space: nowrap; }
.editor code i { color: var(--muted); font-style: normal; }
.editor code b { color: var(--muted); font-weight: 400; }
.editor code strong { color: var(--brand-deep); font-weight: 700; }

.editor-status {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--brand-deep);
  font-size: 11px;
}

.editor-status i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--brand);
  box-shadow: 0 0 12px color-mix(in srgb, var(--brand) 68%, transparent);
}

@media (max-width: 700px) {
  .workspace { grid-template-columns: 1fr; }
  .workspace aside { display: none; }
  .editor-head { align-items: flex-start; flex-direction: column; gap: 5px; }
}
</style>
