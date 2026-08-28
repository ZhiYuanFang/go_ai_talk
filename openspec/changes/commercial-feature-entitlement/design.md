## Context

`cash-service` + `ai_voice_cash` 已承载 VIP 商品/订单/权益（`wx_id`）。Flutter 商业闭环需要设备维功能开通、UCG 喂养门槛、预测数量、邀请码推广与可写 Admin。支付功能尚未上线，可共用 VIP 支付回调并自由演进履约分发。

约束：一服务一库；跨域禁止直查；Redis 经 `cachekit` + platform 键 builder；App 经 gateway-app `/cash/app/api/*`；Admin 对齐 Hub；资格依赖 history 契约。

**Redis 产品确认**：资格日结果、功能定义热读、`allowedCount` 热读、广告幂等走 Redis；收益率：设备维/字典热读、写少读多、失效点清晰。键 ONLY 登记于 `internal/platform/cachekit/keys_*.go`。

**Usage 产品确认**：开通意图链路（建单 / 邀请码兑换 / 广告完成）**计入** usage；查询链路（资格 / 合成目录等）**不计入**，须写入 `maintenance_skip.go`。渠道回调（支付宝 notify）不计入。

## Goals / Non-Goals

**Goals:**

- UCG 入场：连续 7 有效喂养日；与 VIP/功能列表隔离。
- 合成功能列表：登录 + 绑 `device_no`；项内含是否已开通；可直接驱动客户端展示。
- 预测 `allowedCount` 授予与列表暴露；不下发事件 ID。
- 支付 / 邀请码 / 广告开通；支付回调与 VIP 共用分发。
- 邀请码：推广渠道、owner 绑定、一家锁定（成功开通后绑定）、单功能逐次兑换、不可自用、兑换明细。
- Admin：付费功能 CRUD + 独立邀请码管理（含按码追踪兑换人 `wx_id`）。
- V1：客户端 `isVip` 覆盖功能列表全开；服务端 VIP 不写功能表；UCG 不受 VIP 影响。

**Non-Goals:**

- Flutter UI（含隐藏 VIP 开通入口由客户端负责）。
- 邀请码付费购买、推广分润结算（二期；本期只落流水字段）。
- 广告服务端验真；微信收款；资格 MySQL 持久化；新建微服务。
- App 侧独立 `entitlements` / `allowed-count` 接口（能力并入合成列表；内部/Admin 可读明细）。

## Decisions

### D1：落域 — 扩展 cash-service

- 表落 `ai_voice_cash`，代码 `internal/services/cash/**` + `internal/controller` 注册。
- App：`/cash/app/api/...`；Admin：`/cash/admin/api/feature/...` 与 `/cash/admin/api/invite-code/...`（或等价前缀）。
- **备选**独立 feature-service：否决（支付必须回写 cash）。

### D2：表模型（最少集）

```
feature_def
  feature_id PK
  title, description
  unlock_methods          -- payment|invite_code|ad
  duration_days           -- 0=永久（支付/码默认；可被码行覆盖）
  status, sort_order, updated_at

feature_product
  product_code PK
  feature_id
  grant_kind              -- entitlement | allowed_count_delta
  grant_quantity
  price_fen, original_price_fen
  duration_days
  apple_product_id, status, updated_at

feature_invite_code                 -- 邀请码（原激活码）
  code PK/UK
  owner_wx_id UK                    -- 一期：一用户一码
  expires_at                        -- 码本身过期
  max_redemptions, redeemed_count   -- 按「成功兑换次数」或按人；实现固定并注释
  grant_duration_days
  status, created_at, updated_at

feature_invite_code_feature         -- 码可开的功能（可为空=所有 unlock_methods 含 invite_code 的启用功能）
  code, feature_id
  grant_quantity                    -- 对 allowed_count 类有意义
  PK(code, feature_id)

feature_invite_redeemer_bind        -- 一家锁定
  redeemer_wx_id PK
  owner_wx_id
  bound_at                          -- 首次成功开通某功能后写入

feature_invite_feature_grant        -- 人×功能码开去重
  redeemer_wx_id, feature_id PK
  code, device_no, redeemed_at

feature_invite_redemption           -- 兑换流水（Admin 追踪 / 二期分润）
  id AI
  code, owner_wx_id, redeemer_wx_id, device_no, feature_id
  redeemed_at

feature_entitlement
  id AI
  device_no, feature_id
  unlock_method                     -- payment|invite_code|ad
  expires_at                        -- 0=永久
  quantity, source_ref
  created_at, updated_at
  UNIQUE(device_no, feature_id)     -- 同设备同功能一行；续期更新

feature_allowed_count
  device_no PK
  allowed_count INT DEFAULT 0
  updated_at

feature_order
  order_no PK, device_no, wx_id
  product_code, channel, amount_fen, status
  channel_txn_id, paid_at, created_at
```

- VIP 表不变；履约按订单落在 `vip_order` 或 `feature_order` 分流。

### D3：UCG 入场资格

- 日历 `Asia/Shanghai`；有效日：该日该 `device_no` history 行数 ≥10（默认全部 history；产品若改喂养 eventId 则改契约过滤）。
- **算法 A**：以请求日为锚向前连续统计有效日；`qualified = effectiveDays >= 7`；`remainingDays = max(0, 7 - effectiveDays)`；`requiredDays = 7`。
- **VIP / 功能权益 MUST NOT 参与计算或短路**。
- history：`GET /history/internal/api/feeding-day-stats?deviceNo=&days=N`（内部密钥）；cash 经 `HISTORY_SERVICE_URL`；禁止跨库。
- Redis：`cash:ucg:eligibility:{deviceNo}:{yyyyMMdd}`；同日命中不重算；不落 MySQL；无 ticker。
- history 不可用：资格 API **fail-closed**（返回错误，不得伪造 `qualified=true`）。

