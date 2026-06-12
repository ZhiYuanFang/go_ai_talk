## Context

- 现状：帖子 `status=1` 由 `audit_worker` 每 5s 扫描；资料在 Redis `ucg:green:profile:*`；评论无机审；私信 WS 同步 Green（Option C）。
- 约束：变更 **仅 UCG 域**；Green 仍用 `green_client.go`；凭证复用 OSS AK；不新增 `*_test.go`。
- 项目 MQ：HTTP Publisher + HTTP QueueConsumer（Pull + ticker），与 `voice_task_consumer` 同族；消费者部署在 **ucg-service**（禁止 worker 直连 ucg DAO）。

## Goals / Non-Goals

**Goals:**

- 帖/评/资料/聊天机审均经 RabbitMQ 触发；去掉 audit 业务扫表 ticker。
- 帖/评/资料：pending 对公众不可见（Option A）；全实体 CAS + `audit_version` 防并发、重复消费与过期消息写脏。
- 聊天：模式 A（先推后审）；reject 后收件人不可见、发送方见违规；未读不回滚。
- 统一 `ucg audit consumer` 按 routing key 分发；**所有 MQ 载荷 MUST 含 `auditVersion`**。

**Non-Goals:**

- 改造 `recommend_worker`、`chat_persist_worker`、domain_outbox、voice MQ。
- 评论/帖子管理端新 API 的 usage 统计策略（实现前问负责人）。
- 原生 AMQP 长连接 consumer。

## Decisions

### 1. 消费者驻留 ucg-service

Green、DAO、Redis 聊天均在 ucg 进程；worker-service 不消费 ucg 队列。

### 2. Exchange 与队列

- 沿用 topic exchange `voice.events`（与现有 init 脚本一致），新增 binding：
  - `ucg.post.created` → `ucg.post.created.q`
  - `ucg.comment.created` → `ucg.comment.created.q`
  - `ucg.profile.patch.submitted` → `ucg.profile.patch.submitted.q`
  - `ucg.chat.msg.created` → `ucg.chat.msg.created.q`
- `eventkit/routing_keys.go` 增加 `RoutingPrefixUcg` 与注册表项；Publish 前校验合法 key。

### 3. `audit_version` 跨实体语义（MUST）

`audit_version` **仅表示审核轮次/迭代**，与 Green 调用次数对应，**不是**业务数据版本号。

| 实体 | 版本权威列 | MQ 载荷字段 | CAS 条件列 |
|------|-----------|------------|-----------|
| 帖子 | `ucg_post.audit_version` | `postId`, `auditVersion` | `status`, `audit_version` |
| 评论 | `ucg_post_comment.audit_version` | `commentId`, `auditVersion` | `status`, `audit_version` |
| 资料 | `ucg_profile_audit_job.audit_version` | `jobId`, `auditVersion` | `status`, `audit_version`（**仅 job 表**） |
| 私信 | `ucg_chat_message.audit_version` | `messageId`, `conversationId`, `auditVersion` | `audit_status`, `audit_version` |

**递增时机（MUST）：**

- **仅**用户 submit / 再提审时递增（首提审从 1 起始，再提审 `audit_version++` 并将审态置 pending）。
- 消费者 CAS 成功 **MUST NOT** 递增 `audit_version`。

**业务不回滚（MUST）：**

- 再提审 **MUST NOT** 因新版本而 reset 已 published 的业务快照（帖文正文、已 apply 的资料字段等）；仅审态进入 pending 并 bump 版本，等待新一轮 Green。
- 资料 pending 预览读 job 行；公众仍见 `ucg_profile` 旧值直至 CAS 通过并 apply。

**禁止的版本源：**

- 资料：**禁止**用 Redis `ucg:green:profile:*` 或 `ucg_profile` 表列作为 CAS 版本；**必须**读 `ucg_profile_audit_job.audit_version`。
- 聊天：Redis 消息 JSON 中的 `audit_version` **必须**与 `ucg_chat_message.audit_version` 一致（镜像）；CAS 以 DB 行为准，Redis LSET 在 CAS 成功后同步。

**过期消息（MUST）：**

- 用户再提审后 `audit_version` 已递增；队列中旧 `auditVersion` 的消息 CAS `RowsAffected=0` → **ACK、跳过重试**，MUST NOT 覆盖新轮次状态。

### 4. 可靠投递：发帖/评论/资料用「事务 + 直发 MQ + 兜底」

- 单事务写入业务行（+ media / audit job）后 `Publish`（载荷含当前行 `audit_version`）。
- Publish 失败：记录 error 日志 + 指标；提供 **低频 reconciler**（如启动时 + 每 30min）扫描 `status=1` 且超时未审行，补发 MQ（**补发 MUST 读库内当前 `audit_version`**）。
- 聊天：Redis 写入与 WS 投递成功后 Publish；失败同样可 reconciler 按 `audit_status=pending` 补发（读 `ucg_chat_message.audit_version` 或 outbox 镜像值）。

备选（未采用）：MySQL outbox 表 relay——与 chat outbox 类似但本变更优先直发 + reconciler，减少新表数量。

