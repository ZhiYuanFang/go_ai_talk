## 1. APNs 组装层按 event 分流 alert

- [x] 1.1 在 `buildLiveActivityAPNsBody`：仅 `event=start` 写入 `aps.alert`；title/body 空则默认兜底；`update`/`end` 不写 alert（中文注释说明真机约束）
- [x] 1.2 确认 start（含 invalid→restart）走有 alert；update/end 无 alert

## 2. Refresh 恢复伴随文案

- [x] 2.1 恢复 upsert 的 `AlertTitle`/`AlertBody` 赋值
- [x] 2.2 恢复 `liveActivityAlertBody` 函数

## 3. 验收

- [x] 3.1 核对 `[push-la] apns_body`：start 含 `"alert"`；update/end 无 `"alert"`
- [ ] 3.2 真机：无 activity token 能弹出；已有 token 时 update 静默；空列表 end 可摘条
