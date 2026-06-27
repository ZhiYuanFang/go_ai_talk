## Why

sim 各任务经 `aimodel.Invoke` 生成昵称、评论、配文与聊天回复，但请求体未传 `temperature`，完全依赖上游默认值，导致输出风格单一、评论/回复模板感强。须在 aimodel 层支持显式 temperature，并在 sim-user-service 各任务以**任务内常量**传入，提升内容多样性；调参通过改常量并发版 sim，首期不上 admin/DB。

## What Changes

- `aimodel.ChatRequest` 增加可选 `Temperature *float64`；`buildRequestBody` 在非 nil 时写入 JSON `temperature`。
- sim-user-service 新增任务级 temperature 常量与 `simChatRequest` helper；**全部 5 处** `aimodel.Invoke` 必须使用对应常量。
- 生图（`GenerateImage`）、生视频（`SubmitVideoGeneration`）、轮询（`PollVideoGeneration`）首期不改（智谱 API 无 chat 式 temperature 或不在 scope）。
- voice/ucg 未传 `Temperature` 时行为不变（不传字段，走上游默认）。
- 不新增 `*_test.go`；不扩展 `sim_llm_lane_config`。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `aimodel-media-gen`（delta）：chat/completions 经 `Invoke` 时 MUST 支持可选 `temperature` 请求字段。
- `sim-user-service`：各 LLM 文本/vision 任务 MUST 经任务内常量设置 temperature。

## Impact

- **代码**：`internal/services/aimodel/request.go`、`client.go`；`internal/services/simuser/task_llm_temp.go`（新）、`tasks.go`。
- **进程**：`sim-user-service`（主要）；`aimodel` 为共享库，voice/ucg 进程链接同一包但默认无行为变化。
- **配置**：无 env；常量表见 design.md。
