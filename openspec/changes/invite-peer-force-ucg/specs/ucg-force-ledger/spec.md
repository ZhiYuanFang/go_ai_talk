## ADDED Requirements

### Requirement: 原力存储在 ucg 域

系统 MUST 在 ucg 数据库维护每用户原力当前值与积分流水。系统 MUST NOT 再以 device `wx.force_value` 作为原力权威。作者对自己辩论文章投票成功时，系统 MUST 原力 +1 并写入 reason=`debate_self_vote` 的流水。系统 MUST NOT 回填本变更上线前的历史投票流水。

#### Scenario: 辩论加分

- **WHEN** 作者自投成功
- **THEN** ucg 原力 +1 且 ledger 新增一条 `debate_self_vote`

### Requirement: 获客加原力

当邀请码被他人成功兑换一次时，系统 MUST 将码主人获客计数 +1，并经 cash→ucg 契约为码主人原力 +100，ledger reason=`invite_acquisition`。原力加分失败 MUST NOT 撤销已成功的功能开通；系统 MUST 记录可观测错误以便补偿。

#### Scenario: 兑码触发获客积分

- **WHEN** 用户 B 成功使用用户 A 的邀请码
- **THEN** A 的 `redeemed_count` +1，且 A 的原力尝试 +100 并在成功时写入 `invite_acquisition` 流水

### Requirement: App 可读原力与流水

系统 MUST 在 UCG App 可读路径提供当前 `forceValue`（可经 profile 或专用接口）与积分流水分页。系统 MUST NOT 要求返回「距下一档所需积分」；档位展示由客户端计算。

#### Scenario: 流水列表

- **WHEN** 用户请求 force ledger
- **THEN** 返回含 reason、delta、时间的记录，新产生的获客与辩论分均可见
