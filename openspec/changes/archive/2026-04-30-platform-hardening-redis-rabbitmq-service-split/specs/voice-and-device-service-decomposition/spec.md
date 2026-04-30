## ADDED Requirements

### Requirement: Voice 与 Device 领域逻辑必须运行在独立服务中
系统 SHALL 将 voice 领域与 device 领域的业务逻辑部署到独立可部署服务，并 SHALL 定义 gateway 调用所需的明确内部服务契约。

#### Scenario: Voice 请求路由
- **WHEN** gateway 收到 voice 领域请求
- **THEN** gateway SHALL 按既定内部契约将请求路由到 `voice-service`，而不是在本地执行 voice 业务逻辑

#### Scenario: Device 请求路由
- **WHEN** gateway 收到 device 领域请求
- **THEN** gateway SHALL 按既定内部契约将请求路由到 `device-service`，而不是在本地执行 device 业务逻辑

### Requirement: 服务边界遵循领域数据归属
系统 SHALL 按当前数据库/领域归属划分服务边界，并 SHALL 通过显式服务接口处理跨领域访问。

#### Scenario: Voice 流程需要 Device 领域数据
- **WHEN** `voice-service` 需要访问 device 领域数据
- **THEN** `voice-service` SHALL 通过契约化内部 API 或事件交互获取数据，而不是直接嵌入 device 领域实现
