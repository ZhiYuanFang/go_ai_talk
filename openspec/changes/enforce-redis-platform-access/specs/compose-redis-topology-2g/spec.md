## MODIFIED Requirements

### Requirement: 测试 Redis 地址 SHALL 经环境变量注入单机地址

测试部署 MUST 通过 `manifest/docker/.env.test` 设置 **`GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`**（或 runbook documented 等价单地址）。基线 `docker-compose.microservices.yml` SHALL 为需 Redis 的服务提供 `${GF_REDIS_DEFAULT_ADDRESS:-}` 注入；未设置时 SHALL 回退 yaml 默认 cluster 种子（供生产/local cluster 使用）。

#### Scenario: 测试微服务读写 test 单机 Redis

- **WHEN** 测试栈微服务启动且 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`
- **THEN** 经 `internal/platform/cachekit`（如启动探活 Ping）或 `internal/platform/redismsgkit` 建立的 Redis 连接 SHALL 指向测试单机 `redis-test:6379`，**SHALL NOT** 连接生产 cluster 节点
