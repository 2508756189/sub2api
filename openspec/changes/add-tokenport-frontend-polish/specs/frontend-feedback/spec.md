# frontend-feedback

## ADDED Requirements

### Requirement: 写操作成功必须有可感知反馈
所有 mutation（创建/更新/删除/导入/发送）成功后 SHALL 给出可感知反馈：`showSuccess` toast，或当前视口内可见的状态变化（行内更新、跳转、伴随高亮的列表刷新）二选一（D5）。仅失败提示、成功静默的路径视为缺陷（普查计数：`showError` 556 次 vs `showSuccess` 247 次）。

#### Scenario: 用户保存设置成功
- **WHEN** 用户提交任一设置表单且后端返回成功
- **THEN** 界面 MUST 在 1 秒内给出成功反馈，MUST NOT 无任何变化地停留原地

### Requirement: 破坏性操作必须确认
删除、覆盖、吊销、不可逆状态变更 SHALL 经过 ConfirmDialog（或等价确认）并说明后果；确认文案 MUST 指明对象（名称/数量）。现状 57 个 `*Modal/*Dialog` 文件仅 27 个引用 ConfirmDialog，SHALL 盘点补齐（允许结论为「该文件无破坏性操作」，但盘点记录 MUST 留存于 tasks 勾选说明）。

#### Scenario: 删除一个 API Key
- **WHEN** 用户触发删除且尚未经确认弹窗
- **THEN** 系统 MUST NOT 直接执行删除

### Requirement: 等待型页面必须有进行中状态
登录/支付回调等等待型页面 SHALL 呈现进行中指示（spinner 或骨架 + 文案），完成/失败各有明确终态。已核实：`WechatCallbackView.vue`（1102 行）、`OidcCallbackView.vue`（856 行）、`DingTalkCallbackView.vue`（852 行）未检出任何加载指示（含「正在…」文案），而 `OAuthCallbackView.vue` 有 spinner——同族页面 MUST 一致。

#### Scenario: 微信扫码后回调处理中
- **WHEN** 回调页正在与后端交换凭据
- **THEN** 页面 MUST 显示进行中指示，MUST NOT 白屏或静止

### Requirement: 空态视觉必须收敛为一套
空列表呈现 SHALL 收敛：表格场景用 DataTable 内置空态，非表格场景用 `EmptyState` 组件，两者的图标尺寸/字号/间距/文案键 MUST 对齐（D6）；MUST NOT 再新增第三种空态写法。

#### Scenario: 空表格与空卡片列表并排出现
- **WHEN** 同一页面同时出现空 DataTable 与空卡片区
- **THEN** 两个空态 MUST 呈现同一套视觉语言
