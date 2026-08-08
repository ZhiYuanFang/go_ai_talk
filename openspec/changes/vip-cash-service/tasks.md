## 1. cash-service 骨架与库

- [x] 1.1 新增 `cmd/cash-service/main.go`（`GF_GCFG_FILE`→`config.cash-service.yaml`，`CASH_DB_LINK`，`CASH_SERVICE_ADDR` 默认 `:9807`，避开 notify `:9806`）
- [x] 1.2 新增 `manifest/config/config.cash-service.yaml`（仅本域 database.default 占位 + 支付配置占位注释）
- [x] 1.3 新增 `internal/controller/register_cash_service.go` 与空健康可绑路由；`internal/services/cash` 包骨架
- [x] 1.4 Dockerfile + `docker-compose.microservices.yml` 服务块；`.env.example` 增加 `CASH_DB_LINK` / `CASH_SERVICE_URL` / `CASH_SERVICE_ADDR`
- [x] 1.5 runbook 补充：建库 `ai_voice_cash`、部署顺序、端口

## 2. 表结构与一期商品

- [x] 2.1 DDL（`hack/ddl_cash_vip.sql` 或 EnsureSchema）：`vip_product` / `vip_order` / `vip_entitlement`
- [x] 2.2 种子 `vip_monthly_19`（1900 分、30 天、apple_product_id 可配置）
- [x] 2.3 entity/dao（或 cash 包内最小持久化）与中文注释

## 3. 权益读与内部 API

- [x] 3.1 实现 entitlement 判定：`expire_at > now` ⇒ isVip
- [x] 3.2 `GET /cash/internal/api/vip/by-wx-id`（内部密钥；无效 wxId 拒绝；无行 false）
- [x] 3.3 voice 侧 `RemoteIsVipByWxID` 改指向 `CASH_SERVICE_URL`（新 cash 客户端，勿再走 device）

## 4. App API 与 gateway

- [x] 4.1 `GET /cash/app/api/vip/product`、`GET /cash/app/api/vip/status`、`POST /cash/app/api/vip/orders`（强制 wxId）
- [x] 4.2 gateway-app：`installCashProxyMiddleware` 绑定 `/cash/app/api/*`；透传 Bearer / `X-Internal-Wx-Id`
- [x] 4.3 向负责人确认 usage 统计；未答复前不改 `maintenance_skip.go`；若需匿名仅 notify 则登记白名单

## 5. 支付宝

- [x] 5.1 建单返回支付宝调起参数（金额取商品表；env 配置 appId/密钥）
- [x] 5.2 `POST /cash/app/api/vip/alipay/notify`：验签、金额校验、paid + 续期、幂等；gateway Bearer 白名单
- [x] 5.3 配置项写入 cash 配置/env 说明（secret 不进仓库明文）

## 6. Apple IAP

- [x] 6.1 `POST /cash/app/api/vip/apple/verify`：校验交易、productId 映射 `vip_monthly_19`、paid + 续期、幂等
- [x] 6.2 ASC productId 与种子/配置映射；文档注明沙箱验收步骤

## 7. care-alert 改挂与拆除 wx.is_vip

- [x] 7.1 `isAccountVIP` 仅调 cash Remote；保留强制 wxId / 触发者权益 / 降级 Zhipu
- [x] 7.2 删除 device `InternalVipByWxID`、`WxIsVipByWxID`、api `vip-by-wx-id`、entity/dao/do `IsVip`
- [x] 7.3 删除或废弃 `hack/ddl_wx_is_vip.sql`；清除 runbook 中执行该 DDL 的段落；改为 cash 指引
- [x] 7.4 更新 `openspec/changes/llm-care-alert-daily/CONTRACT.md`（VIP→cash-service）

## 8. 自检

- [x] 8.1 `rg 'vip-by-wx-id|is_vip|RemoteIsVipByWxID' internal`：无 device VIP 残留；voice 无 `dao.Wx` VIP 直查
- [x] 8.2 gateway：`/cash/app/api/*` 反代存在；alipay notify 白名单；care-alert 仍须登录
- [x] 8.3 确认主配置未回流 cash 字段；`CASH_DB_LINK` 出现在 compose / `.env.example` / runbook
