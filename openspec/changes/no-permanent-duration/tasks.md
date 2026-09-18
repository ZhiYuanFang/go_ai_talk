## 1. VIP 种子与校验

- [x] 1.1 `EnsureSchema`：`vip_product` 已存在时不再更新 `duration_days`；首次插入仍为 30
- [x] 1.2 Admin 更新 VIP 商品：`durationDays < 1` 拒绝
- [x] 1.3 `ExtendEntitlement`：天数 `<1` 返回错误，不再替换成 30

## 2. 功能天数

- [x] 2.1 Admin 保存功能定义与 SKU：邀请天数、广告天数、SKU `duration_days` 均须 ≥1
- [x] 2.2 支付 / 邀请 / 广告履约：天数 `<1` 拒绝，不再把 `expires_at` 写成 0
- [x] 2.3 确认不批量改写已有 `expires_at=0` 行

## 3. Hub

- [x] 3.1 `cash-feature-admin.html`：VIP、SKU、邀请/广告天数 `min=1`，去掉「0=永久」文案

## 4. 自检

- [x] 4.1 确认未改 `vip_monthly_19`、未改 App 支付字段结构、未改 `maintenance_skip`
- [x] 4.2 `openspec validate no-permanent-duration --strict` 通过
