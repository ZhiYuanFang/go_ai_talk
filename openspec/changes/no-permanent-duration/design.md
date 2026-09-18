## Context

VIP 天数存在 `vip_product.duration_days`。付款成功时 `fulfill` 读该列，经 `ExtendEntitlement` 加到 `vip_entitlement.expire_at`。订单表不存天数。`EnsureSchema` 的 `ON DUPLICATE KEY UPDATE` 每次启动把 `duration_days` 写成常量 30，运维在 Hub 改的值会被盖掉。`ExtendEntitlement` 在天数 ≤0 时静默换成 30。

功能侧 `feature_product.duration_days` 与 `feature_def` 的邀请/广告天数，0 被履约写成 `expires_at=0`（永久）。Hub 文案写着「0=永久」。产品已定：所有功能都不支持永久，采用方案 A（拒绝 0，不实现永久哨兵）。

## Goals / Non-Goals

**Goals:**

- 已有 VIP 商品行的 `duration_days` 在重启后保持运维值。
- VIP、功能 SKU、邀请/广告授予天数，以及对应履约，天数必须 ≥1。
- Hub 不再提供「0=永久」。

**Non-Goals:**

- 把 `vip_monthly_19` 改成 `vip_month`。
- 用 `expire_at=0` 表示永久，或改 `IsVip` 哨兵。
- 批量把历史 `expires_at=0` 改成过期。这些行仍按现有读模型视为有效，直到另一次变更处理。
- 停止覆盖 VIP 标题或强制 `status=1`。试用按小时，不在本次天数校验内。

## Decisions

### D1：种子只在插入时写 VIP 天数

- 从 `vip_product` 的 `ON DUPLICATE KEY UPDATE` 去掉 `duration_days=VALUES(duration_days)`。
- 首次插入仍写 30。已有行不更新该列。与 `price_fen` 不覆盖的做法一致。
- **备选**每次启动按 Hub 值回写：没有第二数据源，否决。

### D2：方案 A，拒绝小于 1 的天数

- Admin 保存 VIP 商品、功能 SKU、功能定义（`duration_days` / `invite_duration_days` / `ad_duration_days`）时，任一授予天数 `<1` MUST 返回参数错误。
- `ExtendEntitlement` 与功能 `upsert*Entitlement`：入参 `<1` MUST 返回错误，MUST NOT 替换成 30，MUST NOT 把 `expires_at` 写成 0。
- 邀请/广告履约读到定义天数为 0 时 MUST 拒绝开通，而不是授予永久。
- **备选**方案 B（0=永久）：否决。产品明确所有功能都不支持永久。

### D3：历史永久行不动

- 不跑 `UPDATE ... SET expires_at=now WHERE expires_at=0`。
- 读路径仍把已有 `expires_at=0` 当成未过期，避免本次上线突然收回已发出的权益。
- 新写入不再产生这种行。

### D4：Hub

- VIP「有效天数」、功能 SKU「有效天数」、功能定义「邀请/广告授予天数」的 `min` 改为 1。
- 去掉「0=永久」文案。

## Risks / Trade-offs

- [库里邀请天数已是 0，兑码失败] → 运维把该功能邀请天数改成 ≥1 后再兑；不自动改库。
- [历史永久用户仍有效] → 符合「不批量收回」；若以后要收回，另开变更。
- [标题仍被种子覆盖] → 本次不做，避免和天数修复绑在一起。

## Migration Plan

1. 部署 cash-service（种子不再覆盖天数；校验与履约拒绝 `<1`）。
2. 部署静态页。
3. 回滚：恢复覆盖与 0=永久会重新引入两个坑；已保存的 ≥1 天数可留在库里。

## Open Questions

- 无。方案 A，且覆盖 VIP 与全部功能授予天数。
