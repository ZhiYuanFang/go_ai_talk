## ADDED Requirements

### Requirement: UCG admin SHALL authenticate post moderation APIs with X-Admin-Password

`ucg-service` Admin 动态审查接口 MUST 与现有 UCG Admin 共用认证：请求 Header `X-Admin-Password` MUST 等于配置项 `ucg.admin.password`；校验失败 MUST 返回未授权，且 SHALL NOT 返回帖子数据或执行驳回。

#### Scenario: 口令正确访问列表

- **WHEN** 管理员携带正确 `X-Admin-Password` 请求 `GET /ucg/admin/api/posts/list`
- **THEN** 系统 SHALL 返回分页帖子列表

#### Scenario: 口令错误拒绝

- **WHEN** 管理员携带错误或未携带 `X-Admin-Password` 请求任一动态审查 Admin API
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改 `ucg_post`

### Requirement: Admin SHALL list all posts with optional status filter

`GET /ucg/admin/api/posts/list` MUST 支持查询参数 `page`（从 1 开始）、`pageSize`（默认 20，最大 100）、可选 `status`（0/1/2/3）。省略 `status` 时 SHALL 返回全部状态的帖子。列表 MUST 按 `updated_at` 降序排序。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；每项 SHALL 至少包含 `id`、`authorWxId`、`content`、`status`、`rejectReason`、`createdAt`、`updatedAt`、`publishedAt` 及媒体展示字段（`media` 含 CDN URL）。

#### Scenario: 按状态筛选待审帖

- **WHEN** 管理员请求 `status=1`
- **THEN** 响应 `list` 中每条 `status` SHALL 为 1（pending_audit）

#### Scenario: 分页默认值

- **WHEN** 管理员未传 `page` 与 `pageSize`
- **THEN** 系统 SHALL 使用 `page=1`、`pageSize=20` 并返回对应分页元数据

### Requirement: Admin SHALL batch reject posts with shared reject semantics

`POST /ucg/admin/api/posts/reject` MUST 接受 JSON body：`postIds`（非空数组，最多 100 个 id）、可选 `reason`（空时 SHALL 使用默认文案「违规已下架」）。对每条目标帖子，若当前 `status` 已为 3（rejected）SHALL 计入 `skipped` 且不更新行；若 `status` 为 0、1 或 2，系统 MUST 将 `status` 置为 3、写入 `reject_reason`、更新 `updated_at` 为当前 unix 秒，并计入 `rejected`。DB 错误计入 `failed`。响应 MUST 包含 `rejected`、`skipped`、`failed` 三个 id 数组。本操作 SHALL NOT 向作者发送通知或站内信。

#### Scenario: 批量驳回已发布帖

- **WHEN** 管理员提交 `postIds` 含 `status=2` 的帖子且口令正确
- **THEN** 对应行 `status` SHALL 变为 3 且 `reject_reason` SHALL 非空；该帖 SHALL NOT 出现在推荐或关注 Feed（`status=published` 查询）

#### Scenario: 已驳回帖幂等跳过

- **WHEN** 管理员对 `status=3` 的帖子再次提交驳回
- **THEN** 该行 SHALL NOT 被修改且该 id SHALL 出现在 `skipped`

#### Scenario: 作者可见驳回原因无推送

- **WHEN** 帖子被管理端驳回
- **THEN** 作者请求「我的动态」SHALL 可见该帖 `status=3` 与 `reject_reason`；系统 SHALL NOT 因本次驳回创建通知或 WS 推送

### Requirement: ucg-admin.html SHALL provide post moderation tab with batch reject UI

静态页 `resource/public/ucg-admin.html` MUST 在现有 UCG Admin 登录态下提供「动态审查」模块（可与 AI 配置以 Tab 切换）。模块 SHALL 调用列表 API 展示表格，对 `status≠3` 的行提供 checkbox；SHALL 提供「全选本页可驳回项」与「批量驳回」按钮；批量驳回前 MUST 经用户确认。`status=3` 的行 checkbox SHALL 禁用或不可选。操作成功后 SHALL 刷新当前列表。

#### Scenario: 全选本页仅选可驳回项

- **WHEN** 管理员点击全选且当前页含已驳回与可驳回帖
- **THEN** 仅 `status≠3` 的行 SHALL 被勾选

#### Scenario: 批量驳回后刷新

- **WHEN** 管理员确认批量驳回且 API 返回成功
- **THEN** 页面 SHALL 刷新列表且已驳回帖显示为不可选状态

### Requirement: device admin entry SHALL link to UCG management

`resource/public/admin.html` 中指向 `/device/admin/ucg-admin.html` 的入口链接文案 SHALL 为「UCG 管理」（或等价中文），以涵盖 AI 配置与动态审查。

#### Scenario: 设备管理页入口文案

- **WHEN** 管理员打开设备管理页
- **THEN** UCG 入口链接可见文案 SHALL 为「UCG 管理」
