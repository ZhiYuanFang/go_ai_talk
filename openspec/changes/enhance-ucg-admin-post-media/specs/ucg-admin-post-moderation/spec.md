## MODIFIED Requirements

### Requirement: Admin SHALL list all posts with optional status filter

`GET /ucg/admin/api/posts/list` MUST 支持查询参数 `page`（从 1 开始）、`pageSize`（默认 20，最大 100）、可选 `status`（0/1/2/3）。省略 `status` 时 SHALL 返回全部状态的帖子。列表 MUST 按 `updated_at` 降序排序。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；每项 SHALL 至少包含 `id`、`authorWxId`、`content`、`status`、`rejectReason`、`createdAt`、`updatedAt`、`publishedAt` 及媒体展示字段。`media` 数组 MUST 包含该帖 **全部** 媒体项（按 `sortOrder`）；每项 SHALL 含 `cdnUrl`、`mediaKind`（1=图片，2=视频），图片 SHALL 含 OSS 缩略图 `thumbnailUrl`，视频 SHALL 含 OSS 首帧截帧 `thumbnailUrl`（`BuildVideoSnapshotURL`），SHALL NOT 仅返回无 thumbnail 的 mp4 `cdnUrl` 供列表 `<img>` 直接使用。

#### Scenario: 按状态筛选待审帖

- **WHEN** 管理员请求 `status=1`
- **THEN** 响应 `list` 中每条 `status` SHALL 为 1（pending_audit）

#### Scenario: 分页默认值

- **WHEN** 管理员未传 `page` 与 `pageSize`
- **THEN** 系统 SHALL 使用 `page=1`、`pageSize=20` 并返回对应分页元数据

#### Scenario: 视频帖返回首帧 thumbnail

- **WHEN** 列表项首条媒体 `mediaKind=2` 且 `objectKey` 有效
- **THEN** 该项 `thumbnailUrl` SHALL 非空且 SHALL 可用于 `<img>` 展示；`cdnUrl` SHALL 为可播放的视频 URL

#### Scenario: 多图帖返回全量 media

- **WHEN** 帖子关联多条 `ucg_post_media`
- **THEN** 列表项 `media` 数组长度 SHALL 等于关联条数且顺序与 `sortOrder` 一致

### Requirement: ucg-admin.html SHALL provide post moderation tab with batch reject UI

静态页 `resource/public/ucg-admin.html` MUST 在现有 UCG Admin 登录态下提供「动态审查」模块（可与 AI 配置以 Tab 切换）。模块 SHALL 调用列表 API 展示表格，对 `status≠3` 的行提供 checkbox；SHALL 提供「全选本页可驳回项」与「批量驳回」按钮；批量驳回前 MUST 经用户确认。`status=3` 的行 checkbox SHALL 禁用或不可选。操作成功后 SHALL 刷新当前列表。表格「媒体」列 SHALL 展示每条动态 **全量** 媒体缩略图（非仅第一条）；图片 SHALL 支持点击后在 modal 中查看原图（`cdnUrl`）；视频 SHALL 展示首帧缩略图并支持点击后在 modal 内播放（`<video controls>` + `cdnUrl`）。

#### Scenario: 全选本页仅选可驳回项

- **WHEN** 管理员点击全选且当前页含已驳回与可驳回帖
- **THEN** 仅 `status≠3` 的行 SHALL 被勾选

#### Scenario: 批量驳回后刷新

- **WHEN** 管理员确认批量驳回且 API 返回成功
- **THEN** 页面 SHALL 刷新列表且已驳回帖显示为不可选状态

#### Scenario: 多图动态展示全部缩略图

- **WHEN** 列表项 `media` 含 3 张图片
- **THEN** 媒体列 SHALL 展示 3 个可辨认缩略图而非仅 1 个

#### Scenario: 图片点击放大

- **WHEN** 管理员点击某图片缩略图
- **THEN** 页面 SHALL 在 modal 中展示该图全分辨率 `cdnUrl`

#### Scenario: 视频点击播放

- **WHEN** 管理员点击某视频缩略图
- **THEN** 页面 SHALL 在 modal 内提供可控制的视频播放器加载该条 `cdnUrl`
