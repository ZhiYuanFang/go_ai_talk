## Why

预测临近推送（`predict_imminent`）在叫醒时仅用「近 30 分钟 `start_time` 历史闸」抑制提醒。长时段进行中事件（如睡眠，`end_time=0` 且开始已超过 30 分钟）不会命中该闸，用户仍可能收到「宝宝要睡觉了」类误扰推送。产品要求：同类事件若已在进行中，则不再推送。

## What Changes

- 在预测临近叫醒消费路径增加**进行中闸**：若该宝宝在同类事件（根则含后代叶子）上存在未闭合历史（`end_time=0`），MUST 跳过可见推送并 Ack。
- 进行中闸 MUST 在五分钟去重占位**之前**执行，命中时 MUST NOT 写入去重键（避免结束后仍被 5 分钟窗误伤）。
- 保留既有 Redis 匹配、五分钟去重、三十分钟 `start_time` 历史闸与推送文案逻辑；**无 BREAKING** 对外 App API 路径与载荷。
- 经 history 服务契约查询未闭合记录（voice MUST NOT 直查 history 库）；若现有 `ListHistoryFilter` 无法按 `end_time=0` 精确命中，则扩展契约/内部查询以最少 DB 往返完成存在性判断。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `predict-imminent-schedule`：推送前闸增加「进行中（`end_time=0`）」抑制；明确与五分钟去重的先后顺序及根事件叶子展开规则。

## Impact

- **voice-service**：`internal/services/voice/predict_imminent.go`（`handlePredictImminentFire` 闸顺序与新查询）
- **history 契约**（按需）：`contracts.DeviceHistoryContract`、history-service HTTP、`internal/clients/history`、history local/adapter — 支持「多 eventId 是否存在 end_time=0」的存在性查询
- **规格**：`predict-imminent-schedule` 增量
- **不涉及**：客户端 pending 同步协议变更、push-service 文案、AASA/微信域名、Redis 读缓存新建（沿用既有 history 契约；去重键语义不变仅调整写入时机相对新闸）
- **边界**：无新 App 对外路由；无背景 ticker；跨域仅 HTTP 契约
