# Tasks

来源：`token-platform/docs/architecture/access-center-audit-20260726.md` 第五节推进顺序。已完成项标注提交。

## 0. 已完成（先于本 spec 落地的无争议修复）

- [x] 0.1 B1 bash `$HOME` 单引号（`5ddd5ed3`，含回归测试）
- [x] 0.2 B2 PowerShell `-AsHashtable` 5.1 兼容（`5ddd5ed3`）
- [x] 0.3 B3 Windows 技能脚本 `$HOME`、`"$SkillId:"` 解析期失败（`5ddd5ed3`）
- [x] 0.4 B6 registry 流水线补 teleagentArchive、schema 1.2（state-of-art-skills `build_market.py` + 测试）
- [x] 0.5 B4 标准归档布局改为 `<id>/` 前缀并全量重建 sha256（同上），端到端验证解压落点
- [x] 0.6 H7 TeleAgent 准备脚本双重引号（本 change 配套提交，spec 断言防回归）
- [x] 0.7 H5 bash 技能脚本单引号化 + `toSkillInstallSelection` 白名单校验（同上）
- [x] 0.8 B2 同批 medium：PS 写盘去 BOM、TOML 空文件 `$null` 崩溃、JSON 数组并集（同上，PS 5.1 实机双跑验证）

## 1. 决策关闭（阻塞后续实现，见 design.md Open Questions）

- [ ] 1.1 向 CC Switch 侧确认 deeplink 三问（usageScript 语义 / config 落盘方式 / 未知参数行为），回写 design.md
- [ ] 1.2 用真实 gemini-cli 验证 direct 路径 `/v1beta` 取舍，敲定 `resolveClientEndpoint` 的 gemini/antigravity 规则
- [ ] 1.3 拍板 WebSocket 形态（独立 tab vs transport 开关），删除另一半死分支
- [ ] 1.4 拍板 Codex OAuth 接管边界与 supportsSkills 范围

## 2. Blocker 收尾（决策后立即动手）

- [ ] 2.1 B5 Codex TOML：按 connector-delivery R4 实现同名表/根键检测中止；模板改完整 table；补「已存在 TokenPort 配置」的确认输出（条目 20）
- [ ] 2.2 sync-skill-market.ps1 补 sha256 复核（对照 index.json 而非纯 Copy-Item）

## 3. High 批次

- [ ] 3.1 H1 密钥可见性：高级编辑强制显示、TeleAgent 文案改明文警示、下载确认（client-access-center R2）
- [ ] 3.2 H2/H3 endpoint：实现共享 `resolveClientEndpoint` 并替换三套归一化函数，按 1.2 结论修 gemini/antigravity；同步改 `accessCenterFiles.spec.ts:23` 断言
- [ ] 3.3 H4 切 tab 清空已选技能并提示（skill-delivery R4）
- [ ] 3.4 H6 registry 信任模型：fallback 默认关 + 来源显示 + content-type/结构校验 + curl `-fL`（skill-delivery R3/R2）
- [ ] 3.5 H8 CMD tab：禁用一键脚本按钮 + title 提示（或 EncodedCommand 方案，二选一写回 spec）；`codexConfigDir` 补 `'cmd'` 分支

## 4. Medium 批次（按 spec 逐条）

- [ ] 4.1 CCS payload 按 clientType 过滤、grokbuild model 进顶层、OpenCode config 分支（connector-delivery R3）
- [ ] 4.2 deeplink 接管探测 + `keys.ccSwitchNotInstalled` 复用 + 「复制 CCS 链接」兜底（client-access-center R4）
- [ ] 4.3 高级编辑 path 键 dirty 跟踪（client-access-center R3）
- [ ] 4.4 备份保留策略 N=5、同秒不覆盖、原始备份保护、还原命令输出（connector-delivery R5）
- [ ] 4.5 CCS 模式隐藏技能选择器并修文案（skill-delivery R8）
- [ ] 4.6 SkillMarketSelector 错误态分离可重试 + 空态文案（skill-delivery R6）
- [ ] 4.7 riskLevel=high 二次确认与安装留痕（skill-delivery R7）
- [ ] 4.8 降级四态与模型加载错误归因（client-access-center R5）
- [ ] 4.9 Grok tab 独立选项（不复用「Codex 模型」标签），无效控件隐藏

## 5. 工程化

- [ ] 5.1 `ClientAccessCenterDialog` 组件测试：platform × clientTab 全组合断言 ccsConfig 非空可 parse、deeplink 参数、非 claude 不带档位、坏 JSON 走 showError
- [ ] 5.2 TokenPort spec 纳入 `FRONTEND_CRITICAL_VITEST`；`check_tokenport_ci_parity.ps1` 补 `test-integration` 与 `govulncheck`
- [ ] 5.3 CI 闸门顺序：先推 sync 分支、GitHub CI 过后再 fast-forward 产品分支（token-platform 同步脚本）
- [ ] 5.4 FileConfig 加显式 `kind`/`writeTarget` 字段，删除 `isSkillInstallFile`/`isWritableClientFile` 的字符串匹配
- [ ] 5.5 合并两个 psQuote 到共享模块；删 `codex-ws` 死分支与 `configFormat:'toml'` 死类型；渲染 `FileConfig.hint`
- [ ] 5.6 access-center 文案迁独立 i18n 命名空间；清理孤儿 key（client-access-center R6）

## 6. 文档

- [ ] 6.1 `docs/sub2api-connector-customization.md` 降级为 Phase 1 历史记录并指向本 change
- [ ] 6.2 deployment-standard §5 替换为 客户端×交付模式×平台 矩阵化验收清单（条目 37–38）
- [ ] 6.3 产品名/CC Switch 写法全仓统一（条目 40–41）
