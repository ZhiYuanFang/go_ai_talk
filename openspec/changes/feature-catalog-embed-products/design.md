## Context

商业功能开通已落地：catalog 返回功能定义 ⊕ 设备开通态；建单走 `feature_product.product_code`。Admin 可配置多 SKU，但 App 侧看不到，形成「能展示开通态、不能付钱」的契约缺口。产品选择方案 A：在 catalog 项内嵌 `products[]`，不新增独立 products 读接口。

## Goals / Non-Goals

**Goals:**

- catalog 每项附带该 `feature_id` 下启用中的可售 SKU 列表。
- 字段足以完成 UI 标价与 `POST /cash/app/api/feature/orders`。
- 写路径（Admin upsert product）失效相关 Redis 缓存。

**Non-Goals:**

- 独立 `GET /cash/app/api/feature/products`。
- Bearer 匿名可读价。
- 改变建单验价逻辑（仍以服务端 SKU 为准）。

## Decisions

### D1：嵌套在 catalog，不拆接口

- 与「进页一次读齐、直接展示」一致；绑机门槛已在 catalog。
- **备选 B**（独立 products）：否决（多余 RTT，且须再 skip usage）。

### D2：SKU 过滤与字段

- 仅 `feature_product.status=1` 且 `feature_id` 匹配当前目录项。
- 字段：`productCode`、`priceFen`、`originalPriceFen`、`durationDays`、`grantKind`、`grantQuantity`、`appleProductId`（可空串）。
- `unlockMethods` 不含 `payment` 时仍可返回空 `products`（或实现省略过滤，以空数组为准）。

### D3：缓存

- 功能定义缓存可继续；SKU 列表可与定义一并重建，或请求路径按 feature 查启用 SKU（字典小，JOIN 成本低）。
- Admin 写 `feature_product` / `feature_def` 时 MUST `Del` `CashFeatureDefCatalogKey`（及若存在的合成缓存键）。

### D4：usage / 鉴权

- 路径不变；继续登录+绑机；继续 maintenance_skip 中的 catalog 查询不统计。

## Risks / Trade-offs

- [一功能多 SKU 响应变大] → MVP SKU 少；无分页目录已接受小字典。
- [客户端误信客户端改价] → 建单金额仍服务端 SKU，与现 VIP 一致。

## Migration Plan

1. 扩展 catalog 响应与组装逻辑；更新 `api/v1` 类型。
2. 确认 Admin 写失效。
3. 回滚：去掉 `products` 字段（客户端需兼容缺省为空）。

## Open Questions

- 无。
