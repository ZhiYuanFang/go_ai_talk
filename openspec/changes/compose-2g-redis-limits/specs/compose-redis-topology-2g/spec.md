## ADDED Requirements

### Requirement: 生产 Redis Cluster SHALL 为 3 主 0 从

`manifest/docker/docker-compose.redis-cluster.yml` SHALL 仅定义 **3** 个 Redis 服务（`redis-node-1`..`redis-node-3`），端口 **7001–7003**。仓库 runbook SHALL 文档化初始化命令：`redis-cli --cluster create` 仅包含上述三节点，且 **`--cluster-replicas 0`**。应用 config 中三主种子地址 `redis-node-1:7001,redis-node-2:7002,redis-node-3:7003` SHALL 与拓扑一致，**无需**为缩容修改 Go 代码。

#### Scenario: 生产 cluster 初始化成功

- **WHEN** 运维在空 volume 上启动 3 节点 compose 并执行 documented `cluster create`
- **THEN** `CLUSTER INFO` SHALL 报告 `cluster_state:ok`，且 `CLUSTER NODES` SHALL 显示 3 个 master、0 个 replica

#### Scenario: 生产微服务连接 Redis

- **WHEN** 生产微服务在 `go-ai-talk-net` 上启动且未设置 `GF_REDIS_DEFAULT_ADDRESS`
- **THEN** 进程 SHALL 通过 yaml 默认三主种子连接生产 3 节点 cluster

### Requirement: 测试 Redis SHALL 为单机 standalone

仓库 SHALL 提供 `manifest/docker/docker-compose.redis-standalone.test.yml`（或后继等价文件），定义 **单** Redis 服务（约定服务名 **`redis-test`**，容器端口 **6379**），且 **仅** 加入 `go-ai-talk-test-net`。**SHALL NOT** 要求测试栈执行 `redis-cli --cluster create`。测试栈 **SHALL NOT** 依赖 `docker-compose.redis-cluster.test.yml` 六节点拓扑作为默认路径。

#### Scenario: 测试 Redis 启动无需 cluster create

- **WHEN** 运维 `up -d` 测试 standalone Redis compose
- **THEN** 容器 running 后 SHALL 可直接 `redis-cli PING` 返回 `PONG`，且 **无需** cluster 初始化步骤

#### Scenario: 测试与生产 Redis 网络隔离

- **WHEN** 生产与测试栈同时运行
- **THEN** 测试 Redis 容器 SHALL 不在 `go-ai-talk-net` 上，生产 Redis 容器 SHALL 不在 `go-ai-talk-test-net` 上

### Requirement: 测试 Redis 地址 SHALL 经环境变量注入单机地址

测试部署 MUST 通过 `manifest/docker/.env.test` 设置 **`GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`**（或 runbook documented 等价单地址）。基线 `docker-compose.microservices.yml` SHALL 为需 Redis 的服务提供 `${GF_REDIS_DEFAULT_ADDRESS:-}` 注入；未设置时 SHALL 回退 yaml 默认 cluster 种子（供生产/local cluster 使用）。

#### Scenario: 测试微服务读写 test 单机 Redis

- **WHEN** 测试栈微服务启动且 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`
- **THEN** `g.Redis()` SHALL 连接测试单机 Redis，**SHALL NOT** 连接生产 cluster 节点

## REMOVED Requirements

### Requirement: 测试栈 SHALL 使用六节点 Redis Cluster 作为默认拓扑

**Reason**: 2G ECS 双栈资源不足；测试环境无 HA 需求，单机 Redis 足够。

**Migration**: 删除或归档 `docker-compose.redis-cluster.test.yml` 的默认 runbook 路径；已有测试 cluster volume 使用 `down -v` 后改用 standalone compose；更新 `.env.test` 与 runbook B.1/B.2。
