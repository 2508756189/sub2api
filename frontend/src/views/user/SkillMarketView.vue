<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import {
  DEFAULT_SKILL_MARKET_REGISTRY_URL,
  fetchSkillMarketWithSource,
  type SkillMarketEntry,
  type SkillMarketRegistry,
} from '@/api/skillMarket'

const loading = ref(false)
const error = ref<string | null>(null)
const registry = ref<SkillMarketRegistry | null>(null)
const registrySource = ref(DEFAULT_SKILL_MARKET_REGISTRY_URL)
const query = ref('')
const activeCategory = ref<string>('all')

async function load() {
  loading.value = true
  error.value = null
  try {
    const result = await fetchSkillMarketWithSource()
    registry.value = result.registry
    registrySource.value = result.registryUrl
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  } finally {
    loading.value = false
  }
}

const categoryNames = computed(() => {
  const map: Record<string, string> = { all: '全部' }
  for (const category of registry.value?.categories ?? []) map[category.id] = category.name
  return map
})

const filteredSkills = computed<SkillMarketEntry[]>(() => {
  const q = query.value.trim().toLowerCase()
  return (registry.value?.skills ?? [])
    .filter((item) => activeCategory.value === 'all' || item.category === activeCategory.value)
    .filter((item) => {
      if (!q) return true
      return [item.name, item.description, item.category, ...(item.tags ?? [])].join(' ').toLowerCase().includes(q)
    })
})

function countByCategory(id: string): number {
  if (!registry.value) return 0
  if (id === 'all') return registry.value.skills.length
  return registry.value.skills.filter((item) => item.category === id).length
}

onMounted(load)
</script>

<template>
  <div class="skill-market min-h-screen bg-[#0c0f11] text-[#f4f0e8]">
    <header class="sm-header">
      <router-link to="/home" class="back-link">
        <Icon name="arrowLeft" size="sm" />
        <span>返回首页</span>
      </router-link>
      <router-link :to="$route.fullPath" class="sm-brand">
        <strong>TokenPort</strong>
        <small>Skill Market</small>
      </router-link>
      <span class="sm-count" v-if="registry">{{ registry.skills.length }} skills indexed</span>
    </header>

    <section class="sm-hero">
      <p class="eyebrow">SKILL MARKET</p>
      <h1>技能包目录</h1>
      <p class="hero-copy">
        精选的智能应用能力包。每个技能包含定义、分类、风险等级与校验信息，
        在「使用 API 密钥」弹窗中勾选后，会生成带 SHA256 校验的一键安装脚本。
      </p>

      <div class="sm-controls">
        <input
          v-model="query"
          type="text"
          class="search-input"
          placeholder="搜索技能名称、描述或标签..."
        />
      </div>

      <div class="category-tabs">
        <button
          :class="['cat-tab', { active: activeCategory === 'all' }]"
          @click="activeCategory = 'all'"
        >
          全部 <i>{{ countByCategory('all') }}</i>
        </button>
        <button
          v-for="category in registry?.categories ?? []"
          :key="category.id"
          :class="['cat-tab', { active: activeCategory === category.id }]"
          @click="activeCategory = category.id"
        >
          {{ category.name }} <i>{{ countByCategory(category.id) }}</i>
        </button>
      </div>
    </section>

    <main class="sm-main">
      <div v-if="loading" class="state-box">加载技能目录中…</div>
      <div v-else-if="error" class="state-box error">
        技能目录加载失败：{{ error }}
        <button class="retry-btn" @click="load">重试</button>
      </div>
      <div v-else-if="!filteredSkills.length" class="state-box">没有匹配的技能。</div>

      <div v-else class="skill-grid">
        <article
          v-for="skill in filteredSkills"
          :key="skill.id"
          class="skill-card"
        >
          <div class="card-head">
            <h3>{{ skill.name }}</h3>
            <span :class="['risk-pill', `risk-${skill.riskLevel}`]">{{ skill.riskLevel }}</span>
          </div>
          <p class="card-desc">{{ skill.description }}</p>
          <div class="card-meta">
            <span class="meta-cat">{{ categoryNames[skill.category] || skill.category }}</span>
            <span v-for="tag in (skill.tags ?? []).slice(0, 3)" :key="tag" class="meta-tag">{{ tag }}</span>
          </div>
          <dl class="card-facts">
            <div>
              <dt>版本</dt>
              <dd>{{ skill.version }}</dd>
            </div>
            <div>
              <dt>大小</dt>
              <dd>{{ skill.archive.size ? `${Math.round(skill.archive.size / 1024)} KB` : '—' }}</dd>
            </div>
            <div class="fact-checksum">
              <dt>SHA256</dt>
              <dd><code>{{ skill.archive.sha256.slice(0, 12) }}…</code></dd>
            </div>
          </dl>
        </article>
      </div>
    </main>

    <footer class="sm-footer">
      <span>Registry: <code>{{ registrySource }}</code></span>
      <router-link to="/home">← 返回首页</router-link>
    </footer>
  </div>
</template>

