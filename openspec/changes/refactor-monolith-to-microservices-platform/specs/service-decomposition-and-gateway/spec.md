## ADDED Requirements

### Requirement: 领域服务 SHALL 可独立部署
系统 SHALL 将单体业务职责拆分为具有独立部署单元与运行时边界的领域服务，至少包括 gateway、device、history、voice-session 以及面向 worker 的接口能力。

#### Scenario: history 能力独立发布
- **WHEN** history service 实现发生变化且 gateway 契约未变更
- **THEN** history service 可独立部署而无需重发其他领域服务

### Requirement: Gateway SHALL 提供统一外部入口
系统 SHALL 通过 gateway 层统一暴露用户与设备流量，并集中完成路由、鉴权和请求关联元数据透传。

#### Scenario: 带身份与关联信息的路由请求
- **WHEN** 客户端向外部 API 端点发送请求
- **THEN** gateway 完成鉴权、转发到目标领域服务，并携带用于追踪的关联元数据

### Requirement: 服务契约 SHALL 显式且版本化
系统 SHALL 为同步服务交互定义显式且可版本化的契约，防止在渐进迁移过程中出现非预期破坏性变更。

#### Scenario: 向后兼容的契约演进
- **WHEN** 领域服务在响应契约中新增可选字段
- **THEN** 现有调用方仍可继续工作且契约不被破坏
