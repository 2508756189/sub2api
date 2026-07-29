<template>
  <div class="flow-preview">
    <div class="flow-grid" aria-hidden="true" />
    <svg viewBox="0 0 560 432" role="img" aria-label="TokenPort 将多种 AI 工具路由至统一治理能力">
      <defs>
        <linearGradient id="tp-flow-line" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stop-color="var(--color-primary)" stop-opacity="0" />
          <stop offset="50%" stop-color="var(--color-primary)" stop-opacity="0.9" />
          <stop offset="100%" stop-color="var(--color-primary)" stop-opacity="0" />
        </linearGradient>
        <radialGradient id="tp-flow-core" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stop-color="var(--color-primary)" stop-opacity="0.9" />
          <stop offset="70%" stop-color="var(--color-primary)" stop-opacity="0.12" />
          <stop offset="100%" stop-color="var(--color-primary)" stop-opacity="0" />
        </radialGradient>
      </defs>

      <g v-for="(node, index) in inputNodes" :key="`line-${node.id}`">
        <path :d="inputPath(node)" class="base-line" />
        <path :d="inputPath(node)" class="motion-line" :style="{ animationDuration: `${3 + index * 0.35}s` }" />
      </g>

      <g v-for="(item, index) in outputs" :key="`output-${item.label}`">
        <path :d="outputPath(item.y)" class="base-line soft" />
        <path :d="outputPath(item.y)" class="motion-line" :style="{ animationDuration: `${3.2 + index * 0.3}s` }" />
        <circle cx="508" :cy="item.y" r="3" class="endpoint" />
      </g>

      <g v-for="node in inputNodes" :key="node.id">
        <rect :x="node.x" :y="node.y" width="124" height="30" rx="8" class="node-card" />
        <circle :cx="node.x + 15" :cy="node.y + 15" r="3.5" class="node-dot" />
        <text :x="node.x + 28" :y="node.y + 19">{{ node.label }}</text>
      </g>

      <circle :cx="port.x" :cy="port.y" r="70" fill="url(#tp-flow-core)" />
      <circle :cx="port.x" :cy="port.y" r="34" class="core" />
      <circle :cx="port.x" :cy="port.y" r="48" class="core-orbit" />
      <text :x="port.x" :y="port.y - 2" text-anchor="middle" class="core-label">TOKEN</text>
      <text :x="port.x" :y="port.y + 12" text-anchor="middle" class="core-name">PORT</text>

      <text
        v-for="item in outputs"
        :key="`label-${item.label}`"
        x="500"
        :y="item.y + 4"
        text-anchor="end"
        class="output-label"
      >{{ item.label }}</text>
    </svg>
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
  { id: 'chatgpt', label: 'ChatGPT', x: 40, y: 46 },
  { id: 'codex', label: 'Codex', x: 34, y: 128 },
  { id: 'claude', label: 'Claude Code', x: 48, y: 214 },
  { id: 'opencode', label: 'OpenCode', x: 40, y: 300 },
  { id: 'gemini', label: 'Gemini CLI', x: 34, y: 386 },
]

const outputs = [
  { label: 'Token 计量', y: 90 },
  { label: '预算护栏', y: 176 },
  { label: 'Skill 路由', y: 262 },
  { label: '审计日志', y: 348 },
]

const port = { x: 300, y: 216 }

function inputPath(node: InputNode) {
  const x = node.x + 62
  const y = node.y + 15
  return `M ${x} ${y} C ${x + 90} ${y}, ${port.x - 90} ${port.y}, ${port.x - 34} ${port.y}`
}

function outputPath(y: number) {
  return `M ${port.x + 34} ${port.y} C ${port.x + 90} ${port.y}, 470 ${y}, 508 ${y}`
}
</script>

<style scoped>
.flow-preview { position: relative; width: 100%; overflow: hidden; padding: 16px; border: 1px solid var(--color-line); border-radius: var(--radius); background: var(--color-band); box-shadow: 0 18px 44px -36px var(--color-shadow); }
.flow-grid { position: absolute; inset: 0; background-image: radial-gradient(circle at center, var(--color-grid-dot) 1px, transparent 1.4px); background-size: 22px 22px; opacity: 0.5; pointer-events: none; }
svg { position: relative; display: block; width: 100%; }
.base-line { fill: none; stroke: var(--color-line-strong); stroke-width: 1.2; }
.base-line.soft { stroke: var(--color-line); }
.motion-line { fill: none; stroke: url(#tp-flow-line); stroke-width: 2; stroke-dasharray: 6 260; animation: tp-flow-dash linear infinite; }
.node-card { fill: var(--color-panel); stroke: var(--color-line-strong); }
.node-dot { fill: var(--color-accent); animation: tp-flow-pulse 2.4s ease-in-out infinite; }
.endpoint { fill: var(--color-primary); }
text { fill: var(--color-fg); font: 500 11px/1 'JetBrains Mono', 'Noto Sans SC', monospace; }
.core { fill: var(--color-panel-2); stroke: var(--color-primary); stroke-width: 1.4; }
.core-orbit { fill: none; stroke: var(--color-primary); stroke-width: 1; stroke-dasharray: 3 5; opacity: 0.5; }
.core-label { fill: var(--color-primary); font-weight: 700; }
.core-name { fill: var(--color-fg); font-weight: 700; }
.output-label { fill: var(--color-muted); font-size: 10.5px; }
@keyframes tp-flow-dash { to { stroke-dashoffset: -1000; } }
@keyframes tp-flow-pulse { 0%, 100% { opacity: 0.35; } 50% { opacity: 1; } }
@media (prefers-reduced-motion: reduce) { .motion-line, .node-dot { animation: none; } }
</style>
