## Context

- T4（`sim-t4-inline-video-poll`）内联调用 `aimodel.PollVideoGeneration` → `GET https://open.bigmodel.cn/api/paas/v4/async-result/{taskID}`。
- 智谱官方 **AsyncVideoGenerationResponse** 字段：`task_status`（`PROCESSING`|`SUCCESS`|`FAIL`）、`video_result[]`（`url`、`cover_image_url`）。
- 当前 `parseVideoPollBody` 仅读 `video_url` / `output.video_url`，与官方格式不一致；`SUCCESS` 且无旧字段 URL 时被映射为 `VideoStatusProcessing`，T4 直至超时。

## Goals / Non-Goals

**Goals:**

- 正确解析 `video_result[0].url` 作为 `VideoPollResult.VideoURL`。
- `task_status` 大小写不敏感；`FAIL`/`failed` → failed；`SUCCESS`+url → success；`PROCESSING` 或 success 无 url → processing。
- 保留对 `video_url`、`output.video_url` 的 fallback（兼容旧响应或第三方代理）。

**Non-Goals:**

- 不改 `SubmitVideoGeneration` 请求体或模型名。
- T4 不使用 `cover_image_url` 发帖（可解析但不写入 UCG post）。
- 不新增 `*_test.go`；不引入 Redis。

## Decisions

### 1. 解析优先级

```
videoURL =
  video_result[0].url
  ?? video_url
  ?? output.video_url

status = lower(task_status || status)
```

| status 值 | videoURL | VideoPollResult |
|-----------|----------|-----------------|
| success/succeed/completed/done | 非空 | Success + URL |
| success/... | 空 | Processing |
| fail/failed/error | — | Failed |
| processing/空/其他 | — | Processing |

**备选**：仅读 `video_result` — 否决，保留 fallback 更安全。

### 2. 结构体扩展（可选）

`VideoPollResult` MAY 增加 `CoverImageURL string`（来自 `video_result[0].cover_image_url`），T4 首期忽略。

### 3. 日志

`PollVideoGeneration` 已有 Debug 全 body；解析失败时 Warning 保留 truncate，不打印 API Key。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 代理返回非标准 JSON | fallback 旧字段；解析 err 仍返回 error |
| SUCCESS 但 video_result 空数组 | 继续 processing，与官方「结果未就绪」语义一致 |
| 临时 URL 过期 | 现有 T4 成功路径立即 download，行为不变 |

## Migration Plan

1. 部署 **sim-user-service**（aimodel 包同镜像）。
2. 无需 DB 迁移；进行中的 T4 poll goroutine 下一 tick 即用新解析。
3. 验收：手动 T4，智谱返回 SUCCESS + video_result 时 job 应 `done` 而非 poll timeout。

## Open Questions

- 无。
