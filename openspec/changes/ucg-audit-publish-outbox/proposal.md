## Why

UCG 审核已以 AMQP consumer + CAS 为主路径，但 Publish 侧仍为「事务后 fire-and-forget HTTP Publish」；`StartUcgAuditReconciler` 通过**定时扫描 pending 业务表**补发事件，与 MQ 语义冲突（重复投递、状态抖动、不必要 DB 压力），也与推荐分改造后「禁止扫 pending/全表兜底」原则不一致。需要 **Publish 失败可自动恢复**，且恢复路径 **MUST NOT 扫描 `ucg_post`/`ucg_post_comment` 等 pending 审态表**。

## What Changes

- 新增 **`ucg_audit_publish_outbox`** 表与 **Audit Publish Relay Worker**：仅轮询 outbox 表，HTTP Publish 至 RabbitMQ，成功标记 `done`，失败指数退避重试。
- 四类审核事件（帖/评/资料/私信）在**业务事务内**写入 outbox 行（载荷冻结 `auditVersion`）；事务提交后 **best-effort 即时 Publish**，失败由 relay worker 重试。
- **删除** `StartUcgAuditReconciler`、`audit_reconciler.go` 及 `UCG_AUDIT_RECONCILE_*` / `UCG_AUDIT_PENDING_TIMEOUT_*` 环境变量语义。
- **删除** 已无引用的 `ucg.green.auditIntervalSeconds` 配置项。
- runbook / 规格：Publish 失败恢复改为 outbox relay；明确 **禁止** pending 业务表 reconciler。

## Capabilities

### New Capabilities

- `ucg-audit-publish-outbox`：审核事件 transactional outbox 写入、relay worker、重试与幂等标记。

### Modified Capabilities

- `ucg-audit-mq`：将「reconciler 补发 pending 条目」改为「outbox relay 重试 Publish」；移除对 pending 表扫描的要求。
- `ucg-green-audit`：明确审核触发 **MUST NOT** 依赖 pending 业务表定时扫描；Publish 可靠投递由 outbox 保证。

## Impact

- **库表**：`ai_voice_ucg` 新增 `ucg_audit_publish_outbox`（DDL `hack/sql/`）。
- **代码**：`post.go`、`social.go`、`profile.go`、`chat_service.go`、`audit_publisher.go`；新增 `audit_publish_outbox.go`、`audit_publish_relay_worker.go`；删除 `audit_reconciler.go`；`ucg_mq_runner.go` 改启动项。
- **配置**：`config.ucg-service.yaml` relay 间隔/最大重试；移除 reconciler 与 dead green interval 配置。
- **部署**：ucg-service 启动 relay worker；存量 pending 可一次性脚本从 outbox/业务表 seed（非 runtime 扫表）。
- **无新 gateway-app App API**；不涉及 usage 统计变更。
