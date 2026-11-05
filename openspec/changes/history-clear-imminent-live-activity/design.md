## Context

今日 Imminent 清场在 `voice.SyncPredictImminentPending`：每次 `PUT …/predict/imminent/pending` 成功先 `clearImminentLiveActivitiesForDevice`，再写 Redis / 重布延时。Timing 起停已挂 history 落库成功路径（`live_activity_timing.go`）。Imminent 日程权威仅 Redis（`PredictImminentPendingKey`），**无** MySQL pending 表；`push_live_activity_token` 存 LA token，end 后 activity 行易残留。

产品已定：清屏彻底离开 pending PUT；history Add/Update/Delete 成功一律清 Imminent（含瞬时 Add、任意 Update）；清屏时清空 Redis pending；**Imminent 与 Timing** 在 end 后均清理无效 DB activity token 行。

## Goals / Non-Goals

**Goals:**

- history 权威写成功后：绑定用户 Imminent LA 立刻 end（立即 dismiss 语义沿用既有组装层）。
- 同钩作废该宝宝 Redis imminent pending；旧 MQ 靠消费对账不匹配则不推。
- **所有** LA `action=end`（Imminent 清屏/换槽 **与** Timing 闭合/删除）后删除对应 `activityId` 的 `token_kind=activity` 行。
- 从 SyncPending **去掉**清屏调用。

**Non-Goals:**

- 不改 Timing start/update 触发条件与 ContentState 载荷；不因 Imminent 清屏去 end Timing。
- 不清空 / 不改写 active-eta Redis（仍由客户端 PUT 与 Timing 路径负责）。
- 不删 `push_to_start` token。
- 不新建 App 对外 API；不引入 history↔voice 业务包互引。
- 不在本变更改 Flutter。

## Decisions

1. **清屏编排落在 history-service 落库成功副作用（与 Timing 同层）。**
   - 在 `AddDeviceHistory` / `UpdateDeviceHistory` / `DeleteDeviceHistory`（及已汇入的闭合旁路如 EnsureEnd）成功后调用统一 `clearImminentOnHistoryMutation(deviceNo)`。
   - 一律清：不区分事件类型、是否瞬时、Update 字段集合。
   - *备选：voice internal HTTP 由 history 经 clients/voice 委托 — 无现成 clients/voice，多一跳；否决。*
   - *备选：仅 Pub/Sub 通知 App 本地 dismiss — 离线锁屏依赖 APNs end；否决。*

2. **history 直清 voice 域 Redis pending + 槽键（经 cachekit）。**
   - 已有先例：history 读写 `PredictTimingActivitySlotKey` / 读 active-eta。
   - 步骤：`RemoteListWxIDsByDeviceNo` → 每 wx `FanoutLiveActivity(end, imminent)` + `Del` slot → `Del PredictImminentPendingKey(deviceNo)`。
   - MUST NOT import `services/voice`；可把 voice 侧 `clearImminent…` 逻辑对等实现到 history（或抽到 `platform` 仅当无业务耦合——优先 history 本地函数，避免过早抽象）。
   - SyncPending **删除**清屏调用；PUT 空/非空列表行为不变（写/Del Redis + MQ），只是不再 end LA。

3. **「数据库相关数据」= `push_live_activity_token` 中已结束活动的 activity 行（Imminent **与** Timing），不是 pending 表。**
   - Imminent pending **无 MySQL**；避免误解写入 design。
   - 在 push `dispatchLiveActivity` 的 `action=end` 路径：**扇出尝试结束后**按 `activityId` **主动**删除该 id 下全部 `token_kind=activity` 行（发送失败仍删，避免脏行无限积压）；与 APNs invalid 删行兼容。
   - **不分** `attributesKind`：`imminent` 与 `timing` 共用同一清理；Timing 闭合/删除走既有 `endTimingLiveActivity` 即可命中，无需 history 再写一遍删库逻辑。
   - Timing Redis 槽键已在 `endTimingLiveActivity` 末尾 `Del`；本变更不重复造槽清理。
   - MUST NOT 删除同用户 `push_to_start`（任一 attributesKind）。

4. **失败语义：清屏/Del Redis 失败不回滚 history 事务。**
   - 与今日 Timing LA 一致：落库已成功则 best-effort 扇出 + 打 WARN；pending Del 失败打 WARN，下次 history 写或客户端 PUT 可再收敛。

5. **开火路径保留单槽换新 end。**
   - `startImminentLiveActivityForWx` 前 `endImminentSlotForWx` 保留；与全员清屏无关。

## Risks / Trade-offs

- [仅预测列表变、无 history 写时锁屏 Imminent 不自动清] → 接受；产品要以历史为准；用户忽略后本机其它变更若只改 pending 不改 history，条可残留直至下次 history 或开火换槽。
- [任意 Update（含改备注）都清 + 清空 pending] → 接受「一律清」简单性；客户端需尽快重报 pending，否则临近提醒空窗。
- [history 写 voice Redis] → 文档化键语义；禁止扩到其它 voice 键；评审盯 cachekit builder。
- [end 后删 activity token，客户端稍后才 register] → 正常；下次 start 用 push_to_start，再上报新 activity token。
- [多监护人绑定列表瞬时不一致] → 与今日清场相同；漏清的 wx 待下次写或换槽。

## Migration Plan

1. 先落地 push end 后删 activity token（可独立验证）。
2. history 钩子：清 LA + Del pending；去掉 SyncPending 清屏。
3. 联调：`PUSH_LIVE_ACTIVITY_ENABLED=true`；Add/Update/Delete 后 Imminent 立刻消失、Redis pending 空；Timing 闭合/删除后对应 activity token 行亦删；pending PUT 不再单独清屏；Imminent 清屏不得误 end Timing。
4. 回滚：恢复 SyncPending 清屏调用并去掉 history 钩子；token 主动删除可保留（有益无害）。

## Open Questions

- （已收敛）pending PUT 彻底搬走清屏；Update 一律清；瞬时 Add 也清；Redis pending 一并 Del；DB 指 LA activity token 行（Imminent+Timing）而非不存在的 pending 表。
- active-eta 是否随 history 清：本变更明确 **不清**；若产品要再开增量。
