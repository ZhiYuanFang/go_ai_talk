# prediction

## Purpose

预测能力、临近推送、Live Activity、predict-pro 与预测计数权益的行为规格。 Server-side prediction schedules, unlocks, and related push hooks.

<!-- migrated-from: event-merge-predict, predict-active-timing-eta, predict-capability-unlock, predict-imminent-capability-gate, predict-imminent-internal-get, predict-imminent-live-activity-push, predict-imminent-push, predict-imminent-push-copy, predict-imminent-schedule, predict-latest-occurrence-batch, predict-pending-parent-optional, predict-pro-feeding-eligibility, predict-pro-idea-api, predict-pro-idea-data-model, predict-pro-unlock-shell, predict-push-ungate, predict-timing-live-activity-push, prediction-allowed-count, prediction-count-grant, prediction-effective-count, remove-prediction-slot-unlock, unified-live-activity-push -->

## Requirements

### Requirement: 事件字典 MUST 支持合并预测标志且默认否

系统 MUST 在全局事件字典（`event`）持久化合并预测标志：数据库列名为 `merge_predict`，对外 JSON 字段名为 `mergePredict`。新建事件时若未显式指定，该标志 MUST 为 false（0）。该标志 MUST NOT 复用或改写既有 `eventType`、`isAppointment` 或 `parent_id` 语义。

#### Scenario: 新建事件默认不合并预测

- **WHEN** 管理员新增事件且未勾选合并预测
- **THEN** 持久化记录的 `merge_predict` MUST 为 0，且后续列表/options 返回中该事件 `mergePredict` 为 false/0

#### Scenario: 与预约标志正交

- **WHEN** 某根事件同时存在 `isAppointment` 与 `mergePredict` 取值
- **THEN** 系统 MUST 独立持久化与暴露两字段，MUST NOT 因其一为真而改写另一字段

### Requirement: 仅有子节点的根事件 MAY 将合并预测置为真

系统 MUST 仅在事件同时满足以下条件时允许将 `merge_predict` 持久化为 1：（1）`parent_id` 为 0（根事件）；（2）存在至少一条 `parent_id` 等于该事件 id 的子行。不满足条件时，写路径 MUST 将标志写为 0（即使请求体传 1）。

#### Scenario: 有子的根可开启合并预测

- **WHEN** 管理员对 `parent_id=0` 且已有直接子节点的事件更新 `mergePredict=1` 且保存成功
- **THEN** 该行 `merge_predict` MUST 为 1，且后续 options/list 返回 `mergePredict` 为真

#### Scenario: 无子的根不可开启

- **WHEN** 管理员对 `parent_id=0` 但无任何子行的事件请求 `mergePredict=1`
- **THEN** 持久化结果 MUST 为 `merge_predict=0`

#### Scenario: 叶子事件不可开启

- **WHEN** 管理员对 `parent_id≠0` 的事件请求 `mergePredict=1`
- **THEN** 持久化结果 MUST 为 `merge_predict=0`

#### Scenario: 根改挂为非根时清零

- **WHEN** 某根事件原 `merge_predict=1` 且更新将其 `parent_id` 改为非 0
- **THEN** 更新后该行 `merge_predict` MUST 为 0

### Requirement: 删除导致父变为无子根时 MUST 清零父的合并预测

当删除事件成功后，若被删事件的原父节点变为「根且无子」，系统 MUST 将该父节点的 `merge_predict` 置为 0，并触发与既有事件变更一致的 EventOptions 缓存重建。

#### Scenario: 删光最后子节点清零父标志

- **WHEN** 根事件 R 仅有一子 C 且 R 的 `merge_predict=1`，管理员成功删除 C
- **THEN** R 的 `merge_predict` MUST 变为 0

### Requirement: 后台 MUST 可按规则编辑合并预测标志

设备域 Hub 事件新增/编辑表单 MUST 提供合并预测勾选；仅当当前编辑对象为有子节点的根时勾选控件可用，否则 MUST 禁用。更新成功后，系统 MUST 使事件字典读模型与既有「变更后重建 EventOptions 缓存」约定一致。

#### Scenario: Hub 对无子根禁用勾选

- **WHEN** 运维打开无子根事件的编辑表单
- **THEN** 合并预测勾选控件 MUST 为禁用状态

#### Scenario: Hub 开启后可读

- **WHEN** 运维对有子根成功勾选合并预测并保存
- **THEN** 之后经事件列表或 options 读取该事件时 MUST 返回 `mergePredict` 为真

### Requirement: 事件 options 与列表 MUST 附带合并预测标志

凡返回 `entity.Event`（或等价事件字典项）的 App/内部 options 与后台列表路径，MUST 包含 `mergePredict` 字段，供 Flutter 预测算法读取。字段缺失 MUST 被客户端视为 false；实现 MUST 保证新列映射到响应（旧缓存淘汰后）。

#### Scenario: Flutter 可从 options 识别合并预测根

- **WHEN** 客户端请求历史事件 options（或等价字典接口）且字典中存在已标记合并预测的根事件
- **THEN** 响应列表中对应项 MUST 包含 `mergePredict` 为真

---

### Requirement: System SHALL accept full-replace active ETA sync from the App

The system SHALL expose an authenticated App HTTP API `PUT /device/api/predict/active-eta/pending` for a bound device to full-replace the baby’s in-progress timing ETA list in Redis (MUST NOT use MySQL as authority; MUST NOT require TTL expiry). The request body MUST use `deviceNo` and `items`; the response MUST include `count` (number of retained items after cleaning). Each retained item MUST have positive `rootEventId` and positive `leafEventId`, positive `startAt` and.

#### Scenario: Replace list

- **WHEN** 客户端先后两次成功提交同一 deviceNo 的不同 items
- **THEN** Redis 中该宝宝进行中 ETA MUST 仅等于第二次内容

#### Scenario: Empty clears

- **WHEN** 客户端提交空 items 且成功
- **THEN** 该宝宝 active-eta Redis MUST 为空且响应 count 为 0

#### Scenario: Discard invalid keep valid

- **WHEN** 同一次 PUT 含一条非法项（leaf≤0 或 expectEndAt 小于 startAt）与一条合法项
- **THEN** 同步 MUST 成功且 Redis 仅含合法项

#### Scenario: Dedup by leafEventId

- **WHEN** 同一次 PUT 含两条相同 leafEventId 的合法项
- **THEN** Redis 中该 leaf MUST 仅保留后一条

#### Scenario: No push side effects

- **WHEN** 同步成功写入一条 ETA
- **THEN** 系统 MUST NOT 因此发布预测临近延时消息或可见推送

#### Scenario: Sync busy

- **WHEN** 同 deviceNo 同步锁已被占用
- **THEN** 新请求 MUST 失败返回同步繁忙类错误且 MUST NOT 改写 Redis 列表

### Requirement: System SHALL expose internal GET of active ETAs for Python

The system SHALL provide `GET /voice/internal/api/predict/active-eta` that returns `{ "items": [...] }` for a `deviceNo` from Redis. When the key is missing or empty, the response MUST be an empty `items` list (not an error solely for emptiness). Field names MUST align with the App sync item schema (`rootEventId`, `leafEventId`, `historyId`, `title`, `startAt`, `expectEndAt`) so Python can match identifiers.

系统 MUST 提供内部 GET 返回 `{ items }`；空键 MUST 返回空列表；字段与 App 同步对齐。

#### Scenario: Get returns stored items

- **WHEN** Redis 已有一条 ETA 且内部 GET 该 deviceNo
- **THEN** 响应 items MUST 包含对应 startAt/expectEndAt 与 rootEventId/leafEventId 字段

#### Scenario: Missing key empty list

- **WHEN** Redis 无该宝宝 active-eta 键
- **THEN** GET MUST 成功返回空 items 列表

### Requirement: Active-eta sync SHALL update Timing Live Activity expectEndAt when historyId is known

After a successful full-replace write of active-eta Redis state for a device, for each cleaned item that has a parseable positive `historyId` and `expectEndAt` > 0, the server MUST fan out a Timing Live Activity **update** (via the existing push Live Activity internal path) to bound users of that device, with `activityId` equal to `timing:{deviceNo}:{historyId}` and ContentState including `expectEndAt` (unix seconds). This path MUST NOT publish delayed MQ messages and MUST.

#### Scenario: ETA sync updates Timing ContentState

- **WHEN** `PUT .../active-eta/pending` 成功写入含 `historyId=42`、`expectEndAt=1730000000` 的条目
- **THEN** 服务端 MUST 对 activityId `timing:{deviceNo}:42` 扇出 Timing update
- **AND** ContentState MUST 含 `expectEndAt=1730000000`
- **AND** MUST NOT 因此 PublishDelayed 或发普通 predict alert

#### Scenario: Clear ETA does not end Timing

- **WHEN** active-eta 同步以空列表清空 Redis
- **THEN** MUST NOT 要求仅因此 end 所有 Timing Live Activity

#### Scenario: Bad historyId skips LA only

- **WHEN** 某条 ETA 的 `historyId` 无法解析为正整数
- **THEN** 该条 MUST 跳过 LA update
- **AND** 其余合法条目与 Redis 写入成功语义 MUST 不受影响

### Requirement: Active-eta storage MUST NOT schedule imminent pushes

The active-eta Redis store remains authoritative for in-progress timing ETAs and MUST NOT publish delayed MQ wakeups or drive Imminent Live Activity / ordinary predict-imminent alerts. After a successful sync write, the server MAY fan out Timing Live Activity **update** payloads that only refresh ContentState fields such as `expectEndAt`, as specified by「Active-eta sync SHALL update Timing Live Activity expectEndAt when historyId is known」. This exception MUST NOT be used.

#### Scenario: No imminent scheduling from active-eta

- **WHEN** 客户端成功同步 active-eta 列表
- **THEN** 服务端 MUST NOT 仅为该同步 PublishDelayed 预测临近 MQ
- **AND** MUST NOT 扇出 Imminent Live Activity start

---

### Requirement: 原双档变更中的 Pro 开通壳 MUST 收敛保留

在回退基础版与推送门禁后，系统 MUST 仍满足 Pro 人对开通（支付/邀请/7 日试用）与喂养前置，见 `predict-pro-unlock-shell` 与 `predict-pro-feeding-eligibility`。

#### Scenario: Pro 仍可支付开通

- **WHEN** 喂养合格用户支付 Pro SKU 成功
- **THEN** 系统 MUST 授予 Pro 人权益

### Requirement: 预测两档开通形态 MUST 对齐成长轨迹（对人）

支付建单、邀请兑换、Admin 授与、catalog 项字段（除本变更明确新增的 `status` 与广告字段清理外）MUST 与成长轨迹同类权益功能同构：权益落 `feature_user_entitlement`；`InviteOncePerUser` 对两档 MUST 为 true。Hub「开通功能管理」MUST 可平铺选中两档并编辑定义/SKU（样式与成长轨迹一致，规则文案为人维）。

#### Scenario: 付费开通基础版写人权益

- **WHEN** 用户支付基础版 SKU 成功
- **THEN** 系统 MUST 为该 `wx_id` 写入/续期 `predict_capability_basic`，`unlockMethod=payment`

#### Scenario: 同账号邀请开通 Pro 仅一次

- **WHEN** 某 wx 已成功用任意邀请码开通 `predict_capability_pro` 后再次用另一码兑换同一功能
- **THEN** 系统 MUST 拒绝，MUST NOT 延长权益

### Requirement: 预测两档 MUST 支持显式点试用且时长为 7 日

系统 MUST 提供已登录 App API（路径以实现为准，如 `POST /cash/app/api/feature/trial/start`）领取预测功能试用。请求 MUST 含 `featureId`（仅允许两档预测功能 ID，或允许其它已配置试用的功能但成长轨迹/值得留意 MUST NOT 改走本接口改变其 24h soft 语义）。领取成功时 MUST：

- 将 `feature_user_trial` 对该 `(wx_id, feature_id)` 标为 used；
- 授予 7 日账号维权益，`unlock_method=trial`，`expires_at` 自领取时刻起算；
- 同机其他账号的试用资格 MUST NOT 被占用（同机试用作废）。

已 used、或已有该功能未过期人权益时 MUST 拒绝或幂等跳过（不得重复再授 7 日）。VIP 覆盖期间 MUST NOT 消耗试用资格（对齐成长轨迹：保留 unused）。

#### Scenario: 点试用起算 7 日

- **WHEN** 未试用过的用户对 `predict_capability_basic` 调用点试用且当前无该功能有效权益且非 VIP
- **THEN** 系统 MUST 写入约 7 日后到期的人权益，且试用状态变为 used

#### Scenario: 同机另一账号仍可试用

- **WHEN** 设备 D 上账号 A 已领取基础版试用，账号 B（不同 wx）绑定 D 且从未试用基础版
- **THEN** B 的 catalog `trialAvailable` MUST 仍可为 true（在无自有权益且试用未用时）

#### Scenario: 成长轨迹试用语义不变

- **WHEN** 用户使用成长轨迹既有 soft/成功后 claim 路径
- **THEN** 时长与触发 MUST 仍为既有 24h 成功后 claim，MUST NOT 被本要求改为 7 日点试用

---

### Requirement: predict-imminent-capability-gate MUST 文档化为已退役推送门禁能力

