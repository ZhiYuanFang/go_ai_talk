## Context

- 当前私信流程：`ProcessOutboundChatMessage` → Green 审核 → `DeliverChatMessage` → Redis `INCR seq` + `RPUSH` JSON；MySQL 仅更新 `ucg_conversation_member` 未读数。
- Redis compose 已启用 `--appendonly yes` 与 Docker volume；生产为 3 节点 Cluster，测试为 standalone。
- Redis 配置 `maxmemory 64mb` + `allkeys-lru`，私信 key 可能被驱逐。
- ucg-service 已有 `StartAuditWorker` / `StartRecommendWorker` 轮询模式，可复用为 persist worker。
- AGENTS.md：UCG 域数据在 `ai_voice_ucg`（`UCG_DB_LINK`），跨域不得直连他域库；不借用 worker-service outbox。

## Goals / Non-Goals

**Goals:**

- 私信正文持久化至 MySQL（utf8mb4），Redis 丢失后可从 MySQL 恢复读取。
- 发送路径保持 Redis 先写，用户感知延迟不变。
- outbox 同步写入保证进程 crash 时不丢持久化意图。
- 读路径 Redis 优先；miss 时 MySQL fallback + lazy warm（含 `seq` 对齐）。
- 提供 Redis 容灾 runbook，覆盖重启、volume、AOF 与数据分层。

**Non-Goals:**

- 不做 Redis 存量 SCAN backfill（未上线）。
- 不改为 MySQL 先写（方案 B）。
- 不使用 worker-service `domain_outbox` 或 RabbitMQ 新队列。
- 不修改 WS 协议、Green 审核逻辑。
- 不在此变更内做全库 utf8mb4 迁移（其它表另开变更）；本变更仅保证新私信表 utf8mb4。

## Decisions

### 1. 方案 A：Redis 热读 + MySQL outbox（非 MySQL 先写）

**选择**：Redis 同步写 → MySQL outbox 同步写 → worker 异步 INSERT `ucg_chat_message`。

**理由**：保持现有「发完即读」UX；outbox 一行 INSERT 开销小（~1–3ms）；MySQL 为持久权威。

**备选**：MySQL 先写再 Redis 投影（history 模式）——读路径改动更大，本次不采用。

### 2. outbox 位于 ucg 本域库，worker 在 ucg-service 进程内

**选择**：`ucg_chat_message_outbox` 在 `ai_voice_ucg`；`StartChatPersistWorker` 仿 `StartAuditWorker`。

**理由**：私信是 ucg 本域表；AGENTS.md 禁止为落本域表绕 worker HTTP。

### 3. 消息 ID 与 Redis seq 对齐

**选择**：`ucg_chat_message.id` = Redis `INCR ucg:chat:conv:{id}:seq` 的值；`UNIQUE(conversation_id, id)`。

**理由**：与现有 `LastReadMsgId`、客户端 msgId 语义一致；回填 Redis 时需 `SET seq = MAX(id)`。

### 4. 读时 lazy warm

**选择**：`LLEN=0` 且 MySQL `COUNT>0` 时查 MySQL；默认回填当前 page；可选配置「消息数 < N 则整会话回填」。

**理由**：无需离线全量重建；未打开会话不占 Redis 内存。

**回填时必须**：`SET ucg:chat:conv:{id}:seq` 为 MySQL `MAX(id)`（若 Redis seq 缺失或更小），避免新消息 ID 冲突。

### 5. outbox 失败语义

**选择**：Redis 写成功、outbox 失败 → 记录 **Error** 日志（含 convId、msgId）；不阻断 WS 推送（可用性优先）。

**理由**：极罕见（MySQL 宕机）；可观测 + 后续 repair；若需强一致可后续改为 fail-fast。

### 6. 幂等与重试

**选择**：worker INSERT 使用 `ON DUPLICATE KEY UPDATE` 或等价幂等；outbox 成功后标记 `done`；失败递增 `attempts` 保留 `last_error`。

### 7. Redis 容灾 runbook 独立文档

**选择**：`docs/runbooks/redis-disaster-recovery.md`，与 `release-deploy-and-run.md` 交叉引用。

**内容**：容器重启 vs volume 丢失、AOF/volume 路径、cluster 勿重复 create、数据分层表、UCG 私信恢复依赖 MySQL fallback。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| Redis 成功、outbox 失败窗口 | Error 日志 + 可选 repair worker 对比 Redis/MySQL |
| lazy warm 并发双写 | `SET rebuild lock NX EX 30` 或接受幂等 RPUSH |
| 64MB LRU 驱逐热数据 | MySQL fallback 保证正确性；仅性能降级 |
| outbox 堆积 | worker 轮询间隔可配置；监控 pending 数量 |
| emoji 入库 | 表与 DSN charset=utf8mb4 |

## Migration Plan

1. 在 test/prod `ai_voice_ucg` 执行 DDL（utf8mb4）。
2. 部署 ucg-service（含 persist worker）。
3. 验证：发私信 → Redis + MySQL 均有；重启 Redis（保留 volume）→ 数据仍在；清空 Redis key → 读消息触发 MySQL fallback + warm。

**Rollback**：回滚 ucg-service 至旧版；新表可保留（无历史数据）；读仍走 Redis-only 旧逻辑。

## Open Questions

- outbox 失败是否改为 fail-fast（阻断发送）——当前默认可用性优先，实现时可配置。
- lazy warm 默认「整会话」还是「仅当前页」——默认当前页，可 yaml 配置阈值 N。
