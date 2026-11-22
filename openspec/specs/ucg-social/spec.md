# ucg-social

## Purpose

UCG 社交、OSS 媒体、审核 MQ、Feed/推荐与聊天 WS 的行为规格。 User-generated content service runtime and media pipelines.

<!-- migrated-from: event-logo-oss-cdn, retire-ucg-polish, ucg-admin-post-moderation, ucg-admin-profile-moderation, ucg-aliyun-secrets-env, ucg-app-http-api, ucg-app-profile, ucg-audit-mq, ucg-audit-mq-reliability, ucg-audit-publish-outbox, ucg-chat-mysql-persist, ucg-chat-ws, ucg-data-model, ucg-device-internal-api, ucg-entry-eligibility, ucg-feed-index-lazy-warm, ucg-feed-index-startup-heal, ucg-feed-no-geo-zset-fallback, ucg-following-feed, ucg-force-first-grant, ucg-force-ledger, ucg-green-audit, ucg-image-thumb, ucg-internal-profile-batch, ucg-media-server-upload-only, ucg-oss-caidoukeji-config, ucg-oss-presign, ucg-recommend-feed, ucg-recommend-mq, ucg-sim-chat-internal, ucg-sim-chat-unread-sample, ucg-sim-feed-sample, ucg-video-thumb, ucg-video-transcode, ucg-video-validate -->

## Requirements

### Requirement: 事件 logo SHALL 存储于 OSS event 前缀

device-service 在管理端上传或迁移脚本写入时，MUST 将 `event.logo` 设为 OSS objectKey，前缀 MUST 为 `event/`，MUST NOT 含 scheme 或域名。

#### Scenario: 管理端上传新 logo

- **WHEN** 客户端 multipart 提交合法图片至 event add/update
- **THEN** device-service MUST 经 HTTP 调用 ucg-service 内部上传接口
- **AND** MUST 将返回的 objectKey 写入 `event.logo`

#### Scenario: 迁移脚本回填

- **WHEN** 迁移脚本成功上传历史 logo
- **THEN** `event.logo` MUST 为 `event/{id}/logo.{ext}` 形式

### Requirement: 对外 API logo 字段 SHALL 返回 CDN 绝对 URL

所有返回事件字典列表的 HTTP 接口（含 admin list、history/internal event options、gateway site home 的 logoUrl）MUST 将 logo 序列化为 `https://{cdnHost}/{objectKey}`，MUST NOT 返回 `/ai_talk_images/` path-only。

#### Scenario: history event options 返回 CDN

- **WHEN** 客户端请求 `GET /device/history/api/event/options`
- **THEN** 每项 `logo` 若有值 MUST 以 `https://` 开头且指向 CDN 域名

### Requirement: ucg-service SHALL 提供内部 OSS 上传供 device 调用

ucg-service MUST 注册 `POST /ucg/internal/api/media/upload`，接受 multipart 文件，鉴权 MUST 使用网关内部密钥；响应 MUST 含 `objectKey` 与 `cdnUrl`。

#### Scenario: device 内部上传成功

- **WHEN** device-service 携带有效内部密钥上传 png
- **THEN** ucg-service MUST 写入 OSS 并返回 objectKey 与 cdnUrl

### Requirement: 迁移完成后 SHALL 下线 ai_talk_images 静态链路

系统 MUST NOT 再注册 `GET /ai_talk_images/*` 静态读或网关反代；Docker compose MUST NOT 再挂载宿主机 `/ai_talk_images` 目录。

#### Scenario: 旧静态 URL 不可用

- **WHEN** 迁移与部署完成后请求 `GET /ai_talk_images/event_1.png`
- **THEN** 网关或 device-service MUST NOT 再提供该静态文件

---

### Requirement: UCG 润笔 polish MUST 删除

ucg-service MUST 删除 `POST /ucg/app/api/posts/polish` 及其控制器/服务实现（`PolishPostText` 等）、polish lane 配置与 Admin AI config 中仅服务润笔的项。App 调用原 polish 路径 MUST 不再执行润笔 LLM。

#### Scenario: polish 接口不可用

- **WHEN** 客户端 POST 原 `/ucg/app/api/posts/polish`
- **THEN** 服务端 MUST NOT 再执行润笔生成逻辑

### Requirement: UCG polish AI 额度 MUST 删除

ucg-service MUST 删除 polish 月度额度（`ai_quota_*` 在 UCG 库中的 polish 列与 App/Admin 额度 API）。MUST NOT 删除 UCG 社区本体能力。

#### Scenario: UCG ai-quota API 下线

- **WHEN** 客户端请求原 `GET /ucg/app/api/ai-quota`
- **THEN** 服务端 MUST NOT 再返回 polish 额度业务响应

### Requirement: UCG 社区本体 MUST 保留

删除润笔与额度时，系统 MUST 保留媒体、发帖/改删、Feed、关注/点赞/投票/评论、私信与 WS chat、资料、Green 审核与 Admin 审帖等社区能力。

#### Scenario: 发帖与 Feed 仍可用

- **WHEN** 用户创建帖子或拉取 recommend/following Feed
- **THEN** 接口 MUST 仍按既有社区契约工作，MUST NOT 因删除 polish 而 5xx

---

### Requirement: UCG admin SHALL authenticate post moderation APIs with X-Admin-Password

`ucg-service` Admin 动态审查接口 MUST 与现有 UCG Admin 共用认证：请求 Header `X-Admin-Password` MUST 等于配置项 `ucg.admin.password`；校验失败 MUST 返回未授权，且 SHALL NOT 返回帖子数据或执行驳回。

#### Scenario: 口令正确访问列表

- **WHEN** 管理员携带正确 `X-Admin-Password` 请求 `GET /ucg/admin/api/posts/list`
- **THEN** 系统 SHALL 返回分页帖子列表

#### Scenario: 口令错误拒绝

- **WHEN** 管理员携带错误或未携带 `X-Admin-Password` 请求任一动态审查 Admin API
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改 `ucg_post`

### Requirement: Admin SHALL list all posts with optional status filter

`GET /ucg/admin/api/posts/list` MUST 支持查询参数 `page`（从 1 开始）、`pageSize`（默认 20，最大 100）、可选 `status`（0/1/2/3）。 省略 `status` 时 SHALL 返回全部状态的帖子。 列表 MUST 按 `updated_at` 降序排序。 响应 MUST 包含 `list`、`total`、`page`、`pageSize`； 每项 SHALL 至少包含 `id`、`authorWxId`、`content`、`status`、`rejectReason`、`createdAt`、`updatedAt`、`publishedAt` 及媒体展示字段。 `media` 数组 MUST 包含该帖 **全部** 媒体项（按 `sortOrder`）； 每项 SHALL 含 `cdnUrl`、`mediaKind`（1=图片，2=视频），图片 SHALL 含物理缩略图 `thumbnailUrl`（`BuildImageThumbnailURL`，path 为.

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

html` MUST 在现有 UCG Admin 登录态下提供「动态审查」模块（可与 AI 配置以 Tab 切换）。 模块 SHALL 调用列表 API 展示表格，对 `status≠3` 的行提供 checkbox； SHALL 提供「全选本页可驳回项」、**「批量通过」**与「批量驳回」按钮。 批量驳回前 MUST 弹出理由输入且 MUST NOT 在理由为空时提交； 批量通过前 MUST 经用户确认。 **批量驳回 MUST 在请求 body 中携带非空 `reason`。 ** `status=3` 的行 checkbox SHALL 禁用或不可选。 操作成功后 SHALL 刷新当前列表。 表格 SHALL 含 **驳回原因** 列（展示 `rejectReason`）。 状态筛选 SHALL 含 0–5（含「发布失败(4)」「机审失败(5)」）。 表格「媒体」列 SHALL 展示每条动态 **全量** 媒体缩略图； 图片 SHALL 支持 modal 原图； 视频 SHALL 展示首帧缩略图并支持 modal 播放。 动态审查.

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

### Requirement: device admin entry SHALL link to UCG management

`resource/public/admin.html` 中指向 `/device/admin/ucg-admin.html` 的入口链接文案 SHALL 为「UCG 管理」（或等价中文），以涵盖 AI 配置与动态审查。

#### Scenario: 设备管理页入口文案

- **WHEN** 管理员打开设备管理页
- **THEN** UCG 入口链接可见文案 SHALL 为「UCG 管理」

### Requirement: Admin SHALL batch approve posts for human publish

`POST /ucg/admin/api/posts/approve` MUST 接受 JSON body：`postIds`（非空数组，最多 100 个 id）。 对每条目标帖子：

- 若 `status=2`（published）SHALL 计入 `skipped` 且不更新行。 - 若 `status=1`（pending_audit）、`4`（apply_failed）或 `5`（moderation_failed），系统 MUST 在不调用 Green 的前提下将帖子发布：`status` 置为 `2`，写入 `published_at` 与 `updated_at`，清空面向作者的 `reject_reason`（及 apply 失败相关字段）； `status=5` 时 MUST 写入 `moderation_verdict=pass` 与 `moderation_at`； `status=1` 且 `moderation_verdict=0` 时 MUST 写入 `moderation_verdict=pass`。 成功后.

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

---

### Requirement: UCG Admin SHALL list profile audit jobs in moderation_failed state

ucg-service MUST 提供 admin HTTP API（Header `X-Admin-Password` 与现有 UCG admin 一致）：

- `GET /ucg/admin/api/profile-audit-jobs/list` — 分页查询 `ucg_profile_audit_job`；默认筛选 `status=ProfileJobStatusModerationFailed(5)`；可选 query 覆盖 status/page/pageSize。

响应 MUST 含：jobId、wxId、auditVersion、nickname/avatarKey/bio（job patch 字段）、rejectReason（机审失败日志）、createdAt、updatedAt。

#### Scenario: 列表默认仅机审失败

- **WHEN** 管理员 GET list 且未传 status
- **THEN** 响应 SHALL 仅包含 `status=5` 的 job 行

#### Scenario: 未授权拒绝

- **WHEN** 请求缺少或错误 `X-Admin-Password`
- **THEN** API SHALL 返回未授权错误且 MUST NOT 返回 job 数据

### Requirement: UCG Admin SHALL manually resolve moderation_failed profile jobs

`POST /ucg/admin/api/profile-audit-jobs/resolve` MUST 接受 `jobId`（必填）、`action`（`approve` | `reject`）、`reason`（reject 时必填）。 MUST 按 job 非空 patch 字段更新 `ucg_profile`（与 `approveProfileJobCAS` 语义一致）； MUST NOT 更新已发布 profile。 操作 MUST 使用 CAS（`status=5` + 匹配 `audit_version`），并发重复请求 MUST 幂等（0 行 affected 返回成功或明确 skip）。

#### Scenario: 人工通过应用 patch

- **WHEN** 管理员对 status=5 的 job 执行 approve 且 job.bio 非空
- **THEN** job SHALL 变为 approved，且 `ucg_profile.bio` SHALL 更新为 job 中 bio

#### Scenario: 人工驳回

- **WHEN** 管理员对 status=5 的 job 执行 reject 并填写 reason
- **THEN** job SHALL 变为 rejected 且 reject_reason SHALL 等于所填 reason，且已发布 profile MUST 不变

#### Scenario: 非 moderation_failed 拒绝

- **WHEN** 管理员对 status≠5 的 job 调用 resolve
- **THEN** API SHALL 返回参数/状态错误且 MUST NOT 变更 job 或 profile

### Requirement: ucg-admin.html SHALL provide profile moderation_failed review UI

静态页 MUST 在「资料机审失败」Tab 内：

- 调用 list API 展示表格（wxId、patch 摘要、失败原因、时间）；
- 每行提供「通过」「驳回」按钮；驳回 MUST 弹窗收集 reason；
- 操作成功后 MUST 刷新当前页列表。

#### Scenario: 通过后列表减少

- **WHEN** 管理员点击某行「通过」且 API 成功
- **THEN** 该行 MUST 从当前列表消失（或刷新后不再出现）

---

### Requirement: ucg-service 阿里云凭证 MUST 经环境变量注入且 yaml 不得含明文

`ucg-service` 运行时使用的阿里云 OSS AccessKey ID/Secret 与 DashScope API Key MUST 来自容器环境变量，不得依赖 `manifest/config/config. dashscope_api_key` MUST 留空字符串。 Green 内容审核 MUST 复用 OSS 凭证（`LoadGreenConfig` fallback），MUST NOT 要求独立的 Green AccessKey 环境变量。 环境变量名 MUST 为：

| 用途 | 环境变量 |
|------|----------|
| OSS AccessKey ID | `UCG_OSS_ACCESS_KEY_ID` |
| OSS AccessKey Secret | `UCG_OSS_ACCESS_KEY_SECRET` |
| DashScope API Key | `UCG_DASHSCOPE_API_KEY` |

#### Scenario: Compose 启动 ucg-service 时注入 OSS 凭证

- **WHEN** 运维使用 `docker compose --env-file manifest/docker/.env.prod` 启动含 `ucg-service` 的栈，且 `.env.prod` 含 `UCG_OSS_ACCESS_KEY_ID` 与 `UCG_OSS_ACCESS_KEY_SECRET`
- **THEN** ucg-service 容器环境 MUST 可见上述变量且 presign 接口 SHALL 可成功返回 uploadUrl

#### Scenario: yaml 无明文时仍可通过 env 润笔

- **WHEN** `config.ucg-service.yaml` 中 `dashscope_api_key` 为空且容器 env 设置 `UCG_DASHSCOPE_API_KEY`
- **THEN** AI 润笔运行时配置 MUST 使用该 key 且 SHALL NOT 因 yaml 为空而单独失败

#### Scenario: Green 复用 OSS env 凭证

- **WHEN** `ucg.green.enabled` 为 true 且 Green yaml 中 accessKey 为空，容器 env 已设置 `UCG_OSS_ACCESS_KEY_*`
- **THEN** Green 客户端 MUST 使用与 OSS 相同的 AccessKey 发起审核

### Requirement: Docker Compose 基线 MUST pass-through ucg 阿里云 env

`manifest/docker/docker-compose.microservices.yml` 的 `ucg-service.environment` MUST 包含 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY` 的 `${VAR:-}` 引用，使 local/test/prod overlay 共享同一注入点。

#### Scenario: 基线 compose 引用 env 文件变量

- **WHEN** 开发者查看 `docker-compose.microservices.yml` 中 `ucg-service` 段
- **THEN** MUST 可见上述三个 environment 条目且格式与其它 secrets pass-through 一致

### Requirement: 部署 env 文件 MUST 含真实凭证且 example 仅含占位符

`manifest/docker/.env.local`、`.env.test`、`.env.prod` MUST 包含真实值的 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY`（与本变更实施时 yaml 中既有凭证一致，由实施者填入）。仓库 MUST 提供 `manifest/docker/.env.example`，列出相同 key 名与注释说明，MUST NOT 含真实密钥。

#### Scenario: example 文件可供新环境复制

- **WHEN** 新成员复制 `.env.example` 为本地 `.env` 并填入凭证
- **THEN** key 名 MUST 与 compose pass-through 及 Go 代码读取的 env 名完全一致

### Requirement: runbook MUST 文档化 ucg 阿里云 env 约定

`docs/runbooks/release-deploy-and-run.md` MUST 说明 ucg-service 部署时必需的三项阿里云相关 env、与 yaml 留空的关系，并指向 `.env.example`。

#### Scenario: runbook 检索 env 名

- **WHEN** 运维查阅 release runbook 中 ucg-service 或 secrets 相关章节
- **THEN** MUST 能找到 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY` 的说明

---

### Requirement: App HTTP API SHALL expose UCG REST under /ucg/app/api

ucg-service SHALL implement REST endpoints (also reachable via gateway proxy) including: profile get/update, feed recommend/following, posts CRUD, media presign, follow/unfollow, likes, comments, conversations and messages list。 **推荐 Feed**（`GET /ucg/app/api/feed/recommend`）MUST 使用 **cursor 分页**（见 ADDED Requirement），**不适用** `{ total, page, pageSize }` 契约。 除推荐 Feed 与评论列表外，其他列表分页 MUST 仍使用 `page`（从 1 开始）与 `pageSize`（默认 20，最大 50），响应 MUST 包含 `{ list, total, page, pageSize }`。.

#### Scenario: 推荐 Feed cursor 分页

- **WHEN** `GET /ucg/app/api/feed/recommend?pageSize=20&lat=31.2&lng=121.5`
- **THEN** 响应 SHALL 仅含 `status=2` 的帖子，且 SHALL 含 `list`、`hasMore`，MAY 含 `nextCursor`；MUST NOT 含 `total`

#### Scenario: 我的动态含全状态

- **WHEN** 作者 `GET /ucg/app/api/posts/mine`
- **THEN** 响应 SHALL 含 draft/pending/rejected/published 本人帖子

#### Scenario: 关注 Feed 需身份

- **WHEN** 未带有效 `X-Internal-Wx-Id` 请求 following feed
- **THEN** 服务 SHALL 返回未授权错误

### Requirement: Profile API SHALL auto-create default nickname from baby name

On first access, ucg-service SHALL create `ucg_profile` with nickname `{babyName}的家长` fetched via device internal API when profile missing.

