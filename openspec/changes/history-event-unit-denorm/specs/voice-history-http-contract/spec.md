## ADDED Requirements

### Requirement: Voice 写入 history SHOULD 传递已知 eventUnit

当 voice-service 经 device HTTP 契约已解析出 `entity.Event` 且 `Unit` 非空时，调用 history 新增/更新接口的请求体 SHOULD 携带 `eventUnit` 字段。history-service MUST 仍支持未携带时经 device 契约补全。

#### Scenario: 语音记录奶量事件

- **WHEN** voice 解析到事件 `unit=ml` 并调用 history 新增一条计数记录
- **THEN** HTTP 请求体 SHOULD 包含 `eventUnit=ml`，且 history 持久化后 `event_unit` SHALL 为 `ml`

#### Scenario: Voice 未传 eventUnit 时由 history 补全

- **WHEN** voice 调用 history 新增记录未传 `eventUnit` 但 `eventId` 对应主档 `unit=ml`
- **THEN** history-service MUST 经 device 契约补全并成功写入 `event_unit=ml`
