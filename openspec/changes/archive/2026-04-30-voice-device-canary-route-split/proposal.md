## Why

当前 gateway 对 `voice` 与 `device` 的代理能力仅支持 `local/proxy`，且共用同一套中间件组织方式，与 `history` 的独立路由与 canary 能力不一致。需要将 voice/device 路由策略升级为与 history 同构，降低发布风险并提升服务边界清晰度。

## What Changes

- 为 `voice` 与 `device` 分别引入独立路由代理组件，移除“一个中间件管理两个领域”的组织方式。
- 为 `voice` 与 `device` 增加 `canary` 模式，支持与 history 一致的分阶段流量切换（`local|proxy|canary`）。
- 增加 voice/device canary 分流配置（如百分比变量）与稳定分流键策略。
- 更新容器编排与部署清单，补齐 voice/device canary 相关环境变量。
- 更新网关路由契约与端点归属文档，明确 voice/device 已具备独立 service 管理与灰度发布能力。

## Capabilities

### New Capabilities
- `voice-route-canary-management`: 定义 voice 路由独立管理与 canary 分流能力。
- `device-route-canary-management`: 定义 device 路由独立管理与 canary 分流能力。
- `gateway-route-middleware-domain-isolation`: 定义 gateway 路由中间件按领域拆分的结构约束。

### Modified Capabilities
- None.

## Impact

- 影响代码：`internal/controller/domain_route_proxy.go`（拆分与重构）、新增 voice/device 独立代理中间件文件。
- 影响配置：`manifest/docker/docker-compose.microservices.yml` 与 kustomize gateway 环境变量。
- 影响运行行为：voice/device 将与 history 一样支持 canary 灰度，流量切换更细粒度。
