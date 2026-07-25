## MODIFIED Requirements

### Requirement: TipStreamRequest 字段类型对齐
Go 侧 `TipStreamRequest` 结构体的 `EventID` 字段 SHALL 为 `int64` 类型，`CurrentTime` 字段 SHALL 为 `int64` 类型（Unix 秒）。

#### Scenario: event_id 类型为 int64
- **WHEN** Go 侧构造 `TipStreamRequest` 并序列化为 JSON
- **THEN** `event_id` 字段 SHALL 为数字类型（int64），而不是字符串
- **AND** 值 SHALL 与事件主键 ID 一致

#### Scenario: current_time 类型为 int64（Unix 秒）
- **WHEN** Go 侧构造 `TipStreamRequest` 并序列化为 JSON
- **THEN** `current_time` 字段 SHALL 为数字类型（int64），表示 Unix 时间戳（秒）
- **AND** 值 SHALL 为当前触发时间，用于 Python 侧作为时间上下文填入 LLM 提示词
