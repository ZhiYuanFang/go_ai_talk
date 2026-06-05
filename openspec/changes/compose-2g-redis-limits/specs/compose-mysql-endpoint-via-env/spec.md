## ADDED Requirements

### Requirement: 参考微服务 Compose MUST 支持通过环境变量注入 Redis 地址

`manifest/docker/docker-compose.microservices.yml`（或其后继官方等价文件）SHALL 允许通过环境变量 **`GF_REDIS_DEFAULT_ADDRESS`** 覆盖 GoFrame `redis.default.address`，作用于 **所有** 依赖 Redis 的微服务（含 gateway、gateway-app、history、voice、device、worker、ucg）。当变量为空或未设置时，SHALL 回退镜像内 yaml 默认地址（cluster 三主种子）。`.env.test.example` SHALL 文档化测试单机地址 `redis-test:6379`；`.env.prod.example` **SHALL NOT** 要求填写该变量（生产使用 yaml 默认 cluster 种子）。

#### Scenario: 测试栈注入单机 Redis 地址

- **WHEN** `.env.test` 设置 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379` 且启动测试微服务栈
- **THEN** 各服务容器环境 SHALL 包含该变量，且 Redis 客户端 SHALL 连接 `redis-test:6379`

#### Scenario: 生产栈不注入时沿用 yaml cluster 种子

- **WHEN** 生产 `.env.prod` 未设置 `GF_REDIS_DEFAULT_ADDRESS`
- **THEN** 微服务 SHALL 使用 config yaml 中的 `redis-node-1:7001,redis-node-2:7002,redis-node-3:7003`
