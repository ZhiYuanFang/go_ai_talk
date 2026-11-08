## Why

联调 Live Activity 时只能看到 `send_ok` / `send_failed` 元数据，看不到真正发往 APNs 的 JSON（`content-state.items`、start attributes、dismissal-date 等），难以核对 lead 窗截断与 update/start 载荷。需要在发送前用 **Info** 打出完整请求体。

## What Changes

- 在 `SendLiveActivityAPNs` 组装 HTTP body 之后、请求 APNs 之前，以 **Info** 级别打印最终消息体（含 `event`、`tokenLen`、body JSON）。
- MUST NOT 将完整 deviceToken 写入日志（仅长度或等价脱敏）。
- **仅加日志**；不改发送语义、截断规则、token 自愈、register 契约。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `unified-live-activity-push`：补充「每次向 APNs 发送 Live Activity 前 MUST 以 Info 记录最终请求体」的可观测要求。

## Impact

- `internal/services/push/push_live_activity_apns.go`（单点，覆盖 update / start / end / invalid→restart start）。
- push-service `logger.level: all` 下 Info 可见；无新接口 / 无 usage 变更。
