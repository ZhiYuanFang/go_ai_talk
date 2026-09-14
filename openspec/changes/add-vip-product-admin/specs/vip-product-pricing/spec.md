## MODIFIED Requirements

### Requirement: VIP 商品承载现价与原价

`ai_voice_cash.vip_product` MUST 提供现价字段 `price_fen` 与原价字段 `original_price_fen`（单位：分）。`original_price_fen=0` 表示无划线原价。改价 MUST 通过更新该表完成；写路径 MUST 为开通功能管理所调用的 VIP 商品 Admin API（及等价手工 SQL）。系统 MUST NOT 使用环境变量或 yaml 作为定价或 Apple 商品 ID 的真相源。

#### Scenario: 原价为零

- **WHEN** 商品行 `original_price_fen` 为 0
- **THEN** 读价 API MUST 返回 `originalPriceFen=0`（客户端可不展示划线价）

#### Scenario: 原价与现价均可读

- **WHEN** 商品行 `price_fen=1900` 且 `original_price_fen=9900`
- **THEN** 读价 API MUST 分别返回对应分值

#### Scenario: Admin 改价后读价一致

- **WHEN** 管理员经 VIP 商品 Admin API 将 `price_fen` 更新为新值
- **THEN** `GET /cash/app/api/vip/product` 与后续建单 `amount_fen` MUST 使用该新值
