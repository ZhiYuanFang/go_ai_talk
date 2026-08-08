## Why

运维需要在管理端查看账号 VIP 权益（含已过期）及最近一次实付金额，以便客服核对开通/续期情况。现有 `cash-service` 仅有 App/内部读接口，Hub 无 VIP 列表页；禁止在 device 域直查 `ai_voice_cash`，因此须在现金域提供只读 Admin API，并按既有 voice-admin / sim-admin 模式挂到运维 Hub。

## What Changes

- 在 **cash-service** 新增只读管理接口：`GET /cash/admin/api/vip/entitlements`（分页），列表范围覆盖 **全部** `vip_entitlement` 行（含已过期）。
- **激活金额**语义锁定：最近一次 `status=paid` 订单的 `amount_fen`，按 `paid_at DESC`（或等价）取该 `wx_id` 最近一条；无 paid 订单时金额为空/0 并明确可展示。
- 列表字段至少含：`wxId`、是否仍有效（`isVip`/`active`）、`expireAt`、剩余时间（派生；过期可为 0 或「已过期」文案）、`lastPaidAmountFen`、支付 `channel`、`paidAt`。
- **gateway-app**：扩展 cash 反代覆盖 `/cash/admin/api/*`；`IsGatewayAdminAPIPath` 纳入该前缀；按 voice/ucg 模式注入 `X-Admin-Password`（复用 `GATEWAY_APP_ADMIN` / Hub JWT，可增加 `CASH_ADMIN_PASSWORD` 回退）。
- **Hub 静态页**：`/device/admin/cash-vip-admin.html` + `admin-modules.js` 入口 + `RegisterAdminStaticPages` / Bearer 静态页白名单登记（对齐 sim-admin）。
- **禁止** device-service（或其它非 cash 进程）直查 `ai_voice_cash`；改价仍手工 SQL（本变更不做价格编辑）。
- Admin API **不计入** App usage 统计（管理端路径，不修改 `maintenance_skip` 假定 App 统计策略）。
- **不**做：Flutter 购买 UI、改 App 支付 API、退款、自动续签。

## Capabilities

### New Capabilities

- `cash-vip-admin-api`：cash-service 只读 VIP 权益列表 Admin API、金额/范围语义、鉴权与 gateway 反代注入。
- `cash-vip-admin-ui`：运维 Hub 静态管理页、模块登记、列展示与只读交互。

### Modified Capabilities

- （无）基线 `openspec/specs/v3.0.0` 中「New admin modules SHALL be registered in admin-modules.js」等 Hub 约定由本变更**复用并满足**，不改写既有 Requirement 文本；归档时再并入版本基线。

## Impact

- **进程**：`cash-service`（Admin API + 本域 DB 查询）；`gateway-app-server`（反代 `/cash/admin/api/*`、Admin JWT、下游口令注入、静态页注册）。
- **库**：仅 `ai_voice_cash` 的 `vip_entitlement` / `vip_order`（`CASH_DB_LINK`）；**无**新表、无跨库。
- **静态资源**：`resource/public/cash-vip-admin.html`、`admin-modules.js`、`admin_static_pages.go`、`gateway_app_auth_exempt.go`（静态页 path）。
- **密钥**：优先复用 Hub/`GATEWAY_APP_ADMIN_PASSWORD`；可选 `CASH_ADMIN_PASSWORD` 覆盖（与 `VOICE_ADMIN_PASSWORD` 同模式）。
- **usage**：管理端 API，不计入 App usage；实现阶段勿为 Admin 路径改 App `maintenance_skip` 策略。
- **依赖变更**：依赖已落地的 `vip-cash-service`（表与进程）；与 `vip-price-db` 无写路径冲突。
- **非目标**：Flutter care-alert 购买页、App 支付契约变更、退款/自动续签、改价 UI。
