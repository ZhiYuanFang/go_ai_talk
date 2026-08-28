## 1. history-service：同 event 重复 start 守卫

- [x] 1.1 在 `internal/services/history/local.go` 抽取 `hasOpenHistoryForEvent(deviceNo, eventID, excludeID)`（查询 `end_time=0` 同 event 行），附中文注释说明与 `EndLatestHistoryIfMatch` 对称
- [x] 1.2 `AddDeviceHistory`：当 `EndTime==0 && EventId>0` 且已存在未闭合同 event 行时返回错误（文案：`该事件已在进行中，请先结束后再开始`）
- [x] 1.3 `UpdateDeviceHistory`：当将 `EndTime` 设为 `0` 时应用相同守卫（排除当前 `id`）
- [x] 1.4 确认 `EventBatch` create 与 `EventAdd` controller 路径错误能透传至 batch item `Reason` / HTTP 错误响应
- [x] 1.5 本地验证：同 device 先 start 睡眠（endTime=0），再 add/batch create 同 eventId start → 应拒绝；不同 eventId start → 应成功；one 瞬时（endTime>0）→ 应成功

## 2. voice-service：删除 dead 写库/pending 代码

- [x] 2.1 删除文件：`event_child_pending.go`、`event_history_end.go`、`event_tree.go`
- [x] 2.2 从 `voice_chat_understanding.go` 移除：`pythonIntentLandPlan`、`mapPythonIntentToLandPlan`、`handleUnifiedIntentAction`、`handleMultiEventIntent`、`matchEventByName`、`historyRowEventName`/`historyRowEventUnit`（若删文件后无引用）
- [x] 2.3 精简 `prepareChatPreamble`：移除 pending child get/clear 逻辑；更新文件/函数级中文注释为「落库由 Python batch」
- [x] 2.4 从 `voice_chat.go` 移除 `eventInfo`、`pendingChildMu`、`pendingChild` 字段及 `NewVoiceService` 初始化
- [x] 2.5 保留并确认仍编译：`parseEventIntentFromReply`、`applyUnifiedIntentResult`、`mapPythonRespToIntent`、`DeviceHistory().ListHistory` 读路径
- [x] 2.6 `go build` voice-service / history-service（或全仓 `./hack/...` 若项目惯例）确保无 dead import

## 3. 规格与自检

- [x] 3.1 `openspec validate voice-dead-code-and-history-start-guard --strict`（若 CLI 支持）
- [x] 3.2 grep 确认 `internal/services/voice` 无 `DeviceHistory().AddHistory` / `UpdateHistory` 喂养写库残留（ListHistory 等读路径除外）
- [x] 3.3 grep 确认 `handleUnifiedIntentAction` 仓库内零引用
- [x] 3.4 评审：未改 `usagestats/maintenance_skip`；未新增 Redis 直连或 background ticker

## 4. 可选 follow-up（兄弟仓，本 PR 不阻塞）

- [ ] 4.1 `python-ai-talk`：batch item `ok=false` 且 reason 含「已在进行中」时调整用户播报（非本仓库任务时可跳过并记 issue）
