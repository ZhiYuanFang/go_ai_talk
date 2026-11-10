## Why

用户划掉锁屏 Live Activity 后，APNs `update` 仍常返回成功，服务端继续持有脏 `activity` token。既有「update 注入 ack + 短超时 failover」依赖原生免鉴权回执与进程内 `AfterFunc`，成本高且易误伤。产品接受划掉到下次打开 App 之间的空窗；改为由已登录 Flutter 在本机无 Live Activity 时主动清服务端 activity token，使后续推送走 `start`。

## What Changes

- **BREAKING（移除 ack 链路）**：删除 `LIVE_ACTIVITY_ACK_TIMEOUT_MS`、ContentState.`ack` 注入、`last_sent_ack` / `last_acked_ack` / `ack_sent_at` 读写、`AfterFunc` 超时与下次 update 前补刀；删除 `POST /app/api/push/live-activity/ack`（含 gateway Bearer 豁免与 usage skip 条目）。
- **Schema**：从 `push_live_activity_token` **DROP** ack 三列；**保留**表级 `updated_at`（upsert / 去重仍依赖）。
- **新鉴权接口**：新增 `POST /app/api/push/live-activity/activity/clear`（**必须 Bearer**；身份走 `X-Internal-Wx-Id`）；body 含 `deviceKey`；仅删除该 wx 该机 `token_kind=activity` 行，**不删** `push_to_start`；无行幂等成功；**不同拍**发 start（下次 upsert 自然走 start）。
- **保留**：`invalidToken → 删 activity → 同机 push_to_start start`；LA `register` 免鉴权不变。
- **产品确认**：划掉到打开 App 的空窗可接受。
- **兄弟仓**：`flutter_ai_talk` twin `live-activity-clear-activity-on-app-open` 已建全套制品（去掉 ack；打开 App 时无 LA 则调 clear）。摘要见同目录 `flutter-twin-proposal.md`。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `unified-live-activity-push`：移除「update ack 超时 failover」与 ack HTTP 要求；新增「已鉴权 clear activity token」要求；明确 ContentState 不再携带 `ack`；表级 `updated_at` 保留。

## Impact

- `internal/services/push`：删 `push_live_activity_ack.go` 及相关 dispatch/schema/entity/upsert 字段；新增 clear 业务。
- `api/v1` + `internal/controller/push`：删 ack 路由；新增 clear 路由。
- `gateway-app`：从 `gateway_app_auth_exempt.go` 移除 ack path；clear **不得**加入豁免。
- `usagestats/maintenance_skip.go`：移除 ack 条目；clear 是否计入 usage **须先问负责人**（未答复前不得假定 skip）。
- 环境变量 / compose / `.env*`：去掉 `LIVE_ACTIVITY_ACK_TIMEOUT_MS`。
- 重部署：`push-service`、`gateway-app-server`（豁免表变更）；env 同步去掉 ack 超时变量。
- 对照基线：`openspec/specs/v3.0.3`；承接未归档变更 `live-activity-update-ack-timeout` 的行为回滚与替换。
- Flutter twin：见 `flutter-twin-proposal.md`。
