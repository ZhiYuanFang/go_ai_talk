# voice-history-http-contract Specification

## Purpose
TBD - created by archiving change enforce-service-boundary-no-cross-db. Update Purpose after archive.
## Requirements
### Requirement: Voice MUST 通过 History HTTP 契约获取历史域数据

`voice-service` 在涉及 **历史记录**（由 `history` 表承载的会话/事件时间线数据）的查询或写入时 MUST 通过 `history-service` 暴露的内部 HTTP 接口完成，MUST NOT 直接访问 history 领域数据库表。用户画像（生日、性别等）、事件类型字典、动作记录等 **非 history 表** 数据 MUST NOT 通过本需求所述的 history 接口冒充权威来源，MUST 分别遵循 device 与 voice 域契约。

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

### Requirement: Voice MUST 通过 Device HTTP 契约访问用户画像与事件字典

`voice-service` 需要读取或更新 **设备用户画像**（如生日、性别）或 **事件/动作** 相关持久化数据时 MUST 通过 `device-service` 暴露的 HTTP 契约完成，MUST NOT 使用 `dao.User`、`dao.Event`、`dao.Action` 直连 device 库表。

#### Scenario: 事件抽取结果落库

- **WHEN** voice 在理解流程中创建或解析事件实体并需持久化
- **THEN** voice MUST 调用 device 服务接口（或已批准的适配层），MUST NOT 在 voice 进程内对 `event` 表执行 DAO Insert
- **AND** 新建事件时 MUST 向 device 传递合法 `eventType`（`number` | `time` | `one`）

#### Scenario: 读取事件选项列表

- **WHEN** voice 需要事件字典列表或 `eventType` 等元数据
- **THEN** voice MUST 从 device 服务获取，MUST NOT 依赖 history 服务返回的 `event` 表投影作为权威来源
- **AND** 响应项 SHALL 含 `eventType`，SHALL NOT 含 `needQuantity`

### Requirement: Voice MUST 在本域处理 suggest 表

`voice-service` 对 **`suggest` 表** 的读写 MUST 仅在 voice 进程内通过本域 DAO 或本域服务接口完成；history-service MUST NOT 作为 suggest 数据的权威存储进程。

#### Scenario: 写入每日建议

- **WHEN** voice 生成并保存建议文案
- **THEN** 持久化 MUST 发生在 voice 库 `suggest` 表路径上，MUST NOT 由 history 进程执行 `dao.Suggest` 写入

