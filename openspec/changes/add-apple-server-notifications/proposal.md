## Why

当前 Apple IAP 开通只依赖客户端 `POST /cash/app/api/vip/apple/verify`：苹果侧扣款后**不会**主动通知本服务；App 断网、杀进程或未调 verify 时，会出现「已付款未开通」。产品要求开通权威与支付宝异步通知对齐，**不依赖客户端**，且不上 MQ。

## What Changes

- 新增 **App Store Server Notifications V2** 入口：`POST /cash/app/api/vip/apple/notifications`（匿名可达、验 `signedPayload`），同步调用既有 `DispatchFulfillPaid` / 功能履约（与 `alipay/notify` 同构），**不上 RabbitMQ**。
- Apple 建单（VIP / 功能）生成符合 StoreKit 要求的 **`appAccountToken`（UUID）** 落库并返回客户端；Flutter 购买时必须带上；ASN 用该 token 反查订单。
- 一期处理支付成功类通知与 **`REFUND`**（收回权益，与「付了就开」对称）。
- 保留 `apple/verify` 作可选加速路径，与 ASN **共用履约与 `(channel, channel_txn_id)` 幂等**；ASN 为权威开通来源。
- gateway-app：Bearer 白名单登记 ASN path；usage **不统计**（渠道回调，对齐支付宝 notify），写入 `maintenance_skip`。
- 跨仓 Flutter：IAP 购买传入 `appAccountToken`；runbook / ASC 配置 Production+Sandbox 通知 URL。

## Capabilities

### New Capabilities

- `apple-server-notifications`：ASN V2 接收、验签、token→订单映射、成功履约与退款收回、网关白名单与 ASC/env 约定。

### Modified Capabilities

- `vip-payment`：Apple 开通权威从「仅客户端 verify」改为「ASN 权威 + verify 可选加速」；建单返回 `appAccountToken`。
- `feature-unlock-fulfillment`：功能 Apple 支付同样走 ASN 共用入口与 `appAccountToken` 映射（与 VIP 共用回调分流）。

## Impact

- **进程**：`cash-service`（handler + 验签 + 履约）；`gateway-app-server`（反代已有 `/cash/app/api/*`、auth exempt、usage skip）。无新微服务；**无 MQ / 无新 ticker**。
- **库**：`ai_voice_cash.vip_order` / `feature_order` 增加 `app_account_token`（UUID，Apple 渠道必填）；EnsureSchema 迁移。
- **API**：新增匿名 POST notifications；建单响应增 `appAccountToken`（Apple 渠道）；不改既有 v1 其它字段语义（仅新增字段）。
- **配置**：ASC 通知 URL；验签所需 Apple 根证书/JWKS 策略见 design；沿用 `CASH_APPLE_BUNDLE_ID`。
- **跨仓**：`flutter_ai_talk` StoreKit 购买带 token。
- **usage**：负责人确认口径——渠道异步通知 **不统计**；tasks 登记 `maintenance_skip`。
- **非目标**：MQ/outbox；App Store Server API 定时对账；订阅续费复杂状态机（非消耗型可后置）；微信收款。
