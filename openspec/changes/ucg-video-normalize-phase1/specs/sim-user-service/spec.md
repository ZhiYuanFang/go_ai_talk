## MODIFIED Requirements

### Requirement: T4 post video SHALL inline poll async-result until post or failure

`RunPostVideoSubmitTask`（调度与手动共用）在 `SubmitVideoGeneration` 成功且 `InsertVideoJob` 写入 `pending` 后 MUST 启动视频结果轮询。轮询 MUST 调用智谱 `GET /paas/v4/async-result/{task_id}`（经 `aimodel.PollVideoGeneration`）。轮询 MUST 使用 submit 阶段已获得的 `loginSession`，MUST NOT 再经分页 list 线性查 account。

- **success**：下载视频 → **ucg internal `upload-video` 转码为 v2** → 用户 token `RegisterMedia`（`transformVersion=v2`）→ `POST /ucg/app/api/posts`（`submit=true`）→ `sim_video_job=done` → `RecordTaskRun("post_video_submit", true, ...)`
- **failed**（上游明确失败）：`sim_video_job=skipped` → `RecordTaskRun(..., false, ...)`
- **processing / pending**：在 `postVideoPollInterval` 后重试，直至 `now >= submitTime + postVideoPollMaxWait` → 超时视为发布失败 → `skipped` + `RecordTaskRun(..., false, ...)`

MUST NOT 在 submit 成功时单独写 `RecordTaskRun` success。MUST NOT 使用 presign 直传未转码视频或 `transformVersion=sim-raw` register。

#### Scenario: Poll success posts video via transcode upload

- **WHEN** T4 提交后 async-result 返回 success 且 internal 转码上传与 register v2、发帖均 OK
- **THEN** job MUST 为 `done` 且 `sim_task_run` MUST 记 post_video_submit 成功

#### Scenario: Poll timeout fails task

- **WHEN** 自 submit 起超过 `postVideoPollMaxWait` 仍为 processing
- **THEN** job MUST 为 `skipped` 且 `sim_task_run` MUST 记 post_video_submit 失败

#### Scenario: T4 does not presign raw video

- **WHEN** T4 下载智谱视频完成
- **THEN** sim-user-service MUST 调用 ucg `POST /ucg/internal/api/media/upload-video` 且 MUST NOT 对原始下载字节执行 OSS presign PUT

## ADDED Requirements

### Requirement: sim T4 video upload SHALL use ucg internal transcode API

sim-user-service 在上传 T4 视频至 UCG 时 MUST 经 HTTP 调用 ucg-service `POST /ucg/internal/api/media/upload-video`（与 device internal 鉴权方式一致），MUST 使用返回的 `objectKey` 与 `contentHash` 调用 App API `RegisterMedia`（`mediaKind=2`、`transformVersion=v2`、`dedupHit=false`）。

MUST NOT 再使用 `transformVersion=sim-raw` 或 presign 直传未验真/未转码字节。

#### Scenario: T4 registers v2 after internal upload

- **WHEN** internal 转码上传成功
- **THEN** 后续 register MUST 使用 `transformVersion=v2` 且 MUST 通过 ucg 侧 v2 验真

#### Scenario: Internal transcode failure skips job

- **WHEN** internal 转码上传失败
- **THEN** T4 流水线 MUST 将 job 标为 skipped 且 MUST NOT 使用 presign 回退直传
