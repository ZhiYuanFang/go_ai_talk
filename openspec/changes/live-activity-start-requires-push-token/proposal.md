## Why

App 关闭消息通知通常会 `unregister` 普通 `push_device`，但 Live Activity 仍可用残留的 `push_to_start` 被 start 出来，造成「通知关了锁屏还在跳」。暂不改 Flutter、不上正式偏好位：用「该 wx 是否仍有任意 push_device」近似代理通知开关，并只拦截 **start**。同时产品确认 **update invalid→同拍 restart 路径用不到**，一并删除以减复杂度。

## What Changes

- Live Activity 在解析为 `event=start`（含 `action=start` 与 `upsert→start`）发送前，MUST 检查该 `wxId` 是否存在至少一条 `push_device`（**任意 channel**）；无则 skip start 并打可观测日志，MUST NOT 发 APNs start。
- `event=update` / `event=end` **MUST NOT** 受此闸约束。
- **BREAKING（服务端 LA 故障恢复）**：删除 `update` 遇 invalidToken 后同拍 `restartLiveActivityViaPushToStart`；invalid 时仍可删除脏 activity 行，但 MUST NOT 同拍补 start。后续呈现依赖客户端再 register / 下次 upsert 自然走 start（且仍受 push_device 闸）。
- **非目标**：不改 Flutter；不新增用户偏好 API；不按 deviceKey 精确对齐；不闸 update/end；不改普通 `PushByBizType`。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `unified-live-activity-push`：start 前必须有 push_device；移除 invalid-update 同拍 restart 要求。

## Impact

- `internal/services/push/push_live_activity.go`：start 前查 `ListPushDevicesForWx`；删除 `restartLiveActivityViaPushToStart` 及其调用。
- 进程：仅 **push-service** 需重部署；history / gateway / Flutter 不必因本变更重部署。
- 近似语义：多机时只要任一设备仍有 push_device，该 wx 的 start 仍可能发出（含仍挂 LA token 的已关通知机）；接受为暂时方案。
