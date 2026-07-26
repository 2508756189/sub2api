# frontend-navigation

## ADDED Requirements

### Requirement: 移动端公开页导航必须保留核心入口
≤720px 视口下，首页顶栏 SHALL 至少保留 Skill Market 入口、主题切换与登录/进入控制台（D4）；MUST NOT 如现状（`frontend/src/tokenport/home/TokenPortHome.vue:1189-1193` 将非主链接与 `.icon-control` 全部 `display:none`）只剩登录一项（375px 实测：顶栏仅「登录平台」可见）。

#### Scenario: 手机用户想从首页进入技能市场
- **WHEN** 用户在 375px 视口打开首页
- **THEN** 顶栏 MUST 提供可点击的 Skill Market 入口与主题切换

### Requirement: 公开页头部必须结构一致
公开页（首页、技能市场未登录页、登录页）的头部 SHALL 共用同一结构定义（logo+副标、导航区、主题切换、主行动按钮）或其明确定义的子集；技能市场未登录页 MUST 提供主题切换（实测现状缺失，仅「返回首页/登录控制台」）；同一元素在不同公开页的位置与样式 MUST 一致。

#### Scenario: 用户在公开页间跳转
- **WHEN** 用户从首页进入技能市场再返回
- **THEN** 头部布局 MUST 保持稳定，MUST NOT 出现主题切换忽有忽无

### Requirement: 登录门控链接必须可预期且可回跳
公开导航中指向需登录页面的链接（如「模型与渠道」，实测未登录点击直接落到登录页且无任何预示）SHALL 有门控提示（锁形图标或等价标识）；未登录点击后 SHALL 跳登录页并在登录成功后回跳原目标。

#### Scenario: 未登录用户点击「模型与渠道」
- **WHEN** 未登录用户从首页导航点击「模型与渠道」
- **THEN** 跳转登录页 MUST 携带回跳信息，登录成功后 MUST 到达模型与渠道页而非默认首页
