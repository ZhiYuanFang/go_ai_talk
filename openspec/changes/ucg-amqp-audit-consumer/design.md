## Context

- **现状**：`ucg-mq-green-audit` 已实现四类 UCG 审核 MQ 事件、HTTP Publisher、HTTP Pull consumer（ticker + `PullOne`/`ack_requeue_false`）、reconciler、CAS + `audit_version`。
- **问题**：Pull 即 ACK，处理前消息已从队列删除；2s 轮询带来延迟；单 worker 四队列 round-robin 吞吐低。
- **约束**：仅 UCG 域；consumer 驻留 `ucg-service`；Publisher **本变更不改**（仍 HTTP 15672）；不新增 `*_test.go`；Go 1.19。

## Goals / Non-Goals

**Goals:**

- UCG 审核 consumer 改为 **AMQP 5672 push + manual ack**（`autoAck=false`）。
- **方案 A**：每个 UCG 审核队列 **一个** consumer goroutine；并发由 **`UCG_AUDIT_MQ_PREFETCH`**（默认 5）控制。
- 处理成功、过期跳过（CAS 0 行）、毒消息（非法 JSON/缺字段）→ **Ack**；可重试错误（Green/DB 临时失败）→ **Nack(requeue)**。
- AMQP 断线自动重连；不阻断 ucg-service 主进程启动（与现有 `RequireRabbitMQ: false` 一致）。
- 复用 `dispatchUcgAuditPayload` 与现有 `audit*FromEvent` 业务逻辑。

**Non-Goals:**

- HTTP Publisher 改 AMQP。
- voice/history/device MQ 脚手架删除。
- worker-service 探活/RabbitMQ 强依赖策略调整。
- DLQ 拓扑（首版可依赖 Nack requeue + reconciler；后续按需加 DLQ）。

## Decisions

### 1. AMQP 客户端库

- **选用** `github.com/rabbitmq/amqp091-go`（官方维护，Go 1.19 兼容）。
- **替代**：继续 HTTP Pull — 无法真正 manual ack，放弃。

### 2. 模块分层

```
internal/platform/eventkit/amqp_consumer.go   # 通用：Dial、Consume、Ack、Nack、Reconnect
internal/services/ucg/audit_amqp_consumer.go # UCG：四队列注册、ACK 决策、调 dispatch
internal/services/ucg/audit_mq_consumer.go     # 保留 StartUcgAuditMQConsumer 入口，内部切 AMQP
```

- **理由**：eventkit 便于后续其他域复用 AMQP；UCG 只关心队列名与 dispatch。

### 3. 连接与消费拓扑（方案 A）

```
1 × amqp.Connection（共享，带 reconnect loop）
4 × amqp.Channel（每队列一个）
4 × goroutine：Channel.Consume(queue, autoAck=false)
prefetch：Qos(prefetch, global=false)  per channel
```

- **不采用** 多 worker 轮询同一队列（方案 B）首版不做；水平扩展时可在 K8s 多副本 ucg-service，每副本各跑四 consumer。
- **删除** `UCG_AUDIT_MQ_POLL_INTERVAL_MS`、`UCG_AUDIT_MQ_WORKERS`（poll/worker 语义废弃）；**新增** `UCG_AUDIT_MQ_PREFETCH`（默认 5）。

### 4. 环境变量

| 变量 | 用途 |
|------|------|
| `RABBITMQ_AMQP_URL` | 优先，如 `amqp://guest:guest@rabbitmq:5672/` |
| `RABBITMQ_HOST` + `RABBITMQ_AMQP_PORT` | 无 URL 时拼装；host 默认 `rabbitmq`，port 默认 `5672` |
| `MQ_USER` / `MQ_PASSWORD` | 与 Publisher 共用凭证 |
| `UCG_AUDIT_MQ_CONSUMER_ENABLED` | 不变 |
| `UCG_AUDIT_MQ_PREFETCH` | 新，默认 5 |

Publisher 仍用 `MQ_HTTP_API_BASE`（15672），两端口并存。

### 5. Ack / Nack 映射

| `dispatchUcgAuditPayload` / handler 结果 | AMQP |
|------------------------------------------|------|
| `nil`（含 stale skip、实体缺失、CAS 0 行已由 handler 返回 nil） | **Ack** |
| JSON 解析失败、缺 `auditVersion`（dispatch 内 return nil 前已 log） | **Ack**（毒消息，避免卡队列；reconciler 兜底 pending） |
| Green API 错误、DB 错误（handler return err） | **Nack(requeue=true)** |
| 进程 crash  mid-handler | 消息 **unacked**，broker 重投 |

- **无限 requeue 风险**：Green 长期不可用时消息在队列与 unacked 间抖动；**缓解**：reconciler 扫 MySQL pending；日志告警；后续变更可加 `x-death` 上限 Ack。

### 6. 重连策略

- Connection 关闭 → exponential backoff（如 1s→30s cap）→ redial → 重建四 channel + Consume。
- 单次 handler 超时不单独 cancel delivery（Green 已有 timeout 配置）；依赖 Nack 或进程重启重投。

### 7. 启动与探活

- `runtimecheck` **不改**（仍 HTTP 探活，供 Publisher）。
- `StartUcgAuditMQConsumer`：若 enabled，启动 AMQP reconnect loop；首次连接失败 **log warning**，不 `Fatalf`。
- `StartUcgAuditReconciler` **不变**，仍随 consumer 启动。

### 8. 队列与 routing key

- 队列名不变：`ucg.post.created.q` 等；init 脚本 **不改**。
- dispatch 仍按 **queue 名** 分支（与现实现一致）；AMQP delivery 的 `RoutingKey` 可作日志，非必须改 dispatch 签名。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| AMQP 连接数/内存 | 单 connection + 4 channel，prefetch 限制 unacked |
| Green 失败无限 Nack | reconciler + 监控队列 depth；后续 DLQ |
| 部署漏配 5672 | compose/runbook 显式 `RABBITMQ_AMQP_URL`；consumer 连不上打 error 日志 |
| Publisher HTTP / Consumer AMQP 双协议 | 短期可接受；下阶段可统一 AMQP publish |

## Migration Plan

1. 部署新版本 ucg-service（含 amqp091-go + env）。
2. 滚动重启：旧 Pull consumer 停止，新 AMQP consumer 接管；队列中未消费消息由新 consumer 立即 push 处理。
3. 无需 DDL；无需重跑 rabbitmq-init（拓扑不变）。
4. **回滚**：回退镜像；Pull consumer 代码若已删则需回滚版本（建议实现时删除 Pull 路径但保留 git 历史）。

## Open Questions

- （已决）方案 A：每队列单 consumer + prefetch — 用户已确认。
- DLQ / max-redeliver 是否首版就上 — **建议否**，reconciler 兜底。
