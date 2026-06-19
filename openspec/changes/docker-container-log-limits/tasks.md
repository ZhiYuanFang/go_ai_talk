## 1. RabbitMQ 日志配置

- [x] 1.1 新增 `manifest/docker/rabbitmq/rabbitmq.conf`（console/connection/channel = warning）
- [x] 1.2 `docker-compose.rabbitmq.yml`：挂载 conf、`logging` 20m×3
- [x] 1.3 `docker-compose.rabbitmq.test.yml`：同上

## 2. Compose 日志轮转（prod/test 共用源）

- [x] 2.1 `docker-compose.microservices.yml`：定义 `x-docker-logging` anchor，六微服务 `logging: *docker-logging`
- [x] 2.2 `docker-compose.redis-cluster.yml`：prod Redis 三节点加 logging 10m×3
- [x] 2.3 `docker-compose.redis-standalone.test.yml`：test Redis 加 logging 10m×3

## 3. 文档

- [x] 3.1 `docs/runbooks/release-deploy-and-run.md`：日志轮转策略、recreate、truncate、`docker system df` 验收
- [x] 3.2 `docs/runbooks/rabbitmq-local.md`：补充 rabbitmq.conf 与 logging 说明

## 4. 验收（测试栈命令写入 runbook，部署后执行）

- [x] 4.1 runbook 含 `docker inspect --format='{{.HostConfig.LogConfig}}'` 示例（微服务 + Rabbit）
- [x] 4.2 确认 prod/test overlay 未覆盖 logging（microservices.prod/test 仅镜像端口，无 logging 冲突）
