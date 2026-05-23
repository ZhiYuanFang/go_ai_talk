# device-event-type-field Specification

## Purpose
TBD - created by archiving change event-type-replace-need-quantity. Update Purpose after archive.
## Requirements
### Requirement: 事件主档必须持久化有效的 event_type

`device-service` 在创建或更新 `event` 表记录时，SHALL 接受并持久化 `event_type`，其值 MUST 为 `number`、`time` 或 `one` 之一。系统 SHALL NOT 再读写 `need_quantity` 列或 API 字段 `needQuantity`。

#### Scenario: 管理端新增事件带 eventType

- **WHEN** 客户端 `POST /device/admin/api/event/add` 提交合法 `eventType=number`
- **THEN** 数据库新行 `event_type` SHALL 为 `number`
- **AND** 随后 `ListEvents` 或缓存命中 SHALL 返回该 `eventType`

#### Scenario: 非法 eventType 被拒绝

- **WHEN** 客户端提交 `eventType` 为空或不在 `number|time|one`
- **THEN** API SHALL 返回参数错误且 SHALL NOT 插入或更新行

### Requirement: 事件选项 Redis 快照含 event_type

写库成功后，系统 SHALL 通过从数据库全量扫描（含 `event_type` 列）重建 Redis 事件 options，且 SHALL NOT 依赖可能过期的 `ListEvents` 缓存读回后写回。

#### Scenario: 更新事件后缓存含新类型

- **WHEN** 管理员成功更新某事件的 `eventType` 为 `one`
- **THEN** 重建后的 Redis 快照中该事件 SHALL 带有 `eventType` 为 `one`

### Requirement: 匹配已有事件时不改 event_type

当仅合并别名（`extra_names`）或命中已有事件名时，系统 SHALL NOT 更新该事件行的 `event_type`。

#### Scenario: DeepSeek 仅追加别名

- **WHEN** 抽取结果匹配已存在事件名且仅合并 `extraNames`
- **THEN** 该事件 `event_type` 列 SHALL 保持不变

### Requirement: 对话新建事件时由模型提供 event_type

经 voice 调用的 `InsertOrGetEventByNeedle` 或 DeepSeek 落库插入新事件时，系统 SHALL 将模型给出的 `event_type` 写入新行；若模型未给出合法值，SHALL 使用规范化默认值（`time`）并仍可观测。

#### Scenario: 语音路径插入新事件

- **WHEN** 用户话术导致新建事件且 DeepSeek 返回 `event_type` 为 `number`
- **THEN** 新插入的 `event` 行 `event_type` SHALL 为 `number`

