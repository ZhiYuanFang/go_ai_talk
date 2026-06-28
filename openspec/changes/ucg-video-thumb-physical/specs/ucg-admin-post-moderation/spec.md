## MODIFIED Requirements

### Requirement: Admin SHALL list all posts with optional status filter

`GET /ucg/admin/api/posts/list` MUST 支持查询参数 `page`（从 1 开始）、`pageSize`（默认 20，最大 100）、可选 `status`（0/1/2/3）。省略 `status` 时 SHALL 返回全部状态的帖子。列表 MUST 按 `updated_at` 降序排序。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；每项 SHALL 至少包含 `id`、`authorWxId`、`content`、`status`、`rejectReason`、`createdAt`、`updatedAt`、`publishedAt` 及媒体展示字段。`media` 数组 MUST 包含该帖 **全部** 媒体项（按 `sortOrder`）；每项 SHALL 含 `cdnUrl`、`mediaKind`（1=图片，2=视频），图片 SHALL 含物理缩略图 `thumbnailUrl`（`BuildImageThumbnailURL`，path 为 `{stem}_thumb.{ext}`，MUST NOT 含 `x-oss-process`），视频 SHALL 含物理首帧缩略图 `thumbnailUrl`（`BuildVideoThumbnailURL`，path 为 `{stem}_thumb.jpg`，MUST NOT 含 `x-oss-process`），SHALL NOT 仅返回无 thumbnail 的 mp4 `cdnUrl` 供列表 `<img>` 直接使用。

#### Scenario: 按状态筛选待审帖

- **WHEN** 管理员请求 `status=1`
- **THEN** 响应 `list` 中每条 `status` SHALL 为 1（pending_audit）

#### Scenario: 分页默认值

- **WHEN** 管理员未传 `page` 与 `pageSize`
- **THEN** 系统 SHALL 使用 `page=1`、`pageSize=20` 并返回对应分页元数据

#### Scenario: 视频帖返回首帧 thumbnail

- **WHEN** 列表项首条媒体 `mediaKind=2` 且 `objectKey` 有效且 OSS 存在对应 `_thumb.jpg`
- **THEN** 该项 `thumbnailUrl` SHALL 非空且 SHALL 可用于 `<img>` 展示；`cdnUrl` SHALL 为可播放的视频 URL；`thumbnailUrl` SHALL NOT 含 `x-oss-process`

#### Scenario: 多图帖返回全量 media

- **WHEN** 帖子关联多条 `ucg_post_media`
- **THEN** 列表项 `media` 数组长度 SHALL 等于关联条数且顺序与 `sortOrder` 一致

#### Scenario: 视频 thumbnail 为物理 jpg path

- **WHEN** 视频 objectKey 为 `social/2026/06/a.mp4` 且 thumb 已 materialize
- **THEN** 该项 `thumbnailUrl` SHALL 指向 `.../a_thumb.jpg` 且 SHALL NOT 包含 `x-oss-process`
