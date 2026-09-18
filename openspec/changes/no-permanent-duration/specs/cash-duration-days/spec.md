## ADDED Requirements

### Requirement: 启动种子 MUST NOT 覆盖已有 VIP 天数

cash-service `EnsureSchema` 在 `vip_product` 行已存在时 MUST NOT 更新 `duration_days`。首次插入该商品时 MUST 仍写入默认 30。`product_code` MUST 保持 `vip_monthly_19`。

#### Scenario: 重启不改运维天数

- **WHEN** `vip_product.duration_days` 已被改为 7，且 cash-service 再次执行 EnsureSchema
- **THEN** 该列 MUST 仍为 7

#### Scenario: 空库首次种子

- **WHEN** 不存在 `vip_monthly_19` 且执行 EnsureSchema
- **THEN** 系统 MUST 插入 `duration_days=30` 的该商品行

### Requirement: VIP 天数 MUST 大于等于 1

Admin 更新 VIP 商品时 `durationDays` MUST ≥1，否则 MUST 拒绝且 MUST NOT 写库。支付履约调用续期时，若商品天数 `<1`，系统 MUST 拒绝续期，MUST NOT 把天数替换为 30，MUST NOT 把 `expire_at` 写成 0。

#### Scenario: 后台保存 0 天

- **WHEN** 管理员将 VIP 有效天数提交为 0
- **THEN** 系统 MUST 拒绝，且库中原天数 MUST 不变

#### Scenario: 履约遇到非正天数

- **WHEN** 一笔 VIP 订单支付成功，但商品 `duration_days` 小于 1
- **THEN** 系统 MUST NOT 延长 `vip_entitlement`，MUST NOT 把该次续期当成 30 天

### Requirement: 功能授予天数 MUST 大于等于 1

功能 SKU 的 `duration_days`、功能定义的邀请授予天数与广告授予天数，在 Admin 保存时 MUST 均 ≥1。支付履约、邀请码兑换、广告开通 MUST NOT 在天数为 0 时写入永久权益（`expires_at=0`）。天数 `<1` 时 MUST 拒绝开通。

#### Scenario: 保存功能套餐天数为 0

- **WHEN** 管理员把某功能 SKU 的有效天数设为 0
- **THEN** 系统 MUST 拒绝保存

#### Scenario: 邀请天数配置为 0 时兑码

- **WHEN** 功能定义的邀请授予天数为 0，用户兑换邀请码
- **THEN** 系统 MUST 拒绝开通，MUST NOT 把权益 `expires_at` 写成 0

### Requirement: Hub MUST NOT 再把 0 标成永久

`cash-feature-admin.html` 中 VIP 有效天数、功能 SKU 有效天数、邀请/广告授予天数的输入 MUST 要求至少 1。页面 MUST NOT 再显示「0=永久」。

#### Scenario: 打开 VIP 天数输入

- **WHEN** 运维打开 VIP 套餐面板
- **THEN** 有效天数控件 MUST 不允许小于 1，且 MUST NOT 出现「0=永久」说明

### Requirement: 历史永久行 MUST NOT 被本变更批量改写

本变更 MUST NOT 执行把已有 `expires_at=0` 批量改为当前时间的迁移。已存在的此类权益行 MUST 仍按变更前的读规则判断是否有效。

#### Scenario: 上线时已有永久权益

- **WHEN** 某账号功能权益在本变更部署前 `expires_at=0`
- **THEN** 部署本身 MUST NOT 把该行改写为已过期
