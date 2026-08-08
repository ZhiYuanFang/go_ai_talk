## 1. cash-service Admin API

- [x] 1.1 新增 `VerifyCashAdminPassword`（或等价）：优先 `CASH_ADMIN_PASSWORD`，否则回退 `GATEWAY_APP_ADMIN_PASSWORD`；常量时间比较；附中文注释
- [x] 1.2 实现权益列表查询：主表 `vip_entitlement` 全量（含已过期）；LEFT JOIN 每 `wx_id` 最近一条 `status=paid` 订单（`paid_at DESC`，并列 `id DESC`）；派生 `isVip`、`remainingSeconds`
- [x] 1.3 注册 `GET /cash/admin/api/vip/entitlements`（page/pageSize/可选 wxId）；口令失败拒绝；响应含 list/page/pageSize/total 与 design 字段；只读、详细中文注释
- [x] 1.4 确认无跨库、无新 Redis 读缓存、无 DDL 强依赖；慢查时再评估 `(wx_id, status, paid_at)` 索引（本期默认可不加）

## 2. gateway-app 反代与鉴权

- [x] 2.1 `installCashProxyMiddleware` 绑定 `/cash/admin/api/*` → `CASH_SERVICE_URL`
- [x] 2.2 `IsGatewayAdminAPIPath` 纳入 `/cash/admin/api/` 前缀
- [x] 2.3 `InjectAdminDownstreamPassword` / `CashAdminPassword()` 对 `/cash/admin/api/` 注入口令（与 voice/sim 同模式）
- [x] 2.4 确认 Admin API **不**计入 App usage；勿仅为该路径改 `maintenance_skip.go`

## 3. Hub 静态页与模块登记

- [x] 3.1 新增 `resource/public/cash-vip-admin.html`：`AdminCommon.requireAdmin` + `adminFetch` 拉列表；分页；可选 wxId；列含 wxId/是否有效/到期/剩余或「已过期」/激活金额/渠道/支付时间；只读无改价/退款
- [x] 3.2 `admin-modules.js` 增加 `cash-vip-admin` 入口（`pagePath=/device/admin/cash-vip-admin.html`，`showInNav: true`）
- [x] 3.3 `admin_static_pages.go` 注册静态 path；`gateway_app_auth_exempt.go` 同步静态页白名单（对齐 sim-admin）
- [x] 3.4 浏览器侧 MUST NOT 发送 `X-Admin-Password`

## 4. 配置与文档

- [x] 4.1 `.env.example` / compose 注释补充可选 `CASH_ADMIN_PASSWORD`（可回退 Hub 口令）；无需新 DB 连接
- [x] 4.2 `docs/runbooks/release-deploy-and-run.md`（或既有 cash VIP sandbox runbook）简述 Hub「VIP 权益」验收路径与口令依赖
- [x] 4.3 自检：device 进程无 `ai_voice_cash` 直查；grep Admin 路径已走 cash-service

## 5. 验收

- [x] 5.1 Hub 登录可见模块；打开页可加载含已过期行的列表；激活金额与最近 paid 订单一致
- [x] 5.2 无 Admin JWT 时 `/cash/admin/api/vip/entitlements` 被网关拒绝；错误口令被 cash-service 拒绝
- [x] 5.3 不改 App 支付 API；无改价 UI；无新增 `*_test.go`
