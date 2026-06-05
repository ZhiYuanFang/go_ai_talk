## Context

- 部署目标：**2C2G ECS**，MySQL 与 Docker 同机；**生产 + 测试双栈常开**；测试需 **真实 ASR / WebSocket 长连接** 验收。
- 约束：**ASR 验收时不停止生产微服务**；**短期不升配 4G**；用户接受 **生产 Redis 缩至 3 主 0 从**、**测试 Redis 单机 1 容器**。
- 现状：生产 `docker-compose.redis-cluster.yml` 6 节点；测试 `docker-compose.redis-cluster.test.yml` 6 节点；config yaml 三主种子 `redis-node-1:7001,...7003`；Compose **无** 容器资源 limits。
- ASR 走百度云端 API，voice 进程主要为连接/缓冲，非本地模型；瓶颈在 **容器数量与 Redis 内存占用**。

## Goals / Non-Goals

**Goals:**

- 生产 Redis：**3 主 0 从** Cluster（7001–7003），与现有 yaml 种子地址一致，无需改 Go 业务代码。
- 测试 Redis：**单机** `redis-test:6379`，无 `cluster create`。
- prod/test **全部容器** 配置 `mem_limit` / `cpus`（2G survival 起步值，runbook 可调）。
- runbook 覆盖：迁移、MySQL 256M 级调优建议、ASR 验收约定、OOM 排错。

**Non-Goals:**

- 升级 ECS 规格、引入云 Redis/RabbitMQ 托管。
- 生产 Redis 改为单机（用户已选 3 节点 cluster）。
- 修改 Redis key 前缀或应用缓存语义。
- 本地开发强制改为 3 节点（`redis-cluster.yml` 可与生产对齐为 3 节点，或文档说明本地仍可用 6 节点 profile；实现时 **生产 compose 改 3 节点**，本地 runbook 同步）。

## Decisions

### 1. 生产 Redis：3 主 0 从（非单机）

- **选择**：保留 Cluster 协议与三主种子地址，去掉 4–6 从节点。
- **理由**：与 `config*.yaml` 已有 `redis-node-1..3` 一致；比 6 节点省 3 容器（~384MB+）；比完全单机保留 slot 分布习惯。
- **备选**：生产也单机 — 更省内存但偏离 cluster 运维习惯；用户已拒绝必须 6 节点但未拒绝 cluster 形态。

### 2. 测试 Redis：standalone + `GF_REDIS_DEFAULT_ADDRESS`

- **选择**：新 compose 文件 `docker-compose.redis-standalone.test.yml`，服务名 `redis-test`，端口 6379；`.env.test` 设 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`；基线 `microservices.yml` 各服务可选 `${GF_REDIS_DEFAULT_ADDRESS:-}`（空则走 yaml cluster 地址）。
- **理由**：单机不走 cluster 客户端；与 prod overlay「不写 environment」策略兼容（注入走基线 + .env）。
- **备选**：test overlay 写 environment — 已明确避免 asymmetry。

### 3. 资源 limits：独立 `resources.{prod,test}.yml` overlay

- **选择**：`mem_limit` + `cpus` 写在独立 overlay，启动命令追加 `-f docker-compose.resources.prod.yml`。
- **理由**：基线/local 不受限；prod/test 配额不同（voice-test 512M）。
- **备选**：`deploy.resources` — 非 Swarm 兼容性参差；Linux 上 `mem_limit`/`cpus` 更直接。

### 4. 2G 起步配额（实现默认值，runbook 为准）

| 组件 | memory | cpus |
|------|--------|------|
| prod redis ×3 | 96m | 0.1 |
| test redis-test | 96m | 0.1 |
| rabbitmq ×2 | 192m | 0.2 |
| voice-test | 512m | 0.8 |
| voice-prod | 256m | 0.3 |
| gateway / gateway-app | 192m | 0.2 |
| 其它微服务 | 128m | 0.15 |

- limits 总和可 > 2G（防单容器暴涨）；宿主机建议 **swap 1G**。

### 5. ASR 验收运营约定（文档，非代码）

- 生产 **7 微服务保持 Up**；验收窗口 **仅对 test 域名做 ASR 压测**，避免 prod 同时语音并发。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 生产 Redis `down -v` 丢缓存 | 维护窗口；版本/会话缓存可重建 |
| 3 主无副本，单节点故障丢 slot | 2G survival 接受；将来升配可扩回 3 主 3 从 |
| limits 过紧导致 voice OOM | 以 `docker stats` 微调；voice-test 最高优先级 |
| 2G 仍偶发 OOM | runbook：swap、MySQL buffer_pool 256M、禁止双栈 ASR 并发 |
| 测试从 6 节点 cluster 迁移 | 文档化 `down -v` + 新 standalone up |

## Migration Plan

1. **文档/compose 合并** 后，在维护窗口：
2. 生产：`docker compose -f redis-cluster.yml down -v` → 更新 compose（3 节点）→ `up -d` → `cluster create` 三节点 `--cluster-replicas 0` → 验 `CLUSTER INFO`。
3. 测试：`docker compose -f redis-cluster.test.yml down -v` → `redis-standalone.test.yml up -d` → 更新 `.env.test` → recreate 微服务。
4. 叠加 `resources.*.yml` recreate 全栈。
5. **回滚**：保留旧 compose 文件 git tag；Redis 回滚仅 `down -v` 恢复旧拓扑（再次丢数据）。

## Open Questions

- 本地 `docker-compose.redis-cluster.yml` 是否同步改为 3 节点（建议 **是**，与生产一致）或保留 6 节点仅文档标注「本地可选」— 实现 tasks 中默认 **同步 3 节点**。
