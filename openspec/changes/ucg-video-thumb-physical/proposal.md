## Why

UCG 视频首帧缩略图当前通过 `BuildVideoSnapshotURL` 在原视频 CDN URL 上拼接 `?x-oss-process=video/snapshot`。CDN 常按 path 缓存，query 截帧与 mp4 原片易窜包或返回错误 Content-Type，列表/聊天 `<img>` 加载失败或展示异常。图片已改为物理 `_thumb` 对象后，视频需同样在上传时 materialize 独立 jpg 对象，读路径返回无 query 的 CDN URL。

## What Changes

- 新增 `mediacdn.VideoThumbObjectKey`（或等价 helper）：视频 `{stem}.mp4` → `{stem}_thumb.jpg`（扩展名固定 jpg，与 mp4 不同）。
- 新增 `EnsureVideoThumb(ctx, videoObjectKey)`：HEAD 幂等跳过；经 OSS `GetObject` + `video/snapshot,t_0` 拉取首帧字节后 PUT 至 thumb key，`Content-Type: image/jpeg`。
- 写路径：`RegisterMedia`（`mediaKind==2`）、`putOSSObject` 视频直传成功后同步调用 `EnsureVideoThumb`（与图片 `EnsureImageThumb` 对称）。
- 读路径：`BuildVideoThumbnailURL` 返回 `BuildCdnURL(VideoThumbObjectKey(key))`；替换 `post.go` `loadPostMedia`、`chat_store.go` `enrichChatMessageMedia` 中的 `BuildVideoSnapshotURL`；聊天视频 MUST 填充 `mediaThumbnailUrl`。
- Chat WebSocket 实时消息（v1）MUST 在 `enrichChatMessageMedia` 后包含视频 `mediaThumbnailUrl`（首版即上线，非后续迭代）。
- 删除媒体时成对删除视频 thumb 对象（沿用 `oss_delete.go` 模式）。
- 同变更内替换 `notification.go`、`post_sample_internal.go` 中 `BuildVideoSnapshotURL` 为物理 thumb URL（通知封面、LLM 采样封面）。
- **不**新增 backfill CLI；历史视频无 thumb 的对象由用户手动重传/重编辑，显式 out of scope。
- API 响应结构（v1/v2 字段名）不变，仅 URL 值由 query 截帧改为物理 jpg path。

## Capabilities

### New Capabilities

- `ucg-video-thumb`：视频 thumb 命名（`_thumb.jpg`）、`EnsureVideoThumb` 生成、`BuildVideoThumbnailURL` 读路径、删除成对、无 backfill 策略。

### Modified Capabilities

- `ucg-oss-presign`：视频 register/服务端直传成功后 MUST 确保物理首帧 thumb 存在。
- `ucg-app-http-api`：帖子/动态媒体列表中视频 `thumbnailUrl` MUST 为物理 thumb CDN URL（非 `x-oss-process`）。
- `ucg-chat-ws`：聊天消息（含 WS 实时推送）中视频 MUST 提供 `mediaThumbnailUrl` 物理 thumb CDN URL。
- `ucg-admin-post-moderation`：管理端帖子列表视频 `thumbnailUrl` MUST 为物理 thumb CDN URL。

## Impact

- `internal/shared/mediacdn/`（扩展 video thumb helper）
- `internal/services/ucg/oss_thumb.go`、`oss_cdn.go`、`media_register.go`、`oss_upload.go`、`oss_delete.go`
- `internal/services/ucg/post.go`、`chat_store.go`、`notification.go`、`post_sample_internal.go`
- 仅 `ucg-service` 进程；无 Redis 变更、无 DB 表结构变更、无新 background ticker
- Flutter 客户端（`d:\work\flutter_ai_talk`）：需验证 Web/Native 对 `mediaThumbnailUrl`（聊天视频）与 `thumbnailUrl`（帖子视频）的展示兼容性；见 tasks.md 末项
