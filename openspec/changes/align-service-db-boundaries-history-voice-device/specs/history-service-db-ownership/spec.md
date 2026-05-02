# history-service-db-ownership Specification

## ADDED Requirements

### Requirement: history-service 进程 MUST 仅直连本域持久化表

`history-service`（及其独立配置所连接的默认数据库）MUST 仅对 `history` 与 `domain_outbox` 表执行 DAO/SQL 读写（不含只读副本或显式配置的跨库迁移工具）。MUST NOT 在 history 进程内对 `user`、`event`、`action`、`qa`、`suggest` 等他域业务表执行直连访问。

#### Scenario: 独立部署 history 库仅含 history 与 outbox

- **WHEN** 运行 `history-service` 且数据库中仅存在 `history` 与 `domain_outbox` 业务表
- **THEN** 服务 MUST 能完成历史记录与 outbox 相关功能，且 MUST NOT 因缺少他域表而依赖本地 DAO 回退直查

#### Scenario: 代码评审检查 history 包 import

- **WHEN** 评审 `internal/services/history` 或 history 进程入口的变更
- **THEN** MUST NOT 引入对 `dao.User`、`dao.Event`、`dao.Suggest`、`dao.Qa`、`dao.Action` 等他域 DAO 的直连依赖用于业务读写

### Requirement: 跨域数据 MUST 通过契约获取

当 history 域逻辑需要设备画像、事件字典或语音建议等非 history 表数据时，MUST 通过 **device-service / voice-service** 的 HTTP（或已批准的事件契约）获取，MUST NOT 在同一进程内直查他域表。

#### Scenario: 元数据或画像由其他服务提供

- **WHEN** 上层仍通过统一 `Contract` 需要「事件选项」或「生日」等能力
- **THEN** 实现 MUST 路由到对应服务客户端，而非 `history/local.go` 内对 `dao.Event` 或 `dao.User` 的查询
