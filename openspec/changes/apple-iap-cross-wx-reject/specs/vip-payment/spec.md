## MODIFIED Requirements

### Requirement: Apple IAP 验单开通权益

cash-service MUST 提供 `POST /cash/app/api/vip/apple/verify`（需登录）：接受 App 提交的 IAP 交易凭证（JWS 或团队选定的等价载荷）。该路径 MAY 用于支付后即时开通（加速 UX），但 **Apple 扣款后的权威开通来源 MUST 为 App Store Server Notifications**（见能力 `apple-server-notifications`）。verify 与 ASN 履约 MUST 共用幂等键（`channel=apple_iap` 与 `channel_txn_id=transactionId`），MUST NOT 因双路径重复叠加权益。校验通过且 `productId` 映射正确时，MUST 幂等将关联订单置 `paid`（或绑定已有订单）并续期 entitlement。本路径 MUST 使用 **Apple IAP**，MUST NOT 使用 Apple Pay（PassKit）作为数字会员开通方式。

当同一 `transactionId` 已对应一笔 `paid` 的 VIP 或功能/MCP 订单时：若该订单 `wx_id` 与本次履约目标订单 `wx_id` 相同，MUST 幂等成功；若不同，MUST 拒绝履约且错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」（见能力 `apple-iap-cross-wx-reject`），MUST NOT 向当前登录账号开通权益。

#### Scenario: 合法 IAP 验单

- **WHEN** 已登录用户提交可验证的 IAP 交易且商品映射正确
- **THEN** 系统 MUST 开通或续期该 `wxId` 的月会员权益

#### Scenario: 重复提交同一 transaction（同 wx）

- **WHEN** 同一 Apple `transactionId`（或等价渠道交易号）被**同一** `wx_id` 再次 verify
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误的时长

#### Scenario: ASN 已履约后再 verify（同 wx）

- **WHEN** 同一 `transactionId` 已由 ASN 履约为同一 `wx_id`，用户再次调用 verify
- **THEN** 系统 MUST 幂等成功且 MUST NOT 再次延长权益

#### Scenario: 异 wx 复用同一 transaction

- **WHEN** 某 `transactionId` 已为账号 A 履约成功，账号 B 登录后对该 `transactionId` 调用 verify
- **THEN** 系统 MUST 拒绝，错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」，MUST NOT 为 B 开通或续期 VIP
