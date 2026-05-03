## Local Redis Cluster Topology

用于 `Task 4.1` 的本地分布式 Redis 拓扑，采用 6 节点（3 主 + 3 从）集群。

### 文件位置

- compose: `manifest/docker/docker-compose.redis-cluster.yml`
- init script: `hack/redis-cluster-init.ps1`

### 启动步骤（Windows / PowerShell）

1. 启动并初始化集群：
   - `powershell -ExecutionPolicy Bypass -File "hack/redis-cluster-init.ps1"`
2. 验证集群状态：
   - `docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 redis-cli -p 7001 cluster info`
   - `docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 redis-cli -p 7001 cluster nodes`

### 连接信息（本机）

- `127.0.0.1:7001`
- `127.0.0.1:7002`
- `127.0.0.1:7003`
- `127.0.0.1:7004`
- `127.0.0.1:7005`
- `127.0.0.1:7006`

应用侧（`manifest/config/config*.yaml` 的 `redis.default.address`）已默认写 **三主种子**：`redis-node-1:7001,redis-node-2:7002,redis-node-3:7003`（与 compose 服务名一致）；本机直连调试仍可用下表各 `127.0.0.1` 端口。

### 停止与清理

- 停止：
  - `docker compose -f manifest/docker/docker-compose.redis-cluster.yml down`
- 清理数据卷（重置集群）：
  - `docker compose -f manifest/docker/docker-compose.redis-cluster.yml down -v`

### 迁移验收要点

- [ ] 集群 `cluster_state:ok`
- [ ] 3 主 3 从拓扑正常
- [ ] 任一节点重启后集群仍可读写
- [ ] 可用作后续 `voice-session` 状态外置基础设施
