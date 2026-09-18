## 1. 手填表与合计

- [x] 1.1 `EnsureSchema` 创建 `cash_manual_ledger`（方向、名称、金额分、发生时间、创建/更新时间）
- [x] 1.2 手填新增、按 id 修改、按 id 物理删除、按发生时间倒序列表；方向非法、名称为空、金额 `<1` 拒绝
- [x] 1.3 收益合计：VIP 实付一笔；功能实付按功能名（空标题用功能编号，对不上商品用商品编码）拆开；手填按方向求和；差额 = VIP + 功能付费 + 进账 − 出账
- [x] 1.4 合计与列表共用可选 `start`/`end`：订单用 `paid_at`，手填用 `occurred_at`；只计 `paid` 且 `alipay`/`apple_iap`

## 2. Admin API

- [x] 2.1 新增 `GET /cash/admin/api/revenue/summary` 与 `GET /cash/admin/api/revenue/entries`
- [x] 2.2 新增 `POST` 创建、`POST .../update`、`POST .../delete`，均走现有 Admin 鉴权

## 3. Hub

- [x] 3.1 `cash-feature-admin.html` 顶栏增加「收益统计」，选中后只显示该面板
- [x] 3.2 展示 VIP、按功能名的付费、进账、出账、差额，以及手填增删改查；金额用分并注明标价不含抽成

## 4. 自检

- [x] 4.1 确认未改 App 支付字段、未改 `maintenance_skip`、未加 Redis 或后台循环
- [x] 4.2 `openspec validate cash-admin-revenue-stats --strict` 通过
