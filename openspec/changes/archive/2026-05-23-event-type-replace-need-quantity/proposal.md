## Why

`ai_voice_device.event` 表已由 `need_quantity`（0/1）改为 `event_type`（`number` / `time` / `one`），GoFrame 生成的 `entity`/`dao` 已对齐，但业务层仍读写 `NeedQuantity`，导致 **device 包无法编译**，且 Redis 事件字典缓存快照缺少 `eventType`。需在全链路引入 `event_type`，并删除全部 `needQuantity` 兼容。

## What Changes

- **BREAKING**：事件增删改 API（管理端 multipart、device 内部 HTTP、voice 委派）表单/JSON 字段 `needQuantity` 改为 `eventType`（`number`|`time`|`one`），不保留旧字段。
- device 域：`AddEvent`/`UpdateEvent`/`InsertOrGetEventByNeedle`/DeepSeek 落库路径写入并校验 `event_type`；`eventListFields` 与写后 `RebuildEventCache` 含 `event_type`。
- voice 域：DeepSeek 提示词与结构化输出用 `event_type`；**对话新建事件**时由模型判定类型并赋值；**不以 `event_type` 驱动** start/end/one 写 history（仍以 **action.target_type** 为主）。
- 前端：`admin.html`、`history.html` 用 `eventType` 控制展示（计数/计时/一次性），移除「需要计数」复选框与 `data-need-quantity`。
- 删除 `needQuantity` 相关结构体字段、prompt 片段与 `event.NeedQuantity` 业务分支（voice 侧）。
- **不修改** `domain_outbox.event_type`（worker 出站消息类型，与母婴事件字典无关）。
- **不做** 旧数据迁移 DML（库表已由运维改好）；运行时依赖写库后 Redis 重建即可。

## Capabilities

### New Capabilities

- `device-event-type-field`：device 事件主档 `event_type` 的校验、CRUD、缓存快照与 API 契约。

### Modified Capabilities

- `voice-history-http-contract`：事件 options 元数据由 `needQuantity` 改为 `eventType`；voice 新建事件时输出 `event_type`。

## Impact

- **服务**：device-service（必重建）、voice-service、gateway（静态 `admin.html`/`history.html`）。
- **代码**：`internal/services/device/*`、`internal/controller/device_*`、`api/v1/device_*`、`internal/services/voice/*`、`internal/services/contracts/runtime_contracts.go`、`resource/public/admin.html`、`resource/public/history.html`、`README.MD`。
- **排除**：`domain_outbox`、`workeroutbox` 信封字段、history 写库主流程（不因 `event_type` 新增服务端校验，除非后续单独变更）。
