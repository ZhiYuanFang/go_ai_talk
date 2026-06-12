## 1. 依赖与配置

- [x] 1.1 `go.mod` 引入 `github.com/rabbitmq/amqp091-go`
- [x] 1.2 `docker-compose.microservices.yml` ucg-service 增加 `RABBITMQ_AMQP_URL`（或 document host/port）；Publisher 用 env 不变
- [x] 1.3 `manifest/docker/.env.example` 与 `docs/runbooks/rabbitmq-local.md` 文档化 AMQP 5672 consumer 与 HTTP 15672 publisher 双端口

## 2. eventkit AMQP 抽象

- [x] 2.1 实现 `internal/platform/eventkit/amqp_consumer.go`：Dial、`Consume(queue, autoAck=false)`、Ack、Nack、Qos prefetch
- [x] 2.2 实现 connection/channel 断线 backoff 重连 loop（可关闭/ctx cancel）

## 3. UCG AMQP consumer

- [x] 3.1 实现 `audit_amqp_consumer.go`：四队列各一 goroutine Consume；读 env `RABBITMQ_AMQP_URL` / host/port + `MQ_USER`/`MQ_PASSWORD`
- [x] 3.2 处理 delivery：调现有 `dispatchUcgAuditPayload`；成功/stale/毒消息 → Ack；可重试 err → Nack(requeue)
- [x] 3.3 重构 `StartUcgAuditMQConsumer`：enabled 时启 AMQP + reconciler；删除 ticker、HTTP `PullOne`、四队列 round-robin
- [x] 3.4 删除废弃 env：`UCG_AUDIT_MQ_POLL_INTERVAL_MS`、`UCG_AUDIT_MQ_WORKERS`；新增 `UCG_AUDIT_MQ_PREFETCH`（默认 5）

## 4. 清理与验证

- [x] 4.1 确认 `audit_publisher.go` 与 HTTP Publisher **未改动**
- [x] 4.2 `go build ./...` 通过
- [x] 4.3 手动验收：发帖 → 无 2s 轮询延迟 → CAS 成功 Ack；Green 失败消息 requeue；再提审旧 version CAS 0 行 Ack
- [x] 4.4 确认无新 gateway-app App API / usage 统计变更
