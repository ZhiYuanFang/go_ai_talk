## Context

VIP 履约 `FulfillPaid` 用 `loadOrderByChannelTxn` → `.One()`：无行时 `nil, nil`，继续按 `order_no` 履约。功能履约 `FulfillFeaturePaid` 先调 `loadFeatureOrderByChannelTxn`：`var r featureOrderDB` + `Scan(&r)`。GoFrame 对已初始化 struct 的空 `Scan` 返回 `sql.ErrNoRows`，该错误被直接返回，支付宝拿到 `failure` 并重试，权益未开通。

用户已确认：VIP 正常，功能单有问题。功能建单也用 `newOrderNo()`（`VIP` 前缀），与本次修复无关。

## Goals / Non-Goals

**Goals:**

- 首次支付宝/Apple 回调时，按交易号查不到行 MUST 视为未履约过，继续按 `order_no` 履约。
- 功能订单与 SKU 读路径空结果语义与 VIP 一致，不向外抛裸 `sql.ErrNoRows`。

**Non-Goals:**

- 不改订单号前缀。
- 不改支付金额校验、渠道校验、退款逻辑。
- 不回填历史未履约订单（部署后支付宝重试或人工补履约即可）。

## Decisions

### 1. 查询改为 `.One()` + 映射

`loadFeatureOrderByNo`、`loadFeatureOrderByChannelTxn`、`loadFeatureOrderByAppAccountToken` 改为与 VIP 相同：

```go
one, err := g.DB().Model("feature_order").Ctx(ctx).Where(...).Limit(1).One()
if err != nil { return nil, err }
if one.IsEmpty() { return nil /* 或 NotFound */, nil/err }
```

按 `order_no` 查空：返回明确的「功能订单不存在」。按交易号 / token 查空：返回 `nil, nil`（幂等「未找到已履约单」）。

`GetActiveFeatureProduct`、`GetFeatureProductByAppleID` 同样改用 `.One()`；空结果返回现有中文业务错误，不返回 `sql.ErrNoRows`。

### 2. Dispatch 分流保持忽略空 Scan

`DispatchFulfillPaid` 里对 `feature_order` 的探测已用 `_ = Scan`；可改为 `.One()` + `IsEmpty` 更清晰，行为不变。

### 3. 不改订单号

功能单仍可能生成 `VIP…` 前缀。本次只修空结果语义；前缀另开变更。

## Risks / Trade-offs

- [已卡在 `created` 的失败单] → 支付宝会重试 notify；部署后重试应成功。若不再回调，需运维按 order_no 补履约（本变更不提供新 Admin 工具）。
- [误把真错误吞掉] → 仅把「无行」当空；其它 DB 错误仍向上返回。

## Migration Plan

发版 cash-service 即可。无 DDL。回滚恢复旧 `Scan` 行为。

## Open Questions

无。
