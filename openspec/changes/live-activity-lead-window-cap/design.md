## Context

`unify-live-activity-list` 已落地：`RefreshBabyLiveActivity` 读 Redis pending 全量 + 未闭合 history，拼 `items` 后扇出；token 分槽字段在 API 侧已废，但表列 `attributes_kind` / `activity_id` 仍保留并恒写空串。

产品收敛：锁屏只关心「现在正在发生的」与「已进普通推送 lead 窗、该提醒的预测」，总条数最多 3；远点预测继续由 pending + delayed MQ 调度，不进 LA。

约束：history 不得 import `services/voice`；跨域经 clients；Redis 经 cachekit；不新增扫 pending 的 ticker（进窗后的 Refresh 仍靠既有 PUT / CUD / MQ fire 触发）。

## Goals / Non-Goals

**Goals:**

- 预测段：仅 `nextAt - leadSeconds ≤ now` 的 pending 项（lead 与 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS` 同语义）。
- `items` 总长 ≤ 3；截断 **进行中优先**：先纳入 open timing（既有 `startTime` 倒序），剩余名额再纳入进窗预测（`nextAt` 升序）。
- 入选后的展示顺序仍为：预测段在前、进行中段在后（与现 Flutter 契约一致）。
- DROP（或等价迁移去掉）token 表 `attributes_kind`、`activity_id` 及代码空写。

**Non-Goals:**

- 修改 leadSeconds 默认值、普通推送对账/1min/进行中闸/30min 历史闸。
- 改变 pending PUT 全量写入或 MQ 投递集合。
- 为「远点变进窗」单独加周期 Refresh；依赖 fireAt 时刻的既有 MQ→Refresh。
- 改 APNs `attributes-type` 常量名或去掉 ContentState item `kind`。
- Flutter Extension UI。

## Decisions

1. **进窗谓词对齐 leadSeconds**  
   `includePredict ⇔ nextAt > 0 && nextAt - leadSeconds ≤ now`（已过 `nextAt` 的 pending 仍纳入，直到 pending 被替换/清空）。  
   *备选：仅 nextAt>now — 会丢掉已到点未清的项，锁屏空白，否决。*  
   *备选：按 dedup 键「已推过」过滤 — 与「进窗应展示」不一致，否决。*

2. **lead 读取放 platform 或 history 本地读同一 env**  
   history 与 voice 共用 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS`（非法/空 → 300）。优先小函数放 `internal/platform/...`（或 history 包内复制读 env，注释对齐 voice），**禁止** import voice。  
   *备选：history HTTP 反查 voice — 过重，否决。*

3. **截断算法（进行中优先 A）**  
   ```
   timings = open histories, startTime desc
   predicts = pending filtered by lead window, nextAt asc
   takeT = min(len(timings), 3)
   takeP = min(len(predicts), 3 - takeT)
   items = predicts[:takeP] + timings[:takeT]   // 展示顺序仍预测在前
   ```  
   若进行中 ≥ 3，预测段为空（仍可有 LA，仅展示 timing）。  
   *备选：预测优先 — 产品已选 A，否决。*

4. **常量 MaxItems=3**  
   代码常量即可；暂不配 env（产品硬顶）。若日后要可配再开变更。

5. **Token 列物理删除**  
   ensure/schema：新库不建这两列；已有库 ALTER DROP（或重建唯一键后 DROP）。实体与 upsert Data map 不再含这两字段。唯一键保持 `(wx_id, device_key, token_kind)`。  
   *备选：继续空列兼容 — 产品明确可删，否决长期保留。*

6. **触发点不变**  
   CUD Del→Refresh、pending PUT Refresh、MQ fire Refresh 足够：远点在 fireAt 进窗时由 fire 路径 Refresh 带上；无需新后台循环。

## Risks / Trade-offs

- [进行中 ≥3 时锁屏看不到任何预测] → 接受；产品选进行中优先。  
- [history 与 voice lead 读法漂移] → 同 env 名 + 同默认 300；注释交叉引用；可选抽 platform 单点。  
- [DROP 列与旧二进制并存] → 先发读忽略列的版本或同发；回滚需保留迁移脚本注意。  
- [远点进窗仅靠 MQ fire Refresh] → 与现「不扫表」一致；若 MQ 失败则普通推送也失败，LA 短暂落后可接受（下次 PUT/CUD 会纠）。  
- [Flutter 仍假设长列表] → 契约字段不变，仅变短；需告知兄弟仓。

## Migration Plan

1. 改 `buildLiveActivityItems`（过滤 + 截断）并验证 lead env。  
2. schema ensure DROP 列；部署 push-service 与 history-service。  
3. 回滚：关 `PUSH_LIVE_ACTIVITY_ENABLED` 或回退二进制；列 DROP 后回滚需迁移加回空列（若仍要兼容旧码）。

## Open Questions

- （无阻塞）Attributes map 中的 `activityId: baby:{wx}` 是否一并删：本变更默认 **保留**（非 token 分槽列）；若 Flutter 确认无用不另开 change。
