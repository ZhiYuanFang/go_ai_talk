## ADDED Requirements

### Requirement: Live Activity start MUST require at least one push_device for the wx

When the push-service resolves a Live Activity send to `event=start` (including `action=start` and `action=upsert` that selects push-to-start because no activity token exists), it MUST load that recipient `wxId`'s `push_device` rows (any channel). If there is no push device row for that wx, the system MUST skip the Live Activity start APNs send, MUST log a skip reason (e.g. `no_push_device`), and MUST NOT invent tokens. `event=update` and `event=end` MUST NOT be gated by this push_device check.

解析为 `event=start` 时 MUST 要求该 wx 至少一条 `push_device`（任意 channel），否则 skip 且 MUST NOT 发 start；update/end MUST NOT 受此闸。

#### Scenario: Start skipped when no push_device

- **WHEN** Refresh/fanout resolves to Live Activity `event=start` for a wx that has push_to_start token(s) but zero `push_device` rows
- **THEN** the system MUST NOT send APNs Live Activity start for that wx
- **AND** MUST emit an observable skip log indicating missing push device

#### Scenario: Start allowed when any push_device exists

- **WHEN** the same start path runs for a wx that has at least one `push_device` row (any channel) and a push_to_start token
- **THEN** the system MUST proceed with the existing start send path (including start `aps.alert` rules from other requirements)

#### Scenario: Update not gated by push_device

- **WHEN** fanout resolves to Live Activity `event=update` for a wx with zero `push_device` rows but a valid activity token
- **THEN** the system MUST still attempt the update send per existing update rules (MUST NOT skip solely due to missing push_device)

## REMOVED Requirements

### Requirement: Update invalid token MUST delete activity and restart via push-to-start

**Reason**: 产品确认同拍 invalid→restart 用不到；改由删脏 activity 后等待后续 upsert/clear 自然走 start（并受 push_device 闸）。

**Migration**: 删除 `restartLiveActivityViaPushToStart` 及其调用；update 仍可在 invalidToken 时删除该 activity 行，但 MUST NOT 同拍发 start。依赖既有 upsert：无 activity 时下次再 start。

#### Scenario: Update invalid deletes without restart

- **WHEN** Live Activity `event=update` returns an APNs invalid-token result
- **THEN** the system MUST delete that activity token row when applicable
- **AND** MUST NOT send a same-dispatch `event=start` via push-to-start as restart failover
