## ADDED Requirements

### Requirement: Invalid activity update token MUST be deleted and restarted via push-to-start

When the system sends Live Activity APNs with `event=update` (including when `action=upsert` resolved to update because an activity token existed), and the send result indicates an **invalid token** per the existing APNs invalid-token rules (e.g. HTTP 410 Gone, BadDeviceToken, Unregistered, DeviceTokenNotForTopic), the system MUST delete that persisted `token_kind=activity` row, and MUST attempt in the **same dispatch** to send `event=start` using a `token_kind=push_to_start` token that shares the **same `device_key`** as the deleted activity row, reusing the same ContentState (and Attributes/alert fields as for a normal start). Temporary APNs or transport failures that are **not** classified as invalid token MUST NOT delete the activity token and MUST NOT trigger this restart. If no matching push-to-start token exists, the system MUST log and skip the restart (MUST NOT invent tokens). A failed restart start send MUST NOT loop retry within the same dispatch. When the original send event is `end`, invalid token cleanup MAY delete the activity row, but the system MUST NOT restart via push-to-start.

当 update（含 upsert→update）遇 invalidToken：MUST 删该 activity 行，并同拍用同 device_key 的 push_to_start 发 start；非 invalid 失败 MUST NOT 删/重启；end MUST NOT 因此 start。

#### Scenario: Update invalid then restart on same device

- **WHEN** upsert/update sends to an activity token and APNs returns an invalid-token outcome
- **THEN** that activity token row MUST be deleted
- **AND** the system MUST send Live Activity start via the push_to_start token for the same `device_key` with the same ContentState

#### Scenario: Temporary update failure does not restart

- **WHEN** update fails with a non-invalid outcome (e.g. rate limit or server error)
- **THEN** the activity token MUST remain
- **AND** the system MUST NOT send a push_to_start restart in that dispatch

#### Scenario: Invalid end does not start

- **WHEN** end is sent to an activity token and APNs returns invalid-token
- **THEN** the system MAY delete the invalid activity token
- **AND** MUST NOT send push_to_start start as a failover for that end

#### Scenario: No push_to_start for device

- **WHEN** update is invalid and the activity row is deleted but no push_to_start exists for that `device_key`
- **THEN** the system MUST NOT send start
- **AND** MUST leave recovery to a later Refresh after the client re-registers tokens
