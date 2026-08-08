## Context

`care-alert-vip-by-wx` 已把 care-alert 强制 `wxId`、触发者选模与「device 读 `wx.is_vip`」落地，但 VIP 属资金域，应独立服务与独立库。本变更合并：**新建 cash-service**、**权益+支付开通**、**care-alert 改挂**，并 **同变更拆除** `wx.is_vip` / device VIP API。

约束：一服务一库；跨域禁止直查；Redis KV/PubSub 经 platform；App 新路由须 gateway 反代 + 白名单自检；支付回调可匿名但必须验签；VIP 读失败对 care-alert 降级 Zhipu。

## Goals / Non-Goals

**Goals:**

- `cash-service` + `ai_voice_cash` 可部署；内部可按 `wxId` 答 `isVip`。
- 一期：单档 ¥19/月；支付宝 + Apple IAP 开通/续期 entitlement。
- voice care-alert 只问 cash；拆除 device/`wx` VIP 路径。
- App 路径前缀 `/cash/app/api/...`。

**Non-Goals:**

- 微信收款、多档 SKU、Apple Pay（PassKit）。
- 完整退款关权益自动化、财务对账中台、发票。
- VIP 结果 Redis 缓存（一期直读 DB；负责人未确认不加新读缓存）。
- 修改 clinic 配额语义。

## Decisions

### D1：进程与库

- 进程名：`cash-service`（`cmd/cash-service`，镜像/compose 对齐现有 `*-service`）。
- 库：`ai_voice_cash`；GoFrame 组 `default`；env **`CASH_DB_LINK`** → `GF_DATABASE_DEFAULT_LINK`。
- 监听：`:9807`（`CASH_SERVICE_ADDR`；`:9806` 已由 notify-service 占用）；他服通过 **`CASH_SERVICE_URL`** 访问。
- 业务代码：`internal/services/cash/**`；注册：`RegisterCashServiceHTTP`。
- 配置：`manifest/config/config.cash-service.yaml`（仅本域 DB/支付占位，禁止回流主 `config.yaml`）。

### D2：表模型（最少三张）

```
vip_product
  product_code PK          -- vip_monthly_19
  title, price_fen         -- 1900
  duration_days            -- 30
  apple_product_id         -- ASC 商品 ID（配置/种子）
  status                   -- active

vip_order
  id, order_no UK
  wx_id, product_code
  channel                  -- alipay | apple_iap
  amount_fen, currency
  status                   -- created | paid | failed | closed
  channel_txn_id           -- 渠道侧交易号（可空至支付成功）
  UNIQUE(channel, channel_txn_id)  -- 支付成功后幂等（NULL 不参与时可应用部分唯一策略或支付成功后再写入）
  created_at, paid_at

vip_entitlement
  wx_id PK/UK
  expire_at                -- unix 秒
  updated_at
```

- **isVip**：`expire_at > now`（上海或 UTC 统一用 unix 秒比较即可）。
- **续期**：`new_expire = max(now, current_expire) + duration_days`。
- 一期种子一行 `vip_monthly_19`（EnsureSchema 或 DDL+seed SQL）。

### D3：HTTP 契约（前缀已定）

**App（Bearer，`X-Internal-Wx-Id`）：**

| Method | Path | 说明 |
|--------|------|------|
| GET | `/cash/app/api/vip/product` | 返回一期 SKU（19 元/月） |
| POST | `/cash/app/api/vip/orders` | body: `productCode`, `channel`∈`alipay\|apple_iap` → 建单并返回调起参数 |
| GET | `/cash/app/api/vip/status` | 当前账号 `isVip` / `expireAt` |
| POST | `/cash/app/api/vip/apple/verify` | App 提交 IAP 交易凭证（JWS 优先）→ 验单开通 |
| POST | `/cash/app/api/vip/alipay/notify` | 支付宝异步通知（**Bearer 白名单**，验签） |

**Internal（内部密钥，如 `X-Device-Gateway-Internal-Secret` 或现金服专用密钥，与现网内部头惯例对齐）：**

| Method | Path | 说明 |
|--------|------|------|
| GET | `/cash/internal/api/vip/by-wx-id?wxId=` | `{ wxId, isVip, expireAt }`；无权益 → `isVip=false` |

gateway-app：

