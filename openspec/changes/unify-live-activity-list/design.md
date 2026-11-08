## Context

现网 Live Activity 双轨：

- Imminent：`attributesKind=imminent`，`activityId=imminent:{wxId}`，history CUD → `clearImminentOnHistoryMutation`（Del Redis pending + Fanout **end**）；MQ 到点 → 先 end 再 start。
- Timing：`attributesKind=timing`，`activityId=timing:{deviceNo}:{historyId}`，未闭合 Add → start，未闭合 Update → update，闭合/删除 → end；可多槽并存。

Token 表 `push_live_activity_token` 唯一键含 `attributes_kind` + `activity_id`；register 强制二者之一分槽。普通预测推送仍走 Redis pending + delayed MQ + `PushByBizType`，与 LA 纠缠在 Imminent 清屏/换槽上。

产品要求：锁屏只留**一条**活动，消息体为列表（预测 + 进行中），Flutter 依次展示；CUD / pending 同步 / 到点都用同一套 Refresh；普通推送与 MQ **不改调度**。

约束：history / voice 禁止互 import；跨域经 `clients/*`；Redis 经 cachekit；禁止默认新增 ticker 扫 pending；已有接口版本不可改结构（register 字段删除属 **BREAKING**，须与 Flutter 同发，path 可保持）。

## Goals / Non-Goals

**Goals:**

- history-service 提供唯一 `RefreshBabyLiveActivity(deviceNo)`（名可调整）：读 pending + 未闭合 history → 排序 → 按绑定 wx upsert/end。
- CUD：Del pending → Refresh（不再 Imminent end 清屏；不再单独 Timing 多槽扇出）。
- pending PUT 成功、MQ fire 闸门通过后：Refresh；fire 仍发普通可见推送。
- Token：每 wx/设备仅 `push_to_start` 与 `activity`；无 attributesKind/activityId 分槽。
- ContentState 列表契约稳定，供 Flutter 消费。

**Non-Goals:**

- 用 Redis TTL / keyspace / 扫表替代 delayed MQ。
- 修改 leadSeconds、对账、1/min、进行中闸、30min 历史闸等普通推送闸门语义。
- Android Live Update；改预测算法字段。
- 在本仓实现 Flutter Extension UI。

## Decisions

1. **Refresh 宿主在 history-service**  
   进行中数据本域权威；pending 经 cachekit 只读。voice 经 `clients/history` 调 internal Refresh。  
   *备选：voice 宿主 — CUD 反调 voice，边界更别扭。否决。*

2. **CUD 顺序：先 Del pending，再 Refresh**  
   保留「历史一变作废旧日程」，旧 MQ 对账失败 → 普通推送不受影响。Refresh 时预测段为空，列表可仅含进行中或空（空则 end）。  
   *备选：CUD 不删 pending — 会让过期预测留在 LA 且旧 MQ 仍可能匹配，否决。*

3. **列表组成与排序**  
   - 预测：当前 Redis pending **全量**（不按 `nextAt>now` 过滤），`nextAt` **升序**（越临近越上）。  
   - 进行中：`end_time=0` 行，`startTime` **倒序**，接在预测段之后。  
   每项带 `kind`（如 `imminent` / `timing`）及既有展示字段（eventId、title/eventName、nextAt/startTime、可选 expect*），MUST NOT 带 logo/color/远程图 URL。

4. **有 activity token → update，否则 start；空 items → end**  
   push 层按 `token_kind=activity` 是否存在选择 APNs event；end 后仍清该 wx 的 activity 行，保留 push_to_start。不再按 activityId 多槽查找。

5. **删掉 DB/API 的 attributesKind 与 activityId**  
   单活动后分槽列无意义；APNs 仍写**单一**固定 `attributes-type`（代码常量或单一 env，如 `PUSH_LA_ATTR_TYPE`）。Flutter 只保留一套 ActivityAttributes。

6. **MQ 保留且到点驱动 LA Refresh**  
   Redis pending 不是时钟器；delayed MQ 仍是准时叫醒。到点：普通推送 + Refresh（同拍更新列表态 LA），禁止先 end 再 start。

7. **pending PUT 也 Refresh**  
   否则预测只在到点或 CUD 后才进锁屏；PUT 写入后立即投影。

8. **与未归档双轨 change 的关系**  
   实现以本 change 为准；旧 Imminent/Timing 专用钩子删除或改道至 Refresh，避免双发。

## Risks / Trade-offs

- [CUD 后至 App 重报 PUT 前锁屏无预测项] → 接受；与「作废旧日程」一致，进行中仍可展示。  
- [Flutter / 旧客户端仍按双 Attributes 注册] → 特性开关 + 同发；旧 kind 注册可拒绝或忽略并打日志。  
- [update ContentState 整包替换丢字段] → Refresh 每次全量重拼 items，不靠增量 patch。  
- [voice→history Refresh 延迟/失败] → 失败仅 WARN，不回滚已写 Redis/已发普通推送；可观测日志。  
- [多监护人各一条 LA] → 仍按绑定 wx 列表各自 Refresh（与现 Imminent 多看护一致）。  
- [表结构迁移] → 启动/ensure 迁移去掉列或重建唯一键；部署期双写窗口与 Flutter 同发。

## Migration Plan

1. 落地 token 表/register 契约 + 单一 attributes-type（开关默认关或与现 `PUSH_LIVE_ACTIVITY_ENABLED` 共用）。  
2. 实现 history Refresh + internal/clients；接线 CUD（Del→Refresh）、pending PUT、MQ fire。  
3. 删除双轨 Timing/Imminent 专用扇出与清屏 end。  
4. 与 Flutter 列表态 Attributes 联调后开开关；回滚：关开关并保留普通推送。

## Open Questions

- 单一 APNs attributes-type 最终字符串名（与 Flutter `ActivityAttributes` 对齐后写入 design/tasks）。  
- register path 是否保持免 Bearer（既有 `live-activity-register-no-auth-go`）；本变更默认**保持**现网鉴权结论，仅改 body 字段。  
- active-eta 是否继续单独 Timing update：本变更默认 **并入 Refresh 列表项字段**（若进行中项需要 expectEndAt，Refresh 可读 active-eta 键合并）；不再按 historyId 发独立 Timing update。
