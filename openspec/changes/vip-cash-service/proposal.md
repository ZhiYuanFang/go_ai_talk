## Why

账号 VIP 属于资金与权益域，不应落在 `wx.is_vip` 或由 device-service 承载。care-alert 选模需要按 `wx.id` 判定 VIP，同时需要支付宝与 Apple IAP 开通月会员；二者合并为单一需求，引入独立 `cash-service`（库 `ai_voice_cash`），并 **supersede** 已完成的 `care-alert-vip-by-wx`（拆除 `wx.is_vip` 路径）。

## What Changes

- 新增微服务 **`cash-service`** 与独立库 **`ai_voice_cash`**（`CASH_DB_LINK`），承载 VIP 商品/订单/权益与支付回调。
- 一期单档 SKU：**¥19 / 月 VIP**（`vip_monthly_19`）；支付渠道：**支付宝** + **Apple IAP**（非 Apple Pay）。
- App API 与回调前缀：`/cash/app/api/...`（gateway-app 反代；回调路径 Bearer 白名单 + 渠道验签）。
- 内部读 VIP：`GET /cash/internal/api/vip/by-wx-id`；**voice-service** care-alert 经此判定触发者 VIP（失败降级 Zhipu）。
- care-alert 保留：强制 `X-Internal-Wx-Id>0`、触发者权益、日缓存仍按 `deviceNo+上海日`、不扣 clinic 配额。
- **BREAKING（相对 care-alert-vip-by-wx）**：同变更拆除 `wx.is_vip`、device `vip-by-wx-id`、`RemoteIsVipByWxID`→device，以及相关 DDL/runbook；**禁止**执行 `hack/ddl_wx_is_vip.sql`。
- 更新 `llm-care-alert-daily/CONTRACT.md`：VIP 真相源改为 cash-service。

## Capabilities

### New Capabilities

- `cash-service-runtime`：`cash-service` 进程、库连接、配置边界、compose/env/runbook 骨架。
- `vip-entitlement`：权益表语义、内部按 `wxId` 查询是否 VIP。
- `vip-payment`：一期月会员下单、支付宝 notify、Apple IAP 验单、幂等开通/续期。
- `care-alert-cash-vip`：care-alert 改挂 cash 读 VIP；拆除 device/`wx.is_vip` 路径。

### Modified Capabilities

- （无）基线 `openspec/specs/v3.0.0` 尚无 cash/VIP 支付 capability；本变更以增量规格引入。归档时再合并版本基线。

## Impact

- **新进程**：`cmd/cash-service`；配置 `manifest/config/config.cash-service.yaml`；Compose / `.env.example` / `docs/runbooks/release-deploy-and-run.md`。
- **库**：`ai_voice_cash`（仅 cash-service 访问）；表含 product / order / entitlement（见 design）。
- **gateway-app**：反代 `/cash/app/api/*`；支付回调精确 path 入 Bearer 白名单；usage 统计须先问负责人后再改 `maintenance_skip`（未答复则不改）。
- **voice-service**：`isAccountVIP` 改调 cash internal；删除对 device VIP 的依赖。
- **device-service**：删除 VIP 读契约与 `is_vip` 模型字段；**不**承载支付。
- **密钥**：支付宝与 Apple IAP 凭据仅进 cash 配置/环境变量。
- **非目标**：微信收款、多档 SKU、退款自动关权益完整流水（可留状态位）、独立账务对账中台、VIP Redis 读缓存。
