# microservice-boundary-final-alignment Specification

## Purpose
TBD - created by archiving change service-dedicated-config-final-boundary. Update Purpose after archive.
## Requirements
### Requirement: 配置边界 MUST 与服务边界一致
系统 MUST 保证“服务职责边界、配置归属边界、运行入口边界”一致；任何服务不得通过共享主配置承担他域职责或访问路径。

#### Scenario: gateway 运行角色审查
- **WHEN** 审查 gateway 启动入口与配置项
- **THEN** gateway MUST 仅包含流量与策略层配置，不得加载他域业务执行配置

#### Scenario: voice 跨服务访问
- **WHEN** voice 需要获取 history/device 领域数据
- **THEN** voice MUST 通过服务契约访问，不得通过主配置回流到跨库直查实现

### Requirement: 最终形态迁移 MUST 包含可回滚路径
面向最终微服务形态的配置与边界收敛 MUST 提供清晰的分阶段切换与回滚策略，避免一次性切换导致生产不可用。

#### Scenario: canary 切换失败
- **WHEN** 配置切换到 canary/remote 后出现异常
- **THEN** 系统 MUST 支持按服务维度快速回滚到 local/上一版本配置

