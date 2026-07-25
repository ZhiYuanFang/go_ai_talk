# 内存规格与用户规模对照（单 prod + 本机 MySQL）

适用：生产与测试 **分机部署**（prod 在 4C8G；test 可留旧 2G）或历史同机单栈；MySQL 跑在 **与对应栈同一台 ECS** 宿主机；Docker Compose 部署。

配置来源：

- 微服务 `mem_limit` / `GOMEMLIMIT`：`manifest/docker/docker-compose.resources.prod.yml`（**4C8G prod-only 默认**）
- Redis Cluster：`manifest/docker/docker-compose.redis-cluster.yml`
- RabbitMQ：`manifest/docker/docker-compose.rabbitmq.yml`（192m）
- 测试栈 limits：`docker-compose.resources.test.yml`（旧机 2G，**不随 prod 升配**）

---

## 1. 当前默认档（4C8G ECS · prod 单栈）

宿主机约 **4 核 8G**；生产独占本机 MySQL + Compose；**不与 test 双栈同机**。

### MySQL（首日）

```ini
innodb_buffer_pool_size = 1G
max_connections = 100
# 可观测稳定后再考虑 1.5G～2G；旧机 test MySQL 维持约 256M，不跟调
```

### 容器 mem_limit 一览（与 compose 一致）

| 组件 | mem_limit | GOMEMLIMIT / maxmemory | 备注 |
|------|-----------|------------------------|------|
| voice-service | 640m | 550MiB | 语音 WS / 流式峰值 |
| gateway / gateway-app | 256m | 220MiB | |
| ucg-service | 256m | 220MiB | Feed 快照组装 |
| history / device | 192m | 165MiB | |
| sim-user-service | 128m | 110MiB | 推广前可不部署 |
| notify / mcp | 96m | — / 110MiB | |
| rabbitmq | 192m | — | idle ~100MB |
| redis-node ×3 | 256m | **200mb**/节点 | `mem_swappiness: 0` |

**idle 粗算**：MySQL ~1.1–1.3G（含 1G buffer）+ 容器常态远低于 limits + OS → `available` 目标 **> 1.5G**，swap 接近 0。

---

## 2. 规格档位与用户量（粗估）

以下为 **语音 + UCG 广场** 混合产品；实际以监控为准，用户数仅作采购参考。

### 档位 A — 2G ECS（历史 survival / 仅作对照）

曾用于同机双栈或单 prod @ 2G；compose 旧值为 voice 384m、Redis maxmemory 96mb/节点、MySQL buffer 192–256M。现 **prod 默认已迁出该档**；旧机若只跑 **test**，MySQL 维持约 **256M**，使用 `resources.test.yml`。

| 指标 | 建议上限（单 prod @ 2G 时） |
|------|------------------------------|
| DAU | **~200–300** |
| 峰值并发语音 | **~10–15 路** |
| 已发布帖（Redis snapshot） | **~3,000** |
| 同机 test 栈 | **关闭**（或整机只留 test） |

**升级信号（出现任二项）**：

- `available` 持续 **< 200MB**
- swap 持续 **> 350MB**
- voice RSS 峰值 **> 280MB** 或 OOM 137
- Redis `evicted_keys` 持续增长
- MySQL 慢查询明显增多且不宜再降 buffer_pool

### 档位 B — 4G ECS

| 指标 | 可容纳（粗估） |
|------|----------------|
| DAU | **~300–1,500** |
| 峰值并发语音 | **~20–40 路** |
| 已发布帖 | **~10,000** |

**建议调整（相对历史 2G survival）**：

| 组件 | 4G 建议 |
|------|---------|
| voice-service | 512–768m / GOMEMLIMIT 450–680MiB |
| ucg / gateway 类 | 256m |
| redis/节点 | 256m / maxmemory **200mb** |
| MySQL buffer_pool | **512M** |
| 可选 | 短期恢复 test 栈做验收 |

### 档位 C — 4C8G prod-only（当前推荐采购档）

| 指标 | 说明 |
|------|------|
| DAU | **~1,500+** 或峰值语音 **~40+ 路**（粗估，盯监控） |
| 架构 | **prod 独占新机**；MySQL 本机；test 可留旧机 |
| MySQL buffer | **首日 1G**（见 §1）；旧机 test **不跟调** |
| compose | 即仓库默认 `resources.prod.yml` + `redis-cluster.yml`（§1 表） |

更高规模再考虑：MySQL 迁 RDS / 第二台，或 voice 独立节点。

---

## 3. 并发语音估算（用于对照档位）

```
峰值语音并发 ≈ DAU × 同时在线率 × 正在使用语音比例

示例（偏乐观）：
  300 DAU × 3% × 30% ≈ 3 路   → 2G 宽裕
  800 DAU × 4% × 35% ≈ 11 路  → 2G 临界，盯 voice
  1500 DAU × 5% × 40% ≈ 30 路 → 需 4G+
```

每路活跃语音粗估 **5–15MB** voice 进程 RSS（流式缓冲差异大）。

---

## 4. 日常巡检命令

```bash
free -h
docker stats --no-stream --format "table {{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}"
ps aux --sort=-%mem | head -5

# Redis 驱逐
docker exec go-ai-talk-redis-redis-node-1-1 redis-cli -p 7001 INFO stats | grep evicted
docker exec go-ai-talk-redis-redis-node-1-1 redis-cli -p 7001 INFO memory | grep used_memory_human

# MySQL
mysql -e "SHOW VARIABLES LIKE 'innodb_buffer_pool_size';"
mysql -e "SHOW GLOBAL STATUS LIKE 'Threads_connected';"
```

---

## 5. 应用新 limit

```bash
# 仓库根目录；prod 微服务 + resources overlay（默认即 4C8G 档）
docker compose --env-file manifest/docker/env/.env.prod \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.prod.yml \
  -f manifest/docker/docker-compose.resources.prod.yml \
  up -d --force-recreate

# Redis cluster（改 maxmemory / mem_swappiness 后需 recreate 各 node）
docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate
```

验证 Redis cgroup 未倾向 swap（recreate 后）：

```bash
docker inspect go-ai-talk-redis-redis-node-1-1 --format '{{.HostConfig.MemorySwappiness}}'
# 期望输出 0
```

生产换机切流步骤见 **`docs/runbooks/release-deploy-and-run.md` §C.5**。
