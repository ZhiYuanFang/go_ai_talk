## Why

智谱 GLM-4.7 系列在 API 未显式传 `thinking` 时**默认开启思考模式**，`max_tokens` 由 `reasoning_content` 与 `content` **共同消耗**。当前 `aimodel` 仅在 `ChatRequest.ThinkingEnabled == true` 时写入 `thinking: enabled`，否则**不传**该字段——调用方（如 sim-user 短文案任务）误以为 `MaxTokens` 仅约束最终正文，实际可能被 reasoning 占满配额，甚至触发 `extractChatContent` 将 `reasoning_content` 当作业务正文。需要在 `aimodel` 层**统一显式关闭 thinking，且默认关闭**，仅 clinic 等明确需要 reasoning 的链路 opt-in 开启。

## What Changes

- `aimodel.ChatRequest` 扩展 thinking 语义：**默认 `disabled`**；调用方显式 opt-in 才发送 `thinking: { type: "enabled" }`（及现有 `ReasoningEffort` 等）。
- 智谱（`zhipuAdapter`）与 DeepSeek（`deepseekAdapter`）在构建请求体时，对支持 thinking 的 provider **始终**写入 `thinking.type`（`disabled` 或 `enabled`），不再出现「省略字段 → 上游默认开启」。
- 保留现有 `ThinkingEnabled bool` 作为 opt-in 开关（`true` → enabled），避免 clinic 等调用方大面积改动；零值 `false` 映射为显式 `disabled`。
- sim-user、UCG 润笔等非 thinking 场景无需改调用代码即可受益（默认关闭）。
- clinic 现有 `ThinkingEnabled: true` 行为不变。
- **非 BREAKING（API 契约）**：对外 HTTP/WS 接口不变；仅 LLM 上游请求体语义更明确。

## Capabilities

### New Capabilities

- `aimodel-thinking-mode`：`aimodel` 统一 thinking 模式（默认 disabled、opt-in enabled）、各 provider 请求体写入规则及与 `MaxTokens` 的边界说明。

### Modified Capabilities

- `llm-lane-gate`：补充 Requirement——经 `aimodel.Invoke`/`InvokeStream` 发起的 chat/completions 请求 MUST 显式携带 thinking 模式；默认 MUST 为 `disabled`。
- `pangbao-ai-clinic`：Clinic LLM MUST 继续 opt-in `thinking: enabled`（经 `ThinkingEnabled` 或等价字段），行为与基线一致。

## Impact

- **代码**：`internal/services/aimodel/request.go`、`provider_zhipu.go`、`provider_deepseek.go`（及 `buildRequestBody` 若需调整）；调用方仅审阅，clinic 已 opt-in，sim/ucg 默认受益。
- **配置/DB**：无。
- **Redis / 部署**：无。
- **App API 使用统计**：无新增 App HTTP 路由。
- **基线**：对照 `openspec/specs/v2.0.5/spec.md` 中 `pangbao-ai-clinic`（thinking 开启）、`llm-lane-gate`（aimodel 统一入口）；sim-user 行为改善属实现层修正，不改变对外契约。
