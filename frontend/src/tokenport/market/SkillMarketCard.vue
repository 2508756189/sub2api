<template>
  <article class="market-skill-card" :class="`variant-${variant}`">
    <header>
      <span class="skill-icon" aria-hidden="true">
        <Icon :name="categoryIcon" size="md" :stroke-width="1.65" />
      </span>
      <span class="skill-heading">
        <strong>{{ displayName }}</strong>
        <small>{{ categoryName }} · v{{ skill.version }}</small>
      </span>
      <em :class="`risk-${riskLevel}`">{{ riskLabel }}</em>
    </header>

    <p>{{ displayDescription }}</p>

    <div class="skill-chips">
      <span v-for="tag in visibleTags" :key="`tag-${tag}`">{{ tag }}</span>
      <span v-for="runtime in visibleRuntimes" :key="`runtime-${runtime}`" class="runtime-chip">
        {{ runtimeLabel(runtime) }}
      </span>
    </div>

    <footer>
      <span class="archive-meta">
        <span>{{ archiveSize }}</span>
        <i aria-hidden="true">·</i>
        <code>{{ archiveHash }}</code>
      </span>
      <slot name="action" />
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import {
  getSkillCategoryName,
  getSkillDisplayDescription,
  getSkillDisplayName,
  getSkillRiskLabel,
  type SkillMarketEntry,
  type SkillMarketRegistry,
} from '@/api/skillMarket'

const props = withDefaults(defineProps<{
  skill: SkillMarketEntry
  registry?: SkillMarketRegistry | null
  variant?: 'catalog' | 'home'
}>(), {
  registry: null,
  variant: 'catalog',
})

const categoryIcons = {
  engineering: 'terminal',
  product: 'lightbulb',
  design: 'sparkles',
  knowledge: 'book',
  workflow: 'users',
} as const

const displayName = computed(() => getSkillDisplayName(props.skill))
const displayDescription = computed(() => getSkillDisplayDescription(props.skill))
const categoryName = computed(() => getSkillCategoryName(props.skill.category, props.registry))
const categoryIcon = computed(() => categoryIcons[props.skill.category as keyof typeof categoryIcons] || 'grid')
const riskLevel = computed(() => props.skill.riskLevel || 'low')
const riskLabel = computed(() => getSkillRiskLabel(props.skill.riskLevel))
const visibleTags = computed(() => (props.skill.tags || []).slice(0, 3))
const visibleRuntimes = computed(() => (props.skill.runtime || []).slice(0, 2))
const archiveSize = computed(() => props.skill.archive.size ? `${Math.round(props.skill.archive.size / 1024)} KB` : '—')
const archiveHash = computed(() => {
  const hash = props.skill.archive.sha256
  return hash ? `${hash.slice(0, 6)}…${hash.slice(-4)}` : '—'
})

function runtimeLabel(runtime: string) {
  return ({ codex: 'Codex', claude: 'Claude Code', portable: '通用运行时' } as Record<string, string>)[runtime] || runtime
}
</script>

<style scoped>
.market-skill-card {
  --skill-card-bg: #ffffff;
  --skill-card-border: rgb(16 52 38 / 12%);
  --skill-card-border-hover: rgb(0 168 120 / 34%);
  --skill-card-fg: #111827;
  --skill-card-muted: #5f6f68;
  --skill-card-faint: #8b9a94;
  --skill-card-soft: #f2f6f4;
  --skill-card-runtime: #eaf8f3;
  --skill-card-icon-bg: #ecf9f4;
  --skill-card-icon-fg: #008f66;
  display: flex;
  min-height: 280px;
  flex-direction: column;
  padding: 20px;
  border: 1px solid var(--skill-card-border);
  border-radius: 12px;
  background: var(--skill-card-bg);
  box-shadow: var(--tp-elev-1, 0 2px 8px rgb(16 52 38 / 6%));
  color: var(--skill-card-fg);
  transition: transform 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.market-skill-card:hover {
  transform: translateY(-2px);
  border-color: var(--skill-card-border-hover);
  box-shadow: var(--tp-elev-2, 0 8px 24px rgb(16 52 38 / 10%));
}

.market-skill-card header {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) auto;
  align-items: start;
  gap: 12px;
}

.skill-icon {
  display: grid;
  width: 44px;
  height: 44px;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--skill-card-icon-fg) 18%, transparent);
  border-radius: 11px;
  background: var(--skill-card-icon-bg);
  color: var(--skill-card-icon-fg);
}

