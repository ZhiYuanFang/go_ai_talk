## MODIFIED Requirements

### Requirement: 语音结束事件 MUST 与 API 结束等价推送 WS

当用户通过语音链路触发「结束」计时类事件（`ActionTargetTypeEnd` / Python `target_type=feeding` 且 `action=end`）且系统向用户播报结束成功时，voice-service MUST 经 `DeviceHistory` 契约完成 history 写库，且该写库 MUST 在 history-service 侧触发与 App 调用 `POST /device/history/api/event/end-latest`（`updated=true`）或等价 `event/update` 相同的 Redis 通知：向 `app:history:notify` PUBLISH，`action` 为 `update` 或 `create`（与实落库操作一致），载荷含 `deviceNo`、记录主键及展示字段。

`EndLatestHistoryIfMatch` / `end-latest` 的匹配权威 MUST 为：该 `deviceNo` 下指定 `eventId` 且 `end_time=0` 的最近一条记录（按 `id` 降序）；MUST NOT 要求该行同时是设备全局最新一条 history。

voice-service MUST NOT 在 `EndLatestHistoryIfMatch` 返回 `updated=false` 且未执行成功降级写库的情况下向用户播报结束成功。

#### Scenario: 结束最近一条同 eventId 未闭合计时记录

- **WHEN** 用户语音或 App 结束事件 E，且 history-service 存在 `eventId=E` 且 `end_time=0` 的记录（取 `id` 最大者），即使全局最新一条 history 的 `eventId` 不等于 E
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=true`
- **AND** 系统 SHALL PUBLISH `action=update` 的 WS 通知
- **AND** 语音路径在写库成功后 MUST 表示结束成功

#### Scenario: EndLatest 未匹配时降级写库仍推送

- **WHEN** 用户语音结束事件 E，且该设备不存在 `eventId=E` 且 `end_time=0` 的记录，因而 `EndLatestHistoryIfMatch` 对 E 返回 `updated=false`
- **THEN** voice-service MUST 执行降级写库（至少 `AddHistory` 写入 E 的瞬时结束记录）
- **AND** 降级写库成功后系统 SHALL PUBLISH `action=create` 或 `update` 的 WS 通知
- **AND** 语音回复 MUST 仅在至少一次写库成功后可表示结束成功

#### Scenario: 与 App end-latest 行为一致

- **WHEN** 同一 `deviceNo` 下 App 调用 `event/end-latest` 成功（`updated=true`）与语音结束同一未闭合 event 在 history 库产生相同终态
- **THEN** 两种路径触发的 WS 载荷在 `action`、记录 `id`、`eventId`、`endTime` 语义上 MUST 一致（允许 `remark` 来源不同）
