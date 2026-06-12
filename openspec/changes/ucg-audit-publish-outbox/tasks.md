## 1. DDL 与配置

- [x] 1.1 新增 `hack/sql/ucg_audit_publish_outbox.sql`（`ucg_audit_publish_outbox` 表）
- [x] 1.2 `config.ucg-service.yaml`：`ucg.auditPublish.relayIntervalMs`、`maxAttempts`；移除 `ucg.green.auditIntervalSeconds`
- [x] 1.3 删除 reconciler 相关 env 文档（`UCG_AUDIT_RECONCILE_*`、`UCG_AUDIT_PENDING_TIMEOUT_*`）

## 2. Outbox 核心

- [x] 2.1 实现 `enqueueAuditPublishOutbox(ctx, tx, routingKey, payload)` 与 outbox 行即时 Publish + 标记 `done`
- [x] 2.2 实现 `audit_publish_relay_worker.go`：`StartAuditPublishRelayWorker`（仅扫 outbox 表，参照 chat persist worker 模式）
- [x] 2.3 重构 `audit_publisher.go`：Publish 入口改为写 outbox + 即时尝试（或统一经 outbox 服务函数）

## 3. 业务挂点（同事务入队）

- [x] 3.1 `post.go`：`CreatePost` / `UpdatePost` 再提审路径同事务入队 `ucg.post.created`
- [x] 3.2 `social.go`：评论创建同事务入队 `ucg.comment.created`
- [x] 3.3 `profile.go`：资料 job 提交同事务入队 `ucg.profile.patch.submitted`
- [x] 3.4 `chat_service.go`：私信投递路径入队 `ucg.chat.msg.created`（与现有 chat 流对齐）

## 4. 清理 reconciler

- [x] 4.1 删除 `audit_reconciler.go` 及 `StartUcgAuditReconciler` 调用（`ucg_mq_runner.go`）
- [x] 4.2 删除 `GreenConfig.AuditIntervalSeconds` 及 yaml 字段

## 5. 启动与文档

- [x] 5.1 `cmd/ucg-service/main.go` 注册 `StartAuditPublishRelayWorker`
- [x] 5.2 更新 `docs/runbooks/rabbitmq-local.md`：outbox relay、禁止 pending reconciler、可选存量 seed 说明
- [x] 5.3 `go build ./...` 通过
- [x] 5.4 确认无新 gateway-app App API / usage 统计变更

## 6. 验收（手动）

- [x] 6.1 停 RabbitMQ 发帖 → API 200 → 启 RabbitMQ → relay 补 Publish → 审核完成
- [x] 6.2 确认无 `[ucg-audit-reconcile]` 日志；有 relay worker 日志
- [x] 6.3 再提审 bump version 后 outbox 载荷为新版、旧版 consumer CAS skip
