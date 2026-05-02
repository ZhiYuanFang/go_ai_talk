## Why

独立部署下各进程已绑定不同数据库（history：`history` + `domain_outbox`；device：`user` + `event` + `action`；voice：`qa` + `suggest`），但 `internal/services/history/local.go` 等仍通过 `dao.User`、`dao.Event`、`dao.Suggest` 在同一实现中混查，违背「库归属=服务归属」与仓库 `AGENTS.md` 跨域规则，存在连错库、事务无法跨库、以及 history HTTP  façade 承载非 history 数据等风险。需要在规格与实现上对齐数据所有权与调用路径。

## What Changes

- **history-service**：本地实现仅持久化/查询 **`history` 与 `domain_outbox`**；移除对 `user`、`event`、`suggest` 表的 DAO 直连；原「生日/事件选项/建议」能力改为通过 **device-service / voice-service 契约** 调用，或由网关聚合；**BREAKING**：独立部署 history 时，原先由 history 单体路径提供的跨表行为将拆分。
- **device-service**：确认 **`user`、`event`、`action`** 为唯一权威存储；修正若存在的 **`domain_outbox` 写入使用错误 DB 连接组** 的问题，使 outbox 写入与「outbox 表所在库」一致（通常与 history 域协调，而非 `dao.User.Group()` 误连）。
- **voice-service**：**`qa`、`suggest`** 仅连 voice 库；对 **`event`、`action`** 的写入/查询改为 **device 契约**，禁止 voice 进程直连 device 表。
- **契约与路由**：更新 `history` 适配器/远程客户端：将非 history 库能力从「history baseURL」下拆出或改为调用对应服务 URL；必要时扩展 `contracts` 中 HTTP 目标解析。
- **缓存与投影**：`RebuildHistoryMetaCache`、`RebuildBirthdayCacheByDevice` 等重建路径改为从正确服务拉取或订阅事件，避免误用本地 DAO。
- **文档**：运行手册中明确各服务允许访问的表清单与禁止项。

## Capabilities

### New Capabilities

- `history-service-db-ownership`：定义 history 进程允许直连的数据库对象（`history`、`domain_outbox`）及禁止项；跨域数据 MUST 走契约。

### Modified Capabilities

- `service-boundary-no-cross-db`：补充 **history-service / device-service / voice-service** 与各表归属的明确场景，以及「outbox 表仅存在于 history 库时禁止由 device 连接组写入」等约束。
- `voice-history-http-contract`：将「生日、事件选项」从「经 history 获取」修正为 **经 device（及必要时 voice 自身 suggest）**；history 契约仅覆盖 **历史记录与 history 域 outbox 相关行为**。

## Impact

- **代码**：`internal/services/history/local.go`、`adapter.go`、`cache_rebuild.go`；`internal/services/device/admin.go`（outbox 连接）；`internal/services/voice/voice_chat_understanding.go` 等使用 `dao.Event`/`dao.Action` 的路径；`internal/services/contracts` 与相关 controller/gateway 路由。
- **配置**：各服务独立 `GF_GCFG_FILE` / 数据库 link；跨服务 URL 与环境变量（`HISTORY_SERVICE_URL`、device/voice 基址等）。
- **运维**：部署拓扑下需保证 device/voice/history 网络互通；迁移期保留 `local|remote|canary` 与可观测日志。
