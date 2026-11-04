## Why

兄弟仓 Flutter change `live-activity-lock-ui` 已按 `eventId` 本地解析锁屏事件图/色，且 Timing 锁屏用系统计时展示已计时。当前 Go Timing start 固定 Attributes **未带** `eventId`（仅 `activityId`/`deviceNo`/`historyId`），客户端无法出事件视觉；若 `startTime`/`nextAt` 误用毫秒，锁屏会卡在 `00:00`。需在服务端补齐契约，且 **不得** 在推送中下发 logo/color。

## What Changes

- **Timing start Attributes**：增加 `eventId`（来自 history 行；与 Flutter catalog `def.id` 十进制对齐）。
- **时间单位**：Imminent / Timing 载荷中的 `nextAt`、`startTime`（及 Attributes 上的 `nextAt`）MUST 为 **unix 秒**（非毫秒）；与 history 表既有秒级字段语义对齐并在实现中核对调用点。
- **禁止视觉字段**：Live Activity start/update ContentState 与 Attributes MUST NOT 携带 `logo` / `color` / `logoFile` / 远程图 URL。
- **Imminent**：已含 `eventId`；本变更复核 `nextAt` 为秒，并明确不推 logo/color。
- **非目标**：不改 1/min、pending 清场、token 注册、Flutter UI。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `predict-timing-live-activity-push`：Timing start 固定字段 MUST 含 `eventId`；`startTime` MUST 为 unix 秒；不得推送 logo/color。
- `predict-imminent-live-activity-push`：明确 `nextAt`/`eventId` 契约与不得推送 logo/color（行为补充，与 Flutter `live-activity-lock-ui` 对齐）。

> 上述 capability 来自未归档 change `predict-live-activity-push`；本变更对其做增量 Requirement。

## Impact

- `internal/services/history/live_activity_timing.go`：`maybeStartTimingLiveActivity` 签名与 Attributes 补 `eventId`；调用点传入 `item.EventId`
- `internal/services/voice/predict_live_activity.go`：复核 Imminent `nextAt`/`eventId`；确保无 logo/color
- 文档/注释与 Flutter `LIVE_ACTIVITY.md` 对齐
- 兄弟仓：`flutter_ai_talk` / `live-activity-lock-ui`（客户端已实现；本仓不改 Flutter）
- 无新 DB 表 / 无新 HTTP path / 无 Android 要求
