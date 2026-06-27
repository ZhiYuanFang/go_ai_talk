## Why

T4 内联轮询经 `aimodel.PollVideoGeneration` 调用智谱 `GET /paas/v4/async-result/{id}`。官方视频完成响应为 `task_status` + `video_result[].url`，当前 `parseVideoPollBody` 仍解析已废弃的顶层 `video_url`/`output.video_url`。当 API 返回 `SUCCESS` 且视频 URL 仅在 `video_result` 中时，解析器误判为 **processing**，导致 T4 轮询直至 `postVideoPollMaxWait` 超时并将 job 标为 skipped，尽管远端视频已生成成功。

## What Changes

- 修正 `parseVideoPollBody`：优先读取 `video_result[0].url`；保留 `task_status`（`PROCESSING`/`SUCCESS`/`FAIL`）及旧字段 fallback 以兼容代理或历史格式。
- `SUCCESS` 且无 URL 时仍视为 processing（结果未就绪）；`FAIL` 映射 failed；有 URL 的 success 映射 success 并供 T4 下载发帖。
- 不改动 T4 调度、`sim_video_job` 状态机或 Admin 配置；不新增 `*_test.go`。

## Capabilities

### New Capabilities

（无 — 在既有 aimodel 媒体生成能力上修正解析。）

### Modified Capabilities

- `sim-user-service`：`PollVideoGeneration` / T4 轮询 MUST 正确识别智谱 `AsyncVideoGenerationResponse`（`video_result`）。
- `aimodel-media-gen`（增量 capability）：智谱视频 async-result 解析语义（若归档时合并至 aimodel 基线）。

## Impact

- **代码**：`internal/services/aimodel/media_gen.go`（`parseVideoPollBody`，可选 `VideoPollResult` 增加 `CoverImageURL` 字段但不强制 T4 使用）。
- **进程**：仅 **sim-user-service**（aimodel 包同进程）；部署 sim-user 即可生效。
- **OpenSpec**：delta 挂于本 change；与 `sim-t4-inline-video-poll` 互补（该 change 负责流水线，本 change 修复轮询结果解析）。