本 capability 仅保留退役说明：imminent / Live Activity / timing 推送与写入 MUST NOT 再校验预测能力开通或 VIP；运行时约束见 `predict-push-ungate`。本文件 MUST 至少保留本条 Requirement，避免空规格。

#### Scenario: 文档指向 ungate

- **WHEN** 读者查阅本 capability
- **THEN** 规格 MUST 指向 `predict-push-ungate` 作为推送开放语义来源，且 MUST NOT 再要求开通校验

---

### Requirement: System SHALL expose internal GET of imminent pending for Python

The system SHALL provide `GET /voice/internal/api/predict/imminent/pending` that returns `{ "events": [. The route MUST be registered on voice-service behind the same internal secret middleware as `GET /voice/internal/api/predict/active-eta` (`X-Device-Gateway-Internal-Secret`). When the key is missing, empty, or contains illegal JSON, the response MUST be an empty `events` list (not an error solely for emptiness or corrupt payload). Field names MUST align with the App sync.

#### Scenario: Get returns stored events

- **WHEN** Redis 已有一条 imminent 待办且内部 GET 该 deviceNo（密钥合法）
- **THEN** 响应 events MUST 包含对应 eventId/nextAt/title 及已存 expect* 字段

#### Scenario: Missing key empty list

- **WHEN** Redis 无该宝宝 imminent pending 键
- **THEN** GET MUST 成功返回空 events 列表

#### Scenario: No push side effects

- **WHEN** 内部 GET 成功
- **THEN** 系统 MUST NOT 因此发布延时消息或可见推送

---

### Requirement: Imminent Live Activity end SHALL dismiss immediately and MUST NOT wipe nextAt

When Go fans out an Imminent Live Activity **end** (pending sync clear of all Imminent slots, or replacing the per-wx slot before a new start), the APNs liveactivity payload MUST include `event=end` and `dismissal-date` set to a unix timestamp in the **past** relative to send time. ContentState MUST include `phase` indicating cleared or ended; when the previous `nextAt`/`title` are known they MUST be included so a delayed dismiss does not decode `nextAt` as 0. Imminent Live.

#### Scenario: Pending PUT end includes past dismissal-date

- **WHEN** Go 因 `PUT …/predict/imminent/pending` 向绑定用户扇出 Imminent end
- **THEN** 每条 APNs end 体 MUST 含 `dismissal-date` 且值 MUST 早于发送时刻
- **AND** MUST NOT 因此结束 Timing Live Activities

#### Scenario: Slot replace end includes past dismissal-date

- **WHEN** Go 在 Imminent start 前对该 wx 先 end 旧槽
- **THEN** 该 end MUST 同样携带过去的 `dismissal-date`

#### Scenario: Imminent end without dismissal-date is forbidden

- **WHEN** push-service 组装 Imminent `event=end` 的 APNs JSON
- **THEN** `aps` MUST 包含键 `dismissal-date`
- **AND** 业务未传 `DismissalDate` 时组装层 MUST 填入过去 unix 秒

### Requirement: Imminent Live Activity start Attributes SHALL carry title

When Go fans out an Imminent Live Activity **start**, the fixed attributes object MUST include `title` set to the same display sentence as ContentState `title` (the imminent alert body). ContentState MUST continue to include `title`. Imminent **start** 固定 Attributes MUST 含 `title`，与 ContentState `title` 为同一句展示文案。 ContentState MUST 继续含 `title`。

#### Scenario: Imminent start attributes include title

- **WHEN** Go 向已注册 token 的用户扇出 Imminent start 且 alert 非空
- **THEN** 固定 Attributes MUST 含键 `title` 且值等于该 alert
- **AND** ContentState MUST 含相同 `title`

### Requirement: Imminent Live Activity payload SHALL use unix-second nextAt and MUST NOT carry visuals

When Go fans out an Imminent Live Activity **start**, fixed attributes MUST continue to include `eventId` and `nextAt`, and ContentState MUST include `nextAt`, each as unix timestamps in **seconds** (not milliseconds). Attributes and ContentState MUST NOT include `logo`, `color`, `logoFile`, or remote image URL fields; Imminent Live Activity **start** 时，固定 Attributes MUST 继续含 `eventId` 与 `nextAt`，ContentState MUST 含 `nextAt`，且均为 unix **秒**（非毫秒）。 Attributes 与 ContentState.

#### Scenario: nextAt is unix seconds

- **WHEN** Imminent start 载荷含 Attributes 或 ContentState 的 `nextAt`
- **THEN** 其值 MUST 为 unix 秒

#### Scenario: eventId present without logo color

- **WHEN** Go 向已注册 token 的用户扇出 Imminent start 且事件 id 有效
- **THEN** Attributes MUST 含 `eventId`
- **AND** MUST NOT 同时下发 `logo` / `color` / `logoFile`

### Requirement: Server SHALL rate-limit Imminent fanout one per user per minute first-wins

When the twin Go service is about to successfully fan out an Imminent Live Activity start (or a non-iOS legacy visible predict-imminent push) to a bound user, it MUST enforce at most one successful Imminent fanout per user per rolling 60-second window. The first eligible fanout in the window MUST proceed; later eligible Imminent fires for that user within the window MUST NOT push. Timing Live Activity start/end MUST NOT consume or be blocked by this gate. 对绑定用户即将成功扇出.

#### Scenario: Second Imminent in same minute dropped

- **WHEN** 同一用户在 60 秒内已成功扇出一条 Imminent，又有另一事件通过既有开火闸
- **THEN** Go MUST NOT 为后一次再推 Imminent（或该用户的等价预测提醒）

#### Scenario: Timing unaffected

- **WHEN** 同一用户在 Imminent 1/min 窗口内需扇出 Timing Live Activity start 或 end
- **THEN** Go MUST 仍按 Timing 规则推送

### Requirement: Server SHALL keep a single Imminent slot per user when starting

When Go is about to start a new Imminent Live Activity for a user and an Imminent slot may still be active for that user, Go MUST end the previous Imminent before or as part of presenting the new content so that at most one Imminent Live Activity is intended per user (wxId). Bound guardians of the same baby/device MUST each have their own Imminent slot; clearing and starting MUST fan out per bound wx. Fanout that fails the one-per-minute gate MUST NOT perform this replace.

#### Scenario: Fire replaces previous Imminent

- **WHEN** Go 对某用户通过 1/min 与既有开火闸并即将 start Imminent，且 Redis（或等价）显示仍有旧 Imminent activity id
- **THEN** Go MUST 先结束旧 id（或同槽更新为最新内容）再呈现单槽

#### Scenario: Multiple guardians each get a slot

- **WHEN** 同一 deviceNo 绑定两个 wx，且开火通过闸并对两用户均有可用 LA token
- **THEN** Go MUST 允许两名用户各自拥有一条 Imminent Live Activity（非共享单一 device 槽）

### Requirement: Live Activity tokens SHALL use a dedicated App path and storage

The twin Go service MUST expose App-facing `POST /app/api/push/live-activity/register` (separate from `POST /app/api/push/register`) for registering and updating Live Activity Push-to-Start and per-activity update tokens. Those tokens MUST be stored in a push-domain table (or equivalent) distinct from `push_device` device-token uniqueness (`uk_token`). The request body MUST include a non-zero `wxId` together with `deviceKey`, `tokenKind` (`push_to_start`|`activity`).

#### Scenario: Unauthenticated register with body wxId succeeds

- **WHEN** 客户端（含 iOS 原生唤醒路径）POST `/app/api/push/live-activity/register`，无 Authorization，body 含非零 `wxId` 与合法既有字段
- **THEN** gateway-app MUST NOT 仅因缺少 Bearer 返回 401
- **AND** Go MUST 将该 token upsert 到该 `wxId` 名下的 LA 专用存储

#### Scenario: Missing wxId rejected

- **WHEN** 请求缺少非零 `wxId`（即使带有有效 Bearer 与内部头）
- **THEN** 服务端 MUST 拒绝该次 register（参数错误类）
- **AND** MUST NOT 静默用内部头 `X-Internal-Wx-Id` 顶替 body.`wxId`（本变更默认不兼容旧客户端）

#### Scenario: LA token not stored as device push token

- **WHEN** 客户端上报 Push-to-Start 或 per-activity update token
- **THEN** Go MUST 经 live-activity register path 持久化到 LA 专用存储
- **AND** MUST NOT 将该 token 作为 `push_device` 全局唯一 device token 写入

#### Scenario: Device register unchanged for alerts and still requires Bearer

- **WHEN** 客户端调用既有 `/app/api/push/register` 或 `unregister` 上报/注销 device alert token
- **THEN** 行为 MUST 保持 device alert 注册语义与既有 Bearer 要求
- **AND** MUST NOT 因本变更被鉴权豁免
- **AND** MUST NOT 要求该调用携带 LA token 才能工作

### Requirement: iOS Imminent delivery SHALL use Live Activity not alert

For APNs/iOS recipients, predict-imminent offline reminders that pass all gates MUST be delivered as ActivityKit Live Activity pushes (Push to Start / liveactivity update as applicable), and MUST NOT use a normal visible APNs alert with `bizType=predict_imminent` as the carrier on that path. Missing Live Activity / Push-to-Start token MUST skip that recipient without failing the whole fire Ack path. 对 iOS/APNs 接收方，通过全部闸的预测临近 MUST 以 Live Activity 推送送达，MUST NOT 以普通可见.

#### Scenario: iOS uses liveactivity

- **WHEN** Go 向已注册 Push to Start（或等价）的 iOS 设备扇出 Imminent
- **THEN** 发送 MUST 为 liveactivity 路径，MUST NOT 依赖普通预测 alert

#### Scenario: No token skips recipient

- **WHEN** 某绑定用户无可用 Live Activity / 推送 token
- **THEN** Go MUST 跳过该用户并继续处理其余接收方

### Requirement: Server SHALL clear all Imminent Live Activities on history mutation

When the twin Go **history** service successfully persists a history create, update, or delete for a device (including instantaneous events with non-zero end time, and any successful update regardless of which fields changed), it MUST send Live Activity end (or equivalent dismiss) for all tracked Imminent Live Activities to bound users of that baby/device, and MUST NOT end Timing Live Activities as part of this clear. The voice handling of `PUT.

#### Scenario: History create clears Imminent and pending

- **WHEN** history 成功新增一条记录（含瞬时事件）
- **THEN** Go MUST 向该 deviceNo 绑定用户发出清空/结束全部 Imminent Live Activity 的推送（或等价）
- **AND** MUST NOT 因此结束 Timing Live Activities
- **AND** MUST 删除该 deviceNo 的 Redis imminent pending 键

#### Scenario: History update clears Imminent and pending

- **WHEN** history 任意成功更新（含仅改备注、闭合、改开始时间或事件名等）
- **THEN** Go MUST 同样清全部 Imminent 并删除 Redis imminent pending
- **AND** MUST NOT 因此结束 Timing

#### Scenario: History delete clears Imminent and pending

- **WHEN** history 删除某行成功
- **THEN** Go MUST 同样清全部 Imminent 并删除 Redis imminent pending

#### Scenario: Pending PUT no longer clears Imminent

- **WHEN** 客户端提交 `PUT …/predict/imminent/pending`（空或非空列表）
- **THEN** Go MUST NOT 仅因该 PUT 向绑定用户扇出 Imminent 全员 end/清屏
- **AND** 既有写 Redis / 延时投递语义 MUST 保持

#### Scenario: Activity token rows removed after Imminent end

- **WHEN** Go 对某 `activityId` 扇出 Imminent `event=end`（history 清屏或开火换槽）
- **THEN** push 域 MUST 删除该 `activityId` 下 `token_kind=activity` 的持久化行
- **AND** MUST NOT 删除该用户的 Push-to-Start 行

### Requirement: Imminent clear on history MUST be best-effort after commit

Imminent Live Activity clear and Redis pending deletion triggered by history mutation MUST run only after the history write has successfully committed, and MUST NOT roll back or fail the history write if clear or Redis delete fails. Failures MUST be logged. Timing Live Activity start/update/end on the same write path MUST remain independently best-effort as today. 由 history 变更触发的 Imminent 清屏与 Redis pending 删除 MUST 仅在 history 写成功提交之后执行； 清屏或 Del Redis 失败 MUST NOT 回滚或导致该次.

#### Scenario: Clear failure does not fail history write

- **WHEN** history 落库成功但 Imminent end 或 pending Del 失败
- **THEN** 该次 history 写路径 MUST 仍视为落库成功
- **AND** MUST 输出可观测失败日志

### Requirement: Gateway-app SHALL exempt live-activity register from mandatory Bearer

The gateway-app Bearer middleware MUST treat `POST /app/api/push/live-activity/register` as an auth-exempt exact POST path (same maintenance style as other entries in `gatewayAppAuthExemptExactPOST`). Spoofed internal identity headers MUST still be stripped before handling. If a non-admin user Bearer is present on an exempt request, the gateway MAY still inject internal headers, but Live Activity register persistence identity MUST follow body `wxId` rules. The gateway MUST.

#### Scenario: Exact path exempt listed

- **WHEN** 检查 gateway-app 鉴权豁免配置
- **THEN** `POST /app/api/push/live-activity/register` MUST 出现在精确 POST 豁免集合中
- **AND** MUST NOT 通过豁免整个 `/app/api/push/` 前缀来实现

