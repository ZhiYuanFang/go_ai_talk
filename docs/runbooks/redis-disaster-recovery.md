# Redis 容灾与恢复 Runbook

适用范围：本项目 Docker 部署的 **测试 standalone Redis** 与 **生产 3 节点 Redis Cluster**。  
日常部署步骤见 [release-deploy-and-run.md](./release-deploy-and-run.md)。

---

## 1. 先分清：重启 vs 数据丢失

| 场景 | volume 是否保留 | 典型操作 | 数据是否还在 |
|------|----------------|----------|-------------|
| 容器 stop / restart | ✅ 是 | `docker compose up -d`、`--force-recreate` | **通常还在**（AOF + volume） |
| 仅 Redis 进程崩溃 | ✅ 是 | 自动/手动 restart | **通常还在** |
| 删除 volume | ❌ 否 | `docker compose down -v` | **Redis 内数据丢失** |
| 磁盘损坏 / 误删 volume | ❌ 否 | — | **Redis 内数据丢失** |
| 内存打满 LRU 驱逐 | volume 在 | 长期 `used_memory` 顶满 | **部分 key 被删** |

**原则**：日常重启 **不要** 使用 `down -v`；生产 Cluster **不要** 重复执行 `cluster create`（卷里已有元数据时）。

---

## 2. 本项目 Redis 持久化配置

### 测试（standalone）

- Compose：`manifest/docker/docker-compose.redis-standalone.test.yml`
- 项目名：`go-ai-talk-redis-test`
- 容器：`go-ai-talk-redis-test`
- 地址：`redis-test:6379`（宿主机 `16379`）
- Volume：`go-ai-talk-redis-test_redis-test-data` → 容器 `/data`
- 持久化：`--appendonly yes`（AOF）

### 生产（Cluster 3 主）

- Compose：`manifest/docker/docker-compose.redis-cluster.yml`
- 项目名：`go-ai-talk-redis`
- 节点：`redis-node-1` ~ `redis-node-3`，端口 `7001`–`7003`
- Volume：各节点 `redis-node-*-data` → `/data`
- 持久化：`--appendonly yes`

AOF 文件位于各容器 `/data` 目录（如 `appendonly.aof`）。

---

## 3. 日常重启恢复（volume 未删）

### 测试 Redis

```bash
cd /path/to/go_ai_talk

docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test redis-cli PING
# 期望 PONG

docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test redis-cli INFO memory | grep used_memory
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test redis-cli DBSIZE
```

### 生产 Redis Cluster

```bash
cd /path/to/go_ai_talk

docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok —— **不要**再跑 cluster create

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 PING
```

若 `cluster_state:ok`，直接重启依赖 Redis 的微服务即可。

**误报 `Node ... is not empty`**：说明卷里已有数据，**忽略 cluster create**，见 [release-deploy-and-run.md §D.4b](./release-deploy-and-run.md#d4b-生产-redisnode--is-not-empty重复-cluster-create)。

---

## 4. 数据分层：丢了 Redis 还能恢复什么

| 数据 | 存储 | Redis 丢/volume 删后 |
|------|------|---------------------|
| 历史记录 history | MySQL 权威 + Redis 短 TTL 缓存 | ✅ 自动回源 MySQL |
| 设备/用户主数据 | MySQL | ✅ 不受影响 |
| UGC 帖子/评论/通知 | MySQL | ✅ 不受影响 |
| **UCG 私信正文** | Redis 热读 + **MySQL `ucg_chat_message` 权威** | ✅ 读时 MySQL fallback，按需 warm Redis |
| 语音多轮会话 | Redis | ❌ 上下文丢失，新会话开始 |
| App refresh token | Redis（TTL） | ❌ 用户重新登录 |
| Pub/Sub 通知 | 无持久化 | ❌ 重启期间通知不补发（MySQL 数据仍在） |
| Outbox 投影标记 | Redis 24h | ⚠️ worker 可重投影 |

UCG 私信在启用 MySQL 持久化后：**Redis volume 丢失不会丢私信正文**，用户拉历史时会从 MySQL 读出并可选回填 Redis（见 §6）。

---

## 5. Volume 备份与还原

### 5.1 备份（维护窗口）

**测试 standalone**（示例，在项目根目录执行）：

```bash
docker run --rm \
  -v go-ai-talk-redis-test_redis-test-data:/data:ro \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/redis-test-$(date +%F-%H%M).tar.gz -C /data .
```

**生产 cluster 单节点**（每节点各备一份，或维护窗口内逐个备份）：

```bash
docker run --rm \
  -v go-ai-talk-redis_redis-node-1-data:/data:ro \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/redis-node-1-$(date +%F-%H%M).tar.gz -C /data .
```

备份前 Redis 仍在写入时，AOF 备份为 **point-in-time**，一般可接受；重要变更前可先 `SAVE`（会阻塞，仅小实例）或停写窗口。

### 5.2 还原

1. `docker compose ... down`（**不加 `-v`** 时先确认是否要覆盖现有卷）
2. 若需从备份恢复：将 tar 解到对应 Docker volume 的 mount 点，或新建 volume 后挂载解压
3. `docker compose ... up -d`
4. 生产执行 §3 验收（`cluster_state:ok`）

**禁止**在已有 cluster 元数据的 volume 上盲目 `cluster create`。

---

## 6. UCG 私信：Redis 空时的应用层恢复

实现见 `ucg-chat-mysql-persist` 变更：

1. 发消息：Redis 先写 + MySQL outbox 同步 → worker 落 `ucg_chat_message`
2. 读消息：Redis 优先；`LLEN=0` 且 MySQL 有记录 → 查 MySQL
3. 会话消息数 ≤ `UCG_CHAT_WARM_MAX_MESSAGES`（默认 200）时，读时 **lazy warm** 全量 RPUSH 回 Redis，并 `SET seq = MAX(id)`

**运维无需手工 SCAN 重建**，用户访问会话即触发回填。

验证（测试环境）：

```bash
# 删除某会话 Redis 列表（模拟丢 key）
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test \
  redis-cli DEL "ucg:chat:conv:123:msgs"

# App 拉该会话历史 → 应仍能看到消息；再次 LLEN 可能 >0（已 warm）
```

---

## 7. 何时必须 `down -v` 重建

仅在 **刻意重置** 或 cluster 元数据损坏且无法修复时：

```bash
# 测试
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml down -v

# 生产（维护窗口，会清空 Redis 全部数据）
docker compose -f manifest/docker/docker-compose.redis-cluster.yml down -v
# 然后 up + 三节点 cluster create（见 release-deploy-and-run §C.1）
```

**后果**：Redis 内所有非 MySQL 备份数据丢失；UCG 私信依赖 MySQL fallback；语音会话、refresh token 等需用户侧重连/重登。

---

## 8. 排查命令速查

```bash
# 内存与驱逐风险
redis-cli INFO memory
redis-cli CONFIG GET maxmemory-policy

# UCG 私信 key 抽样
redis-cli --scan --pattern 'ucg:chat:conv:*:msgs' | head
redis-cli LLEN ucg:chat:conv:CONV_ID:msgs

# MySQL 侧消息数（ucg 库）
# SELECT COUNT(*) FROM ucg_chat_message WHERE conversation_id = ?;
```

---

## 9. 相关文档

- [release-deploy-and-run.md](./release-deploy-and-run.md) — 日常部署、Redis Cluster 验收、D.4/D.4b 排障
- [redis-cluster-local.md](./redis-cluster-local.md) — 本机 Cluster 说明
- DDL：`hack/sql/ucg_chat_persist.sql` — 私信 MySQL 表
