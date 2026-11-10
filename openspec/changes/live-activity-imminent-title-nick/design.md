## Context

- 统一 LA 投影在 `history`：`RefreshBabyLiveActivity` → `buildLiveActivityItems`；imminent 段今天 `item["title"] = pending.Title`（客户端事件名）。
- 普通可见推送在 voice MQ fire：`nick + "要" + title + "了"`，可选再拼 expect 后缀；与 LA 触发面不同（LA 另有 pending PUT Refresh）。
- Flutter 锁屏 Imminent：左侧原样展示推送 `title`，右侧用 `expect*` 拼短版「预计…」；Island 用本地目录短名，不吃推送 title。
- history 已有 `GetBirthday`（经 device 画像契约）可取 `babyName`，无需 import voice。

## Goals / Non-Goals

**Goals:**

- LA imminent `title` 投影为 `{nick}要{事件}了`；nick 空/失败回落「宝宝」。
- 拼装仅发生在 history LA 投影；Redis pending 仍存裸事件名。
- expect 继续独立字段，不进 title / 不进 LA alert 成品句后缀。

**Non-Goals:**

- 不改 voice 普通推送文案算法（已同形基句，可保留各自实现）。
- 不改 timing `eventName`、不改 Flutter UI、不改 register/clear/ack。
- 不把 `expectLeafTitle` 或 amount/duration 中文拼回 title。

## Decisions

1. **拼装落点：`buildLiveActivityItems`（history）**  
   选中进窗 imminent 项后，用 `pending.Title`（trim）作事件段；昵称一次 `GetBirthday(deviceNo)`（或等价 profile）取 `babyName`。  
   *否决：在 voice fire 拼好再写 Redis* — pending PUT 也会 Refresh，无 fire 时仍要成品句；且禁止 LA 依赖普通推送路径。  
   *否决：Flutter 本地拼* — 锁屏契约已约定吃推送 `title`；Push-to-Start 无本地画像时更脆。

2. **句式固定：`{nick}要{事件}了`**  
   与普通推送基句对齐（产品定稿）；不用「需要」。事件段为空则该项可不写 `title` 键（或整项仍带 eventId/nextAt，由客户端「事件即将发生」兜底——实现取「事件空则省略 title 键」，与现状一致）。

3. **expect 仍走 `applyPendingExpect`**  
   与 Flutter 右槽契约不变；title MUST NOT 追加「，预计这次约…」。

4. **alertBody**  
   现有 `liveActivityAlertBody` 读首条 imminent `title`；投影后自然变为成品句，有意保留，无需旁路。

5. **DB**  
   无新表/新连接；画像只读既有 device 契约。

## Risks / Trade-offs

- [昵称变更后锁屏滞后] → 下次 pending PUT / MQ Refresh 会重拼；可接受。  
- [画像失败] → 回落「宝宝」，打 WARN，不阻断扇出。  
- [长 title 与 Island] → Island 不吃推送 title，无 Island 撑宽风险。  
- [与普通推送文案漂移] → 句式刻意同形；后缀策略故意不同（LA 无 expect 后缀）。

## Migration Plan

1. 部署 **history-service**（投影逻辑）。  
2. 无需 Flutter 同发；无需清 Redis。  
3. 回滚：恢复透传裸 `pending.Title` 即可。

## Open Questions

- 无（句式已定为 `{nick}要{事件}了`；拼装落 history LA）。
