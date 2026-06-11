## 1. 后端视频 thumbnail

- [x] 1.1 `post.go` `loadPostMedia`：`mediaKind==2` 时设置 `ThumbnailUrl = BuildVideoSnapshotURL(key)`
- [x] 1.2 确认 `mapAdminPostItem` 已透传 `thumbnailUrl` / `mediaKind`（无额外改动则勾选）

## 2. ucg-admin 媒体列 UI

- [x] 2.1 `renderPostsTable`：遍历 `row.media` 全量渲染缩略图（区分图片/视频）
- [x] 2.2 新增 `#mediaPreviewModal`：图片原图 / 视频 `<video controls>` 两种模式
- [x] 2.3 缩略图点击打开 modal；关闭时暂停视频并清理 src

## 3. 样式

- [x] 3.1 媒体格 flex 布局、视频 ▶ 角标、modal 内大图/视频尺寸（`ucg-admin.html` 或 `admin-pages.css`）
- [x] 3.2 列表缩略图固定 48×48（`#panelPosts` 作用域 + `object-fit: cover`），表格 `table-wrap` 防撑破布局
- [x] 3.3 媒体列 3×3 九宫格（最多 9 张），行间浅灰分割线
- [x] 3.4 视频预览：modal 可见后再 load；`<source>` + 错误提示，修复黑屏

## 4. 校验

- [x] 4.1 `openspec validate enhance-ucg-admin-post-media --strict`
- [x] 4.2 `go build ./...`
- [x] 4.3 手动：含视频待审帖缩略图可见；多图全展示；点击放大/播放正常（部署 ucg-service + git pull 静态页后验收）