### 5. CAS 统一 helper

```text
CasAuditTransition(table, id, fromStatus, fromVersion, toStatus, extraFields)
→ UPDATE ... SET status=toStatus, ... WHERE id=? AND status=fromStatus AND audit_version=fromVersion
→ RowsAffected==0 → 幂等/过期消息，ACK 不重试，记录 info/debug 日志
```

- 帖/评/job：`status` INT 1/2/3 + `audit_version` INT。
- 聊天：`audit_status` ENUM pending/approved/rejected + `audit_version` INT；WHERE 使用 `audit_status` 而非 `status`。
- 消费者 **禁止**在 CAS SET 子句中 `audit_version=audit_version+1`。

**禁止**业务代码直接 `Data(g.Map{status: ...}).Update()` 无 WHERE 版本条件（admin 驳回走专用 CAS 路径，允许 1|2→3，仍带 `audit_version` 条件）。

### 6. 帖子与评论（Option A 公众不可见）

- Create/Update submit：`status=1`，再提审时 `audit_version++`，Publish（`auditVersion` = 行内当前值）。
- Consumer：`auditPost`/`auditComment` 逻辑复用现有 Green 序列；结果经 CAS 写 `status`（2/3）+ `reject_reason`，**不递增** `audit_version`。
- 评论：`comment_count` 与 `NotifyOnComment` **仅在 CAS 到 published 后**执行（与现网立即 +1 不同）。

### 7. 资料：MySQL audit job 替代 Redis

表 `ucg_profile_audit_job`（字段含 wx_id、nickname、avatar_key、bio、status、**audit_version**、reject_reason、timestamps）。  
`UpdateMyProfile` 插入 job（`audit_version=1` 或再提审递增）+ Publish（`jobId`, `auditVersion`）；Consumer Green 通过后 **事务内** CAS job（`WHERE status=1 AND audit_version=?`）+ apply `ucg_profile`。  
**禁止**从 Redis 或 `ucg_profile` 读取 CAS 用版本。  
删除 Redis `ucg:green:profile:pending` 写入路径；`mergeProfileForAuthor` 改读最新 pending job。

### 8. 聊天模式 A

**发送路径（WS，同步）：**

1. `message_ack`
2. 写入/预留 `ucg_chat_message` 行（或 outbox 含 `audit_status=pending`、`audit_version=1`）；Redis RPUSH 消息 JSON **镜像**同 `audit_version`
3. `unread[recipient]++`（不回滚）
4. WS `message_delivered` → 收件人（及发送方 echo 可选）
5. `enqueueChatMessageOutbox`（persist 带 pending + version）
6. Publish `ucg.chat.msg.created`（含 `auditVersion` 来自 `ucg_chat_message`）

**Consumer：**

- Green 文本/图/视频（CDN URL）
- pass：CAS `audit_status=pending→approved`（`WHERE audit_status='pending' AND audit_version=?`），**不递增** version
- fail：CAS `pending→rejected`，写 `reject_reason`；WS `audit_failed`→sender；`msg_hidden`→recipient（在线）

**列表：**

- 收件人：`audit_status!=rejected`
- 发送人：含 rejected，带违规文案

Redis LIST 更新：CAS 成功后按 `msgId` 定位 LSET 更新 JSON；列表 API 以过滤规则为准。

### 9. 移除 audit_worker

删除 `runAuditTick` 中帖/资料分支；`main.go` 改为 `StartUcgAuditMQConsumer`；若 profile 已完全 MQ 化则删除 `StartAuditWorker` entirely。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| DB 提交后 MQ 丢失 | reconciler 补发（读当前 audit_version）；监控 Publish 失败 |
| 聊天先推后审合规 | 产品/法务确认；reject 后收件端隐藏 |
| Redis LIST CAS 复杂 | MySQL `ucg_chat_message` 为审态权威；Redis 懒修正 |
| 重复/过期消费 | audit_version + status/audit_status CAS；RowsAffected=0 ACK |
| Green 长时间 pending（聊天） | 消息对收件人长期可见（已接受） |
| 客户端 BREAKING | 版本说明；WS 新事件 |

## Migration Plan

1. 部署 DDL（各表 `audit_version`、评论列、profile job 表、chat message 审态列）。
2. 存量 `ucg_post.status=1`：reconciler 补发 `ucg.post.created`（**读列 `audit_version`，缺省 1**）。
3. 部署 ucg-service（consumer on）+ rabbitmq-init 新队列。
4. 确认无 pending Redis profile 或脚本导入 job 表。
5. 回滚：关 consumer、恢复 audit_worker 旧镜像（需保留 DDL 兼容或回滚 migration）。

## Open Questions

- 管理端批量驳回帖子是否允许从 `published(2)` → `rejected(3)`（当前 admin 逻辑支持非 rejected 均可驳）——实现保持，CAS 用 `status IN (1,2)` + 当前 `audit_version`。
- `ucg-service` 是否在 API 进程 fail-fast RabbitMQ：建议 consumer 启动时探活，Publish 失败不阻断发帖 API（与 voice TextChat 一致，warning + reconciler）。
