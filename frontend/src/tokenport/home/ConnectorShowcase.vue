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

    <div class="config-workspace">
      <aside aria-label="配置文件">
        <span
          v-for="client in clients"
          :key="client"
          :class="{ active: activeClient === client }"
        ><i>{{ activeClient === client ? '▸' : '·' }}</i>{{ configs[client].file }}</span>
      </aside>

      <div class="config-editor">
        <div class="editor-head">
          <span><i>#</i> {{ configs[activeClient].file }} · {{ activeClient }} 连接器</span>
          <em>CONFIG</em>
        </div>
        <code>
          <span v-for="line in configs[activeClient].lines" :key="line.key">
            <i>{{ line.key }}</i><b> = </b><strong>{{ line.value }}</strong>
          </span>
        </code>
        <div class="editor-status"><i />配置已生成 · 凭证由 TokenPort 托管</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const clients = ['Codex', 'Claude Code', 'OpenCode', 'ChatGPT', 'Gemini CLI'] as const
type Client = typeof clients[number]

const activeClient = ref<Client>('Codex')
const configs: Record<Client, { file: string; lines: Array<{ key: string; value: string }> }> = {
  Codex: {
    file: 'codex.env',
    lines: [
      { key: 'OPENAI_BASE_URL', value: '"https://tokenport.local/v1"' },
      { key: 'OPENAI_API_KEY', value: '"tp-dept-rnd-••••"' },
      { key: 'model', value: '"来自当前分组的真实可用模型"' },
      { key: 'skills', value: '["compound", "diagnosing-bugs"]' },
    ],
  },
  'Claude Code': {
    file: 'settings.json',
    lines: [
      { key: 'ANTHROPIC_BASE_URL', value: '"https://tokenport.local"' },
      { key: 'ANTHROPIC_AUTH_TOKEN', value: '"tp-dept-rnd-••••"' },
      { key: 'model', value: '"claude-sonnet / opus（按授权）"' },
      { key: 'audit', value: 'true' },
    ],
  },
  OpenCode: {
    file: 'opencode.json',
    lines: [
      { key: 'provider.baseURL', value: '"https://tokenport.local/v1"' },
      { key: 'provider.apiKey', value: '"tp-dept-rnd-••••"' },
      { key: 'route', value: '"部门分组 · 故障切换"' },
      { key: 'skills', value: '["differential-review"]' },
    ],
  },
  ChatGPT: {
    file: 'chatgpt.env',
    lines: [
      { key: 'OPENAI_BASE_URL', value: '"https://tokenport.local/v1"' },
      { key: 'OPENAI_API_KEY', value: '"tp-dept-rnd-••••"' },
      { key: 'protocol', value: '"OpenAI 兼容"' },
      { key: 'metering', value: '"按 Key / 模型归集"' },
    ],
  },
  'Gemini CLI': {
    file: 'gemini.env',
    lines: [
      { key: 'CODE_ASSIST_ENDPOINT', value: '"https://tokenport.local"' },
      { key: 'GEMINI_API_KEY', value: '"tp-dept-rnd-••••"' },
      { key: 'model', value: '"来自当前分组的真实可用模型"' },
      { key: 'budget_alert', value: '"90%"' },
    ],
  },
}
</script>

<style scoped>
.connector-showcase {
  overflow: hidden;
  border: 1px solid var(--color-line);
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--color-panel) 52%, transparent);
}

.client-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 8px;
  border-bottom: 1px solid var(--color-line);
  background: color-mix(in srgb, var(--color-ground-2) 60%, transparent);
}

.client-tabs button {
  min-height: 34px;
  padding: 0 13px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: var(--color-muted);
  font: 600 11px/1 var(--font-mono);
  cursor: pointer;
  transition: 160ms ease;
}

.client-tabs button:hover {
  border-color: var(--color-line-strong);
  color: var(--color-fg);
}

.client-tabs button.active {
  border-color: color-mix(in srgb, var(--color-primary) 28%, transparent);
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  color: var(--color-primary);
}

.config-workspace {
  display: grid;
  grid-template-columns: 176px minmax(0, 1fr);
  min-height: 330px;
}

.config-workspace aside {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 20px 12px;
  border-right: 1px solid var(--color-line);
  background: color-mix(in srgb, var(--color-ground-2) 44%, transparent);
}

.config-workspace aside span {
  display: flex;
  align-items: center;
  gap: 7px;
  min-height: 32px;
  padding: 0 8px;
  border-radius: 7px;
  color: var(--color-faint);
  font: 500 11px/1.4 var(--font-mono);
}

.config-workspace aside span i {
  color: var(--color-faint);
  font-style: normal;
}

.config-workspace aside span.active {
  background: color-mix(in srgb, var(--color-primary) 7%, transparent);
  color: var(--color-primary);
}

.config-workspace aside span.active i { color: var(--color-primary); }

.config-editor {
  min-width: 0;
  padding: 20px 22px;
  background: color-mix(in srgb, var(--color-panel) 68%, transparent);
}

.editor-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 15px;
  border-bottom: 1px solid var(--color-line);
}

.editor-head span,
.editor-head em,
.config-editor code,
.editor-status {
  font-family: var(--font-mono);
}

.editor-head span { color: var(--color-fg); font-size: 11px; font-weight: 600; }
.editor-head span i { color: var(--color-primary); font-style: normal; }
.editor-head em { color: var(--color-faint); font-size: 9px; font-style: normal; letter-spacing: .12em; }

.config-editor code {
  display: grid;
  gap: 12px;
  padding: 27px 0 31px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
}

.config-editor code span { white-space: nowrap; }
.config-editor code i { color: var(--color-muted); font-style: normal; }
.config-editor code b { color: var(--color-faint); font-weight: 400; }
.config-editor code strong { color: var(--color-primary); font-weight: 500; }

.editor-status {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-muted);
  font-size: 10px;
}

.editor-status > i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-primary);
  box-shadow: 0 0 10px color-mix(in srgb, var(--color-primary) 60%, transparent);
}

@media (max-width: 700px) {
  .client-tabs { flex-wrap: nowrap; overflow-x: auto; }
  .client-tabs button { flex: 0 0 auto; }
  .config-workspace { grid-template-columns: 1fr; min-height: 300px; }
  .config-workspace aside { display: none; }
  .config-editor { padding: 18px; }
  .editor-head { align-items: flex-start; flex-direction: column; gap: 7px; }
}
</style>
