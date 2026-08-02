## 1. History 权威匹配

- [x] 1.1 修改 `EndLatestDeviceHistoryIfMatch`（`internal/services/history/local.go`）：按 `device_no` + `event_id` + `end_time=0` 查最近一条（`ORDER BY id DESC LIMIT 1`），命中后更新 `end_time`（及可选 remark），保留 cache patch / bumpPiece / `publishHistoryChange(update)`
- [x] 1.2 补充中文注释：说明「按同 eventId 未闭合」语义，明确不再使用「全局 GetLatest 再比 eventId」
- [x] 1.3 核对 controller / contracts / remote adapter：路径与签名不变，仅行为跟随 local；必要时更新接口注释文案

## 2. Voice 路径核对

- [x] 2.1 核对 `applyVoiceEventEndHistory`：仍先 `EndLatestHistoryIfMatch`，`updated=false` 才降级 `AddHistory`；确认在「睡眠未闭合 + 中间其它事件」场景下不再误新建
- [x] 2.2 确认 `handleUnifiedIntentAction` / multi / pending child 的 `end` 均走同一 `applyVoiceEventEndHistory`，无需旁路改写

## 3. 验收与边界

- [x] 3.1 手工/日志验收：开始睡眠 → 记其它事件 → feeding+end 睡眠 → 原行 `end_time` 更新、无新增睡眠行、有 `app:history:notify` update
- [x] 3.2 验收：无未闭合睡眠时 end → `updated=false` + 降级瞬时 AddHistory + create 通知
- [x] 3.3 确认未新增测试文件、未新增 Redis 键/后台 ticker、未改 gateway-app 对外路由；本变更无新 DB 连接（无需改 `*_DB_LINK` / `.env.example`）
