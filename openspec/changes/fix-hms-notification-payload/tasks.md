## 1. HMS 载荷

- [x] 1.1 重写 `HmsSender.Send` 非静默+有 alert 的请求体：顶层 `message.notification`（title/body）+ `android.notification`（click_action 默认 type=3，badge；可选 `PUSH_HMS_CLICK_INTENT`→type=1+intent）+ `android.data`
- [x] 1.2 静默/无 alert 路径：不伪造完整可见通知；保留 badge/data 语义（中文注释说明分支）

## 2. 成功判定与日志

- [x] 2.1 HTTP 2xx 后解析响应 `code`：仅 `80000000` 视为成功；否则返回 error，供 dispatcher 打 `send_failed`（含 code/截断 msg，无 token）
- [x] 2.2 保留/加强 invalid token 识别，与删除 token 逻辑兼容

## 3. 文档与自检

- [x] 3.1 `.env.example`（及需要时 prod 注释）补充可选 `PUSH_HMS_CLICK_INTENT` 说明
- [x] 3.2 `go build ./cmd/push-service`；确认未新增 `*_test.go`
- [x] 3.3 发版后人工：触发 predict-imminent，客户端应见通知；push 日志 `send_ok` 或带真实 `code` 的 `send_failed`
