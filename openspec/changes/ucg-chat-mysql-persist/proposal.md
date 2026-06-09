## Why

UCG 私信正文当前仅存储在 Redis（`ucg:chat:conv:*:msgs`），MySQL 仅有会话元数据。Redis 容器重启、卷丢失或 `allkeys-lru` 驱逐会导致私信不可恢复。产品尚未上线，可在无历史迁移负担下引入 MySQL 持久层，并补充 Redis 容灾 runbook，使运维有明确恢复步骤。

## What Changes

- 新增 MySQL 表 `ucg_chat_message`（utf8mb4）与 `ucg_chat_message_outbox`，私信正文以 MySQL 为持久权威。
- `DeliverChatMessage`：Redis 热写（保持现有 UX）后同步写入 outbox；`StartChatPersistWorker` 异步 flush 至 `ucg_chat_message`。
- `listChatMessages`：Redis 优先；Redis miss 且 MySQL 有数据时回源 MySQL，并按需 lazy warm 回填 Redis（含 `seq` 对齐）。
- 新增 `docs/runbooks/redis-disaster-recovery.md`：说明本项目 Redis 重启、AOF/volume 恢复、数据分层与不可恢复项。
- **不做** Redis 存量消息 backfill（未上线无历史数据）。

## Capabilities

### New Capabilities

- `ucg-chat-mysql-persist`：UCG 私信 Redis 热读 + MySQL outbox 持久与读时回填。
- `redis-disaster-recovery-runbook`：Redis 容灾与恢复运维文档要求。

### Modified Capabilities

- （无）本项目 `openspec/specs/` 尚无 UCG 私信相关基线规格。

## Impact

- `internal/services/ucg/`：`chat_store.go`、`chat_service.go`、新建 `chat_persist_worker.go`
- `cmd/ucg-service/main.go`：启动 persist worker
- `hack/config.yaml` / DAO 生成：`ucg_chat_message`、`ucg_chat_message_outbox` 表
- `ai_voice_ucg` 库：DDL（utf8mb4）
- `docs/runbooks/redis-disaster-recovery.md`（新建）
- 不涉及 worker-service `domain_outbox`、不涉及 WS 协议变更
