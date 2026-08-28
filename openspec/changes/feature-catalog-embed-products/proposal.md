## Why

`commercial-feature-entitlement` 已提供合成 `GET /cash/app/api/feature/catalog`（含开通态）与 `POST .../feature/orders`（需 `productCode`），但 App 读模型未下发可售 SKU（`productCode` / `priceFen` / `durationDays` 等）。客户端无法展示价格与建单，除非硬编码 `productCode`（不可取）。本变更在 catalog 项内嵌套启用中的 `products[]`，一次读齐展示与支付入参。

## What Changes

- 扩展 App 合成目录：每个功能项增加 `products[]`（仅 `status=1` 的 `feature_product`）。
- 每项 SKU 至少含：`productCode`、`priceFen`、`originalPriceFen`、`durationDays`、`grantKind`、`grantQuantity`、可选 `appleProductId`。
- 若该功能 `unlockMethods` 不含 `payment`，`products` 可为空数组。
- **不**新增独立 App `GET .../feature/products`（方案 B 否决，保持单次 RTT）。
- usage：仍属查询链路，**不计入**统计（已有 catalog skip，无需改 denylist 语义）。
- Admin 写 SKU 后 MUST 失效与目录相关的定义/SKU 缓存。

## Capabilities

### New Capabilities

- `feature-catalog-sku-embed`：App 合成目录项嵌套可售 SKU，支撑支付展示与建单。

### Modified Capabilities

- （无）基线尚未归档 `feature-catalog-entitlement`；行为增量以本 capability 为准，与既有 commercial 变更对齐。

## Impact

- **进程**：`cash-service`（catalog 组装）；`api/v1` 响应类型。
- **客户端**：Flutter 用 `products[].productCode` 调建单，价格以服务端为准。
- **非目标**：匿名读价；独立 products App API；改支付回调或订单表。
