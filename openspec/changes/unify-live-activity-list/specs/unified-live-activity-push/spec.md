## ADDED Requirements

### Requirement: System SHALL expose a single Refresh that projects open history and imminent pending into one Live Activity

The twin Go **history-service** MUST provide a shared Refresh operation (name MAY be `RefreshBabyLiveActivity`) keyed by `deviceNo` that: (1) loads the device’s current Redis imminent pending list via `cachekit`; (2) loads open timing histories (`end_time=0`) for that device from the history domain; (3) builds a list ContentState; (4) fans out Live Activity to each bound `wxId` of that device. Voice MUST invoke this Refresh only via `clients/history` (or an equivalent history internal HTTP contract). MUST NOT implement a second divergent assemble path in voice. MUST NOT import `services/voice` from history.

history-service MUST 提供按 `deviceNo` 的统一 Refresh：读 Redis imminent pending、读未闭合 history、组装列表 ContentState、向绑定 wx 扇出 Live Activity。voice MUST 仅经 `clients/history`（或 history internal HTTP）调用。MUST NOT 在 voice 再实现一套不一致的拼装。history MUST NOT import `services/voice`。

#### Scenario: Refresh builds from both sources

- **WHEN** Refresh runs for a device that has two pending predict items and one open history
- **THEN** the Live Activity ContentState `items` MUST include those predict items and that open history item

#### Scenario: Voice calls history contract

- **WHEN** voice needs to refresh Live Activity after pending sync or MQ fire
- **THEN** voice MUST call the history Refresh contract and MUST NOT import history business packages

### Requirement: Live Activity ContentState SHALL be an ordered items list

ContentState MUST carry an `items` array (name MAY be `items`). Predict items from pending MUST appear first, sorted by `nextAt` ascending (sooner first). Open timing items MUST follow, sorted by `startTime` descending. Each item MUST include a discriminant `kind` distinguishing predict vs in-progress (e.g. `imminent` / `timing`) and the fields needed for lock-screen display (at least event identity and the relevant time field). Optional structured expect fields MAY be included when available. ContentState and Attributes MUST NOT include `logo` / `color` / `logoFile` / remote image URLs. When pending is empty, the predict segment MUST be empty; when there are no open histories, the timing segment MUST be empty.

ContentState MUST 含有序列表：预测段在前且 `nextAt` 升序；进行中段在后且 `startTime` 倒序；每项含 kind 与展示所需字段；禁止视觉 URL/色值字段。

#### Scenario: Sort order

- **WHEN** pending has nextAt=100 and nextAt=50, and two open histories with startTime=10 and startTime=20
- **THEN** items MUST order predict(50), predict(100), timing(20), timing(10)

#### Scenario: Pending present means show

- **WHEN** Redis pending still contains an item whose nextAt is in the past
- **THEN** Refresh MUST still include that item until App PUT replaces or clears pending (MUST NOT drop solely because nextAt ≤ now)

### Requirement: Empty items MUST end the Live Activity; otherwise update or start by activity token

After building `items`, for each bound wx: if `items` is empty, the system MUST fan out Live Activity `action=end` for that user’s unified activity and MUST clean persisted `token_kind=activity` rows for that user as on existing end cleanup. If `items` is non-empty and the user has at least one persisted `token_kind=activity` token, the system MUST `update` with the new ContentState. If `items` is non-empty and no activity token exists, the system MUST `start` via push-to-start tokens. MUST NOT clear-then-start when an activity token exists. MUST NOT require separate imminent vs timing activity ids.

空列表 MUST end 并清理 activity token；非空有 activity token MUST update，否则 MUST start（push_to_start）；禁止有 token 时先 end 再 start。

#### Scenario: Upsert when token exists

- **WHEN** Refresh builds a non-empty list and the wx has an activity token
- **THEN** push MUST send Live Activity update (not end+start)

#### Scenario: Start when no activity token

- **WHEN** Refresh builds a non-empty list and the wx has only push_to_start (or no activity token)
- **THEN** push MUST send Live Activity start via push_to_start

#### Scenario: End when nothing to show

- **WHEN** Refresh builds an empty items list
- **THEN** push MUST send Live Activity end for that wx

### Requirement: History CUD MUST delete imminent pending then Refresh Live Activity

On successful history create, update, or delete (including close-by-endTime side paths) for a device, the system MUST delete (or empty) that device’s Redis imminent pending authority key, then MUST invoke Refresh. MUST NOT fan out a dedicated Imminent-only `end` clear as the primary LA side effect, and MUST NOT fan out per-history Timing start/update/end multi-slot Live Activities. Delayed MQ consumers MUST continue to treat missing/empty pending as no-match for ordinary predict push.

history CUD 成功后 MUST 先 Del pending，再 Refresh；MUST NOT 以 Imminent end 清屏或多槽 Timing 扇出为主路径。

#### Scenario: CUD clears pending then refreshes

- **WHEN** history write succeeds while Redis pending was non-empty
- **THEN** pending MUST be cleared and Refresh MUST run afterward
- **AND** subsequent delayed fires matching only the old pending MUST NOT send ordinary predict push

#### Scenario: CUD with open timing still projects timing

- **WHEN** history create opens a timing row and pending was cleared in the same CUD side effect
- **THEN** Refresh MUST project the open timing into items (predict segment empty)

### Requirement: Pending sync and MQ fire MUST Refresh without replacing ordinary push

After a successful App full-replace write of imminent pending Redis, the system MUST invoke Refresh (in addition to existing delayed Publish). When MQ imminent fire passes existing gates and proceeds to notify bound users, the system MUST keep sending ordinary `predict_imminent` visible push as today, and MUST also invoke Refresh for Live Activity. Fire MUST NOT end+start a separate Imminent Live Activity slot.

pending PUT 成功 MUST Refresh；MQ 到点在保留普通推送的同时 MUST Refresh；禁止开火路径 Imminent 先 end 再 start。

#### Scenario: PUT then Refresh

- **WHEN** client successfully PUTs a non-empty imminent pending list
- **THEN** system MUST write Redis, schedule delayed fires as today, and MUST Refresh Live Activity

#### Scenario: Fire keeps alert and refreshes LA

- **WHEN** delayed fire matches pending and passes gates
- **THEN** system MUST attempt ordinary predict_imminent visible push AND MUST Refresh Live Activity
- **AND** MUST NOT perform Imminent end-then-start slot replacement

### Requirement: Live Activity token registration SHALL not use attributesKind or activityId slots

`POST /app/api/push/live-activity/register` MUST accept `tokenKind` of `push_to_start` or `activity` without requiring `attributesKind` or `activityId` for slotting. Persistence MUST uniquely key by at least `(wx_id, device_key, token_kind)` (or equivalent without attributes_kind/activity_id). APNs Live Activity payloads MUST use a single fixed attributes-type string for the unified ActivityAttributes. Lookup for update/end MUST use the user’s activity tokens, not a multi-slot activityId map.

register MUST NOT 再要求 attributesKind/activityId 分槽；表唯一键收敛；APNs 使用单一 attributes-type。

#### Scenario: Register activity token without attributesKind

- **WHEN** client registers tokenKind=activity with a token and no attributesKind/activityId slot fields
- **THEN** server MUST persist the activity token for that wx/device and later Refresh update MUST be able to use it

#### Scenario: Reject obsolete dual-kind requirement

- **WHEN** implementing the unified register contract
- **THEN** server MUST NOT require `attributesKind` in `imminent|timing` as a precondition to store tokens
