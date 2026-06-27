## Context

- T2 已走 `posts/sample` `mode=random` 选帖；sample 返回 `coverObjectKey` 但 sim 未消费。
- OpenSpec 要求 T2「经 simVision 结合正文与图片」；当前为 text-only 假 vision。
- UCG `PolishPostText` 已有成熟模式：`BuildCdnURL` + `image_url` content part + text prompt，单次 `aimodel.Invoke`。
- `sim-user-service` 无 OSS/CDN 配置，**MUST NOT** 在 sim 内拼 CDN URL；URL 由 ucg sample 侧生成。

## Goals / Non-Goals

**Goals:**

- 单次 `LaneSimVision`：有图时 `[image_url, text(prompt+post_content)]` → 评论正文。
- sample 返回 `coverCdnUrl`，按 `mediaType` 选择 builder。
- 纯文字帖（`mediaType=0` 或无有效 URL）文本-only 调用。
- CDN URL 不可用时降级纯文本，任务仍尝试完成。

**Non-Goals:**

- 两阶段「先 pcontent 再评论」（不增加 LLM 次数）。
- 多图全量传入（仅封面/首条 media，与现有 sample 一致）。
- sim 侧新增 CDN env 或跨域 import ucg `BuildCdnURL`。
- 修改 App 对外 API 或 recommend Feed。

## Decisions

### 1. sample 增 `coverCdnUrl`

在 `PostSampleItem` / API 契约增加 `coverCdnUrl`（optional string）：

| mediaType | coverObjectKey 有值时 |
|-----------|----------------------|
| 1 (images) | `BuildCdnURL(key)` |
| 2 (video) | `BuildVideoSnapshotURL(key)`（首帧，非 mp4 直链） |
| 0 (none) | 不填 |

在 `postSampleFromRows` 或行映射 helper 内根据 `media_type` + `cover_object_key` 填充；SQL 不变（仍查首条 media objectKey）。

### 2. T2 多模态消息结构

对齐 `compose_ai.go`：

```go
// 有 coverCdnUrl
contentParts := []map[string]interface{}{
  {"type": "image_url", "image_url": map[string]string{"url": coverCdnUrl}},
  {"type": "text", "text": renderedCommentPrompt},
}
Messages: []aimodel.Message{{Role: "user", Content: contentParts}}

// 无 coverCdnUrl
Messages: []aimodel.Message{{Role: "user", Content: renderedCommentPrompt}}
```

`renderedCommentPrompt` 来自 `LoadRenderedPrompt(ctx, "comment", map[string]string{"post_content": post.Content})`。

### 3. 默认 comment prompt

```
作为宝妈，请结合上方配图与下列帖子正文，写一条简短、口语化的评论。
要求：只输出评论正文，不要解释、不要引号包裹。
帖子正文：{{post_content}}
```

无图时「上方配图」语义无害；或由模板保持原句「结合帖子内容和图片」亦可。

### 4. 降级语义

`mediaType != 0` 且 `coverCdnUrl == ""`（OSS 未配、key 空等）：`glog.Warningf` 后走 text-only，不 abort T2。

### 5. 辅助函数位置

在 `simuser` 包内小 helper `commentVisionContent(cdnURL, text string) interface{}`，避免 tasks.go 膨胀；不新建跨包抽象。

## Risks / Trade-offs

- **[Risk] LLM 无法拉取 CDN URL（内网/鉴权）** → 与 polish 相同前提，CDN 须公网可达；失败时上游报错记 task 失败。
- **[Risk] 视频 snapshot URL 偶发慢** → T2 低频可接受；与 Feed 缩略图一致。
- **[Risk] 单图代表多图帖** → 与 sample 设计一致；Non-Goal 全图。

## Migration Plan

1. 部署 ucg-service（sample 含 `coverCdnUrl`）。
2. 部署 sim-user-service（多模态 T2）。
3. 回滚：sim 回 text-only；ucg 多返回字段对旧客户端无害。
4. 验收：对图文帖 T2 评论与配图相关；日志可见 multimodal 路径；无 `feed/recommend`。

## Open Questions

- 是否在 `RecordTaskRun` 的 success message 附带是否 multimodal — 首期否，仅日志。
