## ADDED Requirements

### Requirement: cash-service SHALL 提供分页只读 VIP 权益 Admin API

cash-service MUST 提供 `GET /cash/admin/api/vip/entitlements`，鉴权 MUST 校验网关注入的 `X-Admin-Password`（与 cash Admin 口令配置一致）。该接口 MUST 为只读，MUST NOT 修改 `vip_entitlement`、`vip_order` 或 `vip_product`。Query MUST 支持 `page`（默认 1）、`pageSize`（默认 20，最大 200），以及可选精确过滤 `wxId`。响应 MUST 含 `list`、`page`、`pageSize`、`total`。

#### Scenario: 合法管理员分页拉取

- **WHEN** 请求携带正确 `X-Admin-Password` 且未指定 `wxId`
- **THEN** cash-service MUST 返回分页后的权益列表与 `total`

#### Scenario: 口令错误

- **WHEN** 请求未携带或携带错误的 `X-Admin-Password`
- **THEN** cash-service MUST 拒绝该请求且 MUST NOT 返回权益数据

#### Scenario: 按 wxId 过滤

- **WHEN** 请求携带合法正整数 `wxId` 与正确口令
- **THEN** 响应 `list` MUST 仅包含该 `wxId` 的权益行（0 或 1 条）

### Requirement: 权益列表范围 SHALL 包含已过期 VIP

列表主数据源 MUST 为 `vip_entitlement` 的全部行，MUST NOT 默认过滤为仅 `expire_at > now`。每行 MUST 提供当前是否仍有效的布尔字段（如 `isVip`：当且仅当 `expire_at` 晚于当前时间），以及 `expireAt`（unix 秒）与可派生的剩余秒数（如 `remainingSeconds`，过期时可为 0 或负数）。

#### Scenario: 已过期权益仍出现在列表

- **WHEN** 某 `wx_id` 存在权益且 `expire_at` 已不晚于当前时间
- **THEN** 该行 MUST 仍可出现在列表中，且有效布尔字段 MUST 为 false

#### Scenario: 未过期权益标记有效

- **WHEN** 某权益 `expire_at` 晚于当前时间
- **THEN** 该行有效布尔字段 MUST 为 true

### Requirement: 激活金额 SHALL 取最近一次已支付订单金额

每条权益行的激活金额（如 `lastPaidAmountFen`）MUST 取该 `wx_id` 在 `vip_order` 中 `status=paid` 的订单，按 `paid_at DESC`（相等时可按 `id DESC`）的最近一条之 `amount_fen`。同条 MUST 附带该订单的 `channel` 与 `paidAt`（字段名以实现为准，语义一致；无 paid 订单时渠道与支付时间 MUST 为空或 0）。MUST NOT 使用未支付订单或商品现价冒充激活金额。若无任何 paid 订单，金额 MUST 为 0 或空语义，且 MUST NOT 伪造渠道支付时间。

#### Scenario: 多次支付取最近一次

- **WHEN** 同一 `wx_id` 存在多笔 `status=paid` 订单且 `paid_at` 不同
- **THEN** 激活金额与渠道、支付时间 MUST 对应 `paid_at` 最新的那一笔

#### Scenario: 无已支付订单

- **WHEN** 某权益行对应 `wx_id` 不存在 `status=paid` 的订单
- **THEN** 激活金额 MUST 不以未支付订单填充，响应 MUST 可区分「无实付」

### Requirement: gateway-app SHALL 反代并鉴权 cash Admin API

gateway-app-server MUST 将 `/cash/admin/api/*` 反代至 `CASH_SERVICE_URL` 指向的 cash-service。`IsGatewayAdminAPIPath` MUST 将 `/cash/admin/api/` 前缀视为须 Admin JWT 的管理 API。通过 Admin JWT 后，网关 MUST 向下游注入 `X-Admin-Password`（优先 `CASH_ADMIN_PASSWORD`，否则回退 `GATEWAY_APP_ADMIN_PASSWORD`）。浏览器客户端 MUST NOT 被要求自带 `X-Admin-Password`。device-service 及其它非 cash 进程 MUST NOT 直查 `ai_voice_cash` 以提供本列表。

#### Scenario: Hub 已登录调用 Admin API

- **WHEN** 运维持有效 Admin JWT 请求 `GET /cash/admin/api/vip/entitlements`
- **THEN** gateway-app MUST 反代至 cash-service 并注入口令头，cash-service MUST 可完成鉴权

#### Scenario: 无 Admin JWT

- **WHEN** 未携带有效 Admin JWT 请求 `/cash/admin/api/vip/entitlements`
- **THEN** gateway-app MUST 拒绝该请求（不得当作普通 App Bearer 用户接口放行）

### Requirement: cash Admin API SHALL 不计入 App usage 统计

`/cash/admin/api/*` MUST 视为运维管理通道，MUST NOT 按 App 对外接口计入 usage 统计策略；实现 MUST NOT 仅为该 Admin 路径修改 App 侧 `maintenance_skip` 假定。

#### Scenario: 管理端拉取列表

- **WHEN** 运维通过 Hub 调用 VIP 权益 Admin API
- **THEN** 系统 MUST NOT 将其记为需计入的 App 功能使用（与既有 voice/sim admin 管理通道一致）
