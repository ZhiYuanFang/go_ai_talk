## 1. Timing 载荷

- [x] 1.1 `maybeStartTimingLiveActivity` 增加 `eventID` 参数，Attributes 写入 `eventId`；`local.go` 调用传入 `item.EventId`
- [x] 1.2 确认并固定 ContentState `startTime` 为 unix 秒（核对落库/`item.StartTime` 无毫秒混入）；禁止在 LA map 中写入 logo/color

## 2. Imminent 复核

- [x] 2.1 复核 `startImminentLiveActivityForWx`：`eventId`/`nextAt` 为秒；载荷无 logo/color

## 3. 验收

- [x] 3.1 与 Flutter `LIVE_ACTIVITY.md` / `live-activity-lock-ui` 字段表对照一致
- [x] 3.2 联调说明：Timing 真机应能按 eventId 出图/色；`startTime` 正确则已计时非恒 `00:00`

### 核对纪要（实现时）

| Flutter `LIVE_ACTIVITY.md` | Go 现状 |
|----------------------------|---------|
| Timing attrs: `activityId`/`deviceNo`/`historyId`/`eventId` | `live_activity_timing.go` 已齐 |
| Timing cs: `eventName`/`startTime`(unix 秒)/`phase` | 已齐；`startTime` 透传 `item.StartTime`（API dc 为 Unix 秒；`piece.go` 亦注明库内 BIGINT 为秒） |
| Imminent attrs: `activityId`/`deviceNo`/`eventId`/`nextAt` | `predict_live_activity.go` 已齐 |
| Imminent cs: `title`/`phase`/`nextAt`(unix 秒) | 已齐；MQ delay 用 `(nextAt-now)*1000` 证 `nextAt` 为秒 |
| 无 `logo`/`color`/`logoFile` | Timing/Imminent map 均未写入 |

**真机联调**：`PUSH_LIVE_ACTIVITY_ENABLED=true` + 已注册 Timing Push-to-Start；落一笔未闭合 history（`eventId>0`）→ 锁屏应出事件图/色，右侧已计时非恒 `00:00`。
