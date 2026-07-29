<template>
  <div class="usage-showcase">
    <div class="usage-head">
      <div>
        <b>Token 用量与预算</b>
        <span>按部门、模型与 API Key 归集</span>
      </div>
      <div class="range-switch" aria-label="时间范围">
        <button
          v-for="item in ranges"
          :key="item"
          type="button"
          :class="{ active: range === item }"
          @click="range = item"
        >{{ item }}</button>
      </div>
    </div>

    <div class="usage-metrics">
      <div><span>模型成本</span><b>{{ current.spend }}</b></div>
      <div><span>Token 总量</span><b>{{ current.tokens }}</b></div>
      <div><span>预算使用</span><b>{{ current.budget }}%</b></div>
    </div>

    <div class="chart-wrap">
      <svg viewBox="0 0 640 220" preserveAspectRatio="none" role="img" :aria-label="`${range} Token 用量趋势`">
        <defs>
          <linearGradient id="tp-usage-area" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="var(--brand)" stop-opacity="0.34" />
            <stop offset="100%" stop-color="var(--brand)" stop-opacity="0" />
          </linearGradient>
        </defs>
        <line v-for="grid in [0.25, 0.5, 0.75]" :key="grid" x1="0" x2="640" :y1="220 * grid" :y2="220 * grid" class="chart-grid" />
        <line x1="0" x2="640" :y1="budgetY" :y2="budgetY" class="budget-line" />
        <path :d="paths.area" fill="url(#tp-usage-area)" />
        <path :d="paths.line" class="usage-line" />
      </svg>
      <span class="budget-label" :style="{ top: `${(1 - current.budget / 100) * 100}%` }">预算阈值</span>
    </div>

    <div class="usage-foot">
      <span><i class="actual" />实际用量</span>
      <span><i class="budget" />预算阈值</span>
      <em>演示数据 · 实际以平台归集结果为准</em>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

type Range = '7D' | '30D' | '90D'

const ranges: Range[] = ['7D', '30D', '90D']
const range = ref<Range>('30D')
const data: Record<Range, { points: number[]; spend: string; tokens: string; budget: number }> = {
  '7D': { points: [42, 55, 48, 63, 58, 71, 66], spend: '¥12,800', tokens: '312M', budget: 62 },
  '30D': { points: [30, 44, 39, 52, 60, 55, 68, 64, 72, 70, 78, 74], spend: '¥53,600', tokens: '1.28B', budget: 71 },
  '90D': { points: [22, 34, 30, 41, 38, 50, 47, 58, 55, 66, 62, 70, 68, 76], spend: '¥146,000', tokens: '3.6B', budget: 68 },
}

const current = computed(() => data[range.value])
const budgetY = computed(() => 220 * (1 - current.value.budget / 100))
const paths = computed(() => buildPath(current.value.points, 640, 220))

function buildPath(points: number[], width: number, height: number) {
  const max = Math.max(...points) * 1.15
  const step = width / (points.length - 1)
  const coords = points.map((point, index) => [index * step, height - (point / max) * height])
  const line = coords.map(([x, y], index) => {
    if (index === 0) return `M ${x} ${y}`
    const [previousX, previousY] = coords[index - 1]
    const controlX = (previousX + x) / 2
    return `C ${controlX} ${previousY}, ${controlX} ${y}, ${x} ${y}`
  }).join(' ')
  return { line, area: `${line} L ${width} ${height} L 0 ${height} Z` }
}
</script>

<style scoped>
.usage-showcase {
  --brand: var(--tp-brand);
  min-width: 0;
  overflow: hidden;
  padding: 24px;
  border: 1px solid color-mix(in srgb, var(--brand) 22%, var(--line));
  border-radius: var(--tp-radius-panel);
  background: color-mix(in srgb, var(--surface) 88%, transparent);
  box-shadow: var(--tp-elev-2);
}

.usage-head,
.usage-head > div:first-child,
.usage-metrics,
.range-switch,
.usage-foot,
.usage-foot span {
  display: flex;
}

.usage-head {
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.usage-head > div:first-child {
  flex-direction: column;
  gap: 4px;
}

.usage-head b { font-size: 16px; }
.usage-head span { color: var(--muted); font-size: 12px; }

.range-switch {
  gap: 2px;
  padding: 4px;
  border: 1px solid var(--line);
  border-radius: var(--tp-radius-control);
  background: var(--soft-2);
}

.range-switch button {
  min-width: 44px;
  min-height: 32px;
  border: 0;
  border-radius: calc(var(--tp-radius-control) - 2px);
  background: transparent;
  color: var(--muted);
  font: 700 11px/1 ui-monospace, Consolas, monospace;
  cursor: pointer;
}

.range-switch button.active {
  background: var(--brand);
  color: #fff;
}

.usage-metrics {
  gap: 1px;
  margin-top: 22px;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: var(--tp-radius-card);
  background: var(--line);
}

.usage-metrics div {
  flex: 1;
  min-width: 0;
  padding: 13px 15px;
  background: var(--surface);
}

.usage-metrics span {
  display: block;
  color: var(--muted);
  font-size: 11px;
}

.usage-metrics b {
  display: block;
  margin-top: 6px;
  font: 800 20px/1.2 ui-monospace, Consolas, monospace;
}

.chart-wrap {
  position: relative;
  height: 240px;
  margin-top: 22px;
}

.chart-wrap svg {
  width: 100%;
  height: 100%;
}

.chart-grid {
  stroke: var(--line);
  stroke-width: 1;
}

.budget-line {
  stroke: color-mix(in srgb, var(--brand) 68%, #5bb8ff);
  stroke-width: 1.2;
  stroke-dasharray: 5 5;
}

.usage-line {
  fill: none;
  stroke: var(--brand);
  stroke-width: 2.6;
}

.budget-label {
  position: absolute;
  right: 0;
  transform: translateY(-120%);
  color: var(--brand-deep);
  font: 700 10px/1 ui-monospace, Consolas, monospace;
}

.usage-foot {
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--line);
  color: var(--muted);
  font-size: 11px;
}

.usage-foot span { align-items: center; gap: 6px; }
.usage-foot i { width: 9px; height: 9px; border-radius: 50%; }
.usage-foot i.actual { background: var(--brand); }
.usage-foot i.budget { border: 1px dashed var(--brand); }
.usage-foot em { margin-left: auto; color: var(--muted); font-style: normal; }

@media (max-width: 560px) {
  .usage-head { align-items: flex-start; flex-direction: column; }
  .usage-metrics { display: grid; grid-template-columns: 1fr; }
  .chart-wrap { height: 190px; }
  .usage-foot em { width: 100%; margin-left: 0; }
}
</style>
