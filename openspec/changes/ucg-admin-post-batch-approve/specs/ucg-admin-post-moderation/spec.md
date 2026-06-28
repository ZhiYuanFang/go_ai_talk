## ADDED Requirements

### Requirement: Admin SHALL batch approve posts for human publish

`POST /ucg/admin/api/posts/approve` MUST 接受 JSON body：`postIds`（非空数组，最多 100 个 id）。对每条目标帖子：

- 若 `status=2`（published）SHALL 计入 `skipped` 且不更新行。
- 若 `status=1`（pending_audit）、`4`（apply_failed）或 `5`（moderation_failed），系统 MUST 在不调用 Green 的前提下将帖子发布：`status` 置为 `2`，写入 `published_at` 与 `updated_at`，清空面向作者的 `reject_reason`（及 apply 失败相关字段）；`status=5` 时 MUST 写入 `moderation_verdict=pass` 与 `moderation_at`；`status=1` 且 `moderation_verdict=0` 时 MUST 写入 `moderation_verdict=pass`。成功后 MUST 调用与 MQ publish 等价的 Feed/Redis 同步（`syncPublishedPostRedis`）。计入 `approved`。
- 若 `status=0`（draft）或 `3`（rejected），SHALL 计入 `failed` 且不更新行。
- DB/CAS 错误计入 `failed`。

响应 MUST 包含 `approved`、`skipped`、`failed` 三个 id 数组。本操作 SHALL NOT 向作者发送通知或站内信。

#### Scenario: 批量通过待审帖

- **WHEN** 管理员提交含 `status=1` 的 `postIds` 且口令正确
- **THEN** 对应行 `status` SHALL 变为 `2` 且 `published_at` SHALL 非空；该帖 SHALL 可出现在推荐/关注 Feed

#### Scenario: 批量通过机审失败帖

- **WHEN** 管理员提交含 `status=5` 的 postId
- **THEN** 行 SHALL 变为 `published` 且 `moderation_verdict` SHALL 为 pass；`reject_reason` SHALL 清空

#### Scenario: 已发布帖幂等跳过

- **WHEN** 管理员对 `status=2` 的帖子提交 approve
- **THEN** 行 SHALL NOT 被修改且 id SHALL 出现在 `skipped`

#### Scenario: 已驳回帖不可批准

- **WHEN** 管理员对 `status=3` 的帖子提交 approve
- **THEN** 行 SHALL NOT 被修改且 id SHALL 出现在 `failed`

## MODIFIED Requirements

### Requirement: Admin SHALL list all posts with optional status filter

`GET /ucg/admin/api/posts/list` MUST 支持查询参数 `page`（从 1 开始）、`pageSize`（默认 20，最大 100）、可选 `status`（**0/1/2/3/4/5**）。省略 `status` 时 SHALL 返回全部状态的帖子。列表 MUST 按 `updated_at` 降序排序。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；每项 SHALL 至少包含 `id`、`authorWxId`、`content`、`status`、`rejectReason`、`createdAt`、`updatedAt`、`publishedAt` 及媒体展示字段。`media` 数组 MUST 包含该帖 **全部** 媒体项（按 `sortOrder`）；每项 SHALL 含 `cdnUrl`、`mediaKind`（1=图片，2=视频），图片 SHALL 含物理缩略图 `thumbnailUrl`，视频 SHALL 含物理首帧缩略图 `thumbnailUrl`（`{stem}_thumb.jpg`），SHALL NOT 仅返回无 thumbnail 的 mp4 `cdnUrl` 供列表 `<img>` 直接使用。

#### Scenario: 按状态筛选待审帖

- **WHEN** 管理员请求 `status=1`
- **THEN** 响应 `list` 中每条 `status` SHALL 为 1（pending_audit）

#### Scenario: 按状态筛选机审失败帖

- **WHEN** 管理员请求 `status=5`
- **THEN** 响应 `list` 中每条 `status` SHALL 为 5（moderation_failed）

#### Scenario: 分页默认值

- **WHEN** 管理员未传 `page` 与 `pageSize`
- **THEN** 系统 SHALL 使用 `page=1`、`pageSize=20` 并返回对应分页元数据

#### Scenario: 视频帖返回首帧 thumbnail

- **WHEN** 列表项含 `mediaKind=2` 且 thumb 已 materialize
- **THEN** 该项 `thumbnailUrl` SHALL 非空且 MUST NOT 含 `x-oss-process`

#### Scenario: 多图帖返回全量 media

- **WHEN** 帖子关联多条 `ucg_post_media`
- **THEN** 列表项 `media` 数组长度 SHALL 等于关联条数且顺序与 `sortOrder` 一致

### Requirement: Admin SHALL batch reject posts with shared reject semantics

