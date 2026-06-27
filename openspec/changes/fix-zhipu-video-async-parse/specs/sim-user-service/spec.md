## MODIFIED Requirements

### Requirement: T4 post video SHALL inline poll async-result until post or failure

`RunPostVideoSubmitTask`（调度与手动共用）在 `SubmitVideoGeneration` 成功且 `InsertVideoJob` 写入 `pending` 后 MUST 启动视频结果轮询。轮询 MUST 调用智谱 `GET /paas/v4/async-result/{task_id}`（经 `aimodel.PollVideoGeneration`）。`PollVideoGeneration` MUST 正确解析响应中的 `video_result[].url`（见 aimodel-media-gen）。轮询 MUST 使用 submit 阶段已获得的 `loginSession`，MUST NOT 再经分页 list 线性查 account。

#### Scenario: Inline poll success posts once

- **WHEN** T4 提交后 async-result 返回 `task_status=SUCCESS` 且 `video_result[0].url` 有效
- **THEN** 系统 MUST 下载、发帖、job=done，且 `RecordTaskRun(post_video_submit)` 在流水线结束时 success

#### Scenario: Poll timeout only when still processing

- **WHEN** 在 maxWait 内始终为 PROCESSING 或 SUCCESS 无 url
- **THEN** job MUST 为 skipped 且 errMsg 含 poll timeout；MUST NOT 在已成功解析 video_result url 后仍 timeout
