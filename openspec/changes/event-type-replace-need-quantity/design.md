## Context

- 表 `ai_voice_device.event`：`need_quantity` 已删除，列 `event_type`（`number` | `time` | `one`）已存在；`entity.Event`、`dao.Event.Columns()` 已生成。
- 业务仍使用 `NeedQuantity`，`go build ./internal/services/device/...` 失败。
- Redis 事件 options 经 `eventListFields()` + `RebuildEventCache`；写库后已有 `refreshEventOptionsCacheAfterMutate`。
- 语音写 history 由 **action.target_type**（start/end/one 等）驱动；`event_type` 主要供管理端与 history 页 UI。
- **`domain_outbox.event_type`** 为 worker 出站路由/信封语义，**禁止**与本变更混用。

## Goals / Non-Goals

**Goals:**

- 全链路使用 `eventType` / `event_type`，删除 `needQuantity`。
- 创建/更新事件时校验并持久化 `event_type`；列表与 Redis 快照含该字段。
- 对话中 **新建事件** 时 DeepSeek 输出 `event_type` 并落库；匹配已有事件时 **不修改** 其 `event_type`。
- voice 写 history 逻辑仍以 action 为主，移除 `event.NeedQuantity` 分支。

**Non-Goals:**

- 不修改 `domain_outbox` 表或 `DomainOutbox.EventType` 相关代码。
- 不做 `need_quantity` 数据迁移、API 双读兼容。
- 不用 `event_type` 在服务端校验 history 行字段组合（前端展示约束即可）。
- 不新增 `*_test.go`（仓库约定）。

## Decisions

### 1. 常量与校验（device 包）

```go
// EventTypeNumber / EventTypeTime / EventTypeOne
func ValidateEventType(t string) error  // 非空且为三者之一
func NormalizeEventType(t string) string // 非法时默认 EventTypeTime（管理端默认「计时」）
```

管理端新建默认建议 `time`（对齐旧 `need_quantity=0` 的主展示路径）。

### 2. API 与契约

| 入口 | 字段 |
|------|------|
| multipart add/update | 表单 `eventType` |
| device internal HTTP | JSON `eventType` |
| `entity.Event` JSON | `eventType` |

**BREAKING**：移除 `needQuantity`。

### 3. 缓存

- `eventListFields()` 增加 `c.EventType`，去掉 `NeedQuantity`。
- 任意 event 写成功后 `refreshEventOptionsCacheAfterMutate`（已有），无需单独 Redis 逻辑。

### 4. Voice / DeepSeek

- `deepSeekUnifiedIntent`：`event_type` 字符串，删除 `need_quantity`。
- 统一意图与实体抽取 prompt：要求新建事件时输出 `"event_type":"number|time|one"`。
- 判定启发（写入 prompt）：
  - 计数/数量语义 → `number`
  - 开始结束/持续计时 → `time`
  - 一次性完成记录 → `one`
- `InsertOrGetEventByNeedle(ctx, needle, eventType string)`：仅插入时设置 `EventType`。
- `ApplyDeepSeekEventExtractPersistence`：新插入用 `out.EventType`；仅合并 `extra_names` 时不改 type。
- 删除 `if event.NeedQuantity > 0` 等分支；数量追问继续依赖 `intent.Quantity` + action。

### 5. 前端

- **admin.html**：下拉三选一；列表列「事件类型」；`FormData`/`buildEventRowData` 用 `eventType`。
- **history.html**：`data-event-type`；按 `number`/`time`/`one` 控制数量框与起止时间展示。

### 6. 与 outbox 投影

`enqueueDeviceProjectionEvent` 仍发 `device.event.changed`；投影仍全表 Scan event。**不**向 outbox payload 或 `domain_outbox.event_type` 写入 `number/time/one`。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 全文搜索 `event_type` 误改 outbox | 评审排除 `domain_outbox`/`workeroutbox`；design 明确 Non-Goals |
| DeepSeek 输出非法 type | `NormalizeEventType` + 日志；管理端人工可改 |
| 旧前端缓存 admin/history | 部署后强刷；API 已 BREAKING |
| history 进程独立 event 缓存 TTL 内旧 type | 可接受；miss 后委派 device |

## Migration Plan

1. 合并代码后构建 device-service、voice-service、gateway。
2. 部署 device → voice → gateway（静态页）。
3. 可选：重启后任意 event 写操作或运维触发一次事件列表刷新以覆盖 Redis。

## Open Questions

- 无（匹配已有事件不改 `event_type` 已确认）。
