## Why

UCG 内容审核当前依赖 `audit_worker` 每 5 秒轮询 MySQL/Redis 待审队列，延迟不可控且与项目 RabbitMQ 事件化方向不一致。需要将 **帖子、评论、资料、私信** 的 Green 机审统一为 **事务提交后发 MQ、消费者异步审核、CAS 更新状态**，并消除 audit 扫表轮询。

## What Changes

- **移除** `StartAuditWorker` 中帖子/资料扫表与 Redis 待审轮询逻辑。
- **新增** RabbitMQ routing keys：`ucg.post.created`、`ucg.comment.created`、`ucg.profile.patch.submitted`、`ucg.chat.msg.created`；在 `ucg-service` 内消费并调 Green。
- **跨实体 `audit_version` 规则（MUST）**：
  - **所有**审核链路（Publish 载荷 + 消费者 CAS）MUST 携带 `auditVersion`；其语义仅为 **审核轮次/迭代号**，不代表业务回滚。
  - **版本递增**仅发生在用户 **提交/再提审** 时；消费者 CAS 成功 **MUST NOT** 递增 `audit_version`。
  - **业务表不因再提审而 reset/rollback**（如已 published 的帖文内容、已 apply 的资料字段）；再提审仅递增版本并将审态置 pending，由 CAS 拦截过期 MQ 消息。
  - **消费者 CAS** MUST 使用 `WHERE ... AND status=? AND audit_version=?`（聊天为 `audit_status` + `audit_version`）；`RowsAffected=0` MUST 视为重复或过期消息，ACK 且 MUST NOT 污染当前状态。
- **帖子**：`ucg_post.audit_version` 为版本源；发帖/再提审事务提交后 Publish（含 `auditVersion`）；消费者 CAS 更新 `status`（1→2/3）；**禁止**无条件 UPDATE status。
- **评论**：`ucg_post_comment.audit_version` 为版本源；新增 `status`、`reject_reason`；发表时 pending 且 `audit_version=1`（再提审递增）；Publish 含 `auditVersion`；审过后才公开、`comment_count++` 与通知；列表过滤非 published。
- **资料/头像**：`ucg_profile_audit_job.audit_version` 为 **唯一**版本源（**禁止** Redis 或 `ucg_profile` 表承载版本）；提交 job 后 Publish（含 `jobId` + `auditVersion`）；通过后 CAS job 并 apply 到 `ucg_profile`。
- **私信**：`ucg_chat_message.audit_version` 为版本源（Redis JSON 镜像同值）；**BREAKING** 由 Option C 改为 **模式 A（先推后审）**——投递后收件人可见 pending；Publish 含 `auditVersion`；异步 MQ 审；reject 后仅发送方可见+违规提示，收件人历史过滤 rejected；未读在收到时 +1 且 reject 不回滚。
- **基础设施**：`eventkit` 注册 `ucg.*` routing keys；`hack/rabbitmq-init.*` 增加队列绑定；compose 为 `ucg-service` 启用 MQ consumer。
- **Non-Goals（本变更不做）**：history/device/voice/worker 域 outbox；推荐 `recommend_worker`；`chat_persist_worker` 机制替换；Green 控制台开通。

## Capabilities

### New Capabilities

- `ucg-audit-mq`：UCG 域 RabbitMQ 事件发布、队列拓扑、消费者框架、载荷 `auditVersion` 契约与 CAS 幂等语义。

### Modified Capabilities

- `ucg-green-audit`：帖子/评论/资料/聊天 Green 触发方式、可见性规则与全实体 CAS + `audit_version`。
- `ucg-data-model`：各实体 `audit_version` 列、评论审态、profile audit job、聊天 `audit_status` 等表结构。
- `ucg-chat-ws`：私信 WS 投递与 `audit_failed` / `msg_hidden` 等事后驳回事件。
- `ucg-app-http-api`：评论列表/发表响应语义；资料 pending 展示（数据源从 Redis 迁 MySQL job）。

## Impact

- **服务**：仅 `ucg-service`（+ 静态资源若 WS 协议文档化）。
- **数据库**：`ai_voice_ucg` DDL migration；存量 pending 帖子/评论需迁移或补发 MQ 策略见 design。
- **中间件**：RabbitMQ 新队列；Redis 移除 profile 待审键（迁移期可兼容读）。
- **客户端**：私信需处理先投递、事后隐藏；评论需展示审核中/违规态。
- **规格**：替换 v2.0.3 `ucg-green-audit` 聊天 Option C 相关 Requirement；全 artifact 统一 `audit_version` + CAS 规则。
