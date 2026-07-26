# frontend-a11y

## ADDED Requirements

### Requirement: 模态框必须圈定键盘焦点
`BaseDialog` SHALL 在开启期间将 Tab/Shift+Tab 循环限制在对话框内（focus trap）。现状已有 Esc 关闭（`frontend/src/components/common/BaseDialog.vue:110-114`）与初始聚焦/关闭还原（`BaseDialog.vue:126-141`），缺循环圈定——Tab 可逃出对话框落到背景页面。自建 `fixed inset-0` 遮罩的约 18 个文件 SHALL 迁移到 BaseDialog 或实现同等圈定。

#### Scenario: 键盘用户在弹窗内连续 Tab
- **WHEN** 弹窗开启且用户反复按 Tab
- **THEN** 焦点 MUST 始终在弹窗内循环，MUST NOT 落到背景页面元素

### Requirement: 键盘焦点样式必须全局定义
`style.css` SHALL 提供全局 `:focus-visible` 样式（现状 0 条，仅个别组件自带 `focus-visible:` 工具类）；`.btn`/`.input` 等全站基类的 `focus:ring` SHALL 改为 `focus-visible:` 系——键盘导航显示焦点环，鼠标点击不显示。

#### Scenario: 键盘 Tab 导航任意页面
- **WHEN** 用户用 Tab 在页面中移动焦点
- **THEN** 当前焦点元素 MUST 有可见焦点环，且鼠标点击同一元素 MUST NOT 出现焦点环

### Requirement: 图标按钮必须有本地化可访问名
纯图标按钮 SHALL 带 `aria-label` 且文案走 i18n：登录页密码可见性切换现状无 aria-label（实测）；`BaseDialog` 关闭按钮现状为硬编码英文 "Close modal"（实测），MUST 本地化。

#### Scenario: 屏幕阅读器聚焦密码切换按钮
- **WHEN** 读屏器聚焦该按钮
- **THEN** MUST 播报本地化名称（如「显示密码」/「隐藏密码」），MUST NOT 静默或播报英文

### Requirement: 移动端触控目标不低于 40px
移动视口下可点击目标高度 SHALL ≥40px（D8）；技能市场卡片「查看详情」按钮现状 30px（375px 实测）SHALL 调整，调整 MUST NOT 以缩小字号为代价。

#### Scenario: 手机用户点击卡片操作按钮
- **WHEN** 375px 视口下浏览技能市场卡片
- **THEN** 操作按钮触控高度 MUST ≥40px