<style scoped>
.skill-market {
  position: relative;
  overflow-x: hidden;
  font-family: "Aptos", "Segoe UI", sans-serif;
  --line: rgba(232, 226, 214, 0.14);
  --muted: #a9b0aa;
  --text: #f4f0e8;
  --green: #2ee58c;
  --cyan: #55d7ff;
  --amber: #ffc35a;
}

.sm-header {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 22px clamp(18px, 5vw, 72px);
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--muted);
  font-size: 13px;
}

.back-link:hover {
  color: var(--text);
}

.sm-brand {
  display: grid;
  gap: 2px;
  text-align: center;
  color: var(--text);
}

.sm-brand strong {
  font-size: 16px;
  line-height: 1;
}

.sm-brand small {
  color: var(--cyan);
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 11px;
}

.sm-count {
  color: var(--muted);
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 12px;
}

.sm-hero {
  position: relative;
  z-index: 1;
  max-width: 980px;
  margin: 0 auto;
  padding: 40px clamp(18px, 5vw, 72px) 28px;
  text-align: center;
}

.eyebrow {
  color: var(--amber);
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 12px;
  font-weight: 800;
  text-transform: uppercase;
}

.sm-hero h1 {
  margin-top: 14px;
  color: var(--text);
  font-size: clamp(36px, 5vw, 56px);
  font-weight: 900;
  letter-spacing: -0.02em;
}

.hero-copy {
  max-width: 640px;
  margin: 18px auto 0;
  color: rgba(244, 240, 232, 0.78);
  font-size: 15px;
  line-height: 1.7;
}

.sm-controls {
  max-width: 520px;
  margin: 28px auto 0;
}

.search-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text);
  font-size: 14px;
}

.search-input::placeholder {
  color: var(--muted);
}

.search-input:focus {
  outline: none;
  border-color: rgba(85, 215, 255, 0.6);
}

.category-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
  margin-top: 22px;
}

.cat-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.04);
  color: var(--muted);
  font-size: 13px;
  cursor: pointer;
  transition: border-color 160ms ease, color 160ms ease;
}

.cat-tab i {
  font-style: normal;
  color: var(--muted);
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 11px;
}

.cat-tab:hover {
  color: var(--text);
  border-color: rgba(85, 215, 255, 0.5);
}

.cat-tab.active {
  background: rgba(46, 229, 140, 0.12);
  border-color: rgba(46, 229, 140, 0.6);
  color: var(--green);
}

.cat-tab.active i {
  color: var(--green);
}

.sm-main {
  position: relative;
  z-index: 1;
  max-width: 1280px;
  margin: 0 auto;
  padding: 16px clamp(18px, 5vw, 72px) 80px;
}

.state-box {
  padding: 60px 20px;
  text-align: center;
  color: var(--muted);
  font-size: 15px;
}

.state-box.error {
  color: #ff6e5d;
}

.retry-btn {
  margin-left: 12px;
  padding: 6px 14px;
  border: 1px solid var(--line);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text);
  font-size: 13px;
  cursor: pointer;
}

.skill-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.skill-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 22px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: rgba(18, 24, 25, 0.82);
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.card-head h3 {
  color: var(--text);
  font-size: 17px;
  font-weight: 800;
}

.risk-pill {
  flex-shrink: 0;
  padding: 3px 9px;
  border-radius: 999px;
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
}

.risk-low {
  background: rgba(46, 229, 140, 0.14);
  color: var(--green);
}

.risk-medium {
  background: rgba(255, 195, 90, 0.14);
  color: var(--amber);
}

.risk-high {
  background: rgba(255, 110, 93, 0.14);
  color: #ff6e5d;
}

.card-desc {
  color: rgba(244, 240, 232, 0.72);
  font-size: 13px;
  line-height: 1.65;
}

.card-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.meta-cat {
  padding: 3px 8px;
  border-radius: 4px;
  background: rgba(85, 215, 255, 0.12);
  color: var(--cyan);
  font-size: 11px;
}

.meta-tag {
  padding: 3px 8px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--muted);
  font-size: 11px;
}

.card-facts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin: 0;
  padding-top: 14px;
  border-top: 1px solid var(--line);
}

.card-facts div {
  display: grid;
  gap: 4px;
}

.card-facts dt {
  color: var(--muted);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.card-facts dd {
  margin: 0;
  color: var(--text);
  font-size: 12px;
  font-family: "Cascadia Mono", Consolas, monospace;
}

.fact-checksum dd code {
  color: var(--cyan);
}

.sm-footer {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  flex-wrap: wrap;
  padding: 24px clamp(18px, 5vw, 72px);
  border-top: 1px solid var(--line);
  color: var(--muted);
  font-size: 12px;
}

.sm-footer code {
  color: var(--cyan);
  font-family: "Cascadia Mono", Consolas, monospace;
  word-break: break-all;
}

.sm-footer a {
  color: var(--muted);
}

.sm-footer a:hover {
  color: var(--text);
}

@media (max-width: 720px) {
  .sm-header {
    padding: 16px;
  }

  .sm-hero {
    padding: 24px 18px 20px;
  }

  .skill-grid {
    grid-template-columns: 1fr;
  }
}
</style>