### D4：App HTTP 与身份

| Method | Path | 鉴权 | usage |
|--------|------|------|-------|
| GET | `/cash/app/api/ucg/eligibility` | 登录 + `X-Internal-Device-No` | 不统计 |
| GET | `/cash/app/api/feature/catalog` | 登录 + device_no | 不统计 |
| POST | `/cash/app/api/feature/orders` | 登录 + device_no | **统计** |
| POST | `/cash/app/api/feature/invite-codes/redeem` | 登录 + device_no + wx_id | **统计** |
| POST | `/cash/app/api/feature/ad/complete` | 登录 + device_no | **统计** |
| （共用） | `/cash/app/api/vip/alipay/notify` | 匿名 + 验签 | 不统计 |
| （共用） | `/cash/app/api/vip/apple/verify` | 登录 | 按既有 |

- **`device_no` 只信**网关注入的 `X-Internal-Device-No`（先 strip 伪造）；写/读设备维接口忽略 query/body `deviceNo`；空则拒绝。
- **`wx_id` 只信** `X-Internal-Wx-Id`（邀请码人维度）；兑码要求 `wx_id > 0`（纯设备会话无 wx 则拒绝兑码）。
- 合成 catalog：**不** Bearer exempt；与 VIP `product` 匿名可读不同。
- catalog 项至少：`featureId`、展示字段、`unlockMethods`、`unlocked`；已开通时 `unlockMethod`/`expiresAt`；预测类附 `allowedCount`。
- **不**提供 App `entitlements` / `allowed-count` 独立路径（并入 catalog）。
- 所有新 App 路由：`api/v1` g.Meta summary（供 apiregistry）。

### D5：权益与 VIP（V1）

- 永久 `expires_at=0`；有期读时过滤过期。
- 同设备同 feature 重复支付/码：**续期** `max(now, expire)+duration`（取较长语义并入该公式）；`allowed_count` **累加**。
- VIP 履约只动 `vip_*`；功能履约只动 feature_*；交叉禁止。
- 客户端：`isVip` 覆盖功能列表全开；**不得**覆盖 UCG 入场；可先隐藏 VIP 开通入口。服务端 catalog 仍返回真实权益/`allowedCount`。

### D6：支付回调共用

- 支付宝仍走现有 notify URL；handler 按 `out_trade_no` 查 `feature_order` 或 `vip_order` 分流履约。
- Apple verify 同入口扩展：按订单/商品映射分流。
- 未上线：无历史订单兼容包袱。

### D7：邀请码规则

1. 生成：必须 `owner_wx_id`；`UNIQUE(owner_wx_id)`（一期一人一码）。
2. 能力：一码可对应多个功能（子表或「所有支持 invite_code 的功能」）；**每次兑换 body 带 `featureId`**，只开这一个。
3. 不可自用：`redeemer_wx_id != owner_wx_id`。
4. 一家锁定：仅当**成功开通某一功能后**写入 `feature_invite_redeemer_bind`；之后只能用该 `owner_wx_id` 的码；失败兑换不绑定。
5. 人×功能：`feature_invite_feature_grant` 保证一 wx 对一 feature 仅码开一次。
6. 同码可多次调用：每次不同尚未码开的 `featureId`。
7. 流水：每次成功写入 `feature_invite_redemption`（owner/redeemer/device/feature/time）。
8. 二期：付费领码、分润——本期非目标，流水字段预留。

### D8：Admin

- 付费功能管理：defs、products、解锁方式。
- **独立**邀请码管理：CRUD、有效期、可开功能、owner_wx_id、按码兑换明细（wx_id + device_no + feature + time）。
- 鉴权同 VIP Admin；静态页 + `admin-modules.js` 两个入口（或明确两个一级模块）。

### D9：Redis

- 资格：按日 key。
- 功能定义：全站短 TTL + Admin 写失效；合成列表在请求路径 JOIN 设备权益（可选 per-device 短 TTL，履约失效）。
- `allowedCount`：设备维；履约失效。
- 广告幂等短 TTL。
- 禁止业务 `g.Redis()`。

### D10：配置与可观测

- cash：`HISTORY_SERVICE_URL`、`DEVICE_GATEWAY_INTERNAL_SECRET`；compose / `.env.example` / runbook 补充。
- 禁止 `config.cash-service.yaml` 回流 history DB link。
- 跨服务调用失败打可观测日志。

## Risks / Trade-offs

- [广告可刷] → MVP 接受；设备日限额 + Redis 幂等；二期验真。
- [一人一码 vs 二期多码] → 一期硬 UNIQUE(owner)；二期再放宽。
- [换绑设备] → 码开去重按 wx，新设备可能无权益行且不能再码开该功能 → 可用支付/广告/VIP(V1)。
- [usage] → 已确认：查询 skip、开通 POST 统计。
- [资格缓存当日不刷新] → 接受「按日算一次」；跨日换 key。

## Migration Plan

1. EnsureSchema 新表；种子可选仅预测相关 `feature_def`。
2. history 内部日统计上线。
3. cash App/Admin + Redis 键 + api/v1 Meta。
4. 扩展支付回调分发 feature 履约。
5. Hub 双模块静态页；maintenance_skip 登记查询 path。
6. 回滚：下线 feature 路由/商品；不影响 VIP。

## Open Questions

- （已关闭）一家绑定时机 → **仅成功开通某一功能后**。
- （已关闭）UCG 算法 → 连续 7 有效日；VIP 无效。
- （已关闭）VIP × 功能 → V1 客户端覆盖。
- （已关闭）catalog 匿名 → 否，须登录+绑机。
- 有效喂养是否过滤 eventId → 默认全部 history；产品变更再改契约。
