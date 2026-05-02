# service-boundary-no-cross-db Specification

## ADDED Requirements

### Requirement: 表归属 MUST 与部署库一致

在分库部署下，`history` 表与 `domain_outbox` MUST 仅由可连接 history 库的进程写入；`user`、`event`、`action` MUST 仅由可连接 device 库的进程写入；`qa`、`suggest` MUST 仅由可连接 voice 库的进程写入。禁止因历史单体代码路径而使用错误默认库连接组访问上述表。

#### Scenario: device 进程不写 history 库中的 outbox（除非显式配置）

- **WHEN** `domain_outbox` 表仅存在于 history 服务数据库中
- **THEN** device-service MUST NOT 使用 `user` 表所在连接组对 `domain_outbox` 执行 Insert，除非运维显式配置为同一物理库且经架构评审

#### Scenario: voice 进程不写 event/action 表

- **WHEN** voice-service 需要新增或查询事件字典、动作记录
- **THEN** voice MUST 通过 device 服务契约完成，MUST NOT 使用 `dao.Event` 或 `dao.Action` 直连 device 库表

### Requirement: history 服务 MUST NOT 冒充他域数据权威

对外 HTTP 或内部契约 MUST NOT 将「生日、事件选项、语音建议」等响应伪装为 history 数据库本地查询结果；若经网关聚合， MUST 在实现上分别调用 device/voice 权威服务，且错误语义可追溯至真实下游。

#### Scenario: 拆分后的 API 归属

- **WHEN** 客户端请求事件选项或用户画像
- **THEN** 响应数据 MUST 来源于 device 域存储与接口，而非 history 进程内对 `event`/`user` 表的 DAO 查询
