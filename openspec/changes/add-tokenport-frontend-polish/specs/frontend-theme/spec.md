# frontend-theme

## ADDED Requirements

### Requirement: 提供主题切换的页面必须两套主题完整
凡可从页面自身或全局入口切换明暗主题的页面，其全部视觉层（根背景、顶栏、底栏、内容卡、粘性元素）SHALL 同时具备浅色与深色样式；MUST NOT 出现「容器保持浅色、子组件激活 dark:」的混合态。已证实的破洞 MUST 修复：

- `frontend/src/views/user/SkillMarketView.vue:6,7,22`——未登录壳根 `bg-[#f4f8f6]`、顶栏 `bg-white/90`、底栏 `bg-white` 均无 dark: 变体（实测深色下为浅底、深色卡片、白顶底栏混合态）；
- `frontend/src/tokenport/home/TokenPortHome.vue`——1226 行内 `dark:` 出现 0 次，仅 3 条 `.dark` 兜底选择器；
- `frontend/src/tokenport/home/ConsolePreview.vue`——深色页面中央 616×804 纯白窗口（实测）；
- `frontend/src/views/admin/SettingsView.vue:11694-11699`——`.settings-tabs-shell` 粘性 Tab 条 `border-white/80 bg-white/90` 与浅色阴影无 dark: 对应（上游文件，按 D9 只做加性补丁）。

实现期复核例外（不计入破洞）：`frontend/src/views/admin/settings/EmailTemplateEditor.vue` 预览 iframe 的 `bg-white` 为有意保留——邮件 HTML 默认按浅色背景设计，深底会让透明背景的邮件不可读；已在代码处加注释说明。此类「内容本身假定浅色」的容器（邮件预览等）允许在深色主题下保持白底，但 MUST 有注释记录理由。

#### Scenario: 深色模式浏览技能市场未登录页
- **WHEN** 用户在深色主题下访问 /skill-market（未登录）
- **THEN** 页面根/顶栏/底栏 MUST 呈现深色样式，与 hero 带、卡片一致，MUST NOT 出现白色顶底栏夹深色内容

#### Scenario: 深色模式下查看首页控制台预览
- **WHEN** 用户在首页切换到深色主题
- **THEN** 控制台预览组件 MUST 呈现深色变体（或按 OQ2 拍板的缓冲方案），MUST NOT 保持整块纯白

### Requirement: 品牌色令牌必须单一来源
品牌色 SHALL 只有一个全局定义点。`--brand: #00a878` 的三份局部副本（`frontend/src/components/layout/AuthLayout.vue:136`、`frontend/src/tokenport/home/ConsolePreview.vue:305`、`frontend/src/tokenport/home/TokenPortHome.vue:405`）MUST 删除并改为引用全局定义；`frontend/src/tokenport/brand/tokenport-console.css` 中与品牌色语义重复的变量（`--tp-mint: #2ad4b8` 等）MUST 映射到全局定义。终值按 design.md OQ1，拍板前取 `#00a878`（D3）。新增代码 MUST NOT 引入新的品牌色硬编码 hex。

#### Scenario: 拍板后调整品牌绿
- **WHEN** OQ1 确定品牌绿终值并修改全局定义点
- **THEN** 登录页、首页、控制台预览、技能市场的品牌色 MUST 随单点修改同步变化，MUST NOT 需要逐文件替换

### Requirement: 定制面必须回归全站设计刻度
TokenPort 定制面（`src/tokenport/**`、`SkillMarketView.vue`、`components/keys/**`）的圆角与字号 SHALL 使用全站刻度（`rounded-xl`、Tailwind 字号刻度）；任意值（`text-[26px]`、`text-[15px]` 等）MUST 收敛，保留的例外 MUST 在代码旁注明理由。落地页正文字号 MUST ≥14px，11px 及以下仅限 eyebrow/徽标类装饰文本（D8；实测现状正文以 12–13px 为主共 87 处、11px 12 处）。

#### Scenario: 定制面与主控制台并排对比
- **WHEN** 用户从技能市场进入控制台（或反向）
- **THEN** 两侧的圆角尺度、正文字号、主色 MUST 呈现同一产品的观感，MUST NOT 出现 emerald 与 primary 两套主色并存

### Requirement: 硬编码色值治理
新增或修改的组件 MUST 使用设计令牌（Tailwind token 或全局 CSS 变量）而非硬编码 hex；存量 391 处（36 个文件，普查计数）SHALL 按文件分批收敛，优先 `AuthLayout.vue`（44 处）、`ConsolePreview.vue`（43 处）、`TokenPortHome.vue`（37 处）。

#### Scenario: 评审含新硬编码 hex 的前端提交
- **WHEN** 代码评审发现新增 `#xxxxxx` 色值且无例外说明
- **THEN** 该提交 MUST 被要求改用令牌后再合入
