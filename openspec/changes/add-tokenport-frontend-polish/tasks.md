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

- [ ] 4.1 盘点 showError-only 的 mutation 路径，按 D5 补 `showSuccess` 或记录豁免
- [ ] 4.2 ConfirmDialog 覆盖盘点：57 个 Modal/Dialog 文件逐一确认破坏性操作是否有确认，无则补
- [ ] 4.3 `WechatCallbackView` / `OidcCallbackView` / `DingTalkCallbackView` 加载态补齐（与 `OAuthCallbackView` 的 spinner 对齐；上游文件按 D9 加性插入）
- [ ] 4.4 空态收敛：DataTable 内置空态与 `EmptyState` 视觉 token 对齐（D6）
- [ ] 4.5 技能卡头像：现为首字母（`skillInitial`），可改按技能语义取图标。来源为 2026-07-22 的 stash `user-ui-enhancements`（已丢弃，其 emoji 映射表针对旧版生物信息学技能目录，对当前 28 个技能仅命中 1 个）。重做时映射 MUST 由 registry 字段驱动（category 或新增 icon 字段），MUST NOT 再硬编码 id→emoji 表

## 5. P2 导航与可达性（frontend-navigation / frontend-a11y）

- [ ] 5.1 `TokenPortHome.vue:1189-1193` 移动端导航改为精简保留（D4：市场 + 主题切换 + 登录）
- [ ] 5.2 技能市场未登录壳头部加主题切换，与首页头部结构对齐
- [ ] 5.3 「模型与渠道」等门控链接加提示，登录后回跳原目标
- [ ] 5.4 `BaseDialog` 补 Tab 循环圈定；自建 `fixed inset-0` 遮罩的约 18 个文件迁移或跟进
- [ ] 5.5 style.css 全局 `:focus-visible`；基类 `focus:` → `focus-visible:`
- [ ] 5.6 图标按钮 aria-label 本地化（密码可见性切换、"Close modal"）
- [ ] 5.7 移动端触控目标 ≥40px（技能卡「查看详情」30px 起步）

## 6. 跨仓与交叉引用（不在本 change 实现）

- [ ] 6.1 registry `source` 字段改指 state-of-art-skills 固定 commit——现指向已转为爬虫仓的 `github.com/anbeime/skill`，详情弹窗「查看来源」会把用户带去非权威仓（state-of-art-skills 仓生成端/元数据修，随下次 sync-skill-market 生效）
- [ ] 6.2 定制面硬编码中文迁 i18n → `add-tokenport-access-center` tasks 5.6（不在此重复）
- [ ] 6.3 Sub2API/TokenPort/sub2api 品牌写法统一 → 同上 tasks 6.3（不在此重复）

## 7. 递延（OQ3 拍板后另行立项，本 change 不做）

- [ ] 7.1 巨型弹窗向导化拆分：`CreateAccountModal`(6281 行) / `EditAccountModal`(4706 行) / `BulkEditAccountModal`(1997 行)（均上游文件，需评估合并负担）
- [ ] 7.2 管理侧栏信息架构：管理 26 + 个人 13 项同屏约 40 链接的分层收纳
