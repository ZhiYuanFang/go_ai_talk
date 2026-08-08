# cash-service VIP 沙箱验收

跨环境变量 / 支付宝 / ASC / Flutter 的**配置总表**见 [vip-commercial-config.md](./vip-commercial-config.md)。

## 前置

1. MySQL 库 `ai_voice_cash` + `CASH_DB_LINK`
2. 启动 `cash-service`（`:9807`），确认 `/api.json` 可达
3. gateway-app 配置 `CASH_SERVICE_URL=http://cash-service:9807`
4. ASC 创建 IAP 商品，将 productId 写入 `CASH_APPLE_PRODUCT_ID`（映射 `vip_monthly_19`）
5. 支付宝开放平台配置应用，注入 `CASH_ALIPAY_APP_ID` / 私钥 / 公钥，`CASH_ALIPAY_NOTIFY_URL` 指向  
   `https://<gateway-app 公网>/cash/app/api/vip/alipay/notify`

## 支付宝

1. 登录 App，`POST /cash/app/api/vip/orders` `{ "productCode":"vip_monthly_19", "channel":"alipay" }`
2. 使用返回的 `alipayOrderStr` 调起 SDK
3. 支付成功后支付宝回调 notify；查 `GET /cash/app/api/vip/status` 应 `isVip=true`

## Apple IAP

1. `POST /cash/app/api/vip/orders` `{ "channel":"apple_iap" }` 拿到 `appleProductId` / `orderNo`
2. StoreKit 购买后 `POST /cash/app/api/vip/apple/verify` 提交 `transactionId`、`productId`、`signedTransaction`（JWS）与可选 `orderNo`
3. 查 status 应为 VIP；重复 verify 应幂等

## 开发旁路

仅非生产：`CASH_PAYMENT_DEV_BYPASS=1` 可跳过支付宝验签 / 允许无 JWS 的 Apple verify（仍校验商品与订单）。**生产禁止开启。**

## 手工改价（DB）

价格**仅**通过 SQL 维护，无管理后台、无环境变量覆盖。

- `price_fen`：现价（分），建单 / 支付宝金额用此字段  
- `original_price_fen`：划线原价（分）；`0` = App 不展示划线  
- App Store Connect 商品价格**不由本服务同步**；改现价后须人肉对齐 ASC 标价

```sql
-- 已有库补列（若 EnsureSchema 未跑）
-- SOURCE hack/ddl_cash_vip_original_price.sql;

-- 改现价为 29 元、划线原价 99 元
UPDATE vip_product
SET price_fen = 2900,
    original_price_fen = 9900,
    updated_at = UNIX_TIMESTAMP()
WHERE product_code = 'vip_monthly_19';
```

匿名可读：`GET /cash/app/api/vip/product`（gateway 白名单）；响应含 `priceFen`、`originalPriceFen`。建单 / 支付仍须登录。

## 运维 Hub「VIP 权益」

1. 配置 `GATEWAY_APP_ADMIN_PASSWORD`（及可选 `CASH_ADMIN_PASSWORD`）；gateway-app 与 cash-service 均需可读到校验口令。
2. 浏览器打开 App 网关 `/device/admin` 登录 → 导航「VIP 权益」→ `/device/admin/cash-vip-admin.html`。
3. 页内调用 `GET /cash/admin/api/vip/entitlements`（分页；可选 `wxId`）；列表含已过期行；激活金额=最近 `paid` 订单 `amount_fen`（分）。
4. 无 Admin JWT 时网关拒绝该 API；错误口令时 cash-service 拒绝。浏览器不得自带 `X-Admin-Password`。
