## Why

统一列表态 Live Activity 的 upsert 若一律附带 `aps.alert`，update 会反复横幅/提示音；但真机确认：**push-to-start 的 `event=start` 若无 `aps.alert`，系统不会呈现 Live Activity**。因此需要「start 必带 alert 才能弹出；update/end 省略 alert 以静默刷新」。

## What Changes

- APNs 组装层：**仅当 `event=start`** 写入 `aps.alert`（title/body；调用方为空时用默认兜底）；`update` / `end` **MUST NOT** 写 `aps.alert`。
- history Refresh upsert **恢复**填充 `AlertTitle`/`AlertBody`（供解析成 start 时使用）；组装层对 update 忽略这些字段。
- invalid→restart 的 start 与首次 push-to-start 同规则（有 alert）。
- ContentState / Attributes / dismissal-date / token 解析 **不变**。
- **非目标**：不改 Android 普通 `predict_imminent`；不引入静音音效资源；不在本变更内强制 start 无声。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `unified-live-activity-push`：Live Activity APNs 仅 start 含 `aps.alert`；update/end 省略；Refresh 继续提供 start 所需伴随文案。

## Impact

- `internal/services/push/push_live_activity_apns.go`：按 event 条件写 alert。
- `internal/services/history/live_activity_refresh.go`：恢复 upsert 的 `AlertTitle`/`AlertBody` 与 `liveActivityAlertBody`。
- 进程：需重部署 **push-service** 与 **history-service**；gateway / voice / device 不必仅为该变更重部署。
