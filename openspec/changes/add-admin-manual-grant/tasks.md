## 1. Schema

- [x] 1.1 `EnsureSchema`：`vip_order` / `feature_order` 幂等增加 `grant_reason VARCHAR(256) NOT NULL DEFAULT ''`（含中文注释）
- [x] 1.2 确认支付建单路径写入空 `grant_reason`（或不写依赖默认）

## 2. VIP 手工授

- [x] 2.1 实现 Admin 授 VIP 服务：校验 reason/wxId/durationDays；事务内插 `channel=admin` paid 订单 + `ExtendEntitlement`
- [x] 2.2 `api/v1` + controller：`POST /cash/admin/api/vip/entitlements/grant`（Admin 口令校验）
- [x] 2.3 VIP 权益列表 JOIN 最近 paid 订单时带回 `grantReason`（可选但推荐，便于审查）

## 3. 功能手工授

- [x] 3.1 实现 Admin 授功能：按 `activation_subject` 校验 wxId/deviceNo；事务内插 admin `feature_order` + 既有 Grant 路径
- [x] 3.2 `api/v1` + controller：`POST /cash/admin/api/feature/grants`
- [x] 3.3 `product_code`：优先该功能在售 SKU，否则 `admin_<featureId>` 占位

## 4. Hub UI

- [x] 4.1 `cash-vip-admin.html`：授 VIP 表单（wxId、天数、理由必填）+ 提交后刷新；更新页头「只读」文案
- [x] 4.2 `cash-feature-admin.html`：授功能表单（featureId、主体字段、天数/数量、理由必填）+ 刷新快照

## 5. 自检

- [x] 5.1 确认未改 App 支付契约、未引入自动退款、未改 `maintenance_skip`；`channel=admin` 不走支付宝/Apple notify 创建路径
- [x] 5.2 `openspec validate add-admin-manual-grant --strict` 通过
