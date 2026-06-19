## ADDED Requirements

### Requirement: aimodel SHALL expose sim lanes sharing Redis gates by model

`internal/services/aimodel` MUST 新增 Lane：`simText`、`simVision`、`simImageGen`、`simVideoGen`。`sim-user-service` 启动时 MUST 注册 `ProfileStore` 提供上述 lane 的 `provider`、`model`、`maxInFlight`、`maxWaiters`。

默认 model MUST 为：

| Lane | Model |
|------|-------|
| simText | glm-4.7-flash |
| simVision | glm-4.6v-flash |
| simImageGen | cogview-3-flash |
| simVideoGen | cogvideox-flash |

`Acquire` MUST 使用与现有服务相同的 Redis 键 `ai:llm:gate:{normalizedModel}:*`，使同 model 跨进程共用并发池。

#### Scenario: Shared gate with voice

- **WHEN** voice-service 与 sim-service 同时调用 `glm-4.7-flash`
- **THEN** 二者 MUST 竞争同一 inflight 上限

### Requirement: aimodel SHALL support image and video generation for sim

包 MUST 导出：

- `GenerateImage(ctx, lane, prompt) (result, err)` — CogView-3-Flash，POST 时 `Acquire`，释放后返回可下载 URL 或字节
- `SubmitVideoGeneration(ctx, lane, prompt) (taskID, err)` — CogVideoX-Flash 提交，POST 时 `Acquire`
- `PollVideoGeneration(ctx, taskID) (status, videoURL, err)` — GET 轮询状态，MUST NOT 占用 inflight 槽

#### Scenario: Image generation acquires gate

- **WHEN** sim 调用 GenerateImage
- **THEN** 上游 HTTP 请求期间 MUST 持有 `cogview-3-flash` 槽位并在完成后释放

#### Scenario: Poll does not acquire gate

- **WHEN** sim 调用 PollVideoGeneration
- **THEN** MUST NOT 调用 `Acquire` inflight

### Requirement: sim LLM text and vision SHALL use Invoke with sim lanes

昵称、文案、聊天、评论生成 MUST 经 `aimodel.Invoke`（或 `InvokeStream` 若需要）与对应 sim lane，MUST NOT 在 sim-service 内直连智谱 HTTP 绕过闸门。

#### Scenario: Comment uses simVision

- **WHEN** T2 生成评论
- **THEN** 调用 MUST 使用 `LaneSimVision` 且受 polish 同 model 池约束
