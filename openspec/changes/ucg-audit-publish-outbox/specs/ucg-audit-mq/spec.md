## MODIFIED Requirements

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

### Requirement: MQ Publish 失败 SHALL 可恢复且 MUST NOT 阻塞用户主路径

帖/评/资料 submit API 在事务与 outbox 入队成功但 HTTP Publish 失败时 MUST 记录 warning/error 日志；系统 MUST 通过 **`ucg_audit_publish_outbox` relay worker** 自动重试 Publish（使用 outbox 冻结载荷）。系统 **MUST NOT** 运行定时扫描 pending 审态业务表的 reconciler 作为恢复机制。聊天 WS 在 Redis 投递成功且 outbox 入队成功但 Publish 失败时 MUST 同样由 relay worker 恢复。

#### Scenario: 发帖 Publish 失败仍返回成功

- **WHEN** 帖子已写入 `status=1` 且 outbox 已 commit 但 RabbitMQ 暂不可用
- **THEN** HTTP 创建接口 MAY 仍返回 200 与帖子 DTO；relay worker MUST 在 MQ 恢复后 Publish 并成功标记 outbox `done`

## REMOVED Requirements

### Requirement: MQ Publish 失败 SHALL 可补发且 MUST NOT 阻塞用户主路径（聊天除外可选）

**Reason**: 原 reconciler 通过扫描 pending 业务表补发，与 MQ 主路径及「禁止 pending 扫表」原则冲突；Publish 恢复改由 transactional outbox relay 承担。

**Migration**: 删除 `StartUcgAuditReconciler`；部署 `StartAuditPublishRelayWorker`；存量 pending 一次性 seed outbox 或手工 Publish；更新 runbook 与 env（移除 `UCG_AUDIT_RECONCILE_*`）。
