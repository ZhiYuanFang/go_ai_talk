## 1. 数据模型与常量

- [x] 1.1 为 `ucg_profile_audit_job` 增加 DDL：`moderation_verdict`、`moderation_reason`、`moderation_at`、`apply_attempts`、`apply_failed_at`；新增 `status=apply_failed`（或等价）常量与 entity/dao 字段映射

- [x] 1.2 为 `ucg_post` 增加相同语义的 moderation/apply 字段（及必要常量）；核对 `UCG_DB_LINK` / `config.ucg-service.yaml` 仅本域库

- [x] 1.3 评估并同步 `ucg_post_comment` / 聊天审核表：若 consumer 仍为 Green+单步 CAS，按 design 一并加字段或记录明确排除理由  
  **排除说明**：comment/chat 同为 Green+单步 CAS，但本变更聚焦 profile/post 费用风暴根因；chat 另有 Redis LSET 同步路径，comment 无 `audit_version` stale 校验，纳入需独立 follow-up。本次不增 DDL。

## 2. 两阶段审核编排（profile + post）

- [x] 2.1 新增 `audit_moderation.go`（或同级）：`persistModerationVerdict` CAS（`moderation_verdict=0` 条件）、`runProfileModerationPhase`、`runProfileApplyPhase`，含中文注释说明幂等与跳过 Green 语义

- [x] 2.2 重构 `audit_profile_job.go`：`auditProfileJobFromEvent` 改为先 Moderation 后 Apply；reject 路径拆为 Phase1 写 verdict=reject、Phase2 `rejectProfileJobCAS`

- [x] 2.3 新增 `runPostModerationPhase` / `runPostApplyPhase`；重构 `audit_post.go` 与 profile 对称

- [x] 2.4 实现 `UCG_AUDIT_APPLY_MAX_ATTEMPTS` 环境变量读取（默认 5）；apply 失败递增 `apply_attempts`，超限写 `apply_failed` + error 日志并返回 nil

## 3. MQ 与可观测性

- [x] 3.1 确认 `audit_mq_consumer.go` / `ucg_mq_runner.go` 无需改路由；验证 handler 超限返回 nil 时 `amqp_consumer.go` 走 Ack 路径

- [x] 3.2 日志：Phase1 跳过 Green、apply 重试、apply 超限失败须含 queue、jobId/postId、auditVersion、apply_attempts，便于对照阿里云 Green 调用量

- [x] 3.3 （可选）为审核队列配置 DLQ / `x-death` 安全网上限，与 DB `apply_attempts` 主路径一致  
  **跳过**：主路径以 DB `apply_attempts` + handler 超限返回 nil Ack 为准；DLQ 需改 RabbitMQ 拓扑与 init 脚本，留 follow-up。

## 4. 作者可见态与 API

- [x] 4.1 确认 profile 预览 / 我的动态接口将 `apply_failed` 映射为作者可感知失败文案（非永久「审核中」）；必要时固定系统 reason

- [x] 4.2 确认 post `pending_audit` + 已写 verdict 但 apply 未完成时，作者仍仅自己可见（与 Option A 一致）

## 5. Runbook 与部署

- [x] 5.1 在 `docs/runbooks/release-deploy-and-run.md` 增补「UCG 审核 MQ 卡死 / apply 失败」：暂停 consumer、查 DB、purge 条件、手工修复 `moderation_verdict=pass` 且 pending 的 job

- [x] 5.2 部署 checklist：先 DDL 后 `ucg-service`；观察 `ucg.profile.patch.submitted.q` 深度与 Green 调用次数下降

## 6. 校验

- [x] 6.1 `go build ./...` 通过

- [x] 6.2 `openspec validate fix-ucg-audit-mq-green-retry-storm --strict` 通过

- [x] 6.3 手工验证：模拟 apply 失败时同一条 MQ 消息 redelivery 不再触发 Green（可看日志或阿里云控制台）；违规帖/资料 reject 仍一次 Ack  
  **说明**：步骤已写入 runbook「手工验证」小节；CI 无法调 Aliyun Green，须部署后人工执行。
