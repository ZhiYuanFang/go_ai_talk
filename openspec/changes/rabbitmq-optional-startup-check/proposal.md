## Why

生产环境 RabbitMQ 短暂不可用或维护时，当前所有微服务在 `runtimecheck.CheckDependencies` 阶段因 MQ 探活失败而无法启动，导致登录、历史记录等不依赖 MQ 的核心 API 整体不可用。容灾目标：**MQ 挂时除 worker 外的服务仍可正常启动**；运行时 MQ 相关业务逻辑保持不变。

## What Changes

- `runtimecheck.CheckDependencies` 增加 `RequireRabbitMQ` 选项：为 `false` 时 MQ 探活失败仅打 Warning 日志，不阻断启动。
- **gateway / gateway-app / device / history / voice / ucg** 启动时 `RequireRabbitMQ: false`（仅校验 Redis）。
- **worker-service** 保持 `RequireRabbitMQ: true`（消费者进程仍 fail-fast）。
- **不修改** 任何 `Publish`、outbox、consumer 等业务路径。

## Capabilities

### New Capabilities

- `runtime-dependency-check`: 定义启动期 Redis 必检与 RabbitMQ 可选探活的进程级策略。

### Modified Capabilities

- `platform-hardening-redis-rabbitmq-service-split`（v1.0.3 基线）：修订「RabbitMQ 不可用时服务启动 SHALL 立即失败」为「API 类进程 MAY 在 MQ 不可达时仍启动；worker 进程 SHALL 保持 fail-fast」。

## Impact

- `internal/platform/runtimecheck/dependencies.go`
- `cmd/*/main.go`（6 个服务入口）、`internal/cmd/cmd.go`（gateway）
- OpenSpec spec delta；runbook 补充一句容灾语义（可选）