`POST /ucg/admin/api/posts/reject` MUST 接受 JSON body：`postIds`（非空数组，最多 100 个 id）、**`reason`（必填，trim 后非空）**。`reason` 为空或仅空白时 MUST 返回参数错误且 SHALL NOT 修改任何帖子。对每条目标帖子，若当前 `status` 已为 3（rejected）SHALL 计入 `skipped` 且不更新行；若 `status` 为 0、1、2、4 或 5，系统 MUST 将 `status` 置为 3、写入 `reject_reason`（管理员提供的 reason）、更新 `updated_at` 为当前 unix 秒，并计入 `rejected`。若原帖为 `status=2`（published），MUST 从推荐/Feed 索引移除。DB 错误计入 `failed`。响应 MUST 包含 `rejected`、`skipped`、`failed` 三个 id 数组。本操作 SHALL NOT 向作者发送通知或站内信。

#### Scenario: 批量驳回已发布帖

- **WHEN** 管理员提交 `postIds` 含 `status=2` 的帖子、非空 `reason` 且口令正确
- **THEN** 对应行 `status` SHALL 变为 3 且 `reject_reason` SHALL 等于提交的 reason；该帖 SHALL NOT 出现在推荐或关注 Feed

#### Scenario: 驳回缺少 reason 拒绝

- **WHEN** 管理员提交 reject 且 `reason` 为空或仅空白
- **THEN** API SHALL 返回参数错误且 MUST NOT 修改 `ucg_post`

#### Scenario: 已驳回帖幂等跳过

- **WHEN** 管理员对 `status=3` 的帖子再次提交驳回
- **THEN** 该行 SHALL NOT 被修改且该 id SHALL 出现在 `skipped`

#### Scenario: 作者可见驳回原因无推送

- **WHEN** 帖子被管理端驳回且 reason 为「含不当用语」
- **THEN** 作者请求「我的动态」SHALL 可见该帖 `status=3` 与 `rejectReason=含不当用语`；系统 SHALL NOT 因本次驳回创建通知或 WS 推送

### Requirement: ucg-admin.html SHALL provide post moderation tab with batch reject UI

静态页 `resource/public/ucg-admin.html` MUST 在现有 UCG Admin 登录态下提供「动态审查」模块（可与 AI 配置以 Tab 切换）。模块 SHALL 调用列表 API 展示表格，对 `status≠3` 的行提供 checkbox；SHALL 提供「全选本页可驳回项」、**「批量通过」**与「批量驳回」按钮。批量驳回前 MUST 弹出理由输入且 MUST NOT 在理由为空时提交；批量通过前 MUST 经用户确认。**批量驳回 MUST 在请求 body 中携带非空 `reason`。** `status=3` 的行 checkbox SHALL 禁用或不可选。操作成功后 SHALL 刷新当前列表。表格 SHALL 含 **驳回原因** 列（展示 `rejectReason`）。状态筛选 SHALL 含 0–5（含「发布失败(4)」「机审失败(5)」）。表格「媒体」列 SHALL 展示每条动态 **全量** 媒体缩略图；图片 SHALL 支持 modal 原图；视频 SHALL 展示首帧缩略图并支持 modal 播放。

动态审查 Tab 内 **工具栏行**（状态筛选、刷新、批量通过、批量驳回、已选提示）SHALL 使用 flex 布局且 **`align-items: center`**，使 label、select、button、hint 文本在同一行内纵向居中对齐；SHALL NOT 因全局 `.row { align-items: flex-start }` 导致工具栏元素顶对齐错位。样式 SHOULD  scoped 至 `#panelPosts`（或等价 class），MUST NOT 改变其它 Tab 的 `.row` 布局。

同一页面 MUST 提供与「动态审查」并列的 **「资料机审失败」** Tab。

#### Scenario: 动态审查批量驳回须填理由

- **WHEN** 管理员勾选帖子并点击批量驳回且在 prompt 中输入非空理由
- **THEN** 系统 SHALL 调用 reject API 且 body 含该 reason，并刷新列表

#### Scenario: 动态审查批量通过

- **WHEN** 管理员勾选含待审/机审失败/发布失败帖并确认批量通过
- **THEN** 系统 SHALL 调用 approve API 且刷新列表

#### Scenario: 驳回理由为空不提交

- **WHEN** 管理员在驳回 prompt 留空或取消
- **THEN** 页面 SHALL NOT 调用 reject API

#### Scenario: 管理页含资料机审 Tab

- **WHEN** 管理员打开 ucg-admin.html 并已通过口令登录
- **THEN** 页面 SHALL 展示「资料机审失败」Tab 入口

#### Scenario: 动态审查工具栏纵向居中

- **WHEN** 管理员打开「动态审查」Tab 且工具栏含状态 select 与批量操作按钮
- **THEN** 工具栏行内各控件 SHALL 纵向居中对齐（视觉同一水平中线），且 AI 配置等其它 Tab 的 `.row` 布局 SHALL 保持不变
