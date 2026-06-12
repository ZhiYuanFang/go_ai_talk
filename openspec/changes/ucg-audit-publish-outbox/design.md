## Context

- **现状**：帖/评/资料/私信在 DB 事务成功后调用 `PublishPostCreated` 等（HTTP Management API → `voice.events`）；失败仅 log warning。`StartUcgAuditReconciler` 每 30min（启动即跑）扫描四类 **pending 业务表**（各 LIMIT 50）补发 MQ。
- **已有参照**：`ucg_chat_message_outbox` + `StartChatPersistWorker`（轮询 outbox 表落 MySQL，**不扫业务 pending 态**）。
- **约束**：UCG 域 only；禁止 pending 业务表 reconciler；不新增 `*_test.go`；Publish 仍走 HTTP（Consumer 仍 AMQP）；`auditVersion` 权威在业务表，outbox 载荷在入队时冻结。

## Goals / Non-Goals

**Goals:**

- Publish 失败 **自动恢复**：outbox relay worker 重试直至 `done` 或达最大 attempts。
- 业务写库与 outbox 入队 **同一事务**（至少同一成功路径），避免「库有、outbox 无」。
- 事务提交后 **best-effort 即时 Publish**，降低常态延迟；失败留给 worker。
- **删除** audit reconciler 及全部 pending 表扫描逻辑。
- Relay worker **仅** `SELECT ... FROM ucg_audit_publish_outbox WHERE status IN (pending,failed)`，ORDER BY id LIMIT 1。

**Non-Goals:**

- 推荐分六类事件 outbox（仍 direct HTTP Publish）。
- Publisher 改 AMQP；Consumer 改 outbox。
- 跨库 outbox（outbox 与 UCG 业务同库 `ai_voice_ucg`）。
- Redis 读缓存。

## Decisions

### 1. 表 `ucg_audit_publish_outbox`

```sql
id BIGINT UNSIGNED PK AUTO_INCREMENT,
routing_key VARCHAR(64) NOT NULL,   -- 如 ucg.post.created
payload JSON NOT NULL,              -- 含 auditVersion 等，入队时冻结
status ENUM('pending','done','failed') DEFAULT 'pending',
attempts INT UNSIGNED DEFAULT 0,
last_error VARCHAR(512) DEFAULT '',
created_at BIGINT NOT NULL,
updated_at BIGINT NOT NULL,
KEY idx_relay (status, attempts, id)
```

- **不存** postId 外键约束；幂等靠 consumer CAS，非 outbox 去重。
- `failed` 且 `attempts < max` 仍可被 worker 捞起重试（与 chat outbox 一致）。

### 2. 写入路径（四类事件）

| 事件 | 写入点 | 事务 |
|------|--------|------|
| `ucg.post.created` | `CreatePost` / `UpdatePost` 再提审 | 与帖 UPDATE/INSERT 同事务 |
| `ucg.comment.created` | `CreateComment` | 与评论 INSERT 同事务 |
| `ucg.profile.patch.submitted` | profile submit | 与 job INSERT 同事务 |
| `ucg.chat.msg.created` | `DeliverChatMessage` | 与消息权威行/outbox 同一提交边界（见 chat 流） |

入队函数：`enqueueAuditPublishOutbox(ctx, tx, routingKey, payload map)`。

### 3. 发布路径：即时 + Relay

```
业务 TX ──▶ INSERT outbox (pending)
     │
     └── commit OK
              │
              ├─▶ tryPublishOutboxRow(id)  ── success ──▶ status=done
              │
              └─▶ fail ──▶ 保持 pending，worker 重试

Relay worker (ticker):
  SELECT 1 row pending|failed (attempts < max)
  HTTP Publish(routing_key, payload)
  ok  → done
  err → attempts++, last_error, status=failed (可再捞)
```

- Worker 间隔：`ucg.auditPublish.relayIntervalMs`（默认 1000）；`maxAttempts` 默认 20。
- **禁止** worker 内 JOIN/扫描 `ucg_post.status=1` 等。

### 4. 移除 audit reconciler

- 删除 `audit_reconciler.go`、`StartUcgAuditReconciler` 调用。
- 删除 env：`UCG_AUDIT_RECONCILE_INTERVAL_MIN`、`UCG_AUDIT_PENDING_TIMEOUT_MIN`。
- 删除 dead config：`ucg.green.auditIntervalSeconds`。

### 5. 与 chat persist worker 并存

- `StartChatPersistWorker`：消息体落 `ucg_chat_message`（不变）。
- `StartAuditPublishRelayWorker`：审核事件 Publish（新增）。
- 两者扫**不同 outbox 表**，语义清晰，非「新 pending 扫表」。

### 6. 存量 pending 迁移

- **非 runtime**：部署时可选手工/SQL 一次性 `INSERT INTO ucg_audit_publish_outbox`（从 pending 行读当前 `audit_version`），或由 relay 消费；**禁止**保留定时 reconciler 作为长期机制。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 即时 Publish 与 worker 双发同一 outbox | 第一路径成功即 `done`；worker 只捞 pending；MQ 侧 CAS 幂等 |
| outbox 堆积 | 监控 `status=pending` 计数；告警 |
| 事务内写 outbox 漏改某路径 | tasks 覆盖四类 submit + 再提审；code review |
| attempts 耗尽仍 failed | 日志 + 运维人工 requeue/update status；runbook |

## Migration Plan

1. DDL `hack/sql/ucg_audit_publish_outbox.sql`。
2. 部署 ucg-service：relay worker 启；reconciler 删。
3. 可选：一次性 seed 存量 pending → outbox 行。
4. 回滚：旧镜像含 reconciler（不推荐）；或手工 Publish + 清 outbox。

## Open Questions

- （已决）不扫 pending 业务表；outbox relay 为唯一 Publish 恢复路径。
- （已决）推荐分事件不在本 change 范围。
