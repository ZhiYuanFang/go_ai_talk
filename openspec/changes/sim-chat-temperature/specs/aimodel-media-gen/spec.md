## ADDED Requirements

### Requirement: aimodel ChatRequest SHALL support optional temperature for chat completions

`aimodel.ChatRequest` MUST 提供可选字段 `Temperature *float64`。当指针非 nil 时，`Invoke` / `InvokeStream`（及 `InvokeWithHeldProfile` 等等价路径）构建的上游 JSON MUST 包含 `temperature` 字段且值为所设浮点数。当指针为 nil 时，MUST NOT 向请求体写入 `temperature`（保持上游 provider 默认行为）。

#### Scenario: Temperature omitted by default

- **WHEN** 调用方构造 `ChatRequest` 且未设置 `Temperature`
- **THEN** 序列化后的请求体 MUST NOT 含 `temperature` 键

#### Scenario: Temperature explicitly set

- **WHEN** 调用方设置 `Temperature` 指向 `0.85`
- **THEN** 请求体 MUST 含 `"temperature": 0.85`

#### Scenario: Stream and non-stream parity

- **WHEN** 同一 `ChatRequest`（含 Temperature）分别用于流式与非流式
- **THEN** 二者 MUST 写入相同 `temperature` 值
