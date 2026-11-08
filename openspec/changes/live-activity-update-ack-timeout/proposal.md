## Why

用户划掉锁屏 Live Activity 后，APNs 对后续 `update` 仍常返回成功，服务端继续以为 activity token 有效，锁屏空白直到下次偶然清脏行。仅靠 `invalidToken` 自愈覆盖不了「静默失败」。需要在 **update 发出后**由 iOS 原生层回执 ack；短超时未回执则同拍改走 `push_to_start` `start`。

## What Changes

- **update 送达回执**：每次 APNs `event=update`（含 `upsert→update`）成功发出前，向 ContentState 注入临时不透明 `ack`；`send_ok` 后写入该 activity 行的 `last_sent_ack`（及发送时刻），并启动可配置超时等待客户端回传。
- **ack App 接口**：新增 `POST /app/api/push/live-activity/ack`（body：`wxId` / `deviceKey` / `ack`）；仅当回传 `ack` 等于该机 activity 行当前 `last_sent_ack` 时写入 `last_acked_ack`。
- **超时 failover（仅 update）**：超时仍 `last_acked_ack != last_sent_ack` 时，删除该 `activity` 行，并对同 `device_key` 的 `push_to_start` 发 `start`（复用本次 ContentState/Attributes；与现有 invalidToken 重启出口一致）。**start / end 路径 MUST NOT** 走该超时逻辑。
- **超时可配**：`LIVE_ACTIVITY_ACK_TIMEOUT_MS`，默认 `3000`。
- **gateway-app**：ack 路径经既有 `/app/api/push/*` 反代；Bearer 豁免 **对齐 register**（精确 path 加入 `gateway_app_auth_exempt`）；**不计入 usage**（`maintenance_skip` 与 register 同策略，负责人已确认）。
- **表字段**：`push_live_activity_token` 为 activity 行增加 `last_sent_ack` / `last_acked_ack`（及实现所需的发送时间字段，若用 DB 补刀丢 timer）。
- **非目标**：不改 start 语义与触发矩阵（仍仅预测 PUT / 到点）；不要求 start 等 ack；不改 lead 窗 / items≤3；不放宽 invalidToken 判定。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `unified-live-activity-push`：补充「update 注入 ack、客户端回执、超时未 ack 则删 activity 并同机 start」；明确 ack HTTP、网关豁免与 usage 排除；start/end 不受该超时约束。

## Impact

- `internal/services/push`：schema、dispatch update 注入 ack / 定时 failover、ack 写入、与 `restartLiveActivityViaPushToStart` 复用。
- `api/v1` + `internal/controller/push`：新 ack 路由；gateway `gateway_app_auth_exempt.go`、`usagestats/maintenance_skip.go`。
- 客户端（仓库外）：iOS 原生在 App 未启动时于超时内回传 ack。
- 环境变量：`LIVE_ACTIVITY_ACK_TIMEOUT_MS`（manifest/docker env 按需登记）。
