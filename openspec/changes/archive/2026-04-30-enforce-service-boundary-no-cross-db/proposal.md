## Why

当前服务拆分已进入网关透传与独立部署阶段，但 `voice-service` 仍存在直接访问 `history/user/device` 领域数据库表的路径，服务边界与微服务目标不一致。需要尽快收敛为“跨服务走接口、不跨库直查”，避免后续演进中的耦合放大与发布风险。

## What Changes

- 为 `voice-service` 与 `history-service` 建立明确的跨服务查询契约，统一通过 HTTP 接口获取历史记录与历史域写入能力。
- 为 `voice-service` 与 `device-service` 建立设备信息/用户画像查询契约，统一通过服务接口获取设备档案相关数据。
- 移除 `voice-service` 对 `dao.History`、`dao.User` 等他域 DAO 的直接依赖，改为通过契约适配层调用。
- 完成 history/device 远程适配器（当前存在 `not implemented yet` 或本地直查路径）并接入 `local|remote|canary` 切换。
- 为跨服务调用增加失败语义与降级策略（可配置 failover 到 local，仅作为迁移期兜底）。
- 更新运行文档与迁移清单，明确“数据库归属=服务归属”的边界规则。

## Capabilities

### New Capabilities
- `voice-history-http-contract`: 定义 voice 调用 history 服务查询历史与画像所需的 HTTP 契约与错误语义。
- `voice-device-profile-http-contract`: 定义 voice 调用 device 服务获取设备信息/用户画像所需的 HTTP 契约与错误语义。
- `service-boundary-no-cross-db`: 定义服务间访问规则，禁止跨服务直接访问对方数据库表。

### Modified Capabilities
- None.

## Impact

- 影响代码：`internal/service/voice_chat_deepseek.go`、`internal/service/voice_chat_understanding.go`、`internal/service/device_history_adapter.go`、`internal/services/contracts/http_targets.go`、history/device controller 路由实现。
- 影响接口：新增/明确 history 服务与 device 服务内部查询 API（供 voice 调用）。
- 影响配置：`HISTORY_SERVICE_MODE`、`HISTORY_SERVICE_URL`、`HISTORY_SERVICE_CANARY_PERCENT`、`HISTORY_SERVICE_REMOTE_FAILOVER_LOCAL` 以及 device 对应远程路由配置在各环境的默认值与发布策略。
- 影响运行：voice 与 history/device 在数据访问上解耦，数据库变更风险由“跨进程耦合”降为“契约兼容管理”。
