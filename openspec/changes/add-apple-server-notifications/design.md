## Context

VIP / 功能 Apple IAP 已落地：建单、`POST /cash/app/api/vip/apple/verify`、与支付宝共用 `DispatchFulfillPaid`。但 verify 依赖客户端提交 JWS，且当前实现以解码 payload 为主、非苹果主动回调。支付宝已有 `POST …/alipay/notify` 同步验签履约模型。产品确认：**ASN V2 同步履约、对齐支付宝通道、不上 MQ**。

约束：cash 一库；gateway `/cash/app/api/*` 已反代；禁止新 ticker；AMQP 本变更不引入；App `api/v1` 仅新增字段/路径，不破坏既有字段；跨仓 Flutter 须配合 `appAccountToken`。

## Goals / Non-Goals

**Goals:**

- Apple 扣款后由 ASN 推送到 cash，验签后开通 VIP/功能（不依赖 App 存活）。
- 建单生成 UUID `appAccountToken`，Flutter 购买必带；通知反查订单。
- 成功类通知履约；`REFUND` 收回。
- `apple/verify` 保留为可选加速，与 ASN 幂等共用履约。

**Non-Goals:**

- RabbitMQ / outbox / 扫表对账。
- 完整订阅生命周期（升降级、宽限期 UI）；一期以消耗型/一次性开通为主。
- 改造支付宝路径。
- 强制删除 `apple/verify`。

## Decisions

### D1：同步 ASN，对齐支付宝，不上 MQ

- **选择**：`HandleAppleNotification` 在 HTTP 请求内验签 → 解析 → `Fulfill*` / 退款 → 对 Apple 返回 200。失败返回非 2xx 以触发苹果重试。
- **理由**：与 `HandleAlipayNotify` 同构；产品明确不要 MQ。
- **备选**：ASN→MQ→履约 — 已否决。

### D2：路径与鉴权

- **选择**：`POST /cash/app/api/vip/apple/notifications`；gateway `ExactPOST` 白名单（同 alipay notify）；cash 内验 `signedPayload`。
- **理由**：前缀已在 cash 反代内；匿名到达、服务端验签。
- **备选**：`/cash/internal/...` — Apple 公网达不到集群内网。

### D3：`appAccountToken` 必须是 UUID，不能直接用 orderNo

- **选择**：`vip_order` / `feature_order` 增加 `app_account_token CHAR(36)`（UNIQUE，可空以兼容历史行）。`channel=apple_iap` 建单时生成 UUID 写入并在建单响应返回 `appAccountToken`。ASN 用 token 查单再履约。
- **理由**：StoreKit 2 / Apple 文档要求 `appAccountToken` 为 UUID；现有 `orderNo`（如 `VIP…`）不合法。
- **备选**：把 orderNo 改成 UUID — 影响面大、日志习惯变。token 与 orderNo 并存更安全。

### D4：通知类型一期范围

| notificationType（及 subtype 按实现细化） | 行为 |
|------------------------------------------|------|
| 表明一笔新扣款成功（如 `ONE_TIME_CHARGE` / 含有效 transaction 的购买成功类） | 按 token 履约 paid |
| `REFUND` | 标记订单退款语义；VIP：缩短/清除 entitlement；功能：停用对应权益行（与既有 entitlement 模型对齐，实现选最小破坏） |
| 其它 | 记日志并 200（避免苹果无限重试），不履约 |

### D5：验签

- **选择**：按 Apple ASN V2 校验 JWS（Apple Root CA / 证书链），拒绝未通过签名的 payload。生产禁止 `CASH_PAYMENT_DEV_BYPASS` 跳过 ASN 验签。
- **理由**：权威开通路径必须密码学可信；与「不能依赖客户端」一致。
- **备选**：仅解码不验签 — 否决。

### D6：`apple/verify` 角色

- **选择**：保留；成功时走同一 `Fulfill*`，`channel_txn_id=transactionId` 幂等。文档标明 ASN 为权威；verify 用于即时 UX。
- **理由**：避免 Flutter 大改砍掉 verify；双路径安全靠幂等。
- **增强（本变更宜做）**：verify 侧至少校验 bundleId；有条件时同样验 JWS 签名（与 ASN 共用验签工具）。

### D7：usage

- **选择**：`POST …/apple/notifications` **不计入** App usage，加入 `maintenance_skip.go`（精确 METHOD+path）。
- **理由**：渠道机器回调，非用户功能使用；对齐支付宝 notify 运维口径。

### D8：Flutter / ASC

- 建单后 StoreKit `Product.purchase(options: Product.PurchaseOptions(appAccountToken: uuid))`（或等价 API）。
- ASC → App → Server Notifications：Production / Sandbox URL 均指向公网  
  `https://<gateway>/cash/app/api/vip/apple/notifications`。

## Risks / Trade-offs

- [ASN 延迟数秒] → verify 加速；最终以 ASN/幂等为准。
- [历史已支付但无 token 的订单] → 仅新单走 ASN 映射；旧单仍可靠 verify。
- [REFUND 与已叠加续期] → VIP 按「该笔 duration」回退或直接置过期，design 实现取可测的一种并在代码注释写明。
- [验签证书轮换] → 使用 Apple 公布的根证书集并允许配置更新路径。

## Migration Plan

1. 发版 cash（DDL token 列）+ gateway（白名单 + skip）。
2. ASC 配置通知 URL（先 Sandbox 验收）。
3. 发 Flutter：购买带 `appAccountToken`。
4. 回滚：关掉 ASC URL 或回退镜像；未带 token 的购买仍可走 verify（降级回客户端依赖，需运维知晓）。

## Open Questions

无。产品已确认：ASN 同步、对齐支付宝、不上 MQ；`appAccountToken` 用 UUID 字段（非原始 orderNo）。
