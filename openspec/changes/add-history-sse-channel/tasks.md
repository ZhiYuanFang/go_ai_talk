# Tasks

## 1. SSE 端点与鉴权

- [x] 1.1 在 `gateway-app-server` 注册 `GET /device/app/sse/history`，写入 `text/event-stream` 必要响应头（含禁缓冲）；确认路径**不在** Bearer 豁免名单；用 curl/浏览器确认 200 长连接
- [x] 1.2 实现 Bearer JWT 校验 + `deviceNo` 与 JWT `device_no` 对齐；用无效 token / 错 deviceNo 验证 401/4xx，合法请求保持长连接

## 2. 扇出与心跳（additive：不改 WS 写路径）

- [x] 2.1 在 Redis 订阅回调中为 SSE 增加扇出（并列集合或独立 hub）；WS 侧继续对匹配连接 `BroadcastText` **原文**；验证 SSE create/update/delete 的 data 与现网 WS 业务字段兼容，且未改 history-service 发布体
- [x] 2.2 增加 SSE 空闲心跳（comment 或 ping event）；确认实现未改写 `gatewayAppHistoryWS` 首帧 auth / `auth_ok` / ping-pong 逻辑

## 3. 旧客户端兼容回归

- [x] 3.1 确认 `/device/app/ws/history` 仍在 `gatewayAppAuthExemptExactGET`（或等价清单），且 `/device/app/sse/history` 不在豁免表；无 HTTP Bearer 的 WS Upgrade 仍可握手
- [x] 3.2 用未升级客户端路径（或等价脚本）连接 history WS：auth → 收到 create/update/delete 帧，且 JSON 仍含顶层 `device_no`、`action`、`payload`
- [x] 3.3 核对部署清单：反代禁缓冲仅绑定 SSE 路径（或精确 location），非全局关闭 buffering

## 4. 新客户端联调与校验

- [ ] 4.1 与 Flutter `history-sse-retire-ucg-clinic-ws` 联调：客户端仅 SSE 可收到历史增量
- [x] 4.2 `openspec validate add-history-sse-channel --strict` 通过

## Workflow follow-up

- Flutter 硬切稳定后，另开变更考虑下线 `/device/app/ws/history`。
- 实现完成后归档本 change。
