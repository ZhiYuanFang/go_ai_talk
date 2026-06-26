## ADDED Requirements

### Requirement: Clinic LLM SHALL 经 ChatRequest 显式 opt-in thinking

胖宝诊疗 LLM 调用（`streamClinicLLM` / `streamClinicLLMHeld`）在构造 `aimodel.ChatRequest` 时 MUST 设置 `ThinkingEnabled=true`（或后续等价的显式 enabled 字段），以确保经 aimodel 层发送 `thinking: enabled` 而非继承默认 disabled。Clinic WS 下行 `thinking_delta` / `answer_delta` 协议 MUST 保持不变。

#### Scenario: clinic 流式仍返回 thinking

- **WHEN** 客户端发送合法 `question` 且 clinic lane 上游返回 reasoning 流
- **THEN** voice-service MUST 继续映射为 `thinking_delta` 帧且上游请求 MUST 为 thinking enabled

#### Scenario: clinic 不受 aimodel 默认 disabled 影响

- **WHEN** aimodel 全局默认 thinking 为 disabled
- **THEN** clinic 路径 MUST 仍为 enabled 且 MUST NOT 因零值误关 thinking
