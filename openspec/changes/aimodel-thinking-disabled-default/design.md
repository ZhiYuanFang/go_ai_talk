## Context

- `internal/services/aimodel` 经 `buildRequestBody` → `getProviderAdapter().ApplyThinkingOptions` 合并 thinking 相关字段后调用上游 chat/completions。
- 当前 `ChatRequest.ThinkingEnabled` 为 `bool`：`true` 时写入 `thinking: enabled`（智谱）或 `extra_body.thinking: enabled`（DeepSeek）；`false` 时**不写任何 thinking 字段**。
- 智谱文档：GLM-4.7 系列在未传 `thinking` 时**默认开启** Thinking；须显式 `"thinking": { "type": "disabled" }` 方可关闭。
- sim-user（`LaneSimText` / `LaneSimVision`，默认 `glm-4.7-flash` / `glm-4.6v-flash`）及 UCG 润笔（`LanePolish`）调用均未设 `ThinkingEnabled`，实际走「省略字段 → 上游默认 thinking」，与 `MaxTokens` 短文案预算冲突。
- clinic（`LaneClinic`）经 `clinic_llm.go` 显式 `ThinkingEnabled: true`，须保持行为不变。

## Goals / Non-Goals

**Goals:**

- 对智谱 provider：**每次** chat/completions 请求 MUST 显式携带 `thinking.type`（`disabled` 或 `enabled`）。
- `ChatRequest` 零值（未 opt-in）映射为 **`disabled`**，sim/润笔/voice 非 thinking 路径无需改调用点即可关闭 thinking。
- clinic 等已设 `ThinkingEnabled: true` 的调用方行为不变。
- 在 `request.go` 补充中文注释，说明 `MaxTokens` 在 thinking enabled 时覆盖 reasoning+content 总预算。

**Non-Goals:**

- 不修改 `extractChatContent` 的 `reasoning_content` fallback（clinic 与非流式兼容仍需要；thinking disabled 后 sim 场景应主要命中 `content`）。
- 不调整 sim-user 各任务的 `MaxTokens` 数值（可在后续变更单独评估）。
- 不为 DashScope 等非 thinking provider 引入 disabled 字段（无默认 thinking 问题）。
- 不新增 Admin 配置项或 DB 字段。

## Decisions

### 1. 保留 `ThinkingEnabled bool`，语义改为 opt-in

**选择**：`false`（零值）→ 显式 `thinking: disabled`（智谱）；`true` → `thinking: enabled`（现有 clinic 逻辑）。

**备选**：新增 `ThinkingMode string`（`disabled|enabled|default`）→ 否决：仅两态且 clinic 已用 bool，扩 enum 增加迁移面。

**备选**：Profile 级默认 thinking → 否决：同 model 在 clinic 与 sim 需求相反，须 per-request 控制。

### 2. 智谱 adapter 始终写入 `thinking.type`

`provider_zhipu.go` 的 `ApplyThinkingOptions` 改为：

```go
thinkingType := "disabled"
if req.ThinkingEnabled {
    thinkingType = "enabled"
}
payload["thinking"] = map[string]interface{}{"type": thinkingType}
```

不再因 `!ThinkingEnabled` 而 early return。

### 3. DeepSeek adapter 保持「仅 enabled 时写入」

DeepSeek 路径仅在 `ThinkingEnabled == true` 时写入 `extra_body.thinking` 与 `reasoning_effort`；`false` 时不写 thinking 相关字段（DeepSeek 非默认强制 thinking）。

若后续 DeepSeek 也出现默认 thinking，再对齐智谱「显式 disabled」模式。

### 4. 调用方改动范围

| 调用方 | ThinkingEnabled | 变更后上游 |
|--------|-----------------|-----------|
| `clinic_llm.go` | `true` | enabled（不变） |
| `simuser/tasks.go` | 未设（false） | **disabled**（修复） |
| `ucg/compose_ai.go` | 未设（false） | **disabled** |
| voice 其他 LLM | 未设（false） | **disabled** |

voice 喂养 LLM 关闭 thinking 符合短问答场景，且降低延迟/成本。

### 5. 文档注释

在 `ChatRequest` 上补充：

- `ThinkingEnabled`：是否启用上游 thinking；默认 false，智谱会发 `disabled`。
- `MaxTokens`：completion 上限；thinking enabled 时 reasoning 与 content **共用**该预算。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 智谱对 `thinking: disabled` 与省略字段行为不一致 | 部署后在 test 环境对 `glm-4.7-flash` 打一条 sim 昵称 prompt，确认 `reasoning_content` 为空或极短、`content` 有值 |
| voice 喂养关闭 thinking 略降复杂推理质量 | 喂养场景以短回复为主；若回归可对该 lane 单独 opt-in（本变更不默认开启） |
| clinic 误删 `ThinkingEnabled` | spec 要求 clinic MUST opt-in；code review grep `ThinkingEnabled` |
| `extractChatContent` 仍将 reasoning 当正文（content 空时） | thinking disabled 后 sim 应极少触发；后续可另开变更禁止 sim 路径 fallback |

## Migration Plan

1. 合并并部署含 aimodel 改动的服务（`voice-service`、`ucg-service`、`sim-user-service` 凡依赖 aimodel 的进程）。
2. 无 DB/Redis 迁移；无 env 新增。
3. 回滚：revert aimodel adapter 改动即可恢复「省略 thinking 字段」旧行为。

## Open Questions

- `glm-4.6v-flash`（simVision 评论）在显式 disabled 后是否仍偶发 reasoning 字段——需 test 环境抽测 T2 comment 任务。
- 是否在 sim-user 任务失败日志中记录 `finish_reason=length` 以便观察 MaxTokens 是否仍偏紧（本变更不实现，可 follow-up）。
