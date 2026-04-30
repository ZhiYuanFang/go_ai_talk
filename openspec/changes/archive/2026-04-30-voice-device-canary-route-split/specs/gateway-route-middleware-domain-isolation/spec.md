## ADDED Requirements

### Requirement: Gateway 路由中间件 MUST 按领域拆分管理
gateway MUST 将 voice 与 device 路由代理逻辑拆分为独立中间件与配置读取路径，不得在同一中间件实现中混合管理两个领域。

#### Scenario: 修改 voice 路由逻辑
- **WHEN** 开发者调整 voice 路由代理策略
- **THEN** 变更 MUST 限定在 voice 独立中间件实现内，且不应直接影响 device 路由行为

#### Scenario: 修改 device 路由逻辑
- **WHEN** 开发者调整 device 路由代理策略
- **THEN** 变更 MUST 限定在 device 独立中间件实现内，且不应直接影响 voice 路由行为

### Requirement: 领域路由配置 MUST 互相隔离
voice 与 device 的路由模式、目标地址、canary 百分比配置 MUST 分别独立，禁止共享配置键。

#### Scenario: 仅调整 voice canary 百分比
- **WHEN** 运维仅修改 voice 的 canary 百分比配置
- **THEN** device 路由行为 MUST 保持不变
