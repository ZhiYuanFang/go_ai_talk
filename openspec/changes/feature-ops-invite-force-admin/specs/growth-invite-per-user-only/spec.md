## ADDED Requirements

### Requirement: 成长轨迹邀请按人一次且允许同机多码

对 `growth_trajectory_predict`，系统在邀请码兑换时 MUST NOT 因同一 `device_no` 已对本功能邀请开通而拒绝（MUST NOT 再适用 `InviteOncePerDevice` / MUST NOT 读写本功能的 `feature_invite_device_grant` 作为拒绝依据）。系统 MUST 继续：若同一 `redeemer_wx_id` 已对任意邀请码成功开通过本功能，则拒绝再次邀请开通；MUST 保留人×码×功能去重、自用禁兑与同宝宝禁兑。`care_alert_smart_remind` 的设备邀请一次约束 MUST 保持不变。

#### Scenario: 同宝宝第二人兑不同码

- **WHEN** 家长 A 已用码 X 邀请开通成长轨迹，家长 B（不同 wx、同一 `device_no`）使用另一好友码 Y 兑换成长轨迹且 B 从未邀请开通过本功能
- **THEN** 系统 MUST 允许兑换并写入 B 的 user 权益

#### Scenario: 同一人第二次邀请拒绝

- **WHEN** 同一 `redeemer_wx_id` 已对成长轨迹邀请开通成功后再次用任意码兑换本功能
- **THEN** 系统 MUST 拒绝

#### Scenario: 值得留意仍同机一次

- **WHEN** 某 `device_no` 已邀请开通 `care_alert_smart_remind` 后，同机另一账号用另一码再兑该功能
- **THEN** 系统 MUST 拒绝
