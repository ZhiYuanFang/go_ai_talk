## MODIFIED Requirements

### Requirement: Gateway MUST 保持无业务后台任务职责

`gateway-service` MUST 仅承担请求入口、路由转发与横切能力，MUST NOT 在进程启动阶段承担消息消费、事件中继、domain outbox relay 等业务后台任务职责。

#### Scenario: 网关处理请求

- **WHEN** gateway 接收 HTTP/WS 请求
- **THEN** gateway MUST 仅执行入口与代理逻辑，不应存在后台任务消费副作用

#### Scenario: 部署配置审查

- **WHEN** 审查 gateway 部署配置与启动流程
- **THEN** 必须能够确认 gateway 未启动 ticker 扫表或 MQ 消费；history/device 缓存由各自服务同步 patch；UCG 后台任务仅在 ucg-service

## REMOVED Requirements

### Requirement: 角色边界变更 MUST 伴随文档更新（原「worker 为唯一执行者」场景）

**Reason**: 「worker 独占后台任务」场景失效；合并入 `background-loop-task-governance` 的 MODIFIED 文档更新 Requirement。

**Migration**: 更新 runbook 描述无 worker 架构。
