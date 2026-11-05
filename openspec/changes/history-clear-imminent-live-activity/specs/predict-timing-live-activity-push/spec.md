## ADDED Requirements

### Requirement: Timing Live Activity end MUST delete activity token rows

When Go fans out a Timing Live Activity **end** (history closed by setting endTime, or the history row is deleted), after the end fanout attempt the push domain MUST delete persisted `token_kind=activity` rows for that Timing `activityId` (e.g. `timing:{deviceNo}:{historyId}`), so ended timing activities do not accumulate invalid update tokens. Push-to-Start tokens for Timing MUST NOT be deleted by this cleanup. This MUST use the same push-layer `action=end` cleanup as Imminent ends (MUST NOT require a separate history-only delete path). Redis Timing slot key deletion on end MAY remain as already implemented on the history Timing end path.

Timing Live Activity **end**（闭合写入 endTime，或删除该行）扇出尝试结束后，push 域 MUST 删除该 Timing `activityId` 下 `token_kind=activity` 持久化行，避免计时结束后无效 update token 累计。MUST NOT 删除 Timing 的 Push-to-Start。MUST 与 Imminent end 共用 push 层 `action=end` 清理，MUST NOT 仅在 history 再写一套删库。history Timing end 路径既有 Redis 槽 `Del` MAY 保持。

#### Scenario: Close timing removes activity tokens

- **WHEN** Go 因 history 闭合对 `timing:{deviceNo}:{historyId}` 扇出 Timing `event=end`
- **THEN** push 域 MUST 删除该 `activityId` 下全部 `token_kind=activity` 行
- **AND** MUST NOT 删除该用户 Timing 的 Push-to-Start 行

#### Scenario: Delete open timing removes activity tokens

- **WHEN** Go 删除未闭合 history 并扇出 Timing end
- **THEN** 该 Timing `activityId` 的 activity token 行 MUST 同样被删除

#### Scenario: Timing update does not delete tokens

- **WHEN** Go 对未闭合 Timing 仅扇出 `action=update`
- **THEN** MUST NOT 因 update 删除该 `activityId` 的 activity token 行
