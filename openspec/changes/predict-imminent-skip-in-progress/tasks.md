## 1. History 进行中存在性契约

- [x] 1.1 在 `contracts.DeviceHistoryContract` 增加 `HasOpenHistory(ctx, deviceNo, eventIds []int64) (bool, error)`（或等价命名），语义：任一 eventId 存在 `end_time=0` 即 true；空 eventIds 返回 false
- [x] 1.2 history-service：实现 DB 查询（`device_no` + `event_id IN` + `end_time=0`，Exist/One，最少读）并挂到 local / switchAdapter
- [x] 1.3 暴露 internal/HTTP 读接口（路径对齐现有 `/device/history/api/event/*`），controller 绑定；`internal/clients/history` 实现远程调用
- [x] 1.4 确认 voice→history 走契约客户端，**无**跨库直查；不改已有 ListHistoryFilter 响应结构

## 2. Voice 叫醒闸顺序

- [x] 2.1 在 `handlePredictImminentFire`：Redis 匹配成功后、五分钟 `SetNX` 去重**之前**，调用展开后的 `HasOpenHistory`；命中则日志 `reason=in_progress`（或等价）、Ack skip、不写去重键
- [x] 2.2 进行中查询错误：fail-closed（不推、Ack、不占去重）；保留其后既有去重 + 30 分钟 `start_time` 闸与推送逻辑
- [x] 2.3 根事件展开复用 `expandEventIDsForHistory`；不同根互不抑制（由 eventId 集合自然保证）

## 3. 验收与边界

- [x] 3.1 手工/日志场景：长时段 `end_time=0`（start 早于 30 分钟）叫醒被 skip；结束后未占去重时可再次进入后续闸
- [x] 3.2 确认**不新增** App 对外路由、**不新增** Redis 业务读缓存键、**不新增**背景 ticker；发版顺序建议 history 先于或同于 voice
