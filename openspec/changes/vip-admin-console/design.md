## Context

`vip-cash-service` / `vip-price-db` 已落地：`cash-service` + 库 `ai_voice_cash`（`vip_entitlement` / `vip_order` / `vip_product`），App 与内部读 VIP 契约已存在。运维 Hub（`admin.html` + `admin-modules.js` + `admin-common.js`）已有 voice-admin / sim-admin 等模块：静态页挂在 `gateway-app-server` 的 `/device/admin/*`，API 走领域前缀 `/<domain>/admin/api/*`，网关校验 Admin JWT 后注入 `X-Admin-Password`。当前 **无** VIP 权益列表管理页；改价仍手工 SQL（刻意不进本变更）。

约束：一服务一库；**禁止** device 或其它进程直查 `ai_voice_cash`；Redis 本期不引入新读缓存；Admin 路径不计入 App usage。

## Goals / Non-Goals

**Goals:**

- cash-service 提供分页只读 Admin API，列出**全部**权益行（含已过期），并附最近一次 paid 订单金额与渠道信息。
- gateway-app 反代 `/cash/admin/api/*` + Admin JWT + 下游口令注入。
- Hub 静态页只读展示：是否有效、到期、剩余/已过期、激活金额、渠道、支付时间。
- 鉴权与现有 admin 模块一致（Hub JWT → 网关注入 `X-Admin-Password`）。

**Non-Goals:**

- Flutter / App 购买 UI；改 App 支付 API。
- 改价、退款、自动续签、写权益、手工开通。
- VIP 列表 Redis 缓存；跨库 JOIN 账号资料（昵称等一期可不展示）。

## Decisions

### D1：架构 Option A（域 Admin API + Hub 静态页）

- **选型**：cash-service 承载 API；静态 HTML 注册在 device Hub（与 voice-admin / sim-admin 同模式）。
- **否决**：在 device-service 查 VIP（跨库违规）；单独新管理进程（过重）。
- 静态 path：`/device/admin/cash-vip-admin.html`（文件 `resource/public/cash-vip-admin.html`）。
- 模块 id：`cash-vip-admin`；`admin-modules.js` 增加入口（`title` 如「VIP 权益」）；`admin_static_pages.go` + `gateway_app_auth_exempt.go` 静态页白名单同步（与 sim-admin 一致：页本身可匿名加载，API 须 Admin JWT）。

### D2：Admin API 契约

| Method | Path | 说明 |
|--------|------|------|
| GET | `/cash/admin/api/vip/entitlements` | 分页列表；只读 |

Query（对齐 sim users 习惯）：

- `page`：默认 1
- `pageSize`：默认 20，最大 200
- 可选 `wxId`：精确过滤（便于客服查单号；无则全量分页）

响应（示意）：

```json
{
  "list": [
    {
      "wxId": 123,
      "isVip": false,
      "expireAt": 1710000000,
      "remainingSeconds": -3600,
      "lastPaidAmountFen": 1900,
      "channel": "alipay",
      "paidAt": 1709000000
    }
  ],
  "page": 1,
  "pageSize": 20,
  "total": 42
}
```

- `isVip`：`expire_at > now`（unix 秒，与内部 `by-wx-id` 一致）。
- `remainingSeconds`：`expire_at - now`（可负）；UI 可将 `<=0` 显示为「已过期」。
- `lastPaidAmountFen` / `channel` / `paidAt`：来自该 `wx_id` **最近一次** `status='paid'` 订单；无则金额为 `0` 或省略、`channel`/`paidAt` 空/0。
- 排序默认：`expire_at DESC`（或 `updated_at DESC`）；实现选定一种并在代码注释写明；带 `wxId` 时返回 0～1 行即可。

### D3：激活金额 SQL 语义（产品锁定）

