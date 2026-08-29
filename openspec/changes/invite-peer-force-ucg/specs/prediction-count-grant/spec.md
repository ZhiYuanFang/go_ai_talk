## ADDED Requirements

### Requirement: 预测开通三通道每次永久 +1

对 `prediction_unlock`，系统在支付履约、邀请码兑换成功、广告完成开通时 MUST 各将设备永久 `allowed_count` 增量 +1（或 SKU/`grantQuantity` 为 1 的等价增量）。预测邀请兑换 MUST NOT 再写入临时/永久全开（`full_access` / catalog `allowedCount=-1`）。预测支付商品的 `grantQuantity` MUST 为 1。

#### Scenario: 支付一次

- **WHEN** 用户支付预测开通商品成功
- **THEN** 该设备永久预测条数 +1，且 MUST NOT 仅因本次支付变为全开哨兵

#### Scenario: 广告一次

- **WHEN** 用户完成预测广告开通
- **THEN** 该设备永久预测条数 +1

#### Scenario: 邀请一次

- **WHEN** 用户用好友邀请码兑换 `prediction_unlock`
- **THEN** 该设备永久预测条数 +1

### Requirement: 非预测邀请开通为权益一次

当非 `prediction_unlock` 的功能启用 `invite_code` 且用户兑换成功时，系统 MUST 授予 entitlement 一次（按功能定义），MUST NOT 默认按条数累加。

#### Scenario: 其它功能兑码

- **WHEN** 用户对支持邀请码的非预测功能兑码成功
- **THEN** 系统写入对应 entitlement，且该人×码×功能不可再兑
