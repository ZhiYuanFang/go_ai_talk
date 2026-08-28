## ADDED Requirements

### Requirement: 禁止同 eventId 重复 start 进行中记录

`history-service` 在 `AddDeviceHistory`（及经其的 `POST /device/history/api/event/add`、`POST /device/history/api/event/batch` 之 `op=create`）插入新行时，若请求满足 `eventId>0` 且 `endTime=0`（进行中），MUST 先查询该 `deviceNo` 下是否已存在同 `eventId` 且 `end_time=0` 的历史行；若存在 MUST 拒绝插入并返回可读错误（语义等价：`该事件已在进行中，请先结束后再开始`）。`UpdateDeviceHistory`（及 `event/update`、batch `op=update`）将某行 `endTime` 改为 `0` 时 MUST 应用相同规则（排除当前行 `id`）。本守卫 MUST NOT 限制不同 `eventId` 的并行进行中； MUST NOT 拦截 `endTime!=0` 的瞬时记录（含 `startTime=endTime` 的 one 类记录）。

#### Scenario: 同 event 已有进行中时拒绝再次 start

- **WHEN** 设备 `deviceNo=D` 已存在 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方经 add 或 batch create 请求插入 `eventId=E`、`endTime=0` 的新行
- **THEN** history-service MUST 拒绝该插入
- **AND** MUST NOT 新增第二条同 E 的未闭合行

#### Scenario: 不同 event 允许并行进行中

- **WHEN** 设备已有 `eventId=E1` 且 `end_time=0` 的睡眠记录
- **AND** 调用方插入 `eventId=E2`（E2≠E1）且 `endTime=0` 的计时记录
- **THEN** history-service MUST 允许插入

#### Scenario: 瞬时 one 记录不受限

- **WHEN** 设备已有 `eventId=E` 且 `end_time=0` 的进行中行
- **AND** 调用方插入同 E 但 `endTime>0` 且 `endTime>=startTime` 的瞬时记录
- **THEN** history-service MUST 允许插入

#### Scenario: batch 单条失败不整单回滚

- **WHEN** Python 经 `event/batch` 提交多条 items，其中一条因重复 start 失败
- **THEN** 该 item MUST 在 results 中 `ok=false` 且 `reason` 含可读拒绝语义
- **AND** 其它 items MUST 仍按既有 batch 语义独立执行

## MODIFIED Requirements

### Requirement: 语音 end 优先闭合未结束记录再降级新建

喂养事件 CRUD（含 start/end/one/multi）MUST 由 `python-ai-talk` 在意图落地阶段经 `POST /device/history/api/event/batch` 写入 history；Go `voice-service` MUST NOT 在 `applyUnifiedIntentResult` 或等价主路径内调用 `DeviceHistory().AddHistory` / `UpdateHistory` 执行喂养写库。当 Python 处理 `action=end` 且目标 event 为 E 时，batch MUST 优先使用 `op=end` 或等价语义经 `EndLatestHistoryIfMatch` 闭合；仅当无未闭合同 E 记录时 MAY 降级 `create` 瞬时结束行（`start_time=end_time`）。MUST NOT 在仍存在未闭合 E 记录时新建一条 E 的历史行作为「结束」手段。

#### Scenario: 孩子醒了结束已开始的睡眠

- **WHEN** Python 返回结束睡眠意图且该设备存在未闭合睡眠历史（同 eventId）
- **THEN** Python batch MUST 闭合该未闭合睡眠行（经 end-latest 或 batch op=end）
- **AND** MUST NOT 再新增一条睡眠历史行作为结束

#### Scenario: 从未开始睡眠时 end 可降级新建

- **WHEN** Python 返回结束睡眠意图，且该设备不存在未闭合的同 eventId 历史
- **THEN** batch MAY 写入瞬时结束记录
- **AND** Go voice MAY 透传 Python 成功话术

#### Scenario: 重复 start 被 history 拒绝

- **WHEN** Python batch 提交同 eventId 的 start（`endTime=0`）且该 event 已在进行中
- **THEN** history-service MUST 拒绝该 create
- **AND** batch 对应 item MUST 返回失败 reason
