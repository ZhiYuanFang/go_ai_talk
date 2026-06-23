## ADDED Requirements

### Requirement: Image thumb suffix SHALL be globally defined as _thumb before extension

平台 MUST 在 `internal/shared/mediacdn` 定义缩略图后缀常量（值为 `_thumb`）及 `ThumbObjectKey` helper。任意服务与脚本 MUST 经该 helper 派生缩略图 objectKey，MUST NOT 在业务代码中散落 `_thumb` 字面量。

对原图 objectKey `{path}/{stem}.{ext}`，缩略图 objectKey MUST 为 `{path}/{stem}_thumb.{ext}`，且 `{ext}` MUST 与原图扩展名一致（如 `abc.jpg` → `abc_thumb.jpg`，`abc.png` → `abc_thumb.png`）。

`IsThumbObjectKey` MUST 识别已为缩略图的 key；对这类 key 调用 `ThumbObjectKey` MUST 原样返回，MUST NOT 产生 `_thumb_thumb` 双后缀。

#### Scenario: JPG 原图派生 thumb key

- **WHEN** 原图 objectKey 为 `social/2026/06/xyz.jpg`
- **THEN** `ThumbObjectKey` SHALL 返回 `social/2026/06/xyz_thumb.jpg`

#### Scenario: PNG 扩展名保持一致

- **WHEN** 原图 objectKey 为 `social/2026/06/xyz.png`
- **THEN** 缩略图 objectKey SHALL 为 `social/2026/06/xyz_thumb.png`

### Requirement: EnsureImageThumb SHALL create idempotent physical thumb objects via OSS process

ucg-service MUST 提供 `EnsureImageThumb(ctx, objectKey)`：对图片原图 objectKey，若缩略图对象不存在，MUST 经 OSS `GetObject` 携带图片处理参数（`auto-orient,1/resize,m_lfit,w_200/quality,q_90`，并按扩展名保持输出格式）获取字节后 `PutObject` 至 `ThumbObjectKey(objectKey)`；若缩略图已存在 MUST 跳过（幂等）。原图不存在时 MUST 返回明确错误。

处理参数 MUST NOT 在返回给客户端的 CDN URL 中出现；仅用于服务端生成物理对象。

#### Scenario: 首次生成 thumb

- **WHEN** 原图 `social/.../a.jpg` 存在于 OSS 且 `social/.../a_thumb.jpg` 不存在
- **THEN** `EnsureImageThumb` SHALL 上传 `a_thumb.jpg` 且后续 `HEAD` 可命中

#### Scenario: 重复调用幂等

- **WHEN** `a_thumb.jpg` 已存在
- **THEN** `EnsureImageThumb` SHALL 成功返回且 MUST NOT 覆盖已有对象

### Requirement: BuildImageThumbnailURL SHALL return physical thumb CDN URL without x-oss-process

`BuildImageThumbnailURL(objectKey)` MUST 返回 `BuildCdnURL(ThumbObjectKey(objectKey))`。图片缩略图 CDN URL MUST NOT 包含 `x-oss-process` query 参数。

视频首帧 `BuildVideoSnapshotURL` 本 capability 不修改，MAY 仍使用 `x-oss-process`。

#### Scenario: 图片列表缩略图 URL 无 query

- **WHEN** 服务为图片 objectKey 拼装 `thumbnailUrl`
- **THEN** URL path SHALL 以 `_thumb.{ext}` 结尾且 SHALL NOT 含 `x-oss-process`

### Requirement: Media deletion SHALL remove paired thumb objects

当 ucg-service 删除用户拥有的图片 OSS 原图对象时，MUST 同时尝试删除 `ThumbObjectKey(原图 key)`；thumb 对象不存在时 MUST NOT 导致整次删除失败。

#### Scenario: 删除原图同时删除 thumb

- **WHEN** `DeleteOwnedMedia` 删除 `social/.../a.jpg` 且 blob 允许删 OSS
- **THEN** OSS 上 `social/.../a_thumb.jpg` SHALL 被删除或已不存在

### Requirement: Backfill CLI SHALL populate historical image thumbs before read-path cutover

仓库 MUST 提供 `cmd/ucg-image-thumb-backfill`：从 UCG 数据库收集去重图片 objectKey（至少含 `ucg_media_blob` media_kind=1、`ucg_post_media` 图片、`ucg_profile.avatar_key`、`ucg_chat_message.image_key`），对每条调用与线上一致的 `EnsureImageThumb`（或等价逻辑）。

CLI MUST 支持 `--dry-run`（仅打印将处理 key）、`--limit`、`--concurrency`；执行结束 MUST 输出 ok/skipped/missing/failed 汇总。运维 runbook MUST 规定：各环境 backfill 验证通过后再部署读路径切换。

#### Scenario: dry-run 不写 OSS

- **WHEN** 运维执行 `go run ./cmd/ucg-image-thumb-backfill --dry-run`
- **THEN** 脚本 SHALL 列出将处理的 key 且 MUST NOT 调用 `PutObject`

#### Scenario: 漏网 key 可重跑

- **WHEN** 首次 backfill 部分失败
- **THEN** 再次执行脚本 SHALL 跳过已成功 thumb 并仅处理剩余失败项
