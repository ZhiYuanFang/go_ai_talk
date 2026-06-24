# 内存规格与用户规模对照（单 prod + 本机 MySQL）

适用：生产与测试 **不同时常驻**（推广后停 test）；MySQL 跑在 **同一台 ECS** 宿主机；Docker Compose 部署。

配置来源：

- 微服务 `mem_limit` / `GOMEMLIMIT`：`manifest/docker/docker-compose.resources.prod.yml`
- Redis Cluster：`manifest/docker/docker-compose.redis-cluster.yml`
- RabbitMQ：`manifest/docker/docker-compose.rabbitmq.yml`（192m）

---

## 1. 当前档（2G ECS · prod 单栈）

宿主机可见内存约 **1.7G** + **1G swap**；MySQL 固定占用约 **350–450MB**（建议 `innodb_buffer_pool_size=192M`~`256M`）。

### 容器 mem_limit 一览

| 组件 | mem_limit | GOMEMLIMIT / maxmemory | 备注 |
|------|-----------|----------------------|------|
| voice-service | 384m | 330MiB | 语音 WS / 流式 STT·LLM·TTS 峰值 |
| gateway | 192m | 165MiB | |
| gateway-app | 192m | 165MiB | |
| ucg-service | 192m | 165MiB | Feed 快照组装 |
| history / device | 128m | 110MiB | |
| sim-user-service | 96m | 85MiB | 推广前建议不部署 |
| rabbitmq | 192m | — | idle ~100MB |
| redis-node ×3 | 128m | **96mb**/节点 | 集群逻辑 ~288MB；`mem_swappiness: 0` 避免 Redis 页被 swap |

**idle 粗算**：MySQL ~400MB + 容器 ~200MB + OS ~200MB → `available` 目标 **> 400MB**，swap **< 200MB**。

### MySQL 与本机 Redis 争抢

Feed 读路径偏 Redis 时，2G 上推荐：

```ini
innodb_buffer_pool_size = 192M   # 或 256M（帖量少、DB 压力大时用 256M）
max_connections = 100
```

---

## 2. 规格档位与用户量（粗估）

以下为 **语音 + UCG 广场** 混合产品；实际以监控为准，用户数仅作采购参考。

### 档位 A — 保持 2G（当前 compose）

| 指标 | 建议上限 |
|------|----------|
| DAU | **~200–300** |
| 峰值并发语音 | **~10–15 路** |
| 已发布帖（Redis snapshot） | **~3,000** |
| 同机 test 栈 | **关闭** |

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

**建议调整（相对 2G compose）**：

| 组件 | 4G 建议 |
|------|---------|
| voice-service | 512–768m / GOMEMLIMIT 450–680MiB |
| ucg / gateway 类 | 256m |
| redis/节点 | 256m / maxmemory **200mb** |
| MySQL buffer_pool | **512M** |
| 可选 | 短期恢复 test 栈做验收 |

### 档位 C — 8G 或拆机

| 指标 | 说明 |
|------|------|
| DAU | **> 1,500** 或峰值语音 **> 40 路** |
| 架构 | prod 独占；MySQL 迁 RDS/第二台；或 voice 独立节点 |

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
# 仓库根目录；prod 微服务 + resources overlay
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

双栈同机时 memory 压力叠加；推广后 **先 down test project** 再评估 prod 余量。
