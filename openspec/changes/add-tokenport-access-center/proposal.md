# TokenPort 接入配置中心正式规格

## Why

2026-07-26 的对抗式审计（token-platform `docs/architecture/access-center-audit-20260726.md`，127 条复核后保留的发现）得出的核心结论：接入配置中心、连接器交付与 Skill 交付从未有过可作为依据的规格文档，「代码即需求」导致五类系统性矛盾——仓库自带的 spec-driven 流程被绕过、客户端×平台矩阵靠 if 链硬堆且漏掉 composite、同一个「导入」在同一弹窗里指五件事、文案承诺与实现相反（脱敏 Key 实为明文、「保留已有设置」实为数组覆盖）、「接入成功」无定义且上游的失败探测在重写中丢失。

6 个 blocker 中 B1–B3 已修（`5ddd5ed3`），B4/B6 与 H5/H7 及三条同源 medium 已随本 change 的配套提交修复；其余 blocker（B5）与多数 high（H1–H4、H6、H8）的修法**取决于本规格先定下来**。本 change 把审计第三节的 42 条需求条目落成正式 capability 规格：凡审计给出明确建议且不涉及外部契约的，直接采纳为规格默认值；真正需要产品拍板或向 CC Switch 侧确认的，集中列在 `design.md` 的 Open Questions，未确认前实现不得自行漂移。

## What Changes

- 固定四种交付路径的唯一命名与副作用声明句式，统一「导入」语义（条目 1–4）。
- 以显式表格定义客户端×平台能力矩阵（含 composite 的产品口径与 WebSocket 形态归属），并要求 direct 与 CCS 两路共用同一份 endpoint 解析定义（条目 5–8）。
- 定义 CC Switch deeplink 的逐参数契约、config payload 统一 schema、usageScript 执行语义与版本协商要求（条目 9–14）。
- 定义密钥可见性矩阵：六个出口逐一声明明文/脱敏与默认值，高级编辑强制显示密钥，TeleAgent 文案改为明文警示（条目 15–18）。
- 定义幂等/冲突/备份策略：TOML 合并方案、JSON 深度合并语义（数组并集，已实现）、备份保留与还原口径、Codex 认证接管边界（条目 19–24）。
- 定义 Skill 交付语义：归档布局契约（standard=`<id>/` 前缀、teleagent-root=根布局，已在 registry 流水线落地）、registry 字段白名单校验（已实现）、切换客户端时清空已选技能、riskLevel=high 二次确认、registry 信任模型与外网 fallback 默认关闭、CCS 模式不承载 Skill（条目 25–33）。
- 定义四类降级状态与「接入成功」验收口径，矩阵化验收清单替换 deployment-standard §5 的单条目验收（条目 34–39）。
- 统一产品名（中文「接入配置中心」/ 英文 "Access Center"）与第三方名 "CC Switch"，定义多语言口径（条目 40–42）。

## Capabilities

### New Capabilities

- `client-access-center`：交付路径命名与副作用声明、密钥可见性矩阵、高级编辑语义、交付动作反馈、四类降级状态、命名与多语言。
- `connector-delivery`：客户端×平台能力矩阵、endpoint 解析单一来源、CC Switch deeplink 契约、TOML/JSON 合并与备份策略、Codex 认证接管边界。
- `skill-delivery`：归档布局契约、registry 白名单校验与信任模型、runtime 适配矩阵、客户端切换行为、风险确认与安装留痕、版本语义。

### Modified Capabilities

无。仓库当前没有已发布的 access-center 相关 capability；`docs/sub2api-connector-customization.md` 降级为 Phase 1 历史记录并在文首指向本 change。

## Impact

- **前端**：`frontend/src/tokenport/access-center/`、`frontend/src/components/keys/`、`frontend/src/utils/ccswitchImport.ts`、`frontend/src/api/skillMarket.ts`——按规格收敛行为；`ClientAccessCenterDialog` 补组件测试并纳入 `FRONTEND_CRITICAL_VITEST`。
- **registry 流水线**：`state-of-art-skills/scripts/build_market.py` 双归档 + schema 1.2（已落地）；`scripts/sync-skill-market.ps1` 补 sha256 复核。
- **文档**：`docs/operations/deployment-standard.md` §5 替换为矩阵化验收清单；产品名/CC Switch 写法全仓统一。
- **不改变**：后端网关协议行为、现有 key/分组模型、上游 UseKeyModal 的原有交付能力（只回补丢失能力，不重定义上游语义）。
