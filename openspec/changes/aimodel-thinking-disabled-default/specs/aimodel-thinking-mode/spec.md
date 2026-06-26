## ADDED Requirements

### Requirement: aimodel SHALL 对智谱请求显式声明 thinking 模式且默认 disabled

`internal/services/aimodel` 在 `provider=zhipu` 的 chat/completions 请求体中 MUST **始终**包含 `thinking.type` 字段。当 `ChatRequest.ThinkingEnabled` 为 `false` 或未设置（零值）时 MUST 发送 `thinking.type=disabled`。当 `ThinkingEnabled` 为 `true` 时 MUST 发送 `thinking.type=enabled`。MUST NOT 因未启用 thinking 而省略 `thinking` 字段（避免 GLM-4.7 等模型上游默认开启 thinking）。

#### Scenario: sim 短文案默认关闭 thinking

- **WHEN** sim-user 经 `aimodel.Invoke(LaneSimText, ChatRequest{MaxTokens: 64})` 调用且未设置 `ThinkingEnabled`
- **THEN** 发往智谱的请求体 MUST 含 `"thinking": {"type": "disabled"}`

#### Scenario: clinic opt-in 开启 thinking

- **WHEN** voice clinic 经 `aimodel.InvokeStream` 调用且 `ThinkingEnabled=true`
- **THEN** 发往智谱的请求体 MUST 含 `"thinking": {"type": "enabled"}`

### Requirement: ChatRequest.MaxTokens SHALL 文档化与 thinking 的预算关系

`aimodel.ChatRequest` 的 `MaxTokens` 字段注释 MUST 说明：该值为上游 completion token 上限；当 thinking enabled 时 reasoning 与 content **共用**该预算，调用方不得将其仅理解为最终正文长度上限。

#### Scenario: 调用方阅读契约

- **WHEN** 开发者查阅 `internal/services/aimodel/request.go` 中 `ChatRequest`
- **THEN** MUST 可见 thinking 默认 disabled 语义及 MaxTokens 与 reasoning 共预算的说明

### Requirement: DeepSeek adapter SHALL 仅在 opt-in 时启用 thinking

`provider=deepseek` 时，仅当 `ChatRequest.ThinkingEnabled=true` MUST 写入 `extra_body.thinking`（或等价配置）与可选 `reasoning_effort`。`ThinkingEnabled=false` MUST NOT 写入 thinking 启用字段。

#### Scenario: 喂养 voice 未 opt-in

- **WHEN** voice 经 aimodel 调用 DeepSeek provider 且 `ThinkingEnabled=false`
- **THEN** 请求体 MUST NOT 含 thinking enabled 配置
