## Why

运维「动态审查」列表中，含视频的动态缩略图加载失败：后端 `loadPostMedia` 对视频（`mediaKind=2`）未填充 `thumbnailUrl`，前端用 `<img src=cdnUrl>` 指向 mp4，浏览器无法解码。同时媒体列仅展示 **第一条** 媒体、无点击放大/视频播放，审查员无法看清全量图片与视频内容。

## What Changes

- **ucg-service**：`loadPostMedia` 对视频填充 `thumbnailUrl`（复用已有 `BuildVideoSnapshotURL`，与通知封面一致）
- **ucg-admin.html**：媒体列展示帖子 **全量** 媒体缩略图；图片点击 modal 放大；视频点击 modal 内 `<video controls>` 播放
- 复用现有 `pangbao-theme.css` 的 `.modal` 样式；不新增 App 接口

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-admin-post-moderation`：列表媒体展示与 PostDTO 视频 thumbnail 语义；审查页全量媒体 + 预览交互

## Impact

- `internal/services/ucg/post.go` — `loadPostMedia` 视频 thumbnail
- `resource/public/ucg-admin.html` — 媒体列渲染与预览 modal
- 可选：`resource/public/pangbao-theme.css` 或 `admin-pages.css` 少量样式（媒体格、播放角标）
- 部署：**ucg-service** 重建；HTML/CSS `git pull` 即可（不经 gateway 镜像）
