## Why

当前 `gateway-service` 与 `worker-service` 都会触发 `StartBackgroundWorkers`，导致网关进程也接触异步业务任务，偏离“网关只做流量与策略层”的架构目标。需要将后台任务职责收敛到 `worker-service`，避免角色混淆与运行时竞争。

## What Changes

- 调整启动职责：仅 `worker-service` 启动后台任务，`gateway-service` 不再启动任何业务 worker。
- 明确后台任务的服务角色约束，防止后续在网关入口回归启动业务任务。
- 补齐配置与文档，明确不同服务的后台任务开关和默认值语义。
- 增加运行验证步骤，确保 worker 独占消费/中继且 gateway 无后台任务副作用。

## Capabilities

### New Capabilities
- `worker-exclusive-background-runtime`: 定义后台任务仅由 worker 进程启动的运行时约束。
- `gateway-no-business-workers`: 定义 gateway 进程禁止启动业务后台任务的边界要求。

### Modified Capabilities
- None.

## Impact

- 影响代码：`internal/cmd/cmd.go`、`cmd/worker-service/main.go`、`internal/service/background_workers.go` 及相关启动路径。
- 影响配置：`MQ_CONSUMER_ENABLED`、`OUTBOX_RELAY_ENABLED` 在 gateway/worker 的默认值与说明。
- 影响运行：消息消费与 outbox relay 统一由 worker 承担，gateway 回归无状态流量入口角色。
