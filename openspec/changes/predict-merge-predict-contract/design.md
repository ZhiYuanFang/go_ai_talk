## Context

- 字典侧 `mergePredict` 已由 `event-merge-predict` 落地；本 change 只改预测契约与 history 批量读。
- Flutter 新客户端：非合并子 pending 带可选根名 `parentEventName`；合并根 active-eta 可无结束；不再依赖 `expectLeafTitle`；需批量 last。
- Python 意图：经 voice internal GET 读 pending，用 `parentEventName` 在用户说根事件名时反推命中预测项。
- 旧客户端仍发旧形（无父名、可能带 leafTitle、active 正结束）。

## Goals / Non-Goals

**Goals:**

- pending/active 契约向前兼容（加字段 / 放宽校验）。
- `expectLeafTitle` 全链路丢弃；alert 不再拼 leaf。
- `parentEventName` 存储并经 App/internal GET 带回（仅 imminent pending）。
- 批量最近发生：纯 history、`start_time`、含进行中；不计入 usage。

**Non-Goals:**

- Go 内按 `mergePredict` 重算预测。
- active-eta 挂 `parentEventName`。
- 合并预约表；改 appointment API。
- 实现 Flutter/Python 算法（仅契约对齐说明）。

## Decisions

1. **`parentEventName`（替代原草案 parentEventId）**  
   - optional string；trim 后空=无父。  
   - Flutter 应写字典根事件权威 `name`（非叶子名、非别名）。  
   - Go 只原样存 Redis 并投影 GET；不做名→id 解析、不参与调度/去重（去重仍按 `eventId`）。  
   - 供 Python 自然语言反推；改名后靠下次 PUT 刷新。

2. **`expectLeafTitle`**  
   - API/DTO **不绑定**该字段；旧 JSON 键静默丢弃。  
   - 不写 Redis、不进 fire payload；`buildPredictImminentExpectSuffix` amount 分支只拼数量+单位。  
   - 有意削弱旧客户端推送文案中的叶子后缀。

3. **`expectEndAt`**  
   - 省略 / null / **0** = 无预计结束，项仍可保留（root/leaf/start 合法即可）。  
   - `>0` 时 MUST `>= startAt`。  
   - LA：无结束则省略 ContentState 键，禁止写 `0` 哨兵。

4. **批量 latest-by-event**  
   - 路径：`GET /device/history/api/event/latest-by-event?deviceNo=`（与既有 `/device/history/api/event/latest` 同族）。  
   - 实现在 **history-service**；鉴权同既有 history App（登录+绑机）；Bearer **不**豁免。  
   - usage：负责人确认 **不计入** → `maintenance_skip.go` 精确 path。  
   - 语义：纯 `history` 表；每 `event_id` 取 `start_time` 最大的一条（含 `end_time=0`）；返回至少 `eventId`、`occurredAt`（=start_time）；MAY 含 `remark`、`historyId`。  
   - **不**读 `appointment_next`（预约是下次约定，不是最近发生）。

5. **发版**  
   - 先发 Go，再 Flutter/Python；三仓字段名锁定 `parentEventName`。

## Risks / Trade-offs

- [旧推送少 leaf] → 产品已接受；文档标明有意变更。  
- [根改名短暂不一致] → 下次 pending PUT 纠正。  
- [混版同宝宝整表覆盖] → 旧 PUT 无 parentName / 强制结束；属发版策略，非协议硬破坏。  
- [LA 与 expect-lock 并行] → 无结束省略副行，勿写 0。

## Migration Plan

1. 部署 voice-service（pending/active 契约）+ history-service（latest-by-event）+ gateway-app（skip 列表）。  
2. Flutter/Python 发版对齐 `parentEventName`。  
3. 回滚：勿再强制正 `expectEndAt`（会伤新客户端）；leaf 拼接可不恢复。

## Open Questions

- 无（`parentEventName` 仅 pending；active-eta 不加；latest 纯 history + start_time + 含进行中；usage 不计入）。
