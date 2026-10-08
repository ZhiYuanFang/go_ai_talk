## Why

运维需在 Hub 设备详情页（`/device/admin/history/{deviceNo}`）为宝宝设备登记/管理小智 MCP 音箱绑定。后端 App CRUD 已由 `xiaozhi-mcp-multi-device-binding` 提供；缺的是详情页入口。通过复用页内 **App 用户登录**（与 AI 调试台同模式）取得 `wx_id`，直接调用既有 App 绑定 API，避免再造 Admin CRUD 与 wx 归属歧义。

## What Changes

- 在 `history.html` 设备详情页新增「小智 MCP 绑定」区块：未 App 登录时禁用绑定操作；登录后支持列表 / 添加（token+备注）/ 改备注 / 删除。
- **复用**现有 `/device/app/api/xiaozhi-mcp/bindings*`（Bearer = App JWT）；列表仅展示脱敏 `tokenMask`，不展示完整 token。
- 登录账号绑定的 `deviceNo` 与当前页 `deviceNo` 不一致时 MUST 告警，添加绑定 SHOULD 拒绝（或与调试台同等严格策略，见 design）。
- **不新增** `/device/admin/api/xiaozhi-mcp/*`；不改 Flutter；不改 Manager / 表结构。
- 无新 App 路由 → **无需** usage 统计确认；不改 `maintenance_skip.go`。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-hub-binding-ui`：Hub 设备详情页小智绑定 UI 与 App 登录门禁、调用既有 App API、设备号一致性校验与脱敏展示。

### Modified Capabilities

- （无；不修改既有 App 绑定 REQUIREMENTS。）

## Impact

- **代码**：主要为 `resource/public/history.html`（及必要时抽取与调试台共享的 App token 辅助逻辑）；网关反代已覆盖 App `xiaozhi-mcp` 路径。
- **依赖**：依赖已上线的 App 小智绑定 API 与 `xiaozhi-mcp-service` 写路径通知。
- **运维**：操作者需知道宝宝账号的 App 用户名密码；列表范围为**当前登录 wx** 的绑定，非设备上全部 wx 的跨账号视图。
