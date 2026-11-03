## ADDED Requirements

### Requirement: Imminent fire MUST apply user one-per-minute gate before send

After existing Redis match, in-progress, per-event dedup, and recent-history gates have been evaluated for an Imminent fire, and before a successful user-visible or Live Activity Imminent fanout, the twin Go service MUST apply the per-user rolling 60-second first-wins gate defined in `predict-imminent-live-activity-push`. A user who fails that gate MUST NOT receive that fire’s Imminent push and MUST Ack the delivery without treating it as a requeueable failure. Timing Live Activity traffic MUST NOT use this gate.

在既有 Redis 匹配、进行中、按事件去重、近史闸评估之后、成功扇出 Imminent 之前，Go MUST 再施加 `predict-imminent-live-activity-push` 中的每用户 60 秒先到占坑闸。未过闸 MUST NOT 发送该次 Imminent，MUST Ack 且 MUST NOT 作为可 requeue 失败。Timing MUST NOT 使用本闸。

#### Scenario: Rate limit after other gates

- **WHEN** 某开火已通过 Redis 匹配与进行中/去重/近史闸，但该用户 60 秒内已成功扇出过 Imminent
- **THEN** Go MUST NOT 再向该用户发送本次 Imminent，MUST Ack

### Requirement: iOS Imminent fire delivery MUST follow live-activity-push rules

When fanout targets an iOS/APNs recipient, the send path MUST follow `predict-imminent-live-activity-push` (Live Activity, not normal predict alert). Non-iOS recipients MAY keep legacy visible predict-imminent notifications subject to the same user one-per-minute gate when that gate applies to Imminent fanout.

扇出目标为 iOS/APNs 时，发送 MUST 遵循 `predict-imminent-live-activity-push`（Live Activity，非普通预测 alert）。非 iOS MAY 保留遗留可见预测通知，且在适用时 MUST 同受用户 1/min 闸约束。

#### Scenario: Schedule delegates iOS carrier

- **WHEN** 开火对 iOS 用户进入发送阶段且已过全部闸（含 1/min）
- **THEN** 发送 MUST 走 Live Activity 路径而非普通 `predict_imminent` alert