#### Scenario: 首次进入 UCG
- **WHEN** 已登录用户首次请求 `/profile/me` 且无 profile 行
- **THEN** 服务 SHALL 创建 profile 且 nickname SHALL 使用 device 返回的 baby_name 拼接

### Requirement: 评论列表 SHALL 单次返回升序全量并批量填充作者 profile

`GET /ucg/app/api/posts/{id}/comments` MUST 仅对已发布（`status=2`）帖子返回评论。 评论 MUST 按 `created_at` **升序**排序（最早在前、最新在列表底部）。 响应 MUST 为 JSON：

- `list`：评论数组，每项 SHALL 含 `id`、`postId`、`authorWxId`、`content`、`createdAt`，且 SHOULD 含 `author`（公开 profile，与帖子作者 profile 同为实时 `ucg_profile` 语义，MUST NOT 依赖评论表快照列）
- `total`：SHALL 优先为帖子 `comment_count`； 若不可用 SHALL 回退为 `len(list)`
- `truncated`：布尔； 当评论数超过配置上限且仅返回部分行时为 `true`，否则为 `false`

服务 MUST NOT 对评论列表执行独立 `COUNT(*)` 查询。 服务 MUST 通过单次.

#### Scenario: 常规模帖单次拉取全量评论

- **WHEN** 客户端 `GET /ucg/app/api/posts/{id}/comments` 且该帖已发布、评论数不超过 `listMax`
- **THEN** 响应 SHALL 含全部评论且 `list` 按 `created_at` 升序
- **AND** `truncated` SHALL 为 `false`
- **AND** `total` SHALL 等于帖子 `comment_count`
- **AND** 每条评论的 `author` SHALL 经批量 profile 查询填充

#### Scenario: 超长帖评论截断

- **WHEN** 帖子 `comment_count` 为 600 且 `listMax=500`
- **THEN** 响应 `list` SHALL 含 500 条（最早 500 条）
- **AND** `truncated` SHALL 为 `true`
- **AND** `total` SHALL 为 600

#### Scenario: 发表评论响应可供乐观追加

- **WHEN** 客户端 `POST /ucg/app/api/posts/{id}/comments` 成功
- **THEN** 响应 SHALL 含完整评论字段及 `author`
- **AND** 客户端 SHALL 可将该条追加至本地 `list` 末尾而无需再次 GET 全列表

#### Scenario: 评论列表不使用 Redis

- **WHEN** ucg-service 处理评论列表读请求
- **THEN** 系统 MUST NOT 读写 Redis 中的评论列表或 profile 缓存键

### Requirement: 评论列表 GET SHALL 不计入 App API 使用统计

负责人已确认：`GET /ucg/app/api/posts/{id}/comments` MUST NOT 计入 gateway-app App API 使用统计。gateway-app SHALL 在 `maintenance_skip.go` 排除该 GET 路径（原始 `req.URL.Path` 为 `/ucg/app/api/posts/<postId>/comments`；归一化 apiKey 为 `GET /ucg/app/api/posts/{id}/comments`）。`POST /ucg/app/api/posts/{id}/comments`（发表评论）SHALL 仍计入统计。

#### Scenario: 成功 GET 评论列表不写入 usage

- **WHEN** 客户端成功调用 `GET /ucg/app/api/posts/{id}/comments` 经 gateway-app
- **THEN** gateway-app MUST NOT 将该请求计入 App API 使用统计

#### Scenario: 发表评论仍计入 usage

- **WHEN** 客户端成功调用 `POST /ucg/app/api/posts/{id}/comments` 经 gateway-app
- **THEN** gateway-app SHALL 将该请求计入 App API 使用统计

### Requirement: Comments API SHALL reflect audit pending and rejection

`POST .../posts/{id}/comments` 在 Green 完成前 MUST 返回 `status=1`（或等价 pending 字段）。`GET .../posts/{id}/comments` MUST 仅返回 `status=2`（published）评论给 **非作者** 视角；作者 MAY 看到自身 pending/rejected 评论及 reject_reason。

#### Scenario: 他人看不到待审评论

- **WHEN** 用户 A 发表评论且未审过，用户 B 拉取评论列表
- **THEN** 列表 MUST NOT 包含 A 的该条评论

#### Scenario: 作者见违规评论

- **WHEN** 用户 A 评论 Green fail
- **THEN** A 拉取评论或评论详情 MUST 可见 reject_reason

### Requirement: Profile me API SHALL read pending patch from audit job

`GET /ucg/app/api/profile/me` 对作者 MUST 合并 MySQL 最新 pending `ucg_profile_audit_job`（`auditPending=true`、预览 nickname/avatar/bio），MUST NOT 依赖 Redis profile pending 键作为长期权威。待审预览与 MQ/CAS 的版本语义 MUST 以 job 表 `audit_version` 为准，MUST NOT 从 `ucg_profile` 或 Redis 读取审核轮次。

#### Scenario: 待审头像预览

- **WHEN** 用户提交新 avatar 且 job pending
- **THEN** profile/me MUST 返回 `auditPending=true` 且 avatar 预览为新 key

### Requirement: Post and feed media DTOs SHALL expose physical video thumbnailUrl

App HTTP API 返回的帖子/动态媒体项（含推荐 Feed、关注 Feed、帖子详情、我的动态等经 `loadPostMedia` 或等价路径组装的 `media` 数组）中，当 `mediaKind=2`（视频）且 `objectKey` 非空时，MUST 填充 `thumbnailUrl` 为 `BuildVideoThumbnailURL(objectKey)` 物理首帧 jpg CDN URL。该 URL MUST NOT 含 `x-oss-process` query。`cdnUrl` MUST 仍为可播放的 mp4 CDN URL。

本要求 MUST NOT 改变既有 v1/v2 响应 JSON 字段名或结构，仅变更 `thumbnailUrl` 的 URL 形态。

#### Scenario: Feed 中视频帖含物理 thumbnailUrl

- **WHEN** 客户端 `GET /ucg/app/api/feed/recommend` 且列表含已发布视频帖（OSS 已存在 `_thumb.jpg`）
- **THEN** 对应 `media` 项 `thumbnailUrl` SHALL 以 `_thumb.jpg` 结尾且 SHALL NOT 含 `x-oss-process`

#### Scenario: 视频 media 仍保留可播放 cdnUrl

- **WHEN** 帖子含视频媒体
- **THEN** 该项 `cdnUrl` SHALL 指向 mp4 原片且 `thumbnailUrl` SHALL 与 `cdnUrl` path 不同

### Requirement: v2 创建帖 API SHALL 支持可选坐标

系统 MUST 提供 `POST /ucg/app/api/v2/posts`，请求体 MUST 兼容 v1 创建帖字段并 MAY 含可选 `lat`、`lng`（WGS84 十进制度）。成功响应 MUST 与 v1 创建帖相同结构（`UcgPostItem`）。该 endpoint MUST 计入 gateway-app App API 使用统计。

#### Scenario: v2 创建带坐标

- **WHEN** 客户端 `POST /ucg/app/api/v2/posts` 含合法 body 与 `lat`/`lng`
- **THEN** 服务 MUST 持久化坐标（MySQL 或等价）且在 publish 后 GEOADD 与 snapshot 含坐标

#### Scenario: v2 创建无坐标

- **WHEN** 客户端 `POST /ucg/app/api/v2/posts` 不含 lat/lng
- **THEN** 服务 MUST 成功创建且帖 MUST NOT 进入 GEO 索引

#### Scenario: v2 创建计入 usage

- **WHEN** 经 gateway-app 的 `POST /ucg/app/api/v2/posts` 返回 HTTP 2xx
- **THEN** gateway-app MUST 将该请求计入 App API 使用统计

### Requirement: 帖子读写 API SHALL 可选接受 viewer 坐标并返回 distanceMeters

下列接口 MUST 支持可选 viewer 坐标；当 viewer 与目标帖均有坐标时，响应 MUST 含 `distanceMeters` 字符串（米，纯数字字符串，如 `"1234"`），MUST NOT 含距离文案标签；否则 MUST 省略该字段：

- `GET /ucg/app/api/feed/recommend` — query `lat`、`lng`（可选）；**例外**：当 viewer 已登录且为帖子作者时，该 item MUST omit `distanceMeters`（见 `ucg-recommend-feed` 本人帖场景）
- `GET /ucg/app/api/posts/{id}` — query `lat`、`lng`（可选）
- `PUT /ucg/app/api/posts/{id}` — body 可选 `lat`、`lng`（更新帖坐标）

#### Scenario: Feed item 含距离

- **WHEN** `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且帖有坐标且 viewer 非该帖作者
- **THEN** 该帖 JSON MUST 含 `distanceMeters` 且值为米数字符串

#### Scenario: 推荐 Feed 本人帖 omit 距离

- **WHEN** `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且帖有坐标且 viewer 为该帖作者
- **THEN** 该帖 JSON MUST NOT 含 `distanceMeters`

#### Scenario: 无坐标省略距离

- **WHEN** `GET /ucg/app/api/posts/{id}` 无 lat/lng query 或帖无坐标
- **THEN** 响应 MUST NOT 含 `distanceMeters` 字段

#### Scenario: 更新帖坐标

- **WHEN** `PUT /ucg/app/api/posts/{id}` body 含新 lat/lng
- **THEN** 服务 MUST 更新存储坐标并同步 Redis GEO 与 post snapshot

### Requirement: 推荐 Feed cursor 参数 SHALL 冻结 session 上下文

`GET /ucg/app/api/feed/recommend` MUST 接受 query：

- `cursor`（可选，opaque）
- `pageSize`（可选，默认 20，最大 50）
- `lat`、`lng`（可选，仅首屏无 cursor 时生效）

有 `cursor` 时 MUST 忽略新的 lat/lng。响应 `nextCursor` MUST 在 `hasMore=true` 时存在。

#### Scenario: 翻页携带 cursor

- **WHEN** 客户端用上一页 `nextCursor` 请求第二页
- **THEN** 系统 MUST 使用 cursor 内冻结坐标与 session 且 MUST 返回不重复 postId

#### Scenario: 下拉刷新无 cursor

- **WHEN** 客户端不传 `cursor` 请求 Feed
- **THEN** 系统 MUST 创建新 feed session 且 MAY 使用本次 lat/lng

### Requirement: 关注 Feed API SHALL 保持 page/total 分页并可选接受 viewer 坐标

`GET /ucg/app/api/feed/following` MUST 继续要求有效 `X-Internal-Wx-Id`。MUST 接受 query `page`（从 1）、`pageSize`（默认 20，最大 50）及可选 `lat`、`lng`。响应 MUST 为 `{ list, total, page, pageSize }`（`UcgPageRes`），**MUST NOT** 含 `nextCursor` 或 `hasMore`。本 endpoint 为既有路由的 query 扩展，**MUST NOT** 作为新 App API 计入 usage 统计。

#### Scenario: 关注 Feed 分页契约不变

- **WHEN** `GET /ucg/app/api/feed/following?page=2&pageSize=20`
- **THEN** 响应 MUST 含 `list`、`total`、`page`、`pageSize`
- **AND** MUST NOT 含 `nextCursor` 或 `hasMore`

#### Scenario: 关注 Feed 可选坐标 query

- **WHEN** `GET /ucg/app/api/feed/following?page=1&lat=31.2&lng=121.5`
- **THEN** 服务 MUST 接受 lat/lng 并用于列表 item 距离计算
- **AND** 分页字段 MUST 仍为 `{ list, total, page, pageSize }`

#### Scenario: 关注 Feed 需身份

- **WHEN** 未带有效 `X-Internal-Wx-Id` 请求 following feed
- **THEN** 服务 SHALL 返回未授权错误

---

### Requirement: 公开 profile 返回社交统计

`GET /ucg/app/api/profile/{wxId}` MUST 在响应中返回与 `GET /ucg/app/api/profile/me` 一致的 `followingCount`、`followerCount`、`postCount`，数值 MUST 经 ucg 库内实时 COUNT（`ucg_follow`、`ucg_post`），与 `enrichProfileStats` 语义相同。

#### Scenario: 他人主页展示关注数

- **WHEN** 客户端请求 `GET /ucg/app/api/profile/{wxId}` 且该 wx 存在 profile 行
- **THEN** 响应 MUST 包含 `followingCount` 等于 `COUNT(ucg_follow WHERE follower_wx_id = wxId)`
- **AND** 响应 MUST 包含 `followerCount` 等于 `COUNT(ucg_follow WHERE followee_wx_id = wxId)`
- **AND** 响应 MUST 包含 `postCount` 等于 `COUNT(ucg_post WHERE author_wx_id = wxId)`

#### Scenario: profile 不存在

- **WHEN** 请求的 wxId 无 `ucg_profile` 行
- **THEN** MUST 返回 404，行为与变更前一致

---

### Requirement: UCG 审核事件 SHALL 使用已注册的 RabbitMQ routing keys

`ucg-service` 发布 UCG 审核事件时 MUST 使用下列 routing key 之一，且 MUST 经 `eventkit` 注册校验通过：

- `ucg.post.created`

- `ucg.comment.created`

- `ucg.profile.patch.submitted`

- `ucg.chat.msg.created`

#### Scenario: 发帖后发布事件

- **WHEN** 用户 submit 创建帖子且数据库事务提交成功

- **THEN** 系统 MUST Publish `ucg.post.created`，载荷至少含 `postId` 与 `auditVersion`（等于 `ucg_post.audit_version` 当前值）

#### Scenario: 未注册 routing key 拒绝发布

- **WHEN** 代码尝试 Publish 未在 eventkit 注册的 ucg 路由键

- **THEN** Publish MUST 失败且 MUST NOT 静默丢弃

### Requirement: 所有 UCG 审核 MQ 载荷 MUST 携带 auditVersion

四类审核事件的 JSON 载荷 MUST 含非空 `auditVersion`（INT），且 MUST 等于 **入队 outbox 时刻**对应权威表列的当前值（冻结在 outbox `payload` 内）：

- 帖子：`ucg_post.audit_version`
- 评论：`ucg_post_comment.audit_version`
- 资料：`ucg_profile_audit_job.audit_version`（载荷 MUST 含 `jobId`）
- 私信：`ucg_chat_message.audit_version`（载荷 MUST 含 `messageId` 与 `conversationId`）

relay worker Publish MUST 使用 outbox 内冻结的 `payload`，MUST NOT 在重试时从业务表重新读取版本（避免与用户再提审后的新版本混淆）。

#### Scenario: 资料审核载荷含 job 版本

- **WHEN** 用户提交资料变更且 job 行 `audit_version=2` 入队 outbox
- **THEN** outbox `payload` MUST 含 `jobId` 与 `auditVersion=2`

#### Scenario: 评论审核载荷含版本

- **WHEN** 用户发表评论且评论行 `audit_version=1` 入队 outbox
- **THEN** outbox `payload` MUST 含 `commentId` 与 `auditVersion=1`

### Requirement: RabbitMQ 拓扑 SHALL 为 UCG 审核队列绑定 topic exchange

仓库 `hack/rabbitmq-init` 脚本 MUST 为上述四个 routing key 创建 durable 队列并完成与 `voice.events` topic exchange 的 binding。runbook MUST 文档化队列名与初始化步骤。

#### Scenario: 本地/测试环境 init 后队列存在

- **WHEN** 运维执行 rabbitmq init 脚本

- **THEN** 管理台 SHALL 可见 `ucg.post.created.q` 等队列且 binding 正确

### Requirement: UCG 审核消费者 SHALL 驻留 ucg-service 并按 routing key 分发

`ucg-service` MUST 启动 UCG 审核 **AMQP push consumer**（TCP 5672，`autoAck=false`），从四个 UCG 审核 durable 队列接收 broker 推送的消息并调用 Green 审核逻辑。consumer MUST NOT 部署在 worker-service 内直连 ucg 数据库。消费者 MUST 从载荷读取 `auditVersion` 用于 CAS。每个 UCG 审核队列 MUST 运行 **一个** AMQP consumer goroutine；并发 MUST 由 channel **`prefetch`**（环境变量 `UCG_AUDIT_MQ_PREFETCH`，默认 5）控制。MUST NOT 使用 HTTP Management API `/queues/.../get` 轮询作为 UCG 审核消费路径。

#### Scenario: 消费帖子审核消息

- **WHEN** 队列 `ucg.post.created.q` 收到合法 `ucg.post.created` 载荷（含 `postId`、`auditVersion`）
- **THEN** consumer MUST 执行 Green 审核并在成功后 CAS 更新帖子状态（条件含 `status` 与 `audit_version`），且 MUST 在处理完成后 **manual Ack**

#### Scenario: 四队列并行消费

- **WHEN** ucg-service 启动且 `UCG_AUDIT_MQ_CONSUMER_ENABLED=true`
- **THEN** 系统 MUST 对 `ucg.post.created.q`、`ucg.comment.created.q`、`ucg.profile.patch.submitted.q`、`ucg.chat.msg.created.q` 各建立 AMQP Consume，且 MUST NOT 依赖 ticker 轮询依次拉取四队列

### Requirement: 审核消费者 MUST 通过 audit_version CAS 更新且过期消息 SHALL 优雅跳过

