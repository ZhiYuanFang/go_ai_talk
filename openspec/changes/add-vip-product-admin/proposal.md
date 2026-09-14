## Why

VIP 月会员的现价、划线价与 Apple IAP 商品 ID 目前靠手工 SQL 或环境变量 `CASH_APPLE_PRODUCT_ID` 维护，而功能 SKU 已能在「开通功能管理」改价与填 Apple ID。运维无法在同一后台配齐四档付费商品；Hub 上独立的「VIP 权益」入口与商品配置职责混在导航里，不便使用。

## What Changes

- 在 **开通功能管理** 页增加 **VIP 套餐** 编辑区：配置一期 `vip_monthly_19` 的现价、原价、时长、Apple 商品 ID（及标题/上架，若表单需要）。数据仍写 `vip_product`，**不**把 VIP 并入 `feature_def` / `feature_product`，App catalog 与 VIP overlay 语义不变。
- cash-service 新增 VIP 商品 Admin 读/写 API（`/cash/admin/api/vip/product`）；建单/读价/Apple 验单 **只认库**，去掉 `CASH_APPLE_PRODUCT_ID` 与 yaml `cash.appleProductId` 回退。
- 从 `manifest/docker/.env.example`、compose 注入与 runbook 中 **删除** `CASH_APPLE_PRODUCT_ID`。**保留** `CASH_APPLE_BUNDLE_ID`（整 App Bundle 验 JWS）。
- Hub 导航 **去掉**「VIP 权益」模块入口（`showInNav: false`）；静态页 `cash-vip-admin.html` 保留为只读权益列表。开通功能管理页提供按钮，文案 **「VIP列表」**，链到该页。
- **不**改 App `api/v1` 的 `/cash/app/api/vip/*` 响应结构；**不**新增多档 VIP SKU；**不**做退款/手工开通。

## Capabilities

### New Capabilities

- `vip-product-admin`：VIP 商品 Admin 读写、开通功能管理页 VIP 套餐表单、Apple/支付宝建单仅读库、移除 `CASH_APPLE_PRODUCT_ID`。

### Modified Capabilities

- `vip-product-pricing`：改价写路径从「仅 SQL、禁止 Admin/env」改为「Admin（及等价 SQL）写库；禁止 env 控价/控 Apple 商品 ID」。匿名读价与建单用库内 `price_fen` 不变。
- `cash-vip-admin-ui`：Hub 主导航不再展示 VIP 权益模块；权益只读页仍注册，入口改为开通功能管理页「VIP列表」按钮。
- `feature-admin-ui`：开通功能管理页 MUST 含 VIP 套餐配置与「VIP列表」入口（VIP 仍非功能定义行）。

## Impact

- **进程**：`cash-service`（EnsureSchema 种子、读价/验单、Admin API）；`gateway-app-server`（既有 `/cash/admin/api/*` 反代、静态页与 `admin-modules.js`）。无新微服务、无新库。
- **库**：仅 `ai_voice_cash.vip_product`（`CASH_DB_LINK` / cash-service `database.default`）；不增表。
- **API**：新增 Admin `GET`/`POST` `/cash/admin/api/vip/product`（`g.Meta` + cash 反代前缀）。无新增 App HTTP，**不**改 `maintenance_skip.go`，usage 无需新确认。
- **Redis**：不加新读缓存；VIP 商品读写走 MySQL。
- **配置**：删除 `CASH_APPLE_PRODUCT_ID`；同步 `.env.example`、compose、`config.cash-service.yaml` 注释、`docs/runbooks/vip-commercial-config.md` 与 `release-deploy-and-run.md`。
- **非目标**：VIP 写入功能权益表；Flutter 契约变更；多 VIP 档位；支付宝密钥进 Admin；新增测试文件。
