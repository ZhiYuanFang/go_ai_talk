# VIP 商业闭环 · 配置清单

跨 **cash-service / gateway-app / 支付宝 / ASC / Flutter** 的运维配置总表。  
沙箱接口验收见 [cash-vip-sandbox.md](./cash-vip-sandbox.md)；Flutter 联调命令见兄弟仓 `flutter_ai_talk/app/README.md`「VIP 开通」。

## 0. 职责分工

| 侧 | 要配什么 |
|----|----------|
| **cash-service / gateway** | DB、反代、支付宝密钥、Apple Bundle ID、Admin 口令 |
| **支付宝开放平台** | 应用、密钥、异步通知 URL |
| **App Store Connect** | IAP 商品 + 沙箱账号 |
| **MySQL `ai_voice_cash` / 开通功能管理** | VIP 与功能 SKU 的现价、原价、Apple 商品 ID |
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

| 项 | 说明 |
|----|------|
| `CASH_APPLE_BUNDLE_ID` | App Bundle ID（验 JWS `bundleId` / ASN） |
| VIP Apple 商品 ID | 开通功能管理 → **VIP 套餐**，写入 `vip_product.apple_product_id`（`vip_monthly_19`） |
| 功能 Apple 商品 ID | 开通功能管理 → 各功能售卖套餐的 Apple 商品 ID |
| **ASN URL（ASC 配置，非 env）** | Production + Sandbox 均指向公网 `https://<gateway>/cash/app/api/vip/apple/notifications` |
| `appAccountToken` | 建单 `channel=apple_iap` 返回 UUID；Flutter StoreKit **必须**传入；ASN 用此反查订单 |

ASC：每个价格各建一个 IAP（建议**消耗型**便于续期）；标价与库内 `price_fen` **人肉对齐**（服务不自动同步）。

**权威履约**：Apple Server Notifications V2（同步验签 → 履约，对齐支付宝 notify，**无 MQ**）。`POST …/apple/verify` 为可选加速，与 ASN 共用 `channel_txn_id` 幂等。

联调：沙箱 Apple ID；`CASH_PAYMENT_DEV_BYPASS=1` **仅非生产**（缺 JWS 时旁路 verify；**ASN 始终验签**，生产禁止开启 bypass）。

可选证书轮换：`CASH_APPLE_ROOT_CA_PEM`（默认内置 Apple Root CA - G3）。

---

## 4. 价格（开通功能管理，非 env）

VIP 现价/原价/Apple 商品 ID 在 **开通功能管理** 的「VIP 套餐」保存；功能 SKU 同页「售卖套餐」。仍可用 SQL 等价改库。

```sql
UPDATE vip_product
SET price_fen = 1900,           -- 现价（分）→ 建单/支付宝金额
    original_price_fen = 9900,  -- 划线；0=不展示
    apple_product_id = '你的ASC商品ID',
    updated_at = UNIX_TIMESTAMP()
WHERE product_code = 'vip_monthly_19';
```

建库：cash-service 启动时 `EnsureSchema`。改现价后须人肉对齐 ASC 标价。

---

## 5. 运维 Hub：开通功能管理 + VIP列表

| 变量 | 说明 |
|------|------|
| `GATEWAY_APP_ADMIN_PASSWORD` | Hub 登录 / 网关注入 `X-Admin-Password` |
| `CASH_ADMIN_PASSWORD` | 可选；空则 cash-service 回退 Hub 口令 |

入口：App 网关 `/device/admin` →「开通功能管理」→ 配置 VIP 套餐；按钮 **「VIP列表」** 打开只读权益页 `/device/admin/cash-vip-admin.html`（Hub 主导航不再展示「VIP 权益」）。  
API：`GET/POST /cash/admin/api/vip/product`；`GET /cash/admin/api/vip/entitlements`（分页；含已过期；激活金额=最近 `paid` 订单 `amount_fen`）。

---

## 6. Flutter（客户端）

| 项 | 说明 |
|----|------|
| `--dart-define=API_BASE_URL=https://<gateway>` | 指向已配好 cash 的网关 |
| 登录态 | `status` / `orders` / `apple/verify` 需 Bearer |
| Android | 真机 + 支付宝 App；发版前 `flutter build apk --release` |
| iOS | StoreKit：建单取 `appAccountToken`，`PurchaseParam.applicationUserName` 传入（须 UUID）；缺 token **不得**静默购买；商品 ID = API `appleProductId` |
| Web | 不支持支付（提示用手机 App） |

客户端**不需要** `CASH_ALIPAY_*` / Apple 共享密钥；ASN URL 只配在 ASC。

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
| POST | `/cash/app/api/vip/orders` | 登录 | `channel=alipay\|apple_iap`；Apple 返回 `appAccountToken` |
| POST | `/cash/app/api/vip/alipay/notify` | 支付宝验签 | 履约（对齐 ASN 同步模型） |
| POST | `/cash/app/api/vip/apple/notifications` | ASN 验签 | **权威**履约 / REFUND；gateway ExactPOST + maintenance_skip |
| POST | `/cash/app/api/vip/apple/verify` | 登录 | 可选加速验单（与 ASN 幂等） |

匿名 product、支付宝 notify、Apple ASN 不计入 App usage 统计（`maintenance_skip.go`）。

---

## 8. 联调验收最短路径

```
登录 App
  → 留意详情非 VIP 见「开通 VIP」
  → 购买页价格 = product.priceFen（划线=originalPriceFen）
Android: 支付宝成功 → notify → status.isVip=true，CTA 消失
iOS: StoreKit（带 appAccountToken）→ ASN 开通（可无 verify）→ status.isVip=true；verify 仅加速
Hub: 可见 wxId / 到期 / 最近 paid amount_fen
```

更细步骤见 [cash-vip-sandbox.md](./cash-vip-sandbox.md)。

## 9. AI 权益（VIP∪额度）

- 账号 VIP 时各 AI feature（voice/clinic/care_alert/polish）视为有额度且不计次。
- 非 VIP 走月度额度；用尽后使用 Admin「AI 模型与并发」中该 lane 的 free 模型（可空，Python 自选）。
- 硬件 /voice/chat/ws 与 MCP 文本对话按硬件特权 premium、不计次。
- care-alert 独立额度 `care_alert` 与 lane `careAlert`：月度默认/用户 override 在 **Voice 运维**（`voice-admin.html`）；模型/并发/free 在 **AI 模型与并发**（`ai-model-admin.html`）。详见 OpenSpec `vip-quota-joint-entitlement`。