所有 MQ 审核消费者在写审态时 MUST 使用条件更新：`WHERE id=? AND status=? AND audit_version=?`（私信为 `audit_status` + `audit_version`）。CAS 成功时 MUST NOT 递增 `audit_version`。`RowsAffected=0` MUST 视为重复投递或过期版本（如用户已再提审 bump 版本），MUST ACK 且 MUST NOT 无限重试，MUST NOT 覆盖新轮次状态。

#### Scenario: 重复消费同一 post 事件

- **WHEN** 同一 `postId` 与 `auditVersion` 的事件被投递两次且首次已 CAS 成功（status 已非 pending）

- **THEN** 第二次 CAS MUST 影响 0 行且 MUST NOT 将已 published 帖改回 pending

#### Scenario: 再提审后旧版本消息过期

- **WHEN** 用户再提审使 `ucg_post.audit_version` 从 2 递增为 3，队列中仍存在 `auditVersion=2` 的消息

- **THEN** consumer CAS `status=1 AND audit_version=2` MUST 影响 0 行且 MUST ACK，当前帖 MUST 保持 version=3 的 pending 状态

#### Scenario: 资料 CAS 仅针对 job 表版本

- **WHEN** consumer 处理 `ucg.profile.patch.submitted` 且载荷 `auditVersion=1`

- **THEN** UPDATE MUST 针对 `ucg_profile_audit_job` 且 MUST 使用 `status=1 AND audit_version=1`；MUST NOT 以 Redis 或 `ucg_profile` 列作为 CAS 版本

### Requirement: MQ Publish 失败 SHALL 可补发且 MUST NOT 阻塞用户主路径（聊天除外可选）

帖/评/资料 submit API 在事务与 outbox 入队成功但 HTTP Publish 失败时 MUST 记录 warning/error 日志；系统 MUST 通过 **`ucg_audit_publish_outbox` relay worker** 自动重试 Publish（使用 outbox 冻结载荷）。系统 **MUST NOT** 运行定时扫描 pending 审态业务表的 reconciler 作为恢复机制。聊天 WS 在 Redis 投递成功且 outbox 入队成功但 Publish 失败时 MUST 同样由 relay worker 恢复。

#### Scenario: 发帖 Publish 失败仍返回成功

- **WHEN** 帖子已写入 `status=1` 且 outbox 已 commit 但 RabbitMQ 暂不可用
- **THEN** HTTP 创建接口 MAY 仍返回 200 与帖子 DTO；relay worker MUST 在 MQ 恢复后 Publish 并成功标记 outbox `done`

### Requirement: UCG AMQP 审核 consumer MUST 使用 manual ack 且 SHALL 按处理结果 Ack 或 Nack

UCG 审核 AMQP consumer MUST 以 `autoAck=false` 订阅队列。消息 MUST 在 Green + CAS 业务处理完成且 handler 返回成功后 **Ack**。下列情况 MUST **Ack** 且 MUST NOT requeue：JSON 非法、缺 `auditVersion`、实体不存在、CAS `RowsAffected=0`（过期/重复）。Green 或数据库 **可重试** 错误 MUST **Nack(requeue=true)**。进程在处理完成前崩溃时，broker MUST 能重投未 Ack 的消息。

#### Scenario: 处理成功后 Ack

- **WHEN** consumer 收到合法帖子审核消息且 Green 通过、CAS 更新成功
- **THEN** consumer MUST 向 broker 发送 Ack，该消息 MUST NOT 再次投递给同一 consumer

#### Scenario: 可重试错误 Nack requeue

- **WHEN** Green API 返回临时错误或 MySQL 更新失败且非 CAS 0 行跳过
- **THEN** consumer MUST Nack(requeue=true)，消息 MUST 可再次被消费

#### Scenario: 毒消息 Ack 丢弃

- **WHEN** 消息体非合法 JSON 或缺少 `auditVersion`
- **THEN** consumer MUST 记录 warning 日志并 Ack，MUST NOT 无限 requeue

### Requirement: UCG AMQP consumer SHALL 支持连接断线重连

当 RabbitMQ AMQP 连接或 channel 异常关闭时，ucg-service 中的 UCG 审核 consumer MUST 以 backoff 策略自动重连并恢复四队列 Consume。AMQP 连接失败 MUST NOT 导致 ucg-service HTTP API 进程退出；MUST 记录可观测 warning/error 日志。

#### Scenario: RabbitMQ 短暂不可用后恢复

- **WHEN** ucg-service 运行中 RabbitMQ 重启导致 AMQP 断开
- **THEN** consumer MUST 在连接恢复后继续消费队列消息，reconciler MUST 仍可补发 pending 条目

### Requirement: UCG 审核 Publisher MAY 保持 HTTP Management API

本变更 MUST NOT 要求 UCG 审核事件 Publisher 改为 AMQP。`ucg-service` MAY 继续通过 HTTP `MQ_HTTP_API_BASE` 向 `voice.events` exchange 发布审核事件；AMQP consumer 与 HTTP Publisher 并存 MUST 为受支持的部署形态。

#### Scenario: HTTP 发布 AMQP 消费

- **WHEN** 发帖后 HTTP Publisher 成功发布 `ucg.post.created` 至 exchange
- **THEN** 绑定队列中的消息 MUST 由 AMQP push consumer 接收并处理

---

### Requirement: UCG audit MQ consumer SHALL bound apply retries and stop infinite requeue

`ucg-service` 内 UCG 审核 AMQP consumer（含 `ucg.profile.patch.submitted.q`、`ucg.post.created.q` 及本变更纳入的其它审核队列）在 handler 处理单条 delivery 时 MUST 区分 **机审阶段**与 **apply 阶段**。当 apply 阶段失败且实体 `apply_attempts` 未达配置上限时，handler MAY 返回 error 触发 `Nack(requeue=true)`。当 `apply_attempts` 达到上限 `UCG_AUDIT_APPLY_MAX_ATTEMPTS`（默认 5，可通过环境变量覆盖）时，系统 MUST 将实体标记为 apply 失败终态、记录可观测告警日志，且 handler MUST 返回 nil 以 **Ack** 该 delivery，MUST NOT 无限 requeue。

#### Scenario: apply 失败未达上限时 requeue

- **WHEN** 资料 job 机审 verdict 已为 pass，但 `approveProfileJobCAS` 因 DB 错误失败，且 `apply_attempts` 为 2、上限为 5
- **THEN** 系统 MUST 将 `apply_attempts` 递增为 3，handler MUST 返回非 nil error，且该 MQ 消息 MUST 被 Nack requeue

#### Scenario: apply 失败达上限时 Ack 停止风暴

- **WHEN** 同上失败场景且递增后 `apply_attempts` 等于上限 5
- **THEN** 系统 MUST 标记 job 为 apply 失败终态（如 `status=apply_failed` 或等价字段组合），MUST 输出含 jobId 与 auditVersion 的 error 级日志，且 handler MUST 返回 nil 使消息 Ack，后续同 payload 重投 MUST NOT 再次调用 Green

#### Scenario: 队列 redelivery 不触发无限 Nack 环

- **WHEN** 同一 delivery 因历史缺陷已被 requeue 超过 apply 上限，且 DB 已记录 apply 失败终态
- **THEN** handler MUST 返回 nil Ack，队列 `messages_unacknowledged` MUST NOT 因该 job 无限增长

### Requirement: UCG audit handler SHALL treat persisted moderation verdict as Green skip signal

当 `ucg_profile_audit_job` 或 `ucg_post`（及纳入本能力的同类审核实体）在对应 `audit_version` 下已持久化 `moderation_verdict`（非 0）时，MQ consumer handler 在 Phase 2 apply 重试路径 MUST NOT 再次调用 `GreenModerator` 的 `ModerateText` / `ModerateImageURL` / `ModerateVideoURL`。

#### Scenario: MQ 重投跳过 Green 仅重试 apply

- **WHEN** profile job `id=100`、`audit_version=2` 的 `moderation_verdict=1`（pass），`status` 仍为 pending，且 MQ 消息因先前 apply 失败被 redeliver
- **THEN** handler MUST NOT 调用 Green API，MUST 仅执行 apply 阶段（`approveProfileJobCAS` 或等价逻辑）

#### Scenario: 首次消费执行 Green 并落库 verdict

- **WHEN** 新提交 profile job `moderation_verdict=0` 且收到首次 audit MQ 消息
- **THEN** handler MUST 调用 Green 完成机审，且 MUST 在 apply 之前将 `moderation_verdict` 与 `moderation_reason`（若 reject）持久化到 MySQL

#### Scenario: 并发双消费避免重复 Green

- **WHEN** 两条相同 job 的 delivery 几乎同时被不同 consumer 处理，且 `moderation_verdict` 仍为 0
- **THEN** Phase 1 持久化 MUST 使用带 `moderation_verdict=0` 条件的 CAS/UPDATE，至多一条成功写入 verdict；另一条 MUST 读取已写入 verdict 后跳过 Green 进入 apply

### Requirement: UCG audit runbook SHALL document stuck queue remediation

`docs/runbooks/release-deploy-and-run.md` MUST 包含 UCG 审核 MQ 卡死与 apply 失败的人工处理步骤，至少覆盖：暂停 consumer（`UCG_AUDIT_MQ_CONSUMER_ENABLED`）、核对 DB 中 `moderation_verdict` 与 `status`、在安全条件下 purge 队列积压、对 `moderation_verdict=pass` 且长期 pending 的 job 的修复指引。

#### Scenario: 运维按 runbook 处理 profile 队列积压

- **WHEN** `ucg.profile.patch.submitted.q` ready 消息数异常升高，且 DB 显示多条 job 已 `moderation_verdict=1` 但 `status=pending`
- **THEN** 运维人员 MUST 能按 runbook 暂停 consumer、确认无需重复 Green 后清理或重放消息，且 MUST NOT 仅无限重启 consumer 而不查 DB

---

### Requirement: 审核 Publish MUST 经 transactional outbox 持久化

`ucg-service` 在触发四类 UCG 审核 MQ 事件（`ucg.post.created`、`ucg.comment.created`、`ucg.profile.patch.submitted`、`ucg.chat.msg.created`）时，MUST 在与对应业务写库**同一数据库事务**内 INSERT 一行至 `ucg_audit_publish_outbox`（`status=pending`）。outbox 行 MUST 含 `routing_key` 与 JSON `payload`，且 `payload.auditVersion` MUST 等于入队时刻权威表上的当前 `audit_version`（冻结快照，非 relay 时再读库）。

#### Scenario: 发帖 submit 同事务入队

- **WHEN** 用户 submit 创建帖子且事务 INSERT `ucg_post` 成功
- **THEN** 同一事务 MUST INSERT outbox 行，`routing_key=ucg.post.created`，载荷含 `postId` 与 `auditVersion`

#### Scenario: 再提审 bump 版本后入队

- **WHEN** 用户再提审使 `ucg_post.audit_version` 递增为 3 且事务提交
- **THEN** outbox 新行 MUST 含 `auditVersion=3`（非旧版本 2）

### Requirement: Audit Publish Relay Worker MUST 仅轮询 outbox 表

`ucg-service` MUST 运行 `StartAuditPublishRelayWorker`，按配置间隔从 `ucg_audit_publish_outbox` 选取 `status IN (pending, failed)` 且 `attempts < maxAttempts` 的行（`ORDER BY id ASC LIMIT 1`），经 HTTP 向 `voice.events` Publish。成功 MUST 将行标记为 `done`；失败 MUST 递增 `attempts`、记录 `last_error` 并标记 `failed`（未达 maxAttempts 时 MUST 可被后续 tick 重试）。Worker MUST NOT 扫描 `ucg_post`、`ucg_post_comment`、`ucg_profile_audit_job`、`ucg_chat_message` 的 pending 审态以发现漏发事件。

#### Scenario: RabbitMQ 短暂不可用后恢复

- **WHEN** outbox 行 Publish 连续失败且 RabbitMQ 恢复
- **THEN** relay worker MUST 重试该行直至 Publish 成功并标记 `done`

#### Scenario: Worker 不扫 pending 帖表

- **WHEN** 存在 `ucg_post.status=1` 但无对应 outbox 行（历史脏数据）
- **THEN** relay worker MUST NOT 因扫表而补发；恢复 MUST 依赖一次性运维 seed 或新业务路径

### Requirement: 事务提交后 SHOULD best-effort 即时 Publish

业务事务成功提交后，系统 SHOULD 尝试对刚写入的 outbox 行执行一次即时 HTTP Publish；若成功 MUST 将该行标记为 `done` 且 relay worker MUST NOT 重复 Publish 已成功行。

#### Scenario: 即时 Publish 成功

- **WHEN** 发帖事务提交且 RabbitMQ 可用
- **THEN** outbox 行 SHOULD 在 relay worker 介入前变为 `done`

### Requirement: Publish 失败 MUST NOT 阻塞用户主路径

帖/评/资料 submit API 与聊天 WS 投递在 outbox 入队成功但即时 Publish 失败时 MUST 仍完成用户可见的成功路径（与现网一致）；恢复 MUST 由 relay worker 承担，MUST NOT 依赖 pending 业务表 reconciler。

#### Scenario: 发帖 Publish 失败仍返回成功

- **WHEN** 帖子与 outbox 已 commit 但 RabbitMQ 暂不可用
- **THEN** HTTP 创建接口 MAY 仍返回 200；outbox 行 MUST 保持 `pending`/`failed` 直至 relay 成功

---

### Requirement: 私信发送 SHALL 同步写入 Redis 与 MySQL outbox

`DeliverChatMessage` 在 Green 审核通过后，SHALL 先执行 Redis `INCR seq` 与 `RPUSH` 消息 JSON 至 `ucg:chat:conv:{conversationId}:msgs`；随后 SHALL 同步 INSERT 一行至 `ucg_chat_message_outbox`（payload 含完整消息字段，status=pending）。Redis 写入失败 SHALL 中止发送；outbox 写入失败 SHALL 记录 Error 日志且 SHALL NOT 阻止 WS `message_delivered` 推送。

#### Scenario: 正常发送双写

- **WHEN** 用户发送私信且 Green 审核通过
- **THEN** Redis LIST 中 SHALL 包含该消息 JSON
- **AND** MySQL `ucg_chat_message_outbox` SHALL 存在对应 pending 行

#### Scenario: Redis 写入失败

- **WHEN** Redis RPUSH 失败
- **THEN** 系统 SHALL NOT 写入 outbox
- **AND** 系统 SHALL NOT 向接收方推送 message_delivered

### Requirement: persist worker SHALL 异步将 outbox flush 至 ucg_chat_message

ucg-service SHALL 运行 `StartChatPersistWorker`，轮询 `ucg_chat_message_outbox` 中 status=pending 或 failed（未超最大 attempts）的记录；SHALL INSERT 至 `ucg_chat_message`（utf8mb4），幂等处理重复 `(conversation_id, id)`；成功后 SHALL 将 outbox 行标记为 done。

#### Scenario: outbox 成功落库

- **WHEN** worker 处理 pending outbox 行且 MySQL 可用
- **THEN** `ucg_chat_message` SHALL 存在与 Redis msg id 一致的记录
- **AND** outbox 行 status SHALL 变为 done

#### Scenario: 重复 flush 幂等

- **WHEN** worker 对同一 outbox 行重试 INSERT
- **THEN** `ucg_chat_message` SHALL NOT 出现重复 `(conversation_id, id)` 行

### Requirement: 读消息 SHALL Redis 优先并在 miss 时回源 MySQL

`listChatMessages` SHALL 优先从 Redis LRANGE 分页返回。当 Redis `LLEN` 为 0 且 MySQL `ucg_chat_message` 对该会话 `COUNT(*) > 0` 时，SHALL 从 MySQL 按 id 升序分页查询并返回等效 `ChatMessage` 结构。

#### Scenario: Redis 有缓存

- **WHEN** Redis LIST 非空
- **THEN** 系统 SHALL 从 Redis 返回消息，SHALL NOT 查询 MySQL

#### Scenario: Redis 空且 MySQL 有历史

- **WHEN** Redis LIST 为空且 MySQL 存在该会话消息
- **THEN** 系统 SHALL 从 MySQL 返回消息列表

### Requirement: MySQL 回源读 SHALL 按需 lazy warm Redis 并对齐 seq

从 MySQL fallback 读取时，系统 SHALL 将读到的消息 JSON RPUSH 回 Redis（至少当前页；可配置整会话阈值）；SHALL 将 `ucg:chat:conv:{id}:seq` 设为 MySQL 该会话 `MAX(id)`（当 Redis seq 缺失或小于该值时）。lazy warm SHOULD 使用短 TTL 分布式锁避免并发重复重建。

#### Scenario: Redis 丢数据后首次读会话

- **WHEN** Redis 消息 LIST 为空但 MySQL 有消息且用户拉取历史
- **THEN** 用户 SHALL 看到 MySQL 中的消息
- **AND** Redis seq SHALL 对齐至 MySQL MAX(id)

### Requirement: ucg_chat_message 表 SHALL 使用 utf8mb4 存储正文

`ucg_chat_message.content` 及表默认字符集 SHALL 为 utf8mb4，以支持 emoji 等四字节 Unicode 入库。

#### Scenario: 含 emoji 的私信持久化

- **WHEN** 用户发送含 emoji 的私信且 worker flush 完成
- **THEN** MySQL 中该消息 content SHALL 原样保留 emoji

