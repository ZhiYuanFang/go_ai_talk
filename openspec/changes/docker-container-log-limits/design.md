## Context

当前 `manifest/docker/**` 未配置 `logging`，Docker 默认 `json-file` 无限增长。微服务 stdout 来自 GoFrame `level: all`；RabbitMQ 使用 `rabbitmq:3.13-management-alpine`，默认 info，且本栈 6 个服务连 `:15672` Publish、ucg 长连 `:5672`。生产与测试分别用独立 compose project，但 **compose 文件应对齐同一套日志策略**（用户要求 prod/test 一起改）。

## Goals / Non-Goals

**Goals:**

- 所有 **长期运行** 的 prod/test 容器（微服务六件套 + RabbitMQ + Redis）均有 json-file 轮转上限。
- RabbitMQ stdout 降到 warning，显著减少连接/channel 噪音。
- runbook 写清 recreate 与旧 log 清理，便于 ECS 运维。

**Non-Goals:**

- 不改应用 `logger.level`（留作后续可选）；不改 `daemon.json` 全局默认（compose 显式即可）。
- 不覆盖 `docker-compose.observability.yml`（可选栈，非 prod/test 主路径）。
- 不启用 `logging: none` 或关闭 management 插件。

## Decisions

### D1：Compose YAML anchor 统一微服务与 Redis 默认值（采用）

在 `docker-compose.microservices.yml` 顶部定义：

```yaml
x-docker-logging: &docker-logging
  driver: json-file
  options:
    max-size: "10m"
    max-file: "3"
```

各 service 增加 `logging: *docker-logging`。约 **30MB/容器** 上限。

**理由**：prod/test/local 共用基线，一处修改两处环境一致。

### D2：RabbitMQ 单独放宽轮转 + rabbitmq.conf（采用）

Rabbit compose（prod + test）使用：

```yaml
x-docker-logging-rabbitmq: &docker-logging-rabbitmq
  driver: json-file
  options:
    max-size: "20m"
    max-file: "3"
```

挂载 `manifest/docker/rabbitmq/rabbitmq.conf`：

```conf
log.console = true
log.console.level = warning
log.connection.level = warning
log.channel.level = warning
```

**理由**：Rabbit 日志量最大；warning 仍保留 alarm/认证失败；排障时可临时改 info。

### D3：Redis prod cluster / test standalone 复用 `&docker-logging`（采用）

在 `docker-compose.redis-cluster.yml` 的 `x-redis-node` 增加 `logging: *docker-logging`（需在文件内定义 anchor 或复制相同块）。test standalone 同理。

### D4：生效方式 recreate，不 prune volume（采用）

变更合并后部署：`docker compose ... up -d --force-recreate` 对应 stack。已有巨型 `*-json.log` 在 **删除旧容器** 或 `truncate` 后释放；**不加 `-v`**。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| warning 下排障信息变少 | runbook 说明临时改 conf 为 info 并 recreate |
| recreate 短暂断连 | 按 runbook 顺序：中间件 → 微服务 |
| 旧 log 不自动缩小 | 文档写 truncate / 删容器 |

## Migration Plan

1. 合并含 logging + rabbitmq.conf 的 compose。
2. **测试栈**：先 recreate RabbitMQ → Redis → 微服务；`docker system df -v` 观察。
3. **生产栈**：同样步骤；低峰执行。
4. 可选一次性清理：对已知巨型 log 路径 `truncate -s 0`（runbook 示例）。

## Open Questions

（无）
