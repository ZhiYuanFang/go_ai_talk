# Spec Delta

## ADDED Requirements

### Requirement: gateway-app-server SHALL expose history SSE endpoint

`gateway-app-server` MUST register `GET /device/app/sse/history` (or the path locked with Flutter) that responds with `Content-Type: text/event-stream` after successful JWT Bearer authentication. The handler MUST NOT perform WebSocket Upgrade on this path. `gateway-app-server` MUST 注册历史 SSE 端点；鉴权成功后 MUST 返回 `text/event-stream`；MUST NOT 在该路径 Upgrade WebSocket。

#### Scenario: 合法 Bearer 打开 SSE

- **WHEN** 客户端携带有效 `Authorization: Bearer` 请求 `GET /device/app/sse/history` 且 `Accept` 含 `text/event-stream`
- **THEN** 响应 MUST 为 HTTP 200 与 `text/event-stream`
- **AND** 连接 MUST 保持打开直至客户端断开或服务端结束流

#### Scenario: 无/无效 Bearer

- **WHEN** 请求缺少有效 Bearer
- **THEN** 响应 MUST 为 HTTP 401（或与 App HTTP 一致的未授权语义）
- **AND** MUST NOT 建立 SSE 流

### Requirement: History SSE SHALL bind device_no like history WebSocket

After JWT validation, the history SSE subscription MUST associate the connection with the device identity: query/header `deviceNo` (lowerCamelCase) MUST match the access JWT `device_no` claim when the claim is non-empty. Mismatch MUST reject the subscription. 历史 SSE MUST 按与历史 WS 等价规则绑定 `device_no`；声明非空时与请求 deviceNo 不一致 MUST 拒绝。

#### Scenario: deviceNo 与 JWT 一致

- **WHEN** JWT `device_no` 为 D 且请求 `deviceNo=D`
- **THEN** 服务端 MUST 接受订阅并将该连接加入 D 的推送集合

#### Scenario: deviceNo 与 JWT 不一致

- **WHEN** JWT `device_no` 为 D 且请求 `deviceNo=E`（E≠D）
- **THEN** 服务端 MUST 拒绝订阅（4xx）
- **AND** MUST NOT 向该连接推送 D 的历史事件

### Requirement: History SSE SHALL fan out the same Redis-sourced events as history WS

When history-service publishes a history mutation to the agreed Redis channel, gateway-app-server MUST push an SSE event to all authenticated history SSE subscribers whose bound `device_no` matches, using JSON data compatible with existing history WebSocket application frames (`action` + `payload`). history-service Redis 推送 MUST 扇出到匹配 device_no 的历史 SSE 订阅者；data JSON MUST 与现网 history WS 业务帧兼容。

#### Scenario: create 推送

- **WHEN** Redis 收到某 `device_no` 的 create 消息且存在该 device 的 SSE 订阅
- **THEN** 网关 MUST 向这些订阅写入至少一条含 create 语义的 SSE data 帧

### Requirement: History SSE MUST NOT count as WebSocket usage and MUST avoid proxy buffering

History SSE requests MUST follow gateway usage exclusions appropriate for long-lived streams (MUST NOT be mis-counted as successful short HTTP if existing rules exclude WS upgrades—document whether SSE is counted). Reverse-proxy responses for this path MUST disable response buffering where applicable so events flush promptly. 历史 SSE 长连接的统计与反代缓冲策略 MUST 明确；适用处 MUST 禁用缓冲以保证及时刷出。

#### Scenario: 事件及时到达

- **WHEN** 网关向 SSE 订阅写入一帧业务 data
- **THEN** 在未缓冲配置下客户端 MUST 能在短时间内收到该帧（联调可观察）

### Requirement: Adding history SSE MUST NOT break existing history WebSocket clients

本变更 MUST 为 additive：`gateway-app-server` MUST 继续提供 `/device/app/ws/history`，且 MUST 保持现网握手与推送契约（GET Bearer 豁免、首帧 auth/`auth_ok`、ping/pong、Redis 原文下行含 `device_no`+`action`+`payload`）。MUST NOT 从豁免表移除该路径；MUST NOT 为 SSE 包装改写 WS 下行。Adding history SSE MUST NOT change existing history WebSocket client-visible behavior.

#### Scenario: 未升级客户端仍可握手并鉴权

- **WHEN** 未升级 App 对 `/device/app/ws/history` 发起 WebSocket Upgrade 且未携带 HTTP `Authorization`
- **THEN** 握手 MUST 成功（路径仍在 Bearer 豁免）
- **AND** 首帧合法 `auth` 后 MUST 收到 `auth_ok`（或与现网等价的成功语义）

#### Scenario: 未升级客户端仍可收到增量

- **WHEN** 某 `device_no` 的 history WS 已认证订阅，且 Redis 发布该设备的 create/update/delete
- **THEN** 网关 MUST 仍向该 WS 连接写入与现网同形的业务 JSON（含顶层 `device_no`、`action`、`payload`）

#### Scenario: 豁免表不误伤 WS

- **WHEN** 本变更合并后检查 `gatewayAppAuthExemptExactGET`（或等价豁免清单）
- **THEN** `/device/app/ws/history` MUST 仍在清单中
- **AND** `/device/app/sse/history` MUST NOT 在 Bearer 豁免清单中

### Requirement: SSE proxy buffering disable MUST be path-scoped

禁用响应缓冲的反代/网关配置（若有）MUST 限定于历史 SSE 路径（或等价精确 location），MUST NOT 以「全局关闭 buffering」作为本变更的默认部署方式，以免改变其它 HTTP/WS 入口特性。SSE proxy buffering disable MUST be scoped to the history SSE path.

#### Scenario: 仅 SSE location 禁缓冲

- **WHEN** 运维按本变更部署清单配置反代
- **THEN** 禁缓冲规则 MUST 仅匹配历史 SSE 路径（或文档明示的等价精确集合）
- **AND** MUST NOT 要求全局 `proxy_buffering off` 作为唯一合规方式
