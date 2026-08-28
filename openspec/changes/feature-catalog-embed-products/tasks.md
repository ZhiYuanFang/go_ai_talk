## 1. 契约与组装

- [x] 1.1 扩展 `api/v1` catalog 响应类型：项内 `products[]`（productCode/priceFen/originalPriceFen/durationDays/grantKind/grantQuantity/appleProductId）
- [x] 1.2 `GetFeatureCatalog`（或等价）按 feature 加载启用中 `feature_product` 填入 `products`；无则 `[]`
- [x] 1.3 Admin upsert `feature_product` / `feature_def` 后失效定义相关 Redis 键（与现 invalidate 对齐）

## 2. 自检

- [x] 2.1 路径仍须登录+绑机；不新增 Bearer exempt；catalog 仍不计入 usage
- [x] 2.2 中文注释；不新建 `*_test.go`；建单仍只信服务端 productCode