.skill-heading { min-width: 0; padding-top: 2px; }
.skill-heading strong { display: block; overflow: hidden; font-size: 16px; font-weight: 700; letter-spacing: 0; text-overflow: ellipsis; white-space: nowrap; }
.skill-heading small { display: block; overflow: hidden; margin-top: 5px; color: var(--skill-card-faint); font-size: 12px; line-height: 1.35; text-overflow: ellipsis; white-space: nowrap; }

.market-skill-card header > em {
  padding: 5px 9px;
  border-radius: 999px;
  font-size: 11px;
  font-style: normal;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
}

.risk-low { background: #dcfce7; color: #166534; }
.risk-medium { background: #fef3c7; color: #92400e; }
.risk-high { background: #fee2e2; color: #991b1b; }

.market-skill-card > p {
  display: -webkit-box;
  min-height: 72px;
  margin: 18px 0 0;
  overflow: hidden;
  color: var(--skill-card-muted);
  font-size: 14px;
  line-height: 1.72;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.skill-chips { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 16px; }
.skill-chips span { min-height: 25px; padding: 5px 9px; border-radius: 999px; background: var(--skill-card-soft); color: var(--skill-card-muted); font-size: 11px; font-weight: 500; line-height: 15px; }
.skill-chips .runtime-chip { background: var(--skill-card-runtime); color: var(--skill-card-icon-fg); }

.market-skill-card footer { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-top: auto; padding-top: 17px; border-top: 1px solid var(--skill-card-border); }
.archive-meta { display: flex; min-width: 0; align-items: center; gap: 7px; overflow: hidden; color: var(--skill-card-faint); font-size: 11px; }
.archive-meta i { font-style: normal; }
.archive-meta code { overflow: hidden; font: 500 11px/1.4 'JetBrains Mono', ui-monospace, monospace; text-overflow: ellipsis; white-space: nowrap; }
.market-skill-card footer :deep(.skill-card-action) { display: inline-flex; min-height: 38px; flex: 0 0 auto; align-items: center; justify-content: center; gap: 6px; padding: 0 11px; border: 1px solid var(--skill-card-border); border-radius: 9px; background: transparent; color: var(--skill-card-fg); font-size: 12px; font-weight: 600; line-height: 1; }
.market-skill-card footer :deep(.skill-card-action:hover) { border-color: var(--skill-card-border-hover); color: var(--skill-card-icon-fg); }

.variant-home {
  --skill-card-bg: var(--color-card);
  --skill-card-border: var(--color-line);
  --skill-card-border-hover: var(--color-line-strong);
  --skill-card-fg: var(--color-fg);
  --skill-card-muted: var(--color-muted);
  --skill-card-faint: var(--color-faint);
  --skill-card-soft: var(--color-chip);
  --skill-card-runtime: var(--color-primary-soft);
  --skill-card-icon-bg: var(--color-ground-2);
  --skill-card-icon-fg: var(--color-primary);
  min-height: 290px;
  border-radius: var(--radius);
  box-shadow: none;
}

:global(.dark .market-skill-card.variant-catalog) {
  --skill-card-bg: rgb(22 33 29 / 72%);
  --skill-card-border: rgb(120 190 165 / 16%);
  --skill-card-border-hover: rgb(47 212 160 / 42%);
  --skill-card-fg: #f3f7f5;
  --skill-card-muted: #b4c4bd;
  --skill-card-faint: #789087;
  --skill-card-soft: #20302a;
  --skill-card-runtime: #15372d;
  --skill-card-icon-bg: #13251f;
  --skill-card-icon-fg: #4fd6ab;
}

:global(.dark .market-skill-card.variant-catalog .risk-low) { background: rgb(22 101 52 / 32%); color: #86efac; }
:global(.dark .market-skill-card.variant-catalog .risk-medium) { background: rgb(146 64 14 / 34%); color: #fcd34d; }
:global(.dark .market-skill-card.variant-catalog .risk-high) { background: rgb(153 27 27 / 34%); color: #fca5a5; }

@media (max-width: 440px) {
  .market-skill-card { min-height: 270px; padding: 18px; }
  .market-skill-card header { grid-template-columns: 40px minmax(0, 1fr); }
  .skill-icon { width: 40px; height: 40px; }
  .market-skill-card header > em { grid-column: 2; width: max-content; }
  .market-skill-card footer { align-items: flex-start; flex-direction: column; }
  .market-skill-card footer :deep(.skill-card-action) { width: 100%; min-height: 40px; }
}
</style>
