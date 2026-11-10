## REMOVED Requirements

### Requirement: Live Activity update MUST carry ack and failover to start on ack timeout

**Reason**: 产品接受划掉到打开 App 的空窗；改为客户端鉴权 clear activity token，不再做 update 注入 ack / 短超时 failover。

**Migration**: 删除 ContentState.`ack`、超时 env、AfterFunc 与补刀；由 `POST /app/api/push/live-activity/activity/clear` 与 Flutter 打开 App 检测替代。

### Requirement: Live Activity ack HTTP API, gateway exempt, excluded from usage

**Reason**: ack 回执接口随超时 failover 一并下线。

**Migration**: 删除 `POST /app/api/push/live-activity/ack`、gateway Bearer 豁免与 usage skip 条目；Flutter 改为调用鉴权 clear。

## ADDED Requirements

### Requirement: Authenticated clear of Live Activity activity token per device

The system MUST expose `POST /app/api/push/live-activity/activity/clear` for App clients. The request MUST require Bearer authentication (gateway-app MUST NOT add this path to the Bearer auth exempt list). Request identity MUST come from the authenticated internal wx header (`X-Internal-Wx-Id`), not from an untrusted body `wxId`. The request body MUST include a non-empty `deviceKey`. On success, the system MUST delete the persisted `token_kind=activity` row for `(wx_id, device_key)` when present, and MUST treat zero matching rows as success (idempotent). The system MUST NOT delete `token_kind=push_to_start` rows for that device as part of this API. The system MUST NOT send Live Activity `event=start` in the same clear request; subsequent Live Activity upsert/dispatch MUST resolve to `start` when no activity token remains for that user device path per existing upsert rules. The reverse proxy MUST continue to cover this path under the existing `/app/api/push/*` binding.

系统 MUST 提供鉴权 `POST /app/api/push/live-activity/activity/clear`；MUST NOT 豁免 Bearer；身份 MUST 来自鉴权后的内部 wx 头；MUST 按 `(wx, deviceKey)` 删 activity 行且无行幂等成功；MUST NOT 删 push_to_start；MUST NOT 同拍 start。

#### Scenario: Clear deletes activity token for device

- **WHEN** an authenticated user POSTs clear with `deviceKey=D` and an activity token row exists for that wx and `D`
- **THEN** that `token_kind=activity` row MUST be deleted
- **AND** any `token_kind=push_to_start` row for the same device MUST remain

#### Scenario: Clear is idempotent when no activity token

- **WHEN** an authenticated user POSTs clear with `deviceKey=D` and no activity token exists for that wx and `D`
- **THEN** the system MUST respond successfully without erroring solely due to absence of the row

#### Scenario: Clear does not start Live Activity

- **WHEN** clear successfully deletes (or finds no) activity token
- **THEN** the system MUST NOT send APNs `event=start` as part of handling that clear request

#### Scenario: Gateway requires Bearer for clear

- **WHEN** a client calls `POST /app/api/push/live-activity/activity/clear` without Bearer
- **THEN** gateway-app MUST NOT treat the path as auth-exempt
- **AND** the unauthenticated call MUST NOT reach successful clear semantics

### Requirement: Live Activity update ContentState MUST NOT include ack delivery fields

When the system sends Live Activity APNs with `event=update` (including upsert resolved to update), the ContentState MUST NOT include an `ack` delivery-receipt field, and the system MUST NOT persist `last_sent_ack` / `last_acked_ack` / `ack_sent_at` on `push_live_activity_token`. The table-level `updated_at` column MUST remain for register upsert and deduplication. The system MUST NOT schedule ack-timeout delete-and-restart after update send success. Existing invalid-token delete-and-restart behavior for update MUST remain unchanged by this requirement.

update MUST NOT 再注入/依赖 ack；MUST NOT 保留 ack 三列；MUST 保留 `updated_at`；MUST NOT 再跑 ack 超时 failover；invalidToken 自愈 MUST 保持。

#### Scenario: Update payload has no ack field

- **WHEN** the system successfully prepares an APNs Live Activity update ContentState
- **THEN** the ContentState MUST NOT contain an `ack` key used for delivery receipt

#### Scenario: No ack timeout restart after update

- **WHEN** an update is sent successfully
- **THEN** the system MUST NOT start an in-process ack-timeout timer that deletes the activity token solely for missing client ack
