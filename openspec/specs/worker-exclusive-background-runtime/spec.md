# worker-exclusive-background-runtime Specification

## Purpose
TBD - created by archiving change worker-exclusive-background-tasks. Update Purpose after archive.
## Requirements
### Requirement: 后台任务 MUST 仅由 worker-service 启动
系统中的业务后台任务（至少包括 voice task consumer 与 domain outbox relay）MUST 仅由 `worker-service` 进程启动，其他服务进程 MUST NOT 启动这些任务。

#### Scenario: worker 启动后台任务
- **WHEN** `worker-service` 完成依赖检查并进入启动流程
- **THEN** 系统 MUST 启动后台任务并持续执行队列消费与 outbox relay

#### Scenario: gateway 启动流程
- **WHEN** `gateway-service` 启动 HTTP 服务
- **THEN** 系统 MUST NOT 启动业务后台任务

### Requirement: 后台任务启动语义 MUST 保持幂等
后台任务启动入口 MUST 保持幂等语义，避免重复调用导致同进程内重复启动 goroutine。

#### Scenario: 重复调用启动入口
- **WHEN** 在同一进程内重复触发后台任务启动入口
- **THEN** 系统 MUST 只保留一份后台任务实例运行

