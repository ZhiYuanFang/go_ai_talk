## Why

commit `827886d unit` 已为 `history.event_unit` 增加字段与 `enrichHistoryEventUnit` 写入逻辑，意图是在事件主档配置了单位（如 `ml`）时，新增/更新历史记录时反规范化写入 `history.event_unit`。但在微服务分库部署下（`HISTORY_DB_LINK` 与 `DEVICE_DB_LINK` 分离），`lookupEventUnit` 仍在 history-service 内直查 `dao.Event`，查的是 history 库而非 device 库，导致单位恒为空。该需求在 prod/test 栈上实际不可用，需补齐契约路径并闭环展示与调用方传参。

## What Changes

- 修正 history-service 补全 `event_unit` 的实现：`MUST NOT` 直查本进程 `default` 库中的 `event` 表；`MUST` 经 device-service HTTP 契约（或等价已批准的委托路径）按 `eventId` 解析 `unit`。
- voice-service 写入 history 时，当已解析出 `entity.Event.Unit`，`SHOULD` 显式传入 `History.EventUnit`，减少对远程查主档的依赖。
- history 管理页（`history.html`）计数类事件展示数量时附带单位；手动新增/编辑时可传 `eventUnit`（或依赖后端补全，二者至少一种路径在微服务下可用）。
- 修复 `mergeHistoryUpdateFromReq` 未合并 `req.EventUnit` 的遗漏。
- 异步投影（`history.record.*` outbox / Redis cache projection）消息体补齐 `event_unit` 字段，与 DB 一致。
- **不**修改 `event` 主档 schema；**不**新增 `*_test.go`；**不**要求将历史 `.env` 恢复进 git。

## Capabilities

### New Capabilities

- `history-event-unit-denorm`：history 写入路径在微服务分库下 MUST 正确反规范化事件单位；列表/piece/实时通知 MUST 暴露 `eventUnit`；调用方与投影链路 MUST 与 DB 语义一致。

### Modified Capabilities

- `history-piece-and-realtime-notify`：实时通知与 piece 列表载荷 MUST 包含 `eventUnit`（与 DB 反规范化值一致）；投影事件体 MUST 携带 `event_unit`。
- `voice-history-http-contract`：voice 经 HTTP 调用 history 新增记录时，当事件主档已知 `unit`，请求体 SHOULD 携带 `eventUnit`；history-service 仍 MUST 在缺失时通过 device 契约补全。

## Impact

- **进程**：`history-service`（`internal/services/history/history_row.go`、`local.go`、`cache_repo.go`、`delegate_http.go`）、`voice-service`（`voice_chat_understanding.go`、`event_child_pending.go`）、`gateway-app` 静态页 `resource/public/history.html`。
- **契约**：device internal HTTP（复用 event options 或新增按 id 查 unit 的内部接口）；history HTTP 请求/响应已有 `eventUnit` 字段，行为需与实现一致。
- **数据**：`ai_voice_history_* .history.event_unit` 新写入应非空（当对应 `event.unit` 非空）；既有 NULL 行不在本变更强制回填（可选运维脚本，非 scope 必须项）。
- **依赖**：`DEVICE_SERVICE_URL` 在 history-service 环境 MUST 可达（与现有 `ListEventOptions` 委托相同前提）。
