## ADDED Requirements

### Requirement: Hub 用户反馈卡片 MUST 位于设备记录下方且常驻全页入口

`resource/public/admin.html` 在管理员登录后展示的卡片顺序中，用户反馈卡片（`#feedbackCard`）MUST 紧挨设备记录卡片（`#deviceRecordCard`）之后。反馈卡片头部 MUST 常驻可见链接，文案为「查看全部」（或等价），目标为 `/device/admin/feedback-records`。该入口 MUST NOT 依赖反馈条数是否超过预览页大小才显示。

#### Scenario: 登录后可见常驻入口

- **WHEN** 管理员在运维 Hub 登录成功且反馈卡片已显示
- **THEN** 页面 MUST 展示可点击的「查看全部」链至反馈全页
- **AND** 即使用户反馈预览不足 10 条，该链接 MUST 仍可见

#### Scenario: 卡片相对位置

- **WHEN** 管理员查看登录后的 Hub 主列卡片
- **THEN** 用户反馈卡片 MUST 出现在设备记录卡片下方，且 MUST 在事件管理等后续卡片之前（相对本变更前的「问答库之后」位置已上移）

### Requirement: 反馈问题列 MUST 顶对齐展示

运维 Hub 反馈预览表与 `/device/admin/feedback-records` 全页反馈表中，「问题」列单元格 MUST 使用顶对齐（`vertical-align: top` 或等价），MUST NOT 使用导致多行问题相对单元格纵向居中的样式。其它列 MUST NOT 因本要求改变既有对齐行为。

#### Scenario: 多行问题顶对齐

- **WHEN** 某条反馈问题文本换行超过一行
- **THEN** 问题列内容 MUST 从单元格顶部开始排布，MUST NOT 垂直居中于单元格

### Requirement: 使用统计页设备号 MUST 可跳转设备数据运维

`api-usage-stats.html` 与 `client-usage-stats.html` 用户维度列表中的非空 `deviceNo` MUST 渲染为可点击链接，导航至 `/device/admin/history/{deviceNo}`（deviceNo MUST URL 编码）。空设备号 MUST NOT 渲染为有效运维链接。链接样式 SHOULD 与 Hub 设备记录 `.device-link` 一致。

#### Scenario: 点击设备号进入运维页

- **WHEN** 管理员在功能使用统计或客户端使用统计的用户列表中点击非空设备号
- **THEN** 浏览器 MUST 导航至同源 `/device/admin/history/{deviceNo}`

#### Scenario: 空设备号不可点

- **WHEN** 某行 `deviceNo` 为空
- **THEN** 该行设备号单元格 MUST NOT 提供指向 history 运维页的有效链接

### Requirement: 用户维度列表 MUST 展示序号且支持本页宝宝名筛选与最近使用排序

两页使用统计的「按用户」列表 MUST：

1. 在表格最前列展示序号，值为当前**渲染结果**从上到下的 1…N；序号 MUST NOT 作为排序键写入服务端。
2. 提供「有无宝宝名」筛选（全部 / 有 / 无），依据响应项 `babyName` trim 后是否非空，且 MUST 仅作用于**当前已加载页**的行集。
3. 展示「最近使用时间」列（格式化展示 `lastAt`；为 0 时显示空或「—」）。
4. 支持点击「最近使用时间」表头对当前已加载页按 `lastAt` 降序重排；未启用该排序时 MUST 按 `wxId`（`id`）升序。
5. 重排或筛选后序号 MUST 按新的可视顺序重新编号。

#### Scenario: 默认按 wxId

- **WHEN** 管理员打开用户维度且未点击最近使用时间排序
- **THEN** 列表行 MUST 按 wxId 升序展示（在服务端返回顺序基础上由前端保证），序号从 1 递增

#### Scenario: 本页按最近使用时间排序

- **WHEN** 管理员点击「最近使用时间」表头
- **THEN** 当前页已加载行 MUST 按 `lastAt` 降序排列，且序号 MUST 按新顺序从 1 重新编号

#### Scenario: 本页筛有宝宝名

- **WHEN** 管理员选择「有宝宝名」且当前页存在 `babyName` 为空与非空的行
- **THEN** 表格 MUST 仅展示 `babyName` 非空的行，且 MUST NOT 因此自动请求其它页
