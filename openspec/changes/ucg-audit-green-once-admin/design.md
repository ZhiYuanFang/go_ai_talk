# 设计

## Context

UCG 审核由 `ucg-service` AMQP consumer 驱动（`audit_mq_consumer.go` → `audit*FromEvent`）。handler 返回非 nil → `Nack(requeue=true)` 无限重投。

当前状态（实现调研）：

| 实体 | 两阶段 | verdict 跳过 Green | Green 失败行为 |
|------|--------|-------------------|----------------|
| profile | ✅ | ✅ `verdict≠0` return | `markProfileModerationFailed(5)`，Phase1 void，但 Green err 后仍可能 persist |
| post | ✅ | ✅ | `return err` → **重复 Green** |
| comment | ❌ | — | Green err → requeue 风暴 |
| chat | ❌ | — | 同 comment |

资料 `ProfileJobStatusModerationFailed=5` 已定义；作者 `mergeProfileForAuthor` 不查 status=5，作者无感知。管理页仅有「动态审查」，无资料 job 队列。

约束（`openspec/project.md` / AGENTS.md）：不新增测试文件；不默认 Redis 读缓存；UCG AMQP consumer 已批准；跨服务边界不变；admin API 不计 usage 统计。

## Goals / Non-Goals

**Goals:**

- 对每个 `(entity_id, audit_version)`，Green API **业务意义上至多一次**；MQ 重投、apply 重试、CAS 竞争均不得再次调 Green。
- profile/post/comment/chat 共用 Phase1/Phase2 编排模式（`audit_moderation.go` 扩展）。
- Green 或 persist verdict 失败 → 机审失败终态 + Ack（profile 已有；post/comment/chat 对齐）。
- UCG 管理页展示 profile `moderation_failed` job，支持人工 approve/reject。
- 作者 App **不**展示 moderation_failed 专用状态或文案。

**Non-Goals:**

- 帖子/评论/私信 moderation_failed 的管理页（后续变更）。
- Green fail-open 自动放行（本变更采用 **fail-closed + 人工复核**）。
- 修改 Green SDK、`green_client.go` 判定逻辑。
- 回填历史超额 Green 账单。

## Decisions

### 1. 「Green 已发起」判定（统一规则）

**选择**：对每个 `(id, audit_version)` Phase1 入口顺序检查：

```
if moderation_verdict != 0 → skip Green（仅 Phase2 或 noop）
if status 已机审失败终态（profile/post moderation_failed）→ skip Green，Ack
else → 调 Green；结果写入 verdict 或进入机审失败终态
```

Green **API 返回 err**（网络/额度/5xx）：**不再** Nack 重试 Green；写入机审失败终态（profile/post `moderation_failed`；comment/chat 同族 status 或 `audit_status` 扩展），handler 返回 nil → Ack。

**理由**：用户明确要求「任意 Green 发起后不得重复调用」；比 `fix-ucg-audit-mq-green-retry-storm` 设计 §3「verdict 未落库可重试 Green」更严。

**备选**：仅 `moderation_verdict` 防重 — 拒绝，API err 时 verdict 仍为 0 会 requeue 重调。

### 2. 资料 Phase1 收口

**选择**：

- `runProfileModerationPhase`：`verdict≠0` 已 `return`；Green err 或 persist err → **仅** `markProfileModerationFailed`，**不再**继续 `persistModerationVerdictProfile`。
- `markProfileModerationFailed`：CAS `WHERE id=? AND status=pending`，写 `status=5`、`reject_reason=运维日志`。
- `auditProfileJobFromEvent`：Phase1 后不依赖 err；Phase2 见 `status≠pending` 直接 skip。

**作者侧**：不读 status=5；作者见已发布 profile（patch 未生效直至人工通过）。

### 3. 帖子 Phase1 对齐

**选择**：

- 新增 `PostStatusModerationFailed = 5`（与 profile 数值对齐，注释区分实体）。
- `runPostModerationPhase` 改 void 语义：Green/persist err → `markPostModerationFailed`，不 return err。
- `auditPostFromEvent` Phase1 后不因 Phase1 失败 requeue。

