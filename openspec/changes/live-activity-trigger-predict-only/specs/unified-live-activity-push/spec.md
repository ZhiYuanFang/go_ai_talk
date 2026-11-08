## MODIFIED Requirements

### Requirement: History CUD MUST delete imminent pending then Refresh Live Activity

**Reason for rewrite**: LA Refresh 从 CUD 挪到预测 PUT/到点；CUD 仅作废 pending。

On successful history create, update, or delete (including close-by-endTime side paths) for a device, the system MUST delete (or empty) that device’s Redis imminent pending authority key. The system MUST NOT invoke Live Activity Refresh or fan out Live Activity as a side effect of that history mutation. Delayed MQ consumers MUST continue to treat missing/empty pending as no-match for ordinary predict push. MUST NOT fan out dedicated Imminent-only end clear or per-history Timing multi-slot Live Activities as the CUD primary path.

history CUD 成功后 MUST Del pending；MUST NOT 因 CUD 触发 LA Refresh/扇出。

#### Scenario: CUD clears pending without LA Refresh

- **WHEN** history write succeeds while Redis pending was non-empty
- **THEN** pending MUST be cleared
- **AND** the system MUST NOT Refresh or fan out Live Activity solely due to that CUD

#### Scenario: Create does not project timing-only LA

- **WHEN** history create opens a timing row and pending is cleared in the same CUD side effect
- **THEN** the system MUST NOT Refresh Live Activity from that CUD
- **AND** Live Activity projection MUST wait for a later predict pending PUT and/or MQ fire Refresh

### Requirement: Pending sync and MQ fire MUST Refresh without replacing ordinary push

After a successful App full-replace write of imminent pending Redis, the system MUST schedule/invoke the unified Live Activity Refresh (subject to deviceNo trailing coalesce). When MQ imminent fire passes existing gates and proceeds to notify bound users, the system MUST keep sending ordinary `predict_imminent` visible push as today, and MUST also schedule/invoke Refresh for Live Activity (same coalesce). Fire MUST NOT end+start a separate Imminent Live Activity slot.

pending PUT 成功 MUST 触发（可合并的）LA Refresh；MQ 到点在保留普通推送的同时 MUST 触发 Refresh。

#### Scenario: PUT then Refresh

- **WHEN** client successfully PUTs an imminent pending list (empty or non-empty)
- **THEN** system MUST write Redis, schedule delayed fires as today when non-empty, and MUST schedule unified Live Activity Refresh

#### Scenario: Fire keeps alert and refreshes LA

- **WHEN** delayed fire matches pending and passes gates
- **THEN** system MUST attempt ordinary predict_imminent visible push AND MUST schedule Live Activity Refresh
- **AND** MUST NOT perform Imminent end-then-start slot replacement

## ADDED Requirements

### Requirement: Live Activity Refresh SHALL coalesce by deviceNo with trailing 300ms

All Live Activity Refresh requests that are allowed by the trigger matrix MUST be coalesced per `deviceNo` with a trailing debounce of **300 milliseconds**: each new schedule for the same device MUST reset the timer; when the timer fires, the system MUST run exactly one `RefreshBabyLiveActivity` (or equivalent) that rebuilds ContentState and fans out to bound wxIds. Within that single run, each bound wx MUST receive at most one Live Activity APNs attempt (excluding invalid-token restart failover on the same send). This coalesce is an in-process short-window merge, MUST NOT be implemented as a periodic table/outbox scanner ticker.

同 deviceNo trailing 300ms 合并 Refresh；每轮每 wx 至多一次 APNs（不含同拍 invalid restart）。

#### Scenario: Burst schedules one Refresh

- **WHEN** pending PUT and MQ fire both request Refresh for the same device within 300ms
- **THEN** the system MUST execute a single Refresh after the trailing window
- **AND** MUST NOT execute one Refresh per request in that burst

### Requirement: Active-eta sync MUST NOT trigger Live Activity Refresh

After a successful write or clear of the device’s predict active-eta Redis list, the system MUST NOT invoke or schedule Live Activity Refresh. Expect-end fields on in-progress items remain merged when Refresh runs for other allowed triggers (pending PUT / MQ fire).

active-eta PUT/清空 MUST NOT 触发 LA。

#### Scenario: Active-eta write alone

- **WHEN** client successfully syncs active-eta for a device
- **THEN** Redis active-eta MUST be updated as today
- **AND** the system MUST NOT schedule Live Activity Refresh from that sync alone
