## MODIFIED Requirements

### Requirement: UCG 内容审核 SHALL 由 MQ 事件异步触发且 MUST 使用 audit_version CAS

Submitted posts (text/images/video) and profile fields (avatar/nickname/bio) SHALL enter pending visibility: ONLY author MAY see content in feeds/profile until Green pass sets published or profile active state; on fail content SHALL be rejected with reason visible to author as「违规已下架」或等价文案。**触发方式 MUST 为事务提交后写入 `ucg_audit_publish_outbox` 并经 relay HTTP Publish RabbitMQ 事件（载荷含冻结 `auditVersion`），由 ucg-service AMQP consumer 异步执行 Green；MUST NOT 依赖定时扫 pending 业务表 worker 或 reconciler 发现漏审条目。** 全链路 MUST 遵循各实体 `audit_version` 权威列与 CAS 规则（见 data-model / audit-mq 规格）。

#### Scenario: 发帖不再依赖 audit_worker 或 pending reconciler

- **WHEN** 用户 submit 发帖且 outbox 与 relay 正常
- **THEN** 系统 MUST Publish `ucg.post.created`（含 `auditVersion`）且 MUST NOT 依赖 `StartUcgAuditReconciler` 或 audit_worker 扫表触发首次审核

#### Scenario: Publish 失败由 outbox relay 恢复

- **WHEN** 帖子已 pending 且 outbox 行 Publish  initially 失败
- **THEN** relay worker MUST 重试 Publish；MUST NOT 通过扫描 `ucg_post.status=1` 补发
