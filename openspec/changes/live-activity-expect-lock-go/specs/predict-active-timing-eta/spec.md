## ADDED Requirements

### Requirement: Active-eta sync SHALL update Timing Live Activity expectEndAt when historyId is known

After a successful full-replace write of active-eta Redis state for a device, for each cleaned item that has a parseable positive `historyId` and `expectEndAt` &gt; 0, the server MUST fan out a Timing Live Activity **update** (via the existing push Live Activity internal path) to bound users of that device, with `activityId` equal to `timing:{deviceNo}:{historyId}` and ContentState including `expectEndAt` (unix seconds). The update SHOULD also include available `phase`/`eventName`/`startTime` (from the ETA item’s title/startAt when present) so a full ContentState replace does not drop the lock-screen title or elapsed base. This path MUST NOT publish delayed MQ messages and MUST NOT send ordinary visible predict alert notifications. When the sync clears to an empty list, the server MUST NOT be required to end Timing activities solely due to ETA clear (Timing end remains history close/delete). Items with unusable `historyId` or non-positive `expectEndAt` MUST be skipped for LA update without failing the Redis sync success.

宝宝 active-eta 全量写入 Redis 成功后，对每条清洗后条目：若 `historyId` 可解析为正整数且 `expectEndAt`&gt;0，则 MUST 经既有 push Live Activity 内部路径，向该宝宝绑定用户扇出 Timing **update**（`activityId=timing:{deviceNo}:{historyId}`），ContentState 含 `expectEndAt`（unix 秒）。update SHOULD 尽量带上可用的 `phase`/`eventName`/`startTime`（可自主题 title / startAt），以免整包替换丢掉锁屏标题或已计时基准。本路径 MUST NOT 发延时 MQ，MUST NOT 发普通可见预测 alert。同步清空为空列表时，MUST NOT 仅因 ETA 清空而 end Timing（Timing end 仍由 history 闭合/删除驱动）。`historyId` 不可用或 `expectEndAt`≤0 的条目 MUST 跳过 LA update，且 MUST NOT 因此使 Redis 同步失败。

#### Scenario: ETA sync updates Timing ContentState

- **WHEN** `PUT .../active-eta/pending` 成功写入含 `historyId=42`、`expectEndAt=1730000000` 的条目
- **THEN** 服务端 MUST 对 activityId `timing:{deviceNo}:42` 扇出 Timing update
- **AND** ContentState MUST 含 `expectEndAt=1730000000`
- **AND** MUST NOT 因此 PublishDelayed 或发普通 predict alert

#### Scenario: Clear ETA does not end Timing

- **WHEN** active-eta 同步以空列表清空 Redis
- **THEN** MUST NOT 要求仅因此 end 所有 Timing Live Activity

#### Scenario: Bad historyId skips LA only

- **WHEN** 某条 ETA 的 `historyId` 无法解析为正整数
- **THEN** 该条 MUST 跳过 LA update
- **AND** 其余合法条目与 Redis 写入成功语义 MUST 不受影响

## MODIFIED Requirements

### Requirement: Active-eta storage MUST NOT schedule imminent pushes

The active-eta Redis store remains authoritative for in-progress timing ETAs and MUST NOT publish delayed MQ wakeups or drive Imminent Live Activity / ordinary predict-imminent alerts. After a successful sync write, the server MAY fan out Timing Live Activity **update** payloads that only refresh ContentState fields such as `expectEndAt`, as specified by「Active-eta sync SHALL update Timing Live Activity expectEndAt when historyId is known」. This exception MUST NOT be used to start new Timing activities or to send user-visible alert banners unrelated to Live Activity.

active-eta Redis 仍为进行中 ETA 权威，MUST NOT 发延时 MQ，MUST NOT 驱动 Imminent Live Activity 或普通 predict-imminent alert。同步写入成功后，MAY 按「Active-eta sync SHALL update Timing Live Activity expectEndAt when historyId is known」扇出仅刷新 ContentState（如 `expectEndAt`）的 Timing **update**。该例外 MUST NOT 用于新 start Timing，也 MUST NOT 借此发与 Live Activity 无关的可见 alert 横幅。

#### Scenario: No imminent scheduling from active-eta

- **WHEN** 客户端成功同步 active-eta 列表
- **THEN** 服务端 MUST NOT 仅为该同步 PublishDelayed 预测临近 MQ
- **AND** MUST NOT 扇出 Imminent Live Activity start
