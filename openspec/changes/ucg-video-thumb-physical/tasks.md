## 1. 共享命名与 URL

- [x] 1.1 扩展 `internal/shared/mediacdn/thumb.go`：新增 `VideoThumbExt`、`VideoThumbObjectKey`（`{stem}.mp4` → `{stem}_thumb.jpg`；已是 thumb key 时幂等返回）
- [x] 1.2 新增 `BuildVideoThumbnailURL`：`BuildCdnURL(VideoThumbObjectKey(key))`；移除或废弃 `BuildVideoSnapshotURL` 对外用法

## 2. 视频缩略图生成与 OSS 集成

- [x] 2.1 在 `oss_thumb.go` 新增 `EnsureVideoThumb`：HEAD 幂等、`GetObject`+`video/snapshot,t_0`、`PutObject` thumb jpg、`Content-Type: image/jpeg`
- [x] 2.2 `RegisterMedia` 在 `mediaKind==2` 成功路径调用 `EnsureVideoThumb`；失败时 register 返回错误
- [x] 2.3 `putOSSObject` 在 `mediaKind==2` 上传成功后调用 `EnsureVideoThumb`
- [x] 2.4 扩展删除逻辑：删除 mp4 时 `DeleteObject(VideoThumbObjectKey(key))`（404 忽略）；图片仍用 `ThumbObjectKey`

## 3. 读路径替换

- [x] 3.1 `post.go` `loadPostMedia`：视频 `thumbnailUrl` 改用 `BuildVideoThumbnailURL`
- [x] 3.2 `chat_store.go` `enrichChatMessageMedia`：视频分支设置 `mediaThumbnailUrl = BuildVideoThumbnailURL(videoKey)`
- [x] 3.3 `notification.go` `resolvePostCoverSnapshot`：视频封面改用 `BuildVideoThumbnailURL`
- [x] 3.4 `post_sample_internal.go` `postSampleCoverCdnURL`：视频封面改用 `BuildVideoThumbnailURL`

## 4. 部署与验证（ucg-service）

- [ ] 4.1 test 环境：上传新视频 → 验证 OSS 成对 `*.mp4` + `*_thumb.jpg` → 验证 Feed/帖子/聊天/Admin 列表 thumb URL 无 `x-oss-process`
- [ ] 4.2 prod 环境重复 4.1；确认无 backfill CLI 依赖；历史无 thumb 视频由用户重传

## 5. Flutter 客户端（sibling repo）

- [x] 5.1 检查 `d:\work\flutter_ai_talk`：Web/Native 聊天视频气泡是否使用 `mediaThumbnailUrl`；帖子/动态列表是否使用 `thumbnailUrl` 展示视频封面
- [x] 5.2 若客户端仍依赖 `x-oss-process` 截帧或忽略视频 `mediaThumbnailUrl`，在 `flutter_ai_talk` 直接修改并验证物理 jpg thumb 加载
