## 1. Info 日志

- [x] 1.1 在 `SendLiveActivityAPNs`：`buildLiveActivityAPNsBody` 成功后、`client.Do` 前增加 `Infof`，含 `event`、`tokenLen`、完整 body；禁止打印完整 deviceToken
- [x] 1.2 中文注释标明覆盖 update/start/end 及 invalid→restart 的 start

## 2. 校验

- [x] 2.1 `gofmt`；编译 `internal/services/push`
- [x] 2.2 确认未改发送分支语义、未改 history 截断
