## ADDED Requirements

### Requirement: Gateway 必须收敛为流量与策略层
`gateway-service` SHALL 仅提供边缘层能力，包括鉴权、路由、策略执行、元数据透传和流量控制。

#### Scenario: 请求进入 gateway
- **WHEN** 客户端请求到达 `gateway-service`
- **THEN** gateway SHALL 执行边缘策略并转发至对应领域服务，不得执行领域业务规则

### Requirement: Gateway 在委派领域执行时必须保持外部契约稳定
`gateway-service` SHALL 在服务拆分过程中保持对外 API 契约稳定，并 SHALL 将领域业务执行委派给下游领域服务。

#### Scenario: 拆分后的既有外部 API 调用
- **WHEN** 客户端调用既有公开 API 端点
- **THEN** gateway SHALL 在调用下游领域服务处理业务的同时返回契约兼容响应
