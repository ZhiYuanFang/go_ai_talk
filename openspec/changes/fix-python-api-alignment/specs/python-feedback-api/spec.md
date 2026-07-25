## MODIFIED Requirements

### Requirement: Python feedback 接口接收 JSON Body
Python 侧 `/v1/clinic/feedback` 和 `/v1/tip/feedback` 接口 SHALL 接收 JSON Body 格式的反馈请求，而不是 Query 参数。

#### Scenario: 反馈接口接收 JSON Body
- **WHEN** 调用方以 `POST /v1/clinic/feedback` 或 `POST /v1/tip/feedback` 发送请求
- **THEN** 请求 Body SHALL 为 JSON 格式，包含 `answer_id`（string）和 `feedback`（int，1 或 -1）字段
- **AND** Python 侧 SHALL 从 JSON Body 中解析参数，而不是从 Query 参数解析
- **AND** 响应格式 SHALL 保持不变（`{"code": 0, "message": "...", "data": {...}}`）

#### Scenario: 无效反馈参数
- **WHEN** 请求体中 `feedback` 字段值不是 1 或 -1
- **THEN** Python 侧 SHALL 返回 400 状态码和错误提示

#### Scenario: 反馈频率超限
- **WHEN** 同一 answer_id 在 60 分钟内反馈次数超过 5 次
- **THEN** Python 侧 SHALL 返回 429 状态码和频率限制提示
