## MODIFIED Requirements

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