### 4. 评论 / 私信两阶段化

**选择**：DDL 为两表增加与 `ucg_post` 同族字段：

| 字段 | comment | chat |
|------|---------|------|
| moderation_verdict | int | int |
| moderation_reason | varchar | varchar |
| moderation_at | int64 | int64 |
| apply_attempts | int | int |
| apply_failed_at | int64 | int64 |

可选终态：`CommentStatusModerationFailed=5`；chat 增加 `audit_status=moderation_failed` 字符串或 int 映射（实现时与现有 `pending|approved|rejected` 一致扩展）。

**Phase1**：`runCommentModerationPhase` / `runChatModerationPhase` + `persistModerationVerdict*`（CAS `moderation_verdict=0`）。

**Phase2**：复用现有 `publishCommentCAS` / `approveChatMessageCAS` + apply 有界重试（`UCG_AUDIT_APPLY_MAX_ATTEMPTS`）。

**Chat Redis 路径**：MySQL 行存在时以 MySQL 为准；Redis-only 竞态路径在 Phase1 前若仍无 MySQL 行，**不调 Green**，返回 nil（Ack）或短暂 requeue（**仅 DB 未就绪**，非 Green 重试）——实现时在 tasks 中明确：Green 单次以 MySQL `(conversation_id, message_id, audit_version)` 为准。

### 5. 管理端人工复核（仅 profile job）

**API**（`X-Admin-Password`，ucg-service / gateway-app 现有 admin 路由组）：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/ucg/admin/api/profile-audit-jobs/list` | 分页；默认 `status=5`；返回 jobId、wxId、nickname/avatarKey/bio patch、rejectReason、auditVersion、createdAt |
| POST | `/ucg/admin/api/profile-audit-jobs/resolve` | body: `jobId`, `action`: `approve` \| `reject`, `reason?`（reject 必填） |

**Service**：

- `approve`：`CasAuditTransition` `moderation_failed(5) → approved(2)`（或 `pending` 不适用）+ 事务更新 `ucg_profile` 非空 patch 字段（复用 `approveProfileJobCAS` 逻辑，扩展 `FromStatus` 含 5）。
- `reject`：`→ rejected(3)` + `reject_reason`；**不**更新 profile 已发布行。

**UI**（`ucg-admin.html`）：新 Tab「资料机审失败」；表格 + 行内通过/驳回（驳回弹 reason）；操作后刷新列表。

### 6. 共享编排

**选择**：在 `audit_moderation.go` 保留 entity 专用函数，但统一注释与失败语义；可选提取 `markModerationFailed` 模式，避免四个 copy-paste 分叉（实现阶段小步提取即可）。

## Risks / Trade-offs

- **[Risk] Green 瞬态错误直接 moderation_failed** → 作者无感知、内容未更新；**Mitigation**：管理页人工通过；日志 `[ucg-audit-mq]` 可检索。
- **[Risk] comment/chat DDL 遗漏环境** → **Mitigation**：tasks 含 migration + deploy runbook。
- **[Risk] 人工 approve 与并发 MQ** → **Mitigation**：resolve CAS 带 `status=5 AND audit_version`；成功后 Ack 积压消息因 stale skip。
- **[Trade-off] 帖子 moderation_failed 无管理页** → 运维暂靠 SQL/日志；profile 先行验证流程。

## Migration Plan

1. 执行 DDL（comment/chat moderation 字段；post status=5 若需枚举文档化）。
2. 部署 `ucg-service` + gateway-app admin 路由。
3. 观察 Green 调用量、队列深度、`status=5` 积压；管理页处理历史 failed job。
4. **回滚**：还原代码；新字段可保留；若回滚后 post 仍 requeue 风暴，暂停 consumer（runbook 既有小节）。

## Open Questions

- chat `moderation_failed` 用新 `audit_status` 字符串还是 comment 同款 int status（**建议**：`audit_status=moderation_failed` 与 approved/rejected 并列，避免 int 与 post 混淆）。
- 人工 approve 是否写 `moderation_verdict=pass` 补录（**建议**：approve 时写 verdict=pass + moderation_at，便于审计）。
