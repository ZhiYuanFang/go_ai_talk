## ADDED Requirements

### Requirement: UCG audit MQ consumer SHALL bound apply retries and stop infinite requeue

`ucg-service` 内 UCG 审核 AMQP consumer（含 `ucg.profile.patch.submitted.q`、`ucg.post.created.q` 及本变更纳入的其它审核队列）在 handler 处理单条 delivery 时 MUST 区分 **机审阶段**与 **apply 阶段**。当 apply 阶段失败且实体 `apply_attempts` 未达配置上限时，handler MAY 返回 error 触发 `Nack(requeue=true)`。当 `apply_attempts` 达到上限 `UCG_AUDIT_APPLY_MAX_ATTEMPTS`（默认 5，可通过环境变量覆盖）时，系统 MUST 将实体标记为 apply 失败终态、记录可观测告警日志，且 handler MUST 返回 nil 以 **Ack** 该 delivery，MUST NOT 无限 requeue。

#### Scenario: apply 失败未达上限时 requeue

- **WHEN** 资料 job 机审 verdict 已为 pass，但 `approveProfileJobCAS` 因 DB 错误失败，且 `apply_attempts` 为 2、上限为 5
- **THEN** 系统 MUST 将 `apply_attempts` 递增为 3，handler MUST 返回非 nil error，且该 MQ 消息 MUST 被 Nack requeue

#### Scenario: apply 失败达上限时 Ack 停止风暴

- **WHEN** 同上失败场景且递增后 `apply_attempts` 等于上限 5
- **THEN** 系统 MUST 标记 job 为 apply 失败终态（如 `status=apply_failed` 或等价字段组合），MUST 输出含 jobId 与 auditVersion 的 error 级日志，且 handler MUST 返回 nil 使消息 Ack，后续同 payload 重投 MUST NOT 再次调用 Green

#### Scenario: 队列 redelivery 不触发无限 Nack 环

- **WHEN** 同一 delivery 因历史缺陷已被 requeue 超过 apply 上限，且 DB 已记录 apply 失败终态
- **THEN** handler MUST 返回 nil Ack，队列 `messages_unacknowledged` MUST NOT 因该 job 无限增长

### Requirement: UCG audit handler SHALL treat persisted moderation verdict as Green skip signal

当 `ucg_profile_audit_job` 或 `ucg_post`（及纳入本能力的同类审核实体）在对应 `audit_version` 下已持久化 `moderation_verdict`（非 0）时，MQ consumer handler 在 Phase 2 apply 重试路径 MUST NOT 再次调用 `GreenModerator` 的 `ModerateText` / `ModerateImageURL` / `ModerateVideoURL`。

#### Scenario: MQ 重投跳过 Green 仅重试 apply

- **WHEN** profile job `id=100`、`audit_version=2` 的 `moderation_verdict=1`（pass），`status` 仍为 pending，且 MQ 消息因先前 apply 失败被 redeliver
- **THEN** handler MUST NOT 调用 Green API，MUST 仅执行 apply 阶段（`approveProfileJobCAS` 或等价逻辑）

#### Scenario: 首次消费执行 Green 并落库 verdict

- **WHEN** 新提交 profile job `moderation_verdict=0` 且收到首次 audit MQ 消息
- **THEN** handler MUST 调用 Green 完成机审，且 MUST 在 apply 之前将 `moderation_verdict` 与 `moderation_reason`（若 reject）持久化到 MySQL

#### Scenario: 并发双消费避免重复 Green

- **WHEN** 两条相同 job 的 delivery 几乎同时被不同 consumer 处理，且 `moderation_verdict` 仍为 0
- **THEN** Phase 1 持久化 MUST 使用带 `moderation_verdict=0` 条件的 CAS/UPDATE，至多一条成功写入 verdict；另一条 MUST 读取已写入 verdict 后跳过 Green 进入 apply

### Requirement: UCG audit runbook SHALL document stuck queue remediation

`docs/runbooks/release-deploy-and-run.md` MUST 包含 UCG 审核 MQ 卡死与 apply 失败的人工处理步骤，至少覆盖：暂停 consumer（`UCG_AUDIT_MQ_CONSUMER_ENABLED`）、核对 DB 中 `moderation_verdict` 与 `status`、在安全条件下 purge 队列积压、对 `moderation_verdict=pass` 且长期 pending 的 job 的修复指引。

#### Scenario: 运维按 runbook 处理 profile 队列积压

- **WHEN** `ucg.profile.patch.submitted.q` ready 消息数异常升高，且 DB 显示多条 job 已 `moderation_verdict=1` 但 `status=pending`
- **THEN** 运维人员 MUST 能按 runbook 暂停 consumer、确认无需重复 Green 后清理或重放消息，且 MUST NOT 仅无限重启 consumer 而不查 DB
