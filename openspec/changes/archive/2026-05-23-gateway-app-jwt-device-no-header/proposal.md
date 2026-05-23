## Why

gateway-app 在 Bearer 鉴权后，为向下游注入身份，当前实现会 **JWT `sub`（wx.id）→ 调 device `internal/by-id` → unionid → `X-Internal-Wx-Union-Id`**，每次受保护请求在缓存未命中时多一次 RTT。业务侧多数接口最终仍要 **device_no**（请求体/查询参数或设备域逻辑），希望 **由网关在解析 access JWT 后直接注入 `device_no`**，去掉 **id→unionid** 这条热路径，同时 **不改变 App 客户端可见的 HTTP 契约**（路径、登录/刷新 body 与 JSON 字段名保持与现网一致；JWT 仍为不透明字符串）。

## What Changes

- **gateway-app-server**：签发 access JWT 时在载荷中增加 **`device_no` 声明**（已绑定设备时非空；未绑定可为空）；Bearer 中间件 **仅本地验签解析 JWT**，向代理请求注入 **`X-Internal-Device-No`**（来自声明，空则省略或按路由策略处理），并注入 **`X-Internal-Wx-Id`**（来自标准 **`sub`**，即 wx 表主键 id，供 device 用户域写路径识别 wx 行）。**移除或停用**「每请求 **`GET .../internal/by-id` 取 unionid**」的网关热路径（及对应 Redis `id→unionid` 缓存职责可收缩或删除）。
- **device-service**：`/device/app/api/user/bindwx`、`auto_save`、`detail` 等原依赖 **`X-Internal-Wx-Union-Id`** 的接口，改为依赖 **`X-Internal-Wx-Id`**（网关从 JWT `sub` 注入，整数 wx 主键）进行 wx 行定位；**若仍需 unionid 落库/审计**，仅在 **登录换票写库路径**使用，不在网关 per-request 链上读取。
- **历史 WebSocket**：首帧仍传 **JWT**；服务端校验 **JWT 内 `device_no` 声明与首帧 `device_no` 一致**（及签名、`exp`），**不再**走「id→unionid→detail 拉 device_no」链。
- **refresh**：仍为不透明 refresh + Redis，旋转策略不变；**重新签发 access 时**必须写入 **与当前会话一致的 `device_no` 声明**（以 device 域权威或登录时缓存为准）。

**对前端/客户端的保证（非 BREAKING）**：`POST /device/app/api/login`、`POST /device/app/api/token/refresh` 的请求/响应 JSON **字段集合与语义**与当前一致（仍含 `accessToken`、`refreshToken`、`wxId`、`deviceNo`、`isNewUser` 等）；客户端 **不解析 JWT 内容**，仅存储与携带 Bearer 字符串即可。

**BREAKING（内部/运维）**：下游 device/history 若曾依赖 **`X-Internal-Wx-Union-Id`** 由网关注入，将迁移为 **`X-Internal-Device-No` + `X-Internal-Wx-Id`**；**直连 device 绕过网关** 的调用方必须同步改头（不属于 App 前端范畴）。

## Capabilities

### New Capabilities

- `gateway-app-jwt-device-no-header`：gateway-app access JWT 载荷扩展、`X-Internal-Device-No` / `X-Internal-Wx-Id` 注入、WS 校验与 unionid 热路径移除的规格说明。

### Modified Capabilities

- （本变更以 change 内增量规格为主；若后续归档到全局 `openspec/specs/`，再对 `gateway-app-server`、`device-wx-profile-apis` 做全局 delta。）

## Impact

- `internal/services/gatewayapp`：`SignAccess` / `ParseAccess*`、Bearer 中间件、`device_client`（可能删除或闲置 unionid 拉取）、历史 WS 控制器。
- `internal/controller/gateway_app_ctrl.go`：登录/刷新签发 access 时传入 **device_no**。
- `internal/controller/device_app_user.go` 与 `internal/services/device/wx.go`：由 **unionid 头** 改为 **wxId 头** 定位 wx 行；必要时保留 unionid **仅 DB 列与登录换票**，不依赖网关 per-request unionid。
- `api/v1`：OpenAPI 注释与内部头约定文档。
- 配置：可删除或闲置 `gatewayApp.wxIdUnionCacheSeconds` 等与 id→unionid 相关的网关缓存项（实现阶段清理）。
