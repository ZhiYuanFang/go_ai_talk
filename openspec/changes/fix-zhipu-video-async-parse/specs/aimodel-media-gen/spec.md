## MODIFIED Requirements

### Requirement: aimodel SHALL support image and video generation for sim

包 MUST 导出：

- `GenerateImage(ctx, lane, prompt) (result, err)` — CogView-3-Flash，POST 时 `Acquire`，释放后返回可下载 URL 或字节
- `SubmitVideoGeneration(ctx, lane, prompt) (taskID, err)` — CogVideoX-Flash 提交，POST 时 `Acquire`
- `PollVideoGeneration(ctx, taskID) (VideoPollResult, err)` — GET `/paas/v4/async-result/{taskID}` 轮询，MUST NOT 占用 inflight 槽

`PollVideoGeneration` 解析 MUST 支持智谱 **AsyncVideoGenerationResponse**：

- 状态：`task_status`（`PROCESSING`、`SUCCESS`、`FAIL`，大小写不敏感）；MAY fallback 顶层 `status`
- 视频 URL MUST 优先 `video_result[0].url`；MAY fallback `video_url` 或 `output.video_url`
- `SUCCESS` 且 URL 非空 → `VideoStatusSuccess`；`SUCCESS` 无 URL → `VideoStatusProcessing`；`FAIL` → `VideoStatusFailed`；`PROCESSING`/其他 → `VideoStatusProcessing`

#### Scenario: Image generation acquires gate

- **WHEN** sim 调用 GenerateImage
- **THEN** 上游 HTTP 请求期间 MUST 持有 `cogview-3-flash` 槽位并在完成后释放

#### Scenario: Poll does not acquire gate

- **WHEN** sim 调用 PollVideoGeneration
- **THEN** MUST NOT 调用 `Acquire` inflight

#### Scenario: Official SUCCESS with video_result

- **WHEN** async-result 响应为 `{"task_status":"SUCCESS","video_result":[{"url":"https://example.com/a.mp4"}]}`
- **THEN** PollVideoGeneration MUST 返回 status=success 且 VideoURL 为上述 url

#### Scenario: PROCESSING without video_result

- **WHEN** 响应为 `{"task_status":"PROCESSING"}`
- **THEN** MUST 返回 status=processing 且 VideoURL 为空

#### Scenario: FAIL task

- **WHEN** 响应为 `{"task_status":"FAIL"}`
- **THEN** MUST 返回 status=failed

#### Scenario: Legacy video_url fallback

- **WHEN** 响应为 `{"task_status":"SUCCESS","video_url":"https://legacy.example/v.mp4"}` 且无 video_result
- **THEN** MUST 返回 status=success 且 VideoURL 为 legacy url