---

### Requirement: Chat SHALL use Redis for durable message storage and WebSocket delivery

ucg-service SHALL persist chat messages in Redis without TTL (forever retention in MVP), expose internal `GET /ws/chat` WebSocket with JWT auth first frame, and push real-time events to conversation members after audit pass. App clients SHALL connect via gateway-app external path `/ucg/app/ws/chat` (upgrade proxy to internal `/ws/chat`).

#### Scenario: WS 首帧认证
- **WHEN** 客户端经 gateway 连接后首帧 JSON 含合法 JWT
- **THEN** ucg-service SHALL 保持连接并注册 wxId；非法 JWT SHALL 关闭连接

#### Scenario: Redis 永久保留
- **WHEN** 消息审核通过并投递
- **THEN** 消息 SHALL 写入 Redis 键空间且 SHALL NOT 设置过期淘汰（MVP）

#### Scenario: 内部 WS 不经公网暴露
- **WHEN** 部署 ucg-service
- **THEN** `/ws/chat` MAY 仅集群内可达；App 对外入口 MUST 为 gateway `/ucg/app/ws/chat`

### Requirement: Conversation list SHALL support unread counts and pin/delete flags

API SHALL return conversations with unread_count, pinned, last message preview; List ordering SHALL use `pinned DESC, updated_at DESC` from `ucg_conversation_member`. `GET /ucg/app/api/conversations` MUST NOT 因单条会话的对方成员缺失或对方账号已注销而返回整页错误。 对每条当前用户未软删的会话，响应 MUST 仍包含 `id`、`unreadCount`、`lastPreview`、`pinned`、`updatedAt`。 当 `ucg_conversation_member` 中仍存在对方成员行时，列表项 `peerWxId` MUST 为该行的历史 `wx_id`。 当对方成员行不存在，或 device internal `ValidateWx` 表明该 `peerWxId` 对应 wx 不存在时，列表项 MUST NOT 填充.

#### Scenario: 未读计数

- **WHEN** 收件人收到新消息
- **THEN** 其 `unread_count` SHALL 递增直至调用 read API

#### Scenario: 对方成员行缺失时会话仍出现在列表

- **WHEN** 用户请求 `GET /ucg/app/api/conversations` 且某 direct 会话中除本人外无其他 `ucg_conversation_member` 行
- **THEN** 该会话 MUST 仍出现在 `list` 中且接口 MUST 返回成功
- **AND** 该项 `peerWxId` MAY 为 `0`
- **AND** `peerNickname` 与 avatar 相关字段 MUST 为空

#### Scenario: 对方已注销时保留 peerWxId 且展示为空

- **WHEN** 用户请求会话列表且某会话对方 `ucg_conversation_member` 行仍存在 `wx_id=W`，但 `ValidateWx(W)` 返回 `exists=false`
- **THEN** 该项 `peerWxId` MUST 为 `W`
- **AND** `peerNickname` 与 avatar 相关字段 MUST 为空
- **AND** 接口 MUST 返回成功

#### Scenario: 发消息路径仍要求有效对方

- **WHEN** 客户端经 WebSocket 向会话发送消息且对方成员不存在
- **THEN** 系统 MUST 拒绝发送且 MUST NOT 因本变更而静默成功

### Requirement: Outbound chat messages SHALL be processed via WebSocket handler

`ucg-service` WebSocket handler MUST 在收到 `type=message` 后：校验成员关系；向发送方发送 `message_ack`；**立即**写入 Redis 消息（`audit_status=pending`，`audit_version` 与 `ucg_chat_message` 权威列一致）、增加收件人未读、向收件方发送 `message_delivered`；Publish `ucg.chat.msg.created`（载荷含 `auditVersion`）。**MUST NOT** 在 handler 内同步调用 Green 阻塞投递。

#### Scenario: 发送方先收到 ack

- **WHEN** 客户端发送合法 message 帧
- **THEN** 发送方 MUST 先收到 `message_ack` 再收到异步审核结果

#### Scenario: 收件方先于审核收到消息

- **WHEN** 消息已写入 Redis
- **THEN** 收件方 MUST 收到 `message_delivered`，且 body MUST 含 pending 审态标识与 `audit_version`

### Requirement: WebSocket SHALL emit post-audit rejection events

Green 事后驳回时，系统 MUST：

- 向发送方 MUST 推送 `audit_failed`（含 `clientMsgId`、`reason`）
- 向在线收件方 MUST 推送 `msg_hidden`（含 `conversationId`、`messageId`）

仅当 CAS（`audit_status='pending' AND audit_version=?`）成功将消息置 rejected 后 MUST 推送上述事件；CAS 0 行（过期/重复消息）MUST NOT 推送。

#### Scenario: 在线收件人隐藏违规消息

- **WHEN** 消息 CAS 为 rejected 且收件方 WS 在线
- **THEN** 系统 MUST 向收件方推送 `msg_hidden`

#### Scenario: 重复 reject 消息不重复推送

- **WHEN** 同一 `auditVersion` 的 reject 事件重复消费且 CAS 影响 0 行
- **THEN** 系统 MUST NOT 再次推送 `audit_failed` 或 `msg_hidden`

### Requirement: Chat messages with video SHALL include mediaThumbnailUrl as physical thumb CDN URL

ucg-service 在 enrich 聊天消息媒体（含 HTTP 消息列表、WS 实时 `message_delivered` 及审核通过后推送）时，当消息含非空 `videoKey` 时 MUST 设置 `mediaThumbnailUrl` 为 `BuildVideoThumbnailURL(videoKey)`。该 URL MUST NOT 含 `x-oss-process`。`mediaCdnUrl` MUST 仍为视频原片 CDN URL。

本要求 MUST 在 Chat WebSocket v1 首版即满足，MUST NOT 推迟至后续 API 版本。MUST NOT 改变 v1 消息 JSON 字段结构，仅填充/修正 `mediaThumbnailUrl` 值。

#### Scenario: WS 实时视频消息含 mediaThumbnailUrl

- **WHEN** 收件方经 WS 收到含 `videoKey` 的 `message_delivered` 且 OSS 存在对应 `_thumb.jpg`
- **THEN** 消息 body SHALL 含非空 `mediaThumbnailUrl` 且 URL SHALL NOT 含 `x-oss-process`

#### Scenario: HTTP 消息历史与 WS 语义一致

- **WHEN** 客户端 `GET` 会话消息列表且行含 `videoKey`
- **THEN** 每项 SHALL 含与 WS 相同规则的 `mediaThumbnailUrl`

#### Scenario: 图片消息行为不变

- **WHEN** 消息仅含 `imageKey`
- **THEN** `mediaThumbnailUrl` SHALL 仍为 `BuildImageThumbnailURL(imageKey)` 物理缩略图 URL

---

### Requirement: UCG data SHALL live in ai_voice_ucg with defined post status enum

Database `ai_voice_ucg` SHALL contain tables: `ucg_profile`, `ucg_post`, `ucg_post_media`, `ucg_follow`, `ucg_post_like`, `ucg_post_comment`, `ucg_conversation`, `ucg_conversation_member`, and MAY contain `ucg_post_recommend`. Post `status` MUST use: 0=draft, 1=pending_audit, 2=published, 3=rejected.

#### Scenario: 创建待审核帖
- **WHEN** 用户提交帖子
- **THEN** `ucg_post.status` SHALL 为 1（pending_audit），且 SHALL NOT 为 2 直至 Green 通过

#### Scenario: 拒绝态记录原因
- **WHEN** Green 审核失败
- **THEN** `ucg_post.status` SHALL 为 3 且 `reject_reason` SHALL 非空

### Requirement: Timestamps SHALL use unix seconds

All `created_at`/`updated_at`/`published_at` columns MUST store unix seconds consistent with `database-unix-timestamp-storage` baseline.

#### Scenario: 写入创建时间
- **WHEN** 插入新 post
- **THEN** `created_at` SHALL 为 unix 秒级整数

### Requirement: Conversation member list SHALL be sortable by pin and last activity

`ucg_conversation_member` MUST include `updated_at` (unix seconds) maintained on new messages or pin changes; index `idx_wx_list (wx_id, pinned, updated_at)` SHALL support per-user conversation list ordering.

#### Scenario: 新消息刷新排序
- **WHEN** 会话成员收到审核通过的新消息
- **THEN** 各成员行的 `updated_at` SHALL 更新为当前 unix 秒

### Requirement: ucg_post SHALL include audit_version for CAS

表 `ucg_post` MUST 新增列 `audit_version INT NOT NULL DEFAULT 1`。新建 pending 帖 MUST 从 1 起始。用户再提审进入 pending 时 MUST `audit_version++`。消费者 CAS 审态变更成功 MUST NOT 递增该列。

#### Scenario: 新帖默认版本

- **WHEN** 用户 submit 创建帖子

- **THEN** 插入行 MUST 含 `audit_version=1` 且 `status=1`

#### Scenario: 再提审递增版本

- **WHEN** 作者对已发布或驳回帖 submit 再提审

- **THEN** 行 MUST `audit_version` 递增且 `status=1`；MUST NOT 回滚帖文正文等业务字段

### Requirement: ucg_post_comment SHALL include audit status columns

表 `ucg_post_comment` MUST 新增：

- `status TINYINT NOT NULL` — 0 draft（保留）、1 pending_audit、2 published、3 rejected

- `audit_version INT NOT NULL DEFAULT 1`

- `reject_reason VARCHAR`（可空）

首评 `audit_version=1`；再提审 MUST 递增。消费者 CAS MUST 使用 `status` + `audit_version`。

#### Scenario: 新评论 pending

- **WHEN** 用户发表评论

- **THEN** 插入 MUST `status=1`、`audit_version=1`

### Requirement: ucg_profile_audit_job SHALL store pending profile patches with audit_version

数据库 MUST 提供 `ucg_profile_audit_job`（或 design 等价命名）存储待审资料 patch，至少含：`wx_id`、`nickname`、`avatar_key`、`bio`、`status`（1/2/3）、**`audit_version`**、`reject_reason`、时间戳。

`audit_version` 为该 job 的 **唯一**审核轮次源；资料 CAS 与 MQ 载荷 MUST 读此列。MUST NOT 用 Redis 或 `ucg_profile` 表列作为审核版本。

#### Scenario: 资料提交创建 job

- **WHEN** 用户 PUT profile 变更

- **THEN** MUST 插入 pending job 行（`audit_version=1` 或再提审递增）而非仅写 Redis

### Requirement: ucg_chat_message SHALL include audit_status and audit_version

表 `ucg_chat_message` MUST 新增（或规范既有 `status` 字符串）：

- `audit_status`：`pending` | `approved` | `rejected`

- `audit_version INT NOT NULL DEFAULT 1`

- `reject_reason`（可空）

Redis 消息 JSON MUST 镜像 `audit_status` 与 `audit_version`，且与 MySQL 列语义一致。私信 CAS 与 MQ 载荷 MUST 以 `ucg_chat_message.audit_version` 为权威。

#### Scenario: 聊天消息初始 pending

- **WHEN** WS 发送消息并投递 Redis

- **THEN** `ucg_chat_message`（或等价 outbox 镜像）MUST `audit_status=pending` 且 `audit_version=1`；Redis JSON MUST 含相同 `audit_version`

---

### Requirement: device-service SHALL expose internal HTTP for ucg wx and baby name

device-service SHALL provide internal endpoints callable only with header `X-Device-Gateway-Internal-Secret` matching `DEVICE_GATEWAY_INTERNAL_SECRET`: validate wx id, batch fetch display fields, and get baby_name for default nickname. ucg-service MUST use these APIs and MUST NOT query `wx` table directly.

#### Scenario: 校验 wxId
- **WHEN** ucg-service 内部调用 validate 且 secret 正确
- **THEN** device-service SHALL 返回 wx 是否存在及必要展示字段

#### Scenario: 错误 secret 拒绝
- **WHEN** internal 请求 secret 不匹配
- **THEN** device-service SHALL 返回 403 且 SHALL NOT 返回 wx 数据

#### Scenario: ucg 禁止直连 device 库
- **WHEN** 代码评审发现 ucg-service import device DAO 查询 wx
- **THEN** 变更 MUST 拒绝合入

---

### Requirement: UCG 入场资格 API SHALL 按 device_no 返回连续有效喂养日结果

cash-service MUST 提供 `GET /cash/app/api/ucg/eligibility`，按网关注入的 `X-Internal-Device-No` 返回至少：`qualified`（bool）、`requiredDays`（固定 7）、`effectiveDays`（int）、`remainingDays`（int），并 MAY 返回可选文案字段。请求 MUST 经登录到达；`device_no` 为空时 MUST 返回错误，MUST NOT 伪造合格。该接口 MUST NOT 计入 App usage 统计（维护型 skip）。

#### Scenario: 连续满 7 有效日判定合格

- **WHEN** 某 `device_no` 按规则计算出的 `effectiveDays >= 7`
- **THEN** 响应 MUST 含 `qualified=true`、`requiredDays=7`、`remainingDays=0`

#### Scenario: 不足 7 日返回剩余天数

- **WHEN** `effectiveDays = 3`
- **THEN** 响应 MUST 含 `qualified=false`、`requiredDays=7`、`remainingDays=4`

#### Scenario: 缺少 device_no

- **WHEN** 请求未提供有效 `X-Internal-Device-No`
- **THEN** 系统 MUST 拒绝并返回错误，MUST NOT 返回 `qualified=true`

### Requirement: 有效喂养日 MUST 为上海日历日且连续统计且当日 history 不少于 10

有效喂养日判定 MUST 使用 `Asia/Shanghai`。某日有效当且仅当该 `device_no` 在该日内 history 记录数 ≥ 10。连续有效日 MUST 以请求当日为锚向前连续统计。资格计算 MUST 经 history-service HTTP 契约取数，MUST NOT 由 cash-service 直查 history 库。

#### Scenario: 单日满 10 条计为有效日

- **WHEN** 某上海日该设备 history 行数 = 10
- **THEN** 该日 MUST 计入有效喂养日

#### Scenario: 单日不足 10 条不计有效且中断连续

- **WHEN** 某上海日该设备 history 行数 = 9
- **THEN** 该日 MUST NOT 计入有效喂养日，且连续计数在该日中断

#### Scenario: 禁止跨库直查

- **WHEN** 实现资格计算
- **THEN** cash-service MUST 仅通过 history 内部 HTTP 获取按日计数，MUST NOT 连接 history 域数据库

### Requirement: UCG 入场资格 MUST NOT 受 VIP 或功能权益影响

资格计算与响应 MUST 仅基于喂养有效日规则。账号 VIP 状态、功能权益、邀请码/支付/广告开通 MUST NOT 将 `qualified` 置为 true 或缩短 `remainingDays`。客户端 MUST NOT 被本规格要求以 `isVip` 绕过入场门（服务端亦不提供 VIP 短路字段）。

#### Scenario: VIP 用户仍按喂养日计算

- **WHEN** 请求方账号为 VIP 但 `effectiveDays < 7`
- **THEN** 响应 MUST 为 `qualified=false`，且结果与非 VIP 在相同喂养数据下一致

### Requirement: 资格结果 MUST 按日缓存于 Redis 且不得落库

同一 `device_no` 在同一上海日内的资格结果 MUST 在缓存命中时直接返回，并 MUST 写入 Redis（经 `cachekit` + platform 键 builder）。MUST NOT 将资格结果作为权威数据写入 MySQL。MUST NOT 为此新增后台 ticker；miss 时在请求路径同步计算。history 调用失败时 MUST fail-closed（返回错误）。

#### Scenario: 同日第二次请求读缓存

- **WHEN** 某设备当日已成功计算并写入 Redis 后再次请求资格 API
- **THEN** 系统 MUST 返回与缓存一致的结果，且 MUST NOT 再次全量重算（除非缓存丢失）

#### Scenario: 资格不落 MySQL

- **WHEN** 资格计算完成
- **THEN** 系统 MUST NOT 将资格结果写入 MySQL 业务表

#### Scenario: Redis 经 platform

- **WHEN** 读写资格缓存
- **THEN** 代码 MUST 经 `cachekit`（含 Observer），键 MUST 来自 platform builder，MUST NOT 业务层直连 `g.Redis()`

---

### Requirement: 推荐 Feed 索引冷启动 SHALL 按需从 MySQL warm Redis

