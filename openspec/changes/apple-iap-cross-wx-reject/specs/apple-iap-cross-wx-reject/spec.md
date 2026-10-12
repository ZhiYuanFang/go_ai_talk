## ADDED Requirements

### Requirement: Apple IAP 已履约交易号异 wx MUST 拒绝授权

当 `channel=apple_iap` 且存在同 `channel_txn_id`（Apple `transactionId`）的已 `paid` 订单时，cash-service 在 VIP（`vip_order`）、功能与小智 MCP（`feature_order`）履约路径上 MUST 比较该已履约订单的 `wx_id` 与本次履约目标订单的 `wx_id`：

- 二者相同：MUST 幂等成功，MUST NOT 重复叠加权益或重复授永久能力。
- 二者不同：MUST 拒绝履约，MUST NOT 将目标订单标为 `paid`，MUST NOT 向目标 `wx_id` 授予权益；错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」。

本规则 MUST 适用于 `FulfillPaid`、`FulfillFeaturePaid`、`FulfillXiaozhiMcpPaid`（含经 `VerifyAppleIAP` 与 `DispatchFulfillPaid` 的调用）。`channel` 非 `apple_iap` 时，本 Requirement MUST NOT 改变既有幂等语义。

#### Scenario: 换号复用同一 Apple transactionId

- **WHEN** 账号 A 的订单已因某 `transactionId` 处于 `paid`，账号 B 以新订单对该 `transactionId` 发起 Apple 履约（含 `POST /cash/app/api/vip/apple/verify`）
- **THEN** 系统 MUST 返回业务错误，文案为「一个 Apple 账号仅支持开通一个胖宝账号」，且 MUST NOT 为 B 写入权益或将 B 的订单标为 `paid`

#### Scenario: 同一 wx 重复 verify 或 ASN 后再 verify

- **WHEN** 同一 `wx_id` 对已履约的同一 `transactionId` 再次走 verify 或 ASN 履约
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加权益

#### Scenario: 非 Apple 渠道不受本规则影响

- **WHEN** `channel=alipay` 且同 `channel_txn_id` 已对应一笔 `paid` 订单再次回调
- **THEN** 系统 MUST 按既有支付宝幂等成功处理，MUST NOT 仅因本能力返回 Apple 专用文案
