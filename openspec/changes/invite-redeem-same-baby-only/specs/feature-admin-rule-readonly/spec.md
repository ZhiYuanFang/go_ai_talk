## MODIFIED Requirements

### Requirement: 功能定义展示只读开通规则

开通功能管理（Admin）在展示某一功能定义时 MUST 提供只读开通规则说明（服务端按 `feature_id` 派生或常量映射），涵盖：开放主体（对人/对机）、邀请开通约束与效果口径、付费主体口径，以及共用「同宝宝禁兑」说明（MUST NOT 再将「不可使用自己的邀请码」列为独立共用前置）。运维 MUST NOT 通过该字段编辑并改写真实闸门逻辑。

#### Scenario: 成长轨迹规则文案

- **WHEN** 运维查看 `growth_trajectory_predict` 的功能定义/详情
- **THEN** 只读规则 MUST 表明对人开通、同机可兑不同邀请码、每人邀请仅一次，且提及同宝宝账号互兑禁止，MUST NOT 再单独要求「不可使用自己的邀请码」作为共用前置句

#### Scenario: 值得留意规则文案

- **WHEN** 运维查看 `care_alert_smart_remind`
- **THEN** 只读规则 MUST 表明对机开通且同宝宝邀请开通仅一次，且共用说明含同宝宝禁兑、不含独立自用禁兑句

### Requirement: 同宝宝禁兑继续生效

系统兑换邀请码时 MUST 拒绝码主人当前绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空的兑换；MUST NOT 因此写入开通或获客原力。拒绝提示 MUST 明确不可使用同一宝宝下其他账号的邀请码（或同等语义）。系统 MUST NOT 再以 `owner_wx_id == redeemer_wx_id` 作为独立拒绝条件。

#### Scenario: 同宝宝拒兑

- **WHEN** 兑换者 `deviceNo` 与码主人当前绑定 `device_no` 相同且均非空
- **THEN** 系统 MUST 拒绝兑换

#### Scenario: 不再因同 wx 单独拒兑

- **WHEN** 兑换者与码主人 `wx_id` 相同，但双方当前 `device_no` 不同或主人未绑机，且其它校验通过
- **THEN** 系统 MUST NOT 仅因同 wx 拒绝兑换
