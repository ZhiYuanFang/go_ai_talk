## ADDED Requirements

### Requirement: aimodel 统一入口 SHALL 默认关闭上游 thinking

凡经 `aimodel.Invoke` / `InvokeStream` / `InvokeWithHeldProfile` / `InvokeStreamWithHeldProfile` 发起的 chat/completions 请求，若调用方未将 `ChatRequest.ThinkingEnabled` 设为 `true`，aimodel MUST 按 provider 规则关闭 thinking（智谱 MUST 显式 `disabled`）。该默认语义 MUST 适用于所有 lane（含 `voiceUnderstanding`、`polish`、`simText`、`simVision` 等），MUST NOT 依赖上游模型默认 thinking 行为。

#### Scenario: 润笔未 opt-in

- **WHEN** `PolishPostText` 调用 `aimodel.Invoke(LanePolish, ...)` 且未设置 `ThinkingEnabled`
- **THEN** 智谱请求 MUST 显式 `thinking.type=disabled`

#### Scenario: voiceUnderstanding 闲聊

- **WHEN** 喂养语音 LLM 经 `LaneVoiceUnderstanding` 调用且未设置 `ThinkingEnabled`
- **THEN** 对智谱 provider 的请求 MUST 显式 `thinking.type=disabled`
