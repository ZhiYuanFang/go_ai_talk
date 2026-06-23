## Context

- 历史 WS 链路：`history-service` 写库成功 → `publishHistoryChange` → Redis `app:history:notify` → `gateway-app` 订阅 → `/device/app/ws/history` 广播。publish **仅**在 history-service 本地写路径触发；voice remote 模式不在 voice 进程 publish。
- `EndLatestDeviceHistoryIfMatch`：仅当 DB 最近一条 `eventId` 匹配时更新 `endTime` 并 `publish action=update`；否则 `updated=false`、不写库、不 publish。
- voice `end` 分支现状：先用 voice 侧 `GetLatestHistory`（可命中 Redis 缓存）比较 `lastEvent.EventId == event.Id`，匹配则调 `EndLatestHistoryIfMatch` 但**丢弃 `updated`**；不匹配则 `AddHistory` + 可选二次 `EndLatest`（亦丢弃 `updated`）。
- 失败模式：缓存认为「同事件」→ 调 EndLatest → 服务端不匹配 → `updated=false` → 无 WS，仍播报成功。

## Goals / Non-Goals

**Goals:**

- 语音 `end` 成功播报时，history 表 MUST 有对应 create/update，且 history-service MUST publish 与 App API 等价的 WS 载荷。
- 消除 `updated=false` 时的静默成功。
- 统一三处重复 end 逻辑，降低后续遗漏。

**Non-Goals:**

- 语音 `start` / `one` 路径（用户反馈的「记录睡觉计时无 WS」留待后续变更）。
- 修改 `gateway-app` Redis 订阅实现或 `redismsgkit` 配置来源。
- 新增 App HTTP 接口或变更 WS 帧格式。
- history-service 修改 `EndLatest` 匹配语义（仍由 eventId + latest 行决定）。

## Decisions

### 1. 结束落库策略：先 EndLatest，失败再降级

```
end 动作 + 已解析叶子 event
    │
    ▼
EndLatestHistoryIfMatch(deviceNo, event.Id, now, remark)
    │
    ├─ err != nil → 返回失败，不播报成功
    ├─ updated == true → 播报「已结束」；WS action=update ✓
    └─ updated == false → 走 fallback（见下）
```

- **理由**：以 history-service 为权威，避免 voice 侧 `GetLatestHistory` 缓存预判 `eventId` 相等/不等导致误分支。
- **替代方案（未采用）**：结束前强制 bypass 缓存再比较 `eventId`——仍有两份真值，不如直接以 EndLatest 结果为准。

### 2. fallback：复用现有「非同事件结束」语义

当 `updated=false`：

1. `GetLatestHistory`（用于判断是否需要自动结束上一条；读缓存可接受，仅影响附加 update）
2. `AddHistory`：`startTime=endTime=now`，写入当前事件瞬时结束 → publish `create`
3. 若 `lastEvent.EndTime==0 && lastEvent.EventId>0 && lastEvent.EventId != event.Id`：对 `lastEvent.EventId` 再调 `EndLatestHistoryIfMatch`；若仍 `updated=false` 且 `lastEvent.Id>0`，**MAY** `UpdateHistory` 设 `endTime=now`（保证上一条闭合且 publish update）

- **理由**：与现有产品语义一致（新结束 + 自动闭合上一条计时）；且每次成功写库均走 history-service publish。

### 3. 抽取 `applyVoiceEventEndHistory`

- **位置**：`internal/services/voice/event_history_end.go`（或同级命名）
- **签名**：`applyVoiceEventEndHistory(ctx, deviceNo string, event entity.Event, targetName, remark string, nowTime int64) (reply string, err error)`
- **调用方**：`handleUnifiedIntentAction`、`handleActionRecord`、`applyEventActionTarget` 的 `end` case。

### 4. 用户可见错误语义

- `EndLatest` / `AddHistory` / `UpdateHistory` 任一步返回 error → 沿用现有「更新结束时间失败,请重试」类文案，**不**播报成功。
- fallback 中上一条自动结束失败：主记录已写入时，播报「已记录 X 结束，Y 结束失败,请手动结束」（与 `handleActionRecord` 现有文案对齐）。

## Risks / Trade-offs

- [EndLatest 先调一次额外 HTTP] → 每次结束多 1 次 remote 调用；可接受，结束非高频热路径。
- [fallback AddHistory 产生瞬时记录] → 与现有「非同事件 end」一致；无新语义。
- [GetLatestHistory 缓存用于 auto-end 判断] → 极端并发下 auto-end 可能漏闭合；与改前行为同级，不在本变更消除。

## Migration Plan

- 部署 `voice-service` 即可；history-service / gateway-app 无强制联动。
- 回滚：还原 voice end 分支旧逻辑。

## Open Questions

- 无。
