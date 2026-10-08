## Why

预测临近推送的「提前叫醒」已由 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS` 配置，但「同一事件已推送则不再推」的 Redis 去重 TTL 仍硬编码 5 分钟。运维将 lead 调到 600（10 分钟）后，去重窗仍停在 5 分钟，窗口语义不一致，易在 lead 窗内出现重复可见推送。

## What Changes

- 消费路径推送去重键的 TTL：**改为等于**当前生效的 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS`（非法/未配时回退与 lead 相同的默认秒数；`LEAD=0` 时 dedup TTL MUST 至少 1 秒）。
- **不改**：提前叫醒算法（已读该 env）；近 **30 分钟** history 有同类操作则不推；**进行中**（`end_time=0`）闸另算。
- **不改** API / MQ 拓扑；本变更不强制改各环境 env 值（由运维后续设为 600）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `predict-imminent-schedule`：将「五分钟去重」改为「去重窗口等于 leadSeconds」；保留「三十分钟历史闸」与进行中闸既有语义。

## Impact

- 代码：`internal/services/voice/predict_imminent.go`（去掉固定 `predictImminentPushDedupTTL`，SetNXEX 用 lead 秒数）；`internal/platform/cachekit/keys_voice.go` 注释同步。
- 规格：`predict-imminent-schedule` 去重相关 Requirement / Scenario 文案。
- 运维：改 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS` 后需重启 voice 使 lead 与 dedup 同时生效；已存在的 dedup 键仍按写入时 TTL 过期。
