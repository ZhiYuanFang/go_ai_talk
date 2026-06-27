## MODIFIED Requirements

### Requirement: sim LLM text and vision SHALL use Invoke with sim lanes

昵称、文案、聊天、评论生成 MUST 经 `aimodel.Invoke`（或 `InvokeStream` 若需要）与对应 sim lane，MUST NOT 在 sim-service 内直连智谱 HTTP 绕过闸门。

sim-user-service 内每一处 `aimodel.Invoke`（chat/completions）MUST 在 `ChatRequest` 中设置非 nil 的 `Temperature`，取值 MUST 来自 `simuser` 包内任务级命名常量（集中定义于单文件，如 `task_llm_temp.go`），MUST NOT 在 `tasks.go` 内联魔法浮点数。首期 MUST 覆盖：T1 注册昵称、T2 评论、T3 图文配文、T4 视频文案、E1 聊天回复。

#### Scenario: Comment invoke sets temperature

- **WHEN** `RunCommentTask` 调用 `aimodel.Invoke` with `LaneSimVision`
- **THEN** `ChatRequest.Temperature` MUST 非 nil 且 MUST 等于包内 `simTempComment` 常量

#### Scenario: All sim Invoke sites use task constants

- **WHEN** 代码评审 `internal/services/simuser` 内全部 `aimodel.Invoke`
- **THEN** 每一处 MUST 经统一 helper 或等价路径传入任务 temperature 常量；MUST 恰好 5 处（不含 `GenerateImage` / `SubmitVideoGeneration` / `PollVideoGeneration`）

#### Scenario: Shared gate with voice

- **WHEN** voice-service 与 sim-service 同时调用 `glm-4.7-flash`
- **THEN** 二者 MUST 竞争同一 inflight 上限

## ADDED Requirements

### Requirement: sim chat temperature constants SHALL be documented for redeploy tuning

任务 temperature 常量 MUST 附中文注释说明业务用途；变更常量 MUST 通过 redeploy `sim-user-service` 生效。首期 MUST NOT 依赖 env 或 sim-admin 配置项。

#### Scenario: Constants live in one file

- **WHEN** 代码评审 sim temperature 定义
- **THEN** 全部任务常量 MUST 位于同一 Go 源文件（如 `task_llm_temp.go`）
