## MODIFIED Requirements

### Requirement: Apple IAP 验单开通权益

cash-service MUST 提供 `POST /cash/app/api/vip/apple/verify`（需登录）：接受 App 提交的 IAP 交易凭证（JWS 或团队选定的等价载荷）。该路径 MAY 用于支付后即时开通（加速 UX），但 **Apple 扣款后的权威开通来源 MUST 为 App Store Server Notifications**（见能力 `apple-server-notifications`）。verify 与 ASN 履约 MUST 共用幂等键（`channel=apple_iap` 与 `channel_txn_id=transactionId`），MUST NOT 因双路径重复叠加权益。校验通过且 `productId` 映射正确时，MUST 幂等将关联订单置 `paid`（或绑定已有订单）并续期 entitlement。本路径 MUST 使用 **Apple IAP**，MUST NOT 使用 Apple Pay（PassKit）作为数字会员开通方式。

#### Scenario: 合法 IAP 验单

- **WHEN** 已登录用户提交可验证的 IAP 交易且商品映射正确
- **THEN** 系统 MUST 开通或续期该 `wxId` 的月会员权益

#### Scenario: 重复提交同一 transaction

- **WHEN** 同一 Apple `transactionId`（或等价渠道交易号）再次 verify
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误的时长

#### Scenario: ASN 已履约后再 verify

- **WHEN** 同一 `transactionId` 已由 ASN 履约成功，用户再次调用 verify
- **THEN** 系统 MUST 幂等成功且 MUST NOT 再次延长权益

### Requirement: 创建 VIP 订单

cash-service MUST 提供 `POST /cash/app/api/vip/orders`：从 Header 读取 `X-Internal-Wx-Id`（>0），body 含 `productCode` 与 `channel`（`alipay` 或 `apple_iap`）。成功时 MUST 创建 `created` 状态订单并返回客户端调起支付所需参数（支付宝为调起串/参数；Apple 为 orderNo、预期 `appleProductId`，以及 UUID 格式的 `appAccountToken`）。金额 MUST 以服务端商品表为准。

#### Scenario: 支付宝建单

- **WHEN** 合法 `wxId` 以 `channel=alipay` 且 `productCode=vip_monthly_19` 建单
- **THEN** 系统 MUST 落库订单且状态为 `created`，并返回可用于调起支付宝的参数

#### Scenario: Apple IAP 建单

- **WHEN** 合法 `wxId` 以 `channel=apple_iap` 建单
- **THEN** 系统 MUST 落库订单且返回与 ASC 商品映射一致的 `appleProductId`（或等价字段）、订单号，以及非空 UUID `appAccountToken`

#### Scenario: 未登录建单

- **WHEN** 缺少有效 `X-Internal-Wx-Id`
- **THEN** 系统 MUST 拒绝建单
