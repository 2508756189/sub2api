<template>
  <div class="console-preview" aria-label="TokenPort 经营控制台交互预览">
    <div class="window-bar">
      <span class="dot red" /><span class="dot yellow" /><span class="dot green" />
      <b>TokenPort 经营控制台</b>
      <em>功能示意 · 非当前账号数据</em>
    </div>

    <div class="console-body">
      <aside class="side">
        <div class="brand">
          <img :src="logoSrc" alt="天翼云" />
          <div>
            <strong>TokenPort</strong>
            <small>天翼云智能接入</small>
          </div>
        </div>

        <div class="side-section">管理菜单</div>
        <button
          v-for="item in primaryMenus"
          :key="item.id"
          type="button"
          class="side-item"
          :class="{ active: activeId === item.id }"
          @click="select(item.id)"
        >
          <span class="icon" v-html="item.icon" />
          <span>{{ item.label }}</span>
          <i v-if="item.badge">{{ item.badge }}</i>
        </button>

        <div class="side-section">我的账户</div>
        <button
          v-for="item in accountMenus"
          :key="item.id"
          type="button"
          class="side-item"
          :class="{ active: activeId === item.id }"
          @click="select(item.id)"
        >
          <span class="icon" v-html="item.icon" />
          <span>{{ item.label }}</span>
        </button>
      </aside>

      <section class="main">
        <header class="main-top">
          <div>
            <p class="crumb">管理控制台 / {{ activeMenu.label }}</p>
            <h3>{{ activeMenu.title }}</h3>
          </div>
          <div class="top-meta">
            <span>CN · ZH</span>
            <span class="balance">示例环境</span>
            <span class="admin-chip">管理员</span>
          </div>
        </header>

        <div class="panel">
          <template v-if="activeId === 'dashboard'">
            <div class="kpi-grid">
              <article v-for="kpi in kpis" :key="kpi.label" class="kpi">
                <span>{{ kpi.label }}</span>
                <b>{{ kpi.value }}</b>
                <small>{{ kpi.hint }}</small>
              </article>
            </div>
            <div class="split">
              <div class="card">
                <div class="card-head"><b>模型分布</b><em>近 24 小时</em></div>
                <div class="donut-wrap">
                  <div class="donut" />
                  <ul>
                    <li v-for="row in models" :key="row.name"><i :style="{ background: row.color }" /><span>{{ row.name }}</span><b>{{ row.share }}</b></li>
                  </ul>
                </div>
              </div>
              <div class="card">
                <div class="card-head"><b>Token 使用趋势</b><em>Input / Output</em></div>
                <div class="chart">
                  <svg viewBox="0 0 320 120" preserveAspectRatio="none">
                    <path d="M0 90 C40 88, 60 70, 90 72 S140 40, 170 48 S230 20, 260 28 S300 50, 320 42" fill="none" stroke="#00a878" stroke-width="3" />
                    <path d="M0 100 C50 98, 80 92, 110 88 S170 80, 210 78 S280 70, 320 68" fill="none" stroke="#5bb8ff" stroke-width="2.5" opacity="0.8" />
                  </svg>
                </div>
              </div>
            </div>
          </template>

          <template v-else-if="activeId === 'users'">
            <div class="table-card">
              <div class="card-head"><b>用户列表</b><em>{{ activeMenu.hint }}</em></div>
              <table>
                <thead><tr><th>用户</th><th>角色</th><th>状态</th><th>本月 Token</th></tr></thead>
                <tbody>
                  <tr v-for="row in users" :key="row.name"><td>{{ row.name }}</td><td>{{ row.role }}</td><td><em class="ok">{{ row.status }}</em></td><td>{{ row.token }}</td></tr>
                </tbody>
              </table>
            </div>
          </template>

          <template v-else-if="activeId === 'accounts'">
            <div class="grid-2">
              <article v-for="acc in accounts" :key="acc.name" class="card account">
                <div class="card-head"><b>{{ acc.name }}</b><em :class="acc.statusClass">{{ acc.status }}</em></div>
                <p>{{ acc.desc }}</p>
                <div class="tags"><span v-for="tag in acc.tags" :key="tag">{{ tag }}</span></div>
              </article>
            </div>
          </template>

          <template v-else-if="activeId === 'channels'">
            <div class="grid-3">
              <article v-for="ch in channels" :key="ch.name" class="card">
                <div class="card-head"><b>{{ ch.name }}</b><em>{{ ch.latency }}</em></div>
                <div class="meter"><i :style="{ width: ch.load }" /></div>
                <small>{{ ch.desc }}</small>
              </article>
            </div>
          </template>

          <template v-else-if="activeId === 'usage'">
            <div class="card">
              <div class="card-head"><b>使用记录</b><em>按部门 / 模型归集</em></div>
              <table>
                <thead><tr><th>时间</th><th>部门</th><th>模型</th><th>Token</th><th>成本</th></tr></thead>
                <tbody>
                  <tr v-for="row in usage" :key="row.time + row.model"><td>{{ row.time }}</td><td>{{ row.dept }}</td><td>{{ row.model }}</td><td>{{ row.token }}</td><td>{{ row.cost }}</td></tr>
                </tbody>
              </table>
            </div>
          </template>

          <template v-else-if="activeId === 'skills'">
            <div class="grid-2">
              <article v-for="skill in skills" :key="skill.name" class="card">
                <div class="card-head"><b>{{ skill.name }}</b><em>{{ skill.cat }}</em></div>
                <p>{{ skill.desc }}</p>
              </article>
            </div>
          </template>

          <template v-else-if="activeId === 'keys'">
            <div class="card">
              <div class="card-head"><b>API 密钥</b><em>一键生成客户端配置</em></div>
              <div class="key-list">
                <div v-for="key in keys" :key="key.name" class="key-row">
                  <div><b>{{ key.name }}</b><small>{{ key.mask }}</small></div>
                  <span>{{ key.scope }}</span>
                </div>
              </div>
            </div>
          </template>

          <template v-else>
            <div class="card empty-card">
              <div class="card-head"><b>{{ activeMenu.title }}</b><em>{{ activeMenu.hint }}</em></div>
              <ul class="bullet-list">
                <li v-for="point in activeMenu.points" :key="point">{{ point }}</li>
              </ul>
            </div>
          </template>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const props = withDefaults(defineProps<{ logoSrc?: string }>(), {
  logoSrc: '/ctyun-logo.svg',
})

