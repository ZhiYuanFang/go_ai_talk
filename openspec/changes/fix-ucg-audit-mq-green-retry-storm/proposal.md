## Why

资料（profile）审核 MQ 消费者在 Green 机审已通过、但后续 `approveProfileJobCAS` 落库失败时向 RabbitMQ 返回 error，触发 `Nack(requeue=true)` 无限重投。每次重投都会重新调用阿里云 Green API，导致单条 bio 修改产生约 4 万次 Green 调用与费用风暴；违规帖驳回路径因 CAS 成功可 Ack，故未暴露同类问题。

用户已确认策略：**允许有限次只 retry 落库** — 每个审核 job 最多调用 Green 一次；MQ 重投时若 verdict 已持久化则跳过 Green；仅对 apply 阶段做有界重试；禁止无限 requeue。

## What Changes

- 将帖子与资料审核拆为 **两阶段**：Phase 1 Green 机审（幂等、verdict 落库）→ Phase 2 apply（CAS 更新帖态/资料、写 profile 行）。
- 在 job/post 行或专用字段持久化机审结论（pass/reject + reason），MQ 重投时若 verdict 已存在则 **跳过 Green**，仅重试 apply。
- apply 阶段失败时递增 `apply_attempts`（或等价计数），在配置上限内 Nack requeue；超限后标记 job/post 为 **failed**、记录告警日志，**Ack 或投递 DLQ**，停止重投。
- 统一 `audit_profile_job.go` 与 `audit_post.go`（及必要时 comment/chat）的处理模式，避免 profile 与 post 行为不一致。
- 在 `docs/runbooks/release-deploy-and-run.md` 补充运维说明：清理卡死队列、修复 pending/failed 审核 job 的步骤。
- **不新增**测试文件；**不**改变 App HTTP 协议字段。

## Capabilities

### New Capabilities

- `ucg-audit-mq-reliability`：UCG 审核 AMQP consumer 的投递语义 — 有界 requeue、verdict 幂等、apply 重试上限、超限失败处理与 DLQ/Ack 策略。

### Modified Capabilities

- `ucg-green-audit`：帖子与资料 Green 异步审核 MUST 保证每个 `(entity_id, audit_version)` 至多一次 Green 调用；MQ 重投 MUST 基于已持久化 verdict 仅重试 apply，不得重复计费调用 Green。

## Impact

- **代码**：`internal/services/ucg/audit_profile_job.go`、`audit_post.go`、`audit_mq_consumer.go`；可能新增 `audit_moderation.go` 等共享两阶段逻辑；`internal/platform/eventkit/amqp_consumer.go`（若需 redelivery 感知或 Nack 策略扩展）。
- **数据模型**：`ucg_profile_audit_job`、可能 `ucg_post` 增加 moderation verdict / apply_attempts 字段（或等价中间态 status）；需 DDL 迁移说明。
- **服务**：仅 `ucg-service`（AMQP push consumer 宿主，见 `ucg_mq_runner.go`）。
- **外部依赖**：阿里云 Green API 调用次数显著下降；RabbitMQ 队列 `ucg.profile.patch.submitted.q`、`ucg.post.created.q` 等 redelivery 行为变化。
- **运维**：runbook 增补卡死消息 purge 与 pending job 人工修复流程。
- **基线依据**：`openspec/specs/v2.0.3/spec.md` 中 `ucg-green-audit`、`ucg-service-runtime` 章节。