indexAutoWarmEnabled` 为 true 时，`ListRecommendFeed` MUST 在组装候选集之前执行有界 warm：分页读取 published 帖并调用与 `cmd/ucg-feed-backfill` 等价的 `syncPublishedPostRedis`（ZADD score、有坐标则 GEOADD、写 post/profile snapshot）。 warm 完成后 MUST 继续现有 GEO+ZSET+cursor 读路径，**MUST NOT** 改用 MySQL 排序降级。 **需要 warm 的判据** MUST 为：MySQL 存在 `status=published` 帖子，且（`ZCARD ucg:recommend:score` 为 0，**或** `publishedCount - zcard >=` 配置阈值 `indexHealGapThreshold`（默认 **50**））。 仅 `ZCARD > 0` 但相对 published **无明显短缺**时 MUST NOT.

#### Scenario: ZSET 空且 DB 有 published 帖

- **WHEN** `ZCARD ucg:recommend:score` 为 0 且 MySQL published 计数 > 0，且 auto warm 开启
- **THEN** 系统 MUST warm 至少一批帖至 Redis 并重试 Feed 候选收集，响应 SHOULD 含帖子（在 cap 覆盖范围内）

#### Scenario: ZSET 非空但相对 published 明显短缺

- **WHEN** `ZCARD` 为 20、published 为 500、缺口不低于阈值，且 auto warm 开启
- **THEN** 系统 MUST 触发有界 warm 补齐索引（受 `indexWarmMaxPosts` 限制），MUST NOT 因 ZCARD>0 而跳过

#### Scenario: 索引充足不 warm

- **WHEN** `ZCARD` 与 published 的差低于阈值（含大致齐平）
- **THEN** 系统 MUST NOT 触发有界 warm

#### Scenario: auto warm 关闭

- **WHEN** `indexAutoWarmEnabled=false`
- **THEN** 系统 MUST 保持空/短缺 ZSET 时不在请求路径 warm，依赖启动 heal 或运维 backfill

#### Scenario: 并发 Feed 请求

- **WHEN** 多个请求同时检测到需要 warm
- **THEN** 仅一个请求 MUST 持有 warm 锁执行灌库；其他请求 MUST 等待或短退避后读 ZCARD，不得无界阻塞

#### Scenario: warm 与 publish 写路径一致

- **WHEN** warm 处理某 published 帖
- **THEN** Redis MUST 写入 `ucg:recommend:score` 与（若有坐标）`ucg:feed:geo` 及 snapshot，语义与 `publishPostCAS` 后同步一致

---

### Requirement: ucg-service SHALL self-check and auto-heal recommend feed index on startup

`ucg-service` 进程启动时（依赖检查通过后、HTTP 服务可接受请求的路径上）MUST 执行一次推荐 Feed 索引自检（任务名 **FeedIndexStartupHeal** 的检测阶段），比较：

- MySQL `ucg_post` 中 `status=published` 的计数 `publishedCount`
- Redis `ZCARD` of `ucg:recommend:score`（经 cachekit / `UCGRecommendScoreKey`）得 `zcard`

当 `publishedCount > 0` 且（`zcard == 0` 或 `publishedCount - zcard >=` 配置阈值 `indexHealGapThreshold`，默认 **50**）时，系统 MUST 打 **ERROR** 级别可观测日志（含 `publishedCount`、`zcard`、gap）。 当自检开关开启且自动补齐开关开启时，系统 MUST **异步**（MUST NOT 阻塞 HTTP.

#### Scenario: 启动发现缺口并异步补齐

- **WHEN** ucg-service 启动且 publishedCount=500、zcard=10、阈值=50、check 与 heal 均开启
- **THEN** 系统 MUST 打 ERROR 日志含两计数，且 MUST 在后台尝试 heal 灌入索引（获锁成功时），HTTP 监听 MUST 仍可启动

#### Scenario: 无缺口不 heal

- **WHEN** publishedCount=100、zcard=100（或 gap 低于阈值）
- **THEN** 系统 MUST NOT 因启动自检触发 heal 灌库

#### Scenario: 仅检查不补齐

- **WHEN** check 开启且 heal 关闭，且存在缺口
- **THEN** 系统 MUST 打 ERROR 日志且 MUST NOT 自动 ZADD 补齐

#### Scenario: 多副本争锁

- **WHEN** 两副本同时启动且均检出缺口
- **THEN** 至多一个副本 MUST 持有 startup-heal 锁执行灌库；另一副本 MUST 跳过或等待锁策略后跳过，MUST NOT 无锁双全量狂写

#### Scenario: 自检关闭

- **WHEN** startup check 开关为 false
- **THEN** 启动 MUST NOT 执行缺口比较与 heal

---

### Requirement: viewer 有坐标时 unlimited 半径步 SHALL 从 ZSET 补全无 GEO 帖

当 viewer 请求含有效 lat/lng 且 GEO 半径阶梯已执行至 `radiusKm=0`（unlimited）时，Feed 候选收集 MUST 从 `ucg:recommend:score` ZSET 分页扫描成员并加入候选集，**MUST NOT** 在 unlimited 步仍仅使用 GEOSEARCH。ZSET 中 **不在** `ucg:feed:geo` 的 published 帖（无坐标）MUST 可进入 Feed，按 `baseScore` 参与 `finalScore`（`distanceTerm=0`），响应 MUST NOT 含 `distanceMeters`。

已在较小 GEO 半径（50–500km）或 pool 中的帖 MUST 由 pool/session 去重，不得重复下发。

#### Scenario: 帖仅在 ZSET 不在 GEO，viewer 带坐标

- **WHEN** published 帖在 `ucg:recommend:score` 有 member 但 `GEOPOS ucg:feed:geo` 为空，且 `GET /feed/recommend` 首屏含有效 lat/lng
- **THEN** 响应 `list` MUST 含该帖（在 pageSize 与候选 batch 覆盖范围内）

#### Scenario: unlimited 步不使用 GEO 替代 ZSET

- **WHEN** viewer 带坐标且当前半径阶梯为 `radiusKm=0`
- **THEN** 候选收集 MUST 使用 ZSET 读路径，MUST NOT 以 GEO 20000km 搜索作为 unlimited 步的唯一来源

#### Scenario: 无 viewer 坐标行为不变

- **WHEN** `GET /feed/recommend` 不含有效 lat/lng
- **THEN** 候选收集 MUST 仍按 ZSET/baseScore 排序，与变更前一致

---

### Requirement: 关注 Feed SHALL 经 MySQL 分页取 postId 并由 Redis snapshot 组装展示

关注 Feed 读路径 MUST 按下列顺序执行：

1. 无关注时 MUST 返回空 `list` 与 `total=0`。 从 MySQL `ucg_post` 查询 `status=2`（published）且 `author_wx_id IN (followees)` 的帖子，**MUST** 按 `published_at DESC` 排序，**MUST** 使用 `page`/`pageSize` 分页并返回准确 `total`。 DB 查询 MUST 仅获取组装所需 postId（及排序字段），**MUST NOT** 在关注 Feed 读路径 JOIN profile、like 或媒体表。 对当前页 postId 列表 MUST 调用与推荐 Feed 共用的 snapshot 组装逻辑（post snapshot、profile snapshot、liked SET batch），**MUST NOT** 使用 `postsFromResult` 全量 MySQL 组装。 snapshot miss MUST.

#### Scenario: 有关注时分页返回 published 帖

- **WHEN** 已登录用户 `GET /ucg/app/api/feed/following?page=1&pageSize=20` 且其关注的人中有 published 帖
- **THEN** 响应 MUST 含 `{ list, total, page, pageSize }`
- **AND** `list` 中每条 MUST 为 `status=2` 且作者 MUST 在当前用户 followee 集合内
- **AND** `list` 顺序 MUST 按 `published_at` 降序

#### Scenario: 组装走 Redis snapshot 而非帖表 N+1

- **WHEN** 关注 Feed 返回非空 `list` 且对应 post snapshot 已存在于 Redis
- **THEN** 系统 MUST 从 post/profile snapshot 与 liked SET 填充 item 字段
- **AND** MUST NOT 对该页帖子逐条 SELECT `ucg_post` 完整行用于展示组装

#### Scenario: snapshot miss 回源

- **WHEN** 关注 Feed 当前页某 postId 在 Redis 无 post snapshot
- **THEN** 系统 MUST 尝试 `backfillPostSnapshot` 回源并回填
- **AND** 回填成功时该帖 SHOULD 出现在 `list` 中

#### Scenario: 无关注时空列表

- **WHEN** 已登录用户未关注任何人
- **THEN** 响应 MUST 为 `{ list: [], total: 0, page, pageSize }`
- **AND** MUST NOT 查询 `ucg_post` 全表

#### Scenario: 不使用 Redis 关注 timeline

- **WHEN** 实现关注 Feed 读路径
- **THEN** 系统 MUST NOT 使用 ZUNIONSTORE 或 per-author 帖 ZSET 作为分页来源
- **AND** MUST NOT 新增 followees SET 或 author timeline 键

### Requirement: 关注 Feed SHALL 可选计算并返回 distanceMeters

当 viewer 请求含有效 `lat`/`lng` 且帖子 snapshot（或回源后）含坐标时，关注 Feed item MUST 含 `distanceMeters` 字符串（米，纯数字字符串）；否则 MUST 省略该字段。距离计算 MUST 复用推荐 Feed 的 haversine 逻辑，**MUST NOT** 将 post 坐标下发客户端。

#### Scenario: 关注 Feed item 含距离

- **WHEN** `GET /ucg/app/api/feed/following?lat=31.2&lng=121.5&page=1` 且某帖 snapshot 含坐标
- **THEN** 该帖 JSON MUST 含 `distanceMeters` 且值为米数字符串

#### Scenario: 无 viewer 坐标省略距离

- **WHEN** `GET /ucg/app/api/feed/following?page=1` 不含 lat/lng
- **THEN** 响应 item MUST NOT 含 `distanceMeters`

#### Scenario: 帖无坐标省略距离

- **WHEN** viewer 含 lat/lng 但帖无坐标
- **THEN** 该 item MUST NOT 含 `distanceMeters`

---

### Requirement: 原力首次增量可写入空账户

当 `ucg_user_force` 尚无该 `wx_id` 行时，系统执行原力增量（含获客 `invite_acquisition` +100 与辩论自投 +1）MUST 成功插入余额行并写入 `ucg_force_ledger`，MUST NOT 因空结果集 `Scan`/`sql.ErrNoRows` 失败。已有行时 MUST 在事务内累加并写流水。

#### Scenario: 首次获客加分

- **WHEN** cash 通知 ucg 对无原力行的码主人执行获客加分且入参合法
- **THEN** 该用户 `force_value` 为 100（或既有 0+100），且 ledger 存在 `invite_acquisition` 记录

#### Scenario: 已有余额再加分

- **WHEN** 用户已有 `force_value=100` 再次获客成功加分
- **THEN** `force_value` 变为 200 且新增一条 ledger

### Requirement: 我的资料返回原力值

`GET` 当前用户 profile（如 `/ucg/app/api/profile/me`）成功时 MUST 填充与本域 `ucg_user_force` 一致的 `forceValue`（无行视为 0），MUST NOT 因未 enrich 而长期省略真实非零原力。

#### Scenario: 有原力时 me 可见

- **WHEN** 用户 `ucg_user_force.force_value=100` 且请求我的资料
- **THEN** 响应 MUST 含 `forceValue` 为 100

---

### Requirement: 原力存储在 ucg 域

系统 MUST 在 ucg 数据库维护每用户原力当前值与积分流水。系统 MUST NOT 再以 device `wx.force_value` 作为原力权威。作者对自己辩论文章投票成功时，系统 MUST 原力 +1 并写入 reason=`debate_self_vote` 的流水。系统 MUST NOT 回填本变更上线前的历史投票流水。

#### Scenario: 辩论加分

- **WHEN** 作者自投成功
- **THEN** ucg 原力 +1 且 ledger 新增一条 `debate_self_vote`

### Requirement: 获客加原力

当邀请码被他人成功兑换一次时，系统 MUST 将码主人获客计数 +1，并经 cash→ucg 契约为码主人原力 +100，ledger reason=`invite_acquisition`。原力加分失败 MUST NOT 撤销已成功的功能开通；系统 MUST 记录可观测错误以便补偿。

#### Scenario: 兑码触发获客积分

- **WHEN** 用户 B 成功使用用户 A 的邀请码
- **THEN** A 的 `redeemed_count` +1，且 A 的原力尝试 +100 并在成功时写入 `invite_acquisition` 流水

### Requirement: App 可读原力与流水

系统 MUST 在 UCG App 可读路径提供当前 `forceValue`（可经 profile 或专用接口）与积分流水分页。系统 MUST NOT 要求返回「距下一档所需积分」；档位展示由客户端计算。

#### Scenario: 流水列表

- **WHEN** 用户请求 force ledger
- **THEN** 返回含 reason、delta、时间的记录，新产生的获客与辩论分均可见

---

### Requirement: Post and profile content SHALL use Green async audit Option A

Submitted posts (text/images/video) and profile fields (avatar/nickname/bio) SHALL enter pending visibility: ONLY author MAY see content in feeds/profile until Green pass sets published or profile active state; on fail content SHALL be rejected with reason visible to author as「违规已下架」或等价文案。 **触发方式 MUST 为事务提交后写入 `ucg_audit_publish_outbox` 并经 relay HTTP Publish RabbitMQ 事件（载荷含冻结 `auditVersion`），由 ucg-service AMQP consumer 异步执行 Green； MUST NOT 依赖定时扫 pending 业务表 worker 或.

#### Scenario: 发帖不再依赖 audit_worker 或 pending reconciler

- **WHEN** 用户 submit 发帖且 outbox 与 relay 正常
- **THEN** 系统 MUST Publish `ucg.post.created`（含 `auditVersion`）且 MUST NOT 依赖 `StartUcgAuditReconciler` 或 audit_worker 扫表触发首次审核

#### Scenario: Publish 失败由 outbox relay 恢复

- **WHEN** 帖子已 pending 且 outbox 行 Publish  initially 失败
- **THEN** relay worker MUST 重试 Publish；MUST NOT 通过扫描 `ucg_post.status=1` 补发

### Requirement: Green text rejection reason SHALL expose human-readable riskTips

当 Green **文本**机审判定违规（`Labels` 非空且非 `nonLabel`）时，系统 MUST 将 `AuditVerdict.Reason` 设为用户可读文案：**若 API `Data.Reason` 为 JSON 字符串，MUST 解析并取 `riskTips` 字段**；若 `Reason` 为纯文本，MUST 使用该纯文本；若解析后仍为空，MUST 使用默认文案「违规已下架」（或代码中等价常量）。系统 MUST NOT 将 `Data.Reason` 的原始 JSON 字符串直接写入 `reject_reason` 或经 App/WS 暴露给作者。

该规则 MUST 适用于所有经 `ModerateText` / `parseTextModeration` 的实体（帖子、评论、资料 job、私信）。图片/视频机审路径不在本 Requirement 范围内。

#### Scenario: Reason 为 JSON 时展示 riskTips

- **WHEN** Green 文本审核返回 `Labels` 命中且 `Data.Reason` 为 `{"riskTips":"命中违禁内容","riskLevel":"high"}`
- **THEN** 机审 consumer 写入的 `reject_reason`（及 WS `audit_failed` 的 reason）MUST 为 `命中违禁内容`，MUST NOT 为完整 JSON 字符串

#### Scenario: Reason 为纯文本时原样使用

- **WHEN** Green 返回 `Data.Reason` 为纯文本 `内容不合规`
- **THEN** `AuditVerdict.Reason` MUST 为 `内容不合规`

#### Scenario: JSON 无 riskTips 时回退默认文案

- **WHEN** Green 返回 `Data.Reason` 为 `{"riskLevel":"high"}` 且无有效 `riskTips`
- **THEN** `AuditVerdict.Reason` MUST 为「违规已下架」（或等价默认文案）

#### Scenario: 帖子作者可见修正后的驳回原因

- **WHEN** 用户帖子文本机审失败且 CAS 驳回成功
- **THEN** 作者「我的动态」接口返回的 `rejectReason` MUST 为上述可读文案，MUST NOT 含 JSON 花括号包裹的原始 Reason

### Requirement: Green image and video moderation dataId SHALL comply with Alibaba constraints

dataId`，该值 MUST 满足：长度不超过 64 字符； 系统 MUST NOT 将完整 HTTP(S) URL（含 scheme、`:`、`/` 等）作为 `dataId`。 若无法推导合规值，MUST 省略 `dataId` 字段而非传入非法值。 Code != 200` 或 HTTP 非 200 时，解析层 MUST 在 error 中包含 business `Code`（及 `Msg` 若 API 提供），以便运维区分参数错误与额度/限流等故障。

#### Scenario: 帖子图片审核使用合规 dataId

- **WHEN** 用户提交带 `social/` 前缀 objectKey 的图片帖且 Phase1 调用 `ModerateImageURL`
- **THEN** Green `ImageModeration` 请求的 `ServiceParameters` MUST NOT 含完整 CDN URL 作为 `dataId`；若含 `dataId` 则 MUST 为规范化 object path（如 `social_2026_06_xxx.jpg`）且长度 ≤64

#### Scenario: 资料头像与私信媒体同步约束

- **WHEN** 资料审核 job 调用头像 `ModerateImageURL`，或私信调用 `ModerateImageURL` / `ModerateVideoURL`
- **THEN** `dataId` 约束 MUST 与帖子媒体相同，MUST NOT 使用完整 URL

#### Scenario: Green 参数错误可观测

- **WHEN** Green 图片 API 因参数校验返回 `body.Code != 200`
- **THEN** ucg-service 日志中的 error MUST 包含 `green image: code <n>` 及 Msg（若存在），MUST NOT 仅返回无 code 的泛化文案

#### Scenario: 合规 dataId 下图片 Phase1 可完成

- **WHEN** 新发图片帖且 Green 配置有效、CDN URL 公网可访问
- **THEN** Phase1 MUST 成功发起 `baselineCheck` 并持久化 `moderation_verdict`（pass 或 reject），MUST NOT 因非法 `dataId` 直接进入 `moderation_failed`（status=5）

### Requirement: Profile patch SHALL use MySQL audit job and MQ instead of Redis pending queue

资料变更（nickname/avatar/bio）MUST 写入 MySQL `ucg_profile_audit_job` 并 Publish `ucg.profile.patch.submitted`。入队前 MUST 将请求字段与 **`ucg_profile` 已发布行**对比，**仅**将相对已发布值发生变化的非空字段写入 job；若 nickname、avatarKey、bio 均无实质变更，MUST NOT enqueue 且 MUST NOT 触发 Green。Consumer 机审时 SHOULD 跳过 job 中与已发布 profile 相同的字段（兼容历史全量 job 消息）。

#### Scenario: 仅 bio 变更

- **WHEN** 作者 PUT profile 携带未改 nickname 与新 bio
- **THEN** job MUST 仅含 bio 非空；Green MUST NOT 调用 `nickname_detection` 审该 nickname

#### Scenario: 全量相同无变更

- **WHEN** PUT 三字段均与已发布 profile 相同
- **THEN** MUST NOT 创建新 audit MQ 消息；HTTP MAY 返回 200 与当前作者 profile

### Requirement: Post status updates SHALL use CAS with audit_version

对 `ucg_post` 的机审与管理端审态变更（publish/reject）MUST 使用条件更新：`WHERE id=? AND status=? AND audit_version=?`。CAS 成功时 MUST 更新 `status`（及 `reject_reason` 等），**MUST NOT** 递增 `audit_version`。MUST NOT 存在无条件 `UPDATE status` 的机审路径。

#### Scenario: 机审通过 CAS

- **WHEN** consumer 审核 postId=1，载荷 `auditVersion=2`，且当前 `status=1`、`audit_version=2`
- **THEN** UPDATE MUST 使用 `status=1 AND audit_version=2` 条件；成功后 `status=2`，`audit_version` MUST 仍为 2

#### Scenario: 再提审递增版本

- **WHEN** 作者对已发布或驳回帖 submit 再提审
- **THEN** 系统 MUST 将 `status` 置 1 且 `audit_version` MUST 递增，并 Publish 新 `auditVersion`；帖文业务字段 MUST NOT 因再提审而 reset

#### Scenario: 过期 post 消息跳过

- **WHEN** 再提审后行内 `audit_version=3`，consumer 收到 `auditVersion=2` 的旧消息
- **THEN** CAS MUST 影响 0 行且 MUST ACK，MUST NOT 覆盖 version=3 的 pending 状态

### Requirement: Comments SHALL use Green async audit before public visibility

用户发表评论时 MUST 写入 `status=1`（pending_audit）与 `audit_version`（首评 1），MUST Publish `ucg.comment.created`（含 `auditVersion`）；Green 通过且 CAS（`WHERE status=1 AND audit_version=?`）成功后 status=2，评论 SHALL 出现在评论列表且 MAY 触发 `comment_count` 递增与通知；失败则 status=3，仅作者可见违规信息。CAS 成功 MUST NOT 递增 `audit_version`。

#### Scenario: 评论待审不对公众展示

- **WHEN** 用户发表评论且 Green 未完成
- **THEN** `GET .../comments` 响应 MUST NOT 含该条（其他用户视角）；作者 MAY 在响应或单独字段看到审核中

#### Scenario: 评论审核通过后计数

- **WHEN** Green pass 且 CAS 将评论置 published
- **THEN** 帖子 `comment_count` MUST 递增且 MAY 触发评论通知

### Requirement: Chat messages SHALL use post-delivery async Green audit (Mode A)

私信发送后 MUST 立即向收件人 WS 投递消息（`audit_status=pending`），MUST 写入 Redis 并增加收件人未读；MUST Publish `ucg.chat.msg.created`（含 `auditVersion`，权威为 `ucg_chat_message.audit_version`）异步 Green。pending 期间收件人 MUST 可见该消息。Green pass MUST CAS `audit_status` 从 pending 到 approved（`WHERE audit_status='pending' AND audit_version=?`）；Green fail MUST CAS 为 rejected，且 MUST 仅发送方可见并含违规提示，收件人 MUST NOT 在历史与列表中看到 rejected 消息。CAS 成功 MUST NOT 递增 `audit_version`。收件人未读在投递时 +1，reject MUST NOT 回滚未读。

#### Scenario: pending 期间收件人可见

- **WHEN** 用户发送聊天消息且 Green 未完成
- **THEN** 收件人 WS MUST 已收到 `message_delivered` 且拉取历史 MUST 含该条（非 rejected）

#### Scenario: 事后驳回仅发送方可见

- **WHEN** Green fail 且 CAS 为 rejected
- **THEN** 发送方 MUST 收到含 reason 的 `audit_failed`（或等价事件）且列表仍可见该条；收件方 MUST NOT 见该条，在线时 SHOULD 收到 `msg_hidden`

#### Scenario: 未读不回滚

- **WHEN** 消息已投递并已对收件人 unread+1，随后 Green reject
- **THEN** 收件人 unread_count MUST NOT 因 reject 减少

#### Scenario: 异步审核经 MQ

- **WHEN** 聊天消息已写入 Redis
- **THEN** 系统 MUST Publish `ucg.chat.msg.created`（含 `auditVersion`）且 MUST NOT 在 WS handler 内同步阻塞 Green

#### Scenario: 过期 chat 消息 CAS 跳过

- **WHEN** 消息已 CAS 为 approved 或 rejected，重复 MQ 消息携带旧 `auditVersion` 到达
- **THEN** CAS MUST 影响 0 行且 MUST ACK，MUST NOT 翻转审态

---

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

---

### Requirement: ucg internal SHALL batch query public profiles by wxIds

`ucg-service` MUST 提供 `POST /ucg/internal/api/profiles/batch`，鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。 请求体 MUST 含 `wxIds`（int64 数组，允许空）。 响应 MUST 含 `list` 数组； 对请求体中**每个有效 wxId（>0）** MUST 在 `list` 中返回**恰好一条**记录。 每项 MUST 含：

- `wxId`（uint64 或 int64）
- `nickname`（string，展示用昵称）
- `avatarUrl`（string，可选）
- `avatarThumbnailUrl`（string，可选）

当 wxId 存在 `ucg_profile` 行时，`nickname` MUST 与 `GetPublicProfilesByWxIDs` 语义一致（含默认昵称刷新逻辑）。 当 wxId **不存在**.

#### Scenario: Batch returns profile nickname for existing profile

- **WHEN** 受信调用方 POST `{ "wxIds": [1001] }` 且 1001 有 ucg_profile
- **THEN** 响应 `list` MUST 含 wxId=1001 的 nickname

#### Scenario: Batch synthesizes nickname without profile row

- **WHEN** 受信调用方 POST `{ "wxIds": [2002] }` 且 2002 无 ucg_profile 但 device 返回 babyName=`"小宝"`
- **THEN** 响应 `list` MUST 含 wxId=2002 且 nickname=`"小宝的家长"`

#### Scenario: Batch synthesizes default parent nickname

- **WHEN** 受信调用方 POST `{ "wxIds": [2003] }` 且 2003 无 ucg_profile 且 babyName 为空
- **THEN** 响应 `list` MUST 含 wxId=2003 且 nickname=`"家长"`

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 profiles batch
- **THEN** HTTP MUST 为 403

#### Scenario: Empty wxIds

- **WHEN** POST `{ "wxIds": [] }`
- **THEN** 响应 `list` MUST 为空数组且 MUST NOT 500

---

### Requirement: All app clients MUST upload media via server multipart API

UCG 媒体上传主路径 MUST 为服务端 multipart 上传（`POST /ucg/app/api/media/upload` 或经 gateway 同域代理的等价路径）。iOS/Android 客户端 MUST 与 Web 一样走该路径，MUST NOT 再依赖「预签名 URL + 客户端直传 OSS」作为主上传流程。

#### Scenario: 原生上传经服务端

- **WHEN** 原生 App 上传图片或视频且 resolve 未命中去重
- **THEN** 客户端 SHALL 将文件以 multipart 提交至服务端上传接口，SHALL NOT PUT 至 OSS 预签名 URL

### Requirement: Media presign MUST NOT issue client-usable upload URLs

`POST /ucg/app/api/media/presign`（若路由仍保留）MUST 拒绝签发可供客户端直传的预签名上传 URL（返回明确错误）。服务端 MUST NOT 向客户端返回指向 OSS 内网域名的上传地址。

#### Scenario: 调用 presign 被拒绝

- **WHEN** 客户端请求 media presign
- **THEN** 接口 SHALL 失败并提示改用服务端 upload，且 SHALL NOT 返回可用的 uploadUrl

### Requirement: Historical pang-bao objects are out of scope for dual-read

系统 MUST NOT 实现运行时回退读取旧桶 `pang-bao`。历史可见性依赖运维将对象以相同 objectKey 拷贝到 `caidoukeji` 且 CDN 已指向新桶。

#### Scenario: 同 objectKey 在新桶可读

- **WHEN** 数据库中已有 objectKey 且该对象已按相同 key 存在于 `caidoukeji`，CDN 指向新桶
- **THEN** 既有展示 URL（cdnBaseUrl + objectKey）SHALL 可访问，无需改库

---

### Requirement: UCG OSS MUST use caidoukeji bucket on Hangzhou internal endpoint

`ucg-service` 的 OSS 配置 MUST 使用 bucket `caidoukeji`、region `cn-hangzhou`。服务端 SDK endpoint MUST 为 `oss-cn-hangzhou-internal.aliyuncs.com`（不含 bucket 虚拟主机前缀）。AccessKey MUST 继续经 `UCG_OSS_ACCESS_KEY_*` 注入，yaml 明文 MUST 留空。展示 URL MUST 继续基于配置的 `cdnBaseUrl` 与 objectKey 拼接。系统 MUST NOT 再将 `pang-bao` 或 `oss-cn-beijing.aliyuncs.com` 作为默认生产配置。

#### Scenario: 默认配置指向新桶内网

- **WHEN** 使用仓库默认 `config.ucg-service.yaml` 且未设置 endpoint 覆盖 env
- **THEN** `LoadOSSConfig`（或等价）SHALL 得到 bucket=`caidoukeji` 且 endpoint 为杭州内网主机

#### Scenario: 非 VPC 可用 env 覆盖 endpoint

- **WHEN** 开发环境设置了约定的 OSS endpoint 覆盖变量为公网杭州 endpoint
- **THEN** 服务端 OSS 客户端 SHALL 使用该覆盖值连接（便于本地联调）

### Requirement: Server-side OSS operations MUST use configured endpoint only

服务端 `PutObject`、删除、缩略图、转码读写、校验下载等 MUST 使用上述配置的 endpoint/bucket，MUST NOT 硬编码旧桶名。

#### Scenario: 媒体服务端上传写入新桶

- **WHEN** 调用 `POST /ucg/app/api/media/upload`（经网关）成功上传文件
- **THEN** 对象 SHALL 写入 `caidoukeji`，响应 SHALL 含 objectKey 与基于 cdnBaseUrl 的 cdnUrl

---

### Requirement: OSS presign SHALL use pang-bao bucket with social/ prefix

ucg-service SHALL provide presigned upload for bucket `pang-bao`, region `cn-beijing`, endpoint `oss-cn-beijing. Database and API DTOs MUST store objectKey only; OSS AccessKey credentials MUST NOT be stored as plaintext in `manifest/config/config. accessKeySecret` MUST be empty in the repository. At runtime, credentials MUST be supplied via environment variables `UCG_OSS_ACCESS_KEY_ID` and `UCG_OSS_ACCESS_KEY_SECRET` (typically injected through `manifest/docker/.

#### Scenario: 获取 presign

- **WHEN** 客户端 `POST /ucg/app/api/media/presign` with media kind and extension，且 ucg-service 已通过 env 配置有效 OSS 凭证
- **THEN** 响应 SHALL 含 uploadUrl、objectKey（以 `social/` 开头），且 SHALL NOT 要求客户端自定义 bucket

#### Scenario: DB 仅存 objectKey

- **WHEN** 帖子媒体写入 `ucg_post_media`
- **THEN** 行 SHALL 仅保存 objectKey 字段，且 SHALL NOT 保存完整 CDN URL

#### Scenario: 未配置 OSS 凭证时 presign 失败

- **WHEN** 容器 env 与 yaml 均未提供有效 `UCG_OSS_ACCESS_KEY_*`
- **THEN** presign 接口 SHALL 返回明确错误且 SHALL NOT 使用硬编码或 yaml 明文 fallback

### Requirement: Image register and server upload SHALL ensure physical thumb exists

当 `RegisterMedia` 成功登记**新**图片 blob（原图已在 OSS）或 `putOSSObject` 成功上传图片（`mediaKind=1`）后，ucg-service MUST 同步调用 `EnsureImageThumb` 生成物理缩略图对象。`EnsureImageThumb` 失败时 register/直传 MUST 返回错误，MUST NOT 仅登记原图而无 thumb。

`PresignUpload` MUST NOT 生成缩略图（客户端尚未完成 PUT）。

dedup hit 路径 MAY 依赖已有 thumb；`EnsureImageThumb` 幂等调用 MUST 可接受。

#### Scenario: 新图 register 成功后 OSS 有成对对象

- **WHEN** 客户端完成原图 PUT 且 `RegisterMedia` 成功登记新图片 blob
- **THEN** OSS MUST 同时存在原图 objectKey 与对应 `ThumbObjectKey`

#### Scenario: 服务端直传图片后生成 thumb

- **WHEN** `UploadMediaObject` 成功上传 `mediaKind=1` 文件
- **THEN** OSS MUST 存在对应 `_thumb.{ext}` 对象

#### Scenario: thumb 生成失败阻止 register

- **WHEN** 原图存在但 `EnsureImageThumb` 失败
- **THEN** `RegisterMedia` SHALL 返回错误且 MUST NOT 仅完成 blob 登记而无 thumb

### Requirement: Video register and server upload SHALL ensure physical first-frame thumb exists

当 `RegisterMedia` 成功登记**新**视频 blob（原视频已在 OSS，`mediaKind=2`）或 `putOSSObject` 成功上传视频（`mediaKind=2`）后，ucg-service MUST 同步调用 `EnsureVideoThumb` 生成物理首帧缩略图对象（`{stem}_thumb.jpg`）。`EnsureVideoThumb` 失败时 register/直传 MUST 返回错误，MUST NOT 仅登记原视频而无 thumb。

`PresignUpload` MUST NOT 生成缩略图（客户端尚未完成 PUT）。

dedup hit 路径 MAY 依赖已有 thumb；`EnsureVideoThumb` 幂等调用 MUST 可接受。

#### Scenario: 新视频 register 成功后 OSS 有成对对象

- **WHEN** 客户端完成 mp4 PUT 且 `RegisterMedia` 成功登记新视频 blob
- **THEN** OSS MUST 同时存在原视频 objectKey 与对应 `{stem}_thumb.jpg`

#### Scenario: 服务端直传视频后生成 thumb

- **WHEN** 服务端成功上传 `mediaKind=2` 视频文件
- **THEN** OSS MUST 存在对应 `_thumb.jpg` 对象

#### Scenario: thumb 生成失败阻止 register

- **WHEN** 原视频存在但 `EnsureVideoThumb` 失败
- **THEN** `RegisterMedia` SHALL 返回错误且 MUST NOT 仅完成 blob 登记而无 thumb

### Requirement: Video media SHALL use transformVersion v1 or v2 with validation gate

视频 blob 索引 MUST 使用 `transformVersion` 区分管线产出：

- `v2`：canonical（H.264 Main + AAC + faststart）；原生客户端转码与 sim 服务端转码产出 MUST 使用 `v2` register
- `v1`：Web Phase 1 宽验真直传；**非 canonical**（可 Baseline、可无 faststart），但 **必须有 AAC 音轨**

`contentHash` MUST 为 OSS 上最终对象字节的 SHA-256 hex 小写。`v1` 与 `v2` MUST NOT 跨版本 dedup 命中（键为 `(contentHash, transformVersion)`）。

新写入 MUST NOT 使用 `sim-raw` 或其他未登记版本。

#### Scenario: v1 and v2 blobs do not dedup each other

- **WHEN** 同一源文件分别作为 v1 直传与 v2 转码后 contentHash 不同或 version 不同
- **THEN** `media/resolve` MUST NOT 跨 v1/v2 返回 hit

### Requirement: Web video upload response SHALL include contentHash for register

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 且上传成功时，响应 `data` MUST 含：

- `contentHash`：64 位小写 hex SHA-256，对 **OSS 上最终对象字节** 计算（v1 直传为原始字节；服务端转码为转码后字节）
- `transformVersion`：字符串 `v1` 或 `v2`，指示客户端 `RegisterMedia` 应使用的版本（v1 直传路径为 `v1`；服务端转码路径为 `v2`）

Web 客户端与 Flutter Web（经 gateway 同域 upload）MUST 使用响应中的 `transformVersion` 与 `contentHash` 配对 register；MUST NOT 在转码路径仍 register `v1`。

#### Scenario: Web direct upload returns v1 hint

- **WHEN** Web 成功上传 v1 合规视频（直传路径）
- **THEN** JSON `data` MUST 含 `contentHash`（长度 64）且 `transformVersion` MUST 为 `v1`

#### Scenario: Web transcode upload returns v2 hint

- **WHEN** Web 上传 v1 不合规但服务端转码成功
- **THEN** JSON `data` MUST 含 `contentHash` 且 `transformVersion` MUST 为 `v2`

#### Scenario: Gateway forwards transformVersion unchanged

- **WHEN** 请求经 gateway-app 反向代理至 ucg-service
- **THEN** 响应 JSON MUST 透传 `transformVersion` 字段（无裁剪）

---

### Requirement: Recommend feed SHALL use mixed ranking algorithm

Recommend feed SHALL rank published posts using mixed score combining new-post weight and engagement decay (likes/comments age decay). Implementation MAY persist scores in `ucg_post_recommend` refreshed by热区 reconciler（默认 1h tick）与冷区 MQ 互动重算。

对 **尚无** `ucg_post_recommend` 行的 published 帖，Feed MUST 在排序中优先于已有 score 的帖（置顶区），置顶区内 MUST 按 `published_at` 降序；已有 score 的帖 MUST 按 `score` 降序（同分则 `published_at`、`id` 降序）。该置顶 MUST 持续直至 reconciler 或冷区 MQ 路径首次写入 recommend 行。

#### Scenario: Unscored new post appears before scored posts

- **WHEN** 用户请求 `GET /ucg/app/api/feed/recommend` 且存在 published 帖尚无 `ucg_post_recommend` 行
- **THEN** 响应 `list` 中该类帖 MUST 排在已有 score 帖之前（同置顶区内按 `published_at` 新者优先）

#### Scenario: After reconciler scores post leaves pin tier

- **WHEN** reconciler 已为帖 UPSERT `ucg_post_recommend.score`
- **THEN** 该帖 MUST 按 score 与全站其他已算分帖一起排序，不再仅因「无 recommend 行」置顶

### Requirement: 推荐 Feed 读路径 SHALL 在索引冷启动时回填 Redis 再返回

在 `ucg-feed-geo-composite-score` 已采用的 Redis 复合分 Feed 读路径之上，当推荐索引冷启动（ZSET 空且 MySQL 有 published 帖）时，Feed 读路径 MUST NOT 直接返回空列表；MUST 先执行有界索引 warm（见 `ucg-feed-index-lazy-warm`），再继续 GEO/ZSET/cursor 分页。snapshot miss 的 `backfillPostSnapshot` 语义不变。

#### Scenario: 未 backfill 环境首次打开推荐

- **WHEN** 用户请求 `GET /ucg/app/api/feed/recommend` 且无 cursor，Redis 尚无 recommend score，MySQL 有 published 帖
- **THEN** 响应 `list` MUST NOT 因索引缺失而恒为空（在 warm cap 内应有帖）

#### Scenario: 已有索引行为不变

- **WHEN** Redis ZSET 已有成员
- **THEN** Feed 排序、cursor、distance 行为 MUST 与 `ucg-feed-geo-composite-score` 一致，且无额外 warm 开销

### Requirement: 推荐 Feed 对作者本人帖 SHALL omit distanceMeters

当已登录 viewer 请求 `GET /ucg/app/api/feed/recommend` 且请求含有效 `lat`/`lng`、帖子 snapshot 含坐标时，若帖子 `authorWxId` 等于 viewer 的 `wxId`，该 item **MUST NOT** 含 JSON 字段 `distanceMeters`（即使 viewer 与帖坐标可计算 haversine）。本要求 **MUST NOT** 改变该帖的 `finalScore` 或 composite 排序语义。

#### Scenario: 本人帖 omit 距离

- **WHEN** 已登录用户 `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且列表含其本人发布的 published 帖（帖含坐标）
- **THEN** 该 item JSON MUST NOT 含 `distanceMeters`
- **AND** 同页他人帖在 viewer 与帖均有坐标时 MUST 仍含 `distanceMeters`

#### Scenario: 本人帖排序不变

- **WHEN** viewer 本人帖与他人帖参与同一推荐页 composite 排序
- **THEN** 系统 MUST 仍按既有 `finalScore = baseScore + distanceTerm` 排序
- **AND** omit `distanceMeters` MUST NOT 单独改变该帖的 `finalScore` 计算

#### Scenario: 未登录无本人概念

- **WHEN** 匿名或未带 `X-Internal-Wx-Id` 的有效登录上下文请求推荐 Feed
- **THEN** 距离字段行为 MUST 与变更前一致（有坐标则按他人帖规则返回 `distanceMeters`）

---

### Requirement: 推荐分更新事件 SHALL 使用已注册的 ucg.recommend routing keys

`ucg-service` 在推荐分需更新时 MUST Publish 下列 routing key 之一（经 `eventkit` 注册校验）：

- `ucg.post.published`
- `ucg.post.unpublished`
- `ucg.post.liked`
- `ucg.post.unliked`
- `ucg.comment.published`
- `ucg.comment.removed`

载荷 MUST 至少含 `postId`；comment 类事件 MUST 含 `commentId`。

#### Scenario: 审核通过后发布推荐更新

- **WHEN** 帖子 Green 审核 CAS 成功且 `status=published`
- **THEN** 系统 MUST Publish `ucg.post.published`，载荷含 `postId`

#### Scenario: 删帖或下架统一 unpublish

- **WHEN** 作者删除帖子或帖子从 published 变为 rejected/unpublished
- **THEN** 系统 MUST Publish `ucg.post.unpublished`，载荷含 `postId`

### Requirement: RabbitMQ 拓扑 SHALL 为推荐分队列绑定 topic exchange

仓库 `hack/rabbitmq-init` MUST 声明 durable 队列 `ucg.recommend.score.q` 并与 `voice.events` exchange 完成 binding（覆盖上述 routing key 或 `ucg.recommend.#`）。

#### Scenario: init 后队列可见

- **WHEN** 运维执行 rabbitmq init
- **THEN** 管理台 SHALL 可见 `ucg.recommend.score.q` 且 binding 正确

### Requirement: 推荐分 AMQP consumer MUST 驻留 ucg-service 且 SHALL 单帖更新

`ucg-service` MUST 经 AMQP push（`autoAck=false`）消费 `ucg. removed` → 若帖处于**冷区**（`published_at` 早于 `now - hotWindowHours`，与热区 reconciler 窗口一致），MUST `RecomputeRecommendScore(postId)` UPSERT `ucg_post_recommend`； 若帖处于**热区**，MUST Ack 跳过本次重算（不得 Nack 重试）。 - `published` → MUST Ack 跳过且不 UPSERT（新帖 score 由热区 reconciler 写入； MUST NOT 在 consumer 内对全部 published 帖做无 LIMIT 全表扫描。

#### Scenario: Cold zone like triggers recompute

- **WHEN** consumer 收到 `ucg.post.liked` 且 `postId` 对应帖 `published_at` 早于热区 cutoff
- **THEN** 系统 MUST 执行 `RecomputeRecommendScore` 且 MUST Ack

#### Scenario: Hot zone like skips recompute

- **WHEN** consumer 收到 `ucg.post.liked` 且帖 `published_at` 在热区窗口内
- **THEN** 系统 MUST NOT 调用 `RecomputeRecommendScore` 且 MUST Ack

#### Scenario: unpublished 删除推荐行且永远 Ack

- **WHEN** consumer 收到 `ucg.post.unpublished` 且 `postId` 合法
- **THEN** 系统 MUST 执行 DELETE；`RowsAffected=0` MUST 视为成功且 MUST NOT 报错；处理完成后 MUST Ack

### Requirement: like 类事件 throttle MUST 仅限制重算频率且 SHALL 允许短期误差

对 **冷区** `ucg.post.liked`、`ucg.post.unliked`、`ucg.comment.published`、`ucg.comment.removed`，consumer MUST 对 `postId` 使用 Redis 单 key `SET NX EX` throttle（默认 500ms）。窗口内 NX 失败 MUST 跳过本次重算并 Ack。throttle MUST 仅保证 500ms 内最多 1 次 `RecomputeRecommendScore`；MUST NOT 保证反映每一次 like/unlike 方向变化。Publisher MUST NOT 在发送侧合并事件。

热区事件因跳过 Recompute，MUST NOT 依赖 throttle 收敛（由 reconciler 负责）。

#### Scenario: Cold zone 500ms 内多次 like 只重算一次

- **WHEN** 冷区帖 `postId` 在 500ms 内收到 3 条 `ucg.post.liked`
- **THEN** consumer MUST 最多执行 1 次 `RecomputeRecommendScore`，且 3 条消息 MUST 均 Ack

### Requirement: 热区 reconciler MUST 分页增量且轮首固定 hotCutoff

系统 MUST 使用 MySQL 表 `ucg_recommend_hot_scan_state`（含 `last_post_id` 与 `round_hot_cutoff`）驱动热区分页 reconciler。当 `last_post_id=0` 开始新轮时 MUST 计算并持久化 `round_hot_cutoff`；同一轮后续分页 MUST 使用已存 `round_hot_cutoff` 且 MUST NOT 在分页过程中用 `NOW()` 重新计算 cutoff。每 tick MUST 使用 `published_at >= round_hot_cutoff AND id > last_post_id ORDER BY id LIMIT pageSize`；MUST NOT 一次加载全部 published 帖。

默认 tick 间隔 MUST 为 **3600s**（1 小时），可由 env `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS` 覆盖。

#### Scenario: 热区一轮扫完重置 cursor

- **WHEN** 某 tick 返回行数小于 `pageSize`
- **THEN** reconciler MUST 将 `last_post_id` 置 0；下轮开始 MUST 重新计算 `round_hot_cutoff`

#### Scenario: Default interval one hour

- **WHEN** 未设置 `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS` 且 yaml 为变更后默认值
- **THEN** reconciler 日志 MUST 显示 interval 为 1h 量级

### Requirement: 热区 reconciler MUST 周期性重算即使无互动

在冷区零兜底前提下，热区 reconciler MUST 对每个扫到的 published 热区帖执行 `RecomputeRecommendScore`，即使该帖在扫描周期内无任何 like/comment 事件。该行为 MUST 用于热区时间衰减与热区互动改由 reconciler 收敛后的 score 更新。

#### Scenario: 无互动热区帖仍更新 score

- **WHEN** 热区 reconciler 扫描到 `postId` 且该帖在上一扫描周期内 like/comment 计数未变
- **THEN** 系统 MUST 仍执行 `RecomputeRecommendScore(postId)` 以反映 `exp(-age/τ)` 变化

### Requirement: 冷区 MUST NOT 运行分页 reconciler

`published_at < round_hot_cutoff`（冷区）的帖子 MUST 通过 MQ 互动事件更新推荐分（见上文冷区分流）；系统 MUST NOT 为冷区启动定时全量或分页扫表任务。

#### Scenario: 冷区帖靠互动翻红

- **WHEN** 冷区帖收到 `ucg.post.liked` 且 throttle 允许重算
- **THEN** 系统 MUST 更新该帖 `ucg_post_recommend.score` 且 Feed MUST 可将该帖按 score 前排展示

### Requirement: ucg-service AMQP audit 与 recommend consumer SHALL 共用单 connection

`ucg-service` 内 UCG 审核队列 consumer 与推荐分 consumer MUST 共用 **一条** AMQP connection；每个消费队列 MUST 使用 **独立 channel** 与独立 prefetch 配置。连接断线 MUST 统一 backoff 重连并恢复全部 channel Consume。

#### Scenario: 单 connection 多 channel

- **WHEN** ucg-service 启动且 audit 与 recommend consumer 均 enabled
- **THEN** 进程 MUST 仅建立 1 条到 RabbitMQ 5672 的 AMQP connection，且 audit 4 队列与 recommend 1 队列各占用独立 channel

### Requirement: publishPostCAS 成功路径 SHALL 同步 Redis 推荐索引

帖子 Green 审核 CAS 成功变为 published 时，写路径 MUST 同步：

1. ZADD `ucg:recommend:score`（initial baseScore）
2. 若帖含 lat/lng → GEOADD `ucg:feed:geo`
3. SET post snapshot 与 author profile snapshot

MUST NOT 写入 `ucg_post_recommend`。

#### Scenario: 过审发帖同步 GEO 与 snapshot

- **WHEN** 帖 publish 成功且含 lat/lng
- **THEN** 系统 MUST GEOADD、`ZADD` 且写入 post/profile snapshot 于同一写路径事务语义内（best-effort 顺序：MySQL commit 后 Redis）

### Requirement: ucg-service 热区 reconciler SHALL 独立于 recommend MQ consumer 开关

`StartRecommendHotReconciler` MUST 在 `ucg-service` 启动时运行（与 `UCG_RECOMMEND_MQ_CONSUMER_ENABLED` 无关）。`UCG_RECOMMEND_MQ_CONSUMER_ENABLED=false` MUST 仅停止订阅 `ucg.recommend.score.q`，MUST NOT 停止热区 reconciler。

#### Scenario: Consumer disabled reconciler still runs

- **WHEN** `UCG_RECOMMEND_MQ_CONSUMER_ENABLED=false` 且进程启动
- **THEN** 日志 MUST 含 `[ucg-recommend-hot] started` 且 MUST NOT 订阅 recommend 队列

### Requirement: 发帖过审 MUST NOT 发布 recommend post.published 事件

帖 `pending_audit → published` 成功后，系统 MUST NOT 调用 `PublishPostPublished` / 向 `ucg.recommend.score.q` 发送 `published` 路由事件。该帖首次 score MUST 由热区 reconciler 写入。

#### Scenario: Publish CAS does not emit recommend published

- **WHEN** `publishPostCAS` 成功将帖设为 published
- **THEN** MUST NOT 向 recommend MQ 发送 `post.published` 事件

---

### Requirement: ucg internal chat send SHALL allow simulated users only

`ucg-service` MUST 提供 `POST /ucg/internal/api/chat/send`，要求有效内部网关密钥（与 device ucg internal 一致）。 请求体 MUST 含 `senderWxId`、`conversationId`、`clientMsgId`，以及 `content` 或 `imageKey`/`videoKey` 之一（互斥规则与 App WS 一致）。 发送前 MUST 经 device internal 确认 `senderWxId` 对应 `is_simulated=1`。 否则 MUST 返回 403。 发送方 MUST 为会话成员。 成功 MUST 调用 `ProcessOutboundChatMessage`（含 Green 异步审核与真人 peer push）。 **消息投递成功后**，MUST 对 `senderWxId` 调用 `MarkConversationRead`，将其在该会话的 `unread_count` 置 0（T5 未读闭环）；

#### Scenario: Sim user sends message

- **WHEN** `senderWxId` 为 sim 且为会话成员且 content 非空
- **THEN** 消息 MUST 持久化并进入正常聊天审核流程

#### Scenario: Sim sender marked read after send

- **WHEN** internal send 成功且发送前 sim 侧 `unread_count > 0`
- **THEN** 发送完成后 sim 侧 `unread_count` MUST 为 0

#### Scenario: Real user rejected

- **WHEN** `senderWxId` 的 `is_simulated=0`
- **THEN** API MUST 返回 403 且 MUST NOT 发送消息

#### Scenario: Not a member

- **WHEN** sender 非 conversation 成员
- **THEN** API MUST 返回 404 或等价业务错误

### Requirement: sim-user-service MUST NOT use WebSocket for outbound chat

模拟用户出站聊天 MUST 经上述 internal HTTP 契约；sim-user-service MUST NOT 建立 `/ucg/app/ws/chat` 长连接。

#### Scenario: No WS from sim service

- **WHEN** sim 用户回复私聊
- **THEN** 实现 MUST 使用 `POST /ucg/internal/api/chat/send` 而非 WebSocket

---

### Requirement: ucg-service SHALL expose internal sim unread conversation sample API

系统 MUST 提供 internal 接口 `POST /ucg/internal/api/chat/sim-unread-sample`，供 `sim-user-service` 等受信内部调用方抽取 **一条** eligible 未读会话样本。 鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。 请求体 MUST 含 `simWxIds`（`int64` 数组，非空）。 deleted_at = 0`

响应 MUST NOT 调用 App 层 author 富化或 recommend 路径。 命中时 MUST 返回 `conversationId`、`simWxId`、`peerWxId`、`unreadCount` 及 `messages[]`（最近消息，元素含 `senderWxId`、`content`）。 无 eligible 时 MUST 返回 `found=false`（或等价空结果）与 HTTP 200（code=0），MUST.

#### Scenario: Sample returns one eligible unread conversation

- **WHEN** 存在 sim 用户 S∈simWxIds 对真人 P 的未读会话，且请求含完整 simWxIds
- **THEN** 响应 MUST 含 `found=true`、`simWxId=S`、`peerWxId=P`、`conversationId>0` 及非空 `messages`

#### Scenario: No eligible unread

- **WHEN** simWxIds 非空但无任何 eligible 未读会话
- **THEN** 响应 MUST 含 `found=false` 且 MUST NOT 返回会话 id

#### Scenario: Invalid secret

- **WHEN** internal 密钥缺失或错误
- **THEN** MUST 返回 403 且 MUST NOT 查询业务表

#### Scenario: Sim-sim unread excluded

- **WHEN** 仅存在 sim↔sim 双方均在 simWxIds 内的未读会话
- **THEN** 响应 MUST 为 `found=false`

### Requirement: sim unread sample MUST use bounded ID probe without retry loop

抽样 MUST 在 eligible 集合上对 `ucg_conversation_member.id` 使用 **MIN/MAX + 均匀锚点 + `id >= anchor ORDER BY id LIMIT 1`**（空洞回退 minId），固定 **2 条**有界 SQL。MUST NOT 使用 `ORDER BY RAND()`。MUST NOT 对 `simWxIds` 内单个 wx 循环重试或分页扫描会话列表。

消息加载 MUST 为单条有界查询（`LIMIT messageLimit`），MUST NOT 调用 device HTTP。

#### Scenario: Bounded SQL on eligible subset

- **WHEN** 代码评审 sample 实现
- **THEN** MUST 为 eligible 子集上的 MIN/MAX + LIMIT 1 探测，MUST NOT 全表 `ORDER BY RAND()`

#### Scenario: No cross-domain DAO in ucg

- **WHEN** 代码评审 ucg internal sample 实现
- **THEN** MUST NOT import device 域 DAO；sim 身份过滤 MUST 依赖请求体 `simWxIds`

#### Scenario: Message limit clamp

- **WHEN** 请求 `messageLimit` 为 100
- **THEN** 实际返回 MUST 最多 50 条消息

---

### Requirement: ucg-service SHALL expose internal published post sample API for sim workloads

系统 MUST 提供 internal 接口 `POST /ucg/internal/api/posts/sample`，供 `sim-user-service` 等受信内部调用方抽取已发布帖子样本。 鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。 响应 MUST 仅包含评论任务所需最小字段，MUST NOT 触发 recommend Feed 的 author/media 富化或 device HTTP 调用。 每条样本 MUST 含 `postId`、`content`、`mediaType`。 图文帖（`mediaType=1`）且存在首图 `coverObjectKey` 时 MUST 含 `coverCdnUrl`（全分辨率 CDN URL）。 视频帖（`mediaType=2`）且存在封面 media key 时 MUST 含 `coverCdnUrl`（视频首帧 snapshot URL，MUST NOT 为 mp4 直链）。 LLM 输入.

#### Scenario: Image post sample includes cover CDN URL

- **WHEN** sample 返回 `mediaType=1` 且首条 media 有 objectKey
- **THEN** 响应项 MUST 含非空 `coverCdnUrl`，且 MUST 由 ucg OSS CDN 配置拼装

#### Scenario: Video post sample includes snapshot URL

- **WHEN** sample 返回 `mediaType=2` 且首条 media 有 objectKey
- **THEN** `coverCdnUrl` MUST 为首帧 snapshot URL（含 `x-oss-process=video/snapshot` 语义），MUST NOT 为原始视频文件 URL

#### Scenario: Text-only post omits cover CDN URL

- **WHEN** sample 返回 `mediaType=0`
- **THEN** `coverCdnUrl` MUST 为空或省略

#### Scenario: Invalid or missing secret

- **WHEN** internal 密钥缺失或错误
- **THEN** MUST 返回 403 且 MUST NOT 查询业务表

#### Scenario: Empty plaza

- **WHEN** 无 `status=published` 帖子
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

### Requirement: sample API MUST use bounded single-query read pattern

抽样读路径 MUST 在 ucg 库内完成，MUST NOT 调用 `postsFromResult`、`GetPublicProfile` 或 `Device(). `limit` 默认 20，超出 MUST 截断为 50； - **random 模式**：MUST 使用有界读模式：一次 `MIN(id)/MAX(id)` 聚合（`WHERE status=published` 及可选 author 排除）加一次 `id >= R LIMIT 1` 探测（共 2 次 SQL）； 每次 MUST 带 `LIMIT 1` 或聚合有界，MUST NOT 全表加载。 - 当请求 body 含非空 `excludeAuthorWxIds` 时，latest 与 random 模式的 published 帖查询 MUST 附加 `author_wx_id NOT IN (.

#### Scenario: Limit clamp in latest mode

- **WHEN** 请求 `limit` 为 100 且未指定 random 模式
- **THEN** 实际查询 MUST 最多返回 50 条

#### Scenario: Random mode bounded queries

- **WHEN** 请求 `mode=random`
- **THEN** 读路径 MUST 最多 2 次 SQL且最终 MUST NOT 返回超过 1 条

#### Scenario: Exclude sim authors in random mode

- **WHEN** 请求 `mode=random` 且 `excludeAuthorWxIds` 含全部 sim wxId，且存在至少一条 published 帖其 author 不在该集合
- **THEN** 响应 `list[0].authorWxId` MUST 不在 `excludeAuthorWxIds` 中

#### Scenario: No cross-domain DAO

- **WHEN** 代码评审 ucg internal sample 实现（含 random 分支）
- **THEN** MUST NOT import device 域 DAO 或直连 device 库表

### Requirement: sample API SHALL support random mode via ID probe with mild recency bias

当请求 body `mode` 为 `random` 时，系统 MUST 在 ucg 库内通过有界 ID 探测返回 **0 或 1** 条满足 **status 与 excludeMediaTypes（若提供）** 条件的已发布帖，MUST NOT 使用 `ORDER BY RAND()`。探测 MUST 在过滤后的 published 集合上覆盖全库（非仅最新 N 条）。锚点 MUST 在 eligible 帖的 `[minId, maxId]` 上按 `R = minId + floor((maxId - minId) * U^α)` 生成（`U` 均匀随机，默认 `α = 0.65`），随后 `WHERE … AND id >= R ORDER BY id ASC LIMIT 1`。响应字段 MUST 含 `postId`、`content`、`mediaType`、可选 `coverObjectKey`/`coverCdnUrl`。

#### Scenario: Random mode returns one published post

- **WHEN** 有效 internal 密钥、body `{ "mode": "random" }`（或无 exclude），且存在 eligible published 帖
- **THEN** 响应 `list` MUST 含 1 条 published 帖

#### Scenario: Random mode empty plaza

- **WHEN** 无 eligible published 帖（含 exclude 后为空）
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Latest mode unchanged without exclude

- **WHEN** `mode=latest`（或缺省）、`limit=20`、无 `excludeMediaTypes`
- **THEN** MUST 仍按 `published_at DESC` 返回最多 20 条

### Requirement: sample API SHALL support excludeMediaTypes filter

`POST /ucg/internal/api/posts/sample` 请求 body MAY 含 `excludeMediaTypes`（整型数组）。当数组非空时，抽样 MUST 仅返回 `media_type` **不在**该集合中的已发布帖。`mode=random` 的 ID 探测（MIN/MAX 与 `id>=R` probe）MUST 应用与列表查询相同的 `excludeMediaTypes` filter。未传或空数组时 MUST 保持变更前行为（含视频帖）。

#### Scenario: Random sample excludes video

- **WHEN** 请求为 `{ "mode": "random", "excludeMediaTypes": [2] }` 且存在 published 非视频帖
- **THEN** 返回的每条样本 `mediaType` MUST NOT 为 `2`

#### Scenario: Random bounds respect exclude filter

- **WHEN** 请求 `mode=random` 且 `excludeMediaTypes` 为 `[2]`
- **THEN** MIN/MAX id 聚合 MUST 仅在 `media_type NOT IN (2)` 的 published 帖上计算

#### Scenario: No eligible posts after exclude

- **WHEN** 所有 published 帖均为 `media_type=2` 且 `excludeMediaTypes` 含 `2`
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Backward compatible without exclude

- **WHEN** 请求未含 `excludeMediaTypes`
- **THEN** random/latest 行为 MUST 与引入本字段前一致

### Requirement: sample API response SHALL include authorWxId for internal consumers

`POST /ucg/internal/api/posts/sample` 响应 `list` 每项 MUST 含 `authorWxId`（`ucg_post.author_wx_id`）。T2 等既有调用方 MAY 忽略该字段。

#### Scenario: Author field present

- **WHEN** sample 返回非空 `list`
- **THEN** 每条 MUST 含 `authorWxId` 且 MUST 大于 0

### Requirement: sample API SHALL support excludeDebate and onlyDebate filters

二者 MUST NOT 同时为 true。 判定 MUST 与 UCG `isDebatePost` 一致：`(debate_left_label != '' AND debate_right_label != '') OR type = 'debate'`。 - `excludeDebate=true`：MUST 仅返回非辩论帖； random 的 MIN/MAX 与 probe MUST 应用相同 filter。 - `onlyDebate=true`：MUST 仅返回辩论帖； random 的 MIN/MAX 与 probe MUST 应用相同 filter。 - 均未传或为 false：MUST 保持变更前行为（含辩论帖）。 响应 `list` 每项 MUST 含 `debateLeft`、`debateRight`（字符串； 非辩论帖 MUST 为空字符串或省略）。 T8 MUST 使用。

#### Scenario: T2 exclude debate in SQL

- **WHEN** 请求 `{ "mode": "random", "excludeDebate": true }` 且存在 published 非辩论帖
- **THEN** 返回样本 MUST NOT 为辩论帖

#### Scenario: T8 only debate in SQL

- **WHEN** 请求 `{ "mode": "random", "onlyDebate": true }` 且存在 published 辩论帖
- **THEN** 返回样本 MUST 为辩论帖且 MUST 含非空 `debateLeft` 与 `debateRight`

#### Scenario: Mutually exclusive flags rejected

- **WHEN** 请求同时 `excludeDebate=true` 且 `onlyDebate=true`
- **THEN** MUST 返回 400 且 MUST NOT 查询

#### Scenario: Empty when only debate and none exist

- **WHEN** `onlyDebate=true` 且无 published 辩论帖
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Debate fields on debate sample

- **WHEN** sample 返回辩论帖
- **THEN** `debateLeft` 与 `debateRight` MUST 各 ≤5 字且非空

---

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

---

### Requirement: Internal upload-video SHALL validate then PUT without transcode

ucg-service MUST 注册 `POST /ucg/internal/api/media/upload-video`，接受 multipart 视频文件，鉴权 MUST 与现有 `POST /ucg/internal/api/media/upload` 一致（内部网关密钥）。

处理流程 MUST 为：读取 body（上限与 `MaxMediaUploadBytes` 一致）→ 对字节执行 **v1** 验真 → 通过则 PUT OSS（`video/mp4`）→ 响应 `objectKey`、`cdnUrl`、`contentHash`（SHA-256 hex 小写，对 **OSS 上最终字节** 即原始合规字节计算）、`transformVersion=v1`。

v1 验真失败 MUST 返回 4xx 且 MUST NOT 在 OSS 创建对象。MUST NOT 调用 `NormalizeVideo` 或任何进程内转码。本接口 MUST NOT 自动 `RegisterMedia`。

#### Scenario: Internal upload returns v1 object

- **WHEN** 内部密钥有效且上传通过 v1 验真的 mp4
- **THEN** 响应 MUST 含 objectKey、contentHash 与 `transformVersion=v1`，且 OSS 对象字节 MUST 与上传字节一致

#### Scenario: Internal upload unauthorized

- **WHEN** 未提供有效内部密钥
- **THEN** MUST 返回 403

#### Scenario: Internal upload rejects non-compliant video

- **WHEN** 上传未通过 v1 验真（含不可解码）
- **THEN** MUST 返回 4xx 且 MUST NOT 留下 OSS 对象，MUST NOT 转码

### Requirement: ucg-service container SHALL include ffprobe for validation

`manifest/docker/Dockerfile.ucg-service` 构建的镜像 MUST 包含可执行的 `ffprobe`，供视频验真使用。部署 ucg-service MUST 使用含 ffprobe 的镜像。镜像 MAY 仍包含 `ffmpeg` 二进制，但上传路径 MUST NOT 依赖其做转码。

#### Scenario: Probe available in container

- **WHEN** ucg-service 容器启动
- **THEN** 进程 MUST 能成功执行 `ffprobe -version`

---

### Requirement: ucg-service SHALL validate video by transformVersion with forked rules

ucg-service MUST 提供按 `transformVersion` 分叉的视频 ffprobe 验真能力。 **v2（canonical）** MUST 满足：

- 容器 format 为 mp4
- 视频轨 codec 为 h264，profile 为 **Main**，pix_fmt 为 yuv420p
- **必须有音轨且 codec 为 aac**（静音 AAC 合法）
- **faststart**：moov MUST 位于 mdat 之前（可播放渐进下载）
- 大小 ≤ ucg-service 配置的单文件上传上限

**v1（Web Phase 1 宽规）** MUST 满足：

- 容器 format 为 mp4
- 视频轨 codec 为 h264，profile 为 **Main 或 Baseline**，pix_fmt 为 yuv420p
- **必须有音轨且 codec 为 aac**（静音 AAC 合法）； **无音轨 MUST 拒绝**（Phase 1 不补轨）
- faststart **不**.

#### Scenario: v2 rejects Baseline profile

- **WHEN** OSS 对象为 h264 Baseline + AAC + faststart 的 mp4 且 register 请求 `transformVersion=v2`
- **THEN** 验真 MUST 失败且 register MUST 返回 4xx

#### Scenario: v1 accepts Baseline with AAC

- **WHEN** Web 上传 h264 Baseline + AAC 的 mp4（可无 faststart）且验真版本为 v1
- **THEN** 验真 MUST 通过

#### Scenario: v1 rejects missing audio track

- **WHEN** 视频仅有视频轨、无音轨，且验真版本为 v1
- **THEN** 验真 MUST 失败且 MUST NOT 直传 OSS（Web upload）或 MUST 拒绝 register

#### Scenario: v2 rejects missing AAC

- **WHEN** 视频含非 AAC 音轨（如 mp3）且验真版本为 v2
- **THEN** 验真 MUST 失败

#### Scenario: v2 rejects missing faststart

- **WHEN** mp4 满足 h264 Main + AAC 但 moov 不在 mdat 前且验真版本为 v2
- **THEN** 验真 MUST 失败

### Requirement: Web video proxy upload SHALL validate v1 before OSS PUT

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 时 MUST 对上传字节执行 **v1 验真**，且 MUST NOT 进行服务端转码：

1. **v1 通过**：MUST 将 **原始上传字节** 直传 OSS；`contentHash` MUST 为原始字节 SHA-256；响应 MUST 提示 `transformVersion=v1`。
2. **v1 未通过**（含不可解码、编码不合规等）：MUST 返回 4xx 且 MUST NOT 在 OSS 上创建对象。

MUST NOT 在 v1 失败后调用 `NormalizeVideo`、`ProbeVideoDecodable` 转码门禁或任何进程内转码后再 PUT。v1 已合规但缺少 faststart 的对象 MUST 直传 `v1`（faststart 非 v1 强制项）。

#### Scenario: Web compliant video uploads direct v1

- **WHEN** Web 上传满足 v1 规则的 mp4（可无 faststart）
- **THEN** 响应 MUST 含 `objectKey`、`cdnUrl`、`contentHash`，OSS MUST 存在与上传字节一致的对象，且 MUST 含 `transformVersion=v1`

#### Scenario: Web non-compliant video rejected without transcode

- **WHEN** Web 上传 HEVC mp4 或无 AAC 音轨或其它未通过 v1 验真的文件（无论 ffprobe 是否可解析）
- **THEN** API MUST 返回 4xx，MUST NOT 创建 OSS 对象，MUST NOT 返回 `transformVersion=v2` 转码成功响应

#### Scenario: Web undecodable video rejected

- **WHEN** Web 上传损坏或非视频文件
- **THEN** API MUST 返回 4xx 且 OSS MUST NOT 存在新对象

#### Scenario: v1 compliant without faststart stays direct

- **WHEN** Web 上传 h264 Main/Baseline + AAC 的 mp4 但 moov 不在 mdat 前（v1 允许）
- **THEN** MUST 直传原始字节且 `transformVersion` MUST 为 `v1`（MUST NOT 转码）

### Requirement: Video RegisterMedia SHALL validate OSS object by transformVersion

`RegisterMedia` 在 `mediaKind=2` 时 MUST：

- 仅接受 `transformVersion` 为 `v1` 或 `v2`；其他值（含 `sim-raw`）MUST 返回 400
- 在登记 blob/ownership 之前 MUST 对 `objectKey` 对应 OSS 对象执行与请求 `transformVersion` 匹配的验真
- 验真失败 MUST 返回 4xx 且 MUST NOT 完成 blob 登记

`mediaKind=1` 行为 MUST 不变（图片 thumb 逻辑不受本 requirement 影响）。

#### Scenario: Native v2 register succeeds

- **WHEN** 客户端 PUT 满足 v2 规则的 mp4 且 `RegisterMedia` 带 `transformVersion=v2` 与正确 contentHash
- **THEN** 登记 MUST 成功

#### Scenario: sim-raw register rejected

- **WHEN** `RegisterMedia` 请求 `transformVersion=sim-raw` 且 `mediaKind=2`
- **THEN** MUST 返回 400 且 MUST NOT 登记

#### Scenario: v1 register validates relaxed rules

- **WHEN** OSS 对象为 v1 合规（Baseline + AAC、无 faststart）且 register `transformVersion=v1`
- **THEN** 登记 MUST 成功

#### Scenario: Register v2 on v1-only object fails

- **WHEN** OSS 对象仅满足 v1（如无 faststart）但 register `transformVersion=v2`
- **THEN** MUST 返回 4xx

---
