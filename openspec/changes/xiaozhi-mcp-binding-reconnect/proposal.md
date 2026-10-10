## Why

已绑定小智音箱在列表中显示 `connected=false`（断线或退避重拨中）时，用户与 Hub 无法主动触发立即重拨：现有 Upsert 在 token/deviceNo/wxId 未变时为 noop，只能等待 Manager 内置退避。需要一条专用「重连」路径，在不动 Python、不改既有 List/Add 契约的前提下，让 Flutter 与 Hub 可一键强制取消会话并立即 redial。

## What Changes

- 新增 App API：`POST /device/app/api/xiaozhi-mcp/bindings/{id}/reconnect`（经 gateway-app 反代 + Bearer；**不**加入 auth exempt）。
- device-service：校验绑定归属当前 wx，加载 token/deviceNo/wxId，经 `clients/xiaozhimcp` 调用 mcpbridge 专用内部重连接口。
- `clients/xiaozhimcp` + `mcpbridge`：新增内部 `POST .../bindings/reconnect`（或等价 ForceRestart）；Manager **MUST** cancel 现会话 + `startLocked`，即使 key 未变；重置退避并立即 redial；接口快速返回，客户端再刷列表看 `connected`。
- Hub：`resource/public/history.html` 小智区在 `!connected` 时展示「重连」并调用上述 App API。
- **不**修改既有 v1 List/Add 请求/响应结构；仅新增 reconnect 类型。
- **不**改 Python；**不**引入 Redis；**不**新增后台 ticker。
- Flutter 客户端重连按钮：本仓不做，作为兄弟仓后续跟进（见 Impact）。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-binding-reconnect`：App/device/clients/mcpbridge 强制重连契约与归属校验、错误语义、与 entitlement 的交互说明。
- `xiaozhi-mcp-hub-reconnect-ui`：Hub 设备详情页小智区在断连时展示并调用重连。

### Modified Capabilities

- （无）基线 `openspec/specs/` 为版本归档，无独立 capability 名需 delta；本变更以新增 capability 规格描述行为。

## Impact

- **Go / 本仓**：`api/v1` 新增 reconnect DTO；`internal/controller/device`、`internal/services/device`；`internal/clients/xiaozhimcp`；`internal/services/mcpbridge`（InternalHTTP + Manager ForceRestart）；`resource/public/history.html`；gateway-app 反代/exempt/usage 自检（usage 是否计入须向负责人确认后再动 `maintenance_skip`）。
- **Flutter（兄弟仓）**：API 落地后另开变更做「断连显示重连」；**本提案不实现 Flutter**。
- **Python**：无；不要求、不依赖 Python 改动。
- **依赖**：device → `clients/xiaozhimcp` HTTP；禁止 device 直 import `services/mcpbridge`。
