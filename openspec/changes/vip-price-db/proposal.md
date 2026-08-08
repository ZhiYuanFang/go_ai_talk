## Why

客户端需要展示 VIP **现价**与**原价（划线价）**，且落地页在未登录时也要能拉价。定价真相源为 `ai_voice_cash.vip_product`（手工 SQL 维护），不引入 env 控价或改价后台；Apple ASC 标价由运维自行与库内现价对齐，本变更不管理 ASC。

## What Changes

- `vip_product` 增加 `original_price_fen`（原价，分；`0` 表示不展示划线价）。
- 扩展 `GET /cash/app/api/vip/product`：返回 `priceFen` + `originalPriceFen`（及既有商品字段）；**允许未登录访问**。
- gateway-app：将该 GET 精确路径列入 Bearer 白名单；支付宝建单金额仍读库内 `price_fen`。
- 种子/DDL/EnsureSchema 同步原价默认值（可与现价相同或略高，实现时写明）。
- **不**做：改价 Admin 页、env 定价、退款、自动续签、ASC 同步。

## Capabilities

### New Capabilities

- `vip-product-pricing`：库表现价/原价语义、匿名读价 API、与建单金额一致性。

### Modified Capabilities

- （无）基线版本规格中尚无独立 VIP 定价 capability；增量规格引入即可。

## Impact

- **进程**：`cash-service`（读库、product API）；`gateway-app`（白名单）。
- **库**：`ai_voice_cash.vip_product`（`CASH_DB_LINK`）。
- **写价**：运维手工 SQL（无管理 UI）。
- **usage 统计**：放宽匿名前须问负责人；未答复不改 `maintenance_skip.go`。