const logoSrc = computed(() => props.logoSrc || '/ctyun-logo.svg')

type MenuId =
  | 'dashboard'
  | 'ops'
  | 'users'
  | 'groups'
  | 'channels'
  | 'accounts'
  | 'usage'
  | 'settings'
  | 'skills'
  | 'keys'

interface MenuItem {
  id: MenuId
  label: string
  title: string
  hint: string
  icon: string
  badge?: string
  points: string[]
}

const icon = {
  dash: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 4h7v7H4V4zm9 0h7v5h-7V4zM4 13h7v7H4v-7zm9 3h7v4h-7v-4z"/></svg>',
  ops: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 19V5m5 14V9m5 10V7m5 12V3"/></svg>',
  users: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="3"/><path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/></svg>',
  folder: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"/></svg>',
  channel: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 7h16M4 12h10M4 17h13"/></svg>',
  globe: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18"/></svg>',
  chart: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 19V5m0 14h16M8 17V9m4 8v-5m4 5V7"/></svg>',
  cog: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="3"/><path d="M12 3v2m0 14v2M4.9 4.9l1.4 1.4m11.4 11.4 1.4 1.4M3 12h2m14 0h2M4.9 19.1l1.4-1.4m11.4-11.4 1.4-1.4"/></svg>',
  skill: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="m12 3 2.5 6.5L21 12l-6.5 2.5L12 21l-2.5-6.5L3 12l6.5-2.5L12 3z"/></svg>',
  key: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M15 7a4 4 0 1 1-4 4m0 0 8 8m-3-3 2 2"/></svg>',
}

