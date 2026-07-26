# 设计决策与开放问题

## 已采纳的默认决策

| # | 决策 | 采纳方案 | 落点 |
| --- | --- | --- | --- |
| D1 | 门面页深色策略 | TokenPortHome / ConsolePreview / SkillMarketView 壳做完整深色适配（首页头部自带主题切换，「禁用深色」不成立）；ConsolePreview 首期做变量级替换，不做精细重设计 | frontend-theme R1 |
| D2 | 品牌令牌落点 | 在全局层（style.css `:root` 或 tailwind token）建立唯一品牌色定义；三份局部 `--brand` 副本删除并改为引用；`--tp-*` 中语义重复变量映射到该定义；Tailwind primary 色阶不动 | frontend-theme R2 |
| D3 | 品牌绿过渡值 | 终值待 OQ1 拍板；拍板前单点定义取现状最广的 `#00a878`，避免修复被决策阻塞 | frontend-theme R2 |
| D4 | 移动端首页导航形态 | 精简保留（Skill Market + 主题切换 + 登录/进入控制台），「模型与渠道」/Docs 移动端隐藏；不新增汉堡组件（最小正确方案） | frontend-navigation R1 |
| D5 | 成功反馈豁免口径 | 操作结果在当前视口内有可见状态变化（行内更新/跳转/伴随高亮的列表刷新）可豁免 toast，否则必须 showSuccess | frontend-feedback R1 |
| D6 | 空态收敛方向 | 表格场景用 DataTable 内置空态，非表格场景用 EmptyState，两者视觉 token（图标尺寸/字号/间距/文案键）对齐为一套 | frontend-feedback R4 |
| D7 | 键盘焦点样式 | style.css 增加全局 `:focus-visible` 规则；`.btn`/`.input` 等基类的 `focus:ring` 改为 `focus-visible:` 系 | frontend-a11y R2 |
| D8 | 字号与触控下限 | 落地页正文 ≥14px（11px 仅限 eyebrow/徽标类装饰）；移动视口可点目标高度 ≥40px | frontend-theme R3 / frontend-a11y R4 |
| D9 | 上游文件边界 | 上游视图只做加性修补（补 dark: 类名、插入加载态），禁止结构性重构；fork 自有文件（`src/tokenport/**`、`components/keys/**` 等）可自由重构。判断依据与同步分类见 `scripts/sync/tokenport-sync-policy.json` | 全部 |

## Open Questions（实现前必须逐条确认，未确认项不得自行漂移）

1. **品牌绿终值**：`#00a878`（天翼云绿，登录/首页现用）vs `#2ad4b8`（`--tp-mint`，控制台预览用）vs 收敛到 `primary-500 #14b8a6`（上游 teal）。涉及品牌认知，需用户/品牌方拍板；D3 规定了拍板前的过渡口径。
2. **ConsolePreview 深色的设计深度**：D1 先做变量级替换；若首版观感不达标，是否投入精细重设计（层次/阴影/发光），或改为「深色页面中保留浅色窗口但加边框与投影缓冲」——视首版效果评审再定。
3. **结构性改造是否立项**：巨型弹窗拆分（CreateAccountModal 6281 行 / EditAccountModal 4706 行 / BulkEditAccountModal 1997 行）与管理侧栏信息架构（管理 26 + 个人 13 项同屏约 40 链接）不在本 change 范围；若立项应各自另开 change（涉及交互重构与上游合并负担评估，三个巨型弹窗均为上游文件）。

## 证据方法说明

- **「实测」**= 2026-07-26 对 localhost:8080 运行实例的浏览器实测（浅色/深色主题 × 1280/375 视口，读取 DOM 结构与计算样式取证），非猜测。
- **「普查计数」**= 对 frontend/src 289 个 .vue 的静态 grep 计数。
- 初始普查中三条发现经逐行复核**不实，已剔除**，不得再被引用：`RiskControlView.vue:535` 与 `OpsDashboardHeader.vue:1001` 的深色 hover 实际齐全（均有 `dark:hover:bg-dark-800` 系）；`<img>` 32 处与 `alt=` 32 处持平，不存在「12 处图片缺 alt」。
- 测试环境提示：浏览器窗格未显示时 rAF 挂起会让 Vue 过渡卡住，曾误判「弹窗关不掉/Esc 无效」——BaseDialog 的 Esc 与焦点还原代码齐全（`BaseDialog.vue:110-114,126-141`），复核结论以代码为准。
