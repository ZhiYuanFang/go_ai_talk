## Why

语音收到 Python `target_type=feeding` + `action=end`（如「孩子醒了」→ 睡眠 `event_id=3`）时，现网 `EndLatestHistoryIfMatch` 只判断「全局最近一条 history 的 eventId 是否等于目标」，中间若夹了换尿布等其它记录，会误判未匹配并降级 `AddHistory` 新建瞬时结束行，而真正未闭合的睡眠仍挂着。需要把结束语义改为「按 eventId 闭合最近一条未结束记录」，App `end-latest` 与语音共用同一权威行为。

## What Changes

- **BREAKING（契约语义）**：`EndLatestHistoryIfMatch` / `POST /device/history/api/event/end-latest` 匹配条件从「全局最新一条 eventId 相等」改为「该 deviceNo 下 `eventId` 匹配且 `end_time=0`（未闭合）的最近一条（按 id DESC）」。
- 命中未闭合记录时：更新其 `end_time`（及可选 remark），返回 `updated=true`，并走既有 cache patch + `app:history:notify` update 推送。
- 不存在该 eventId 的未闭合记录时：返回 `updated=false`；voice 侧既有降级（瞬时 `AddHistory`）保持不变。
- 已结束（`end_time!=0`）的同 eventId 记录不得被本接口再次「结束」覆盖；不得因全局最新是其它事件而漏闭合更早的未结束同 event。
- 不新增 HTTP 路径、不改请求/响应字段形状；不引入新 Redis 键；不新增后台循环任务。

## Capabilities

### New Capabilities

- `end-open-history-by-event`：结束计时类 history 时按 eventId 定位并闭合最近一条未结束记录；无未结束记录才允许 voice 降级新建。

### Modified Capabilities

- `history-piece-and-realtime-notify`：澄清/修正「语音结束事件 / EndLatest」场景中「最近一条未闭合」的匹配范围为「同 eventId 的最近未闭合」，而非「全局最近一条且 eventId 碰巧相等」（见 v2.0.24 基线同名章节）。

## Impact

- **代码**：`internal/services/history/local.go`（`EndLatestDeviceHistoryIfMatch` 查询条件）；voice `applyVoiceEventEndHistory` 调用契约不变，行为随 history 权威语义自动变正确；remote/switch adapter 透传不变。
- **API**：内部契约 `POST /device/history/api/event/end-latest` 语义变更（路径与字段不变）；App 与语音共享。
- **依赖**：history 库 `history` 表；既有 list/latest 缓存 patch 路径；无新 DB 连接、无 gateway-app 新路由。
- **基线**：对照 `openspec/specs/v2.0.24/spec.md` 中 `history-piece-and-realtime-notify`；归档合并时须同步场景表述。
- **非范围**：Python intent schema、事件字典、start/one 落库、澄清/quota、gateway-app 对外接口。
