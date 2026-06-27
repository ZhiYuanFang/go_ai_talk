## Why

sim T2 评论虽配置 `simVision` lane 且 prompt 要求「结合帖子内容和图片」，但实现仅将帖子正文文本传入 LLM，`coverObjectKey` 未使用，导致图文/视频帖评论与配图无关。须在单次 `simVision` 调用中传入封面 CDN URL、帖子正文与评论目标 prompt，使模型同时理解图与文并直接输出评论。

## What Changes

- ucg internal `posts/sample` 响应增加 `coverCdnUrl`（图文用全图 CDN，视频用首帧 snapshot URL；无媒体或无法拼装时省略）。
- sim T2 `RunCommentTask`：有媒体 URL 时以 OpenAI 兼容多模态 `contentParts`（`image_url` + `text`）调用 `LaneSimVision`；纯文字帖仍仅传文本。
- 默认 `comment` prompt 微调：明确只输出评论正文、结合配图与帖子正文。
- 有媒体但 `coverCdnUrl` 为空时降级为纯文本评论（打 warning 日志），不失败整任务。
- 不新增第二次 LLM 调用；不新增 Redis；不新增 `*_test.go`。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-sim-feed-sample`：sample 项 MUST 在有媒体时提供 LLM 可用的 `coverCdnUrl`。
- `sim-user-service`：T2 MUST 经单次 `simVision` 多模态调用结合正文与封面图（或视频首帧）生成评论；纯文字帖 MUST 仍可用文本-only 调用。

## Impact

- **代码**：`internal/services/ucg/post_sample_internal.go`、`api/v1/ucg_internal_posts_sample_http.go`、`internal/services/simuser/tasks.go`、`internal/services/simuser/schema.go`（默认 prompt）。
- **进程**：`ucg-service`、`sim-user-service`；无 DB 迁移。
- **部署**：先 ucg（含 `coverCdnUrl`）→ 再 sim；向后兼容（旧 sim 忽略新字段仍可跑，但无图）。
