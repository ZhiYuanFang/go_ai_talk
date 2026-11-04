## Why

用户改未闭合计时 history（开始时间、事件名等）时，锁屏 Timing Live Activity 仍停在 start 那一帧。`UpdateDeviceHistory` 仅在写入 `endTime` 时 end，未闭合 update 不发 `action=update`。

## What Changes

- **未闭合 update 扇出 Timing update**：`UpdateDeviceHistory` 成功且 `endTime==0` 时，MUST 按 `timing:{deviceNo}:{historyId}` 向绑定用户发 Live Activity **update**（activity token）。
- **只改 ContentState**：`eventName`、`startTime`（unix 秒）、`phase=timing`。MUST NOT 为换 `eventId` 而 end+start；图标可沿用 start 时 Attributes。
- **仍闭合则只 end**：`endTime!=0` 继续走既有 end，MUST NOT 再 update。
- **非目标**：不改 Imminent、删除/闭合路径、token 分层、Flutter Attributes 解码（state 非空即覆盖标题/计时）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `predict-timing-live-activity-push`：未闭合计时 history 更新 MUST 扇出 Timing ContentState update；换事件类型不重建活动。

> capability 来自未归档 `predict-live-activity-push` / `fix-live-activity-end-dismiss`。v3.0.3 基线尚无此章。

## Impact

- `internal/services/history/live_activity_timing.go`：新增 update 扇出
- `internal/services/history/local.go`：`UpdateDeviceHistory` 未闭合分支调用
- 无新 HTTP path / 无新表 / 无 Flutter 必改（ContentState 已有 eventName/startTime）
