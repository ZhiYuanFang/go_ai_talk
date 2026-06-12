## Context

### 现状与根因

UCG 审核由 `ucg-service` 内 AMQP push consumer 驱动（`ucg_mq_runner.go` → `audit_mq_consumer.go`）。`eventkit.amqp_consumer` 约定：**handler 返回 nil → Ack；非 nil → `Nack(false, true)` 无限 requeue**。

资料审核 `auditProfileJobFromEvent`（`audit_profile_job.go`）在单函数内串联：

1. 校验 job 仍为 `pending` 且 `audit_version` 匹配；
2. 依次对 nickname / bio / avatar 调用 `EffectiveGreen()`；
3. 全部通过后调用 `approveProfileJobCAS`（事务内 CAS `pending→approved` 并更新 `ucg_profile`）。

当 Green 已通过（阿里云控制台显示 PASS）但 **Phase 2 落库失败**（如 `approveProfileJobCAS` 事务错误、profile 行更新失败）时，handler 返回 error → 消息 requeue → **每次重投重新执行 Green**，产生费用风暴。违规帖 `rejectPostCAS` 路径通常一次 CAS 成功即 Ack，故未复现。

帖子审核 `auditPostFromEvent`（`audit_post.go`）结构相同，存在同类隐患。

### 约束

- 用户决策：**允许有限次只 retry 落库**；每个 `(entity_id, audit_version)` Green 至多一次；MQ 重投跳过已持久化 verdict；apply 有界重试；禁止无限 requeue。
- 仓库约定：不新增测试文件；UCG AMQP consumer 为已批准 push consumer（`ucg-service`）。
- 不引入 Redis 读缓存；以 MySQL 为 verdict 权威。
- 帖子与资料审核 MUST 采用一致两阶段模式。

## Goals / Non-Goals

**Goals:**

- 将 Green 机审与业务 apply 拆为两阶段，verdict 在 MySQL 持久化后方可在 MQ 重投时跳过 Green。
- apply 失败仅重试落库，计数有界；超限标记失败、告警、Ack（或 DLQ），停止 requeue。
- profile 与 post（至少）统一重构；comment/chat 若结构相同则一并纳入。
- 增补 runbook：清理卡死队列、修复 pending/failed job。

**Non-Goals:**

- 改变 Green API 调用参数、service 名或 pass/fail 判定逻辑（`green_client.go` 解析层另属其他变更）。
- 修改 App HTTP 协议或作者可见字段语义（仍 pending / approved / rejected）。
- 回填历史已产生的超额 Green 账单。
- 新增 `*_test.go` 或端到端自动化测试文件。
- 改造 UCG 推荐 consumer 或其它非审核队列。

## Decisions

### 1. 两阶段状态机（MySQL 权威）

**选择**：在 `ucg_profile_audit_job` 与 `ucg_post` 增加机审结果字段（DDL），不新增中间 `status` 值对外暴露：

| 字段 | 类型语义 | 说明 |
|------|----------|------|
| `moderation_verdict` | `0=未审` `1=pass` `2=reject` | Phase 1 完成后写入 |
| `moderation_reason` | varchar | reject 时 Green reason；pass 时空 |
| `moderation_at` | int64 unix | 机审结论落库时间 |
| `apply_attempts` | int 默认 0 | Phase 2 失败计数 |
| `apply_failed_at` | int64 可空 | 超限失败时间 |

profile job 终态仍用现有 `status`（1 pending / 2 approved / 3 rejected）；新增 **`4=apply_failed`**（或等价命名）表示机审已完成但 apply 超限失败，供运维与作者查询。

帖子侧：在 `pending_audit` 期间用上述字段表达 Phase 1 完成；终态仍为 `published`/`rejected`；可增 `apply_failed` 等价态或保留 `pending_audit` + `apply_attempts` 超限标记（design 实现时二选一，**spec 以「超限后不得再 requeue」为准**）。

**Phase 1 — Moderate**（幂等）：

```
if moderation_verdict != 0:
    skip Green
else:
    run Green checks
    CAS/UPDATE: moderation_verdict, moderation_reason, moderation_at
    (job/post 仍为 pending / pending_audit)
```

**Phase 2 — Apply**（有界重试）：

```
if status already terminal (approved/rejected/published): return nil
if moderation_verdict == 0: return error (不应发生，打日志)
if verdict == reject: reject*CAS
if verdict == pass: approve/publish*CAS
on success: return nil
on error:
    increment apply_attempts
    if apply_attempts >= MAX: mark apply_failed, log alert, return nil (Ack)
    else: return error (Nack requeue)
```

**理由**：verdict 与 apply 解耦后，MQ 重投仅读 DB 即可跳过 Green；不依赖 Redis 或消息体携带 verdict。

**备选**：仅用 RabbitMQ `x-death` 计数限制重投 — 拒绝，无法区分「Green 未做」与「仅 apply 失败」，仍会重复调 Green。

### 2. 共享编排函数

**选择**：新增 `audit_moderation.go`（或同级包内函数）提供：

- `runProfileModerationPhase(ctx, job) (skipGreen bool, err error)`
- `runProfileApplyPhase(ctx, job, auditVersion) error`
- `runPostModerationPhase` / `runPostApplyPhase` 对称实现

`auditProfileJobFromEvent` / `auditPostFromEvent` 瘦身为：加载行 → stale 跳过 → 调两阶段。

**理由**：避免 profile/post 复制粘贴；后续 comment/chat 可复用。

### 3. Green 调用次数保证

