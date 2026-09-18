## Why

手工授 VIP/功能已经能落库，但授错或赠送结束后无法从列表收回，权益会一直挂到自然到期。开通功能管理页同时堆着定义、SKU、二维码、VIP 套餐和快照，运维找不到当前要改的那一块；已下架的「预测槽位开通」仍留在种子、履约和表单里，继续占位置。

## What Changes

- 权益列表对「最近一笔 paid 且 `channel=admin`、且仍有效」的行提供撤销。撤销将该主体该项权益 `expire_at` 设为当前时间（直接过期），并把对应 admin 订单标为已撤销。付费/邀请/Apple 渠道不可撤销。
- 确认文案必须说明：权益只有一根到期时钟，直接过期会一并清掉此前叠加上去的付费剩余。
- Hub 开通页改为顶部平铺：VIP、各上架功能、群二维码。点选后只展示该项的配置、列表、手工授与撤销。VIP 列表不再单独成页。
- **BREAKING（仅 Admin 静态页与内部履约分支）**：删除 `prediction_unlock` 的种子、特判与 Admin「增加可看数量」表单。不 DROP 历史表；不改 App 支付 API 结构。

## Capabilities

### New Capabilities

- `cash-admin-grant-revoke`: Admin 仅可撤销手工授，语义为权益立即过期并标记 admin 订单已撤销。
- `cash-admin-hub-tiles`: 开通管理页顶部平铺 VIP / 功能 / 群二维码，点选后展示该项内容。
- `remove-prediction-slot-unlock`: 移除已下架预测槽位开通的种子与死代码，保留历史表。

### Modified Capabilities

- （无。`openspec/specs/` 下没有独立的 cash-admin / prediction 能力规格；`add-admin-manual-grant` 仍停在 changes，本期以新能力覆盖撤销与 Hub 信息架构。）

## Impact

- **进程**：`cash-service`（撤销 API、订单状态、权益过期、去掉 `prediction_unlock` 分支）；`gateway-app` 静态页 `cash-feature-admin.html`，`cash-vip-admin.html` 收进 VIP 面板或改为跳转。
- **库**：不新增审查表；admin 订单使用新状态（不用 `refunded`）。不 DROP `feature_allowed_count` 或历史订单。
- **API**：新增 Admin POST 撤销路径（精确 path，既有 Admin 口令）。不改 App v1/v2 支付契约，不改 `maintenance_skip`。
- **非目标**：按天数回退（`ShrinkEntitlement`）、验签失败自动退款、支付宝/Apple 退款、把撤销记入 App usage。
