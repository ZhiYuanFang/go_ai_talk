## MODIFIED Requirements

### Requirement: Live Activity ContentState SHALL be an ordered items list

ContentState MUST carry an `items` array (name MAY be `items`). Predict candidates MUST be taken from Redis imminent pending and MUST include only items that satisfy the lead window: `nextAt - leadSeconds ≤ now`, where `leadSeconds` MUST use the same configuration source and default as ordinary predict imminent push (`VOICE_PREDICT_IMMINENT_LEAD_SECONDS`, default 300 seconds). Pending items outside that window MUST NOT appear in `items` (they MAY remain in Redis for scheduling). Among included predict items, sort by `nextAt` ascending (sooner first). Open timing items MUST be `end_time=0` histories for the device, sorted by `startTime` descending.

The total length of `items` MUST be at most **3**. Truncation MUST prefer in-progress: first take up to 3 open timing items (in the sort above); any remaining slots MUST be filled from lead-window predict items (in the sort above). After selecting the set, the emitted `items` order MUST still place the selected predict segment first, then the selected timing segment (MUST NOT time-interleave). Each item MUST include a discriminant `kind` distinguishing predict vs in-progress (e.g. `imminent` / `timing`) and the fields needed for lock-screen display (at least event identity and the relevant time field). Optional structured expect fields MAY be included when available. ContentState and Attributes MUST NOT include `logo` / `color` / `logoFile` / remote image URLs. When no lead-window pending remains after filter/cap, the predict segment MUST be empty; when there are no open histories (or none selected under the cap), the timing segment MUST be empty.

ContentState MUST 含序列表：预测仅 lead 进窗且 `nextAt` 升序；进行中 `startTime` 倒序；总条数 ≤ 3 且进行中优先截断；展示仍预测段在前；禁止视觉 URL/色值字段。

#### Scenario: Sort order within selected set

- **WHEN** lead-window pending has nextAt=100 and nextAt=50, and two open histories with startTime=10 and startTime=20, and the selected set includes all four under the cap
- **THEN** items MUST order predict(50), predict(100), timing(20), timing(10)

#### Scenario: Past nextAt still included while in pending

- **WHEN** Redis pending still contains an item whose nextAt is in the past (hence inside the lead window)
- **THEN** Refresh MUST still include that item until App PUT replaces or clears pending (MUST NOT drop solely because nextAt ≤ now)

#### Scenario: Far pending excluded from Live Activity

- **WHEN** Redis pending contains an item with `nextAt - leadSeconds > now`
- **THEN** Refresh MUST NOT include that item in `items`
- **AND** the item MAY remain in Redis pending for ordinary MQ scheduling

#### Scenario: Cap prefers in-progress over predict

- **WHEN** there are 2 open histories and 4 lead-window predict items
- **THEN** items MUST contain exactly 3 entries: the 2 timing items and the single soonest predict item
- **AND** emitted order MUST be that one predict, then the two timings (`startTime` desc)

#### Scenario: Three open histories fill the cap

- **WHEN** there are 3 or more open histories and any number of lead-window predicts
- **THEN** items MUST contain at most 3 timing items and MUST NOT include predict items

### Requirement: Live Activity token registration SHALL not use attributesKind or activityId slots

`POST /app/api/push/live-activity/register` MUST accept `tokenKind` of `push_to_start` or `activity` without requiring `attributesKind` or `activityId` for slotting. Persistence MUST uniquely key by at least `(wx_id, device_key, token_kind)` (or equivalent). The `push_live_activity_token` schema MUST NOT retain business columns `attributes_kind` or `activity_id` (existing deployments MUST drop them; new installs MUST NOT create them). APNs Live Activity payloads MUST use a single fixed attributes-type string for the unified ActivityAttributes. Lookup for update/end MUST use the user’s activity tokens, not a multi-slot activityId map.

register MUST NOT 再要求 attributesKind/activityId；表唯一键收敛；**MUST** 去掉 attributes_kind/activity_id 列；APNs 使用单一 attributes-type。

#### Scenario: Register activity token without attributesKind

- **WHEN** client registers tokenKind=activity with a token and no attributesKind/activityId slot fields
- **THEN** server MUST persist the activity token for that wx/device and later Refresh update MUST be able to use it

#### Scenario: Reject obsolete dual-kind requirement

- **WHEN** implementing the unified register contract
- **THEN** server MUST NOT require `attributesKind` in `imminent|timing` as a precondition to store tokens

#### Scenario: Schema has no attributes_kind column

- **WHEN** ensure/migrate runs for `push_live_activity_token`
- **THEN** the table MUST NOT expose `attributes_kind` or `activity_id` as required business columns
- **AND** upsert MUST NOT write those fields