**选择**：以 `(id, audit_version)` + `moderation_verdict != 0` 作为「已机审」判定；Phase 1 写 verdict 使用 **单条 CAS/UPDATE**（`WHERE status=pending AND audit_version=? AND moderation_verdict=0`），避免并发 consumer 双调 Green。

Green **瞬时错误**（网络/5xx）：Phase 1 返回 error → Nack requeue → **允许重试 Green**（因 verdict 未写入）。这与「verdict 已持久化则跳过」不冲突。

**理由**：用户要求「每个 audit job 至多一次 Green」指 **业务意义上的一次判定**，非绝对禁止 API 瞬态重试；一旦 verdict 落库则永不重调。

### 4. apply 重试上限与 MQ 语义

**选择**：

- 默认 `UCG_AUDIT_APPLY_MAX_ATTEMPTS=5`（可 env 覆盖）。
- 未超限时 Phase 2 错误 → handler 返回 error → 现有 `Nack(requeue=true)`。
- 超限时：DB 标记 `apply_failed`（及 `apply_failed_at`），`glog.Error` 结构化告警（含 jobId/postId、wxId、auditVersion、attempts），handler 返回 **nil → Ack**。
- **可选增强**（tasks 中实现）：为审核队列配置 DLQ + `x-death` 上限作安全网；主路径以 DB `apply_attempts` 为准。

**不修改** `amqp_consumer.go` 全局 Nack 语义（避免影响其它 consumer）；仅在 UCG handler 层通过「超限返回 nil」停止 requeue。

**备选**：扩展 `AMQPMessageHandler` 返回 `(retry bool, err error)` — 改动面过大，本变更不采用。

### 5. 帖子与资料一致性

**选择**：`audit_post.go` 与 `audit_profile_job.go` **必须**采用相同两阶段与字段语义；`audit_comment.go` / `audit_chat.go` 若仍为「Green + 单步 CAS」结构，本变更 **一并重构** 或至少在 tasks 中列为 follow-up（proposal 要求一致性时优先一并改）。

帖子驳回路径当前 `rejectPostCAS` 一步完成；重构后改为 Phase 1 写 verdict=reject，Phase 2 执行 reject CAS（与 profile 对称）。

### 6. 与现有 stale 跳过逻辑的关系

现有逻辑：`status != pending` 或 `audit_version` 不匹配 → 返回 nil（Ack）。

保留；另增：若 `moderation_verdict != 0` 且 `status` 仍为 pending → **仅执行 Phase 2**，不调 Green。

### 7. Runbook 运维流程

在 `docs/runbooks/release-deploy-and-run.md` 增补 **「UCG 审核 MQ 卡死 / apply 失败」** 小节：

1. **识别**：队列 `ucg.profile.patch.submitted.q` / `ucg.post.created.q` 的 `messages_ready` 持续增长；日志 `[ucg-audit-mq]` + `apply_attempts` 或 Green 重复调用。
2. **止血**：可设 `UCG_AUDIT_MQ_CONSUMER_ENABLED=false` 暂停 consumer；部署含本修复的版本。
3. **清理队列**（仅当消息对应 job 已在 DB 终态或已 `apply_failed`）：Management API `purge` 或删除指定 `delivery_tag` 重投积累（须先核对 DB）。
4. **修复 pending job**：`moderation_verdict=1` 且 `status=1` → 可手工执行 apply SQL 或重发 outbox；`apply_failed` → 运维确认后重置 `apply_attempts` 并重新入队或人工 approve。
5. **禁止**：在 verdict 未落库时直接 Ack 消息而不处理。

## Risks / Trade-offs

- **[Risk] DDL 迁移遗漏测试/生产** → tasks 须含 migration 脚本与 `UCG_DB_LINK` 核对；deploy 前备份。
- **[Risk] Phase 1 写 verdict 成功、Phase 2 长期失败** → 作者长时间见「审核中」；超限后转 `apply_failed` 并告警，runbook 定义人工介入。
- **[Risk] 双 consumer 竞态** → Phase 1 CAS 带 `moderation_verdict=0` 条件；Phase 2 仍用现有 `CasAuditTransition`。
- **[Risk] 历史积压消息无新字段** → 迁移后旧行 `moderation_verdict=0`，首次消费仍调 Green（可接受，一次性）；若需避免可对积压先 purge（runbook）。
- **[Trade-off] Green 瞬态错误仍会 requeue 并重调 API** → 仅在 verdict 未落库时发生，次数受 RabbitMQ 与运维关注限制；远优于无限 apply 失败风暴。
- **[Trade-off] 新增 `apply_failed` 状态** → App 需将此类 job 展示为审核失败或「系统异常」（若当前仅映射 1/2/3，实现时确认 UI 文案，可能映射为 rejected + 固定 reason）。

## Migration Plan

1. 执行 DDL：为 `ucg_profile_audit_job`、`ucg_post`（及 comment/chat 若纳入）增加 moderation/apply 字段；默认值使旧行等价「未机审」。
2. 部署新版 `ucg-service`；观察 Green 调用量与队列深度。
3. 对线上已卡死队列：按 runbook purge 或等 consumer 用新逻辑消费（verdict 落库后不再调 Green）。
4. **回滚**：还原代码；新字段可保留（兼容）；若回滚后仍无限 requeue，须再次暂停 consumer 或 purge。

## Open Questions

- `apply_failed` 对 App 作者展示文案是否复用「违规已下架」或单独「审核异常，请稍后重试」——实现前与产品确认（默认：固定系统文案，非 Green reason）。
- comment/chat 是否本变更一并两阶段化，或仅 profile+post——**建议一并改**；若工期紧可在 tasks 拆为 4.1/4.2 子项。