const primaryMenus: MenuItem[] = [
  { id: 'dashboard', label: '仪表盘', title: '经营总览', hint: '系统概览与统计数据', icon: icon.dash, points: [] },
  { id: 'ops', label: '运维监控', title: '运维监控', hint: '延迟、错误率与通道健康', icon: icon.ops, points: ['实时观察上游可用性', '错误率与超时阈值告警', '按渠道定位异常波动'] },
  { id: 'users', label: '用户管理', title: '用户管理', hint: '部门账号与权限边界', icon: icon.users, points: [] },
  { id: 'groups', label: '分组管理', title: '分组管理', hint: '按业务线组织资源', icon: icon.folder, points: ['按研发 / 产品 / 客服分组', '分组级预算与模型策略', '支持差异化限流配置'] },
  { id: 'channels', label: '渠道管理', title: '渠道管理', hint: '模型供给与路由策略', icon: icon.channel, badge: '3', points: [] },
  { id: 'accounts', label: '账号管理', title: '上游账号', hint: '供应商密钥与配额', icon: icon.globe, points: [] },
  { id: 'usage', label: '使用记录', title: '使用记录', hint: 'Token 与成本明细', icon: icon.chart, points: [] },
  { id: 'settings', label: '系统设置', title: '系统设置', hint: '品牌、安全与部署参数', icon: icon.cog, points: ['站点品牌与登录策略', '安全白名单与审计开关', '私有化部署参数管理'] },
]

const accountMenus: MenuItem[] = [
  { id: 'skills', label: 'Skill Market', title: '能力市场', hint: '可复用智能体能力', icon: icon.skill, points: [] },
  { id: 'keys', label: 'API 密钥', title: 'API 密钥', hint: '接入配置与密钥治理', icon: icon.key, points: [] },
]

const activeId = ref<MenuId>('dashboard')

const activeMenu = computed(() => {
  return [...primaryMenus, ...accountMenus].find((item) => item.id === activeId.value) || primaryMenus[0]
})

function select(id: MenuId) {
  activeId.value = id
}

const kpis = [
  { label: 'API 密钥', value: '--', hint: '登录后查看' },
  { label: '账号', value: '--', hint: '登录后查看' },
  { label: '今日请求', value: '--', hint: '实时统计' },
  { label: '用户', value: '--', hint: '按部门归集' },
  { label: '今日 Token', value: '--', hint: '实时核算' },
  { label: '总 Token', value: '--', hint: '按周期汇总' },
  { label: '性能指标', value: '--', hint: 'RPM / TPM' },
  { label: '平均响应', value: '--', hint: '调用质量' },
]

const models = [
  { name: '高性能模型', share: '示例', color: '#00a878' },
  { name: '快速模型', share: '示例', color: '#5bb8ff' },
  { name: '多模态模型', share: '示例', color: '#f5b942' },
  { name: '其他模型', share: '示例', color: '#c5d4cd' },
]

const users = [
  { name: '研发部门', role: '开发者', status: '示例', token: '--' },
  { name: '产品部门', role: '业务用户', status: '示例', token: '--' },
  { name: '服务部门', role: '业务用户', status: '示例', token: '--' },
  { name: '平台管理员', role: '管理员', status: '示例', token: '--' },
]

const accounts = [
  { name: '高性能模型池', status: '示例', statusClass: 'ok', desc: '承载复杂推理、编码和智能体任务。', tags: ['模型路由', '高可用'] },
  { name: '通用兼容池', status: '示例', statusClass: 'ok', desc: '兼容常用 API 协议与工具调用场景。', tags: ['兼容协议', '工具调用'] },
  { name: '多模态通道', status: '示例', statusClass: 'warn', desc: '承载图片、文档和批量处理任务。', tags: ['多模态', '预算控制'] },
  { name: '自有模型接入', status: '示例', statusClass: 'ok', desc: '统一纳管企业内部模型与私有化推理资源。', tags: ['私有化', '内网'] },
]

const channels = [
  { name: '高性能通道', latency: '--', load: '68%', desc: '复杂推理与编码任务示意' },
  { name: '通用兼容通道', latency: '--', load: '54%', desc: '工具调用与通用对话示意' },
  { name: '多模态通道', latency: '--', load: '36%', desc: '图片与长文档任务示意' },
]

