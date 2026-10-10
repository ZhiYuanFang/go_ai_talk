## Context

小智 MCP 绑定（`xiaozhi_mcp_binding`）已上线：Flutter 与 Hub（`history.html` App 登录）共用 `/device/app/api/xiaozhi-mcp/bindings*`，写成功后通知 `xiaozhi-mcp-service`。当前无付费门禁。

现有商业功能开通（care / growth）走 `feature_def` + `feature_product` + `ActivateFeature`，且 **强制 durationDays≥1、禁止永久**；配置集中在「开通功能管理」。小智 MCP 需要 **wx 维一次性永久能力**，且运维配置/手工授 **完全独立** 于该页。

约束：跨服务禁止 device 直查 cash 库；新 App 接口须确认 usage；不默认加 Redis 读缓存；不新增未批准的背景 ticker。

## Goals / Non-Goals

**Goals:**

- wx 账号一次开通永久「可添加小智绑定」能力（支付或 Admin 手工授）。
- 仅 Add 校验；Flutter 与 Hub 同一服务端闸门。
- 支付复用现有支付宝/Apple 建单与回调，履约处分 MCP 类型分支。
- SKU 专用表 + 独立小 Admin（一页两区：SKU + 手工授/撤销）。
- device 经 `clients/cash` 内部接口查开通态。

**Non-Goals:**

- 不把 MCP 纳入开通功能管理、`feature/catalog`、邀请码、试用、VIP 旁路。
- 不改 List/Alias/Delete 语义；不自动迁移/grandfather 存量绑定。
- 不在本仓改 Flutter UI（仅契约）；不做按音箱计费。
- 不为开通态新增 Redis 缓存（首版直查 DB）。

## Decisions

### 1. 权益与 SKU 专用表（不用 `feature_product` / `feature_user_entitlement`）

- **选择**：`xiaozhi_mcp_product`（种子单行永久 SKU：price、apple_product_id、status…）+ `xiaozhi_mcp_entitlement`（`wx_id` 唯一；永久有效；`unlock_method`=`payment`|`admin`；`channel_ref`；可 `status`/撤销时间）。
- **理由**：与「禁止永久」的 feature 域隔离；开通功能管理天然看不到；符合「SKU 专用表」。
- **备选**：复用 `feature_product` 并过滤 Admin → 易误入 catalog/收益 JOIN，否决。

### 2. 订单复用 `feature_order`，履约按 product 类型分支

- **选择**：MCP 建单仍写入 `feature_order`（`product_code` = MCP 种子码，金额取自 `xiaozhi_mcp_product`），以便共用 `DispatchFulfillPaid` / Apple 验单 / 支付宝回调。
- 履约：`FulfillFeaturePaid`（或前置分流）先查 `xiaozhi_mcp_product`；命中则 `GrantXiaozhiMcpPermanent`，**禁止**调用 `ActivateFeature`；未命中走原 feature 逻辑。
- Apple `productId` 反查：在既有 feature 反查旁增加 MCP product 查找。
- **理由**：满足「通道一致、多一个类型判断」；少一张订单表。
- **备选**：独立 `xiaozhi_mcp_order` → 回调三路分流，改动面更大。

### 3. Add 门禁位置与跨域

- **选择**：`device` 的 `AddXiaozhiMcpBinding`（或 App Add 控制器）在写库前调用 `clients/cash` 内部接口（如 `GET /cash/internal/api/xiaozhi-mcp/entitlement`）；未开通返回明确业务错误。
- cash fail → fail-closed（不可 Add）。
- **理由**：绑定属 device；权益属 cash；符合服务边界。

### 4. App 契约（独立于 feature catalog）

- `GET /cash/app/api/xiaozhi-mcp/unlock`：`unlocked` + 可售 SKU（productCode、价格、appleProductId；未上架则无 products）。
- `POST /cash/app/api/xiaozhi-mcp/orders`：建单（channel 同功能开通）；后续支付/验单走既有共用入口，靠 `order_no` 履约。
- **usage**：实现前向负责人确认是否计入；未确认不改 `maintenance_skip.go`。

### 5. 独立小 Admin：一页两区

- 静态页：`/device/admin/cash-xiaozhi-mcp-admin.html`；`admin-modules.js` 导航「小智 MCP 开通」；`admin_static_pages.go` 登记。
- **区 A — SKU**：读/改种子商品（价格、原价、Apple 商品 ID、上下架）；`product_code` 只读。
- **区 B — 手工授**：`wxId` + `grant_reason` → 永久授；可选「撤销最近一笔手工授」（对齐 feature 语义：仅最近 admin 且仍有效）。
- Admin API 前缀建议：`/cash/admin/api/xiaozhi-mcp/product`、`.../grants`、`.../grants/revoke`（口令同其它 cash Admin）。
- **不**改 `cash-feature-admin.html`。

### 6. Hub 绑定页

- 服务端闸门已覆盖 Hub（同一 App Add）。
- UX：`history.html` 小智区展示当前登录 wx 是否已开通；未开通禁用 Add 并提示去付费或运维在小 Admin 手工授（可选增强，任务中列为 SHOULD）。

### 7. 退款与撤销

- Apple/渠道退款：若订单为 MCP product，撤销该 wx 永久权益（绑定行保留，仅阻断后续 Add）。
- Admin 撤销：仅最近一笔 `unlock_method=admin` 且仍有效时可撤。

## Risks / Trade-offs

- [Risk] `feature_order` 混入 MCP 单，开通功能管理收益统计可能把 MCP 算进 feature → **Mitigation**：收益 SQL/展示按 `product_code IN (mcp)` 排除或单独分类；Admin 收益页若受影响则加过滤。
- [Risk] 存量已绑未付费用户无法再 Add → **Mitigation**：proposal 明确首版不 grandfather；运维可手工授。
- [Risk] device→cash 内部调用延迟/故障导致无法绑 → **Mitigation**：fail-closed + 明确错误文案；依赖既有 internal 密钥与超时惯例。
- [Risk] 永久权益误授难追 → **Mitigation**：手工授必填理由落库；支持撤销最近 admin 授。

## Migration Plan

1. cash `EnsureSchema`：建 `xiaozhi_mcp_product` / `xiaozhi_mcp_entitlement`，种子一笔永久 SKU（`apple_product_id` 空，价可由运维填）。
2. 部署 cash（履约分支 + Admin/App/internal API）→ 部署 device（Add 门禁）→ 发布 Admin 静态页与导航。
3. 运维在小 Admin 填写 Apple 商品 ID 与价格后，Flutter 再开放购买。
4. 回滚：关掉 Add 门禁或下架 SKU 仅停售；已授权益保留除非显式撤销。

## Open Questions

- 新 App 接口（unlock 查询、建单）是否计入 usage —— **须负责人确认**。
- 收益统计是否要单独展示「小智 MCP」收入行（首版可仅排除出 feature 汇总）。
