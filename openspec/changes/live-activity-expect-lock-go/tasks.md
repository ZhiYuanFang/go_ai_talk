## 1. Imminent ContentState + 标题

- [x] 1.1 扩展 `startImminentLiveActivityForWx`（或等价）接收 expect 结构化参数；有值才写入 ContentState；不写 leaf、不写中文 label
- [x] 1.2 开火路径拆分：`laTitle="{昵称}要{事件名}了"` 进 LA；`alert=laTitle+预计后缀` 仅 `PushByBizType`；并把 MQ/pending 的 `expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit` 传入 LA

## 2. Timing + active-eta

- [x] 2.1 Timing ContentState 契约允许可选 `expectEndAt`（秒；>0 才写）
- [x] 2.2 `SyncPredictActiveEta` 成功后：按 `historyId`+`expectEndAt` 扇出 Timing update（带齐 phase/title/startAt 尽量不丢字段）；空列表不清 end Timing
- [x] 2.3 修正注释/文档：active-eta 禁止 MQ/普通 alert，但允许 LA update
- [x] 2.4 history `updateTimingLiveActivity`：只读 `PredictActiveEtaKey`，匹配 `historyId` 后回写 `expectEndAt`；禁止 import voice；无匹配则 omit

## 3. 联调

- [x] 3.1 对照 Flutter twin `live-activity-expect-lock`：有 expect 的 Imminent/Timing ContentState 键齐全；LA title 无「预计这次约」；无 expect 省略键
- [x] 3.2 手工：开计时 → sync active-eta → 副行出现 → 改计时名/开始时间后副行仍在；imminent 开火右侧有字且标题为基句
