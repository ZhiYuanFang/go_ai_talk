## 1. 投影拼装

- [x] 1.1 在 `buildLiveActivityItems`（或等价）读取设备 nick：经既有 `GetBirthday`/profile，trim 空则「宝宝」；失败 WARN 并回落「宝宝」
- [x] 1.2 对选中的 imminent 项：pending.title trim 非空时写入 `title`=`{nick}要{事件}了`；事件名为空则省略 `title` 键
- [x] 1.3 确认 `applyPendingExpect` 仍独立写 expect*；title 不含「预计」/leaf/普通推送后缀

## 2. 边界与回归

- [x] 2.1 确认 voice 普通推送 alert 拼装未改；LA 拼装不依赖 fire 路径
- [x] 2.2 确认 `liveActivityAlertBody` 取首条 imminent title 时自然为成品句
- [x] 2.3 `go build` 相关包（至少 history）通过

## 3. 对照验收

- [x] 3.1 pending PUT Refresh：锁屏 imminent title 为`{nick}要{事件}了`，右槽仍仅来自 expect*
- [x] 3.2 无宝宝名时 title 为`宝宝要{事件}了`
