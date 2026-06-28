## Context

- 现状：`BuildVideoSnapshotURL` 对 `BuildCdnURL(videoObjectKey)` 追加 `?x-oss-process=video/snapshot,t_0`。用于 `post.go` `loadPostMedia`、`notification.go` 通知封面、`post_sample_internal.go` LLM 采样封面等；聊天 `enrichChatMessageMedia` 对视频**未**填充 `mediaThumbnailUrl`（仅图片有）。
- 图片物理缩略图（`ucg-image-thumb-physical`）已落地：`mediacdn.ThumbObjectKey` → `{stem}_thumb.{ext}`，`EnsureImageThumb` 写路径 materialize，读路径 `BuildImageThumbnailURL` 无 query。
- 约束：仅 `ucg-service`；无 Redis 新增；无 DB 变更；禁止新 background ticker；不新增 `*_test.go`；v1/v2 API 字段结构不变。

## Goals / Non-Goals

**Goals:**

- 视频首帧缩略图使用独立 OSS 对象 `{stem}_thumb.jpg`，CDN URL 与 mp4 原片 path 隔离，避免 CDN 窜缓存与 Content-Type 混乱。
- 写路径（register、服务端直传）同步 materialize thumb；读路径统一 `BuildVideoThumbnailURL`。
- 帖子/动态 `thumbnailUrl`、聊天 `mediaThumbnailUrl`（含 WS 实时 v1）对视频可用。
- 删除视频原片时成对删除 `_thumb.jpg`。

**Non-Goals:**

- 历史视频 backfill CLI（用户手动重传/重编辑）。
- 修改 v1/v2 HTTP/WS 响应 JSON 结构（仅 URL 值变化）。
- 事件 logo、图片 thumb 逻辑改动。
- Flutter 仓库代码变更（本仓库 tasks 末项仅跟踪验证/必要时在 sibling repo 修改）。

## Decisions

### 1. 命名与 helper（mediacdn）

- **决定**：在 `internal/shared/mediacdn` 新增 `VideoThumbExt = "jpg"`、`VideoThumbObjectKey(videoObjectKey)`。
- **规则**：对 `{path}/{stem}.mp4`（及受支持视频 ext），thumb key 为 `{path}/{stem}_thumb.jpg`；扩展名**固定 jpg**，与 mp4 不同（与图片 `{stem}_thumb.{原 ext}` 区分）。
- **示例**：`social/2026/06/xyz.mp4` → `social/2026/06/xyz_thumb.jpg`。
- **IsThumbObjectKey**：已有 `_thumb` stem 判定对 `*_thumb.jpg` 仍有效；`VideoThumbObjectKey` 对已是 thumb key 的输入 MUST 原样返回。
- **理由**：列表/聊天统一 jpg 预览；path 与原视频不同，CDN cache key 独立。

### 2. CDN 缓存与 OSS snapshot 角色（用户关切）

- **问题**：使用 `video.mp4?x-oss-process=video/snapshot` 是否导致 CDN 缓存问题？
- **答案**：这正是本变更要**避免**的问题。query 截帧 URL 与原 mp4 URL 共享 path 前缀，CDN/客户端常按 path 或错误 Content-Type 缓存，导致 `<img>` 拿到 mp4 字节或截帧与播放 URL 窜包。物理对象 `xyz_thumb.jpg` 拥有**独立 path 与 cache namespace**，与图片 `_thumb` 模式一致——读路径**无 query**。
- **OSS snapshot 仅用于写路径一次**：`EnsureVideoThumb` 内 `GetObject(mp4, oss.Process("video/snapshot,t_0"))` 取字节 → `PutObject(xyz_thumb.jpg)`；客户端与 CDN 读路径 NEVER 带 `x-oss-process`。

### 3. EnsureVideoThumb 实现