### Requirement: Imminent Live Activity start ContentState SHALL carry optional structured expect fields

When Go fans out an Imminent Live Activity **start**, ContentState MUST continue to include `title`, `phase`, and `nextAt` (unix seconds). When the fire/pending payload has a usable structured expect, ContentState MUST also include: `expectKind` set to `duration` or `amount`; The server MUST NOT write a preformatted Chinese expect label into ContentState for this path, MUST NOT require `expectLeafTitle` in ContentState, and MUST NOT add `logo` / `color` / `logoFile`. When.

#### Scenario: Duration expect on Imminent start

- **WHEN** Go 扇出 Imminent start 且开火载荷 `expectKind=duration`、`expectDurationSec=5400`
- **THEN** ContentState MUST 含 `expectKind=duration` 与 `expectDurationSec=5400`
- **AND** MUST NOT 含成品「预计1时30分钟」类 label 键（除非另有与本需求无关的既有 alert 字段）

#### Scenario: Amount expect on Imminent start

- **WHEN** 开火载荷 `expectKind=amount`、`expectAmount=120.5`、`expectUnit=ml`
- **THEN** ContentState MUST 含上述三键且 amount 为浮点语义
- **AND** MUST NOT 要求写入 `expectLeafTitle`

#### Scenario: No expect omits keys

- **WHEN** 开火载荷无 expect 或 duration/amount 不可用
- **THEN** ContentState MUST 省略 `expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit`

---

### Requirement: 系统 MUST 提供可按业务类型发送的公共可见推送

系统 MUST 提供跨域可调用的公共推送契约（HTTP internal 或等价），入参至少包含接收方 `wxId`、业务类型 `bizType`、可见文案（及可选深链/data）。当 `bizType` 为预测临近（实现常量名 MAY 为 `predict_imminent`）时，系统 MUST 向该用户已注册的推送 token（APNs/HMS/MiPush）发送可见通知，且 MUST NOT 使用 UCG 社区未读总和作为该通知的角标语义（MAY 省略 badge 或使用与社区无关的固定策略）。调用方（含 voice）MUST 经 `clients/*` 契约调用，MUST NOT 直接 import `internal/services/ucg` 业务包或直查 ucg 库表。

#### Scenario: 预测临近推送不套用社区角标

- **WHEN** voice 以预测临近 `bizType` 请求向某 `wxId` 发送可见推送
- **THEN** 下发载荷 MUST 为可见提醒，且 MUST NOT 把 UCG 会话/通知未读之和写入该推送的业务角标语义

### Requirement: 预测临近推送 MUST 扇出到宝宝全部绑定用户

当预测临近叫醒校验通过后，系统 MUST 解析该 `deviceNo` 下全部绑定 `wxId`，并对每一用户尝试公共推送。单个用户无有效 token 或发送失败 MUST NOT 阻止对其余用户的尝试；整体流程 MUST 仍按调度规格 Ack 消息。

#### Scenario: 两名看护均收到尝试

- **WHEN** 某 `deviceNo` 绑定两个 `wxId` 且均有有效推送 token，且叫醒校验通过
- **THEN** 系统 MUST 对两个 `wxId` 均发起可见推送调用

#### Scenario: 无 token 用户跳过

- **WHEN** 某绑定 `wxId` 没有可用推送 token
- **THEN** 系统 MUST 跳过该用户且 MUST 继续处理其他绑定用户

---

### Requirement: 预测临近可见推送正文 MUST 使用昵称与 title 模板

当 `voice-service` 在预测临近叫醒路径上准备向绑定用户发送 `bizType=predict_imminent` 的**可见**系统推送时，通知正文（传入推送契约的 `alert`）MUST 以基句 `{宝宝昵称}要{事件名}了` 开头，其中：

- **事件名** MUST 为该叫醒载荷中的 `title`（trim 后的非空字符串），MUST NOT 由服务端展开事件树选取子事件名，MUST NOT 使用「事件将在约 N 分钟内发生」类时间回落文案作为事件名。 - **宝宝昵称** MUST 在发送前经 device 画像契约按 `deviceNo` 读取； 当返回的 `babyName` 为空、或画像读取失败时，昵称 MUST 回落为字面量「宝宝」。 - 当载荷含**合法** expect 时，系统 MUST 在基句之后追加统一「约」后缀，且 MUST NOT 使用「需要」措辞。 合法判定与时长算法 MUST 遵循本变更 design D1/D3：
  - `expectKind=duration` 且.

#### Scenario: 有昵称与 title

- **WHEN** 画像 `babyName` 为「小宝」且叫醒载荷 `title` 为「换尿布」，无 expect，且其它推送闸均通过
- **THEN** 发往推送契约的 `alert` MUST 等于「小宝要换尿布了」

#### Scenario: 昵称为空回落

- **WHEN** 画像 `babyName` 为空串（或画像读取失败）且 `title` 为「吃奶」，无 expect，且其它推送闸均通过
- **THEN** `alert` MUST 等于「宝宝要吃奶了」

#### Scenario: 不使用旧前缀

- **WHEN** 校验通过并发送可见预测临近推送
- **THEN** `alert` MUST NOT 以「宝宝提醒：」开头

#### Scenario: 追加 duration 后缀

- **WHEN** 基句条件满足，且载荷 `expectKind=duration`、`expectDurationSec=3900`，其它闸通过
- **THEN** `alert` MUST 等于「{昵称}要{title}了，预计这次约1小时5分钟」

#### Scenario: duration 不足一分钟

- **WHEN** 基句条件满足，且 `expectKind=duration`、`expectDurationSec=59`，其它闸通过
- **THEN** `alert` MUST 以「，预计这次约不足1分钟」结尾

#### Scenario: duration 整点小时

- **WHEN** 基句条件满足，且 `expectKind=duration`、`expectDurationSec=3600`，其它闸通过
- **THEN** `alert` MUST 以「，预计这次约1小时」结尾（无「0分钟」）

#### Scenario: 追加 amount 后缀含叶子名

- **WHEN** 基句条件满足，昵称为「小宝」，`title` 为「喂养」，`expectKind=amount`、`expectAmount=120`、`expectUnit=毫升`、`expectLeafTitle=配方奶`，其它闸通过
- **THEN** `alert` MUST 等于「小宝要喂养了，预计这次约120毫升配方奶」

#### Scenario: amount 无单位仍追加

- **WHEN** 基句条件满足，`expectKind=amount`、`expectAmount=120`、`expectUnit` 为空、无叶子，其它闸通过
- **THEN** `alert` MUST 等于「{昵称}要{title}了，预计这次约120」

#### Scenario: 非法 expect 不追加

- **WHEN** 基句条件满足，但 `expectKind=duration` 且 `expectDurationSec` 缺失或 ≤0
- **THEN** `alert` MUST 仅为基句，MUST NOT 追加预计后缀

#### Scenario: 未知 kind 不追加

- **WHEN** 基句条件满足，但 `expectKind` 为空或非 `duration`/`amount`
- **THEN** `alert` MUST 仅为基句

### Requirement: 空 title 时 MUST NOT 发送可见推送

当叫醒载荷中的 `title` 经 trim 后为空时，系统 MUST NOT 调用推送发送（MUST NOT 向任何绑定 wx 下发该次可见 `predict_imminent` 通知），MUST 记录可观测日志，且消费路径 MUST 仍按现有约定 Ack（MUST NOT 因空 title 而 Nack requeue 或再次 Publish 延时重试）。

#### Scenario: 缺 title 跳过

- **WHEN** Redis 匹配与其它闸均可通过，但载荷 `title` 为空或仅空白
- **THEN** 系统 MUST NOT 发送该次可见推送，且 MUST Ack 该叫醒消息

#### Scenario: 有 title 才发送

- **WHEN** 载荷 `title` 为非空「睡觉」且其它闸通过
- **THEN** 系统 MUST 按模板拼装正文并尝试向绑定用户发送可见推送

---

### Requirement: 客户端 MUST 能全量替换宝宝待发生预测节点

系统 MUST 提供经 gateway-app 暴露的 App HTTP 接口，供已登录且绑定设备的客户端在预测更新后提交该宝宝（`deviceNo`）的待发生事件全量列表。每条事件 MUST 包含 `eventId` 与 `nextAt`（unix 秒）。服务端 MUST 以该列表**整表替换**该宝宝在 Redis 中的待办（MUST NOT 写入业务 MySQL 作为权威）。待办键 MUST 经 `cachekit` 登记且 MUST NOT 依赖 TTL 过期清理（靠后续替换维护）。后一次成功写入 MUST 覆盖前一次（最后写入赢）。空列表 MUST 清空该宝宝待办。

#### Scenario: 预测刷新后覆盖旧列表

- **WHEN** 客户端先后两次成功提交同一 `deviceNo` 的不同 `events` 列表
- **THEN** Redis 中该宝宝待办 MUST 仅等于第二次提交内容，且 MUST NOT 残留仅存在于第一次列表中的事件节点

#### Scenario: 空列表清空

- **WHEN** 客户端提交空 `events` 且请求成功
- **THEN** 该宝宝 Redis 待办 MUST 为空，后续基于旧节点的叫醒校验 MUST 失败并跳过推送

### Requirement: 同步成功后 MUST 按提前量投递延时叫醒消息

对提交列表中每一条有效事件，系统 MUST 计算 `fireAt = nextAt - leadSeconds`（秒； 实现 MUST 使用显式常量并注释对齐该上限）时，系统 MUST 将消息发布到已启用 `rabbitmq_delayed_message_exchange` 的延时交换机（headers `x-delay=delayMs`），routing 到 voice 预测临近队列； 消息体 MUST 足以在消费时定位 `deviceNo`、`eventId`、`nextAt`。 当 `fireAt <= now` 时 MUST 仍投递（delay 为 0 或等价即时投递），不得因「已过理想叫醒点」而拒绝入队。 当 `delayMs` **大于**上述上限时，系统 MUST NOT 为该事件调用延时发布（MUST NOT 投递会导致插件 nodelay 立刻路由的超大 `x-delay`），MUST 仍将该事件保留在 Redis 待办中，MUST NOT 因此使整表同步失败，MUST 记录可观测跳过日志（至少含 `eventId`.

#### Scenario: 五分钟前提醒入队

- **WHEN** 当前时间为 T0，某事件 `nextAt = T0 + 600`
- **THEN** 系统 MUST 发布延时消息且延迟约 300 秒量级（允许实现取整误差），消息携带该 `deviceNo`、`eventId`、`nextAt`

#### Scenario: 已进入窗口仍入队

- **WHEN** 某事件 `nextAt - leadSeconds <= now`
- **THEN** 系统 MUST 仍发布叫醒消息（零延迟或即时），MUST NOT 静默丢弃该事件的叫醒

#### Scenario: 超过插件延时上限则只写 Redis 不发 MQ

- **WHEN** 某有效事件计算得 `delayMs > 4294967295`（例如 `nextAt` 约在两年后）且整表同步成功
- **THEN** 该事件 MUST 出现在 Redis 待办中，系统 MUST NOT 为其发布延时叫醒消息，同步 API MUST 仍成功，且 MUST 有可观测跳过日志

#### Scenario: 同批近端远点互不影响

- **WHEN** 同一次提交同时包含 `delayMs` 在上限内与超过上限的有效事件
- **THEN** 近端事件 MUST 照常入队延时消息，远端事件 MUST 仅保留于 Redis 且不入队，整表同步 MUST 成功

### Requirement: 消费叫醒时 MUST 以 Redis 校验为准且业务失败 MUST Ack

`voice-service` MUST 以 AMQP push consumer（`autoAck=false`）消费预测临近队列。处理单条 delivery 时 MUST 先读取该宝宝当前 Redis 待办：若不存在匹配的 `eventId` 与相同 `nextAt`，MUST 跳过推送并对该 delivery **Ack**。下列情况 MUST **Ack** 且 MUST NOT `Nack(requeue=true)`，MUST NOT 在消费路径再次 Publish 延时重试消息：载荷非法、Redis 不匹配、**进行中闸命中或进行中查询失败**、推送去重命中、历史闸命中、列绑定用户失败、系统推送发送失败。进程崩溃导致未 Ack 时 broker 可重投；应用层 MUST 依赖去重键限制重复可见推送（进行中闸跳过 MUST NOT 依赖去重键）。

#### Scenario: 列表已替换则旧消息作废

- **WHEN** 延时消息携带的 `nextAt` 与 Redis 当前该 `eventId` 的 `nextAt` 不一致或不存在该事件
- **THEN** consumer MUST NOT 发送系统推送，且 MUST Ack 该消息

#### Scenario: 推送失败不回队

- **WHEN** 校验通过但向厂商推送返回错误
- **THEN** consumer MUST 记录可观测日志、MUST Ack，且 MUST NOT Nack requeue，MUST NOT 再发布一条延时消息作为重试

### Requirement: 推送前 MUST 做五分钟去重与三十分钟历史闸

