## ADDED Requirements

### Requirement: 详情页小智绑定依赖 App 登录

Hub 设备详情页（`/device/admin/history/{deviceNo}`）SHALL 提供「小智 MCP 绑定」区块。未完成 App 用户登录时，MUST NOT 允许发起添加/改备注/删除绑定请求；MUST 提示需先 App 登录。

#### Scenario: 未登录禁用写操作

- **WHEN** 操作者仅持有 Hub Admin 会话、未 App 登录
- **THEN** 小智绑定添加/改备注/删除控件 MUST 不可用或提交被前端拒绝
- **AND** MUST 展示需 App 登录的说明

#### Scenario: App 登录后可操作

- **WHEN** 操作者在详情页完成 App 用户名密码登录并取得 App JWT
- **THEN** 小智绑定区块 MUST 允许调用绑定列表与写操作（在设备号一致前提下）

### Requirement: 复用 App 小智绑定 API

小智绑定 UI MUST 使用既有 App 接口完成 CRUD，且请求携带 App Bearer（非 Admin JWT）：

- 列表：`GET /device/app/api/xiaozhi-mcp/bindings`
- 添加：`POST /device/app/api/xiaozhi-mcp/bindings`（`mcpToken`、`alias`）
- 改备注：`PUT /device/app/api/xiaozhi-mcp/bindings/{id}/alias`
- 删除：`DELETE /device/app/api/xiaozhi-mcp/bindings/{id}`

MUST NOT 新增依赖 Admin JWT 的小智绑定业务 API 作为本能力的实现路径。

#### Scenario: 添加绑定

- **WHEN** App 已登录、设备号一致，操作者提交非空 mcpToken 与 alias
- **THEN** 前端 MUST 以 App Bearer 调用 POST bindings
- **AND** 成功后刷新列表

#### Scenario: 删除绑定

- **WHEN** App 已登录，操作者删除列表中某条绑定
- **THEN** 前端 MUST 以 App Bearer 调用 DELETE bindings/{id}
- **AND** 成功后自列表移除或刷新

### Requirement: Token 脱敏展示

列表与详情展示 MUST 使用接口返回的脱敏字段（如 `tokenMask`）；MUST NOT 在页面上持久展示或回显完整 `mcpToken`（添加表单输入框在提交前的明文输入除外）。

#### Scenario: 列表无全文 token

- **WHEN** 绑定列表加载成功
- **THEN** 每行展示 alias 与 tokenMask（或等价脱敏）
- **AND** 页面 DOM/文案 MUST NOT 包含完整 mcpToken

### Requirement: 登录设备与页内设备一致

App 登录后，系统 MUST 比较登录账号绑定的 `deviceNo` 与当前详情页 `deviceNo`。不一致时 MUST 向操作者告警，且 MUST 禁止添加小智绑定（避免串号落库）。

#### Scenario: 串号禁止添加

- **WHEN** App 登录成功但账号 deviceNo 与 URL 中 deviceNo 不同
- **THEN** UI MUST 提示不一致
- **AND** MUST NOT 成功发起添加绑定请求

#### Scenario: 一致允许添加

- **WHEN** App 登录账号 deviceNo 与当前页 deviceNo 相同（忽略首尾空白）
- **THEN** 在已登录前提下允许添加绑定
