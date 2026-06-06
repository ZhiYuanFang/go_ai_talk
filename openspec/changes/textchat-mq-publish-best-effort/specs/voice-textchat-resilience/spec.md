## ADDED Requirements

### Requirement: TextChat SHALL 在 voice.task 发布失败时仍完成对话

`VoiceService.TextChat` 在调用 `publishTaskRequested`（routing key `voice.task.requested`）时，若 RabbitMQ 发布失败或 publisher 初始化失败，SHALL 记录 Warning 日志并继续执行 `chat`；SHALL NOT 因事件发布失败而向客户端返回错误。

#### Scenario: MQ 不可达时文本对话成功

- **WHEN** 客户端调用文本对话 API 且 RabbitMQ 管理 API 发布失败
- **THEN** 响应 SHALL 仍包含 LLM 对话结果（在 chat 阶段本身成功的前提下）
- **AND** 系统 SHALL 记录发布降级 Warning

#### Scenario: MQ 可达时文本对话行为不变

- **WHEN** RabbitMQ 发布成功
- **THEN** 系统 SHALL 照常发布 `voice.task.requested` 后继续 chat

## MODIFIED Requirements

### Requirement: RabbitMQ 必须作为唯一事件通道

（保留运行时其它必需事件发布失败阻断语义；TextChat 的 voice.task.requested 发布除外，见 `voice-textchat-resilience`。）

#### Scenario: 必需事件发布失败

- **WHEN** 某请求路径要求发布**必需**事件且 RabbitMQ 发布失败或超时（不含 TextChat 的 best-effort voice.task.requested）
- **THEN** 该请求 SHALL 被阻断，并返回事件发布失败错误响应
