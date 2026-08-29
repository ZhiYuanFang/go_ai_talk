## ADDED Requirements

### Requirement: VIP 与永久单事件开通写路径分离

VIP 支付/续期履约 MUST 只更新 VIP 权益（含 `expireAt`），MUST NOT 修改 `feature_allowed_count` 或把 VIP 写成永久 feature entitlement 来「买断」预测条数。单事件支付/邀请/广告 MUST 只写入对应永久条数或 entitlement，MUST NOT 修改 VIP `expireAt`。

#### Scenario: 开通月卡

- **WHEN** 用户 VIP 履约成功
- **THEN** VIP 有效期延长且预测永久 `allowed_count` 不变

#### Scenario: 单次预测支付

- **WHEN** 用户支付预测 +1 成功
- **THEN** 永久条数 +1 且 VIP 状态不变

### Requirement: 读模型叠加使用权

对需开通的功能（含预测锁），业务使用态 MUST 在「永久已开通/条数内」与「当前 isVip」之间按 OR 放行。catalog 返回的预测 `allowedCount` MUST 反映永久合成值，MUST NOT 仅因 isVip 改为全开哨兵。UCG 入场资格 MUST 继续不受 isVip 影响。

#### Scenario: VIP 未买断条数

- **WHEN** 用户 isVip=true 且永久预测条数小于 totalActivatableCount
- **THEN** 使用态可查看预测事项，但 catalog `allowedCount` 仍为永久值
