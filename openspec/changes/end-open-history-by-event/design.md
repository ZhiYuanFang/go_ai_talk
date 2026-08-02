## Context

Python 意图已正确返回 `feeding` + `end` + `event_id`（如睡眠）。voice 经 `applyVoiceEventEndHistory` 先调 `DeviceHistory().EndLatestHistoryIfMatch`，未命中再降级瞬时 `AddHistory`。

history 侧 `EndLatestDeviceHistoryIfMatch` 当前实现：

1. `GetLatestDeviceHistory(deviceNo)` —— 全局按 id 最新一条（不限 event、不限是否未闭合）
2. 若 `last.EventId != eventID` → `updated=false`
3. 否则更新该行 `end_time`（**未**校验 `end_time==0`）

因此：开始睡眠后若中间写入其它事件，「孩子醒了」会降级新建瞬时睡眠，原未闭合睡眠残留。基线 `history-piece-and-realtime-notify` 场景标题已写「同 eventId」，但实现落成「全局最新碰巧相等」。

约束：跨服务仍走既有 `end-latest` HTTP；最少 DB 操作；不新增 Redis 读缓存键；不新增测试文件；中文注释。

## Goals / Non-Goals

**Goals:**

- `EndLatestHistoryIfMatch` 权威语义：闭合该 device 下指定 `eventId` 的最近一条未闭合（`end_time=0`）记录
- App `end-latest` 与语音 `applyVoiceEventEndHistory` 共用同一语义（中间夹其它事件仍可结束睡眠）
- 无未闭合同 event 时 `updated=false`，voice 继续既有降级新建
- 保持 notify / piece 缓存失效行为与现网 update 路径一致

**Non-Goals:**

- 不改 Python 字段或意图模型
- 不新增/改版对外 App HTTP 路径或请求字段（内部契约路径不变，仅语义）
- 不重写 start/one/multi 落库
- 不引入按时间窗扫全表的 reconciler / ticker
- 不新建 Redis 读缓存；latest 列表缓存仍走既有 patch

## Decisions

### 1. 在 history 权威路径改匹配条件（方案 A），不另开 voice 旁路 API

```
EndLatestDeviceHistoryIfMatch(deviceNo, eventID, endTime, remark):
  查 history WHERE device_no=? AND event_id=? AND end_time=0
       ORDER BY id DESC LIMIT 1
  命中 → UPDATE end_time (+ optional remark)
       → patch cache + bumpPiece + publish update
       → return true
  未命中 → return false（不写库）
```

- **理由**：单一真源；App 与语音一致；voice 降级逻辑可不动。
- **替代（否决）**：voice 侧自查未闭合再 UpdateHistory —— 双语义、易与 App 分叉，且 voice 不应加深 history 查询形态。

### 2. 「未闭合」定义：`end_time = 0`

- 与 start 落库（不写 EndTime / 默认 0）及现网「未结束」展示一致。
- **替代**：`end_time IS NULL` —— 当前列语义为 Unix 秒，0 已表示未结束，不改为 NULL。

### 3. 已结束同 event 不再被本接口覆盖

- 全局最新若是「已结束的睡眠」，且更早还有未闭合睡眠：按 `event_id + end_time=0` 仍能命中未闭合行（修复旧实现「改写已结束最新行 / 或因最新已结束但 event 匹配而误更新」）。
- 若同 event **没有任何** `end_time=0`：`updated=false` → voice 降级新建瞬时结束（保留「只说了醒了、从未开始」的可用路径）。

### 4. 一次查询 + 一次更新（最少 DB）

- 用条件查询代替「先 GetLatest 再比 eventId」；命中后再按主键 Update。
- 不引入新索引为强制项；若线上 `device_no + event_id + end_time` 慢查询再单独立项（本变更不扩表）。

### 5. voice 降级与旁路闭合逻辑保持

- `applyVoiceEventEndHistory` 仍：EndLatest → 失败则 AddHistory；若降级前缓存的 last 为「其它事件且未结束」则尝试闭合该其它事件（既有行为）。
- 本变更后，常见「睡眠未结束但中间有尿布」路径应在第一步 `updated=true` 结束，不再误新建。

### 6. 契约形状不变

- 路径、`DeviceHistoryEndLatestReq/Res`、`contracts.DeviceHistory.EndLatestHistoryIfMatch` 签名不变。
- 注释与 OpenSpec 场景 MUST 写明新匹配语义（**BREAKING** 行为，字段兼容）。

## Risks / Trade-offs

- [App 曾依赖「只能结束全局最新」] → 语义变宽：可结束非最新的未闭合同 event；需在发布说明中点名；若产品要坚持「仅全局最新」则本方案不适用（已否决，产品选 A）
- [多条未闭合同 event（历史脏数据）] → 只闭合 id 最大的一条；其余仍挂起，可后续运维清理，本变更不批量扫合
- [并发双 end] → 两条请求可能都查到同一行再 update，终态仍闭合；或一条 updated=false 后降级新建 —— 可接受；不引入分布式锁
- [缓存 latest 与 DB 不一致] → 写路径仍 patch + bumpPiece；查询未闭合行以 DB 为准，避免只信 latest 缓存误判

## Migration Plan

1. 部署 history-service（含 local 实现变更）；voice 无需改调用，建议同批或随后部署以便联调日志一致。
2. 回归：开始睡眠 → 中间记其它事件 → 「孩子醒了」→ 原睡眠行 end_time 更新、无多余瞬时睡眠行；无未闭合睡眠时 end → 新建瞬时行；App end-latest 同语义。
3. 回滚：回退 history-service 二进制即可恢复旧「全局最新」匹配。

## Open Questions

- 无（产品已确认方案 A：按 eventId 找未结束记录，中间夹其它事件仍应结束）。
