# connector-delivery

## ADDED Requirements

### Requirement: 客户端与平台的能力矩阵必须显式声明并由数据驱动
系统 SHALL 维护一张显式能力矩阵（行：claude / codex / gemini / grok / opencode / teleagent；列：anthropic / openai / gemini / antigravity / grok / composite），每格取值为 支持 / 不支持 / 需 `allow_messages_dispatch`，并注明依据（目标客户端支持的 wire 协议）。客户端 tab 的生成 MUST 读取该矩阵，MUST NOT 以 if 链散落在组件里。composite 平台按 D10 口径：MUST 显示引导文案，MUST NOT 落入任何默认分支生成错误协议配置。WebSocket 形态按 Open Question 4 拍板前，`codex-ws` 死分支 MUST NOT 扩散。

#### Scenario: composite 平台的 key 打开配置中心
- **WHEN** 用户为 platform=composite 的 key 打开配置中心
- **THEN** 界面 MUST 显示 composite 引导说明
- **THEN** 系统 MUST NOT 用 `buildAnthropicFiles` 或任何单协议模板为其生成配置

#### Scenario: 未声明的组合出现
- **WHEN** 出现矩阵中标注「不支持」的 客户端×平台 组合
- **THEN** 对应 tab MUST 不出现或显示不支持说明，MUST NOT 生成配置文件

### Requirement: endpoint 解析必须是单一共享定义
系统 SHALL 提供唯一的 `resolveClientEndpoint(platform, client)` 定义（含版本段与 antigravity 前缀规则），使用 `new URL()` 解析而非字符串尾部匹配；direct 与 CCS 两路 MUST 调用同一实现。三套既有归一化函数（`ensureApiVersion` / `ensureOpenAIBaseUrl` / `withV1Endpoint`）MUST 合并进该定义。同一 key 的同一客户端在 direct 与 CCS 两路得到的 endpoint MUST 一致（gemini 与 antigravity 的最终取值以 Open Questions 2/3 的确认结果为准）。

#### Scenario: base_url 带查询串
- **WHEN** 管理员把 api_base_url 配成带查询串或尾斜杠的 URL
- **THEN** 解析结果 MUST 是合法 URL，MUST NOT 产出 `https://h/v1?x=1/v1` 形态

#### Scenario: 同一 key 在两种交付模式间切换
- **WHEN** 用户在同一客户端 tab 切换 direct 与 CCS
- **THEN** 两路生成的 endpoint MUST 逐字符一致

### Requirement: CC Switch deeplink 契约必须逐参数成文
deeplink 的每个参数（resource/app/name/homepage/endpoint/apiKey/model/haikuModel/sonnetModel/opusModel/config/configFormat/usageEnabled/usageScript/usageAutoInterval）SHALL 有类型、必填性、编码方式与引入版本的成文定义（契约事实见 design.md「CC Switch deeplink 契约确认记录」）。payload MUST 按 clientType 过滤：非 Claude 客户端 MUST NOT 携带 Claude 档位参数；grokbuild 必选 model MUST 进入顶层参数；OpenCode 走 CCS 时 MUST NOT 发送 `settings.json` 形态的 config（CC Switch 只从 config 提取 apiKey/endpoint/model 后自建 opencode 供应商结构，其余内容丢弃）——首选省略 config 仅发 URL 参数，若发送则 MUST 为可提取的 opencode provider 形态；无 config 时 MUST NOT 声明 `configFormat=json`。`providerName` 来源为 `publicSettings.site_name`，回落 `'TokenPort'`。

已确认的协议事实（2026-07-26）：homepage/endpoint/apiKey MUST 始终随 URL 参数发出（兼容 v3.7.x 必填要求）；未知参数被所有版本静默忽略，但 `app=opencode` 要求 CCS ≥ v3.12.0、`app=grokbuild` 要求 CCS ≥ v3.18.0，UI 提示 CC Switch 版本要求时 MUST 按此口径；协议无版本协商与回执，导入成败只在 CC Switch 内呈现，发起页 MUST 依赖 focus 启发式提示未安装/未接管（见 client-access-center R4）。

#### Scenario: 用户在 Claude tab 填过档位后切到 codex 导入
- **WHEN** connectorOptions 中存在 Claude 档位残留且当前 clientType=codex
- **THEN** deeplink MUST NOT 包含 haikuModel/sonnetModel/opusModel

