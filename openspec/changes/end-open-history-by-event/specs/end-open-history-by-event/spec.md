## ADDED Requirements

### Requirement: EndLatest 按 eventId 闭合最近未结束记录

`history-service` 的 `EndLatestHistoryIfMatch`（及内部 HTTP `POST /device/history/api/event/end-latest`）MUST 在指定 `deviceNo` 与 `eventId` 下，查找 `end_time=0`（未闭合）的历史行，按 `id` 降序取最近一条并更新其结束时间；MUST NOT 仅因「全局最新一条 history 的 eventId 不等于目标」而返回未匹配。当 remark 非空时 MUST 同时覆盖该行备注；remark 为空串时 MUST NOT 修改原备注。

#### Scenario: 中间夹其它事件仍能结束睡眠

- **WHEN** 设备已有未闭合睡眠记录（`eventId=E` 且 `end_time=0`），其后又写入其它事件记录，使全局最新一条不再是睡眠
- **AND** 调用方对同一 `deviceNo` 请求结束 `eventId=E`（语音 `feeding+end` 或 App `end-latest`）
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=true`
- **AND** SHALL 更新该未闭合睡眠行的 `end_time`（而非新建一条睡眠历史）

#### Scenario: 无未闭合同 event 时返回未匹配

- **WHEN** 该 `deviceNo` 下不存在 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方请求结束 `eventId=E`
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=false`
- **AND** SHALL NOT 修改任何已结束（`end_time!=0`）的同 `eventId` 行

#### Scenario: 多条未闭合同 event 只闭合最近一条

- **WHEN** 同一 `deviceNo` 下存在多条 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方请求结束 `eventId=E`
- **THEN** 系统 SHALL 仅更新其中 `id` 最大的一条的 `end_time`
- **AND** 其余未闭合行保持 `end_time=0`

### Requirement: 语音 end 优先闭合未结束记录再降级新建

当 voice 处理 Python 意图 `target_type=feeding` 且 `action=end` 并已解析到事件 E 时，MUST 先调用 `EndLatestHistoryIfMatch(deviceNo, E, ...)`；仅当返回 `updated=false` 时 MAY 降级执行瞬时 `AddHistory`（`start_time=end_time`）。MUST NOT 在仍存在未闭合的 E 记录时新建一条 E 的历史行作为「结束」手段。

#### Scenario: 孩子醒了结束已开始的睡眠

- **WHEN** Python 返回 `target_type=feeding`、`action=end`、`event_id` 对应睡眠，且该设备存在未闭合睡眠历史
- **THEN** voice SHALL 经契约闭合该未闭合睡眠
- **AND** SHALL NOT 再新增一条睡眠历史行

#### Scenario: 从未开始睡眠时 end 可降级新建

- **WHEN** Python 返回结束睡眠意图，且该设备不存在未闭合睡眠历史
- **THEN** `EndLatestHistoryIfMatch` SHALL 为 `updated=false`
- **AND** voice MAY 降级 `AddHistory` 写入瞬时结束记录并向用户播报成功
