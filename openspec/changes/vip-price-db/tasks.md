## 1. 库表与种子

- [x] 1.1 为 `vip_product` 增加 `original_price_fen`（EnsureSchema ALTER + `hack/ddl_cash_vip_original_price.sql` 或更新 `ddl_cash_vip.sql`）
- [x] 1.2 种子/回填：`vip_monthly_19` 设置合理 `original_price_fen`（默认 9900；现价仍 1900）

## 2. 读价 API

- [x] 2.1 `GetActiveProduct` / `CashVipProductRes` 增加 `originalPriceFen`；`Product` handler **取消**强制 wxId
- [x] 2.2 确认建单路径仍只用 `price_fen`，不受 `original_price_fen` 影响

## 3. gateway 与文档

- [x] 3.1 `gateway_app_auth_exempt.go`：`GET /cash/app/api/vip/product` 加入白名单
- [x] 3.2 runbook / `cash-vip-sandbox.md`：手工 SQL 改价示例；注明 ASC 需人肉对齐现价
- [x] 3.3 usage：问负责人是否计入匿名 product；未答复不改 `maintenance_skip.go`
