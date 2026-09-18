## Context

`add-admin-manual-grant` 已落地：Hub 可授 VIP/功能，写入 `channel=admin` 的 paid 订单和 `grant_reason`，续期走 `ExtendEntitlement` / `ActivateFeature`。权益表只有一根 `expire_at`，不是按笔叠加。退款回退 `ShrinkEntitlement` 按天数扣减，但手工授的天数没有落在订单上。

开通页 `cash-feature-admin.html` 一页堆着定义、SKU、群二维码、VIP 套餐和快照；VIP 列表在 `cash-vip-admin.html`。`prediction_unlock` 已在 EnsureSchema 里 `status=0`，catalog 会跳过，但种子、履约特判和 Admin「增加可看数量」仍在。

## Goals / Non-Goals

**Goals:**

- 仅当该主体该项的最近一笔 `status=paid` 订单为 `channel=admin` 且权益仍有效时，允许撤销。
- 撤销把对应 `expire_at` 设为当前时间，并把该笔 admin 订单标为 `revoked`。
- Hub 顶部平铺 VIP、上架功能、群二维码；点选后只渲染该项。
- 删除 `prediction_unlock` 代码路径与 Admin 表单项；不 DROP 历史表。

**Non-Goals:**

- 按天数回退；把 `durationDays` 补写进旧订单。
- 撤销付费、邀请码、广告、试用、Apple 退款。
- 验签失败自动退款。
- DROP `feature_allowed_count` 或历史 `feature_order`。
- 改 App 支付 API、Flutter、`maintenance_skip`。

## Decisions

### D1：撤销 = 立即过期，不按天数回退

- 服务端把 `vip_entitlement.expire_at`（或功能权益 `expires_at`）写成 `now`。`IsVip` / 功能有效判断立刻为 false。
- **不**调用 `ShrinkEntitlement`。天数未落库，扣天数会猜错。
- 确认框必须写明：此前叠加上去的付费剩余也会一起失效。这是产品已接受的代价。
- **备选**补列 `duration_days` 再 Shrink：否决。用户要求直接过期，且旧单无法回填天数。

### D2：资格以最近一笔 paid 订单为准，服务端再校验

- VIP 列表已 JOIN 最近 paid 订单的 `channel`。行上 `channel=admin` 且未过期才显示撤销。
- 功能快照行：`unlock_method=admin` 且仍有效才显示撤销。服务端按 `featureId` + 主体找到最近一笔 `feature_order.status=paid`，MUST 为 `channel=admin`，否则拒绝。
- 付费后又手工续上时，最近一笔是 admin，撤销会清掉付费剩余。这与 D1 一致，不另做「底下还有付费则禁止」。
- 手工授之后又付费：最近一笔不是 admin，按钮不出现，服务端拒绝。此前手工天数留在时钟里，本期不拆。

### D3：订单状态用 `revoked`，不用 `refunded`

- 新增 `OrderRevoked = "revoked"`。撤销成功后把那笔 admin 订单从 `paid` 改为 `revoked`。
- **不**复用 `refunded`（Apple ASN 退款语义）。
- 列表 JOIN 仍只看 `status=paid`。撤销后最近 paid 若是更早的支付宝单，渠道变回付费；若没有更早 paid，渠道为空，按钮消失。
- 幂等：目标订单已是 `revoked` 或权益已过期 → 返回明确错误，不重复写。

### D4：API

- `POST /cash/admin/api/vip/entitlements/revoke`：body `wxId`。口令与其它 cash Admin 相同。
- `POST /cash/admin/api/feature/grants/revoke`：body `featureId`，以及 `wxId` 或 `deviceNo`（按 `activation_subject`）。
- 不新增 App 路径，不进 usage 统计。

### D5：Hub 信息架构

```
[ VIP ] [ 值得留意 ] [ 成长轨迹 ] [ 群二维码 ]
────────────────────────────────────────────
只渲染当前选中项
```

- 平铺项 = 写死的 VIP + `feature_def.status=1` 的功能 + 群二维码。`prediction_unlock` MUST NOT 出现。
- VIP 面板收进现有 VIP 套餐表单 + `cash-vip-admin.html` 的列表、授表、撤销按钮。`cash-vip-admin.html` 改为跳到开通页并选中 VIP，或保留只读跳转，避免两套可写入口。
- 功能面板：该功能的定义、仅该 `featureId` 的 SKU、开通快照、授/撤。不再一页展示全部功能的编辑表和全部 SKU。
- 群二维码面板保留现有上传与有效期，不并进某个功能。

### D6：删除预测槽位代码，表留下

- 停止 EnsureSchema 插入/更新 `prediction_unlock` 种子（已有行可留，不再每次写回）。
- 删除 `FeatureIDPredictionUnlock` 及 `feature_activate` / `feature_grant` / `feature_order` / `apple_notify` / `feature_admin` / `feature_catalog` / `feature_admin_snapshot` / `feature_admin_rules` 中的特判。
- Admin 去掉 `allowed_count_delta`、默认开通条数、预测数量授予字段。`GrantKindAllowedCountDelta` 若再无调用则删除常量与分支。
- **不** DROP `feature_allowed_count`。历史订单保持原样。

## Risks / Trade-offs

- [付费剩余被手工撤销清掉] → 仅 admin 最近一笔可撤；确认框写明后果。
- [撤销后 JOIN 显示更早的付费单] → 预期行为；订单 `revoked` 仍可在库里审查。
- [删预测分支后旧客户端再买预测 SKU] → 商品已 `status=0`；履约若遇到该 `feature_id` MUST 拒绝并打日志，不要静默当普通权益。
- [静态页缓存] → 页头保留强制刷新提示。

## Migration Plan

1. 部署 cash-service（新状态常量、撤销 API、去掉预测分支）。无需新环境变量。
2. 部署静态页。
3. 回滚：停用撤销接口；`revoked` 行保持，不自动恢复权益。预测代码回滚不影响已停用的种子行。

## Open Questions

- 无。撤销语义、平铺范围、预测表不删，已在探索中定下。