对同一 `(deviceNo, eventId)`，在**进行中闸已通过**（无同类 `end_time=0` 记录且查询成功）之后，成功进入推送发送路径前 MUST 检查推送去重状态：若五分钟内已推送过，MUST 跳过并 Ack。若未去重命中，MUST 查询 history：该宝宝在过去 1800 秒内是否存在同类事件记录（`start_time` 落在窗口内）；若存在 MUST 跳过推送并 Ack。当上报 `eventId` 为事件树一级根时，历史查询 MUST 包含该根及其后代叶子 `eventId` 集合（经 device 事件目录解析），MUST NOT 仅用根 id 查询导致漏判。通过闸后 MUST 写入五分钟去重标记（成功发送或「已尝试发送」的语义由实现固定为「进入发送前占位」或「发送成功后写入」，但 MUST 保证五分钟内不会对用户重复可见推送风暴）。进行中闸命中或进行中查询失败时 MUST NOT 适用本要求的去重占位（见「推送前 MUST 抑制同类进行中事件」）。

#### Scenario: 用户已记录喂养则不推

- **WHEN** Redis 仍匹配某喂养类预测节点，但 history 在近 30 分钟内已有对应叶子事件
- **THEN** 系统 MUST NOT 发送该节点的临近推送

#### Scenario: 五分钟内不重复推同一事件

- **WHEN** 同一 `(deviceNo, eventId)` 在五分钟内第二次通过 Redis 匹配进入推送逻辑（且进行中闸已通过）
- **THEN** 系统 MUST 跳过第二次可见推送

### Requirement: 叫醒消费宿主与拓扑 MUST 声明

本能力的 AMQP consumer MUST 运行于 `voice-service`。系统 MUST 使用独立于即时 `voice.events` 业务的延时交换机（启用 delayed message 插件），并声明专用队列与绑定。变更与部署文档 MUST 说明插件启用方式；插件不可用时同步路径 MUST 失败可观测，MUST NOT 静默假装已调度。

#### Scenario: voice 进程订阅专用队列

- **WHEN** `voice-service` 启动且 MQ 配置可用
- **THEN** 系统 MUST 订阅预测临近专用队列以消费延时到期消息

### Requirement: Pending items MAY carry optional expect fields through Redis and MQ

The `PUT /device/api/predict/imminent/pending` event item schema SHALL accept optional fields `expectKind`, `expectDurationSec`, `expectAmount`, `expectUnit`, and `expectLeafTitle` in addition to `eventId`, `nextAt`, and `title`. On successful full-replace sync, the voice service MUST persist these fields on the Redis pending item when present, and MUST copy them onto the delayed fire message payload alongside `deviceNo`, `eventId`, `nextAt`, and `title`. Redis match for.

#### Scenario: Sync stores duration expect

- **WHEN** 客户端提交含 `expectKind=duration` 与正 `expectDurationSec` 的事件且同步成功
- **THEN** Redis 该宝宝待办中对应项 MUST 保留上述 expect 字段，且随后发布的延时消息 MUST 携带相同字段

#### Scenario: Sync stores amount expect

- **WHEN** 客户端提交含 `expectKind=amount`、正 `expectAmount`、单位与可选叶子名的事件且同步成功
- **THEN** Redis 与延时消息 MUST 保留这些 expect 字段

#### Scenario: Omitting expect still syncs

- **WHEN** 客户端仅提交 `eventId`/`nextAt`/`title`（无 expect）
- **THEN** 同步 MUST 成功，行为与扩展前一致（待办可调度）

### Requirement: 推送前 MUST 抑制同类进行中事件

当 `voice-service` 处理预测临近叫醒且 Redis 待办已匹配该 `eventId`/`nextAt` 后，在五分钟去重占位与三十分钟 `start_time` 历史闸之前，系统 MUST 判断该宝宝是否存在同类**进行中**历史：进行中定义为 history 行满足 `end_time=0`。 当叫醒载荷 `eventId` 为事件树一级根时，查询 MUST 包含该根及其后代叶子 `eventId` 集合（与三十分钟历史闸同一展开规则）； 若展开集合中**任一** id 存在 `end_time=0` 行，系统 MUST 跳过可见推送并对该 delivery **Ack**，MUST NOT 写入五分钟去重键，MUST 记录可观测跳过原因（实现日志 reason MAY 为 `in_progress`）。 查询 MUST 经 history 服务契约完成，MUST NOT 由 voice 直查 history 库表。 契约查询 MUST 以存在性为目标（最少行扫描，例如 `end_time=0` 且 eventId IN.

#### Scenario: 长时段睡眠进行中不推

- **WHEN** Redis 仍匹配某睡眠类预测节点，且 history 存在对应叶子（或根）`end_time=0` 的未闭合行，即使该行 `start_time` 早于近 1800 秒窗口
- **THEN** 系统 MUST NOT 发送该节点的临近推送，MUST Ack，且 MUST NOT 写入该 `(deviceNo, eventId)` 的五分钟去重键

#### Scenario: 根预测因叶子进行中而抑制

- **WHEN** 叫醒 `eventId` 为一级根，其某后代叶子存在 `end_time=0` 记录
- **THEN** 系统 MUST 跳过该根节点的可见推送

#### Scenario: 不同根并行进行中不互相抑制

- **WHEN** 设备存在事件根 A 的进行中记录，叫醒载荷为不同根 B 且 B 子树无进行中、其它闸可通过
- **THEN** 系统 MUST NOT 仅因根 A 进行中而跳过根 B 的推送

#### Scenario: 进行中查询失败不推且不占去重

- **WHEN** Redis 已匹配，但进行中存在性查询返回错误
- **THEN** 系统 MUST NOT 发送可见推送，MUST Ack，且 MUST NOT 写入五分钟去重键

### Requirement: Imminent fire MUST apply user one-per-minute gate before send

After existing Redis match, in-progress, per-event dedup, and recent-history gates have been evaluated for an Imminent fire, and before a successful user-visible or Live Activity Imminent fanout, the twin Go service MUST apply the per-user rolling 60-second first-wins gate defined in `predict-imminent-live-activity-push`. A user who fails that gate MUST NOT receive that fire’s Imminent push and MUST Ack the delivery without treating it as a requeueable failure. Timing Live.

#### Scenario: Rate limit after other gates

- **WHEN** 某开火已通过 Redis 匹配与进行中/去重/近史闸，但该用户 60 秒内已成功扇出过 Imminent
- **THEN** Go MUST NOT 再向该用户发送本次 Imminent，MUST Ack

### Requirement: iOS Imminent fire delivery MUST follow live-activity-push rules

When fanout targets an iOS/APNs recipient, the send path MUST follow `predict-imminent-live-activity-push` (Live Activity, not normal predict alert). Non-iOS recipients MAY keep legacy visible predict-imminent notifications subject to the same user one-per-minute gate when that gate applies to Imminent fanout.

扇出目标为 iOS/APNs 时，发送 MUST 遵循 `predict-imminent-live-activity-push`（Live Activity，非普通预测 alert）。非 iOS MAY 保留遗留可见预测通知，且在适用时 MUST 同受用户 1/min 闸约束。

#### Scenario: Schedule delegates iOS carrier

- **WHEN** 开火对 iOS 用户进入发送阶段且已过全部闸（含 1/min）
- **THEN** 发送 MUST 走 Live Activity 路径而非普通 `predict_imminent` alert

### Requirement: History mutation MUST invalidate imminent pending Redis before Live Activity Refresh

When history create, update, or delete for a device succeeds, the twin Go service MUST delete (or empty) that device’s Redis imminent pending authority key as a post-commit side effect, and MUST do so before invoking the unified Live Activity Refresh for that device. Subsequent delayed fire consumers MUST continue to treat a missing or empty pending list as no-match (MUST NOT push ordinary `predict_imminent` solely because an old delayed message arrived). This invalidation.

#### Scenario: Pending key gone after history write

- **WHEN** 某 deviceNo 在 Redis 仍有非空 imminent pending，随后 history 对该宝宝写入成功
- **THEN** 该 deviceNo 的 imminent pending Redis 键 MUST 被删除或等价清空
- **AND** 之后到达的、仅对账旧 pending 的延时消息 MUST NOT 因旧列表再发普通预测推送

#### Scenario: Invalidate happens before Refresh

- **WHEN** history write triggers both pending invalidation and Live Activity Refresh
- **THEN** Refresh MUST observe the cleared pending (predict segment empty) rather than the pre-CUD list

### Requirement: Successful pending sync MUST trigger unified Live Activity Refresh

After a successful full-replace write of the device’s imminent pending Redis (including empty clear), the system MUST invoke the unified Live Activity Refresh for that `deviceNo`. This MUST NOT replace or skip delayed MQ publish rules already required by「同步成功后 MUST 按提前量投递延时叫醒消息」. This MUST NOT perform Imminent-only Live Activity end clear as the sync side effect.

imminent pending 整表替换成功后 MUST 触发统一 LA Refresh；MUST NOT 因此取消延时投递；MUST NOT 以 Imminent end 清屏作为 PUT 副作用。

#### Scenario: Non-empty PUT refreshes LA

- **WHEN** client successfully PUTs a non-empty pending list
- **THEN** Redis MUST hold that list, delayed fires MUST still be scheduled per existing rules, and unified Live Activity Refresh MUST run

#### Scenario: Empty PUT refreshes LA

- **WHEN** client successfully PUTs an empty pending list
- **THEN** Redis pending MUST be empty and unified Live Activity Refresh MUST run

### Requirement: Imminent MQ fire MUST keep ordinary push and also Refresh Live Activity

When delayed imminent fire matches Redis pending and passes existing business gates, the system MUST retain ordinary visible `predict_imminent` push behavior for bound users, and MUST also invoke unified Live Activity Refresh for the device. The fire path MUST NOT end-then-start a separate Imminent Live Activity. 延时开火在闸门通过后 MUST 保留普通可见推送，并 MUST 调用统一 LA Refresh； MUST NOT Imminent 先 end 再 start。

#### Scenario: Fire dual channel

- **WHEN** fire matches pending and gates pass for a device with bound users
- **THEN** system MUST attempt ordinary predict_imminent visible push AND MUST Refresh unified Live Activity

#### Scenario: Stale fire still skips both when pending mismatch

- **WHEN** fire payload does not match current Redis pending
- **THEN** system MUST NOT send ordinary predict push and MUST NOT be required to Refresh solely for that stale message (Ack as today)

### Requirement: History mutation MUST invalidate imminent pending Redis

