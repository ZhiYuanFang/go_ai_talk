## Why

UCG 审核 consumer 当前通过 HTTP Management API 轮询（ticker 2s + `ack_requeue_false`），消息在 Pull 时即被 ACK，Green/DB 处理失败或进程崩溃会导致消息丢失，只能依赖 reconciler 补发；延迟也不可控。真实业务已依赖 MQ 驱动审核，应将 consumer 改为 **AMQP push + 手动 ACK**，在 CAS 成功或过期跳过后才 Ack，Publisher 暂保持 HTTP 不变。

## What Changes

- **新增** `eventkit` AMQP consumer 抽象（`amqp091-go`）：连接、重连、Consume、`autoAck=false`、Ack/Nack。
- **替换** `ucg-service` 审核 consumer：四 UCG 队列各 **一个** AMQP consumer goroutine，并发由 **`UCG_AUDIT_MQ_PREFETCH`** 控制（方案 A）。
- **删除** UCG 侧 HTTP Pull consumer、ticker、`UCG_AUDIT_MQ_POLL_INTERVAL_MS`、四队列 round-robin 轮询。
- **保留** HTTP Publisher（`audit_publisher.go`）、`dispatchUcgAuditPayload`、四个 `audit*FromEvent`、reconciler、`UCG_AUDIT_MQ_CONSUMER_ENABLED`。
- **新增** 部署 env：`RABBITMQ_AMQP_URL`（或 host/port 组合）；compose/runbook 文档化。
- **不在本变更**：Publisher AMQP 化、voice/history/device MQ 脚手架清理、worker 探活策略调整。

## Capabilities

### New Capabilities

（无独立新 capability；消费语义并入 `ucg-audit-mq` delta。）

### Modified Capabilities

- `ucg-audit-mq`：消费者 MUST 使用 AMQP 5672 push + manual ack；明确 Ack/Nack 与 CAS/过期/毒消息语义；Publisher 仍可 HTTP。

## Impact

- **代码**：`internal/platform/eventkit/`（AMQP consumer）、`internal/services/ucg/audit_mq_consumer.go`（或 `audit_amqp_consumer.go`）、`go.mod`。
- **配置**：`manifest/docker/docker-compose.microservices.yml` ucg-service env；`docs/runbooks/rabbitmq-local.md`。
- **依赖**：新增 `github.com/rabbitmq/amqp091-go`。
- **运行时**：ucg-service 需可达 RabbitMQ **5672**（consumer）；15672 仍供 Publisher 与现有探活。
- **API/数据**：无对外 App API 变更；无 DDL。
