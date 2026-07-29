# client-access-center

## ADDED Requirements

### Requirement: 交付路径必须有唯一命名和常驻副作用声明
配置中心 SHALL 只提供四种交付路径，全局统一文案：「一键写入本机（脚本）」「交给 CC Switch 导入」「下载配置文件自行放置」「在目标 App 内手工导入（TeleAgent/Skill）」。每种路径 MUST 常驻显示副作用声明（是否改本机文件、是否产生备份、是否由第三方 App 落盘），而不是只靠动态提示。「一键写入本机」与「交给 CC Switch 导入」是同级互斥的交付动作，MUST NOT 与「复制当前文件」并列在页脚同一按钮组。

#### Scenario: 用户在直接配置模式查看页脚
- **WHEN** 用户在任一客户端 tab 的直接配置模式停留
- **THEN** 界面 MUST 显示当前路径的副作用声明（例如「将修改 ~/.codex/config.toml，写入前自动备份」）
- **THEN** 不可用的交付按钮 MUST 以禁用态 + 原因提示呈现，MUST NOT 无解释地消失

#### Scenario: KeysView 行内按钮
- **WHEN** 用户点击密钥列表的行内接入按钮
- **THEN** 其行为定义为「打开配置中心」，i18n 文案 MUST 与该语义一致，MUST NOT 写成「导入到 CCS」

### Requirement: 密钥可见性必须按出口逐一定义且开关语义一致
六个出口的密钥可见性 SHALL 固定为：屏幕预览默认脱敏（可切换）；高级编辑 textarea 始终明文，进入时 MUST 强制将「显示密钥」置为开且置灰并说明原因；剪贴板、下载文件、一键脚本、deeplink 均为明文，对应操作旁 MUST 有明文风险提示。TeleAgent 提供商字段文件的文案 MUST 声明为「API Key（明文，请勿外发）」，MUST NOT 声称脱敏。

#### Scenario: 用户在隐藏密钥状态点击高级编辑
- **WHEN** 预览处于「隐藏密钥」状态且用户进入高级编辑
- **THEN** 「显示密钥」开关 MUST 自动置为开并置灰
- **THEN** textarea 内容与开关状态 MUST NOT 出现「界面称隐藏、实际明文」的矛盾

#### Scenario: 用户下载包含明文密钥的文件
- **WHEN** 用户点击下载当前文件且文件含明文 apiKey
- **THEN** 系统 MUST 给出一次明文确认提示后才写入下载

### Requirement: 高级编辑是逃生舱，编辑内容不得被静默丢弃
高级编辑内容 SHALL 以文件 path 为键跟踪 dirty 状态。左侧选项变动重新生成文件时，系统 MUST 只重置未编辑文件，已编辑文件 MUST 保留内容并显示「已修改」标记；文件集合变化导致某已编辑文件不再存在时 MUST 显式提示而非错位。交付动作（脚本/复制/下载/deeplink）使用的内容 MUST 与用户当前看到的编辑内容一致。

#### Scenario: 用户编辑 auth.json 后切换 shell
- **WHEN** 用户在高级编辑中修改了 auth.json 再切换 shell 选项
- **THEN** auth.json 的编辑内容 MUST 保留且带「已修改」标记，其余文件按新选项重新生成

#### Scenario: 用户把 JSON 编辑坏后执行交付动作
- **WHEN** 高级编辑中的 JSON 不合法且用户点击「导入 CCS」或复制脚本
- **THEN** 系统 MUST 阻止该动作并给出指明文件的用户可见错误，MUST NOT 静默无响应

### Requirement: 每个交付动作必须有用户可见的成功或失败反馈
所有交付动作 SHALL 产生可见回执。CCS deeplink 跳转 MUST 恢复协议接管探测（try/catch + 延时 `document.hasFocus()` 判定），未接管时 MUST 提示未安装 CC Switch（复用 i18n key `keys.ccSwitchNotInstalled`）并提供「复制 CCS 链接」兜底入口。脚本路径 MUST 在脚本末尾输出可回贴的验证命令或自检输出；「接入成功」的弱判定口径为：生成配置后用当前 baseUrl+key 调一次 `/v1/models`。

#### Scenario: 未安装 CC Switch 的用户点击导入
- **WHEN** 用户点击「交给 CC Switch 导入」而系统未注册 ccswitch 协议
- **THEN** 界面 MUST 在探测超时后提示未安装并展示「复制 CCS 链接」按钮

### Requirement: 四类降级必须有区分的状态与文案
系统 SHALL 区分四类降级并各自定义状态与出路：模型接口失败（可重试，显示错误态而非「分组未配置模型」）；分组无可用模型（引导去渠道页，手填模型标记「未经平台验证」）；分组未绑定（隐藏交付模式切换）；Skill registry 不可达（区分同源未同步与外网 fallback 失败）。模型列表加载失败 MUST NOT 与「分组没有模型」共用同一文案。

#### Scenario: 模型接口返回 500
- **WHEN** 分组模型接口请求失败
- **THEN** 界面 MUST 显示可重试的错误态，MUST NOT 显示「分组未配置可用模型」

### Requirement: 产品名、第三方名与多语言口径必须唯一
产品名 SHALL 固定为中文「接入配置中心」、英文 "Access Center"，并统一替换 TOKENPORT.md / AGENT.md / README / repository-map / deployment-standard 与 UI 标题。第三方名 SHALL 统一写作 "CC Switch"（缩写场景 CCS）。TokenPort 定制 UI 的文案 MUST 迁入独立 i18n 命名空间文件（新增文件避免上游合并冲突）；在未翻译 locale 下 MUST 有明确降级说明而非中英混排。

#### Scenario: 英文用户打开配置中心
- **WHEN** locale 为 en 的用户打开配置中心
- **THEN** 标题与按钮 MUST 显示 Access Center 命名空间下的英文文案或声明式降级说明
