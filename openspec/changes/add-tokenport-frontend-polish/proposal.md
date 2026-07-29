# TokenPort 前端观感与可用性打磨

## Why

2026-07-26 对运行实例的浏览器实测（localhost:8080 三个公开页 × 浅色/深色主题 × 1280/375 视口，DOM 与计算样式取证）与全库静态普查（289 个 .vue、约 12.9 万行）的联合分析结论：全站基础是健康的——`dark:` 覆盖 261/289 文件、DataTable/BaseDialog/Toast 共享组件体系、双语 locale 齐全且有测试、公开页无 console 错误、无横向溢出、表单 label/autocomplete 规范。但存在两类系统性问题：

1. **TokenPort 定制面（首页/技能市场/登录页，恰是第一印象页面）游离在全站规范之外**：
   - 深色模式在门面页面破洞：技能市场未登录壳三处容器无 `dark:` 变体，深色下实测呈「浅色画布 + 深色卡片 + 白色顶底栏」混合态；`TokenPortHome.vue`（1226 行）`dark:` 出现 0 次；「控制台交互预览」在深色页面中央是一块 616×804 纯白窗口（实测）。
   - 一个产品三个「品牌绿」：`--brand: #00a878` 在三个文件各自复制一份局部定义、`--tp-mint: #2ad4b8`（tokenport-console.css 约 60 个 `--tp-*` 变量）、Tailwind `primary-500: #14b8a6`，三者互不引用；定制面用 emerald 色系 + `rounded-2xl` + `text-[26px]` 任意值，与主站 primary + `rounded-xl` + 标准刻度并存；36 个文件散布 391 处硬编码 hex。
   - 移动端（≤720px）首页导航把 Skill Market、模型与渠道、Docs、主题切换全部 `display:none` 且无替代入口，顶栏只剩「登录平台」（375px 实测）。

2. **跨页面的反馈与可达性缺口**：
   - 反馈不对称：普查计数 `showError` 556 次 vs `showSuccess` 247 次，大量写操作成功后静默；57 个 Modal/Dialog 文件仅 27 个引用 ConfirmDialog。
   - 共享加载/空态原语采用率极低（Skeleton 2 处、LoadingSpinner 3 处引用）；微信/OIDC/钉钉登录回调页（1102/856/852 行）未检出任何加载指示，而 OAuthCallbackView 有 spinner——同族页面不一致；DataTable 内置空态与 EmptyState 组件两套视觉并存。
   - 可达性：BaseDialog 有 Esc 关闭与焦点还原但无 Tab 循环圈定；全站基类用 `focus:ring`（鼠标点击也触发），style.css 无 `:focus-visible` 规则；纯图标按钮存在无可访问名（密码可见性切换）或英文 aria（"Close modal"）；移动端卡片按钮触控高度 30px。

证据行号已逐条复核，三条初始发现复核不实已剔除（见 design.md「证据方法说明」）。与 `add-tokenport-access-center` 的关系：该 change 管交付语义与外部契约，本 change 管视觉一致性与交互反馈；两处已有条目（定制面文案迁 i18n、品牌写法统一）不在此重复，只交叉引用。

## What Changes

- 定义「主题完整性」要求：凡提供主题切换入口的页面，其全部视觉层必须同时具备明暗两套样式；修复已证实的 5 处破洞（技能市场壳、TokenPortHome、ConsolePreview、SettingsView 粘性 Tab 条、EmailTemplateEditor 预览容器）。
- 品牌色令牌单一来源：删除三份局部 `--brand` 副本，`--tp-*` 与品牌色语义重复的变量映射到共享定义；定制面圆角/字号回归全站刻度；落地页正文字号下限 14px；新增代码禁止新的硬编码 hex。
- 移动端公开页导航保底：≤720px 保留 Skill Market 入口与主题切换；公开页头部结构统一；登录门控链接带提示与回跳。
- 写操作成功反馈准则、破坏性操作确认覆盖盘点、登录回调页加载态、空态视觉收敛为一套。
- BaseDialog 焦点圈定、全局 `:focus-visible` 规则、图标按钮本地化可访问名、移动端触控目标 ≥40px。

## Capabilities

### New Capabilities

- `frontend-theme`：主题完整性、品牌令牌单一来源、设计刻度回归、硬编码色值治理。
- `frontend-feedback`：mutation 成功反馈、破坏性操作确认、等待页加载态、空态统一。
- `frontend-navigation`：移动端公开页导航保底、公开页头部一致性、登录门控提示与回跳。
- `frontend-a11y`：弹窗焦点圈定、全局键盘焦点样式、图标按钮可访问名、触控目标下限。

### Modified Capabilities

无。`client-access-center` R6（命名与多语言）与本 change 的品牌令牌治理相邻但不重叠：R6 管文案写法，本 change 管视觉令牌。

## Impact

- **fork 自有/深度定制文件（可自由重构）**：`src/tokenport/**`（TokenPortHome、ConsolePreview、SkillMarketCatalog、tokenport-console.css）、`src/views/user/SkillMarketView.vue`、`src/components/keys/**`、`src/components/layout/AuthLayout.vue`、`src/style.css`、`src/components/common/BaseDialog.vue`。
- **上游文件（最小加性修补）**：`SettingsView.vue`（settings-tabs-shell 一处 dark: 补丁）、`EmailTemplateEditor.vue:214`（补一个 dark:bg）、各登录回调视图（加载态为加性插入）。本 change 明确禁止对上游视图做结构性重构，避免加重上游合并负担（同步策略见 `scripts/sync/tokenport-sync-policy.json`）。
- **跨仓**：registry `source` 字段改指 state-of-art-skills 固定 commit（在 skills 仓生成端修，前端无改动）。
- **不改变**：业务逻辑、路由结构、后端、上游交付能力、`add-tokenport-access-center` 已定契约。
