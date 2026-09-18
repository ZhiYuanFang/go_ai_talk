## ADDED Requirements

### Requirement: 功能订单按交易号查空 MUST 不阻断履约

`FulfillFeaturePaid` 在带有渠道交易号时 MUST 先查是否已有同渠道同交易号的已支付单。查无行时 MUST 视为尚未履约，MUST 继续按 `order_no` 加载订单并履约。MUST NOT 因该次空查询向调用方返回 `sql.ErrNoRows` 或导致支付宝回调失败。

#### Scenario: 首次支付宝回调功能单

- **WHEN** 功能订单状态为 `created`，支付宝以新的 `trade_no` 回调，且库中尚无该 `channel_txn_id`
- **THEN** 系统 MUST 将该单置为 `paid` 并完成功能开通，MUST NOT 因「按交易号查无」失败

#### Scenario: 同交易号重复回调

- **WHEN** 同一 `channel` 与 `channel_txn_id` 已对应一笔 `paid` 功能订单
- **THEN** 系统 MUST 幂等成功且 MUST NOT 重复开通

### Requirement: 功能订单与 SKU 读空 MUST 使用明确语义

按 `order_no` 查功能订单无行时，系统 MUST 返回明确的业务错误（订单不存在），MUST NOT 返回裸的 `sql.ErrNoRows`。按 `app_account_token` 查无行时 MUST 视为未找到（与 VIP 同语义）。按商品编码或 Apple 商品 ID 查启用 SKU 无行时，系统 MUST 返回明确的业务错误（商品不存在或未映射），MUST NOT 返回裸的 `sql.ErrNoRows`。

#### Scenario: 不存在的功能订单号

- **WHEN** 履约请求的 `order_no` 在 `feature_order` 中不存在
- **THEN** 错误信息 MUST 表明订单不存在，且 MUST NOT 仅为 `sql: no rows in result set`
