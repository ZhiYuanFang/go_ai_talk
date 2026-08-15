## 1. API 与契约

- [x] 1.1 在 `api/v1/device_history_http.go` 的 `DeviceHistoryFilterReq` 增加 `ignoreTimeRange`（bool 或经确认的兼容类型），`dc` 注明默认 false、为真时完全忽略 start/end
- [x] 1.2 扩展 `contracts.HistoryService.ListHistoryFilter` 及所有实现签名（local / remote / switchAdapter）增加 `ignoreTimeRange bool`
- [x] 1.3 `historyRemoteClient.ListHistoryFilter` 在为真时透传 query `ignoreTimeRange`

## 2. 筛选实现

- [x] 2.1 `ListDeviceHistoryFilter`：当 `ignoreTimeRange` 为真时跳过 start/end WHERE；为假时保持现网逻辑；补充中文注释
- [x] 2.2 `HistoryCtrl.Filter` 解析并下传该参数；必要时对 `"1"`/`"true"` 做轻量规范化

## 3. 调用方与验收

- [x] 3.1 检索 voice/App 是否已调用 `ListHistoryFilter` 或 filter HTTP；若有读历史映射则对称接线 `ignoreTimeRange`，若无则仅保证 HTTP 契约可用并在 PR 说明
- [x] 3.2 向负责人确认 usage：本变更默认不改 `maintenance_skip.go`；若要求排除/计入则按答复处理
- [x] 3.3 手工或日志验收：未传开关 + 有时间窗 = 旧行为；`ignoreTimeRange=true` + 非零时间 = 无时间过滤且 limit/排序正确；remote/canary 透传一致
- [x] 3.4 确认无新测试文件、无新增 Redis 读缓存、无跨域 DAO；评审对照 `history-filter-api` 增量 spec 与 v3.0.0 基线章节
