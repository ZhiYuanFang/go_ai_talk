## 1. 去重 TTL 对齐 leadSeconds

- [x] 1.1 在 `predict_imminent.go`：去掉固定 `predictImminentPushDedupTTL`；SetNXEX 使用 `max(1, predictImminentLeadSeconds())` 秒作为 TTL；中文注释说明与 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS` 对齐及 LEAD=0 兜底
- [x] 1.2 更新 `internal/platform/cachekit/keys_voice.go` 中 `PredictImminentPushedKey` 注释（勿再写死「五分钟」）
- [x] 1.3 确认未改动：history 1800s、进行中闸、lead 读取逻辑、远 delay 跳过

## 2. 自检

- [x] 2.1 代码审阅：默认 LEAD=300 时行为等价原 5 分钟去重；LEAD=600 时 TTL 为 10 分钟
