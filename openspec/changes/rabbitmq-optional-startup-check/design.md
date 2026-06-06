## Context

- 启动探活集中在 `internal/platform/runtimecheck/dependencies.go`，所有进程调用同一 `CheckDependencies`。
- MQ 运行时发布/消费逻辑分散在 `eventkit`、`async`、`workeroutbox` 等，本次 **不触碰**。
- v1.0.3 基线要求 MQ 不可用启动失败；本变更仅放宽 **API 类进程** 的启动探活。

## Goals / Non-Goals

**Goals:**

- API 类服务（gateway、gateway-app、device、history、voice、ucg）在 RabbitMQ 管理 API 不可达时仍可启动。
- Redis 探活保持 fail-fast。
- MQ 探活失败时输出可观测 Warning（含 `metric=rabbitmq_startup_degraded` 或等价日志）。
- worker-service 保持 RabbitMQ 启动强依赖。

**Non-Goals:**

- 不改为 best-effort 发布、不引入 noop publisher、不调整 TextChat 阻断语义。
- 不新增测试文件。
- 不通过环境变量切换 worker 的 MQ 强依赖（worker 固定 required）。

## Decisions

1. **`DependencyOptions.RequireRabbitMQ`**：`true` 时行为与现网一致（失败返回 error）；`false` 时跳过或 warn-only。
2. **MQ 配置缺失**（`MQ_HTTP_API_BASE` 为空）：`RequireRabbitMQ=false` 时跳过探活；`true` 时仍失败（worker 需显式配置）。
3. **调用方显式传参**，避免默认行为歧义；各 `main.go` 一眼可见策略。

## Risks / Trade-offs

- MQ 宕机期间 API 已启动但部分请求仍会因发布失败而报错（既有行为）→ 可接受，符合「不改业务逻辑」。
- worker 无法启动时 outbox relay 停止 → 已知；API 仍可写库与 Redis 通知。
