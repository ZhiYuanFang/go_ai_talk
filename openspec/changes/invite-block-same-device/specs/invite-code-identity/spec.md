## MODIFIED Requirements

### Requirement: 兑码去重为人×码×功能

系统兑换邀请码时 MUST 拒绝自用（码主人 `owner_wx_id` 与兑换者 `redeemer_wx_id` 相同）。系统 MUST 拒绝码主人**当前**绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空的兑换（同一宝宝下不同账号互兑）。系统 MUST 以 `(redeemer_wx_id, code, feature_id)` 唯一约束防止重复兑换。系统 MUST 允许同一兑换者使用不同 owner 的码（且双方设备号不同时）。系统 MUST NOT 再执行「一家锁定」（redeemer 绑定单一 owner）。系统 MUST 仅当目标 `feature_def` 启用且 `unlock_methods` 含 `invite_code` 时允许兑换。码主人设备号 MUST 经 device-service 契约按 `owner_wx_id` 查询，MUST NOT 由 cash 直查 device 库。当 device 查询失败时，系统 MUST fail-closed 拒绝本次兑换。当查询成功但主人为空 `device_no`（未绑机）时，系统 MUST NOT 仅因同设备规则拒绝。

#### Scenario: 多好友码累加预测

- **WHEN** 用户依次兑换好友 A、B 的码且 featureId 为 `prediction_unlock`，且兑换者与 A、B 当前绑定设备号均不同
- **THEN** 两次均成功且该设备预测永久条数各 +1

#### Scenario: 不同设备互兑

- **WHEN** A 兑 B 的码开通某功能且 B 兑 A 的码开通同一功能，且双方当前绑定 `device_no` 不同
- **THEN** 两次均成功

#### Scenario: 同设备拒绝兑码

- **WHEN** 兑换者当前 `device_no` 与码主人当前绑定 `device_no` 相同且均非空
- **THEN** 系统 MUST 拒绝兑换，MUST NOT 写入开通或获客原力

#### Scenario: 主人未绑机可兑

- **WHEN** 码主人经 device 查询得到空 `device_no`，且非自用、其它校验通过
- **THEN** 系统 MUST NOT 因同设备规则拒绝

#### Scenario: device 查询失败 fail-closed

- **WHEN** cash 调用 device 按主人 wxId 取 `device_no` 失败
- **THEN** 系统 MUST 拒绝本次兑换

#### Scenario: 同码同功能重复

- **WHEN** 同一用户对同一码同一 featureId 再次兑换
- **THEN** 系统拒绝
