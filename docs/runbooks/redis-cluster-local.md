## Local Redis Cluster Topology

用于本地开发的 Redis Cluster 拓扑，与生产一致：**3 主 0 从**（7001–7003）。

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

应用侧（`manifest/config/config*.yaml` 的 `redis.default.address`）已默认写 **三主种子**：`redis-node-1:7001,redis-node-2:7002,redis-node-3:7003`（与 compose 服务名一致）。

### 停止与清理

- 停止：
  - `docker compose -f manifest/docker/docker-compose.redis-cluster.yml down`
- 清理数据卷（重置集群）：
  - `docker compose -f manifest/docker/docker-compose.redis-cluster.yml down -v`

### 迁移验收要点

- [ ] 集群 `cluster_state:ok`
- [ ] 3 主 0 从拓扑正常
- [ ] 任一节点重启后集群仍可读写
- [ ] 可用作后续 `voice-session` 状态外置基础设施
