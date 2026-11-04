## Context

现网 Imminent：`SyncPredictImminentPending` 写 Redis + MQ 延时 → `handlePredictImminentFire`（pending 匹配、进行中闸、event 去重、30 分钟历史闸）→ `PushByBizType(predict_imminent)` 普通可见推送。active-eta 只写 Redis。history 变更经 Redis Pub/Sub 通知 App WS，无 Live Activity。

兄弟仓 Flutter 将只等推送、两类 Attributes、点击不 dismiss Imminent；本仓负责权威起停。

现网 `POST /app/api/push/register` 与表 `push_device` 仅存 **Device APNs / 厂商 device token**（`uk_token` 全局唯一）。Apple Live Activity 另有 **Push-to-Start token** 与 **per-activity update token**，与 device token 互不可替，禁止共用该唯一键语义。

## Goals / Non-Goals

**Goals:**

- 用户级 Imminent 1/min 先到占坑。
- pending PUT 清全部 Imminent LA；开火单槽先 end 后 start。
- iOS Imminent/Timing 走 liveactivity；Android 可见通知可保留。
- Timing 挂 history 开/闭；多槽；无 10s 确认延迟。
- activity id 可配对 end；LA token 可查（新 path + 分表）。
- 多监护人：同一宝宝绑定的每个 wx 各自一条 Imminent 槽。

**Non-Goals:**

- Flutter Extension / ActivityKit UI / Info.plist（兄弟仓）。
- Android 锁屏 Live Update。
- 改间隔预测算法或 pending 事件字段语义。
- 用 active-eta PUT 作为 Timing 唯一触发。
- 为 Live Activity 在 Apple Developer 另申请新推送证书（复用现有 APNs `.p8`）。

## Decisions

1. **清场挂 `SyncPredictImminentPending`。** 每次成功处理 pending 全量同步先 end 全部 Imminent，再写日程。客户端约定「有变化才发 PUT」，故服务端按次清 = 按变更清；不必再做内容哈希闸。备选（仅开火换新）无法覆盖「用户忽略后本机其它变更」。

2. **1/min 用 cachekit SETNX，键按 wxId（用户），TTL 60s。** 放在真扇出之前；未过闸不 start。Timing 不写此键。

3. **Imminent 槽维度：按 wx（用户）单槽。** 清场按宝宝绑定用户列表扇出；每个绑定 wx 各自维护一条 Imminent activity id（如 Redis `imminent:{wxId}` 或含 deviceNo 的复合键但意图为 per-wx）。多监护人同宝宝时各亮一条。Timing：`timing:{deviceNo}:{historyId}`（或多监护人各自一份，实现与绑定列表扇出对齐）。

4. **Timing 触发：单一落库成功副作用。** 扇出 MUST 挂在 history-service **持久化成功**路径一次触发（典型：local 落库成功后）。所有写入口（App HTTP、voice 经 `clients/history` remote 等）MUST 汇入该路径，使间接触发一次即可。MUST NOT 在 voice remote adapter 与 history local 各推一次导致双发。不依赖 active-eta。

5. **iOS 渠道分支。** 有可用 LA / Push-to-Start token 时发 liveactivity；无 token 跳过该接收方；其它通道（Android 等）MAY 保留可见 `predict_imminent`。

6. **LA Token：新 App path + 落库分表。** MUST NOT 扩展把 LA token 写入 `push_device` 同一 `token`/`uk_token` 语义。已定 App path：`POST /app/api/push/live-activity/register`（`deviceKey`、`tokenKind`=`push_to_start`|`activity`、`attributesKind`=`imminent`|`timing`、`token`；`activity` 时 `activityId`）。权威表 `push_live_activity_token`。活动态跟踪键经 cachekit builder。`/app/api/push/register` 继续只管 device alert token。新 path MUST 走 gateway-app 反代 / Bearer / usage 检查表。

7. **APNs / Apple 配置。** 复用现有 `PUSH_APNS_*` Auth Key（`.p8`）；一般无需在 Developer 后台另开 Live Activity 专用证书。发送时 `apns-push-type: liveactivity`，`apns-topic: {bundleId}.push-type.liveactivity`（后缀由发送代码拼接，**勿**写入 env Bundle ID）。App 侧 `NSSupportsLiveActivities`、Widget Extension、描述文件归 Flutter change。

8. **与 Flutter 协作。** 载荷字段、深链、两类 Attributes 名与 `flutter_ai_talk` change `predict-ios-live-activity` 对齐；本仓 tasks 不含客户端勾选。

## Risks / Trade-offs

- [pending 频繁导致 LA 刚亮被清] → 接受「有变更即 PUT、服务端按次清」。
- [1/min 丢后条] → 产品接受先到占坑。
- [多监护人各一条] → 锁屏条目随绑定人数增加；产品已接受。
- [仅 Go 上线 / 仅 App 上线] → 特性开关；无效推送可忽略。
- [history 钩子漏路径或双发] → tasks 审计：写入口汇入单次落库触发，禁止双钩。
- [包边界] → Timing 扇出经 clients/push（或内部契约），禁止 services 互引。
- [两类 Attributes 可能两套 Push-to-Start] → 分表须能按 attributes/kind 存；与 Flutter 对齐。

## Migration Plan

1. 新 path + 分表落地 token；APNs liveactivity 发送（开关默认关）；复用 `.p8`。
2. pending 清场 + 1/min + 开火 LA（per-wx 单槽）。
3. Timing：history 单次落库成功钩子。
4. 与 Flutter 联调后开开关；回滚关开关恢复 iOS alert。

## Open Questions

- （已收敛）LA register：`POST /app/api/push/live-activity/register`；按 `attributesKind` 分列存 Push-to-Start；表 `push_live_activity_token`。
- usage 是否计入：实现时已按负责人结论处理则关闭本条；否则归档前再确认。
- 服务端 end 对 per-activity update token 的强制程度：与 Flutter 联调（对端客户端点开不 dismiss，依赖 end 推送）。