- `installCashProxyMiddleware`（或等价）绑定 `/cash/app/api/*`、`/cash/internal/api/*`（internal 仅集群内，可按现有 device internal 是否经 gateway 的惯例：voice **直连** `CASH_SERVICE_URL`，不必经 gateway）。
- 白名单：`/cash/app/api/vip/alipay/notify`（及若 Apple Server Notification 二期再加）。

### D4：支付流

```
支付宝：
  建单(created) → 返回 orderStr/调起参数
  → notify 验签 → 校验金额/商户订单号 → paid + 续期 entitlement（幂等）

Apple IAP：
  建单(created, channel=apple_iap) 可选（或 verify 时按 transaction 建/绑单）
  → App POST verify（transaction JWS）
  → App Store Server API / 本地验签（实现选一种，design 倾向 Server API 验交易）
  → 校验 productId 映射 vip_monthly_19 → paid + 续期（幂等 channel_txn_id=transactionId）
```

- **开通只信服务端验签/验单**，不信客户端「已支付」布尔。
- 金额以服务端 `vip_product.price_fen` 为准，拒绝篡改价。

### D5：care-alert 改挂与拆除

- `device.RemoteIsVipByWxID` → 迁移为 `cash.RemoteIsVipByWxID`（或 `voice` 内 cash HTTP 客户端），URL=`CASH_SERVICE_URL`。
- **删除**：`WxIsVipByWxID`、`InternalVipByWxID`、api `vip-by-wx-id`、entity/dao/do `IsVip`、`hack/ddl_wx_is_vip.sql` 标注废弃或删除且 runbook 删除执行提示。
- 保留 care-alert：`careAlertRequireWxID`、触发者权益、降级 Warning。

### D6：配置与密钥

- `CASH_DB_LINK`、`CASH_SERVICE_URL`、`CASH_SERVICE_ADDR`
- 支付宝：`CASH_ALIPAY_APP_ID` / 私钥 / 支付宝公钥（或等价 yaml 段，secret 走 env）
- Apple：`CASH_APPLE_BUNDLE_ID`、`CASH_APPLE_ISSUER_ID` / key 等（按所选验单方式列全）
- 内部密钥：复用 `DEVICE_GATEWAY_INTERNAL_SECRET` 或新增 `CASH_INTERNAL_SECRET`（实现时与 voice 客户端一致，推荐复用现有网关内部密钥减少配置爆炸，若复用须在 design 任务写明）。

### D7：与旧变更关系

- OpenSpec：`care-alert-vip-by-wx` 标记为被本变更 supersede；归档策略可在 archive 时处理，实现以本变更 tasks 为准。
- `llm-care-alert-daily/CONTRACT.md` 同步 VIP→cash。

## Risks / Trade-offs

- [新服务体积大] → tasks 按骨架→权益读→支付→拆除分阶段；可先 internal 读 + 手工插 entitlement 验收 care-alert，再接通支付。
- [支付宝/IAP 沙箱依赖外部] → 凭据与 ASC 商品为发布前置；缺凭据时建单可返回明确「未配置」错误。
- [双源窗口] → 同 PR 拆除 device VIP，部署顺序：先 cash 再 voice，device 拆除可同发。
- [notify 重放] → `(channel, channel_txn_id)` 幂等 + order 状态机。
- [IAP 与支付宝价格] → App Store 定价档可能无法精确 19 元；**以 ASC 实际档位为准映射同一 product_code**，支付宝侧固定 1900 分；文档注明可能存在轻微差异。

## Migration Plan

1. 创建 MySQL 库 `ai_voice_cash`，配置 `CASH_DB_LINK`。
2. 部署 `cash-service`（DDL/EnsureSchema + 种子 SKU）。
3. gateway-app 挂载 `/cash/app/api/*` 与 notify 白名单。
4. 部署 voice（改挂 cash）+ device（拆除 is_vip）同窗口。
5. **不执行** `ddl_wx_is_vip.sql`；若误执行可保留列但代码不再读写。
6. 回滚：voice 临时 `isAccountVIP=false`；cash 停写不影响主站其它域。

## Open Questions

- Apple IAP 验单实现细节（App Store Server API v2 vs 本地 JWS）——实现阶段按团队现有证书选定，不阻塞表结构。
- usage 统计是否计入 `/cash/app/api/*`——**实现前须问负责人**；未答复不改 `maintenance_skip.go`。
