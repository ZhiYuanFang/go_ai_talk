## Why

预测临近多事件同时到点时，按用户连发多条可见推送易刷屏；锁屏缺少服务端可起停的「即将发生 / 计时中」通道。客户端（见兄弟仓 `flutter_ai_talk` change `predict-ios-live-activity`）将用 iOS Live Activity 承接展示，本仓负责限流、清场、Live Activity 推送与计时 history 起停扇出。

## What Changes

- **Imminent 用户级限流**：同一用户滚动 60 秒内至多成功扇出一次 Imminent；**谁先到谁占坑**，窗口内后续不推。Timing start/end **不受**此闸。
- **pending 清场**：每次成功处理 `PUT /device/api/predict/imminent/pending` 时，MUST 先向该宝宝绑定用户结束全部 Imminent Live Activity，再写 Redis / 重布延时；MUST NOT 误 end Timing。客户端有变化才 PUT，服务端按次清即可。
- **单槽换新（每 wx）**：开火真推 Imminent 前若该用户仍有旧 Imminent，MUST 先 end 再 start（或同槽更新）。多监护人同宝宝时 **各 wx 各亮一条**。
- **iOS 载体**：**BREAKING（iOS 送达形态）**：预测临近对 iOS MUST 走 Live Activity（Push to Start / liveactivity），MUST NOT 再发普通可见 `predict_imminent` APNs alert。Android 等厂商通道 MAY 继续可见通知。
- **Timing**：history **单次持久化成功**路径：未闭合计时 create → Timing LA start；写入 endTime 闭合 → 同 activity id end。多 history 可并存。MUST NOT 在多处钩子双发。不设固定秒级「确认仍在计时」延迟。
- **Token**：**新 App HTTP path + push 域分表** 存储 Push-to-Start / per-activity update token；MUST NOT 写入 `push_device` 的 device `uk_token` 语义。既有 `/app/api/push/register` 不变。
- **APNs**：复用现有 `.p8`；发送侧使用 `liveactivity` push-type 与 `{bundle}.push-type.liveactivity` topic；一般无需 Apple Developer 新证书。
- **边界**：不改客户端 pending JSON 字段契约（除非 LA 载荷另需文档化扩展）；active-eta Redis 同步可不直接推送。

## Capabilities

### New Capabilities

- `predict-imminent-live-activity-push`：Imminent Live Activity 扇出、1/min、pending 清场、per-wx 单槽、iOS 取消普通预测 alert、LA token 新 path/分表。
- `predict-timing-live-activity-push`：history 单次落库开/闭驱动 Timing Live Activity start/end；多槽；不受 Imminent 1/min。

### Modified Capabilities

- `predict-imminent-schedule`：开火路径在既有闸之外增加用户级 1/min；iOS 送达改为 Live Activity；与 pending 清场/单槽规则衔接（细节以新 capability 为准，本能力补充调度/闸顺序）。

## Impact

- **voice-service**：`predict_imminent.go`（限流、清场调用、开火 LA）
- **push-service / APNs**：liveactivity 发送；LA token 新表与新 App register path；复用 `PUSH_APNS_*`
- **gateway-app**：新 push LA 路径反代 / Bearer / usage（问负责人）
- **Redis（cachekit）**：用户 1/min 键、Imminent/Timing activity id 跟踪
- **history-service**：create/update 闭合后 **一次** 触发 Timing 扇出（经 clients/push）；禁止与 voice 双钩
- **兄弟仓**：`flutter_ai_talk` / `predict-ios-live-activity`（Extension、token 上报、UI、`NSSupportsLiveActivities`）；本提案不实现 Flutter
- **无 Android 锁屏要求**
