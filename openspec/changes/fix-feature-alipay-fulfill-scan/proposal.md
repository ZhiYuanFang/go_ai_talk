## Why

功能单支付宝回调履约失败，日志为 `sql: no rows in result set`。VIP 正常。原因是功能订单按 `channel_txn_id` 做幂等查询时，用 `Scan` 扫到已初始化的 struct；首次回调交易号尚未写入，GoFrame 返回 `sql.ErrNoRows` 并被当成失败，订单其实按 `order_no` 是存在的。

## What Changes

- 功能订单查询（按 `order_no`、按 `channel`+`channel_txn_id`、按 `app_account_token`）对齐 VIP：空结果表示「没有」，MUST NOT 把 `sql.ErrNoRows` 原样返回给履约。
- `GetActiveFeatureProduct` / `GetFeatureProductByAppleID` 同类 `Scan` 空结果 MUST 返回明确的业务错误（商品不存在），MUST NOT 露出裸的 `sql.ErrNoRows`。
- 不改 App 支付字段、不改订单号前缀、不改支付宝验签与金额校验。

## Capabilities

### New Capabilities

- `feature-order-query`: 功能订单与功能 SKU 读路径的空结果语义，保证支付宝首次履约可继续按 order_no 处理。

### Modified Capabilities

- （无）不改 VIP 履约行为。

## Impact

- **进程**：`cash-service`。不新增表或配置。
- **代码**：主要在 `internal/services/cash/feature_order.go`。
- **不做**：不改功能订单号仍为 `VIP` 前缀（可另开变更）；不改 `maintenance_skip`；不加 Redis。
