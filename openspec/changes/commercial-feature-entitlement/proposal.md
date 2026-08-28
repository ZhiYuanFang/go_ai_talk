## Why

Flutter 商业闭环需要服务端提供：UCG 入场资格（连续有效喂养日，与 VIP/功能开通无关）、按设备的功能开通读模型（列表项含是否已开通）、预测「开通数量」、支付/邀请码/广告履约，以及运维 Hub 的付费功能与邀请码管理（含兑换追踪）。现有 `cash-service` 仅有账号 VIP（`wx_id`），无设备维功能目录、邀请码推广模型与预测 `allowedCount`。本变更在 Go 侧补齐该商业后端；支付尚未上线，回调与 VIP **共用**且无历史兼容包袱。

## What Changes

- 在 **`cash-service`**（库 `ai_voice_cash`）扩展：功能定义、功能 SKU、邀请码（owner 绑定）、兑换流水、按 `device_no` 的权益与预测 `allowedCount`、功能订单表。
- **UCG 入场资格 API**（独立）：连续 **7** 个上海有效喂养日（当日 history ≥10）；**VIP / 功能权益 MUST NOT 改变** `qualified`；结果按日 Redis 缓存、不落库；经 history HTTP 契约取数。
- **功能列表 API**（合成读）：须登录且 JWT 已绑 `device_no`（只信 `X-Internal-Device-No`）；返回启用功能清单，**每项含是否已开通**（及开通方式/到期等）；**不含** UCG 资格；MVP 可仅配置预测数量类功能，模型按多功能预留。
- **VIP（V1）**：购买/续期 **MUST NOT** 写功能权益表；功能列表「全解锁」由客户端 `isVip` 覆盖（客户端可先隐藏 VIP 开通入口）；UCG 入场 **禁止** 用 `isVip` 绕过。
- **开通方式**：独立功能 SKU 支付（回调与 VIP **共用**分发）、邀请码（推广/渠道）、广告完成（MVP 信客户端）。
- **邀请码**：一码可覆盖所有支持邀请码的功能；每次兑换指定 **一个** `featureId`（不自动开其余）；人=`wx_id`；一人一功能仅能码开一次；一人只能绑定 **一家** owner（**仅成功开通某一功能后**才绑定）；owner 不可自兑；一期一 owner 一码；流水支持按码追踪兑换人。
- **Admin**：付费功能 CRUD 页 + **独立**邀请码管理页（有效期、功能能力、owner、兑换明细）。
- **Redis**：资格日缓存、目录定义热读、`allowedCount`/广告幂等经 `cachekit` + platform 键 builder（产品确认加 Redis）。
- **gateway-app**：反代核验；查询类 App 接口 **不计入** usage（`maintenance_skip`）；开通意图 POST **计入**；功能类路径 **不** Bearer 匿名豁免；`api/v1` g.Meta 登记；静态页与 Hub 模块登记。

## Capabilities

### New Capabilities

- `ucg-entry-eligibility`：UCG 入场资格（连续 7 有效日）、与 VIP/功能隔离、Redis 日缓存、history 契约。
- `feature-catalog-entitlement`：绑机合成功能列表（含 `unlocked`）；权益权威在 MySQL、主体 `device_no`。
- `prediction-allowed-count`：预测开通数量权威与授予语义；App 经列表项暴露 `allowedCount`（不下发事件 ID）。
- `feature-unlock-fulfillment`：支付共用回调分发、邀请码兑换规则、广告开通；VIP 履约不写功能表。
- `feature-admin-api`：付费功能与邀请码 Admin HTTP（含兑换明细）。
- `feature-admin-ui`：Hub「开通功能管理」+「邀请码管理」静态页与模块登记。

### Modified Capabilities

- （无强制改写 v3.0.0 既有 Requirement 文本）复用 gateway-app 反代、Admin Hub、Redis platform、一服务一库等基线；归档时再合并版本基线。

## Impact

- **进程**：`cash-service`；`history-service`（内部按日统计）；`gateway-app-server`（usage skip、静态页、auth 自检）。
- **库**：仅 `ai_voice_cash`；禁止 cash 直查 history/device。
- **部署**：cash 增加 `HISTORY_SERVICE_URL` 等；不新建微服务/ACR 矩阵项。
- **非目标**：Flutter UI；微信收款；广告服务端验真；邀请码付费购买与分润结算（流水预留）；资格落 MySQL；VIP 写功能授权行；UCG 进功能列表。
