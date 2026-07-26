# Tasks

来源：2026-07-26 前端观感与可用性分析（浏览器实测 + 全库普查，证据行号已逐条复核；三条复核不实的初始发现已剔除，见 design.md「证据方法说明」）。P1/P2 为建议顺序；上游文件改动一律遵循 D9 加性边界。

## 1. 决策关闭（见 design.md Open Questions）

- [ ] 1.1 OQ1 品牌绿终值拍板（过渡期按 D3 用 `#00a878`，不阻塞 2/3 节动手）
- [ ] 1.2 OQ2 ConsolePreview 深色首版效果评审，决定是否加深设计
- [ ] 1.3 OQ3 巨型弹窗拆分 / 管理侧栏 IA 是否另行立项

## 2. P1 主题急救（frontend-theme R1）

- [x] 2.1 `SkillMarketView.vue:6,7,22` 三处补 dark: 变体（根/顶栏/底栏）
- [x] 2.2 `TokenPortHome.vue` 深色适配——实现期复核：该页 `.dark .home-shell` 变量覆盖早已存在且生效，真正的欠账只在 ConsolePreview（见 2.3）；本次将 8 条冗余 `.dark X { color: #7fe0bc }` 手工覆盖删除，改由主题化的 `--brand-deep` 单点解析
- [x] 2.3 `ConsolePreview.vue` 深色变体（D1：局部变量扩展为 20 个 + `.dark .console-preview` 整套重取值，硬编码浅色全部改 var）
- [x] 2.4 `SettingsView.vue` `.settings-tabs-shell` 加性 dark: 补丁 + 深色阴影覆盖（上游文件，D9）
- [x] 2.5 `EmailTemplateEditor.vue` 复核改判：邮件预览 iframe 白底为有意保留（邮件 HTML 假定浅色背景），已加注释，spec 同步记录例外
- [x] 2.6 防回归：`src/tokenport/__tests__/frontendPolishTheme.spec.ts` 12 条字符串断言（令牌单一来源/主题完整性/刻度回归）

## 3. P1 品牌令牌统一（frontend-theme R2/R3/R4）

- [x] 3.1 全局品牌色定义点落地：style.css `:root` 六个 `--tp-brand*` + `.dark` 六个深色取值（过渡值 `#00a878`，OQ1 拍板后单点改）
- [x] 3.2 三处局部 `--brand` 副本改为引用 `var(--tp-brand)`/`var(--tp-brand-deep)`；三个文件内 `rgba(0,168,120,…)` 品牌光晕改 `color-mix(var(--brand))`
- [x] 3.3 `tokenport-console.css` mint 系五变量映射到全局令牌；`.dark` 块的 mint 重定义删除（随全局 .dark 解析）；`#36dcc0/#42e4c8` 按钮渐变顶改 `color-mix(…, white)`，按钮文字改白（新品牌绿更深，原深墨绿文字对比不足）
- [x] 3.4 `SkillMarketCatalog.vue` / `SkillMarketView.vue` emerald→primary（风险徽章语义色 emerald-100/700 保留）；`rounded-2xl`→`rounded-xl`；`text-[26px]`→`text-2xl`、`text-[15px]`→`text-base`
- [x] 3.5 落地页正文字号：导航/按钮/市场卡描述/架构条目/文字链 13→14px，预览提示/页脚 12→13px；ConsolePreview 仿真窗内部 10–13px 为示意性微缩 UI，按装饰豁免；徽章/eyebrow 保持 11–12px
- [ ] 3.6 硬编码 hex 分批令牌化：ConsolePreview/TokenPortHome 已基本令牌化（余下为变量定义值、恒暗横幅与图表系列色）；`AuthLayout.vue` 品牌相关已接令牌，其余表面色值待后续批次；其他文件随触碰收敛

## 4. P2 反馈补齐（frontend-feedback）

- [x] 4.1 盘点 showError-only 的 mutation 路径（588 对 try/catch 全量过一遍，326 处 catch 有 showError，判定 25 处属写操作缺反馈）。已补 9 处开关型写操作的 `showSuccess`：余额通知开关、Ollama 自动刷新、支付渠道启用/退款开关、支付渠道支持类型、套餐上下架、渠道监控启用、错误透传规则启用、渠道状态、账号可调度。**反馈来源是 Pinia store（`stores/app.ts:137/146` 的 `appStore.showSuccess/showError`），不是 composable**——后续搜索按 `appStore.showSuccess` 而非 `showSuccess(`
  - 复核改判：`BackupView` 恢复备份**不缺反馈**——成功提示在它启动的轮询里（`BackupView.vue:501` `restoreSuccess`），初始审计只看了 try 块
  - 待续（本批未动，均为「有行内状态变化但无成功语义」，按 D5 属可豁免的边界情形）：支付渠道拖拽排序、上游计费探测 snapshot 为空分支、通知邮箱启停、批量图片导出、验证码已发送、取消支付订单三处（`PaymentQRDialog` / `StripePaymentInline` / `PaymentStatusPanel`；注意 `PaymentQRCodeView:176` 同为取消订单却走路由跳转，口径不一致需统一）
