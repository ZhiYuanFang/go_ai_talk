## Why

同机双栈（生产 + 测试）部署在 **2C2G ECS** 且 **MySQL 同机** 时，当前 **生产 6 节点 + 测试 6 节点 Redis Cluster** 共 12 个 Redis 容器，叠加 14 个微服务与 RabbitMQ，内存远超物理上限；测试 **真实 ASR / 长连接** 验收时 `voice-service` 峰值进一步吃紧。需要在 **不停止生产微服务、短期不升配** 的前提下，缩小 Redis 拓扑并为全部容器增加 CPU/内存上限，防止单容器异常拖死宿主机。

## What Changes

- **BREAKING（生产 Redis）**：`docker-compose.redis-cluster.yml` 从 6 节点（3 主 3 从）改为 **3 主 0 从** Cluster；`cluster create` 仅含 `redis-node-1..3`，`--cluster-replicas 0`。迁移需 `down -v` 重建，**清空 Redis 数据**（缓存/会话可重建）。
- **BREAKING（测试 Redis）**：废弃 `docker-compose.redis-cluster.test.yml` 六节点方案，改为 **`docker-compose.redis-standalone.test.yml` 单机 Redis**（`redis-test:6379`）；测试栈 **不再** 执行 `cluster create`。
- 基线 `docker-compose.microservices.yml` 为各服务增加可选 **`GF_REDIS_DEFAULT_ADDRESS`** 注入；测试 `.env.test` 指向 `redis-test:6379`，生产不填则仍用 yaml 三主种子（与 3 节点 prod cluster 一致）。
- 新增 **`docker-compose.resources.prod.yml` / `docker-compose.resources.test.yml`**（或等价 overlay），为 prod/test 微服务及 Redis/RabbitMQ 定义 **`mem_limit` / `cpus`**（2G survival 起步配额，runbook 文档化）。
- **`docs/runbooks/release-deploy-and-run.md`**：2G 宿主机约束、MySQL `innodb_buffer_pool_size` 建议、Redis 迁移步骤、资源 limits 表、ASR 验收约定（test 压测、避免 prod 并发 ASR）、OOM 排错。
- 更新 `docs/runbooks/redis-cluster-local.md` 注释：生产/本地仍可 6 节点或对齐 3 节点（实现阶段二选一文档化）。

## Capabilities

### New Capabilities

- `compose-redis-topology-2g`：生产 3 主 0 从 Cluster 与测试单机 Redis 的 Compose 拓扑、地址注入与迁移/runbook 要求。
- `compose-container-resource-limits`：prod/test 栈容器 CPU/内存 limits 的 Compose 表达与 runbook 验收。

### Modified Capabilities

- `compose-mysql-endpoint-via-env`：测试 Redis 地址经基线 compose 可选 `GF_REDIS_DEFAULT_ADDRESS` 注入（非 test overlay 独占），单机与 cluster 地址语义文档化。
- `runtime-docs-centralization-and-governance`：runbook 须含 2G 双栈 survival 配置、资源 limits 与 Redis 拓扑变更后的启动顺序。

## Impact

- **Compose**：`docker-compose.redis-cluster.yml`、新建 `redis-standalone.test.yml`、`resources.*.yml`、基线 `microservices.yml` 环境变量、测试/生产启动命令叠加 `-f resources.*.yml`。
- **运维**：生产 Redis 重建窗口；测试 Redis 自 6 节点 cluster 迁移至单机；宿主机建议配置 swap；MySQL 内存调优。
- **应用代码**：无 API 行为变更；Redis 客户端仍通过 GoFrame `redis.default.address` / `GF_REDIS_DEFAULT_ADDRESS`。
- **不受影响**：生产 MySQL 库、RabbitMQ 拓扑、微服务镜像与端口隔离策略。
