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

