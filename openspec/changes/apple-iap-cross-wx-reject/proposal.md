## Why

Apple 非消耗/已购商品在换 App 账号（`wx.id`）后再次「购买」时，StoreKit 常免扣款并返回同一 `transactionId`。服务端履约对已 `paid` 的 `channel_txn_id` 做幂等 `return nil`，异号也会接口成功，客户端易误报「开通成功」，但新账号不会落权益。需要在 JWS 校验通过、且拒绝向当前 `wx` 授权时，返回明确业务错误。

## What Changes

- Apple IAP 履约（VIP / 功能 / 小智 MCP）在 `channel=apple_iap` 且同一 `channel_txn_id`（`transactionId`）已对应一笔 `paid` 订单时：
  - **同 `wx_id`**：保持幂等成功（本人重试、ASN + verify 双路径）。
  - **异 `wx_id`**：拒绝履约，返回业务错误，文案为「一个 Apple 账号仅支持开通一个胖宝账号」。
- 覆盖 `FulfillPaid`、`FulfillFeaturePaid`、`FulfillXiaozhiMcpPaid`（及经 `DispatchFulfillPaid` / `VerifyAppleIAP` 的调用链）。
- 不改支付宝语义（预期不会出现同 txn 跨号）；不新增 `originalTransactionId` 绑定表；不改 Flutter（假定已展示接口 `message`）。
- 非 **BREAKING** API 形态：路径与请求体不变；异号场景由「空成功」变为业务错误（客户端若已 toast `message` 则行为更正确）。

## Capabilities

### New Capabilities

- `apple-iap-cross-wx-reject`：Apple IAP 渠道交易号跨 `wx.id` 拒绝授权的统一规则与错误文案。

### Modified Capabilities

- `vip-payment`：澄清「重复同一 transaction」仅对**同一 wx** 幂等成功；异号 MUST 拒绝。
- `feature-order-query`：澄清功能/MCP 订单「同交易号已 paid」在 Apple 异号时 MUST NOT 空成功。

## Impact

- **代码**：`internal/services/cash` 中 `FulfillPaid` / `FulfillFeaturePaid` / `FulfillXiaozhiMcpPaid`（建议抽公共判断）；`VerifyAppleIAP` / ASN / 支付宝分流间接受益。
- **API**：`POST /cash/app/api/vip/apple/verify` 及一切经上述 Fulfill 的 Apple 履约路径；响应错误 `message` 对用户可见。
- **客户端**：本仓无 Flutter；约定现网已展示业务 `message` 则无需改包。
- **数据**：无新表；沿用 `vip_order` / `feature_order` 的 `channel` + `channel_txn_id` + `wx_id`。
