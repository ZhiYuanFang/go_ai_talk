## ADDED Requirements

### Requirement: history 文本 chat MUST HTTP 委派 voice-service

history-service 处理 `POST /device/history/api/chat` 时 MUST 经 HTTP 调用 voice-service internal 文本 chat 契约，MUST NOT 在 history 进程内 import 或调用 `voice.TextChat` / `voice.Voice()` 执行业务。

history-service MUST NOT 配置 `VOICE_DB_LINK` 或访问 voice 库表（含 `ai_quota_default`、`qa`、`suggest`、`llm_lane_config`）。

对外路径、请求体字段（`deviceNo`、`transcript`）与成功响应结构 MUST 保持不变。

#### Scenario: App chat 经委派完成

- **WHEN** App 调用 `POST /device/history/api/chat` 且 voice-service 可达
- **THEN** history-service MUST 向 voice-service 发起 internal text chat HTTP 请求并返回等价 `reply`

#### Scenario: 额度错误透传

- **WHEN** voice-service internal chat 返回 40301 或 40302
- **THEN** history-service 对外响应 MUST 携带相同 business code 与 message 语义

#### Scenario: 无 voice 库配置

- **WHEN** history-service 进程未配置 voice 数据库连接
- **THEN** chat 路径 MUST 仍可工作（依赖 voice-service HTTP，不依赖本地 voice 库）
