## Why

测试/生产 ECS 上 Docker 默认 `json-file` 日志无大小上限，微服务 `logger.level: all` 与 RabbitMQ 默认 info 连接日志叠加，`*-json.log` 易占满磁盘；运维很少查看 `docker logs`，需要统一限制保留量。RabbitMQ 因多服务 HTTP Publish + ucg AMQP 长连接，日志量尤为突出。

## What Changes

- 在 **生产与测试** 共用 compose 基线及中间件 compose 中为各 service 增加 `logging`（`json-file` + `max-size` + `max-file`）。
- 新增 `manifest/docker/rabbitmq/rabbitmq.conf`，将 console/connection/channel 日志级别降为 **warning**；prod/test Rabbit compose 挂载该配置。
- runbook 补充：recreate 后生效、已有巨型 log 清理方式、`docker system df` 验收。

## Capabilities

### New Capabilities

- `docker-deploy-logging`：Compose 部署栈容器 stdout 日志轮转上限与 RabbitMQ 日志级别约束（prod/test 一致策略）。

### Modified Capabilities

（无）

## Impact

- **Compose**：`docker-compose.microservices.yml`（六微服务）、`docker-compose.rabbitmq.yml`、`docker-compose.rabbitmq.test.yml`、`docker-compose.redis-cluster.yml`、`docker-compose.redis-standalone.test.yml`。
- **配置**：新增 `manifest/docker/rabbitmq/rabbitmq.conf`。
- **文档**：`docs/runbooks/release-deploy-and-run.md`、`docs/runbooks/rabbitmq-local.md`。
- **运行时**：需 `--force-recreate` 相关容器后生效；不删 volume；业务 API 无契约变更。
