## ADDED Requirements

### Requirement: 用户邀请码由 cash Ensure 持有

系统 MUST 在 cash 库为每个有效 `owner_wx_id` 最多保留一条邀请码（一用户一码）。系统 MUST 在用户注册完成后触发 Ensure，或在 App 首次读取「我的邀请码」时懒创建。邀请码 MUST NOT 写入 device `wx` 表。邀请码 MUST NOT 要求有效期、码级可开功能范围或 grant 时长才能用于兑换。

#### Scenario: 首次读取生成

- **WHEN** 某 wx 尚无邀请码且请求我的邀请码
- **THEN** 系统创建唯一码并返回，且 `redeemedCount` 为 0

#### Scenario: 重复 Ensure 幂等

- **WHEN** 同一 wx 再次 Ensure
- **THEN** 系统返回同一码且 MUST NOT 插入第二行

### Requirement: 兑码去重为人×码×功能

系统兑换邀请码时 MUST 拒绝自用。系统 MUST 以 `(redeemer_wx_id, code, feature_id)` 唯一约束防止重复兑换。系统 MUST 允许同一兑换者使用不同 owner 的码。系统 MUST NOT 再执行「一家锁定」（redeemer 绑定单一 owner）。系统 MUST 仅当目标 `feature_def` 启用且 `unlock_methods` 含 `invite_code` 时允许兑换。

#### Scenario: 多好友码累加预测

- **WHEN** 用户依次兑换好友 A、B 的码且 featureId 为 `prediction_unlock`
- **THEN** 两次均成功且该设备预测永久条数各 +1

#### Scenario: 互兑

- **WHEN** A 兑 B 的码开通某功能且 B 兑 A 的码开通同一功能
- **THEN** 两次均成功

#### Scenario: 同码同功能重复

- **WHEN** 同一用户对同一码同一 featureId 再次兑换
- **THEN** 系统拒绝

### Requirement: App 可读我的码与邀请列表

系统 MUST 提供 App 接口返回当前用户邀请码与获客数量。系统 MUST 提供 App 接口返回成功使用该码的用户列表（含 UCG 展示昵称与兑换时间）。系统 MUST NOT 再提供面向运营的邀请码造码 Admin 页面。

#### Scenario: 邀请列表

- **WHEN** owner 查询 invitees
- **THEN** 每条含昵称与 `redeemedAt`，且仅含成功兑换记录
