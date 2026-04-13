## ADDED Requirements
### Requirement: 文字对话接口
系统 SHALL 提供需鉴权的 POST `/text/chat` 接口，接受 JSON 请求体（包含 `deviceNo` 与 `text`），并返回 JSON 回复（包含 `reply`）。

#### Scenario: 合法文本请求返回回复
- **WHEN** 客户端携带有效 `Token`，提交 `deviceNo` 与非空 `text`
- **THEN** 系统返回 200 且响应体包含非空 `reply`

#### Scenario: 文本为空被拒绝
- **WHEN** 客户端提交空 `text`
- **THEN** 系统返回 400，说明参数校验失败

### Requirement: 跨语音/文字共享连续对话
系统 SHALL 基于 `deviceNo` 在进程内维护临时会话历史（最近 N 轮 + TTL），并在调用 DeepSeek 时拼接历史消息；`/text/chat` 与 `/voice/chat` MUST 共享同一份设备会话历史。

#### Scenario: 文本后接语音可延续上下文
- **WHEN** 同一 `deviceNo` 先调用 `/text/chat` 并成功获得回复
- **AND** 在 TTL 内调用 `/voice/chat`
- **THEN** `/voice/chat` 发起的 DeepSeek 请求包含先前 `/text/chat` 写入的历史消息

#### Scenario: 语音后接文本可延续上下文
- **WHEN** 同一 `deviceNo` 先调用 `/voice/chat` 并成功获得回复
- **AND** 在 TTL 内调用 `/text/chat`
- **THEN** `/text/chat` 发起的 DeepSeek 请求包含先前 `/voice/chat` 写入的历史消息

#### Scenario: DeepSeek 失败不污染历史
- **WHEN** `/text/chat` 调用 DeepSeek 超时或返回错误
- **THEN** 系统返回结构化错误
- **AND** 不写入本轮不完整消息到该设备会话缓存
