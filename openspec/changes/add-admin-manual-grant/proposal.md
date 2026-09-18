## Why

运维需要在 Hub 手工授予 VIP 或商业功能权益：一是用户已付款但因服务端缺陷未履约时的补开通；二是推广期向合作伙伴赠送。现有 VIP 列表与开通功能页均为只读/配置向，缺少带审查理由的写路径。

## What Changes

- cash-service 新增 Admin 手工授 VIP / 授功能 API（Hub JWT + 下游 `X-Admin-Password`，复用既有 `/cash/admin/api/*` 反代）。
- 授予语义与支付续期一致：`durationDays` → `max(now, 当前到期) + N 天`（功能按 `activation_subject` 写 device/user 权益或预测数量增量）。
- **复用订单表**落审查记录：插入 `vip_order` / `feature_order`，`channel=admin`、`amount_fen=0`、`status=paid`；两表新增 **`grant_reason`** 列，表单 **必填** 授权理由并写入该列。
- Hub：`cash-vip-admin.html` 增加授 VIP 表单；`cash-feature-admin.html` 增加授功能表单（按主体填 wxId 或 deviceNo）。
- **不**做：验签失败自动退款、App 支付契约变更、伪造真实支付金额、缩短/撤销 UI（本期仅授予/续期）。

## Capabilities

### New Capabilities

- `cash-admin-manual-grant`: Admin 手工授 VIP/功能 API、订单审查落库（含 `grant_reason`）、Hub 表单与列表展示约定。

### Modified Capabilities

- （无基线独立 capability 文件需 MODIFIED；实现时更新既有「VIP 权益页只读」行为——该要求仅存在于已归档 change `vip-admin-console`，本变更以新 capability 覆盖写路径。）

## Impact

- **进程**：`cash-service`（EnsureSchema 加列、Admin 写接口、履约复用 `ExtendEntitlement` / `Grant*`）；`gateway-app` 静态页（无需新反代前缀）。
- **库**：`ai_voice_cash.vip_order`、`feature_order` 增 `grant_reason`；写入 admin 订单行 + 权益表。
- **API**：新 `POST` Admin 路径（精确 path，Admin JWT）；不改 App v1 结构。
- **usage**：Admin 通道，不计入 App usage；勿为 Admin 改 `maintenance_skip`。
- **非目标**：支付宝/Apple 退款自动化、Flutter UI、专用授记表（否决，采用订单表方案 A）。
