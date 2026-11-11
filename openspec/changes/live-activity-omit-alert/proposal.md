## Why

统一列表态 Live Activity 的 upsert/start 目前会附带 `aps.alert`（`AlertTitle`/`AlertBody`），系统会额外出横幅并常伴随提示音；锁屏条本身由 `content-state` 驱动，并不依赖 alert。产品希望 Live Activity 远程推送不再「喊一声」，只静默更新锁屏 / Dynamic Island。

## What Changes

- Live Activity APNs 组装层在存在 `content-state` 的 start/update/end 路径上 **MUST NOT** 再写入 `aps.alert`（含 title/body/sound）；实现上注释或移除现有 alert 拼装逻辑。
- history Refresh upsert **MUST NOT** 再填充 `AlertTitle`/`AlertBody`（注释或清空调用）；`end` 本就无 alert，行为不变。
- ContentState / Attributes / `event` / `dismissal-date` / token 解析与 invalid→restart **不变**；锁屏列表展示继续只靠 `content-state`。
- **非目标**：不改 Android / 非 iOS 的普通 `predict_imminent` 可见推送；不改 UCG 等非 LA 通道；不引入静音音效资源。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `unified-live-activity-push`：统一 Live Activity 远程推送 MUST 省略 `aps.alert`；Refresh 扇出不再携带伴随通知文案。

## Impact

- `internal/services/push/push_live_activity_apns.go`：注释/停用 `aps.alert` 写入。
- `internal/services/history/live_activity_refresh.go`：注释/停用 upsert 的 `AlertTitle`/`AlertBody`。
- 进程：需重部署 **push-service**（APNs 组装）与 **history-service**（Refresh 入参）；gateway / voice / device 不必仅为该变更重部署。
- 风险：个别 iOS 版本对 **push-to-start 且完全无 alert** 的呈现可能变差；需真机验收 start/update/end。
