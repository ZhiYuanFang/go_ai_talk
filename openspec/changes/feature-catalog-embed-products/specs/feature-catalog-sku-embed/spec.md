## ADDED Requirements

### Requirement: App 合成功能目录项 MUST 嵌套可售 products 列表

`GET /cash/app/api/feature/catalog` 返回的每一功能项 MUST 包含 `products` 数组（可为空）。数组元素 MUST 为该 `featureId` 下状态启用的功能 SKU，且至少含：`productCode`、`priceFen`、`originalPriceFen`、`durationDays`、`grantKind`、`grantQuantity`；MAY 含 `appleProductId`。客户端 MUST 能仅凭本接口完成支付展示，并用所选 `productCode` 调用功能建单接口，MUST NOT 依赖硬编码商品码。系统 MUST NOT 为此新增独立的 App `GET /cash/app/api/feature/products` 作为必调接口。

#### Scenario: 启用 SKU 出现在对应功能项下

- **WHEN** Admin 为 `prediction_unlock` 配置了启用中的功能 SKU，且客户端请求合成 catalog
- **THEN** 该项 `products` MUST 含该 SKU 的 `productCode` 与 `priceFen` 等字段

#### Scenario: 无 payment 或无启用 SKU 时 products 为空

- **WHEN** 某功能无任何启用中的 `feature_product`
- **THEN** 该项 `products` MUST 为空数组（或等价空列表），MUST NOT 省略导致客户端解析失败（实现可用 `[]`）

#### Scenario: 停用 SKU 不下发

- **WHEN** 某 SKU `status` 非启用
- **THEN** catalog 的 `products` MUST NOT 包含该 SKU

#### Scenario: 不新增独立 App products 接口为必调

- **WHEN** 客户端需要展示功能价格并建单
- **THEN** 合成 catalog MUST 足够，MUST NOT 要求再调独立 App products 列表接口
