## ADDED Requirements

### Requirement: Gateway MUST 保持无业务后台任务职责
`gateway-service` MUST 仅承担请求入口、路由转发与横切能力，MUST NOT 在进程启动阶段承担消息消费、事件中继等业务后台任务职责。

#### Scenario: 网关处理请求
- **WHEN** gateway 接收 HTTP/WS 请求
- **THEN** gateway MUST 仅执行入口与代理逻辑，不应存在后台任务消费副作用

#### Scenario: 部署配置审查
- **WHEN** 审查 gateway 部署配置与启动流程
- **THEN** 必须能够确认业务后台任务执行角色仅为 worker-service

### Requirement: 角色边界变更 MUST 伴随文档更新
当后台任务执行角色发生调整时，运行文档与部署说明 MUST 同步更新，以确保运维与开发对角色边界认知一致。

#### Scenario: 后台任务角色收敛到 worker
- **WHEN** 完成“worker 独占后台任务”改造
- **THEN** 文档 MUST 明确 gateway 不启动后台任务、worker 为唯一执行者
