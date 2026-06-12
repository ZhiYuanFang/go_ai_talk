## ADDED Requirements

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
