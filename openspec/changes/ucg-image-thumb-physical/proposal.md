## Why

UCG 图片缩略图当前通过原图 CDN URL 拼接 `?x-oss-process=image/...` 生成。CDN 层常按 path 缓存，导致原图与缩略图响应窜包，列表/头像等场景出现错图。需改为上传时生成独立 OSS 物理对象，读路径使用不含 query 的 CDN URL，从根上隔离缓存键。

## What Changes

- 新增全局缩略图后缀常量 `_thumb`（`internal/shared/mediacdn`），提供 `ThumbObjectKey` 等 helper；业务代码禁止散落 `"_thumb"` 字面量。
- 图片缩略图 objectKey 规则：`{stem}_thumb.{ext}`（扩展名与原图一致），例如 `social/2026/06/abc.jpg` → `social/2026/06/abc_thumb.jpg`。
- 上传登记（`RegisterMedia` miss）、服务端直传（`putOSSObject`）后 MUST 同步生成并 PUT 缩略图对象；删除原图时 MUST 一并删除对应缩略图对象。
- `BuildImageThumbnailURL` 改为返回物理缩略图 CDN URL，图片路径不再使用 `x-oss-process`（视频首帧 `BuildVideoSnapshotURL` 本变更不改动）。
- 缩略图生成：服务端 OSS `GetObject` + `x-oss-process`（resize w_200、auto-orient、quality 90、按原格式输出）取字节后 `PutObject` 至 thumb key；register、直传与 backfill 共用 `EnsureImageThumb`。
- 新增一次性 CLI `cmd/ucg-image-thumb-backfill` 与 runbook，对历史图片 objectKey 幂等补全缩略图（方案 A：数据量不大，先 backfill 再切换读路径）。
- 历史 `ucg_notification.post_thumb_url` 中已固化的 `x-oss-process` URL 本变更不批量改写（影响面小）；可选后续读时重算。

## Capabilities

### New Capabilities

- `ucg-image-thumb`：缩略图命名、物理对象生成、`BuildImageThumbnailURL` 语义、删除成对、backfill CLI 与发布顺序。

### Modified Capabilities

- `ucg-oss-presign`：图片 register/直传成功后 MUST 确保物理缩略图对象存在。
- `ucg-admin-post-moderation`：管理端帖子列表图片 `thumbnailUrl` MUST 为物理缩略图 CDN URL（不含 `x-oss-process`）。

## Impact

- `internal/shared/mediacdn/`（新）、`internal/services/ucg/oss_cdn.go`、`oss_thumb.go`（新）、`media_register.go`、`oss_upload.go`、`oss_delete.go`
- `cmd/ucg-image-thumb-backfill/main.go`（新）、`docs/runbooks/ucg-image-thumb-backfill.md`（新）
- API 字段名不变（`thumbnailUrl`、`avatarThumbnailUrl`、`mediaThumbnailUrl`）；图片 URL 形态变化（无 query），对正确实现客户端透明
- 部署依赖：test/prod 先跑 backfill，再发布读路径切换；需 OSS 凭证与 `UCG_DB_LINK`
- 无新 Redis、无 DB 表结构变更
