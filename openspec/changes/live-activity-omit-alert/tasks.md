## 1. APNs 组装层省略 alert

- [x] 1.1 在 `push_live_activity_apns.go` 的 `buildLiveActivityAPNsBody` 注释或移除写入 `aps.alert` 的分支，并加中文注释说明锁屏只靠 content-state、alert 已省略
- [x] 1.2 确认 start/update/end 及 invalid→restart 共用该组装函数后请求体均无 `alert` 键

## 2. Refresh 停填伴随文案

- [x] 2.1 在 `live_activity_refresh.go` 注释或删除 upsert 的 `AlertTitle`/`AlertBody` 赋值
- [x] 2.2 若 `liveActivityAlertBody` 无其它引用，注释或删除该 helper，并注明原因

## 3. 验收

- [x] 3.1 触发一次非空 Refresh，核对 push-service `[push-la] apns_body` Info 日志：有 `content-state`、无 `"alert"`
- [ ] 3.2 真机确认：锁屏 Live Activity 仍更新；无提示音/横幅；空列表 end 仍可摘条
