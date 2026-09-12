## ADDED Requirements

### Requirement: 功能定义 MUST 声明开通主体 device 或 user

cash-service MUST 在 `feature_def` 上持久化 `activation_subject`，取值仅为 `device` 或 `user`，缺省 MUST 为 `device`。EnsureSchema MUST 幂等加列并对已有行回填 `device`。种子 `growth_trajectory_predict` MUST 为 `user`；`care_alert_smart_remind` 与 `prediction_unlock` MUST 为 `device`。

#### Scenario: 成长轨迹种子为人

- **WHEN** cash EnsureSchema 首次或幂等执行且写入/确认成长轨迹功能定义
- **THEN** 该功能 `activation_subject` MUST 为 `user`

#### Scenario: 值得留意仍为机

- **WHEN** 读取 `care_alert_smart_remind` 功能定义
- **THEN** `activation_subject` MUST 为 `device`

### Requirement: ActivateFeature MUST 按主体写入对应权益表

系统 MUST 支持 `ActivationSubjectDevice` 与 `ActivationSubjectUser`。当主体为 `device` 时 MUST 写入（或续期）`feature_entitlement`（按 `device_no`）。当主体为 `user` 时 MUST 写入（或续期）`feature_user_entitlement`（按 `wx_id`），MUST NOT 为该次履约写入成长轨迹的设备权益行。`prediction_unlock` 与 `allowed_count_delta` MUST 仅允许 device 主体。

#### Scenario: 用户主体支付履约

- **WHEN** 成长轨迹付费订单支付成功且功能定义为 `user`
- **THEN** 系统 MUST 以订单 `wx_id` 为键续期 `feature_user_entitlement`，时长取 SKU `duration_days`（种子 30）

#### Scenario: 拒绝预测写到账号维

- **WHEN** 调用 Activate 且 feature 为预测条数类且 SubjectType=user
- **THEN** 系统 MUST 拒绝且 MUST NOT 修改 `feature_allowed_count`

### Requirement: 支付邀请广告 MUST 按 activation_subject 选择 SubjectKey

支付履约、邀请码兑换、广告完成在授予非预测功能时 MUST 读取功能定义的 `activation_subject`：为 `user` 时 SubjectKey MUST 为触发账号 `wx_id`；为 `device` 时 MUST 为请求/订单中的 `device_no`。广告完成在主体为 `user` 时 MUST 要求有效 `wxId`，否则 MUST 拒绝。

#### Scenario: 邀请开通成长轨迹写人

- **WHEN** 用户兑码开通 `growth_trajectory_predict` 成功
- **THEN** 权益 MUST 落在兑换人 `wx_id` 的 user 权益表，且邀请天数取自 `feature_def.invite_duration_days`

### Requirement: 目录合成 MUST 按主体合并开通态

`GetFeatureCatalog`（或等价 App 目录）MUST 接受设备号，并在调用方提供登录 `wxId` 时：对 `activation_subject=device` 的功能用设备权益判断 `unlocked`；对 `user` 用该 `wxId` 的 user 权益判断。未提供有效 `wxId` 时，user 类功能 MUST 视为未解锁（`unlocked=false`），MUST NOT 误用设备权益冒充。

#### Scenario: 同设备另一账号目录未解锁

- **WHEN** 设备 D 上账号 A 已开通成长轨迹，账号 B（非 VIP）请求目录且绑定 D
- **THEN** 成长轨迹项对 B MUST 为未解锁
