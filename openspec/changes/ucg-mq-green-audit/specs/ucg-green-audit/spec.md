## MODIFIED Requirements

### Requirement: Post and profile content SHALL use Green async audit Option A

Submitted posts (text/images/video) and profile fields (avatar/nickname/bio) SHALL enter pending visibility: ONLY author MAY see content in feeds/profile until Green pass sets published or profile active state; on fail content SHALL be rejected with reason visible to author as「违规已下架」或等价文案。**触发方式 MUST 为事务提交后 Publish RabbitMQ 事件（载荷含 `auditVersion`），由 ucg-service consumer 异步执行 Green；MUST NOT 依赖定时扫表 worker 发现 pending 帖子或 Redis profile 待审集合。** 全链路 MUST 遵循各实体 `audit_version` 权威列与 CAS 规则（见 data-model / audit-mq 规格）。

#### Scenario: 提交后仅作者可见

- **WHEN** 用户发布帖子且 Green 未完成
- **THEN** 其他用户请求 Feed SHALL NOT 包含该帖；作者 我的动态 SHALL 显示审核中

#### Scenario: 审核通过公开

- **WHEN** Green 返回 pass 且 CAS 成功
- **THEN** post status SHALL 变为 2 且 SHALL 出现在推荐/关注 Feed

#### Scenario: 审核失败

- **WHEN** Green 返回 fail 且 CAS 成功
- **THEN** post status SHALL 变为 3，作者 SHALL 见 reject_reason；其他用户 SHALL NOT 见该帖

#### Scenario: 发帖通过 MQ 触发审核

- **WHEN** 用户 submit 创建帖子且 DB 事务提交成功
- **THEN** 系统 MUST Publish `ucg.post.created`（含 `auditVersion`）且 MUST NOT 依赖 audit_worker 扫表触发首次审核

## REMOVED Requirements

### Requirement: Chat messages SHALL use Green audit Option C before delivery

**Reason**: 产品改为聊天模式 A（先投递后异步审核），收件人可在 pending 阶段可见，驳回后仅发送方可见。

**Migration**: 客户端实现 `message_delivered` 对 pending 展示；处理 `msg_hidden` 与事后 `audit_failed`；历史拉取过滤 rejected（收件人）。

## ADDED Requirements

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

### Requirement: Profile patch SHALL use MySQL audit job and MQ instead of Redis pending queue

资料变更（nickname/avatar/bio）MUST 写入 MySQL `ucg_profile_audit_job` 行（status pending，`audit_version` 仅存 job 表）并 Publish `ucg.profile.patch.submitted`（含 `jobId`、`auditVersion`）；公众 profile MUST 保持旧值直至 CAS 通过 job 并 apply；MUST NOT 长期依赖 Redis `ucg:green:profile:pending` 作为待审或版本权威。消费者 CAS MUST 仅针对 job 表：`WHERE id=? AND status=1 AND audit_version=?`。

#### Scenario: 改头像后公众仍见旧头像

- **WHEN** 用户提交新 avatarKey 且 Green 未完成
- **THEN** `GET profile/{wxId}` MUST 返回旧 avatar；作者 `profile/me` MUST 可见 pending 预览

#### Scenario: 资料 CAS 使用 job audit_version

- **WHEN** consumer 处理 jobId=5 且载荷 `auditVersion=1`
- **THEN** UPDATE MUST 条件含 `ucg_profile_audit_job.status=1 AND audit_version=1`；MUST NOT 读 Redis 或 `ucg_profile` 作版本

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
