## 1. 同步路径延时上限闸

- [x] 1.1 在 `internal/services/voice/predict_imminent.go` 增加显式常量（对齐 `4294967295` ms / 插件 `ERL_MAX_T`），附中文注释说明超限会 nodelay
- [x] 1.2 在 `SyncPredictImminentPending` 发布循环中：计算 `delayMs` 后若超过上限则 WARN 日志（含 eventId、跳过原因）并 `continue`；不超过则照常 `PublishDelayed`
- [x] 1.3 确认超限跳过不导致整表同步返回错误；近端条目仍可成功入队

## 2. 文档与验收

- [x] 2.1 （可选）在 `docs/runbooks/rabbitmq-local.md` 的 predict-imminent 小节补一句：超 ~49.7 天 delay 只写 Redis、不发 MQ
- [x] 2.2 自检：同批近端 + 远端（如 +2y）逻辑符合 specs 场景；不改 API / 消费路径 / appointment 表
