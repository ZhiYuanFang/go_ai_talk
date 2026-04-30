## ADDED Requirements

### Requirement: Voice MUST 通过 History HTTP 契约获取历史域数据
`voice-service` 在涉及历史记录、生日信息、事件选项等 history 域数据时 MUST 通过 `history-service` 暴露的内部 HTTP 接口完成查询或写入，MUST NOT 直接访问 history 领域数据库表。

#### Scenario: 查询历史记录用于对话生成
- **WHEN** voice 处理“查询历史记录”或需要最近历史上下文的请求
- **THEN** voice MUST 调用 history 内部查询接口获取数据，并使用返回结果生成回复

#### Scenario: History 服务不可达
- **WHEN** voice 调用 history 内部接口超时或网络失败
- **THEN** voice MUST 返回可观测的错误语义，并按照配置决定是否执行本地兜底（仅迁移期允许）

### Requirement: History 内部接口 MUST 提供稳定错误语义
history 内部接口 MUST 对参数错误、资源不存在、服务异常返回可区分的错误结构，供 voice 做一致错误处理与日志分类。

#### Scenario: 参数不合法
- **WHEN** voice 传入缺失 `deviceNo` 或非法参数
- **THEN** history MUST 返回明确的参数错误码与错误信息，voice MUST 将其映射为调用方可理解的失败结果

#### Scenario: 服务端内部异常
- **WHEN** history 在处理请求时发生内部错误
- **THEN** history MUST 返回统一错误结构，voice MUST 记录失败原因并输出统一告警日志
