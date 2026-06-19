## ADDED Requirements

### Requirement: 生产与测试 Compose 栈 MUST 限制容器 json-file 日志保留量

`manifest/docker` 下用于 **生产** 与 **测试** 长期运行的 Compose 服务（微服务基线六件套、RabbitMQ prod/test、Redis prod cluster / test standalone）MUST 配置 `logging.driver=json-file`，且 MUST 设置 `max-size` 与 `max-file` 轮转选项。微服务与 Redis 默认 MUST 为 `max-size=10m`、`max-file=3`；RabbitMQ MUST 为 `max-size=20m`、`max-file=3`。策略 MUST 在 prod 与 test 对齐（同一 compose 源或等价 anchor）。

#### Scenario: 微服务容器日志有上限

- **WHEN** 运维在 ECS 上对测试或生产微服务栈执行 `docker compose up -d --force-recreate`
- **THEN** 各微服务容器 inspect 的 LogConfig MUST 含 `max-size` 10m 与 `max-file` 3

#### Scenario: RabbitMQ 容器日志有上限

- **WHEN** 运维 recreate 生产或测试 RabbitMQ compose 栈
- **THEN** Rabbit 容器 LogConfig MUST 含 `max-size` 20m 与 `max-file` 3

### Requirement: RabbitMQ MUST 降低 stdout 日志级别至 warning

生产与测试 RabbitMQ compose MUST 挂载仓库内 `manifest/docker/rabbitmq/rabbitmq.conf`，且该文件 MUST 将 `log.console.level`、`log.connection.level`、`log.channel.level` 设为 `warning`。Management 插件与 AMQP/HTTP 业务行为 MUST NOT 因该配置而关闭。

#### Scenario: Rabbit 正常收发时 stdout 更安静

- **WHEN** 客户端 HTTP Publish 与 AMQP consume 正常运行
- **THEN** Rabbit 容器 `docker logs` MUST NOT 以 info 级别刷屏连接/channel 常规行（warning 及以上仍可见）

### Requirement: runbook MUST 说明日志策略生效与磁盘清理

`docs/runbooks/release-deploy-and-run.md` MUST 说明：logging 变更需 recreate 容器；已有巨型 `*-json.log` 的清理方式（删容器或 truncate）；可用 `docker system df -v` 验收。

#### Scenario: 运维按 runbook 部署后验收

- **WHEN** 运维完成 logging 变更的 recreate
- **THEN** runbook MUST 提供检查 LogConfig 与磁盘占用的命令示例
