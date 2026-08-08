## ADDED Requirements

### Requirement: VIP 商品承载现价与原价

`ai_voice_cash.vip_product` MUST 提供现价字段 `price_fen` 与原价字段 `original_price_fen`（单位：分）。`original_price_fen=0` 表示无划线原价。改价 MUST 通过运维更新该表完成；系统 MUST NOT 要求 env 或 Admin 页面作为定价写路径。

#### Scenario: 原价为零

- **WHEN** 商品行 `original_price_fen` 为 0
- **THEN** 读价 API MUST 返回 `originalPriceFen=0`（客户端可不展示划线价）

#### Scenario: 原价与现价均可读

- **WHEN** 商品行 `price_fen=1900` 且 `original_price_fen=9900`
- **THEN** 读价 API MUST 分别返回对应分值

### Requirement: 匿名读取 VIP 商品价格

cash-service MUST 提供 `GET /cash/app/api/vip/product`，响应至少包含 `productCode`、`priceFen`、`originalPriceFen`、`durationDays`。该接口 MUST 允许未携带 Bearer / 未注入有效 `X-Internal-Wx-Id` 的调用方成功读取。gateway-app MUST 将该路径列入 Bearer 鉴权白名单（精确 GET）。支付宝等建单路径 MUST 仍要求登录，且建单金额 MUST 使用库内 `price_fen`。

#### Scenario: 未登录拉价成功

- **WHEN** 客户端不带 Authorization 请求 `GET /cash/app/api/vip/product`
- **THEN** gateway MUST 放行且 cash-service MUST 返回上架商品的现价与原价

#### Scenario: 建单仍用库现价

- **WHEN** 已登录用户创建支付宝 VIP 订单
- **THEN** 订单 `amount_fen` MUST 等于当时 `vip_product.price_fen`，MUST NOT 使用请求体中的任意标价覆盖
