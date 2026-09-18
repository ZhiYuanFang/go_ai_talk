## Why

开通功能管理只能改套餐和权益，管理者看不到付费收入和自己记下的成本，无法对比实际成本与收益。订单金额已经在 `vip_order` / `feature_order` 里，缺的是一块只读合计，加上手填进账/出账。

## What Changes

- 开通功能管理顶栏增加「收益统计」面板（与 VIP、功能、群二维码并列）。
- cash-service 新增一张手填流水表：方向（进账/出账）、名称、金额（分）、发生时间。支持增删改查。名称不唯一。
- 同一面板只读展示：VIP 付费合计、功能付费按功能名拆开的合计、手填进账合计、手填出账合计，以及差额（VIP + 各功能付费 + 手填进账 − 手填出账）。
- 付费合计现场对已支付订单求和，不把订单金额再写入新表。只计 `status=paid` 且渠道为 `alipay` 或 `apple_iap`。`channel=admin` 与 `refunded` 不计入。
- 可选时间范围同时作用于订单 `paid_at` 与手填 `occurred_at`。不传则全部。
- 订单金额是用户支付的标价，不是到账净额。Apple 抽成、支付宝手续费、服务器与模型费用由管理者记为出账，系统不自动扣费率。

## Capabilities

### New Capabilities

- `cash-admin-revenue`: 管理端收益统计面板、手填进账/出账，以及 VIP / 按功能名拆开的付费合计与差额。

### Modified Capabilities

- （无）不改 App 支付、VIP 状态、功能开通或既有 Admin 接口的行为。

## Impact

- **进程**：`cash-service`。表与查询走该进程现有 MySQL，配置组 `default`，环境变量 `CASH_DB_LINK`（兜底 `GF_DATABASE_DEFAULT_LINK`）。与 `vip_order` / `feature_order` 同库。
- **代码**：`internal/services/cash`（`EnsureSchema` 建表、流水 CRUD、合计查询）、`api/v1` 新增 Admin 路由、`internal/controller/cash`、`resource/public/cash-feature-admin.html`。
- **网关**：新接口挂在已有 `/cash/admin/api/*` 反代下，不新增 Bind，不改 Bearer 白名单。
- **不做**：不改 App API 结构；不改 `maintenance_skip`；不加 Redis；不加后台循环任务；不批量改历史订单；不自动计算渠道手续费。
