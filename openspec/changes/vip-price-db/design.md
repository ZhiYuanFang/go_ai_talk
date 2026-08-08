## Context

`vip-cash-service` 已提供 `GET /cash/app/api/vip/product`，但仅返回 `priceFen`，且要求登录。产品需要未登录可读现价+原价；定价以 `vip_product` 为准，运维手工 SQL 改价，无 Admin 页、无 env 定价。

## Goals / Non-Goals

**Goals:**

- 表字段承载原价；读价 API 返回现价+原价。
- 匿名可访问 product GET；建单仍用库内现价。
- DDL / EnsureSchema / 种子同步。

**Non-Goals:**

- 改价后台、env 控价、ASC API 同步、退款、自动续签。

## Decisions

### D1：原价列 `original_price_fen`

- `INT NOT NULL DEFAULT 0`：`0` = 客户端不展示划线；`>0` 为展示原价（分）。
- 现价继续 `price_fen`；支付宝 `CreateOrder` / `BuildAlipayAppPayOrderStr` 只用 `price_fen`。
- 种子：现价 1900；原价默认 1900 或更高展示价（如 9900）——实现取 **9900** 便于划线演示，运维可 SQL 改。

### D2：匿名读价

- 路径仍为 `GET /cash/app/api/vip/product`。
- controller：**不再**强制 `X-Internal-Wx-Id`（仅此接口）。
- gateway：`gatewayAppAuthExemptExactGETHEAD`（或等价）加入精确 path。
- 其它 `/cash/app/api/*` 仍须 Bearer。

### D3：写价 = 手工 SQL

- 文档/runbook 给示例：
  `UPDATE vip_product SET price_fen=?, original_price_fen=?, updated_at=UNIX_TIMESTAMP() WHERE product_code='vip_monthly_19';`
- 无 Admin API。

### D4：Apple

- 不改 StoreKit 价；运维保证库内 `price_fen` 与 ASC 标价心理一致即可（无系统校验）。

## Risks / Trade-offs

- [匿名刷读] → 只读商品元数据，可接受；若滥用再加 CDN/限流。
- [SQL 改错价] → 运维责任；无审批流（非目标）。
- [原价 < 现价] → 不做强校验；客户端可自行忽略不合理划线。

## Migration Plan

1. 对 `ai_voice_cash` 执行加列（EnsureSchema `ALTER` 或 `hack/ddl_cash_vip_original_price.sql`）。
2. 部署 cash-service + gateway-app（白名单）。
3. 回滚：去掉白名单并恢复登录校验；列可保留。

## Open Questions

- usage 是否计入匿名 product GET——实现前问负责人；未答复不改 `maintenance_skip.go`。
