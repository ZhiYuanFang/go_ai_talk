## 1. 下线 ack 超时与字段

- [x] 1.1 从 `push_live_activity.go` dispatch 移除 ack 注入、`MarkLiveActivityAckSent`、`scheduleLiveActivityAckTimeout`、`failoverStaleActivityIfAckTimedOut`
- [x] 1.2 删除 `push_live_activity_ack.go`（含 `LIVE_ACTIVITY_ACK_TIMEOUT_MS` 读取与 AfterFunc）
- [x] 1.3 实体 / `RegisterLiveActivityToken` upsert 去掉 `last_sent_ack` / `last_acked_ack` / `ack_sent_at`；**保留** `updated_at`
- [x] 1.4 schema：CREATE 不再建 ack 三列；ensure 幂等 `ALTER DROP` 三列；**禁止** DROP `updated_at`

## 2. 删除 ack HTTP 与网关条目

- [x] 2.1 删除 `api/v1` ack Req/Res 与 `AppPushCtrl.LiveActivityAckPost`
- [x] 2.2 从 `gateway_app_auth_exempt.go` 移除 `/app/api/push/live-activity/ack`
- [x] 2.3 从 `usagestats/maintenance_skip.go` 移除 ack 条目
- [x] 2.4 删除 compose / `.env.example` / `.env.prod` 等中的 `LIVE_ACTIVITY_ACK_TIMEOUT_MS`

## 3. 新增鉴权 clear 接口

- [x] 3.1 实现按 `(wxId, deviceKey)` 删除 `token_kind=activity` 的服务方法（无行幂等；不删 push_to_start；不同拍 start）
- [x] 3.2 新增 `POST /app/api/push/live-activity/activity/clear`（`api/v1` + controller）；身份 `X-Internal-Wx-Id`；body `deviceKey`
- [x] 3.3 确认 gateway `/app/api/push/*` 反代已覆盖；**不得**把 clear 加入 Bearer 豁免
- [x] 3.4 **先问负责人** clear 是否计入 usage；仅在明确答复后决定是否改 `maintenance_skip`（未答复不得假定）

## 4. 联调与验收

- [x] 4.1 与 Flutter twin `live-activity-clear-activity-on-app-open` 对齐（参考 `flutter-twin-proposal.md`）：无 LA 打开 App → clear → 下次 upsert 为 start
- [x] 4.2 确认 update APNs 体无 `ack`；划掉后 invalidToken 自愈仍可用
- [x] 4.3 重部署清单：`push-service`、`gateway-app-server`；同步去掉 ack 超时 env
