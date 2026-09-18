## 1. 查询空结果

- [x] 1.1 `loadFeatureOrderByChannelTxn` / `loadFeatureOrderByAppAccountToken`：无行返回 `nil, nil`，不返回 `sql.ErrNoRows`
- [x] 1.2 `loadFeatureOrderByNo`：无行返回明确「功能订单不存在」，不返回 `sql.ErrNoRows`
- [x] 1.3 `GetActiveFeatureProduct` / `GetFeatureProductByAppleID`：无行返回明确业务错误，不返回 `sql.ErrNoRows`
- [x] 1.4 `DispatchFulfillPaid` 分流探测与上述空结果语义一致（可选改为 `.One()`）

## 2. 自检

- [x] 2.1 确认未改 VIP 履约、未改 App 支付字段、未改订单号前缀
- [x] 2.2 `openspec validate fix-feature-alipay-fulfill-scan --strict` 通过
