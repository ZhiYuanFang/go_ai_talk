## MODIFIED Requirements

### Requirement: 功能支付履约 MUST 授予 device_no 权益或递增 allowedCount

cash-service MUST 支持独立功能 SKU（`feature_product`）：建单与支付宝/Apple IAP 验签 MUST 与 VIP 支付栈一致（服务端验签）。支付成功后 MUST 按商品配置向订单关联的主体授予功能权益或递增 `allowedCount`（主体按功能 `activation_subject`：device 或 user）。功能订单 MUST 使用独立 `feature_order` 表。支付宝异步通知、Apple ASN 与 Apple 客户端验单入口 MUST 与 VIP **共用分流**（按订单表区分），MUST NOT 污染 VIP 订单语义。`channel=apple_iap` 的功能建单 MUST 生成并返回 UUID `appAccountToken`，供 StoreKit 与 ASN 映射。

#### Scenario: 功能 SKU 支付成功写权益

- **WHEN** 某功能商品支付回调/验单成功且 `grant_kind=entitlement`
- **THEN** 系统 MUST 幂等为订单主体写入/续期对应 `feature_id` 权益，`unlockMethod` MUST 为 `payment`

#### Scenario: 功能 SKU 支付成功增加 allowedCount

- **WHEN** 某功能商品支付成功且 `grant_kind` 为预测数量增量
- **THEN** 系统 MUST 幂等将该主体的 `allowedCount` 增加配置的 `grant_quantity`

#### Scenario: 共用回调按订单分流

- **WHEN** 支付宝通知或 Apple ASN 到达且订单号/token 属于 `feature_order`
- **THEN** 系统 MUST 执行功能履约且 MUST NOT 续期 `vip_entitlement`

#### Scenario: 重复回调不重复授予

- **WHEN** 同一渠道交易号或已 paid 订单再次通知
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误时长或数量

#### Scenario: 功能 Apple 建单返回 appAccountToken

- **WHEN** 用户以 `channel=apple_iap` 创建功能订单
- **THEN** 响应 MUST 含 UUID `appAccountToken` 且落库可被 ASN 反查
