# skill-delivery

## ADDED Requirements

### Requirement: 归档布局是一份被生产端和消费端共同遵守的契约
Skill 归档布局 SHALL 只有两种且由 registry 生成端保证：`standard`——归档根为 `<id>/`（供 `rm -rf $target && 解压到父目录` 语义使用）；`teleagent-root`——SKILL.md 位于归档根。`archiveLayout` 字段 MUST 与归档实际布局一致，manifest 的 hint 文案 MUST 与 layout 一致；`build_market.py` MUST 为每个技能同时产出两种归档并写入 `archive` 与 `teleagentArchive`（schema ≥1.2），校验（`--check` 与 CI）MUST 断言两种归档的 sha256/size/内容与源一致。

#### Scenario: 标准归档安装
- **WHEN** 安装脚本对 `~/.claude/skills/<id>` 执行 rm -rf 后解压 standard 归档到其父目录
- **THEN** 技能文件 MUST 恰好落在 `~/.claude/skills/<id>/` 下，MUST NOT 产生 `skills/skills/<id>` 嵌套

#### Scenario: registry 流水线全量构建
- **WHEN** 运维执行 build_market.py 后运行 sync-skill-market.ps1
- **THEN** 同步 MUST 成功（teleagent 目录存在）且每个条目都带可校验的 teleagentArchive

### Requirement: registry 字段进入脚本前必须通过白名单校验
`toSkillInstallSelection` SHALL 校验：id 匹配 `^[a-z0-9][a-z0-9._-]*$`；sha256 为 64 位十六进制；archiveUrl 解析后为 http(s)；installTargets 每个值以 `~/.` 开头且不含 `..`。校验失败 MUST 抛出用户可见错误并拒绝生成选择。安装脚本中 registry 值 MUST 只进入单引号（bash）或 psQuote 单引号（PowerShell），`$HOME` 保留在引号外展开；bash 下载 MUST 使用 `curl -fL`，4xx/5xx MUST 失败而不是把 HTML 存成 zip。

#### Scenario: registry 被篡改注入命令替换
- **WHEN** 某条目的 installTargets 含 `$(...)`、反引号或 `..`
- **THEN** 选择动作 MUST 在 UI 层被拒绝；即使绕过 UI，生成脚本中的单引号包裹 MUST 阻止 shell 展开

### Requirement: 市场来源必须可见且外网 fallback 默认关闭
Skill 市场加载 SHALL 常驻显示当前来源；非同源来源 MUST 用告警色标注。外网 fallback（jsdelivr/raw.githubusercontent）MUST 默认关闭，仅在部署期显式配置后启用，且引用 MUST 固定到 tag/commit 而非 `@main`。registry 响应 MUST 校验 content-type 与 `Array.isArray(skills)`，后端静态缺失返回的 200+HTML MUST 被识别为「同源未同步」而不是静默回落。

#### Scenario: 镜像未执行技能同步
- **WHEN** 部署镜像缺少 /skill-market/index.json 且未启用外网 fallback
- **THEN** 界面 MUST 显示「内置市场未同步」的明确错误与修复指引，MUST NOT 静默列出第三方仓库的技能

### Requirement: runtime 适配与客户端切换行为必须一致且无残留
弹窗内技能选择器与独立市场页 SHALL 使用同一套 runtime 过滤标准（`skill.runtime` 与 `installTargets` 键）。切换客户端 tab 时 SHALL 清空已选技能并提示「已按新客户端清空技能选择」（D7）；已选技能的归档选择 MUST 始终与当前 runtime 匹配。缺少对应 runtime 安装目标的技能 MUST 不可勾选，fallback MUST NOT 指向不带 `<id>` 的 skills 根目录。

#### Scenario: 从 Codex tab 切到 TeleAgent tab
- **WHEN** 用户在 Codex tab 勾选了技能后切到 TeleAgent tab
- **THEN** 已选技能 MUST 被清空并提示，manifest MUST NOT 出现 standard 归档配 teleagent hint 的组合

### Requirement: 无兼容包与错误态必须可恢复
勾选缺少 TeleAgent 兼容包的技能 SHALL 呈现为该条目的行内错误与禁用态，MUST NOT 吞掉整个列表；列表加载错误与勾选错误 MUST 分离，均可重试；`filteredSkills` 为空时 MUST 有空态文案。

#### Scenario: 勾选无 TeleAgent 包的技能
- **WHEN** registry 某条目缺 teleagentArchive 且用户在 TeleAgent tab 勾选它
- **THEN** 仅该条目显示不兼容说明，其余技能保持可选

### Requirement: 高风险技能需要确认与留痕
riskLevel=high 的技能在勾选时 SHALL 弹出一次二次确认（说明风险与来源）；确认后系统 MUST 在本地留下安装决策记录（至少：技能 id、版本、sha256、时间、来源 URL），满足运维文档「Skill install and risk review records」的要求。license 展示 MUST 本地化处理 `review-required` 等占位值。

#### Scenario: 用户勾选 high 风险技能
- **WHEN** 用户勾选 riskLevel=high 的技能
- **THEN** 界面 MUST 弹出确认并在确认后记录决策，取消则不选中

### Requirement: 版本语义短期口径必须明示
在版本管理落地前，界面与脚本 SHALL 明示「每次安装为全量覆盖（rm -rf 后重装），不支持回滚」（D11）。CCS 模式 MUST NOT 承载 Skill：技能选择器在 CCS 模式隐藏，相关文案不得指向不存在的按钮（D9）。

#### Scenario: CCS 模式下查看技能区域
- **WHEN** 用户切到「交给 CC Switch 导入」模式
- **THEN** 技能选择器 MUST 隐藏，且弹窗内 MUST NOT 出现「Skill 仍通过独立脚本安装」这类指向隐藏按钮的文案
