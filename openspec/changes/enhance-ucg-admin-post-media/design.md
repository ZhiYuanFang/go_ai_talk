## Context

- 管理端列表：`GET /ucg/admin/api/posts/list` → `ListPostsForAdmin` → `postToDTO` → `loadPostMedia`
- 图片 `mediaKind=1`：`thumbnailUrl` = OSS 缩略图 process；`cdnUrl` = 原图
- 视频 `mediaKind=2`：仅有 `cdnUrl`（mp4）；`BuildVideoSnapshotURL` 已在 `notification.resolvePostCoverSnapshot` 使用，帖子 DTO 未复用
- 前端 `ucg-admin.html`：`renderPostsTable` 只取 `media[0]`，统一 `<img class="thumb">`

## Goals / Non-Goals

**Goals**

- 视频审查列表缩略图正常显示（首帧 snapshot）
- 每条动态展示 **全部** 媒体缩略图
- 图片点击全屏/ modal 查看原图；视频点击 modal 播放
- 改动最小，不破坏批量驳回等现有流程

**Non-Goals**

- App 端 Feed/详情 UI 改造（thumbnail 后端补齐后 App 可选后续优化）
- 视频转码、新 OSS 能力
- 审查页 inline 播放（仅 modal）

## Decisions

### 1. 后端：在 `loadPostMedia` 补视频 `thumbnailUrl`

```text
mediaKind == 2 → item.ThumbnailUrl = BuildVideoSnapshotURL(key)
```

与 `resolvePostCoverSnapshot` 对齐；所有 `PostDTO` 消费者（含 App API）一并获得视频封面 URL。

### 2. 前端：媒体列多格 + 统一 preview modal

- 每行 `media.forEach` 渲染 `.media-thumb-wrap`（flex 横向，可 wrap）
- `mediaKind === 1`：`<img>` + `data-full-url=cdnUrl`
- `mediaKind === 2`：`<img src=thumbnailUrl>` + ▶ 角标 + `data-video-url=cdnUrl`
- 单页一个 `#mediaPreviewModal`：图片模式 `<img>`；视频模式 `<video controls>`
- 点击缩略图打开；ESC / 遮罩 / 关闭按钮关闭；打开视频时 `video.play()`，关闭时 `pause()` + `src` 清空

### 3. 样式

- 优先在 `ucg-admin.html` 内 `<style>` 或复用 `admin-pages.css` 增加 `.media-grid`、`.media-play-badge`
- 缩略图尺寸略大于现 48px（如 56px）以兼顾多图

## Risks / Trade-offs

- **[Risk] OSS 视频截帧未开启** → 与通知封面同依赖；snapshot 失败时仍可用 `cdnUrl` 点 modal 播放
- **[Trade-off] 多图行高增加** → flex wrap + max-width 限制媒体列宽度

## Migration Plan

- 部署 ucg-service 后新请求即带 video thumbnail；静态页 pull 后生效
- 无需 DB 迁移

## Open Questions

- 无
