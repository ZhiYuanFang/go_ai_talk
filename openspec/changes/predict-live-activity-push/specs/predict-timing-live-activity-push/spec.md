## ADDED Requirements

### Requirement: Server SHALL push Timing Live Activity start and end from history lifecycle

When the twin Go service successfully persists a new open timing history record (unset end time per existing open-history rules), it MUST fan out a Timing Live Activity start to bound users of that device/baby using a stable activity id that includes that history identity (e.g. `historyId`). When that record is closed (end time set), Go MUST fan out a Timing Live Activity end for the same activity id. Multiple open timings MUST be allowed to map to multiple concurrent Timing Live Activities. This path MUST NOT require a fixed multi-second delay gate to re-check end time before start. The Imminent pending MQ fire MUST NOT be the sole carrier for Timing start/end. The active-eta Redis sync MAY remain for ETA data and MUST NOT be required as the only start trigger.

Fanout MUST be triggered exactly once from the authoritative history persistence success path (so that App HTTP, voice via `clients/history`, and other writers that converge on that persist all cause a single side effect). MUST NOT register duplicate fanout hooks on both a remote client adapter and the history local persist for the same write.

Go 成功落库未闭合计时 history 时 MUST 向该宝宝绑定用户扇出 Timing Live Activity start，activity id MUST 稳定含该 history 身份。闭合（写入 end）时 MUST 对同一 id 扇出 end。多条未闭合 MUST 允许对应多条 Timing。MUST NOT 要求固定多秒延迟再确认 endTime 才 start。MUST NOT 仅用 Imminent 延时开火作为 Timing 唯一载体。active-eta MAY 保留，MUST NOT 作为唯一 start 触发。

扇出 MUST 由 history **权威持久化成功**路径触发一次，使经该落库汇入的所有写入口（含 App HTTP、voice 经 `clients/history` 等）间接触发同一副作用。MUST NOT 在 remote 调用方适配器与 history local 落库对同一次写入各挂一次扇出导致双推。

#### Scenario: Start on open timing

- **WHEN** Go 成功持久化一条未闭合计时 history
- **THEN** Go MUST 向该 deviceNo 绑定用户尝试 Timing Live Activity start

#### Scenario: End on close timing

- **WHEN** Go 将该 history 更新为已设置 endTime
- **THEN** Go MUST 向绑定用户发送对应 activity id 的 Timing Live Activity end（或等价 dismiss）

#### Scenario: Quick start-stop without delay gate

- **WHEN** 计时短时间开始后又结束且两端扇出均触发
- **THEN** 系统 MUST 通过 end 推送关闭对应锁屏组件，MUST NOT 依赖额外固定秒级「确认仍在计时」延迟作为 start 前置条件

#### Scenario: Multiple timings

- **WHEN** 同一宝宝下两条不同 historyId 的计时均未闭合且均已 start 扇出
- **THEN** Go MUST 使用不同 activity id，允许客户端同时展示多条 Timing Live Activity

#### Scenario: Single fanout per persist

- **WHEN** voice 经 HTTP 调用 history-service 成功落库一条未闭合计时
- **THEN** Timing Live Activity start MUST 仅因该次落库成功副作用扇出一次
- **AND** MUST NOT 因 voice 侧额外钩子再扇出第二次
