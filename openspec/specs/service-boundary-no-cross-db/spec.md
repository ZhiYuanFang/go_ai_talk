# service-boundary-no-cross-db Specification

## Purpose
TBD - created by archiving change enforce-service-boundary-no-cross-db. Update Purpose after archive.
## Requirements
### Requirement: 服务边界 MUST 与数据库边界一致
每个微服务 MUST 仅访问其所属数据库/表；跨服务数据访问 MUST 通过服务契约（HTTP/RPC/事件）进行，MUST NOT 通过 DAO 或 SQL 直连他域数据表。

#### Scenario: Voice 访问 history 数据
- **WHEN** voice 需要读取或更新 history 领域数据
- **THEN** voice MUST 通过 history 服务契约调用完成，代码中 MUST NOT 出现对 history DAO 的直接查询或更新

#### Scenario: Voice 访问 device/user 画像数据
- **WHEN** voice 需要读取设备资料、生日、性别或注册状态
- **THEN** voice MUST 通过 device 服务契约调用完成，代码中 MUST NOT 出现对 `user/device` 领域 DAO 的直接查询

#### Scenario: 代码评审发现跨库直查
- **WHEN** 新增代码出现跨服务数据库直连访问
- **THEN** 该变更 MUST 被视为违反架构约束并在合入前整改

### Requirement: 迁移期分流 MUST 支持可控回退
服务边界治理迁移期 MUST 支持 `local|remote|canary` 切换，并对同一分流键保持稳定命中，确保可渐进放量与快速回滚。

#### Scenario: Canary 分流验证
- **WHEN** 开启 canary 模式并设置百分比
- **THEN** 同一设备标识 MUST 稳定命中同一路径，避免在 local 与 remote 之间抖动

#### Scenario: 远程路径故障回退
- **WHEN** remote 路径连续失败且开启 failover 配置
- **THEN** 调用方 MUST 回退到 local 路径并记录可观测日志

### Requirement: 单服务单库部署下禁止进程内跨域 DAO

当进程配置的 `database.default` 仅包含本服务所属库时，业务代码 MUST NOT 通过 import 他域服务包并调用其基于 DAO 的实现来访问他域表；跨服务数据 MUST 经 HTTP/RPC/消息。禁止依赖「同一代码仓库、不同包」造成可连他域表的假象。

#### Scenario: Voice 进程仅配置 voice 库

- **WHEN** `voice-service` 的配置仅连接 `qa`/`suggest` 所在库
- **THEN** 代码路径 MUST NOT 执行对 `user`/`event`/`action` 等表的 DAO；必须通过 device-service 契约

#### Scenario: 评审发现 voice 包引用他域 DAO

- **WHEN** 代码评审发现 `internal/services/voice` 直接或间接触发他域 `dao` 访问
- **THEN** 该变更 MUST 拒绝合入，直至改为 HTTP 客户端或经批准的同进程例外（文档化且仅限非生产）

### Requirement: Device 进程内 outbox 写入 MUST 使用显式 history 库连接

`domain_outbox` 若仅存在于 history 库，device-service MUST 使用独立配置的 `history_relay`（或等价）连接组写入，MUST NOT 误用 `default` 连接组指向 device 库写 outbox。

#### Scenario: 分库部署

- **WHEN** device 与 history 为不同数据库实例
- **THEN** 未正确配置 relay 时 MUST 跳过或失败可观测，MUST NOT 静默写入错误库

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

