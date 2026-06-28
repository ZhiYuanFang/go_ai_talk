## ADDED Requirements

### Requirement: Video thumb objectKey SHALL be stem_thumb.jpg regardless of video extension

平台 MUST 在 `internal/shared/mediacdn` 提供 `VideoThumbObjectKey(videoObjectKey)` helper（及 `VideoThumbExt = "jpg"` 常量）。对视频原 objectKey `{path}/{stem}.{videoExt}`，缩略图 objectKey MUST 为 `{path}/{stem}_thumb.jpg`，且 MUST NOT 使用与视频相同的扩展名（如 `xyz.mp4` → `xyz_thumb.jpg`，NOT `xyz_thumb.mp4`）。

业务代码 MUST 经该 helper 派生视频 thumb key，MUST NOT 散落 `_thumb.jpg` 字面量拼接。

对已是视频 thumb 的 key（stem 以 `_thumb` 结尾且 ext 为 jpg），`VideoThumbObjectKey` MUST 原样返回。

#### Scenario: MP4 原视频派生 thumb key

- **WHEN** 原视频 objectKey 为 `social/2026/06/xyz.mp4`
- **THEN** `VideoThumbObjectKey` SHALL 返回 `social/2026/06/xyz_thumb.jpg`

#### Scenario: thumb key 幂等

- **WHEN** 输入已为 `social/2026/06/xyz_thumb.jpg`
- **THEN** `VideoThumbObjectKey` SHALL 原样返回该 key

### Requirement: EnsureVideoThumb SHALL create idempotent physical first-frame objects via OSS snapshot

ucg-service MUST 提供 `EnsureVideoThumb(ctx, videoObjectKey)`：对视频原 objectKey，若 `{stem}_thumb.jpg` 不存在，MUST 经 OSS `GetObject` 携带 `video/snapshot,t_0` 获取首帧字节后 `PutObject` 至 `VideoThumbObjectKey(videoObjectKey)`，`Content-Type` MUST 为 `image/jpeg`；若 thumb 已存在 MUST 跳过（幂等）。原视频不存在时 MUST 返回明确错误。

OSS `x-oss-process` MUST 仅用于服务端生成阶段，MUST NOT 出现在返回客户端的 CDN URL 中。

#### Scenario: 首次生成视频 thumb

- **WHEN** 原视频 `social/.../a.mp4` 存在于 OSS 且 `social/.../a_thumb.jpg` 不存在
- **THEN** `EnsureVideoThumb` SHALL 上传 `a_thumb.jpg` 且后续 `HEAD` 可命中

#### Scenario: 重复调用幂等

- **WHEN** `a_thumb.jpg` 已存在
- **THEN** `EnsureVideoThumb` SHALL 成功返回且 MUST NOT 覆盖已有对象

### Requirement: BuildVideoThumbnailURL SHALL return physical thumb CDN URL without x-oss-process

`BuildVideoThumbnailURL(videoObjectKey)` MUST 返回 `BuildCdnURL(VideoThumbObjectKey(videoObjectKey))`。视频缩略图 CDN URL MUST NOT 包含 `x-oss-process` query 参数。

#### Scenario: 视频列表缩略图 URL 无 query

- **WHEN** 服务为视频 objectKey 拼装 `thumbnailUrl` 或 `mediaThumbnailUrl`
- **THEN** URL path SHALL 以 `_thumb.jpg` 结尾且 SHALL NOT 含 `x-oss-process`

### Requirement: Video media deletion SHALL remove paired _thumb.jpg objects

当 ucg-service 删除用户拥有的视频 OSS 原对象时，MUST 同时尝试删除 `VideoThumbObjectKey(原视频 key)`；thumb 对象不存在时 MUST NOT 导致整次删除失败。

#### Scenario: 删除 mp4 同时删除 thumb jpg

- **WHEN** `DeleteOwnedMedia` 删除 `social/.../a.mp4` 且 blob 允许删 OSS
- **THEN** OSS 上 `social/.../a_thumb.jpg` SHALL 被删除或已不存在

### Requirement: Historical videos without physical thumbs SHALL NOT be backfilled by platform CLI

平台 MUST NOT 提供批量 backfill CLI 为历史视频生成 thumb。无物理 thumb 的历史视频读路径 MAY 返回 404 thumb URL；补救 MUST 由用户通过重编辑或重上传触发写路径 `EnsureVideoThumb`。

#### Scenario: 无 backfill 命令

- **WHEN** 运维检索仓库 `cmd/` 与 runbook
- **THEN** MUST NOT 存在 `ucg-video-thumb-backfill` 或等价批量补全工具
