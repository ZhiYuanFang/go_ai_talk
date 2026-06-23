## Why

语音控制「结束」喂养/计时事件时，voice-service 在调用 `EndLatestHistoryIfMatch` 后**忽略 `updated` 返回值**，且结束决策依赖 voice 侧 Redis 缓存的 `GetLatestHistory`。当缓存与 history-service 真值不一致时，可能出现：**语音播报成功、库内无更新、无 `app:history:notify` WS 推送**。App 手动调 `event/end-latest` / `event/update` 成功时则会正常 publish。前端依赖历史 WS 做动态刷新，结束路径 MUST 与 API 等价。

本变更**仅对齐结束路径**；start/create 路径与 Redis 订阅配置不在本变更范围（可后续单独变更）。

## What Changes

- 语音 `end` 动作：以 history-service 返回的 `EndLatestHistoryIfMatch.updated` 为权威，**MUST NOT** 在 `updated=false` 时向用户播报「已结束」。
- `updated=false` 时 **MUST** 降级为与现有「非同事件结束」分支等价的写库（`AddHistory` 瞬时结束 + 必要时自动结束上一条未闭合记录），保证产生 create/update 并成功 publish WS。
- 复合结束中「自动结束上一条」**MUST** 检查 `updated` 或改用可观测的 `UpdateHistory`，避免静默 no-op。
- 抽取 voice 内统一的结束落库 helper，消除 `voice_chat_understanding.go` 与 `event_child_pending.go` 重复逻辑。
- history-service 侧 publish 语义不变；修复点在 voice 写路径决策。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `history-piece-and-realtime-notify`：语音触发的结束事件 MUST 与 App `event/end-latest` / 等价 update 一样，在写库成功后向 `app:history:notify` 推送 WS。

## Impact

- `internal/services/voice/voice_chat_understanding.go`（`handleUnifiedIntentAction` / `handleActionRecord` 的 `end` 分支）
- `internal/services/voice/event_child_pending.go`（`applyEventActionTarget` 的 `end` 分支）
- 新增 voice 内 helper（如 `event_history_end.go`）
- 无 DB 迁移；无 Redis 键变更；无新 HTTP 接口
