## Context

`cash-service` 已具备 VIP / 功能支付履约（`ExtendEntitlement`、`GrantEntitlementOrCount` / user 维 Grant）与 Hub Admin 只读 VIP 列表、开通功能配置页。网关已反代 `/cash/admin/api/*` 并注入 `X-Admin-Password`。订单表 `vip_order` / `feature_order` 无理由列；权益表只反映当前状态，不适合单独作审查流水。

产品用途：付款未开通的人工补发；推广赠送合作伙伴。探索结论锁定：复用订单表（方案 A）+ 新建 `grant_reason` 必填列。

## Goals / Non-Goals

**Goals:**

- Admin 可对指定 `wxId` 授 VIP（`durationDays` 续期语义）。
- Admin 可对指定功能授权益：按 `feature_def.activation_subject` 要求 `wxId` 或 `deviceNo`；预测数量类支持数量增量。
- 每次授予插入对应 `channel=admin` 的 paid 订单，并写入非空 `grant_reason`。
- Hub 两页提供表单（理由必填）；列表/快照可识别 admin 渠道。

**Non-Goals:**

- 验签失败自动退款；支付宝/Apple 退款 OpenAPI。
- 专用 `admin_grant_log` 表；缩短/撤销权益 UI。
- 修改 App 支付 API 或 Flutter。
- 将 Admin 授记计入 App usage。

## Decisions

### D1：审查落库 = 订单行 + `grant_reason` 列

- `vip_order` / `feature_order` 均 `ALTER`/`EnsureSchema` 增加 `grant_reason VARCHAR(256) NOT NULL DEFAULT ''`。
- Admin 授予：`channel=admin`、`amount_fen=0`、`status=paid`、`paid_at=now`、`grant_reason=trim(reason)`；支付单 `grant_reason` 保持空串。
- **不**把理由写入 `channel_txn_id`（避免唯一键与语义污染）；admin 单 `channel_txn_id` 可用幂等短码（如 `admin:{orderNo}`）或空串（若唯一键允许空重复则需注意 MySQL 多空串行为——建议写入唯一 `admin-{orderNo}`）。
- **备选**专用授记表：否决（产品选 A，改动面小）。

### D2：履约复用现有函数

- VIP：`ExtendEntitlement(ctx, wxID, durationDays)`，与支付一致。
- 功能：按主体调用既有 user/device Grant；`grant_kind=allowed_count_delta` 时授数量而非天数（表单字段区分）。
- 先写订单再履约，或同一逻辑块内顺序写；失败返回错误且勿返回 success（允许订单已插入时需可观测日志；实现优先「订单+履约」同函数内成功才返回，履约失败则标记订单或回滚——**推荐单事务**：插订单 + 写权益）。

### D3：API 形状

- `POST /cash/admin/api/vip/entitlements/grant`：body `wxId`、`durationDays`（≥1）、`reason`（必填）。
- `POST /cash/admin/api/feature/grants`：body `featureId`、`reason`（必填）、按主体的 `wxId`/`deviceNo`、`durationDays` 和/或 `grantQuantity`。
- `product_code`：VIP 用现有默认商品（如 `vip_monthly_19`）；功能用该 `feature_id` 下一件 `status=1` 的 SKU，若无则 `admin_<featureId>` 占位码（Ensure 或插入时允许非外键，订单表本无 FK）。
- 鉴权：与其它 cash Admin 相同，校验 `X-Admin-Password`。

### D4：Hub UI

- `cash-vip-admin.html`：去掉「仅只读」文案中与手工开通冲突的绝对表述；增加授 VIP 表单。
- `cash-feature-admin.html`：增加授功能表单；提交后刷新开通快照。
- 浏览器仅 Admin JWT；理由客户端非空校验 + 服务端再校验。

### D5：列表展示

- VIP 权益列表「渠道」对最近 paid 为 admin 时显示 `admin`；金额 0 显示为「—」或「0（手工）」——实现选「渠道=admin 时金额文案可为手工授」。
- 可选：列表增加 `grantReason` 仅当最近订单为 admin 时返回（JOIN 已有最近 paid 订单即可带出 `grant_reason`）。

## Risks / Trade-offs

- [admin 订单混入「最近实付」JOIN] → 金额为 0；UI 按 channel 区分，避免显示假实付。
- [无操作员账号字段] → 一期仅 `grant_reason`；操作员依赖 Hub 登录审计（网关日志），不阻塞。
- [履约失败留下 paid 订单] → 事务绑定或失败时订单 status 勿标 paid；实现选事务。
- [伪造 notify 改 admin 单] → 外部回调按 out_trade_no 履约；admin 单号前缀与支付单区分，且无支付宝流水；既有金额校验对 0 元需确认不会被恶意 notify 重放——Dispatch 仍校验订单存在与金额；admin 单勿接受外部 channel 覆盖。

## Migration Plan

1. 部署含 `EnsureSchema` 加列的 cash-service（幂等 ADD COLUMN）。
2. 部署静态页；无需新 env。
3. 回滚：停用 Admin 写接口即可；列可保留。

## Open Questions

- 无（`durationDays=0` 永久授 VIP：本期 **不允许**，VIP `durationDays` MUST ≥1；功能永久按既有 Grant `durationDays=0` 语义仅当产品明确需要时在功能 API 允许——默认功能授亦要求 ≥1 天，预测增量用 `grantQuantity`）。
