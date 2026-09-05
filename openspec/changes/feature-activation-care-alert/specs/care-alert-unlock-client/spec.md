## ADDED Requirements

### Requirement: Flutter 值得留意展示 MUST 区分喂养门槛与开通门槛

Flutter 值得留意卡片流 MUST 继续先请求 cash 喂养资格。当喂养 `qualified` 不为 true（含加载失败 fail-closed）时，客户端 MUST 展示进度或失败提示，MUST NOT 请求值得留意 daily，MUST NOT 仅因未开通跳转开通中心。当喂养合格且设备该功能未有效开通且账号非 VIP 时，客户端 MUST 展示可点击的开通引导态，用户点击后 MUST 进入功能开通中心（宜定位到 `care_alert_smart_remind`）。当喂养合格且（目录显示已开通未过期 **或** 当前 isVip）时，客户端 MUST 按既有逻辑请求值得留意 daily 并展示。

#### Scenario: 未喂养合格不进开通中心

- **WHEN** eligibility `qualified=false`
- **THEN** UI MUST 展示喂养进度类文案，点击 MUST NOT 作为开通中心的主路径要求

#### Scenario: 合格未开通点击进开通中心

- **WHEN** eligibility `qualified=true` 且非 VIP 且 catalog 中 `care_alert_smart_remind` 未有效开通
- **THEN** 用户点击值得留意引导区域 MUST 导航至开通中心

#### Scenario: VIP 或已开通拉 daily

- **WHEN** eligibility `qualified=true` 且（isVip 或功能已开通未过期）
- **THEN** 客户端 MUST 允许请求 daily 并展示结果

### Requirement: 开通中心 MUST 展示值得留意项且 VIP 覆盖与预测模式一致

功能目录拉取后，开通中心 MUST 展示 `care_alert_smart_remind`（功能启用时）。有效开通展示 MUST 使用「设备 `unlocked` 未过期 **或** isVip」的 OR 语义（与非预测项既有 `isFeatureEffectivelyUnlocked` 一致）。catalog 的 `unlocked` MUST 仍反映设备真实权益。客户端 MUST NOT 用 isVip 绕过喂养资格。

#### Scenario: 非 VIP 未开通显示可开通

- **WHEN** 非 VIP 用户打开开通中心且设备未开通值得留意
- **THEN** 该项 MUST 展示为未开通并可走支付/邀请（及若启用的广告）开通

#### Scenario: VIP 覆盖显示

- **WHEN** VIP 用户打开开通中心且设备无值得留意权益行
- **THEN** 有效开通态 MUST 视为可用，且 MUST NOT 要求客户端把 catalog `unlocked` 改写成服务端已开通
