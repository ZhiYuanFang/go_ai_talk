## Why

开通页可以改 VIP「有效天数」，但 cash-service 每次启动都会把 `vip_product.duration_days` 写回 30，改完的天数留不住。天数填 0 时，VIP 会被静默当成 30 天，功能权益则被写成永久（`expires_at=0`）。产品要求所有功能都不支持永久，天数必须是正整数。

## What Changes

- `EnsureSchema` 在 `vip_product` 已存在时 MUST NOT 再覆盖 `duration_days`（与现价一样只在首次插入写入 30）。
- VIP 商品、功能 SKU、功能定义上的邀请/广告授予天数，以及支付/邀请履约，`durationDays` MUST ≥1。小于 1 MUST 拒绝，MUST NOT 静默改成 30，MUST NOT 写成永久。
- Hub 表单去掉「0=永久」；天数输入最小为 1。
- **不**改 `vip_monthly_19`。**不**批量改写已有 `expires_at=0` 的历史权益行。

## Capabilities

### New Capabilities

- `cash-duration-days`: VIP 与功能的授予天数必须 ≥1；启动种子不得覆盖已配置的 VIP 天数。

### Modified Capabilities

- （无。`openspec/specs/` 下没有独立的 VIP 天数能力规格。）

## Impact

- **进程**：`cash-service`（`EnsureSchema`、VIP/功能 Admin 校验、`ExtendEntitlement` / 功能 Grant 履约）；`gateway-app` 静态页 `cash-feature-admin.html`。
- **库**：不改表结构。已有 `vip_product.duration_days` 在重启后保持运维值。
- **API**：不改 App 支付契约字段结构。Admin 保存天数 `<1` 时返回参数错误。
- **非目标**：商品编号改名、永久 VIP、把历史 `expires_at=0` 批量过期、启动时停止覆盖标题或上架状态。
