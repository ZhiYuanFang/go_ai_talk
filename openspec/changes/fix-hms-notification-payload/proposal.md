## Why

华为控制台下发通知客户端能收到，但 push-service 对同一设备 `send_ok channel=hms` 后客户端无通知。根因高度指向 HMS 请求体与官方「通知消息」形态不一致（缺顶层 `message.notification`、`click_action.type=1` 无 `intent`），以及成功判定只认 HTTP 2xx、不校验业务 `code`（可能假成功）。

## What Changes

- 修正 HMS 非静默推送载荷：对齐华为 Push Kit 通知消息结构（顶层 `message.notification` + `android.notification.click_action`；默认 `type:3` 打开应用，或可配置完整 intent）。
- 静默/仅 badge 路径保持可区分（不伪造成完整通知栏文案）。
- 自定义业务字段仍可通过 `data` 传递（与通知并存时不破坏控制台可比的通知展示）。
- HMS 响应 MUST 校验业务成功码（如 `80000000`）后才记 `send_ok`；否则 `send_failed` 并打出 `code`/摘要（脱敏）。
- 可选：Warning 日志带上业务 `code`，便于对齐控制台与服务端。

## Capabilities

### New Capabilities

- `hms-notification-payload`：HMS 通知/静默载荷形态与成功判定契约。

### Modified Capabilities

- （无）既有 `openspec/specs/` 为版本快照；本变更新增 change 内能力规格。

## Impact

- 代码：`internal/services/push/push_hms.go`（及必要时 dispatcher 日志语义）。
- 行为：客户端应能收到与控制台类似的系统通知；原先假 `send_ok` 可能变为 `send_failed`（暴露真实业务错误）。
- 无 API/表结构 BREAKING；不影响 APNs/MiPush（除非发现同类问题另开变更）。
- 配置：若需自定义 `click_action` intent，可用可选 env（如 `PUSH_HMS_CLICK_INTENT`）；默认 `type:3` 无需配置。
