## 1. Schema 与配置

- [x] 1.1 `push_live_activity_token` ensure：为 activity 行增加 `last_sent_ack`、`last_acked_ack`、`ack_sent_at`（或等价发送时刻）；中文注释说明仅 activity 有意义
- [x] 1.2 读取 `LIVE_ACTIVITY_ACK_TIMEOUT_MS`（默认 3000，非法回退默认，建议上限钳制）；`manifest/docker` env 示例按需登记

## 2. update 注入 ack 与超时 failover

- [x] 2.1 dispatch 在 `sendEvent=update` 临发前生成不透明 ack，写入 ContentState；`send_ok` 后持久化 `last_sent_ack` + `ack_sent_at`
- [x] 2.2 `AfterFunc(timeout)`：到期若该行 `last_acked_ack != last_sent_ack`，删 activity 并调用现有 `restartLiveActivityViaPushToStart`；打可观测日志
- [x] 2.3 准备再次 update 前：若已超时未 ack（DB 补刀），先失效重启，禁止继续对脏 activity 发 update
- [x] 2.4 确认 start/end 不注册 ack 超时；与 invalidToken 重启出口复用、互不替代

## 3. ack HTTP 与 gateway

- [x] 3.1 `api/v1` + push controller：`POST /app/api/push/live-activity/ack`（wxId/deviceKey/ack）；仅 `ack==last_sent_ack` 时更新 `last_acked_ack`
- [x] 3.2 `gateway_app_auth_exempt.go` 精确加入 `/app/api/push/live-activity/ack`（对齐 register）
- [x] 3.3 `usagestats/maintenance_skip.go` 排除 `POST /app/api/push/live-activity/ack`（负责人已确认不计入）

## 4. 校验

- [x] 4.1 `gofmt`；编译 push / gatewayapp 相关包
- [x] 4.2 自检：update 带 ack；按时 ack 不 restart；超时走 push_to_start；start/end 无超时；豁免与 usage skip 生效
