## MODIFIED Requirements

### Requirement: UCG 审核消费者 SHALL 驻留 ucg-service 并按 routing key 分发

`ucg-service` MUST 启动 UCG 审核 **AMQP push consumer**（TCP 5672，`autoAck=false`），从四个 UCG 审核 durable 队列接收 broker 推送的消息并调用 Green 审核逻辑。consumer MUST NOT 部署在 worker-service 内直连 ucg 数据库。消费者 MUST 从载荷读取 `auditVersion` 用于 CAS。每个 UCG 审核队列 MUST 运行 **一个** AMQP consumer goroutine；并发 MUST 由 channel **`prefetch`**（环境变量 `UCG_AUDIT_MQ_PREFETCH`，默认 5）控制。MUST NOT 使用 HTTP Management API `/queues/.../get` 轮询作为 UCG 审核消费路径。

#### Scenario: 消费帖子审核消息

- **WHEN** 队列 `ucg.post.created.q` 收到合法 `ucg.post.created` 载荷（含 `postId`、`auditVersion`）
- **THEN** consumer MUST 执行 Green 审核并在成功后 CAS 更新帖子状态（条件含 `status` 与 `audit_version`），且 MUST 在处理完成后 **manual Ack**

#### Scenario: 四队列并行消费

- **WHEN** ucg-service 启动且 `UCG_AUDIT_MQ_CONSUMER_ENABLED=true`
- **THEN** 系统 MUST 对 `ucg.post.created.q`、`ucg.comment.created.q`、`ucg.profile.patch.submitted.q`、`ucg.chat.msg.created.q` 各建立 AMQP Consume，且 MUST NOT 依赖 ticker 轮询依次拉取四队列

## ADDED Requirements

### Requirement: UCG AMQP 审核 consumer MUST 使用 manual ack 且 SHALL 按处理结果 Ack 或 Nack

UCG 审核 AMQP consumer MUST 以 `autoAck=false` 订阅队列。消息 MUST 在 Green + CAS 业务处理完成且 handler 返回成功后 **Ack**。下列情况 MUST **Ack** 且 MUST NOT requeue：JSON 非法、缺 `auditVersion`、实体不存在、CAS `RowsAffected=0`（过期/重复）。Green 或数据库 **可重试** 错误 MUST **Nack(requeue=true)**。进程在处理完成前崩溃时，broker MUST 能重投未 Ack 的消息。

#### Scenario: 处理成功后 Ack

- **WHEN** consumer 收到合法帖子审核消息且 Green 通过、CAS 更新成功
- **THEN** consumer MUST 向 broker 发送 Ack，该消息 MUST NOT 再次投递给同一 consumer

#### Scenario: 可重试错误 Nack requeue

- **WHEN** Green API 返回临时错误或 MySQL 更新失败且非 CAS 0 行跳过
- **THEN** consumer MUST Nack(requeue=true)，消息 MUST 可再次被消费

#### Scenario: 毒消息 Ack 丢弃

- **WHEN** 消息体非合法 JSON 或缺少 `auditVersion`
- **THEN** consumer MUST 记录 warning 日志并 Ack，MUST NOT 无限 requeue

### Requirement: UCG AMQP consumer SHALL 支持连接断线重连

当 RabbitMQ AMQP 连接或 channel 异常关闭时，ucg-service 中的 UCG 审核 consumer MUST 以 backoff 策略自动重连并恢复四队列 Consume。AMQP 连接失败 MUST NOT 导致 ucg-service HTTP API 进程退出；MUST 记录可观测 warning/error 日志。

#### Scenario: RabbitMQ 短暂不可用后恢复

- **WHEN** ucg-service 运行中 RabbitMQ 重启导致 AMQP 断开
- **THEN** consumer MUST 在连接恢复后继续消费队列消息，reconciler MUST 仍可补发 pending 条目

### Requirement: UCG 审核 Publisher MAY 保持 HTTP Management API

本变更 MUST NOT 要求 UCG 审核事件 Publisher 改为 AMQP。`ucg-service` MAY 继续通过 HTTP `MQ_HTTP_API_BASE` 向 `voice.events` exchange 发布审核事件；AMQP consumer 与 HTTP Publisher 并存 MUST 为受支持的部署形态。

#### Scenario: HTTP 发布 AMQP 消费

- **WHEN** 发帖后 HTTP Publisher 成功发布 `ucg.post.created` 至 exchange
- **THEN** 绑定队列中的消息 MUST 由 AMQP push consumer 接收并处理
