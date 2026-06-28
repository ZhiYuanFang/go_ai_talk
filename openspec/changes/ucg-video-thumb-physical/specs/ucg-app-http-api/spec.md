## ADDED Requirements

### Requirement: Post and feed media DTOs SHALL expose physical video thumbnailUrl

App HTTP API 返回的帖子/动态媒体项（含推荐 Feed、关注 Feed、帖子详情、我的动态等经 `loadPostMedia` 或等价路径组装的 `media` 数组）中，当 `mediaKind=2`（视频）且 `objectKey` 非空时，MUST 填充 `thumbnailUrl` 为 `BuildVideoThumbnailURL(objectKey)` 物理首帧 jpg CDN URL。该 URL MUST NOT 含 `x-oss-process` query。`cdnUrl` MUST 仍为可播放的 mp4 CDN URL。

本要求 MUST NOT 改变既有 v1/v2 响应 JSON 字段名或结构，仅变更 `thumbnailUrl` 的 URL 形态。

#### Scenario: Feed 中视频帖含物理 thumbnailUrl

- **WHEN** 客户端 `GET /ucg/app/api/feed/recommend` 且列表含已发布视频帖（OSS 已存在 `_thumb.jpg`）
- **THEN** 对应 `media` 项 `thumbnailUrl` SHALL 以 `_thumb.jpg` 结尾且 SHALL NOT 含 `x-oss-process`

#### Scenario: 视频 media 仍保留可播放 cdnUrl

- **WHEN** 帖子含视频媒体
- **THEN** 该项 `cdnUrl` SHALL 指向 mp4 原片且 `thumbnailUrl` SHALL 与 `cdnUrl` path 不同
