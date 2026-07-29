<template>
  <div class="flow-preview" aria-label="TokenPort 模型与工具统一治理示意图">
    <div class="flow-grid" aria-hidden="true" />
    <div class="flow-status"><i />资源与策略在线</div>
    <svg viewBox="0 0 560 432" role="img" aria-label="多种 AI 工具经 TokenPort 接入计量、预算、技能和审计能力">
      <defs>
        <linearGradient id="tp-flow-line" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stop-color="var(--brand)" stop-opacity="0" />
          <stop offset="50%" stop-color="var(--brand)" stop-opacity="0.95" />
          <stop offset="100%" stop-color="var(--brand)" stop-opacity="0" />
        </linearGradient>
        <radialGradient id="tp-flow-core" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stop-color="var(--brand)" stop-opacity="0.78" />
          <stop offset="72%" stop-color="var(--brand)" stop-opacity="0.12" />
          <stop offset="100%" stop-color="var(--brand)" stop-opacity="0" />
        </radialGradient>
      </defs>

      <g v-for="(node, index) in inputNodes" :key="`line-${node.id}`">
        <path :d="inputPath(node)" class="base-line" />
        <path
          :d="inputPath(node)"
          class="motion-line"
          :style="{ animationDuration: `${3 + index * 0.35}s` }"
        />
      </g>

      <g v-for="(item, index) in outputs" :key="`output-${item.label}`">
        <path :d="outputPath(item.y)" class="base-line soft" />
        <path
          :d="outputPath(item.y)"
          class="motion-line"
          :style="{ animationDuration: `${3.2 + index * 0.3}s` }"
        />
        <circle cx="506" :cy="item.y" r="3.5" class="endpoint" />
      </g>

      <g v-for="node in inputNodes" :key="node.id">
        <rect :x="node.x" :y="node.y" width="126" height="32" rx="8" class="node-card" />
        <circle :cx="node.x + 16" :cy="node.y + 16" r="3.5" class="node-dot" />
        <text :x="node.x + 29" :y="node.y + 20">{{ node.label }}</text>
      </g>

      <circle :cx="port.x" :cy="port.y" r="72" fill="url(#tp-flow-core)" />
      <circle :cx="port.x" :cy="port.y" r="48" class="core-orbit" />
      <circle :cx="port.x" :cy="port.y" r="35" class="core" />
      <text :x="port.x" :y="port.y - 2" text-anchor="middle" class="core-label">TOKEN</text>
      <text :x="port.x" :y="port.y + 13" text-anchor="middle" class="core-name">PORT</text>

      <text
        v-for="item in outputs"
        :key="`label-${item.label}`"
        x="496"
        :y="item.y + 4"
        text-anchor="end"
        class="output-label"
      >{{ item.label }}</text>
    </svg>
    <div class="flow-caption">
      <span><i />统一接入</span>
      <span><i />策略治理</span>
      <span><i />能力交付</span>
    </div>
  </div>
</template>

<script setup lang="ts">
interface InputNode {
  id: string
  label: string
  x: number
  y: number
}

const inputNodes: InputNode[] = [
  { id: 'chatgpt', label: 'ChatGPT', x: 34, y: 44 },
  { id: 'codex', label: 'Codex', x: 28, y: 126 },
  { id: 'claude', label: 'Claude Code', x: 42, y: 208 },
  { id: 'opencode', label: 'OpenCode', x: 34, y: 290 },
  { id: 'gemini', label: 'Gemini CLI', x: 28, y: 372 },
]

const outputs = [
  { label: 'Token 计量', y: 90 },
  { label: '预算护栏', y: 176 },
  { label: 'Skill 路由', y: 262 },
  { label: '审计日志', y: 348 },
]

const port = { x: 300, y: 216 }

function inputPath(node: InputNode) {
  const x = node.x + 63
  const y = node.y + 16
  return `M ${x} ${y} C ${x + 88} ${y}, ${port.x - 88} ${port.y}, ${port.x - 35} ${port.y}`
}

function outputPath(y: number) {
  return `M ${port.x + 35} ${port.y} C ${port.x + 92} ${port.y}, 468 ${y}, 506 ${y}`
}
</script>

<style scoped>
.flow-preview {
  --brand: var(--tp-brand);
  position: relative;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--brand) 24%, var(--line));
  border-radius: var(--tp-radius-panel);
  background: color-mix(in srgb, var(--surface) 76%, transparent);
  box-shadow: var(--tp-elev-3);
}

.flow-grid {
  position: absolute;
  inset: 0;
  opacity: 0.48;
  background-image: radial-gradient(circle at center, color-mix(in srgb, var(--brand) 18%, transparent) 1px, transparent 1.4px);
  background-size: 22px 22px;
  mask-image: linear-gradient(to bottom, transparent, #000 12%, #000 88%, transparent);
}

.flow-status {
  position: absolute;
  z-index: 2;
  top: 16px;
  right: 18px;
  display: flex;
  align-items: center;
  gap: 7px;
  color: var(--muted);
  font: 600 11px/1.3 ui-monospace, Consolas, monospace;
}

.flow-status i,
.flow-caption i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--brand);
  box-shadow: 0 0 12px color-mix(in srgb, var(--brand) 68%, transparent);
}

svg {
  position: relative;
  z-index: 1;
  display: block;
  width: 100%;
}

.base-line {
  fill: none;
  stroke: color-mix(in srgb, var(--brand) 30%, var(--line));
  stroke-width: 1.2;
}

.base-line.soft {
  opacity: 0.65;
}

.motion-line {
  fill: none;
  stroke: url(#tp-flow-line);
  stroke-width: 2;
  stroke-dasharray: 6 250;
  animation: tp-flow-dash linear infinite;
}

.node-card {
  fill: color-mix(in srgb, var(--surface) 88%, transparent);
  stroke: color-mix(in srgb, var(--brand) 26%, var(--line));
}

.node-dot,
.endpoint {
  fill: var(--brand);
  animation: tp-flow-pulse 2.4s ease-in-out infinite;
}

text {
  fill: var(--ink);
  font: 600 11px/1 ui-monospace, Consolas, monospace;
}

.core-orbit {
  fill: none;
  stroke: var(--brand);
  stroke-width: 1;
  stroke-dasharray: 3 5;
  opacity: 0.55;
}

.core {
  fill: var(--soft);
  stroke: var(--brand);
  stroke-width: 1.5;
}

.core-label {
  fill: var(--brand);
  font-weight: 800;
}

.core-name {
  fill: var(--ink);
  font-weight: 800;
}

.output-label {
  fill: var(--muted);
  font-size: 10.5px;
}

.flow-caption {
  position: absolute;
  z-index: 2;
  right: 18px;
  bottom: 14px;
  display: flex;
  gap: 16px;
  color: var(--muted);
  font: 600 10px/1.2 ui-monospace, Consolas, monospace;
}

.flow-caption span {
  display: flex;
  align-items: center;
  gap: 6px;
}

@keyframes tp-flow-dash {
  to { stroke-dashoffset: -1000; }
}

@keyframes tp-flow-pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}

@media (prefers-reduced-motion: reduce) {
  .motion-line,
  .node-dot,
  .endpoint { animation: none; }
}
</style>