const usage = [
  { time: '--:--', dept: '研发部门', model: '高性能模型', token: '--', cost: '--' },
  { time: '--:--', dept: '产品部门', model: '多模态模型', token: '--', cost: '--' },
  { time: '--:--', dept: '服务部门', model: '快速模型', token: '--', cost: '--' },
  { time: '--:--', dept: '平台运维', model: '兼容模型', token: '--', cost: '--' },
]

const skills = [
  { name: '多智能体团队框架', cat: '组织协作', desc: '组织多个角色型智能体协作完成评审与交付。' },
  { name: '前端体验设计', cat: '设计体验', desc: '提升页面质感、信息层级与交互细节。' },
  { name: '内容研究写作助手', cat: '产品与研究', desc: '调研写作、提纲迭代与方案材料生产。' },
  { name: '文档转 Markdown', cat: '知识与记忆', desc: '将 Office / PDF 等资料转为可处理 Markdown。' },
]

const keys = [
  { name: '研发共用 Key', mask: 'sk-****-example', scope: '代码模型池' },
  { name: '服务部门 Key', mask: 'sk-****-example', scope: '快速模型池' },
  { name: '示例环境 Key', mask: 'sk-****-example', scope: '只读预算' },
]
</script>

<style scoped>
.console-preview {
  --ink: #13251f;
  --muted: #62756c;
  --line: #d7e5de;
  --soft: #f3faf6;
  --brand: #00a878;
  --ctyun: #0077e6;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--line) 80%, #8fbaaa);
  border-radius: 18px;
  background: #fff;
  box-shadow:
    0 20px 40px rgba(16, 52, 38, 0.1),
    0 1px 0 rgba(255, 255, 255, 0.85) inset;
  color: var(--ink);
  font-family: "Noto Sans SC", "PingFang SC", "HarmonyOS Sans SC", "Segoe UI", sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: geometricPrecision;
  transform: none;
  filter: none;
  user-select: none;
}

.window-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 14px;
  border-bottom: 1px solid #dce7e1;
  background: linear-gradient(180deg, #f8fbfa, #eef5f1);
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #c5d4cd;
}