- **决定**：`EnsureVideoThumb(ctx, videoObjectKey)` 镜像 `EnsureImageThumb` 流程：
  1. 校验为视频 objectKey（非 thumb、ext 为 mp4 等受支持格式）；
  2. `HEAD` thumb 存在则返回；
  3. `HEAD` 原视频不存在则错误；
  4. `GetObject(原视频, oss.Process("video/snapshot,t_0"))` 读字节；
  5. `PutObject(thumbKey, data, ContentType: image/jpeg)`。
- **备选**：客户端本地截帧上传——增加客户端复杂度与一致性风险；未采用。
- **失败语义**：register/直传路径失败 MUST 阻断（避免登记成功但列表无 thumb）。

### 4. 写路径挂钩

| 路径 | 时机 |
|------|------|
| `RegisterMedia` 成功（`mediaKind==2`） | 事务成功后、`return` 前 `EnsureVideoThumb` |
| `RegisterMedia` dedup hit | 幂等 `EnsureVideoThumb` |
| `putOSSObject` | `PutObject` 原视频成功后，若 `mediaKind==2` 则 `EnsureVideoThumb` |

`PresignUpload` 不生成 thumb（客户端尚未 PUT）。

### 5. 读路径

- 新增 `BuildVideoThumbnailURL(objectKey)` → `BuildCdnURL(VideoThumbObjectKey(objectKey))`。
- **替换** `BuildVideoSnapshotURL` 调用：
  - `post.go` `loadPostMedia`（视频 `thumbnailUrl`）
  - `chat_store.go` `enrichChatMessageMedia`（视频 `mediaThumbnailUrl`）
  - `notification.go` `resolvePostCoverSnapshot`
  - `post_sample_internal.go` `postSampleCoverCdnURL`
- `BuildVideoSnapshotURL` MAY 保留为 deprecated 内部别名或删除（实现阶段若无其他引用则删除）。

### 6. 删除成对

- `deletePairedThumbObject` 需区分媒体类型：图片用 `ThumbObjectKey`，视频用 `VideoThumbObjectKey`（或统一 `PairedThumbObjectKey(objectKey, mediaKind)`）。
- `DeleteOwnedMedia` 删除 mp4 后 MUST 删除对应 `_thumb.jpg`（404 忽略）。

### 7. 历史数据

- **无 backfill**。已上线视频无物理 thumb 时，读路径 URL 将 404——接受；用户通过重编辑/重上传触发写路径。
- 通知表已固化的 `BuildVideoSnapshotURL` 历史行本变更不批量改写（与图片变更一致）。

### 8. Chat WS v1

- `enrichChatMessageMedia` 在 `chat_persist.go`、`chat_store.go`、`chat_service.go` 等所有出站 enrich 点生效；WS `message_delivered` / 历史拉取 MUST 含视频 `mediaThumbnailUrl`（首版发布，非 follow-up）。

## Risks / Trade-offs

- [历史视频列表裂图] → 无 backfill；文档与运营知悉需重传；新上传立即正常。
- [Register 延迟增加] → 单次 OSS snapshot + PUT；与图片 thumb 同级，可接受。
- [thumb 为 jpg 非原视频分辨率] → 列表预览足够；全屏仍用 `cdnUrl` 播放。
- [EnsureVideoThumb 失败阻断 register] → 可观测；与图片策略一致。
- [delete 误删/漏删] → 按 mediaKind 选 key helper；404 忽略。

## Migration Plan

1. 部署含写+读路径的单次 ucg-service 发布（或先写后读两阶段，推荐一次发布减少历史窗口）。
2. test 环境：上传新视频 → 验证 OSS 成对 `*.mp4` + `*_thumb.jpg` → 验证 Feed/帖子/聊天/Admin 列表 thumb URL 无 `x-oss-process`。
3. prod 重复验证；历史视频按需人工重传。
4. 回滚：读路径可临时恢复 `BuildVideoSnapshotURL`；已写入 thumb 对象无害可保留。

## Open Questions

- 无（命名 `_thumb.jpg`、无 backfill、OSS snapshot 仅写路径、WS v1 含 `mediaThumbnailUrl` 均已确认）。
