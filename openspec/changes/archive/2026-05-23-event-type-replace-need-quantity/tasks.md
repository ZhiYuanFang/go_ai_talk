## 1. Device 域与缓存

- [x] 1.1 新增 `ValidateEventType` / `NormalizeEventType` 与常量 `number|time|one`（`internal/services/device`）
- [x] 1.2 `eventListFields`、`AddEvent`、`UpdateEvent`、`InsertOrGetEventByNeedle`、`ApplyDeepSeekEventExtractPersistence` 改用 `EventType`；删除全部 `NeedQuantity`
- [x] 1.3 确认写库成功后仍调用 `refreshEventOptionsCacheAfterMutate`

## 2. API 与契约

- [x] 2.1 更新 `runtime_contracts.go`、`admin_http_client.go`、`api/v1/device_admin_http.go`、`api/v1/device_internal_full_http.go`
- [x] 2.2 `device_admin_event.go` multipart 解析 `eventType`；`device_internal_handlers.go` 同步字段
- [x] 2.3 清理 `device_history.go` 中 `NeedQuantity` 引用（改 `eventType` 或删除未用辅助函数）

## 3. Voice 域

- [x] 3.1 `deepSeekUnifiedIntent` 与 DeepSeek prompt：`event_type` 替代 `need_quantity`；新建事件判定规则写入 system 文案
- [x] 3.2 `InsertOrGetEventByNeedle`、实体抽取落库传递 `eventType`；匹配已有事件不改 type
- [x] 3.3 删除 `event.NeedQuantity` / `intent.NeedQuantity` 业务分支；`voice_chat.go` 中 `eventInfo` 等字段更名
- [x] 3.4 `go build` device + voice + controller 通过

## 4. 前端与文档

- [x] 4.1 `admin.html`：事件类型下拉/列表列/FormData 使用 `eventType`
- [x] 4.2 `history.html`：`data-event-type` 与三分支 UI（number/time/one）
- [x] 4.3 更新 `README.MD` 事件表单字段说明

## 5. 验收

- [x] 5.1 管理端新增/编辑事件后列表与 Redis 快照含正确 `eventType`
- [x] 5.2 语音对话新建事件后 DB `event_type` 与 DeepSeek 输出一致
- [x] 5.3 确认未改动 `domain_outbox` / `workeroutbox` 的 `event_type` 语义
