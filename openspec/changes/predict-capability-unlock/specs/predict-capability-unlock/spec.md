## ADDED Requirements

### Requirement: cash-service MUST 种子预测能力基础版与 Pro 为人维功能

EnsureSchema MUST 幂等确保存在功能定义：

- `predict_capability_basic`：标题面向「预测能力基础版」；`activation_subject=user`；`unlock_methods` 为 `payment,invite_code`（MUST NOT 含 `ad`）；`status=1`；默认主色可配置。
- `predict_capability_pro`：标题面向「预测能力 Pro」；其余同上。

MUST 各种子一颗启用中的限时 entitlement SKU（如 30 天；`INSERT IGNORE` 不覆盖运维改价）。VIP 履约 MUST NOT 写入上述功能的权益行。`prediction_unlock` MUST 继续保持停用/拒绝语义。

#### Scenario: 首次 Ensure 有两档对人定义

- **WHEN** 空库或升级后执行 cash EnsureSchema
- **THEN** `feature_def` MUST 含上述两 `feature_id`，且 `activation_subject=user`，`unlock_methods` MUST NOT 含 `ad`

#### Scenario: VIP 购买不写预测权益行

- **WHEN** 用户购买或续期 VIP
- **THEN** 系统 MUST NOT 为 `predict_capability_basic` 或 `predict_capability_pro` 写入 `feature_user_entitlement`

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

### Requirement: VIP MUST 覆盖预测两档使用权且不改写 catalog unlocked

业务使用态（含 imminent 门禁合成）MUST 在「该 wx 基础版或 Pro 人权益未过期」与「当前 isVip」之间按 OR 放行。合成功能目录 MUST 返回两档真实 `unlocked` / `expiresAt`，MUST NOT 因请求者 `isVip` 将 `unlocked` 改写为 true。

#### Scenario: VIP 未购功能 catalog 仍未解锁

- **WHEN** VIP 用户无基础版/Pro 人权益请求 catalog
- **THEN** 两档 `unlocked` MUST 为 false，且业务侧仍可按 VIP 放行 imminent

### Requirement: Pro 能力 MUST 包含基础版推送语义

当账号仅有有效 `predict_capability_pro` 人权益（无基础版行）时，系统 MUST 仍视为具备基础版推送资格（imminent 可放行）。本变更 MUST NOT 要求实现其它 Pro 专属业务能力。

#### Scenario: 仅 Pro 可推 imminent

- **WHEN** 用户仅开通 Pro 且非 VIP，imminent 到达放行点
- **THEN** 系统 MUST 允许向该用户推送 imminent（在其它既有推送门禁通过时）
