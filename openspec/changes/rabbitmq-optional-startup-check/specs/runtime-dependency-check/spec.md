## ADDED Requirements

### Requirement: API 类进程启动 SHALL 不因 RabbitMQ 不可达而失败

gateway-service、gateway-app-server、device-service、history-service、voice-service、ucg-service 在启动前 SHALL 校验 Redis 连通性；RabbitMQ 管理 API 探活 MAY 失败且不 SHALL 阻断进程启动。探活失败时 SHALL 记录 Warning 级别日志。

#### Scenario: RabbitMQ 宕机时 device-service 启动

- **WHEN** device-service 启动且 Redis 可达但 RabbitMQ 管理 API 不可达
- **THEN** 进程 SHALL 成功启动并监听 HTTP

#### Scenario: Redis 不可达时 API 进程启动

- **WHEN** 任一 API 类进程启动且 Redis 不可达
- **THEN** 进程启动 SHALL 失败

### Requirement: worker-service 启动 SHALL 保持 RabbitMQ 强依赖

worker-service 在启动前 SHALL 同时校验 Redis 与 RabbitMQ；任一失败 SHALL 导致进程启动失败。

#### Scenario: RabbitMQ 宕机时 worker-service 启动

- **WHEN** worker-service 启动且 RabbitMQ 管理 API 不可达
- **THEN** 进程启动 SHALL 失败

## MODIFIED Requirements

### Requirement: RabbitMQ 必须作为唯一事件通道

系统 SHALL 将 RabbitMQ 作为唯一跨服务事件通道。运行时必需事件流程中的 MQ 发布失败语义保持不变。启动期探活策略 SHALL 按进程角色区分：API 类进程 MAY 在 MQ 不可达时仍启动；worker-service SHALL 在 MQ 不可达时启动失败。

#### Scenario: RabbitMQ 不可用时 API 服务启动

- **WHEN** API 类服务进程启动且 RabbitMQ 连通性检查失败
- **THEN** 进程 MAY 继续启动并记录 MQ 降级 Warning

#### Scenario: RabbitMQ 不可用时 worker 启动

- **WHEN** worker-service 启动且 RabbitMQ 连通性或必需拓扑检查失败
- **THEN** 进程启动 SHALL 立即失败，并记录缺失依赖状态

#### Scenario: 必需事件发布失败

- **WHEN** 某请求路径要求发布事件且 RabbitMQ 发布失败或超时
- **THEN** 该请求 SHALL 被阻断，并返回事件发布失败错误响应
