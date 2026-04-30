## ADDED Requirements

### Requirement: Redis 必须作为唯一缓存后端
系统 SHALL 将 Redis 作为 voice、device、history 相关缓存状态的唯一后端，并且 SHALL NOT 在请求处理路径执行内存缓存兜底逻辑。

#### Scenario: Redis 不可用时服务启动
- **WHEN** 服务进程启动且 Redis 连通性检查失败
- **THEN** 进程启动 SHALL 立即失败，并在启动日志中输出依赖失败原因

#### Scenario: 运行时缓存操作失败
- **WHEN** 请求处理中发生 Redis 缓存读写失败
- **THEN** 系统 SHALL 返回明确的依赖错误，且不得切换到内存兜底

### Requirement: RabbitMQ 必须作为唯一事件通道
系统 SHALL 将 RabbitMQ 作为唯一跨服务事件通道，并且 SHALL NOT 保留必需事件流程中的 MQ 关闭分支。

#### Scenario: RabbitMQ 不可用时服务启动
- **WHEN** 服务进程启动且 RabbitMQ 连通性或必需拓扑检查失败
- **THEN** 进程启动 SHALL 立即失败，并记录缺失依赖状态

#### Scenario: 必需事件发布失败
- **WHEN** 某请求路径要求发布事件且 RabbitMQ 发布失败或超时
- **THEN** 该请求 SHALL 被阻断，并返回事件发布失败错误响应
