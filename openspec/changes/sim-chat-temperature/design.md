## Context

- sim 任务已在 `tasks.go` 多处调用 `aimodel.Invoke`，仅设 `MaxTokens`，无 `temperature`。
- `ChatRequest.ExtraTopLevel` 可塞 temperature 但未使用；缺乏类型安全与 spec 约定。
- 负责人决策：**任务内常量**即可，不上 sim-admin / DB；范围覆盖 sim **全部 chat Invoke**（5 处）。

## Goals / Non-Goals

**Goals:**

- aimodel 平台化支持 `Temperature *float64`（nil = 不写入 body）。
- sim 包内集中定义各任务默认温度，经 helper 构造 `ChatRequest`。
- 5 处 Invoke：T1 昵称、T2 评论、T3 配文、T4 视频文案、E1 聊天回复。

**Non-Goals:**

- CogView/CogVideo 生图生视频 temperature（API 非 chat completions）。
- `sim_llm_lane_config` 增列、sim-admin UI、env 覆盖。
- voice/clinic/ucg polish 默认传 temperature（保持 nil）。

## Decisions

### 1. aimodel：`Temperature *float64`

```go
if req.Temperature != nil {
    payload["temperature"] = *req.Temperature
}
```

流式与非流式共用 `buildRequestBody`。不使用 `0` 作为「未设置」哨兵，避免误传极低温度。

### 2. sim 常量（首期默认值）

| 任务 | 常量 | 值 |
|------|------|-----|
| T1 register 昵称 | `simTempRegisterNickname` | 0.90 |
| T2 comment | `simTempComment` | 0.85 |
| T3 post_image 配文 | `simTempPostImageText` | 0.80 |
| T4 post_video 文案 | `simTempPostVideoText` | 0.80 |
| E1 chat_reply | `simTempChatReply` | 0.88 |

文件：`internal/services/simuser/task_llm_temp.go`，含中文注释说明调参需 redeploy。

### 3. Helper

```go
func simChatRequest(temp float64, maxTokens int, messages []aimodel.Message) aimodel.ChatRequest
```

各任务替换内联 `aimodel.ChatRequest{...}` 为 helper 调用，保留现有 `MaxTokens` 数值不变。

### 4. 与 thinking 默认 disabled 关系

sim Invoke 不传 `ThinkingEnabled`（false）；temperature 独立字段，互不影响。

## Risks / Trade-offs

- **[Risk] 温度过高导致废话或 Green 拒评** → T2/T3 用 0.80–0.85；运营反馈后改常量发版。
- **[Risk] 智谱 model 忽略 temperature** → test 环境抽测 T2/T1 各一条；无效则再查 provider adapter。
- **[Risk] 常量分散** → 强制单文件 `task_llm_temp.go`，禁止 tasks.go 内联浮点。

## Migration Plan

1. 部署含 aimodel 改动的 **sim-user-service**（及同二进制依赖的进程若重编，行为不变）。
2. 回滚：revert sim 常量或整个 change；voice/ucg 无影响。
3. 验收：debug 日志或抓包可见 upstream body 含 `temperature`；sim 评论/回复体感多样度提升。

## Open Questions

- 无；admin 可调 temperature 留后续 change。
