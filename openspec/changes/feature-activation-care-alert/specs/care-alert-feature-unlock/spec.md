## ADDED Requirements

### Requirement: 系统 MUST 提供值得留意智能提醒功能定义与设备维开通

cash-service MUST 种子（或等价确保存在）稳定功能编号 `care_alert_smart_remind`（标题面向「值得留意智能提醒」），启用开通方式 MUST 包含 `payment`、`invite_code`、`ad`。权益主体 MUST 为 `device_no`（全家共享）。付费履约 MUST 授予永久 entitlement（SKU `duration_days=0` 或等价）。邀请码与广告开通 MUST 按功能定义上的邀请/广告授予天数限时授予（种子默认 7，Admin 可改）。VIP 购买/续期 MUST NOT 写入该功能的 `feature_entitlement` 行。

#### Scenario: 付费永久开通

- **WHEN** 用户支付值得留意开通商品成功
- **THEN** 该 `device_no` 对应 `care_alert_smart_remind` entitlement 的 `expires_at` MUST 为 0（永久）或等价永久语义

#### Scenario: 邀请限时开通

- **WHEN** 用户用邀请码兑换 `care_alert_smart_remind` 且定义天数为 7
- **THEN** 该设备 entitlement MUST 为约 7 天有效，且 unlock_method 反映邀请码

#### Scenario: VIP 不写功能表

- **WHEN** 账号 VIP 履约成功
- **THEN** 系统 MUST NOT 仅为值得留意插入或更新 `feature_entitlement`

### Requirement: 值得留意邀请开通 MUST 按设备仅成功一次且原力记用户

对 `care_alert_smart_remind`，系统在邀请码兑换成功时 MUST 以 `(device_no, feature_id)` 约束该设备仅能成功邀请开通一次（任意邀请码、任意兑换账号）。重复兑码 MUST 拒绝且 MUST NOT 延长权益或再次发放获客原力。获客原力与兑换流水 MUST 仍按邀请码主人/兑换用户维度记录，MUST NOT 因设备主体而改记到「设备账号」。人×码×功能去重 MUST 继续生效。本约束 MUST NOT 应用于 `prediction_unlock`。

#### Scenario: 同设备第二次邀请拒绝

- **WHEN** 设备 D 已成功用任意码开通 `care_alert_smart_remind`，同设备另一账号再次用其它码兑换该功能
- **THEN** 系统 MUST 拒绝，MUST NOT 更新 entitlement，MUST NOT 再次给码主人加原力

#### Scenario: 预测不受设备一次限制

- **WHEN** 同一设备依次用两个不同好友码兑换 `prediction_unlock`（且同设备互兑规则通过）
- **THEN** 两次均可成功且永久条数各 +1

#### Scenario: 原力仍记用户

- **WHEN** 用户 U 兑码为设备 D 开通值得留意成功
- **THEN** 获客原力 MUST 记到码主人用户侧，兑换流水 MUST 含 U 的 `wx_id` 与 D 的 `device_no`

### Requirement: 目录与 VIP 覆盖语义 MUST 对齐预测槽模式

合成功能目录 MUST 返回 `care_alert_smart_remind` 的真实设备 `unlocked` / `expiresAt`，MUST NOT 因请求者 `isVip` 将 `unlocked` 改写为 true。业务使用态（含 access 合成）MUST 在「设备权益未过期」与「当前 isVip」之间按 OR 放行。UCG 与值得留意**喂养**资格 MUST 继续不受 isVip / 功能开通影响。

#### Scenario: VIP 使用态可看但 catalog 真实

- **WHEN** 用户 isVip=true 且设备无值得留意 entitlement
- **THEN** catalog 该项 `unlocked` MUST 为 false，但业务可看判定 MUST 为允许（在喂养合格前提下）

#### Scenario: 过期后非 VIP 不可用

- **WHEN** 邀请授予已过期且用户非 VIP
- **THEN** catalog `unlocked` MUST 为 false，业务可看判定中 featureActive MUST 为 false
