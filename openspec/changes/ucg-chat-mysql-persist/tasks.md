## 1. 数据库与 DAO

- [x] 1.1 编写 `ucg_chat_message`、`ucg_chat_message_outbox` DDL（utf8mb4，`UNIQUE(conversation_id, id)` 等索引）
- [x] 1.2 在 test/prod `ai_voice_ucg` 执行 DDL；更新 `hack/config.yaml` 并生成 DAO/entity/do

## 2. 发送与 persist worker

- [x] 2.1 `chat_store.go` / `chat_service.go`：`DeliverChatMessage` 后同步 INSERT outbox
- [x] 2.2 新建 `chat_persist_worker.go`：轮询 outbox → 幂等 INSERT `ucg_chat_message` → 标记 done
- [x] 2.3 `cmd/ucg-service/main.go`：启动 `StartChatPersistWorker`
- [x] 2.4 outbox 失败 Error 日志（convId、msgId、err）

## 3. 读取与 lazy warm

- [x] 3.1 `listChatMessages`：Redis 优先；`LLEN=0` 且 MySQL COUNT>0 时 MySQL 分页查询
- [x] 3.2 实现 lazy warm：RPUSH 回填 + `SET seq = MAX(id)`；可选 rebuild 锁
- [x] 3.3 `lastMessagePreview` 在 Redis 空时回源 MySQL 最后一条

## 4. Redis 容灾 runbook

- [x] 4.1 新建 `docs/runbooks/redis-disaster-recovery.md`（重启、AOF/volume、备份还原、数据分层、UCG 私信 MySQL fallback）
- [x] 4.2 在 `release-deploy-and-run.md` 增加指向 redis-disaster-recovery runbook 的链接（若尚无）

## 5. 验收

- [x] 5.1 test 环境：发含 emoji 私信 → Redis + MySQL 均有记录
- [x] 5.2 test 环境：删除某会话 Redis msgs key → 拉历史仍可从 MySQL 读出并 warm Redis
- [x] 5.3 test 环境：Redis `docker compose up --force-recreate`（不 `-v`）后数据仍在

> **部署前**：在目标库执行 `hack/sql/ucg_chat_persist.sql`。§5 验收步骤见 `docs/runbooks/redis-disaster-recovery.md` §6 与 §3。