- [x] 4.2 ConfirmDialog 覆盖盘点：57 个 Modal/Dialog 文件全量扫描，**5 处无确认的破坏性操作已补**（沿用各文件既有写法，均为 `window.confirm` + 新增中英文案）：
  - `GroupRPMOverridesModal.vue` 清空分组全部专属 RPM（服务端立即生效不可撤销）
  - `AccountsView.vue` 重置账号配额（顺带补了缺失的 `showError`，原先只有 `console.error`）
  - `ProfileIdentityBindingsSection.vue` 解绑第三方登录（可能失去唯一登录途径；配套加了「取消则不调接口」的回归用例）
  - `ProfileBalanceNotifyCard.vue` 移除已验证通知邮箱
  - `RiskControlView.vue` 删除风控哈希（其同块的 `clearFlaggedHashes` 早有 confirm，此处缺失属明显不一致）
  - 遗留问题（不阻塞，另行收敛）：全仓三套确认机制并存——ConfirmDialog 40 处、`window.confirm` 27 处、TOTP step-up；且 `AccountsView`/`SettingsView` 内部混用，裸 `confirm(t('common.confirm'))` 的文案不含被删对象名
  - 仍无 `showError` 的破坏性操作（catch 只有 `console.error`）：`AccountsView.vue:2025` 删除账号、`:1494` 批量删除
- [x] 4.3 三个回调页补 spinner（复核改判：它们**有**处理中文案，走 i18n 所以初始 grep 漏判；详见 frontend-feedback spec 的实现期复核修正）
- [x] 4.4 空态收敛：DataTable 两处内置空态与 `EmptyState` 视觉 token 对齐（D6）
- [ ] 4.5 技能卡头像：现为首字母（`skillInitial`），可改按技能语义取图标。来源为 2026-07-22 的 stash `user-ui-enhancements`（已丢弃，其 emoji 映射表针对旧版生物信息学技能目录，对当前 28 个技能仅命中 1 个）。重做时映射 MUST 由 registry 字段驱动（category 或新增 icon 字段），MUST NOT 再硬编码 id→emoji 表

## 5. P2 导航与可达性（frontend-navigation / frontend-a11y）

- [x] 5.1 移动端导航改为精简保留（D4）：≤720px 只隐藏 `.nav-link-optional`（模型与渠道 / Docs），保留 Skill Market、主题切换与登录
- [x] 5.2 技能市场未登录壳头部加主题切换；顺带抽出 `composables/useThemeToggle.ts`，两个门面页共用（全仓原有 5 份复制实现，上游那 3 份按 D9 不动，本次未新增第 6 份）
- [x] 5.3 门控链接加锁形标识与说明。复核发现**回跳本来就通**——路由守卫 `router/index.ts:836` 已写 `?redirect=`，`LoginView.vue:502/536` 已消费，缺的只是点击前的预示
- [x] 5.4 Tab 循环圈定：逻辑抽到 `composables/useFocusTrap.ts` 单点实现（排除 disabled/不可见元素；焦点不在容器内时不介入，避免嵌套弹窗互抢），`BaseDialog` 改为调用它。自建 `fixed inset-0` 的 18 个文件已分类：4 个认证弹窗（`TotpLoginModal` / `TotpStepUpDialog` / `TotpDisableDialog` / `TotpSetupModal`，原先 Esc 与 Tab 圈定**都没有**）已接入；其余为下拉菜单/抽屉/内联面板（`AccountActionMenu`、`AccountGroupsCell`、`AnnouncementBell`、`AppLayout`、`AppSidebar`、`LoginAgreementPrompt` 的 checkbox 分支等），圈定语义不适用，不处理
  - 待续：`BackupView` / `RedeemView` / `SettingsView` / `SubscriptionsView` / `PaymentView` / `AccountTestModal`×2 / `AnnouncementPopup` 里的内联弹窗尚未接入，均为上游文件，随后续触碰逐个补
- [x] 5.5 style.css 全局 `:focus-visible` 兜底规则；`.btn` 焦点环改 `focus-visible:`；`.input` 保留 `focus:border`（点击需可见)但环改 `focus-visible:`
- [x] 5.6 图标按钮 aria-label 本地化（`BaseDialog` 关闭按钮原为硬编码 "Close modal"、登录页密码可见性切换原无可访问名；新增 `common.showPassword/hidePassword` 中英各一）
- [x] 5.7 技能卡「查看详情」移动端 `min-h-[40px]`，桌面端保持 btn-sm 紧凑尺度

## 6. 跨仓与交叉引用（不在本 change 实现）

- [x] 6.1 registry `source` 字段已改正：7 个 `anbeime-*` 条目原指向已转为爬虫仓的 `github.com/anbeime/skill`，现指向本仓对应技能目录（在 state-of-art-skills 的 `market/categories.json` 改，重建 registry 后经 sync-skill-market 同步进 fork）。其余 21 个条目本就指向各自上游的固定 commit，未动
- [ ] 6.2 定制面硬编码中文迁 i18n → `add-tokenport-access-center` tasks 5.6（不在此重复）
- [ ] 6.3 Sub2API/TokenPort/sub2api 品牌写法统一 → 同上 tasks 6.3（不在此重复）

## 7. 递延（OQ3 拍板后另行立项，本 change 不做）

- [ ] 7.1 巨型弹窗向导化拆分：`CreateAccountModal`(6281 行) / `EditAccountModal`(4706 行) / `BulkEditAccountModal`(1997 行)（均上游文件，需评估合并负担）
- [ ] 7.2 管理侧栏信息架构：管理 26 + 个人 13 项同屏约 40 链接的分层收纳
