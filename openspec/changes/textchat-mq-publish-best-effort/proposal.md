## Why

生产容灾场景下 RabbitMQ 不可用时，`TextChat` 在对话前同步发布 `voice.task.requested` 失败会阻断整个文本对话 API（`POST /device/history/api/chat` 等），与「MQ 挂时核心 API 仍可用」目标冲突。`rabbitmq-optional-startup-check` 已解决启动解耦，需补齐运行期最后一处硬依赖。

## What Changes

- `VoiceService.TextChat`：`publishTaskRequested` 失败时改为 **Warning 日志 + 继续执行 chat**，不再返回 `event_publish` 错误。
- **不引入** outbox 路径；不修改 `VoiceTaskProducer`、worker 消费或其它 MQ 发布点。

## Capabilities

### New Capabilities

- `voice-textchat-resilience`: TextChat 对 voice.task.requested 事件发布采用 best-effort 语义。

### Modified Capabilities

- `platform-hardening-redis-rabbitmq-service-split`（v1.0.3 基线）：TextChat 路径不再属于「必需事件发布失败阻断请求」场景。

## Impact

- `internal/services/voice/voice_chat.go`（`TextChat` 一处）
