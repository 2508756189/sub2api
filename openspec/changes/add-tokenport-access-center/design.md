# 设计决策与开放问题

## 已采纳的审计建议（规格默认值）

以下决策直接采纳审计报告的建议方案，已写入三个 capability 的正式需求；改动它们需要修订对应 spec，而不是在实现里绕开：

| # | 决策 | 采纳方案 | 落点 |
| --- | --- | --- | --- |
| D1 | 交付路径枚举 | 固定四种：「一键写入本机（脚本）」「交给 CC Switch 导入」「下载配置文件自行放置」「在目标 App 内手工导入」 | client-access-center R1 |
| D2 | 密钥可见性 | 预览默认脱敏；高级编辑强制显示并置灰开关；脚本/下载/deeplink 明文但常驻警示 | client-access-center R2 |
| D3 | 高级编辑 | 定位为逃生舱：以 path 为键跟踪 dirty，选项变动只重置未编辑项 | client-access-center R3 |
| D4 | JSON 合并 | 两层合并 + 数组并集（已实现于安装脚本生成器） | connector-delivery R5 |
| D5 | TOML 合并 | 默认方案：检测目标文件已含同名表/根键即中止并提示手工合并；独立 profile 文件作为后续增强 | connector-delivery R4 |
| D6 | Skill 归档布局 | standard=`<id>/` 前缀、teleagent-root=根布局；registry 生成端保证，消费端按 archiveLayout 断言（已落地） | skill-delivery R1 |
| D7 | 切换客户端 tab | 清空已选技能并提示重新勾选（最小正确方案） | skill-delivery R4 |
| D8 | registry 信任模型 | 外网 fallback 默认关闭、部署期可配置；界面常驻显示市场来源；安装脚本 curl 加 `-fL` | skill-delivery R7 |
| D9 | CCS 模式与 Skill | CCS 模式不承载 Skill：隐藏技能选择器并修正文案 | skill-delivery R8 |
| D10 | composite 平台 | 默认口径：显式不支持——composite key 在配置中心显示引导文案而非生成错误协议配置 | connector-delivery R1 |
| D11 | 版本语义 | 短期明说「每次安装为全量覆盖，不支持回滚」；版本管理为独立后续 change | skill-delivery R9 |

## Open Questions（实现前必须逐条确认，未确认项不得自行漂移）

1. **CCS deeplink 契约三问（需向 CC Switch 侧确认）**：`usageScript` 的执行语义（`new Function` 函数体还是 `eval` 表达式）；`config` payload 是按 model 参数建档还是按文件落盘；旧版 CCS 对未知参数的行为。——阻塞 connector-delivery R3 的最终参数表与 Grok model / OpenCode config / ws 标记的修法。
2. **gemini direct 路径是否应带 `/v1beta`**：上游 CLI 输出裸 base，TokenPort direct 多拼了一层。需用真实 gemini-cli 抓包确认后统一 `resolveClientEndpoint`。——阻塞 H3 修复方向。
3. **antigravity × OpenCode 的协议选择**：补 `anthropic + /antigravity/v1` 还是恢复上游双文件（anthropic + google 两份）。——阻塞 H2。
4. **WebSocket 形态**：独立 codex-ws tab 还是 transport 开关（当前死分支残留）。二选一后从矩阵与代码中删除另一半。
5. **Codex 认证接管边界**：直接配置模式是否覆盖已有 ChatGPT OAuth 登录态（写 `preferred_auth_method` 还是保留并存并在 UI 说明）。
6. **supportsSkills 范围**：OpenCode / Gemini CLI / Grok CLI 不支持技能是有意还是漏配。
7. **composite 的后续形态**（D10 之后）：是否按 composite-routes 的 `target_platform` 集合合并出多个 tab。

## 与审计条目的映射

- 第一档 1–7 → 本 change 的三个 spec + 本文件。
- 第二档剩余：B5 → connector-delivery R4（D5 确认后动手）；已完成项见 tasks.md 第 0 节。
- 第三档 → tasks.md 第 3–6 节。
