# VIP 商业闭环 · 配置清单

跨 **cash-service / gateway-app / 支付宝 / ASC / Flutter** 的运维配置总表。  
沙箱接口验收见 [cash-vip-sandbox.md](./cash-vip-sandbox.md)；Flutter 联调命令见兄弟仓 `flutter_ai_talk/app/README.md`「VIP 开通」。

## 0. 职责分工

| 侧 | 要配什么 |
|----|----------|
| **cash-service / gateway** | DB、反代、支付宝密钥、Apple 商品映射、Admin 口令 |
| **支付宝开放平台** | 应用、密钥、异步通知 URL |
| **App Store Connect** | IAP 商品 + 沙箱账号 |
| **MySQL `ai_voice_cash`** | 现价/原价（SQL，非 env） |
| **Flutter** | `API_BASE_URL`；装支付宝 App（Android）；ASC 沙箱（iOS）；**无私钥** |

变量模板：`manifest/docker/.env.example`（`CASH_*` 段）。

---

## 1. 基础设施（必配）

```text
# MySQL：库 ai_voice_cash
CASH_DB_LINK=mysql:user:password@tcp(mysql-host:3306)/ai_voice_cash

# cash-service
CASH_SERVICE_ADDR=:9807

# gateway-app（及需调 cash internal 的 voice 等）
CASH_SERVICE_URL=http://cash-service:9807
```

验收：cash `:9807/api.json` 可达；网关可反代 `/cash/app/api/*`、`/cash/admin/api/*`。

---

## 2. 支付宝（Android 闭环 · 服务端）

| 变量 | 说明 |
|------|------|
| `CASH_ALIPAY_APP_ID` | 开放平台应用 ID |
| `CASH_ALIPAY_PRIVATE_KEY` | 应用私钥（勿提交仓库） |
| `CASH_ALIPAY_PUBLIC_KEY` | 支付宝公钥（验签 notify） |
| `CASH_ALIPAY_NOTIFY_URL` | **公网** `https://<gateway>/cash/app/api/vip/alipay/notify` |

开放平台侧：签约 App 支付能力；notify 须能被支付宝 POST 到公网网关。

Flutter：**不配**支付宝密钥；`tobias` 的 `url_scheme: pangbaovip` 已在 `pubspec`；真机需安装支付宝。

---

## 3. Apple IAP（iOS 闭环）

| 变量 | 说明 |
|------|------|
| `CASH_APPLE_PRODUCT_ID` | ASC 商品 ID，映射 SKU `vip_monthly_19`；须与 `GET …/vip/product` 的 `appleProductId` 一致 |
| `CASH_APPLE_BUNDLE_ID` | App Bundle ID（验单用） |

ASC：创建 IAP（建议**消耗型**便于续期）；标价与 DB `price_fen` **人肉对齐**（服务不自动同步）。

联调：沙箱 Apple ID；`CASH_PAYMENT_DEV_BYPASS=1` **仅非生产**（缺 JWS 时旁路，生产禁止）。

---

## 4. 价格（DB，非 env）

价格仅通过 SQL 维护，无管理后台、无环境变量覆盖。

```sql
UPDATE vip_product
SET price_fen = 1900,           -- 现价（分）→ 建单/支付宝金额
    original_price_fen = 9900,  -- 划线；0=不展示
    updated_at = UNIX_TIMESTAMP()
WHERE product_code = 'vip_monthly_19';
```

建库/补列：`hack/ddl_cash_vip.sql`、`hack/ddl_cash_vip_original_price.sql`；或依赖 cash-service 启动时 `EnsureSchema`。

改现价后须人肉对齐 ASC 标价。

---

## 5. 运维 Hub「VIP 权益」（可选）

| 变量 | 说明 |
|------|------|
| `GATEWAY_APP_ADMIN_PASSWORD` | Hub 登录 / 网关注入 `X-Admin-Password` |
| `CASH_ADMIN_PASSWORD` | 可选；空则 cash-service 回退 Hub 口令 |

入口：App 网关 `/device/admin` →「VIP 权益」，或直达 `/device/admin/cash-vip-admin.html`。  
API：`GET /cash/admin/api/vip/entitlements`（分页；列表含已过期；激活金额=最近 `paid` 订单 `amount_fen`）。

---

## 6. Flutter（客户端）

| 项 | 说明 |
|----|------|
| `--dart-define=API_BASE_URL=https://<gateway>` | 指向已配好 cash 的网关 |
| 登录态 | `status` / `orders` / `apple/verify` 需 Bearer |
| Android | 真机 + 支付宝 App；发版前 `flutter build apk --release` |
| iOS | StoreKit + 沙箱账号；商品 ID = API `appleProductId` |
| Web | 不支持支付（提示用手机 App） |

客户端**不需要** `CASH_ALIPAY_*` / Apple 共享密钥。

```bash
cd flutter_ai_talk/app
flutter run -d android --dart-define=API_BASE_URL=https://你的网关
flutter run -d ios --dart-define=API_BASE_URL=https://你的网关
```

Debug 日志过滤：`[CashVip]`。

---

## 7. App API 路径（网关前缀）

| 方法 | 路径 | 鉴权 | 用途 |
|------|------|------|------|
| GET | `/cash/app/api/vip/product` | 可匿名 | 现价/原价/`appleProductId` |
| GET | `/cash/app/api/vip/status` | 登录 | `isVip` / `expireAt` |
| POST | `/cash/app/api/vip/orders` | 登录 | `channel=alipay\|apple_iap` |
| POST | `/cash/app/api/vip/alipay/notify` | 支付宝验签 | 履约 |
| POST | `/cash/app/api/vip/apple/verify` | 登录 | IAP 验单履约 |

匿名 product 不计入 App usage 统计（`maintenance_skip.go`）。

---

## 8. 联调验收最短路径

```
登录 App
  → 留意详情非 VIP 见「开通 VIP」
  → 购买页价格 = product.priceFen（划线=originalPriceFen）
Android: 支付宝成功 → notify → status.isVip=true，CTA 消失
iOS: StoreKit → apple/verify → status.isVip=true
Hub: 可见 wxId / 到期 / 最近 paid amount_fen
```

更细步骤见 [cash-vip-sandbox.md](./cash-vip-sandbox.md)。
