## Context

预测临近消费闸顺序：Redis 对账 → 进行中（`end_time=0`）→ 去重 SetNX → 近 30 分钟 history → 推送。  
提前叫醒 `fireAt = nextAt - leadSeconds` 已读 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS`；去重 TTL 仍为常量 `5 * time.Minute`。产品要将 lead 调到 600s，并要求去重窗与之对齐；历史 30 分钟与进行中保持不变。

## Goals / Non-Goals

**Goals:**

- 去重键 TTL = 当前 `predictImminentLeadSeconds()`（与同步路径同一读取函数）。
- `LEAD=0` 时 dedup 仍有效（最小 1s）。
- 规格与 `keys_voice` 注释不再写死「五分钟去重」。

**Non-Goals:**

- 改 history 窗口（保持 1800s）。
- 改进行中闸。
- 改各环境 `.env` 中的 LEAD 数值（运维另行改为 600）。
- 改提前叫醒公式或远 delay 跳过逻辑。

## Decisions

### D1：复用 `predictImminentLeadSeconds()`，删除固定 dedup 常量

- **选择**：`ttl := time.Duration(sec) * time.Second`，其中 `sec = max(1, predictImminentLeadSeconds())`；去掉 `predictImminentPushDedupTTL`。
- **备选**：独立 env `VOICE_PREDICT_IMMINENT_DEDUP_SECONDS`——多一个旋钮，与「跟 lead」目标不符。
- **理由**：单一配置源，运维只改 LEAD 即可。

### D2：LEAD=0 时 dedup 至少 1 秒

- **选择**：`if sec < 1 { sec = 1 }` 仅用于 dedup TTL；lead 本身仍可为 0（立即延时投递）。
- **理由**：Redis `EX 0` / 零 TTL SetNX 行为不稳定；避免「立刻可再推」打穿去重。

### D3：不改已写入 Redis 的旧 dedup 键

- **选择**：部署后新占位用新 TTL；旧键按原 TTL 自然过期。
- **理由**：无需扫键；短暂窗口不一致可接受。

## Risks / Trade-offs

- **[LEAD 调很大]** → 去重窗变长，合法二次提醒更晚。缓解：与提前窗一致，属预期。
- **[LEAD 调很小但 >0]** → 去重短，重投/多实例窗口变窄。缓解：进行中与 30min history 仍挡误推。
- **[与 skip-far-delay 并存]** → 无冲突；far-delay 管发布，本变更管消费 dedup。

## Migration Plan

1. 发布 voice-service。
2. （可选）运维将 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS=600` 并重启。
3. 回滚：恢复固定 5min 常量即可。

## Open Questions

（无）
