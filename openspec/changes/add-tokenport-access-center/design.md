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

1. ~~**CCS deeplink 契约三问（需向 CC Switch 侧确认）**~~ **已确认（2026-07-26，依据 cc-switch v3.18.0 源码，见文末「CC Switch deeplink 契约确认记录」）**：三问结论——usageScript 是 QuickJS 对整段脚本求值、取完成值为 `{request, extractor}` 对象（两者都不是：非 `new Function` 函数体、也非浏览器 eval）；config 是**建档不落盘**（导入只写 CC Switch 内部供应商库，`enabled=true` 才顺带切换写盘）；未知参数自 v3.7.0 起一律静默忽略，真正的版本闸门在 `app` 白名单与必填参数。connector-delivery R3 已按结论更新。
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

## CC Switch deeplink 契约确认记录（2026-07-26）

确认方式：直接审读 cc-switch 源码（`farion1231/cc-switch`，当前版 v3.18.0 @ 878c26f，旧版按 tag 抽查 v3.7.0 / v3.8.1 / v3.10.0 / v3.11.0 / v3.12.0 / v3.14.0 / v3.17.0）。源码即契约权威；官方文档 `docs/user-manual/*/5-faq/5.3-deeplink.md` 存在偏差（把 usageScript 标成 URL-encoded，实际必须 base64）。

### 三问结论

**1. usageScript 执行语义**（`src-tauri/src/usage_script.rs`、`deeplink/provider.rs::build_provider_meta`）
- 线上编码：**base64（UTF-8 字节）**，解析端 `decode_base64_param` + `String::from_utf8`。我方 `btoa()` 只能编 Latin-1，故脚本内容必须保持 ASCII（已加 spec 断言）。
- 执行引擎：**Rust 侧 QuickJS（rquickjs）`ctx.eval(整段脚本)`**，取求值完成值——既不是 `new Function` 函数体（裸 `return` 会语法错误），也不是浏览器 eval。规范形态是带括号的对象字面量 `({ ... })`。
- 结果契约：对象必须含 `request`（`{url, method, headers?, body?}`，方法非法即失败不回退 GET）和 `extractor`（函数，入参为响应 JSON，返回对象或对象数组；字段全可选但类型受检：isValid/invalidMessage/remaining/used/total/unit/planName/extra）。
- 模板变量在 eval 前做文本替换：`{{apiKey}}`、`{{baseUrl}}`、`{{accessToken}}`、`{{userId}}`；查询时 apiKey/baseUrl 按供应商 live 配置解析（usageApiKey/usageBaseUrl 可覆盖，与本值相同的覆盖会被归一化清掉，#4654）。
- 安全约束：请求 URL 强制 HTTPS（localhost 豁免）且必须与 base_url **同 host+port**；超时钳制 2–30s；usageAutoInterval 上限 1440 分钟。deeplink 导入的 template_type=None → 上述校验全部生效。
- 上游 sub2api `KeysView.vue` 生成的 `({request, extractor})` 模板完全合规；**我方 access-center 曾改写成 `({endpoint, key})`，不合规（查询必报「缺少 request 配置」），已修复回上游形态**（`buildCcsUsageScript()`）。
- 遗留注意（并入 H2/H3）：`{{baseUrl}}` 查询时取 live 配置的 endpoint，antigravity 平台的 endpoint 带 `/antigravity` 前缀 → `{{baseUrl}}/v1/usage` 路径错位；上游同样有此问题。可选修法：deeplink 附 `usageBaseUrl=<裸 origin>`（v3.9.0+）。

**2. config payload 语义**（`deeplink/provider.rs::import_provider_from_deeplink / parse_and_merge_config`）
- **建档不落盘**：导入只调 `ProviderService::add` 写 CC Switch 内部供应商库；客户端配置文件仅在切换到该供应商时才写（`enabled=true` 会导入后立即切换）。导入前有 App 内确认对话框。
- 合并优先级：**URL 参数 > inline config（base64）**；`configUrl` 尚未实现（直接报错）。
- 分 app 语义：claude —— config 的 `env` 对象整体保留（自定义 env 变量存活，#2928），标准键被 URL 参数覆盖；codex/gemini/grokbuild/opencode/openclaw/hermes —— config 仅被提取 apiKey/endpoint/model 用于补缺，之后 CC Switch **重新生成自家规范配置**（codex 固定模板 `wire_api="responses"` + `requires_openai_auth=true`；opencode 生成 `{npm:"@ai-sdk/openai-compatible", options{baseURL,apiKey}, models}`），其余 config 内容一律丢弃。
- 推论：OpenCode 走 CCS **不需要也不应该发 settings.json 形态的 config**——endpoint/apiKey/model URL 参数即足够，CC Switch 自建 opencode 供应商结构；R3 的 OpenCode 分支已按此改写。无 config 时携带 `configFormat` 无害（解析端提前返回），但按 R3 保持不发。

**3. 旧版对未知参数的行为**（v3.7.0 `deeplink.rs` ↔ 现行 `deeplink/parser.rs` 对照）
- 所有版本都是 `HashMap` + 已知键逐个 `.get()`：**未知参数静默忽略，从不报错**。发新参数给旧版是安全的（功能缺失但不炸）。
- 真正的版本闸门：
  - `app` 白名单：claude/codex/gemini（v3.7.0+）→ +opencode/openclaw（**v3.12.0**）→ +hermes（v3.14.0）→ +grokbuild（**v3.18.0**）。旧版收到白名单外的 app 直接报「Invalid app type」。
  - 必填参数：v3.7.x 要求 homepage+endpoint+apiKey（我方始终全发 → 兼容）；v3.8.0 起三者可选（可由 config 补齐）。
  - usage 系参数 v3.9.0 起生效；haiku/sonnet/opusModel v3.7.1 起生效；更早版本均为忽略。
- 协议无版本协商（仅校验 host 段 `v1`）、**无回执机制**：导入结果只在 CC Switch 界面呈现，发起页唯一可用信号是 focus 启发式（上游已实现 100ms `document.hasFocus()` 探测，对应 tasks.md 4.2）。协议错误（非法 app 等）只在 CC Switch 内部弹出，网页侧无感。
