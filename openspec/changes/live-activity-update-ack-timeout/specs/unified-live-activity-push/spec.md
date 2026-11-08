## ADDED Requirements

### Requirement: Live Activity update MUST carry ack and failover to start on ack timeout

When the system sends Live Activity APNs with `event=update` (including when `action=upsert` resolved to update because an activity token existed), it MUST inject a temporary opaque `ack` value into the APNs ContentState for that send, and after a successful send MUST persist that value as the activity token row's `last_sent_ack` (with send time as needed for recovery). The system MUST wait for the client to acknowledge the same `ack` via the Live Activity ack HTTP API. If the acknowledgment for that `last_sent_ack` is not recorded before the configured timeout (`LIVE_ACTIVITY_ACK_TIMEOUT_MS`, default 3000ms) elapses, the system MUST delete that `token_kind=activity` row and MUST attempt to send `event=start` using a `token_kind=push_to_start` token that shares the **same `device_key`**, reusing the same ContentState/Attributes/alert semantics as the existing invalid-token restart path. The system MUST NOT apply this ack-timeout failover to `event=start` or `event=end`. Late acknowledgments after the activity row was deleted MUST be ignored. When preparing a subsequent update, if `last_sent_ack` differs from `last_acked_ack` and the send time already exceeds the timeout (e.g. in-process timer was lost), the system MUST perform the same delete-and-restart failover before relying on the stale activity token.

update（含 upsert→update）MUST 注入临时 ack、send_ok 后等待回执；超时未 ack MUST 删 activity 并用同 device_key 的 push_to_start start；start/end MUST NOT 走该超时；丢 timer 时下次 update 前 MUST 按库内状态补刀。

#### Scenario: Update acked in time keeps activity token

- **WHEN** an update is sent successfully with ContentState `ack=A` and the client posts ack `A` for that `device_key` before the timeout
- **THEN** the activity token row MUST remain
- **AND** the system MUST NOT start a push_to_start failover for that send

#### Scenario: Update ack timeout restarts via push_to_start

- **WHEN** an update is sent successfully with `ack=A` and no matching ack is recorded before `LIVE_ACTIVITY_ACK_TIMEOUT_MS`
- **THEN** that activity token row MUST be deleted
- **AND** the system MUST send Live Activity start via the push_to_start token for the same `device_key`

#### Scenario: Start and end are not subject to ack timeout failover

- **WHEN** the system sends `event=start` or `event=end`
- **THEN** the system MUST NOT schedule ack-timeout delete-and-restart for that send

#### Scenario: Stale sent ack recovered on next update

- **WHEN** an activity row has `last_sent_ack != last_acked_ack` and the recorded send time is already past the ack timeout
- **AND** the system is about to send another update to that token
- **THEN** the system MUST delete the activity row and attempt same-`device_key` push_to_start restart instead of trusting the stale update token

### Requirement: Live Activity ack HTTP API, gateway exempt, excluded from usage

The system MUST expose `POST /app/api/push/live-activity/ack` accepting `wxId`, `deviceKey`, and `ack`. When the request `ack` equals the current `last_sent_ack` for the matching `(wx_id, device_key, token_kind=activity)` row, the system MUST set `last_acked_ack` to that value; acknowledgments that do not match `last_sent_ack` MUST NOT overwrite `last_acked_ack` with a stale value. Gateway-app MUST reverse-proxy the path under the existing `/app/api/push/*` binding and MUST add the exact path to the Bearer auth exempt list **aligned with** `/app/api/push/live-activity/register`. The ack endpoint MUST be excluded from App API usage statistics (same maintenance-skip policy as live-activity register).

ack 接口 MUST 仅在 ack 等于当前 last_sent 时更新 last_acked；gateway MUST 精确豁免该 path（对齐 register）；MUST 不计入 usage。

#### Scenario: Matching ack updates last_acked

- **WHEN** an activity row has `last_sent_ack=A` and the client posts ack with `wxId`, `deviceKey`, and `ack=A`
- **THEN** the row's `last_acked_ack` MUST become `A`

#### Scenario: Mismatched ack does not rewind last_acked

- **WHEN** an activity row has `last_sent_ack=B` and the client posts `ack=A` (not equal to B)
- **THEN** the system MUST NOT set `last_acked_ack` to `A`

#### Scenario: Gateway and usage policy match register

- **WHEN** a client calls `POST /app/api/push/live-activity/ack` without Bearer (as with register)
- **THEN** gateway-app MUST allow the request through the push proxy auth exempt list
- **AND** the call MUST NOT be counted in App API usage statistics
