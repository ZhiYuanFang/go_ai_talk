## 1. history-service 单位补全（核心）

- [x] 1.1 在 `internal/services/history/delegate_http.go` 或 `history_row.go` 新增 `resolveEventUnitFromDevice(ctx, eventID)`：经 `delegateListEventOptions` 按 id 匹配 `unit`，禁止 `dao.Event` 直查
- [x] 1.2 将 `lookupEventUnit` 改为调用上述委托；确认 `internal/services/history` 包内无 `dao.Event` import
- [x] 1.3 修复 `internal/controller/device_history.go` 中 `mergeHistoryUpdateFromReq`：合并 `req.EventUnit`
- [x] 1.4 在 `internal/services/history/local.go` outbox 入队 map（created/updated）增加 `event_unit` 字段

## 2. 投影与实时通知

- [x] 2.1 扩展 `internal/services/history/cache_repo.go` 中 `historyProjectionEvent` 与 `ApplyProjection` 分支，读写 `EventUnit`
- [x] 2.2 确认 `internal/services/history/realtime_notify.go` 推送载荷已含 `eventUnit`（与 DB 一致，必要时补注释）

## 3. voice-service 传参

- [x] 3.1 在 `internal/services/voice/voice_chat_understanding.go` 所有 `AddHistory` 构造处，当 `event.Unit` 非空时设置 `EventUnit`
- [x] 3.2 在 `internal/services/voice/event_child_pending.go` 的 `applyEventActionTarget` 中同样传递 `EventUnit: event.Unit`

## 4. 历史管理页

- [x] 4.1 `resource/public/history.html`：event options 的 `<option>` 增加 `data-unit`；`loadEventOptions` 从 API `unit` 写入
- [x] 4.2 列表计数列展示 `{number}{unit}`；`submitHistoryModal` payload 增加 `eventUnit`

## 5. 验收

- [ ] 5.1 测试栈：device 库 event 配置 `unit=ml` → 手动新增 history → 查 `ai_voice_history_* .history.event_unit` 为 `ml`
- [ ] 5.2 测试栈：语音记录计数事件 → history 行 `event_unit` 非空
- [x] 5.3 页面列表与 WS 推送可见 `eventUnit`；grep 确认 history-service 无 `dao.Event` 跨域直查
