## 1. 基础设施与 DDL



- [x] 1.1 新增 `hack/sql/ucg_mq_green_audit.sql`：`ucg_post.audit_version`、`ucg_post_comment.status`/`audit_version`/`reject_reason`、`ucg_profile_audit_job`（含 `audit_version`）、`ucg_chat_message.audit_status`/`audit_version`/`reject_reason`

- [x] 1.2 更新 GoFrame entity/dao（`ucg_post`、`ucg_post_comment`、`ucg_chat_message`、profile audit job）

- [x] 1.3 `eventkit/routing_keys.go` 注册四个 `ucg.*` routing key

- [x] 1.4 `hack/rabbitmq-init.ps1` / `.sh` 增加 UCG 审核队列与 binding

- [x] 1.5 `docker-compose.microservices.yml` 为 ucg-service 增加 `UCG_AUDIT_MQ_CONSUMER_ENABLED`（或复用文档化 env）

- [x] 1.6 更新 `docs/runbooks/rabbitmq-local.md` 队列列表



## 2. CAS 与 MQ 公共模块



- [x] 2.1 实现 `CasAuditTransition` helper（帖/评/job 用 `status`+`audit_version`；聊天用 `audit_status`+`audit_version`；**SET 子句禁止递增 version**）

- [x] 2.2 实现 `UcgAuditPublisher`：四类 Publish 载荷 **MUST** 含 `auditVersion`（资料含 `jobId`）；reconciler 补发 **MUST 读库内当前版本**

- [x] 2.3 实现 `StartUcgAuditMQConsumer`（Pull + dispatch by routing key；CAS `RowsAffected=0` → ACK 跳过，记录过期/重复日志）

- [x] 2.4 实现 pending 超时 reconciler（启动 + 定时，补发四类事件，版本取自各权威表列）



## 3. 帖子审核 MQ 化



- [x] 3.1 `CreatePost`/`UpdatePost`：单事务 post+media；首提审 `audit_version=1`，再提审 `audit_version++`+`status=1`；Commit 后 Publish `ucg.post.created`（含 `auditVersion`）

- [x] 3.2 重构 `auditPost`：Green 后 CAS `WHERE status=1 AND audit_version=event.auditVersion`；**不递增** version

- [x] 3.3 `post_admin` 驳回改 CAS（带 `audit_version`）；移除无条件 status UPDATE

- [x] 3.4 删除 `audit_worker` 中帖子扫表逻辑



## 4. 评论审核



- [x] 4.1 `AddComment`：insert pending（`audit_version=1` 或再提审递增）；Publish `ucg.comment.created`（含 `auditVersion`）；移除立即 comment_count++ 与 Notify

- [x] 4.2 消费者 `auditComment`：Green + CAS `WHERE status=1 AND audit_version=?`；通过后 increment count + NotifyOnComment

- [x] 4.3 `ListComments`：按 viewer 过滤 status；作者见 rejected+reason

- [x] 4.4 `CommentDTO` 暴露 status/rejectReason/auditVersion（App API，若对外暴露）



## 5. 资料/头像审核



- [x] 5.1 实现 `ucg_profile_audit_job` DAO 与 `EnqueueProfileAuditJob`（`audit_version` 仅存 job 表）

- [x] 5.2 `UpdateMyProfile` 写 job + Publish（`jobId`, `auditVersion`）；**禁止** Redis/`ucg_profile` 作版本源；移除 Redis pending 主路径

- [x] 5.3 消费者 `auditProfileJob`：Green + CAS job 表（`WHERE status=1 AND audit_version=event.auditVersion`）+ apply profile；**不递增** job version

- [x] 5.4 `mergeProfileForAuthor` 改读 job 表；删除 audit_worker profile 分支



## 6. 私信模式 A



- [x] 6.1 `ProcessOutboundChatMessage`：移除同步 Green；先 Redis pending（镜像 `ucg_chat_message.audit_version`）+ unread + WS deliver + outbox + Publish（含 `auditVersion`）

- [x] 6.2 消费者 `auditChatMessage`：Green + MySQL CAS `WHERE audit_status='pending' AND audit_version=?`；Redis LSET 同步；过期消息 ACK 跳过

- [x] 6.3 WS：`audit_failed`（sender）、`msg_hidden`（recipient）

- [x] 6.4 `listChatMessages`：收件人过滤 rejected；发送人保留 rejected+reason

- [x] 6.5 `chat_persist` 写入/更新 `audit_status`、`audit_version`（与 Redis/MQ 载荷一致）



## 7. 收尾



- [x] 7.1 移除或空实现 `StartAuditWorker`；`main.go` 仅启 MQ consumer

- [x] 7.2 存量 `status=1` 帖子 reconciler 验收说明写入 runbook（补发须带当前 `audit_version`）

- [x] 7.3 确认未新增 App gateway usage 统计路由（本变更无新对外 App API）

- [x] 7.4 部署验收：发帖/评论/改头像/私信 + RabbitMQ 载荷含 `auditVersion` + CAS 0 行 ACK + 再提审旧消息不脏写