- **定义**：`vip_order` 中 `wx_id=? AND status='paid'`，按 `paid_at DESC`（`paid_at` 相同可再按 `id DESC`）取 **一条** 的 `amount_fen`。
- 列表实现推荐：`vip_entitlement` **LEFT JOIN** 派生表/子查询「每 wx_id 最新 paid 订单」，避免 N+1。
- **禁止**用商品现价或 `created` 未付订单冒充激活金额。

### D4：列表范围（产品锁定）

- 主表为 **`vip_entitlement` 全量行**（含 `expire_at <= now`）。
- **不**默认只筛有效 VIP；若未来加 `activeOnly` 开关，须另开变更（本期不做）。

### D5：gateway 反代与鉴权

1. `installCashProxyMiddleware`：除 `/cash/app/api/*` 外，绑定 **`/cash/admin/api/*`** → `CASH_SERVICE_URL`。
2. `IsGatewayAdminAPIPath`：增加 `strings.HasPrefix(path, "/cash/admin/api/")`。
3. `InjectAdminDownstreamPassword`：`/cash/admin/api/` → `CashAdminPassword()`：
   - env `CASH_ADMIN_PASSWORD` 非空则用之；
   - 否则回退 `GATEWAY_APP_ADMIN_PASSWORD`（同 Voice/Sim）。
4. cash-service Admin handler：校验 `X-Admin-Password`（`VerifyCashAdminPassword`，常量时间比较），失败 401/403。
5. 浏览器 **禁止**自带 `X-Admin-Password`；页用 `AdminCommon.requireAdmin()` + `AdminCommon.adminFetch`。

Internal `/cash/internal/api/*` 仍由 voice 等直连，**不**经本 Admin 路径。

### D6：UI 列与只读

页面列：wxId | 是否 VIP/有效 | 到期时间 | 剩余时间（或「已过期」）| 激活金额（分或元展示，页内注明单位）| 渠道 | 最近支付时间。

- 分页控件；可选 wxId 查询框。
- **无**编辑/退款/改价按钮。
- 中文界面文案。

### D7：usage / 统计

- Admin API 为运维通道，**不计入** App usage；**不**为 `/cash/admin/api/*` 改 `maintenance_skip.go`（该文件管 App 维护型跳过）。
- 无需询问 App usage 负责人（非 App 对外接口）；若实现误把 Admin 路径纳入 usage 采集，须显式排除。

### D8：DB / Redis

- 仅 cash-service 读本域表；无 DDL 变更（现有索引 `PRIMARY(wx_id)`、`idx_wx_created` 足够；若「最新 paid」查询慢，可后续加 `(wx_id, status, paid_at)`——本期评估后再定，默认子查询可接受）。
- **不加**新 Redis 读缓存（与 vip-cash-service「一期直读 DB」一致）。

## Risks / Trade-offs

- [权益行多时全表分页偏慢] → 默认 pageSize 20；可加 wxId 过滤；必要时再加索引（另开或同 PR 小改）。
- [无 paid 订单的权益行金额为空] → 允许；UI 显示「—」或 0。
- [口令 env 未配导致 Admin 全拒] → 与其它 admin 模块相同；runbook 提示 `GATEWAY_APP_ADMIN_PASSWORD` / 可选 `CASH_ADMIN_PASSWORD`。
- [静态页误入 App Bearer 白名单过宽] → 仅登记 HTML path（与现网其它 admin 页一致），API 仍走 Admin JWT。

## Migration Plan

1. 部署含 Admin handler 的 `cash-service`（表已存在则无需 DDL）。
2. 部署 `gateway-app-server`（proxy + `IsGatewayAdminAPIPath` + 静态页 + modules）。
3. 运维用 Hub 账号登录 → 打开「VIP 权益」模块验收。
4. 回滚：去掉 Hub 入口与 `/cash/admin/api/*` 反代即可；无数据迁移。

## Open Questions

- （无关键阻塞项）可选：列表默认排序字段在实现时选 `expire_at DESC` 即可；金额展示「分」还是「元」由 UI 定（API 固定 fen）。
