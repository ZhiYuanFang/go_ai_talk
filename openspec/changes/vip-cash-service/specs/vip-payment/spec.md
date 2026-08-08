## ADDED Requirements

### Requirement: 一期单档月会员商品

cash-service MUST 提供一期唯一可售 VIP 商品：`product_code=vip_monthly_19`，标价 **1900 分人民币**，时长 **30 天**（续期按日叠加）。App MUST 能通过 `GET /cash/app/api/vip/product`（需登录，`X-Internal-Wx-Id>0`）读取该商品信息。

#### Scenario: 查询一期商品

- **WHEN** 已登录用户请求 VIP 商品接口
- **THEN** 响应 MUST 包含 `vip_monthly_19`、月费语义与 19 元标价（或等价分单位字段）

### Requirement: 创建 VIP 订单

cash-service MUST 提供 `POST /cash/app/api/vip/orders`：从 Header 读取 `X-Internal-Wx-Id`（>0），body 含 `productCode` 与 `channel`（`alipay` 或 `apple_iap`）。成功时 MUST 创建 `created` 状态订单并返回客户端调起支付所需参数（支付宝为调起串/参数；Apple 可为 orderNo + 预期 `appleProductId`）。金额 MUST 以服务端商品表为准。

#### Scenario: 支付宝建单

- **WHEN** 合法 `wxId` 以 `channel=alipay` 且 `productCode=vip_monthly_19` 建单
- **THEN** 系统 MUST 落库订单且状态为 `created`，并返回可用于调起支付宝的参数

#### Scenario: Apple IAP 建单

- **WHEN** 合法 `wxId` 以 `channel=apple_iap` 建单
- **THEN** 系统 MUST 落库订单且返回与 ASC 商品映射一致的 `appleProductId`（或等价字段）及订单号

#### Scenario: 未登录建单

- **WHEN** 缺少有效 `X-Internal-Wx-Id`
- **THEN** 系统 MUST 拒绝建单

### Requirement: 支付宝异步通知开通权益

cash-service MUST 提供 `POST /cash/app/api/vip/alipay/notify`：网关可匿名到达，但 MUST 校验支付宝签章与订单金额/商户订单号。校验通过后 MUST 将订单置为 `paid`（若尚未支付），并按商品时长续期该 `wx_id` 的 `vip_entitlement`（`new_expire = max(now, current_expire) + 30d`）。重复通知 MUST 幂等，MUST NOT 重复叠加错误时长。

#### Scenario: 首次合法 notify

- **WHEN** 支付宝对某 `created` 订单首次投递合法成功通知
- **THEN** 订单 MUST 变为 `paid`，且对应账号 VIP `expire_at` MUST 相对支付前至少延长 30 天

#### Scenario: 重复 notify

- **WHEN** 同一渠道交易对已 `paid` 订单再次通知
- **THEN** 系统 MUST 返回成功应答语义且 MUST NOT 再次延长权益（或仅保持幂等结果）

#### Scenario: 验签失败

- **WHEN** notify 签章非法
- **THEN** 系统 MUST NOT 将订单标为 paid，MUST NOT 开通权益

### Requirement: Apple IAP 验单开通权益

cash-service MUST 提供 `POST /cash/app/api/vip/apple/verify`（需登录）：接受 App 提交的 IAP 交易凭证（JWS 或团队选定的等价载荷），MUST 向 Apple 验真（或等价可信验签）。校验通过且 `productId` 映射到 `vip_monthly_19` 时，MUST 幂等将关联订单置 `paid`（或创建/绑定订单）并续期 entitlement。本路径 MUST 使用 **Apple IAP**，MUST NOT 使用 Apple Pay（PassKit）作为数字会员开通方式。

#### Scenario: 合法 IAP 验单

- **WHEN** 已登录用户提交可验证的 IAP 交易且商品映射正确
- **THEN** 系统 MUST 开通或续期该 `wxId` 的月会员权益

#### Scenario: 重复提交同一 transaction

- **WHEN** 同一 Apple `transactionId`（或等价渠道交易号）再次 verify
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误的时长

### Requirement: 查询当前 VIP 状态

cash-service MUST 提供 `GET /cash/app/api/vip/status`（需登录），返回当前 Header `wxId` 的 `isVip` 与 `expireAt`（无权益时 `isVip=false`）。

#### Scenario: 会员中查询状态

- **WHEN** 权益未过期的用户请求 status
- **THEN** 响应 MUST 含 `isVip=true` 与未来的 `expireAt`
