## ADDED Requirements

### Requirement: Server SHALL rate-limit Imminent fanout one per user per minute first-wins

When the twin Go service is about to successfully fan out an Imminent Live Activity start (or a non-iOS legacy visible predict-imminent push) to a bound user, it MUST enforce at most one successful Imminent fanout per user per rolling 60-second window. The first eligible fanout in the window MUST proceed; later eligible Imminent fires for that user within the window MUST NOT push. Timing Live Activity start/end MUST NOT consume or be blocked by this gate. Existing per-event dedup, in-progress, and recent-history skip gates MAY remain in addition.

对绑定用户即将成功扇出 Imminent（或非 iOS 遗留可见预测推送）时，Go MUST 每用户滚动 60 秒至多一次成功扇出；窗口内首次 MUST 执行，后续 MUST NOT 再推。Timing 起停 MUST NOT 占用或受阻于该闸。既有按事件去重、进行中、近史跳过闸 MAY 并存。

#### Scenario: Second Imminent in same minute dropped

- **WHEN** 同一用户在 60 秒内已成功扇出一条 Imminent，又有另一事件通过既有开火闸
- **THEN** Go MUST NOT 为后一次再推 Imminent（或该用户的等价预测提醒）

#### Scenario: Timing unaffected

- **WHEN** 同一用户在 Imminent 1/min 窗口内需扇出 Timing Live Activity start 或 end
- **THEN** Go MUST 仍按 Timing 规则推送

### Requirement: Server SHALL clear all Imminent Live Activities on pending sync

On each successful handling of `PUT /device/api/predict/imminent/pending` for a device, the twin Go service MUST send Live Activity end (or equivalent dismiss) for all tracked Imminent Live Activities to bound users of that baby/device before or as part of rewriting Redis schedules, and MUST NOT end Timing Live Activities as part of this clear. After clear, existing pending write and delayed-fire scheduling SHALL remain.

每次成功处理某宝宝的 imminent pending 全量同步时，Go MUST 先向绑定用户结束全部已跟踪 Imminent Live Activity，MUST NOT 借此结束 Timing；其后既有写 pending 与延时开火 MUST 保持。

#### Scenario: Pending PUT clears Imminent

- **WHEN** 客户端提交 `PUT …/predict/imminent/pending`
- **THEN** Go MUST 向该 deviceNo 绑定用户发出清空/结束全部 Imminent Live Activity 的推送（或等价）
- **AND** MUST NOT 因此结束 Timing Live Activities

### Requirement: Server SHALL keep a single Imminent slot per user when starting

When Go is about to start a new Imminent Live Activity for a user and an Imminent slot may still be active for that user, Go MUST end the previous Imminent before or as part of presenting the new content so that at most one Imminent Live Activity is intended per user (wxId). Bound guardians of the same baby/device MUST each have their own Imminent slot; clearing and starting MUST fan out per bound wx. Fanout that fails the one-per-minute gate MUST NOT perform this replace start.

即将 start 新 Imminent 且该用户旧槽可能仍在时，Go MUST 先 end 旧 Imminent（或同槽更新），使每用户（wx）意图上至多一条。同一宝宝多监护人 MUST 各自独立槽；清场与开火 MUST 按绑定 wx 列表分别处理。未过 1/min 闸 MUST NOT 执行该换新 start。

#### Scenario: Fire replaces previous Imminent

- **WHEN** Go 对某用户通过 1/min 与既有开火闸并即将 start Imminent，且 Redis（或等价）显示仍有旧 Imminent activity id
- **THEN** Go MUST 先结束旧 id（或同槽更新为最新内容）再呈现单槽

#### Scenario: Multiple guardians each get a slot

- **WHEN** 同一 deviceNo 绑定两个 wx，且开火通过闸并对两用户均有可用 LA token
- **THEN** Go MUST 允许两名用户各自拥有一条 Imminent Live Activity（非共享单一 device 槽）

### Requirement: Live Activity tokens SHALL use a dedicated App path and storage

The twin Go service MUST expose a new App-facing HTTP path (separate from `POST /app/api/push/register`) for registering and updating Live Activity Push-to-Start and per-activity update tokens. Those tokens MUST be stored in a push-domain table (or equivalent) distinct from `push_device` device-token uniqueness (`uk_token`). The existing device push register path MUST continue to store only vendor/device alert tokens. Live Activity tokens MUST NOT be written into `push_device.token` under the global device-token unique-key semantics.

Go MUST 提供并列于 `/app/api/push/register` 的新 App HTTP 路径，用于注册/更新 Live Activity 的 Push-to-Start 与 per-activity update token；MUST 落在 push 域分表（或等价），MUST NOT 写入 `push_device` 的 device `uk_token` 语义。既有 register MUST 仅继续承载厂商/设备 alert token。

#### Scenario: LA token not stored as device push token

- **WHEN** 客户端上报 Push-to-Start 或 per-activity update token
- **THEN** Go MUST 经新 path 持久化到 LA 专用存储
- **AND** MUST NOT 将该 token 作为 `push_device` 全局唯一 device token 写入

#### Scenario: Device register unchanged for alerts

- **WHEN** 客户端继续调用既有 `/app/api/push/register` 上报 APNs/HMS/MiPush device token
- **THEN** 行为 MUST 保持 device alert 注册语义，MUST NOT 要求该调用携带 LA token 才能工作

### Requirement: iOS Imminent delivery SHALL use Live Activity not alert

For APNs/iOS recipients, predict-imminent offline reminders that pass all gates MUST be delivered as ActivityKit Live Activity pushes (Push to Start / liveactivity update as applicable), and MUST NOT use a normal visible APNs alert with `bizType=predict_imminent` as the carrier on that path. Non-iOS channels MAY continue visible vendor notifications. Missing Live Activity / Push-to-Start token MUST skip that recipient without failing the whole fire Ack path.

对 iOS/APNs 接收方，通过全部闸的预测临近 MUST 以 Live Activity 推送送达，MUST NOT 以普通可见 `predict_imminent` alert 为载体。非 iOS MAY 继续可见通知。无 LA/Push to Start token MUST 跳过该接收方，MUST NOT 因此使整次开火 Ack 失败路径崩溃。

#### Scenario: iOS uses liveactivity

- **WHEN** Go 向已注册 Push to Start（或等价）的 iOS 设备扇出 Imminent
- **THEN** 发送 MUST 为 liveactivity 路径，MUST NOT 依赖普通预测 alert

#### Scenario: No token skips recipient

- **WHEN** 某绑定用户无可用 Live Activity / 推送 token
- **THEN** Go MUST 跳过该用户并继续处理其余接收方