When history create, update, or delete for a device succeeds, the twin Go service MUST delete (or empty) that device’s Redis imminent pending authority key as part of the same post-commit side effects that clear Imminent Live Activities. Subsequent delayed fire consumers MUST continue to treat a missing or empty pending list as no-match (MUST NOT push solely because an old delayed message arrived). This invalidation MUST NOT require a `PUT.

#### Scenario: Pending key gone after history write

- **WHEN** 某 deviceNo 在 Redis 仍有非空 imminent pending，随后 history 对该宝宝写入成功
- **THEN** 该 deviceNo 的 imminent pending Redis 键 MUST 被删除或等价清空
- **AND** 之后到达的、仅对账旧 pending 的延时消息 MUST NOT 因旧列表再推 Imminent

#### Scenario: Pending PUT still authoritative for new schedules

- **WHEN** history 已清空 pending 后客户端再次 `PUT …/predict/imminent/pending` 写入新列表
- **THEN** Go MUST 按既有整表替换与延时投递规则接受新日程
- **AND** MUST NOT 因该 PUT 执行 Imminent 全员清屏（清屏挂 history）

---

### Requirement: Server SHALL expose batch latest occurrence by event from history only

The system MUST provide an authenticated App HTTP API `GET /device/history/api/event/latest-by-event` that accepts `deviceNo` (same bind semantics as existing history App APIs). The handler MUST query only the history domain table (MUST NOT read `appointment_next` or other appointment stores). For each `event_id` that has at least one history row for that device, the response MUST include at most one summary whose occurrence time is that row's `start_time` (unix seconds).

#### Scenario: Multiple events one response

- **WHEN** 宝宝 history 含多个 eventId 的记录
- **THEN** 响应 MUST 为每个有记录的 eventId 给出一条最新发生，且 `occurredAt` 等于该条的 `start_time`

#### Scenario: In-progress counts as latest

- **WHEN** 某 eventId 最近一条 history 的 `end_time=0` 且其 `start_time` 最大
- **THEN** 该 eventId 的摘要 MUST 仍返回，且 `occurredAt` 为该未闭合行的 `start_time`

#### Scenario: Empty history

- **WHEN** 宝宝无 history 行
- **THEN** 响应 MUST 为空列表且 MUST NOT 报错

#### Scenario: Appointment next not mixed in

- **WHEN** 宝宝仅有预约下次约定、无对应 history 行
- **THEN** 该 eventId MUST NOT 仅因预约表出现在本接口结果中

### Requirement: latest-by-event MUST require login and MUST NOT count toward usage

The `GET /device/history/api/event/latest-by-event` route MUST require Bearer authentication (MUST NOT be listed on gateway auth exempt). The path MUST be registered in App usage maintenance skip list so it is not counted toward usage stats (owner confirmed). latest-by-event MUST 登录；MUST NOT 计入 usage；MUST NOT Bearer 豁免。

#### Scenario: Unauthenticated rejected

- **WHEN** 无有效 Bearer 调用该 GET
- **THEN** 网关或服务 MUST 拒绝（与既有需登录 history 读接口一致）

---

### Requirement: Imminent pending SHALL accept optional parentEventName

When accepting `PUT` imminent pending items, the service MUST accept an optional `parentEventName` (string). Missing or blank after trim MUST mean no parent. Persistence and subsequent GET / voice internal projections MUST round-trip the trimmed non-empty value when present. The service MUST NOT resolve the name to an event id and MUST NOT use it for scheduling, dedup, or in-progress gates (those remain `eventId`-based). imminent pending MUST 接受可选 parentEventName 并原样投影。

#### Scenario: Legacy body without parent name

- **WHEN** 旧客户端 PUT pending 项无 `parentEventName`
- **THEN** 服务 MUST 接受并调度，该项父名视为空

#### Scenario: Child item with root name

- **WHEN** 新客户端 PUT 子 `eventId` 且 `parentEventName` 为非空根事件名（如「吃奶」）
- **THEN** 存储与后续 GET / internal 读 MUST 能读回该 trim 后的 `parentEventName`

### Requirement: expectLeafTitle MUST be dropped end-to-end

The service MUST NOT require `expectLeafTitle`. Legacy clients MAY still send the JSON key; the service MUST NOT fail the sync because of it. The service MUST NOT persist `expectLeafTitle` to Redis pending, MUST NOT copy it onto delayed fire payloads, and MUST NOT append it in push alert amount suffixes. Amount expect copy MUST use formatted amount and optional unit only. expectLeafTitle MUST 全链路丢弃；alert 不拼 leaf。

#### Scenario: Legacy body with expectLeafTitle still succeeds

- **WHEN** 旧客户端 PUT pending 项含 `expectLeafTitle` 及其它合法字段
- **THEN** 服务 MUST 接受并调度，MUST NOT 因该字段失败

#### Scenario: Amount alert without leaf

- **WHEN** 延时开火拼 amount 预计后缀且 `expectAmount>0`
- **THEN** 后缀 MUST 含格式化数量与（若有）单位，MUST NOT 拼接叶子标题

### Requirement: Active-eta SHALL allow empty or zero expectEndAt

When accepting active-eta items, the service MUST treat omitted, null, or non-positive `expectEndAt` (including `0`) as no estimated end, and MUST still retain the item when `rootEventId` / `leafEventId` / `startAt` are otherwise valid. Positive `expectEndAt` MUST remain valid and MUST satisfy `expectEndAt >= startAt`. Downstream Live Activity / projections MUST tolerate missing end (omit ContentState key; MUST NOT write `0` as sentinel). active-eta MUST 允许无预计结束（含 0）；

#### Scenario: Merge-root active without end

- **WHEN** PUT active-eta 项含合法 root/leaf/start 且 `expectEndAt` 为 0 或省略
- **THEN** 服务 MUST 接受该进行中项且 MUST NOT 因无正结束而丢弃该项或拒绝整次 PUT

#### Scenario: Legacy positive end still accepted

- **WHEN** 旧客户端 PUT active-eta 项含正 `expectEndAt` 且 `>= startAt`
- **THEN** 服务 MUST 按既有语义保留该项

---

### Requirement: 系统 MUST 提供预测 Pro 独立喂养资格场景

cash-service EnsureSchema MUST 幂等种子 `feeding_eligibility_scene.scene_key=predict_pro_entry`，默认 `requiredDays=7`、`minRecordsPerDay=10`，MUST NOT 在重复启动时覆盖运维已改值（INSERT IGNORE 或等价）。该场景 MUST 独立于 `ucg_entry` 与 `care_alert_entry`。资格合成算法 MUST 与现有喂养资格引擎同构（自昨天起向前连续有效日，今日不计入窗口）。

#### Scenario: 首次部署含 predict_pro_entry

- **WHEN** 空库执行 EnsureSchema
- **THEN** 配置表 MUST 含 `predict_pro_entry` 默认 7/10 行

#### Scenario: 改 Pro 场景不影响 UCG

- **WHEN** 运维将 `predict_pro_entry.requiredDays` 改为 5
- **THEN** `ucg_entry.requiredDays` MUST 保持不变

### Requirement: App MUST 可查询预测 Pro 喂养资格且计入 usage

cash-service MUST 提供已登录 App API `GET /cash/app/api/predict-pro/eligibility`（路径精确以实现登记为准）。`device_no` MUST 取自网关注入内部头。响应 MUST 同构 UCG/值得留意资格字段：`qualified`、`requiredDays`、`effectiveDays`、`remainingDays`、可选 `message`。该接口 **MUST 计入** App API 使用统计：MUST NOT 写入 `usagestats/maintenance_skip.go`。MUST 具备 `g.Meta` summary 供 apiregistry。MUST NOT 加入 Bearer 匿名白名单。

#### Scenario: 未满 7 日返回进度

- **WHEN** 绑机用户请求 Pro eligibility 且连续有效日不足配置值
- **THEN** `qualified` MUST 为 false，且 `effectiveDays` / `remainingDays` / `requiredDays` MUST 反映进度

#### Scenario: 请求计入 usage

- **WHEN** 客户端调用 Pro eligibility
- **THEN** gateway-app usage 统计 MUST 可计到该 METHOD+path（维护排除列表 MUST NOT 含该项）

### Requirement: Pro 试用与付费与邀请开通 MUST 校验喂养资格

对 `predict_capability_pro` 的点试用、功能建单（创建订单）、邀请码兑换成功履约前，系统 MUST 确认当前请求设备在 `predict_pro_entry` 下 `qualified=true`；否则 MUST 拒绝且 MUST NOT 授予权益。基础版开通路径 MUST NOT 要求该资格。Pro 开通 MUST NOT 要求基础版已开通。

#### Scenario: 未合格拒绝 Pro 试用

- **WHEN** 用户点 Pro 试用但喂养资格未合格
- **THEN** 系统 MUST 拒绝，MUST NOT 标记试用 used，MUST NOT 写 Pro 权益

#### Scenario: 合格后可付费开通 Pro 且无需基础版

- **WHEN** 设备喂养合格且用户无基础版权益，支付 Pro SKU 成功
- **THEN** 系统 MUST 授予 Pro 人权益

#### Scenario: 基础版开通不查 Pro scene

- **WHEN** 用户开通基础版（支付/邀请/试用）
- **THEN** 系统 MUST NOT 因 `predict_pro_entry` 未合格而拒绝

### Requirement: Pro 商业开通 MUST 仍校验 predict_pro_entry 且不授予推送特权

对 `predict_capability_pro` 的点试用、建单、邀请兑换，系统 MUST 继续要求当前绑机在场景 `predict_pro_entry` 下 `qualified=true`。App `GET /cash/app/api/predict-pro/eligibility` MUST 保持字段同构并 **计入 usage**。通过开通获得的 Pro 权益 MUST NOT 被解释为 imminent 推送许可（推送见 `predict-push-ungate`）。

#### Scenario: 未合格仍拒 Pro 试用

- **WHEN** 用户点 Pro 试用但 `predict_pro_entry` 未合格
- **THEN** 系统 MUST 拒绝且 MUST NOT 写 Pro 权益

#### Scenario: 已开通 Pro 不单独放行推送

- **WHEN** 用户仅有 Pro 权益、非 VIP，且其它推送门禁与未开通用户相同
- **THEN** 推送路径 MUST NOT 额外因 Pro 放行或拒绝（与未开通用户同一套非商业规则）

---

### Requirement: App SHALL allow authors to soft-delete their own ideas

`POST /device/app/api/predict-pro-ideas/delete` MUST accept `{ "id": <ideaId> }` with a valid JWT. The server MUST soft-delete the idea only when it is non-deleted (`deleted_at IS NULL OR deleted_at = 0`) and the row’s `wx_id` equals the authenticated caller. Soft-delete MUST set `deleted_at` to a positive unix timestamp (same semantics as Admin soft-delete). After success, subsequent App `list` MUST NOT include that idea. Requests for another user’s idea, a missing id, or.

#### Scenario: Author deletes own idea

- **WHEN** 已登录用户对本人未删想法调用 delete
- **THEN** 系统 MUST 将该行 `deleted_at` 写为 >0
- **AND** 随后 list MUST 不再返回该 id

#### Scenario: Non-author rejected

- **WHEN** 已登录用户对他人未删想法调用 delete
- **THEN** 系统 MUST 返回业务错误且该行 MUST 保持未删

#### Scenario: Missing or already deleted

- **WHEN** id 不存在或 `deleted_at > 0`
- **THEN** 系统 MUST 返回业务错误且 MUST NOT 冒充成功

### Requirement: App predict-pro-ideas APIs SHALL require JWT but MUST NOT require Pro entitlement

All `/device/app/api/predict-pro-ideas/*` endpoints MUST require a valid access_token (JWT). The server SHALL derive `wx_id` from the token. Unauthenticated requests MUST return 401. Submit MUST require a bound `deviceNo`. The server MUST NOT call cash or read `predict_capability_pro` / `feature_user_entitlement` as a gate.

#### Scenario: Unauthenticated list

- **WHEN** 请求 `GET /device/app/api/predict-pro-ideas/list` 无有效 JWT
- **THEN** 系统 MUST 返回 401

#### Scenario: Non-Pro user may submit

- **WHEN** 已登录已绑定用户未持有预测 Pro 权益并提交合法 text
- **THEN** 系统 SHALL 仍允许插入（若未超日限额），MUST NOT 因缺少 Pro 而拒绝

### Requirement: App SHALL list non-deleted ideas newest-first with public fields

`GET /device/app/api/predict-pro-ideas/list` SHALL return non-deleted rows ordered by `created_at` descending. A row MUST be treated as non-deleted when `deleted_at IS NULL OR deleted_at = 0`； `deleted_at > 0` MUST be excluded. Each item MUST expose `id`, `text` (`display_text`), `babyName`, `likeCount`, `likedByMe`, and `isMine`. The response MUST NOT include `wxId` or `deviceNo`. `isMine` MUST be true iff the row’s `wx_id` equals the authenticated caller. `likedByMe` MUST.

#### Scenario: Empty list

- **WHEN** 无未删除想法（不含 `deleted_at=0`/`NULL` 行）
- **THEN** 系统 SHALL 返回空 list 且业务成功

#### Scenario: Rows with deleted_at zero appear

- **WHEN** 库中存在 `deleted_at = 0` 的想法且无其它过滤
- **THEN** `list` MUST 返回这些想法（`code=0` 且 `data.list` 非空）

#### Scenario: isMine for author

- **WHEN** 调用者是某想法的提交者
- **THEN** 该项 `isMine` MUST 为 true

### Requirement: App SHALL submit an idea with validated text and baby name snapshot

`POST /device/app/api/predict-pro-ideas/submit` MUST accept `{ "text": "<content>" }`. After trim, text MUST be non-empty and MUST NOT exceed 200 characters. On success the server SHALL insert a row with `content` and `display_text` equal to the trimmed text, `wx_id` from JWT, `device_no` from the bound session, and `baby_name` from the bound device user. Missing bind MUST fail without insert. 提交 **必须** 校验并快照宝宝名。

#### Scenario: Successful submit

- **WHEN** 已登录已绑定用户提交合法 text
- **THEN** 系统 SHALL 插入一行并返回新建 `id`（或含 id 的记录）

#### Scenario: Empty text rejected

- **WHEN** text trim 后为空
- **THEN** 系统 MUST 拒绝且 MUST NOT 插入

### Requirement: App submit SHALL rate-limit per wx_id by table daily COUNT

The server MUST enforce a per-`wx_id` submit quota of at most **10** ideas per **Asia/Shanghai calendar day**. The quota MUST be computed by `COUNT(*)` on `predict_pro_idea` for that `wx_id` whose `created_at` falls in the current Shanghai day (soft-deleted rows MUST still count toward the quota). The server MUST NOT introduce Redis (or other cache) solely for this quota. Excess submits MUST be rejected without inserting. 同一账号提交 **必须** 按表按日 COUNT 频控。

#### Scenario: Over quota

- **WHEN** 同 wx_id 在当日 COUNT 已达 10 后再提交
- **THEN** 系统 MUST 返回业务错误且 MUST NOT 插入

#### Scenario: Soft-deleted still consumes quota

- **WHEN** 用户当日已提交并被 Admin 软删后再次提交且当日 COUNT（含软删）已达上限
- **THEN** 系统 MUST 仍拒绝新插入

### Requirement: App SHALL like and unlike an idea idempotently

`POST /device/app/api/predict-pro-ideas/like` and `POST /device/app/api/predict-pro-ideas/unlike` MUST accept `{ "id": <ideaId> }` for a non-deleted idea（`deleted_at IS NULL OR deleted_at = 0`）. Like MUST create the unique like row if absent; unlike MUST remove it if present. Both MUST return current `likeCount` and `likedByMe`. Liking a missing/deleted idea（`deleted_at > 0` 或不存在）MUST fail with a business error. 点赞/取消 **必须** 幂等并返回热度；对 `deleted_at=0` 的想法 MUST 允许操作。

#### Scenario: Like then unlike

- **WHEN** 用户对未赞且未删（含 `deleted_at=0`）想法 like 再 unlike
- **THEN** 最终 `likedByMe` MUST 为 false 且 `likeCount` MUST 反映取消后的数量

### Requirement: Admin idea list and mutations SHALL use the same alive predicate

Admin 分页列表、改 `display_text`、软删前置加载 MUST 与 App 使用同一未删谓词：`deleted_at IS NULL OR deleted_at = 0`。Hub 与 App **不得** 因谓词不一致而一边有数据一边为空。

#### Scenario: Hub shows zero-deleted_at rows

- **WHEN** 库中存在 `deleted_at = 0` 的想法
- **THEN** Admin list MUST 返回这些行且 total 计入它们

---

### Requirement: predict_pro_idea tables SHALL persist ideas and likes with baby name snapshot

The device-service MUST create and maintain via idempotent EnsureSchema in `ai_voice_device`（test: `ai_voice_device_test`）:

1. These tables MUST NOT reuse `feedback`.

#### Scenario: EnsureSchema creates tables

- **WHEN** device-service 执行 EnsureSchema 且表不存在
- **THEN** 系统 SHALL 创建上述两表且可重复执行无错误，且新表 `deleted_at` 默认未删值为 `0`

#### Scenario: Soft-delete keeps row

- **WHEN** Admin 删除一条想法
- **THEN** 系统 SHALL 将 `deleted_at` 设为大于 0 的 unix 秒且 App/Admin list MUST 随后不可见该行

#### Scenario: Default zero is alive

- **WHEN** 新插入行的 `deleted_at` 为 `0`（或存量 `NULL`）
- **THEN** 该行 MUST 被视为未删除，且 MUST 出现在 App/Admin 未删列表中

---

### Requirement: 系统 MUST 仅提供预测能力 Pro 商业开通壳

cash-service MUST 种子并对人维护功能 `predict_capability_pro`（`activation_subject=user`，`unlock_methods` 含 `payment,invite_code`，MUST NOT 含 `ad`）。MUST 支持支付建单、邀请兑换、显式点试用（自点击起算 7 日写入人权益）。App catalog MUST 可返回该项及真实 `unlocked`/`status`。本变更 MUST NOT 要求任何业务路径（含推送）因持有 Pro 权益而改变行为。

#### Scenario: 合格用户可点试用开通 Pro

- **WHEN** 绑机用户喂养资格合格、试用未用、非 VIP、无有效 Pro 权益，调用点试用且 featureId 为 `predict_capability_pro`
- **THEN** 系统 MUST 写入约 7 日人权益且试用标为 used

#### Scenario: 开通 Pro 不改变推送资格

- **WHEN** 用户已开通 Pro 或未开通 Pro
- **THEN** imminent 推送是否发出 MUST NOT 仅因 Pro 权益有无而不同（由既有非商业门禁决定）

### Requirement: 系统 MUST NOT 再提供预测能力基础版开通

系统 MUST NOT 再种子或启用 `predict_capability_basic` 作为可售功能（Ensure 若遇已有行 MUST 将其 `status=0` 或等价停用，MUST NOT 继续 INSERT 启用态）。点试用、建单、邀请 MUST 拒绝以基础版为目标的新开通（若仍收到该 featureId）。

#### Scenario: Ensure 后基础版不可售

- **WHEN** 执行 EnsureSchema 且库中曾有基础版定义
- **THEN** 基础版 MUST 不为上架可售态（status≠1 或不再出现在可购 products）

#### Scenario: 点试用基础版被拒

- **WHEN** 客户端对 `predict_capability_basic` 调用点试用
- **THEN** 系统 MUST 拒绝

---

### Requirement: 预测 imminent 推送 MUST NOT 依赖功能开通或 VIP

向绑定用户发送普通可见 `predict_imminent` 推送时，系统 MUST NOT 查询或要求 `predict_capability_basic`、`predict_capability_pro` 人权益，MUST NOT 要求 isVip 作为发送前提。既有非商业门禁（pending、去重、历史窗口、限流等）MAY 继续生效。

#### Scenario: 未开通非 VIP 仍可收 imminent

- **WHEN** 目标 wx 无 Pro/基础版权益且非 VIP，imminent 触发且其它既有门禁通过
- **THEN** 系统 MUST 允许向该 wx 发送可见 imminent 推送

### Requirement: 统一 Live Activity 预测段 MUST NOT 因未开通而剥离

`RefreshBabyLiveActivity`（或等价统一投影）在扇出 upsert 时 MUST NOT 因缺少预测能力开通或 VIP 而过滤 `kind=imminent` 项。空投影时 end/清屏行为 MUST 保持。系统 MUST NOT 调用已删除的预测能力 access 合成作为扇出前提。

#### Scenario: 未开通用户仍可投影 imminent 段

- **WHEN** 无 Pro 权益且非 VIP 的用户设备有进窗预测 pending，触发统一 Refresh
- **THEN** ContentState 仍可含 imminent 项并 upsert（在开关开启等既有条件下）

---

### Requirement: Timing Live Activity end SHALL dismiss immediately and MUST NOT wipe startTime

When Go fans out a Timing Live Activity **end** (history closed by setting endTime, or the history row is deleted), the APNs liveactivity payload MUST include `event=end` and `dismissal-date` set to a unix timestamp in the **past** relative to send time, so the system removes the Lock Screen activity immediately. ContentState MUST include `phase` indicating ended and MUST NOT set `startTime` to 0 when a real start timestamp is known; if `eventName`/`startTime` are available.

#### Scenario: Close timing dismisses immediately

- **WHEN** Go 将一条 history 更新为已设置 endTime 并扇出 Timing end
- **THEN** APNs 体 MUST 含 `dismissal-date` 且其值 MUST 早于发送时刻
- **AND** ContentState MUST 含 `phase` 表示 ended
- **AND** 若该行 `startTime` > 0，ContentState 的 `startTime` MUST 等于该 unix 秒而非 0

#### Scenario: Delete history ends Timing

- **WHEN** Go 成功删除一条 history（含未闭合计时行）
- **THEN** Go MUST 向该 deviceNo 绑定用户扇出同一 activity id（`timing:{deviceNo}:{historyId}`）的 Timing Live Activity end
- **AND** 该 end MUST 同样携带过去的 `dismissal-date`

#### Scenario: End without dismissal-date is forbidden

- **WHEN** push-service 组装 Timing `event=end` 的 APNs JSON
- **THEN** `aps` MUST 包含键 `dismissal-date`
- **AND** 业务未传 `DismissalDate` 时组装层 MUST 填入过去 unix 秒，MUST NOT 省略该键

### Requirement: Timing Live Activity start Attributes SHALL carry eventName and startTime

When Go fans out a Timing Live Activity **start**, the fixed attributes object MUST include `eventName` (catalog display name when the history row name is empty but `eventId` is valid) and `startTime` as unix seconds when known. ContentState MUST continue to include the same fields. Timing **start** 固定 Attributes MUST 含 `eventName`（行内名为空且 eventId 有效时用事件主档名）及已知的 unix 秒 `startTime`。 ContentState MUST 继续含同名字段。

#### Scenario: Timing start attributes include eventName

- **WHEN** Go 因未闭合计时 history 落库成功而扇出 Timing start，且能解析出非空展示名
- **THEN** 固定 Attributes MUST 含键 `eventName` 且值为该展示名
- **AND** ContentState MUST 同样含该 `eventName`

#### Scenario: Empty row name uses catalog

- **WHEN** 该 history 的 `eventName` 为空且 `eventId` > 0，事件主档有标准名
- **THEN** Attributes 与 ContentState 的 `eventName` MUST 使用该主档名，MUST NOT 以空串作为唯一展示名

### Requirement: Timing Live Activity start payload SHALL include eventId and unix-second startTime without visuals

When Go fans out a Timing Live Activity **start**, the fixed ActivityKit attributes object MUST include `eventId` set from the persisted history row’s event id (same identity the Flutter event catalog uses). ContentState `startTime` MUST be a unix timestamp in **seconds** (not milliseconds). The attributes and ContentState MUST NOT include `logo`, `color`, `logoFile`, or remote image URL fields; Timing Live Activity **start** 扇出时，固定 Attributes MUST 含来自该 history 行的.

#### Scenario: Timing start carries eventId

- **WHEN** Go 因未闭合计时 history 落库成功而扇出 Timing start，且该行 `eventId` > 0
- **THEN** 推送固定 Attributes MUST 含键 `eventId` 且值等于该 history 的事件 id

#### Scenario: startTime is unix seconds

- **WHEN** Timing start ContentState 含 `startTime`
- **THEN** 其值 MUST 表示自 Unix epoch 起的秒数，MUST NOT 使用毫秒

#### Scenario: No logo or color in Timing push

- **WHEN** Go 构造 Timing Live Activity start 或 update 载荷
- **THEN** Attributes 与 ContentState MUST NOT 包含 `logo`、`color`、`logoFile` 或等价远程图字段

### Requirement: Server SHALL push Timing Live Activity start and end from history lifecycle

When the twin Go service successfully persists a new open timing history record (unset end time per existing open-history rules), it MUST fan out a Timing Live Activity start to bound users of that device/baby using a stable activity id that includes that history identity (e. When that record is closed (end time set), Go MUST fan out a Timing Live Activity end for the same activity id. Multiple open timings MUST be allowed to map to multiple concurrent Timing Live.

#### Scenario: Start on open timing

- **WHEN** Go 成功持久化一条未闭合计时 history
- **THEN** Go MUST 向该 deviceNo 绑定用户尝试 Timing Live Activity start

#### Scenario: End on close timing

- **WHEN** Go 将该 history 更新为已设置 endTime
- **THEN** Go MUST 向绑定用户发送对应 activity id 的 Timing Live Activity end（或等价 dismiss）

#### Scenario: Quick start-stop without delay gate

- **WHEN** 计时短时间开始后又结束且两端扇出均触发
- **THEN** 系统 MUST 通过 end 推送关闭对应锁屏组件，MUST NOT 依赖额外固定秒级「确认仍在计时」延迟作为 start 前置条件

#### Scenario: Multiple timings

- **WHEN** 同一宝宝下两条不同 historyId 的计时均未闭合且均已 start 扇出
- **THEN** Go MUST 使用不同 activity id，允许客户端同时展示多条 Timing Live Activity

#### Scenario: Single fanout per persist

- **WHEN** voice 经 HTTP 调用 history-service 成功落库一条未闭合计时
- **THEN** Timing Live Activity start MUST 仅因该次落库成功副作用扇出一次
- **AND** MUST NOT 因 voice 侧额外钩子再扇出第二次

### Requirement: Open timing history update SHALL fan out Timing Live Activity ContentState update

When Go successfully updates a history row that remains open (`endTime` unset), it MUST fan out a Timing Live Activity **update** to bound users using the same activity id as start (`timing:{deviceNo}:{historyId}`). The update payload MUST include ContentState `phase` of timing, `eventName` when a display name is available, and `startTime` as unix seconds when `startTime` > 0. MUST NOT set `startTime` to 0. MUST NOT end and restart the activity solely because `eventId`.

#### Scenario: Edit start time updates lock screen timer

- **WHEN** App 将一条未闭合 history 的 `startTime` 改为另一 unix 秒且更新成功
- **THEN** Go MUST 对该 `timing:{deviceNo}:{historyId}` 扇出 `action=update`
- **AND** ContentState `startTime` MUST 为新的 unix 秒

#### Scenario: Edit event name updates lock screen title

- **WHEN** App 更新未闭合 history 的展示名（或 eventId 对应主档名）且行仍未闭合
- **THEN** ContentState MUST 含非空 `eventName`（主档优先）
- **AND** MUST NOT 因此结束该 Timing 活动

#### Scenario: Close still only ends

- **WHEN** 同一次 Update 将 `endTime` 设为非 0
- **THEN** Go MUST 扇出 Timing end
- **AND** MUST NOT 再对该行发 update

### Requirement: Timing Live Activity end MUST delete activity token rows

When Go fans out a Timing Live Activity **end** (history closed by setting endTime, or the history row is deleted), after the end fanout attempt the push domain MUST delete persisted `token_kind=activity` rows for that Timing `activityId` (e. Push-to-Start tokens for Timing MUST NOT be deleted by this cleanup. This MUST use the same push-layer `action=end` cleanup as Imminent ends (MUST NOT require a separate history-only delete path). Timing Live Activity **end**（闭合写入.

#### Scenario: Close timing removes activity tokens

- **WHEN** Go 因 history 闭合对 `timing:{deviceNo}:{historyId}` 扇出 Timing `event=end`
- **THEN** push 域 MUST 删除该 `activityId` 下全部 `token_kind=activity` 行
- **AND** MUST NOT 删除该用户 Timing 的 Push-to-Start 行

#### Scenario: Delete open timing removes activity tokens

- **WHEN** Go 删除未闭合 history 并扇出 Timing end
- **THEN** 该 Timing `activityId` 的 activity token 行 MUST 同样被删除

#### Scenario: Timing update does not delete tokens

- **WHEN** Go 对未闭合 Timing 仅扇出 `action=update`
- **THEN** MUST NOT 因 update 删除该 `activityId` 的 activity token 行

### Requirement: Timing Live Activity ContentState SHALL support optional expectEndAt

0, the server MUST include it; when unknown, the server MUST omit the key (MUST NOT write `0` as a sentinel that clients treat as epoch). ContentState MUST NOT include a preformatted Chinese subtitle string for the expect-end line. Attributes and ContentState MUST still MUST NOT include `logo` / `color` / `logoFile`. 0 时 MUST 写入； 未知时 MUST 省略该键，MUST NOT 写 `0` 占位。 MUST NOT 写入成品中文「预计 HH:mm 结束」字符串。 Attributes/ContentState 仍 MUST NOT 含 logo/color。

#### Scenario: Update carries expectEndAt

- **WHEN** Go 对已有 Timing 活动扇出 update 且已知 `expectEndAt=1730000000`
- **THEN** ContentState MUST 含 `expectEndAt` 且值为该 unix 秒

#### Scenario: Unknown ETA omits expectEndAt

- **WHEN** Timing start 时尚无预计结束
- **THEN** ContentState MUST 省略 `expectEndAt`
- **AND** MUST 仍含既有 `eventName` / `startTime` / `phase` 契约字段

---

### Requirement: 预测开通数量 MUST 经功能列表暴露且不得下发事件 ID

对预测类功能，合成 catalog 对应项 MUST 暴露非负整数 `allowedCount`（无记录则为 0）。响应 MUST NOT 包含由服务端挑选的预测事件 ID 列表。客户端负责按本地预测列表顺序解锁前 N 项。系统 MUST NOT 将独立 `GET .../prediction/allowed-count` 作为客户端必调接口（可省略）。

#### Scenario: 列表项含 allowedCount 而非事件 ID

- **WHEN** 客户端请求合成功能目录且存在预测类功能
- **THEN** 该项 MUST 含 `allowedCount`，且 MUST NOT 含事件 ID 数组作为开通清单

#### Scenario: 无开通记录时为 0

- **WHEN** 设备从未获得预测开通数量增量
- **THEN** 该项 `allowedCount` MUST 为 `0`

### Requirement: Admin 对预测功能 MUST 只配置开通数量语义

与预测解锁相关的可配置授予物 MUST 表达为开通数量（增量），MUST NOT 要求运营配置具体事件 ID。

#### Scenario: 授予数量后列表可读到

- **WHEN** 经支付/邀请码/广告等路径为某 `device_no` 增加预测开通数量 N
- **THEN** 随后合成目录中该项 `allowedCount` MUST 反映累加后的权威值（无扣减前提下 ≥ 原值 + N）

### Requirement: allowedCount 权威在 MySQL 且热读经 Redis

`allowedCount` 权威 MUST 存于 MySQL（`feature_allowed_count`）。读路径 MUST 可经 Redis 缓存加速（产品已确认）；履约或 Admin 变更后 MUST 失效或写穿。Redis 访问 MUST 经 `cachekit` 与 platform 键 builder。

#### Scenario: 履约后数量可见

- **WHEN** 功能 SKU 支付成功且 `grant_kind` 为预测数量增量
- **THEN** 该 `device_no` 的 `allowedCount` MUST 增加对应 `grant_quantity`，且后续 catalog 读 MUST 看到更新后的值（缓存失效后）

---

### Requirement: 预测开通三通道每次永久 +1

对 `prediction_unlock`，系统在支付履约、邀请码兑换成功、广告完成开通时 MUST 各将设备永久 `allowed_count` 增量 +1（或 SKU/`grantQuantity` 为 1 的等价增量）。预测邀请兑换 MUST NOT 再写入临时/永久全开（`full_access` / catalog `allowedCount=-1`）。预测支付商品的 `grantQuantity` MUST 为 1。

#### Scenario: 支付一次

- **WHEN** 用户支付预测开通商品成功
- **THEN** 该设备永久预测条数 +1，且 MUST NOT 仅因本次支付变为全开哨兵

#### Scenario: 广告一次

- **WHEN** 用户完成预测广告开通
- **THEN** 该设备永久预测条数 +1

#### Scenario: 邀请一次

- **WHEN** 用户用好友邀请码兑换 `prediction_unlock`
- **THEN** 该设备永久预测条数 +1

### Requirement: 非预测邀请开通为权益一次

当非 `prediction_unlock` 的功能启用 `invite_code` 且用户兑换成功时，系统 MUST 授予 entitlement 一次（按功能定义），MUST NOT 默认按条数累加。

#### Scenario: 其它功能兑码

- **WHEN** 用户对支持邀请码的非预测功能兑码成功
- **THEN** 系统写入对应 entitlement，且该人×码×功能不可再兑

---

### Requirement: 预测有效条数 MUST 为默认加永久累加

对 `prediction_unlock`（及约定的预测数量类功能），合成 catalog 暴露的有效可看条数（非临时全开时）MUST 等于 `defaultFree + permanentDelta`，其中 `defaultFree` 来自功能定义可配置默认开通数（非负），`permanentDelta` 来自设备永久累加权威（付费/广告等，非邀请码临时全开）。无设备永久行时 `permanentDelta` MUST 为 0。系统 MUST NOT 在无临时全开时仅因缺行而固定返回 0 并忽略已配置的 `defaultFree`。

#### Scenario: 仅配置默认时返回默认条数

- **WHEN** 某设备无永久累加记录，且该功能 `defaultFree=3`，且无有效临时全开
- **THEN** catalog 该项 `allowedCount` MUST 为 `3`

#### Scenario: 默认加永久累加

- **WHEN** `defaultFree=3` 且该设备 `permanentDelta=5`，且无有效临时全开
- **THEN** catalog 该项 `allowedCount` MUST 为 `8`

### Requirement: 邀请码预测临时全开 MUST 使用哨兵 -1

当设备对预测功能存在有效临时全开（邀请码授予且未过期，或永久全开标志生效）时，catalog 该项 `allowedCount` MUST 为 `-1`（哨兵，表示全部可看）。有限期临时全开 MUST 暴露到期时间（`expiresAt` 为 unix 秒）。到期后 MUST 回落为「默认 + 永久累加」，且 MUST NOT 清除既有永久累加。邀请码兑换预测 MUST NOT 再将 `grant_quantity` 累加进永久 `allowed_count`。

#### Scenario: 期限内返回 -1

- **WHEN** 用户用邀请码成功开通 `prediction_unlock` 且授予有效天数 > 0，且当前时间未超过到期
- **THEN** catalog 该项 `allowedCount` MUST 为 `-1`，且 MUST 含对应 `expiresAt`

#### Scenario: 到期后回落

- **WHEN** 临时全开已过期，且 `defaultFree=2`、`permanentDelta=4`
- **THEN** catalog 该项 `allowedCount` MUST 为 `6`，且 MUST NOT 为 `-1`

#### Scenario: 邀请码不增加永久条数

- **WHEN** 邀请码兑换预测成功
- **THEN** 该设备永久累加权威值 MUST 与兑换前相同（仅更新临时全开状态）

### Requirement: 默认开通数 MUST 可经 Admin 配置

功能定义 MUST 支持配置预测类默认开通数；Admin 更新后，后续无临时全开的 catalog 读 MUST 反映新默认（缓存失效后）。功能编号仍为与客户端约定的稳定 ID，本要求不授权 Admin 修改 `featureId`。

#### Scenario: Admin 修改默认开通数

- **WHEN** 运维将 `prediction_unlock` 的默认开通数从 0 改为 2 并保存
- **THEN** 无永久累加且无临时全开的设备再次拉 catalog 时该项 `allowedCount` MUST 为 `2`

---

### Requirement: 系统 MUST NOT 再种子或特判 prediction_unlock

cash-service `EnsureSchema` MUST NOT 再插入或每次启动写回 `feature_def.feature_id=prediction_unlock`。履约、下单、Apple 通知、Admin 规则与 catalog MUST NOT 再为该 `feature_id` 走「预测条数 / `allowed_count_delta`」专用分支。若请求仍携带该 `feature_id`，系统 MUST 拒绝开通并记录日志，MUST NOT 把它当成普通权益授予。

#### Scenario: 启动不再写预测种子

- **WHEN** cash-service 执行 EnsureSchema
- **THEN** 系统 MUST NOT 执行针对 `prediction_unlock` 的 INSERT 或将其 `status` 写回的 UPDATE

#### Scenario: 旧请求被拒绝

- **WHEN** 支付履约或 Admin 授功能的 `featureId` 为 `prediction_unlock`
- **THEN** 系统 MUST 拒绝，MUST NOT 增加 `feature_allowed_count`

### Requirement: Admin MUST NOT 再提供预测条数表单项

`cash-feature-admin.html` MUST NOT 再展示「增加可看数量」、默认开通条数，或手工授里的授予数量字段。SKU 的 `grant_kind` 表单 MUST NOT 再提供 `allowed_count_delta`。

#### Scenario: 打开功能 SKU 表单

- **WHEN** 运维编辑任一上架功能的套餐
- **THEN** 表单 MUST NOT 出现 `allowed_count_delta` 选项

### Requirement: 删除预测代码 MUST NOT 删除历史表

实现 MUST NOT DROP `feature_allowed_count`，MUST NOT 删除已有 `feature_order` / `feature_product` 历史行。其它功能（值得留意、成长轨迹、邀请码、群二维码、VIP）的履约路径 MUST 保持可用。

#### Scenario: 历史表仍在

- **WHEN** 本变更部署完成
- **THEN** `feature_allowed_count` 表 MUST 仍然存在，且 care / growth 的支付与邀请开通 MUST 仍可成功

---

### Requirement: Live Activity APNs request body MUST be logged at Info before send

Before each successful attempt to HTTP POST a Live Activity payload to APNs, after the final JSON request body has been built, the push-service MUST emit an **Info**-level log that includes at least the Live Activity `event` (`start`|`update`|`end`), the device-token length (or equivalent redaction), and the full request body JSON string. The log MUST NOT include the full device token value. This logging MUST apply to all Live Activity sends that go through the shared APNs.

#### Scenario: Update path logs body at Info

- **WHEN** the system is about to send a Live Activity APNs request with event=update
- **THEN** an Info log MUST be written containing event=update and the full JSON body
- **AND** the log MUST NOT contain the complete device token string

#### Scenario: Start via push-to-start also logs body

- **WHEN** the system sends event=start (including after invalid-update restart)
- **THEN** an Info log MUST likewise include the full start request body before the APNs POST

### Requirement: Imminent Live Activity title MUST be composed as nick要事件了 in history Refresh

When the unified Live Activity Refresh projects a `kind=imminent` item into ContentState `items`, the `title` field MUST be a finished Chinese sentence of the form `{nick}要{eventName}了`, where `eventName` is the trimmed event display name from Redis imminent pending `title`, and `nick` is the device baby display name resolved via the existing history→device profile contract (e. When the baby name is empty or profile lookup fails, `nick` MUST fall back to `宝宝`. The.

#### Scenario: Composed title with baby name

- **WHEN** Refresh 投影 imminent 项且 pending.title=`吃奶`、设备宝宝名为`豆豆`
- **THEN** ContentState 该项 `title` MUST 为`豆豆要吃奶了`

#### Scenario: Fallback nick when profile empty

- **WHEN** Refresh 投影 imminent 项且 pending.title=`睡觉`、宝宝名为空或画像失败
- **THEN** ContentState 该项 `title` MUST 为`宝宝要睡觉了`

#### Scenario: Pending sync Refresh without ordinary push still composes

- **WHEN** App PUT imminent pending 成功触发 Refresh，且尚无对应 MQ 可见推送
- **THEN** 投影出的 imminent `title` MUST 仍为成品句（MUST NOT 因未走普通推送而只透传裸事件名）

#### Scenario: Empty event name omits title

- **WHEN** 进窗 pending 项 eventId/nextAt 合法但 title trim 后为空
- **THEN** 投影项 MUST 省略 `title` 键，MUST NOT 写出仅含 nick 的句子

### Requirement: Imminent Live Activity title MUST NOT embed expect amount or duration copy

When projecting imminent items, structured expect fields (`expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit`) MUST continue to be written as separate ContentState keys when usable (existing apply-pending-expect rules). The composed `title` MUST NOT append ordinary-push expect suffixes (including「，预计这次约…」or leaf titles). 投影 imminent 时 expect MUST 仍为独立 ContentState 键； 成品 `title` MUST NOT 追加普通推送的预计后缀或 leaf；

#### Scenario: Duration expect stays out of title

- **WHEN** pending 项 title=`吃奶`、nick=`豆豆`、expectKind=`duration`、expectDurationSec=`1800`
- **THEN** `title` MUST 为`豆豆要吃奶了`且 MUST NOT 含`预计`
- **AND** ContentState MUST 仍含 `expectKind=duration` 与 `expectDurationSec=1800`

#### Scenario: Amount expect stays out of title

- **WHEN** pending 项 title=`吃奶`、nick=`豆豆`、expectKind=`amount`、expectAmount=`120`、expectUnit=`ml`
- **THEN** `title` MUST 为`豆豆要吃奶了`
- **AND** ContentState MUST 仍含 amount 相关 expect 键

### Requirement: Live Activity APNs start MUST include aps.alert; update and end MUST omit it

The push-service MUST build ActivityKit Live Activity APNs bodies (`apns-push-type=liveactivity`) so that `event=start` includes `aps. For `event=start` (including first push-to-start and same-dispatch invalid-token restart start), `aps` MUST include an `alert` object with at least a non-empty title or body; if the internal op supplies empty alert title and body, the assembler MUST apply a non-empty default title/body so presentation is not dropped. For `event=update` and.

#### Scenario: Start includes alert

- **WHEN** push sends Live Activity `event=start` via push-to-start
- **THEN** the APNs request body `aps` MUST include `alert` with non-empty title or body
- **AND** MUST include `attributes-type` / `attributes` / `content-state` as today

#### Scenario: Update omits alert

- **WHEN** Refresh fans out a non-empty items list and push sends Live Activity `event=update`
- **THEN** the APNs request body `aps` MUST include `content-state` and MUST NOT include `alert`

#### Scenario: End omits alert

- **WHEN** Refresh fans out empty items and push sends Live Activity `event=end`
- **THEN** the APNs request body MUST omit `alert` and MUST keep existing end/`dismissal-date` semantics

#### Scenario: Start with empty caller alert still has default alert

- **WHEN** push sends `event=start` and the op alert title and body are both empty
- **THEN** the APNs `aps.alert` MUST still contain a non-empty default title or body

### Requirement: Live Activity Refresh SHALL supply alert copy for upsert fanout

When history-service Refresh schedules Live Activity fanout for a non-empty items list (`action=upsert` or equivalent), it MUST set companion `AlertTitle`/`AlertBody` (e. Empty items `end` fanout MUST remain without alert fields. MUST NOT change ContentState item projection rules solely due to alert copy. 非空 Refresh upsert MUST 填充伴随 title/body 供 start 使用；

#### Scenario: Non-empty Refresh supplies alert fields

- **WHEN** Refresh builds a non-empty items list and calls the push Live Activity fanout
- **THEN** the fanout op MUST carry non-empty alert title/body for potential start presentation
- **AND** ContentState `items` MUST still be built as before

### Requirement: System SHALL expose a single Refresh that projects open history and imminent pending into one Live Activity

The twin Go **history-service** MUST provide a shared Refresh operation (name MAY be `RefreshBabyLiveActivity`) keyed by `deviceNo` that: (1) loads the device’s current Redis imminent pending list via `cachekit`; Voice MUST invoke this Refresh only via `clients/history` (or an equivalent history internal HTTP contract). MUST NOT implement a second divergent assemble path in voice. MUST NOT import `services/voice` from history. history-service MUST 提供按 `deviceNo` 的统一.

#### Scenario: Refresh builds from both sources

- **WHEN** Refresh runs for a device that has two pending predict items and one open history
- **THEN** the Live Activity ContentState `items` MUST include those predict items and that open history item

#### Scenario: Voice calls history contract

- **WHEN** voice needs to refresh Live Activity after pending sync or MQ fire
- **THEN** voice MUST call the history Refresh contract and MUST NOT import history business packages

### Requirement: Live Activity ContentState SHALL be an ordered items list

ContentState MUST carry an `items` array (name MAY be `items`). Predict candidates MUST be taken from Redis imminent pending and MUST include only items that satisfy the lead window: `nextAt - leadSeconds ≤ now`, where `leadSeconds` MUST use the same configuration source and default as ordinary predict imminent push (`VOICE_PREDICT_IMMINENT_LEAD_SECONDS`, default 300 seconds). Pending items outside that window MUST NOT appear in `items` (they MAY remain in Redis for.

#### Scenario: Sort order within selected set

- **WHEN** lead-window pending has nextAt=100 and nextAt=50, and two open histories with startTime=10 and startTime=20, and the selected set includes all four under the cap
- **THEN** items MUST order predict(50), predict(100), timing(20), timing(10)

#### Scenario: Past nextAt still included while in pending

- **WHEN** Redis pending still contains an item whose nextAt is in the past (hence inside the lead window)
- **THEN** Refresh MUST still include that item until App PUT replaces or clears pending (MUST NOT drop solely because nextAt ≤ now)

#### Scenario: Far pending excluded from Live Activity

- **WHEN** Redis pending contains an item with `nextAt - leadSeconds > now`
- **THEN** Refresh MUST NOT include that item in `items`
- **AND** the item MAY remain in Redis pending for ordinary MQ scheduling

#### Scenario: Cap prefers in-progress over predict

- **WHEN** there are 2 open histories and 4 lead-window predict items
- **THEN** items MUST contain exactly 3 entries: the 2 timing items and the single soonest predict item
- **AND** emitted order MUST be that one predict, then the two timings (`startTime` desc)

#### Scenario: Three open histories fill the cap

- **WHEN** there are 3 or more open histories and any number of lead-window predicts
- **THEN** items MUST contain at most 3 timing items and MUST NOT include predict items

### Requirement: Empty items MUST end the Live Activity; otherwise update or start by activity token

After building `items`, for each bound wx: if `items` is empty, the system MUST fan out Live Activity `action=end` for that user’s unified activity and MUST clean persisted `token_kind=activity` rows for that user as on existing end cleanup. If `items` is non-empty and the user has at least one persisted `token_kind=activity` token, the system MUST `update` with the new ContentState. If `items` is non-empty and no activity token exists, the system MUST `start` via.

#### Scenario: Upsert when token exists

- **WHEN** Refresh builds a non-empty list and the wx has an activity token
- **THEN** push MUST send Live Activity update (not end+start)

#### Scenario: Start when no activity token

- **WHEN** Refresh builds a non-empty list and the wx has only push_to_start (or no activity token)
- **THEN** push MUST send Live Activity start via push_to_start

#### Scenario: End when nothing to show

- **WHEN** Refresh builds an empty items list
- **THEN** push MUST send Live Activity end for that wx

### Requirement: History CUD MUST delete imminent pending then Refresh Live Activity

On successful history create, update, or delete (including close-by-endTime side paths) for a device, the system MUST delete (or empty) that device’s Redis imminent pending authority key. The system MUST NOT invoke Live Activity Refresh or fan out Live Activity as a side effect of that history mutation. Delayed MQ consumers MUST continue to treat missing/empty pending as no-match for ordinary predict push. MUST NOT fan out dedicated Imminent-only end clear or per-history.

#### Scenario: CUD clears pending without LA Refresh

- **WHEN** history write succeeds while Redis pending was non-empty
- **THEN** pending MUST be cleared
- **AND** the system MUST NOT Refresh or fan out Live Activity solely due to that CUD

#### Scenario: Create does not project timing-only LA

- **WHEN** history create opens a timing row and pending is cleared in the same CUD side effect
- **THEN** the system MUST NOT Refresh Live Activity from that CUD
- **AND** Live Activity projection MUST wait for a later predict pending PUT and/or MQ fire Refresh

### Requirement: Pending sync and MQ fire MUST Refresh without replacing ordinary push

After a successful App full-replace write of imminent pending Redis, the system MUST schedule/invoke the unified Live Activity Refresh (subject to deviceNo trailing coalesce). When MQ imminent fire passes existing gates and proceeds to notify bound users, the system MUST keep sending ordinary `predict_imminent` visible push as today, and MUST also schedule/invoke Refresh for Live Activity (same coalesce). Fire MUST NOT end+start a separate Imminent Live Activity slot..

#### Scenario: PUT then Refresh

- **WHEN** client successfully PUTs an imminent pending list (empty or non-empty)
- **THEN** system MUST write Redis, schedule delayed fires as today when non-empty, and MUST schedule unified Live Activity Refresh

#### Scenario: Fire keeps alert and refreshes LA

- **WHEN** delayed fire matches pending and passes gates
- **THEN** system MUST attempt ordinary predict_imminent visible push AND MUST schedule Live Activity Refresh
- **AND** MUST NOT perform Imminent end-then-start slot replacement

### Requirement: Live Activity token registration SHALL not use attributesKind or activityId slots

`POST /app/api/push/live-activity/register` MUST accept `tokenKind` of `push_to_start` or `activity` without requiring `attributesKind` or `activityId` for slotting. Persistence MUST uniquely key by at least `(wx_id, device_key, token_kind)` (or equivalent). The `push_live_activity_token` schema MUST NOT retain business columns `attributes_kind` or `activity_id` (existing deployments MUST drop them; new installs MUST NOT create them). APNs Live Activity payloads MUST use a.

#### Scenario: Register activity token without attributesKind

- **WHEN** client registers tokenKind=activity with a token and no attributesKind/activityId slot fields
- **THEN** server MUST persist the activity token for that wx/device and later Refresh update MUST be able to use it

#### Scenario: Reject obsolete dual-kind requirement

- **WHEN** implementing the unified register contract
- **THEN** server MUST NOT require `attributesKind` in `imminent|timing` as a precondition to store tokens

#### Scenario: Schema has no attributes_kind column

- **WHEN** ensure/migrate runs for `push_live_activity_token`
- **THEN** the table MUST NOT expose `attributes_kind` or `activity_id` as required business columns
- **AND** upsert MUST NOT write those fields

### Requirement: Authenticated clear of Live Activity activity token per device

The system MUST expose `POST /app/api/push/live-activity/activity/clear` for App clients. The request MUST require Bearer authentication (gateway-app MUST NOT add this path to the Bearer auth exempt list). Request identity MUST come from the authenticated internal wx header (`X-Internal-Wx-Id`), not from an untrusted body `wxId`. The request body MUST include a non-empty `deviceKey`. On success, the system MUST delete the persisted `token_kind=activity` row for `(wx_id.

#### Scenario: Clear deletes activity token for device

- **WHEN** an authenticated user POSTs clear with `deviceKey=D` and an activity token row exists for that wx and `D`
- **THEN** that `token_kind=activity` row MUST be deleted
- **AND** any `token_kind=push_to_start` row for the same device MUST remain

#### Scenario: Clear is idempotent when no activity token

- **WHEN** an authenticated user POSTs clear with `deviceKey=D` and no activity token exists for that wx and `D`
- **THEN** the system MUST respond successfully without erroring solely due to absence of the row

#### Scenario: Clear does not start Live Activity

- **WHEN** clear successfully deletes (or finds no) activity token
- **THEN** the system MUST NOT send APNs `event=start` as part of handling that clear request

#### Scenario: Gateway requires Bearer for clear

- **WHEN** a client calls `POST /app/api/push/live-activity/activity/clear` without Bearer
- **THEN** gateway-app MUST NOT treat the path as auth-exempt
- **AND** the unauthenticated call MUST NOT reach successful clear semantics

### Requirement: Live Activity update ContentState MUST NOT include ack delivery fields

When the system sends Live Activity APNs with `event=update` (including upsert resolved to update), the ContentState MUST NOT include an `ack` delivery-receipt field, and the system MUST NOT persist `last_sent_ack` / `last_acked_ack` / `ack_sent_at` on `push_live_activity_token`. The table-level `updated_at` column MUST remain for register upsert and deduplication. The system MUST NOT schedule ack-timeout delete-and-restart after update send success. Existing invalid-token.

#### Scenario: Update payload has no ack field

- **WHEN** the system successfully prepares an APNs Live Activity update ContentState
- **THEN** the ContentState MUST NOT contain an `ack` key used for delivery receipt

#### Scenario: No ack timeout restart after update

- **WHEN** an update is sent successfully
- **THEN** the system MUST NOT start an in-process ack-timeout timer that deletes the activity token solely for missing client ack

### Requirement: Live Activity start MUST require at least one push_device for the wx

When the push-service resolves a Live Activity send to `event=start` (including `action=start` and `action=upsert` that selects push-to-start because no activity token exists), it MUST load that recipient `wxId`'s `push_device` rows (any channel). If there is no push device row for that wx, the system MUST skip the Live Activity start APNs send, MUST log a skip reason (e. `no_push_device`), and MUST NOT invent tokens. `event=update` and `event=end` MUST NOT be gated by this.

#### Scenario: Start skipped when no push_device

- **WHEN** Refresh/fanout resolves to Live Activity `event=start` for a wx that has push_to_start token(s) but zero `push_device` rows
- **THEN** the system MUST NOT send APNs Live Activity start for that wx
- **AND** MUST emit an observable skip log indicating missing push device

#### Scenario: Start allowed when any push_device exists

- **WHEN** the same start path runs for a wx that has at least one `push_device` row (any channel) and a push_to_start token
- **THEN** the system MUST proceed with the existing start send path (including start `aps.alert` rules from other requirements)

#### Scenario: Update not gated by push_device

- **WHEN** fanout resolves to Live Activity `event=update` for a wx with zero `push_device` rows but a valid activity token
- **THEN** the system MUST still attempt the update send per existing update rules (MUST NOT skip solely due to missing push_device)

### Requirement: Live Activity Refresh SHALL coalesce by deviceNo with trailing 300ms

All Live Activity Refresh requests that are allowed by the trigger matrix MUST be coalesced per `deviceNo` with a trailing debounce of **300 milliseconds**: each new schedule for the same device MUST reset the timer; when the timer fires, the system MUST run exactly one `RefreshBabyLiveActivity` (or equivalent) that rebuilds ContentState and fans out to bound wxIds. Within that single run, each bound wx MUST receive at most one Live Activity APNs attempt (excluding.

#### Scenario: Burst schedules one Refresh

- **WHEN** pending PUT and MQ fire both request Refresh for the same device within 300ms
- **THEN** the system MUST execute a single Refresh after the trailing window
- **AND** MUST NOT execute one Refresh per request in that burst

### Requirement: Active-eta sync MUST NOT trigger Live Activity Refresh

After a successful write or clear of the device’s predict active-eta Redis list, the system MUST NOT invoke or schedule Live Activity Refresh. Expect-end fields on in-progress items remain merged when Refresh runs for other allowed triggers (pending PUT / MQ fire).

active-eta PUT/清空 MUST NOT 触发 Live Activity Refresh。

#### Scenario: Active-eta write alone

- **WHEN** client successfully syncs active-eta for a device
- **THEN** Redis active-eta MUST be updated as today
- **AND** the system MUST NOT schedule Live Activity Refresh from that sync alone

---
