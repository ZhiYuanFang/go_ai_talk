## MODIFIED Requirements

### Requirement: 历史 CUD 后发布 Redis 通知

history-service SHALL 在任何导致 history 表新增、更新或删除成功的业务路径完成后，向约定 Redis channel 发布一条消息；消息体 MUST 包含 `device_no`、操作类型 `action`（create、update、delete 之一）以及供前端更新的历史记录载荷。当记录存在非空 `event_unit` 时，载荷 MUST 包含 `eventUnit` 字段且值与数据库一致。

#### Scenario: 新增历史

- **WHEN** 新增一条 history 成功提交且 `event_unit=ml`
- **THEN** 系统 SHALL PUBLISH 一条 action 为 create 的消息且包含新记录标识、展示所需字段及 `eventUnit=ml`

#### Scenario: 更新或删除历史

- **WHEN** 更新或删除 history 成功提交
- **THEN** 系统 SHALL PUBLISH 对应 update 或 delete 的消息且包含受影响记录的主键或 event 关联信息；update 时 MUST 包含最新 `eventUnit`

## ADDED Requirements

### Requirement: History 投影事件 MUST 携带 event_unit

`history.record.created` / `history.record.updated` 域 outbox 与 Redis 缓存投影消费载荷 MUST 包含 `event_unit`（或等价 JSON 字段），与权威库 `history.event_unit` 一致；投影写入 Redis 读模型时 MUST 保留 `EventUnit`。

#### Scenario: 创建投影含单位

- **WHEN** outbox 发布 `history.record.created` 且对应行 `event_unit=次`
- **THEN** 投影载荷 MUST 含 `event_unit=次`，且 Redis 列表缓存中该条记录的 `eventUnit` SHALL 为 `次`