.dot.red { background: #ff796f; }
.dot.yellow { background: #ffc85c; }
.dot.green { background: #45c98c; }

.window-bar b {
  margin-left: 6px;
  color: #4f635a;
  font-size: 12px;
  font-weight: 700;
}

.window-bar em {
  margin-left: auto;
  color: #7d9188;
  font-size: 11px;
  font-style: normal;
}

.console-body {
  display: grid;
  grid-template-columns: 210px 1fr;
  min-height: 480px;
  background: #f7faf8;
}

.side {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 14px 12px;
  border-right: 1px solid var(--line);
  background: #ffffff;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  padding: 4px 6px 12px;
  border-bottom: 1px solid var(--line);
}

.brand img {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  box-shadow: 0 8px 18px rgba(0, 102, 204, 0.2);
}

.brand strong {
  display: block;
  font-size: 13px;
  letter-spacing: -0.02em;
}

.brand small {
  color: var(--muted);
  font-size: 10px;
}

.side-section {
  margin: 10px 8px 6px;
  color: #8aa197;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.side-item {
  display: flex;
  align-items: center;
  gap: 9px;
  width: 100%;
  min-height: 36px;
  padding: 0 11px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #2f453d;
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.side-item:hover {
  background: rgba(0, 168, 120, 0.08);
  color: var(--ink);
}

.side-item.active {
  background: linear-gradient(180deg, rgba(0, 168, 120, 0.16), rgba(0, 168, 120, 0.1));
  color: #066b4c;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px rgba(0, 168, 120, 0.16);
}

.side-item .icon {
  display: grid;
  width: 16px;
  height: 16px;
  place-items: center;
  flex: 0 0 16px;
}

.side-item .icon :deep(svg) {
  width: 15px;
  height: 15px;
}

.side-item i {
  margin-left: auto;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  background: rgba(0, 119, 230, 0.12);
  color: var(--ctyun);
  font-size: 10px;
  font-style: normal;
  font-weight: 700;
  line-height: 18px;
  text-align: center;
}

.main {
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding: 16px 18px 18px;
  background: #f4f8f6;
}

.main-top {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.crumb {
  color: var(--muted);
  font-size: 11px;
}

.main-top h3 {
  margin-top: 4px;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.top-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.top-meta span {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 10px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: #fff;
  color: var(--muted);
  font-size: 11px;
  font-weight: 600;
}

.balance {
  color: #0a7a58 !important;
}

.admin-chip {
  background: #102a21 !important;
  border-color: #102a21 !important;
  color: #fff !important;
}

.panel {
  display: grid;
  gap: 12px;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.kpi,
.card,
.table-card {
  border: 1px solid var(--line);
  border-radius: 14px;
  background: #ffffff;
  box-shadow: 0 6px 16px rgba(16, 52, 38, 0.04);
}

.kpi {
  padding: 12px;
}

.kpi span,
.card small,
.account p,
.key-row small {
  color: var(--muted);
  font-size: 11px;
}

.kpi b {
  display: block;
  margin-top: 6px;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.kpi small {
  display: block;
  margin-top: 4px;
}

.split {
  display: grid;
  grid-template-columns: 1.05fr 1fr;
  gap: 10px;
}

.grid-2 {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.grid-3 {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.card,
.table-card {
  padding: 14px;
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 12px;
}

.card-head b {
  font-size: 13px;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.card-head em,
.ok,
.warn {
  flex-shrink: 0;
  font-size: 11px;
  font-style: normal;
  font-weight: 700;
  white-space: nowrap;
}

.card-head em,
.ok {
  color: #0a7a58;
}

.warn {
  color: #c27b12;
}

.donut-wrap {
  display: grid;
  grid-template-columns: 96px 1fr;
  gap: 14px;
  align-items: center;
}

.donut {
  width: 96px;
  height: 96px;
  border-radius: 50%;
  background:
    conic-gradient(#00a878 0 41%, #5bb8ff 41% 69%, #f5b942 69% 87%, #c5d4cd 87% 100%);
  mask: radial-gradient(circle at center, transparent 46%, #000 47%);
}

.donut-wrap ul,
.bullet-list,
.key-list {
  display: grid;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.donut-wrap li,
.key-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.donut-wrap li i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.donut-wrap li span {
  flex: 1;
  color: var(--muted);
}

.chart {
  height: 120px;
  border-radius: 12px;
  background:
    linear-gradient(180deg, rgba(0, 168, 120, 0.05), transparent),
    repeating-linear-gradient(0deg, transparent, transparent 23px, rgba(16, 52, 38, 0.05) 24px);
}

.chart svg {
  width: 100%;
  height: 100%;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

th,
td {
  padding: 10px 8px;
  border-bottom: 1px solid #e7f0eb;
  text-align: left;
}

th {
  color: var(--muted);
  font-size: 11px;
  font-weight: 700;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 12px;
}

.tags span {
  min-height: 24px;
  padding: 0 8px;
  border-radius: 999px;
  background: var(--soft);
  color: var(--muted);
  font-size: 11px;
  line-height: 24px;
}

.meter {
  height: 8px;
  margin: 10px 0;
  overflow: hidden;
  border-radius: 999px;
  background: #e7f1ec;
}

.meter i {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #12b884, #0077e6);
}

.account p {
  margin: 0;
  line-height: 1.6;
}

.key-row {
  justify-content: space-between;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: #fbfdfc;
}

.key-row b {
  display: block;
  font-size: 12px;
}

.key-row span {
  color: var(--muted);
  font-size: 11px;
  font-weight: 700;
}

.bullet-list li {
  position: relative;
  padding-left: 14px;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.7;
}

.bullet-list li::before {
  content: "";
  position: absolute;
  top: 0.62em;
  left: 0;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--brand);
}

.empty-card {
  min-height: 220px;
}

@media (max-width: 900px) {
  .console-body {
    grid-template-columns: 1fr;
  }

  .side {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 6px;
    border-right: 0;
    border-bottom: 1px solid var(--line);
  }

  .brand,
  .side-section {
    grid-column: 1 / -1;
  }

  .kpi-grid,
  .split,
  .grid-2,
  .grid-3 {
    grid-template-columns: 1fr;
  }
}
</style>
