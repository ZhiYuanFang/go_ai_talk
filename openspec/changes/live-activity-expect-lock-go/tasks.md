## 1. Imminent ContentState

- [ ] 1.1 扩展 `startImminentLiveActivityForWx`（或等价）接收 expect 结构化参数；有值才写入 ContentState
- [ ] 1.2 开火路径把 MQ/pending 的 `expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit` 传入；不写 leaf、不写中文 label

## 2. Timing + active-eta

- [ ] 2.1 Timing ContentState 契约允许可选 `expectEndAt`（秒；>0 才写）
- [ ] 2.2 `SyncPredictActiveEta` 成功后：按 `historyId`+`expectEndAt` 扇出 Timing update（带齐 phase/title/startAt 尽量不丢字段）；空列表不清 end Timing
- [ ] 2.3 修正注释/文档：active-eta 禁止 MQ/普通 alert，但允许 LA update

## 3. 联调

- [ ] 3.1 对照 Flutter twin `live-activity-expect-lock`：有 expect 的 Imminent/Timing 推送 ContentState 键齐全；无 expect 省略键
- [ ] 3.2 手工：开计时 → sync active-eta → 锁屏副行出现；imminent 开火带 duration/amount → 锁屏右侧有字
