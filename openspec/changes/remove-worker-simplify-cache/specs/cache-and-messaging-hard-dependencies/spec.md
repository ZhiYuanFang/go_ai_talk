## MODIFIED Requirements

### Requirement: RabbitMQ 必须作为唯一事件通道

系统 SHALL 将 RabbitMQ 作为跨服务事件通道。**UCG 审核/推荐等必需 consumer 流程** MUST 保持 RabbitMQ；history/device 的 `history.record.*` / `device.*` **fan-out 发布** 随 worker 删除而移除，不再作为必需流程。API 类进程 MAY 在 RabbitMQ 不可达时仍启动。**已删除 worker-service**，不再有「worker 启动 MUST 因 RabbitMQ 失败而失败」的进程角色。

#### Scenario: RabbitMQ 不可用时 API 服务启动

- **WHEN** API 类服务进程启动且 RabbitMQ 连通性检查失败
- **THEN** 进程 MAY 继续启动并记录 MQ 降级 Warning（ucg-service 若启用 consumer 则按 ucg 规格处理）

#### Scenario: 运行时必需 UCG 事件发布失败

- **WHEN** ucg-service 路径要求发布审核/推荐事件且 RabbitMQ 发布失败
- **THEN** 系统 MUST 按 UCG outbox/MQ 规格处理（与 worker 删除无关）

## REMOVED Requirements

### Requirement: RabbitMQ 不可用时 worker 启动失败

**Reason**: worker-service 已删除。

**Migration**: 从 deployment 移除 worker；RabbitMQ 依赖以 ucg-service consumer 为准。