#### Scenario: OpenCode 走 CCS
- **WHEN** 用户在 OpenCode tab 点击导入 CCS
- **THEN** deeplink MUST 携带 endpoint/apiKey（及已选 model）URL 参数，MUST NOT 发送 settings.json 形态的 config，MUST NOT 在无 config 时声明 configFormat

### Requirement: CCS usageScript 必须满足 QuickJS 求值契约
经 deeplink 发送的 `usageScript` MUST 为 UTF-8 base64；解码后的脚本 MUST 是求值结果为对象的表达式（规范形态 `({ ... })`），该对象 MUST 含 `request`（url/method/headers）与 `extractor`（函数）。脚本 MUST 使用 `{{baseUrl}}`/`{{apiKey}}` 模板变量而非烧录字面量（使密钥/端点编辑后查询仍随 live 配置走）；请求 URL MUST 与 base_url 同源（HTTPS），指向 `/v1/usage`。在编码器仍使用 `btoa()` 期间脚本内容 MUST 保持 ASCII。MUST NOT 发送 `({endpoint, key})` 等非契约形态（该形态在 CC Switch 查询时报「缺少 request 配置」）。

#### Scenario: CCS 导入后查询用量
- **WHEN** 用户经任一客户端 tab 导入 CC Switch 并触发用量查询
- **THEN** 脚本求值 MUST 得到含 request/extractor 的对象，查询 MUST 命中 `{{baseUrl}}/v1/usage` 且通过同源校验

#### Scenario: antigravity 平台 key 导入后查询用量
- **WHEN** platform=antigravity 的 key 经 CCS 导入（endpoint 带 `/antigravity` 前缀）
- **THEN** 用量查询 URL MUST 解析到后端真实的 `/v1/usage` 路由（经 `usageBaseUrl` 覆盖或等效手段，与 H2/H3 的 endpoint 决议一并实现）

### Requirement: TOML 交付不得使用无差别文本追加
Codex/Grok 的 TOML 交付 SHALL 遵循 D5：写入前 MUST 扫描目标文件，若已存在同名 table 或将要写入的根级键，MUST 中止并提示手工合并（列出冲突键），MUST NOT 直接追加托管块。托管块以裸键开头的模板 MUST 改为完整 table 或显式根级段，保证追加到空文件时也不产生归属歧义。独立 profile 文件（`config.tokenport.toml` + 主文件单一引用）作为后续增强实现。

#### Scenario: 用户已有 [mcp_servers] 配置
- **WHEN** 用户 `~/.codex/config.toml` 末尾是任意 table 且执行一键脚本
- **THEN** 脚本 MUST 检测到根级键归属风险并中止提示，MUST NOT 把 `model_provider` 写进上一个 table 的作用域

### Requirement: 合并与备份语义必须与文案一致
JSON 合并 SHALL 为两层合并且数组做并集（保留用户已有项，追加新增项）。备份 SHALL 带保留策略：同一目标保留最近 N=5 份、时间戳到秒且同秒不覆盖（追加序号）；首次备份（原始配置）MUST NOT 被后续执行覆盖；脚本输出 MUST 包含还原命令示例。UI 上「保留已有项目设置」的文案 MUST 与实际合并语义一致。

#### Scenario: 用户已有 enabledPlugins 再执行脚本
- **WHEN** 目标 settings.json 已有 `enabledPlugins: ["a","b"]` 且本次交付含 `["b","c"]`
- **THEN** 结果 MUST 为 `["a","b","c"]`

#### Scenario: 同一秒内重复执行
- **WHEN** 用户在同一秒内执行两次一键脚本
- **THEN** 第一次的原始备份 MUST 仍然存在且未被覆盖

### Requirement: Codex 认证接管边界必须显式
在 Open Question 5 确认前，直接配置模式 SHALL 保持「不删除已有 ChatGPT OAuth tokens」的现状，但 MUST 在 Codex tab 显示凭据并存说明（存在 tokens 时提示实际路由取决于 Codex 版本）。确认后按结论补 `preferred_auth_method` 或清理逻辑。

#### Scenario: 已用 ChatGPT 登录过的用户执行脚本
- **WHEN** auth.json 中已有 tokens 字段且用户执行一键脚本
- **THEN** 界面/脚本输出 MUST 包含双凭据并存的说明，MUST NOT 静默留下歧义状态
