## Why

用户划掉锁屏 Live Activity 后，APNs activity update token 常立即失效，但库里仍留着 `token_kind=activity` 行。后续 Refresh 的 `upsert` 会继续走 `update` 失败，锁屏起不来，直到偶然清掉脏行。需要在 **update 判定为 token 失效** 时同拍删库并改用 `push_to_start` 重新 `start`。

## What Changes

- **失效判定用方案 B**：仅当 APNs 返回现有 `isApnsInvalidToken` 语义（如 410 Gone、BadDeviceToken、Unregistered、DeviceTokenNotForTopic）时，才视为「端侧活动已没了」；429 / 5xx / 网络等临时错误 **MUST NOT** 删 token、**MUST NOT** failover。
- **同拍自愈**：对 `update` 或 `upsert→update` 路径，invalid 后 MUST 删除该条 `activity` token；再对**同一 `device_key`** 的 `push_to_start` 发送 `event=start`（复用本次 ContentState / Attributes）。
- **范围限制**：`end` 失败 invalid 只删脏 token，**MUST NOT** 再 start；start 自身失败仅 WARN，禁止同循环无限重试。
- **非目标**：不放宽为「任意 APNs 错误都重启」；不改 register 契约、lead 窗截断、普通 alert 推送。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `unified-live-activity-push`：补充「activity update 在 invalidToken 时删行并同拍 push_to_start 重启」要求；临时错误不得触发该路径。

## Impact

- `internal/services/push/push_live_activity.go`（及必要时 APNs 发送返回值消费处）：update invalid → delete → same-device start。
- 可能需要按 `device_key` 列举 `push_to_start` 的查询辅助（若尚无）。
- history Refresh / voice 调用方契约不变（仍 `action=upsert`）。
- 无新 App HTTP；usage 不改。
