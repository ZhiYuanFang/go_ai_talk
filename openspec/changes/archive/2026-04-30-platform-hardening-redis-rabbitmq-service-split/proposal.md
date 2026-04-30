## Why

当前代码仍存在缓存与消息中间件的多分支运行路径，导致可读性下降且行为不够确定。需要将 Redis 与 RabbitMQ 收敛为强依赖，并完成服务边界拆分，使 gateway 成为纯粹的流量与策略层。

## What Changes

- 移除业务流程中的内存降级与后端切换分支，强制 voice、device、history 仅使用 Redis 缓存。
- 移除 MQ 可选开关与降级逻辑，强制跨服务事件仅通过 RabbitMQ 传递。
- **BREAKING**：调整运行语义为 Redis/RabbitMQ 不可用时启动即失败，且必需事件发布失败时阻断请求。
- 按当前数据库归属与领域边界，将业务实现拆分为独立部署的 `voice-service` 与 `device-service`。
- 重构 gateway-service，仅保留鉴权、路由、策略、流量控制、元数据透传等横切能力，移除嵌入式领域业务实现。
- 删除现有 Go 测试文件，并在本阶段后续迭代中不再新增测试文件。

## Capabilities

### New Capabilities
- `cache-and-messaging-hard-dependencies`：定义 Redis-only 缓存与 RabbitMQ-only 消息的强依赖语义，以及严格的启动/运行失败行为。
- `voice-and-device-service-decomposition`：将 voice 与 device 领域业务拆分为独立服务，并定义明确的内部调用契约。
- `gateway-policy-layer-consolidation`：将 gateway-service 收敛为不承载领域业务的流量与策略层。
- `service-runtime-standardization`：统一缓存键、事件路由键、依赖健康检查与失败处理等运行时规范。

### Modified Capabilities
- 无。

## Impact

- 影响代码范围：`internal/service/*`、`internal/controller/*`、`main.go` 与 `cmd/*` 下服务入口，以及路由注册与配置解析逻辑。
- 影响运行依赖：Redis 与 RabbitMQ 成为启动与请求路径上的必需依赖。
- 影响部署拓扑：新增/完善 voice 与 device 领域服务进程，gateway 作为统一入口转发到领域服务。
- 影响运维侧：需要加强 Redis/RabbitMQ 可用性监控与事件发布链路告警。
- 影响质量流程：本阶段移除 `*_test.go` 文件，验证以启动校验与运行时核验脚本为主。
