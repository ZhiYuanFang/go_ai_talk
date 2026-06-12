## Context

当前架构：

```
AddHistory / device admin 写 MySQL
  → 同步 patch Redis（主路径，毫秒级）
  → HTTP EnqueueDomainOutbox → worker 轮询 → ApplyProjection（兜底，~1.5s）
  → 可选 MQ fan-out（仓库内无 consumer）
```

探索结论：voice 回复与 App Pub/Sub 不等待 worker；stale hit 窗口由 list TTL 60s / latest 30s 界定；device 字典/画像已在写路径 `refreshEventOptionsCacheAfterMutate` / `setUserProfile` 同步更新。

UCG 域已在 `ucg-service` 内运行经 OpenSpec 批准的 outbox relay 与 AMQP consumer，与 worker 无关。

约束：`AGENTS.md` 原要求 history/device 经 `WORKER_SERVICE_URL` 入队；本变更删除该路径。Redis 读缓存约定仍适用——本变更 **不新增** Redis 键，仅加强同步 patch。

## Goals / Non-Goals

**Goals:**

- 彻底删除 `worker-service` 及 domain_outbox / workeroutbox / voice.task 脚手架。
- history/device 缓存一致性由 **写后同步 patch + 读 miss 回源 MySQL + TTL** 保证。
- 加强 `patchHistoryOnAdd`：冷缓存写 `latest`；`setHistoryList` 失败时删除 list key 促使下次 miss 重建。
- 在 `AGENTS.md` 与 `openspec/project.md` 写入 **循环后台任务治理** 约定。
- 更新部署、runbook、RabbitMQ init，移除 worker 与 orphan MQ 资源。

**Non-Goals:**

- 不修改 UCG 审核/推荐/聊天 persist 等已批准的后台任务。
- 不引入 Feed Redis ZSET 或新读缓存。
- 不强制 DROP `domain_outbox` 表（文档标注 deprecated 即可）。
- 不新增测试文件。

## Decisions

### D1：同步 patch 为唯一写后投影路径

**选择**：删除 outbox relay；history `AddDeviceHistory` / Update / Delete 与 device admin mutate 继续在同事务或写库成功后同步更新 Redis。

**理由**：用户热路径已依赖同步 patch；worker 仅重复 patch 或延迟补洞。

**备选**：保留 outbox 但 relay 迁入 history-service——仍引入 ticker，与「减少循环后台任务」目标冲突。**拒绝**。

### D2：可接受的一致性窗口

**选择**：允许 list cache **hit** 时最长约 **60s**、latest **30s** 的 stale 窗口；通过 D3 缩短典型失败场景窗口。

**理由**：与 explore 阶段产品假设一致；MySQL 始终权威。

### D3：加强同步 patch 以降低 stale hit

**选择**：

1. `patchHistoryOnAdd`：列表 cache miss 时仍 `setLatestHistory(item)`。
2. `setHistoryList` 失败时 `DEL` list key（best-effort），避免长期 hit 旧列表。
3. Update/Delete 对称考虑：patch 失败时 bump piece epoch 仍保留；list patch 失败则 DEL list key。

**理由**：无 worker 时减少「等 TTL」依赖；改动面小于引入新后台任务。

### D4：删除 voice.task.* 全链路

**选择**：删除 `publishTaskRequested`、`VoiceTaskConsumer`、相关 routing key 注册与 rabbitmq-init binding。

**理由**：consumer 为 stub；TextChat 不依赖其业务语义；保留仅增加运维噪音。

**备选**：保留 best-effort 发布无 consumer——**拒绝**（无消费者队列堆积无意义）。

### D5：循环后台任务治理

**选择**：新增 OpenSpec capability `background-loop-task-governance`；在 `AGENTS.md` / `project.md` 强制：新增 `NewTicker`、扫 outbox/表 reconciler、HTTP Pull 消费队列等 **MUST** 有 OpenSpec proposal/design 明确批准；**AMQP push consumer**（broker 驱动）不视为 ticker 扫表，但仍须在变更中声明。

**理由**：去掉 worker 后防止 history/voice 等域再次 silently 增加 ticker。

### D6：MQ 拓扑清理范围

**选择**：删除 `history.record.*`、`device.event.changed` 等 **仅 worker fan-out** 的 binding；**保留** UCG、`voice.events` 上 UCG 所用 binding。

**理由**：避免无人消费的队列；降低 RabbitMQ 运维面。

### D7：history `ReconcileProjectionCachesForWorker` HTTP

**选择**：删除 worker 调用的 projection reconcile HTTP 端点（若仅 worker 使用）；`RebuildHistoryCacheByDevice` 保留为 **手动运维** 工具函数（无自动 ticker），文档说明用法。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 同步 patch Redis 失败且 list 仍 hit 旧数据 | D3 DEL list key；TTL 60s 上限；运维可手动 rebuild |
| 多副本 history 写 patch 均失败 | 读 miss 回源 MySQL 仍正确；监控 patch 失败日志 |
| 规格与 v2.0.3 冲突 | 本变更 spec delta 明确 MODIFIED/REMOVED |
| 环境仍配 `WORKER_SERVICE_URL` | 删除 env 与 deployment 引用；代码无 enqueue |
| UCG 与「禁止后台任务」混淆 | governance 明确 UCG 已批准项 + AMQP push 例外 |

## Migration Plan

1. 合并代码并部署 **无 worker** 的版本（history/device/voice/gateway 新版本 + worker 缩容为 0）。
2. 全环境 `OUTBOX_RELAY_ENABLED=false`（或移除变量）；删除 `WORKER_SERVICE_URL`。
3. 可选：RabbitMQ 删除 orphan queue/binding（`voice.task.*`、`history.events.q` 等若存在）。
4. `ai_voice_worker` 库保留只读备份后标记 deprecated；无需阻断 rollout。
5. **Rollback**：恢复 worker deployment 与 env（不推荐长期）；代码回滚需恢复 enqueue 路径。

## Open Questions

- 生产是否曾依赖 `CACHE_PROJECTION_REPAIR_ENABLED`？本变更删除 repair worker；若曾开启，确认可接受仅 TTL/miss 自愈。
- `history.events.q` / `notify.events.q` 是否有集群外消费者？若无，一并从 rabbitmq-init 移除。
