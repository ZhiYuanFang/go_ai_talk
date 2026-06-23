## 1. 统一结束落库 helper

- [x] 1.1 新增 `applyVoiceEventEndHistory`：先 `EndLatestHistoryIfMatch`，`updated=true` 直接成功；`updated=false` 走 AddHistory 降级 + 可选自动结束上一条（检查二次 `updated`，必要时 `UpdateHistory`）
- [x] 1.2 补充中文注释：说明以 history-service `updated` 为权威、避免缓存误判导致无 WS

## 2. 替换 voice 三处 end 分支

- [x] 2.1 `voice_chat_understanding.go`：`handleUnifiedIntentAction` 与 `handleActionRecord` 的 `end` case 改为调用 helper
- [x] 2.2 `event_child_pending.go`：`applyEventActionTarget` 的 `end` case 改为调用 helper
- [x] 2.3 删除三处重复的 `GetLatestHistory` 预判与忽略 `updated` 的逻辑

## 3. 验证

- [ ] 3.1 联调：语音结束当前计时事件 → App 历史 WS 收到 `action=update`
- [ ] 3.2 联调：`EndLatest` 不匹配场景（如跨事件结束）→ 仍有 `create`/`update` WS，且语音文案与改前语义一致
