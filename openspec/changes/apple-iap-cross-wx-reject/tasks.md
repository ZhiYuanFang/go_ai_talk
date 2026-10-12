## 1. 公共判断

- [x] 1.1 在 `internal/services/cash` 增加公共 helper：入参含 `channel`、已履约订单 `wx_id`、目标订单 `wx_id`；`channel=apple_iap` 且异 wx 时返回固定文案「一个 Apple 账号仅支持开通一个胖宝账号」；同 wx 或非 apple_iap 返回「幂等放行」语义
- [x] 1.2 为错误文案定义包级常量（中文），避免三处字符串漂移

## 2. 接入三处履约

- [x] 2.1 改造 `FulfillPaid`：`channel_txn` 已 paid 早退前加载目标 `orderNo` 的 `wx_id`（或等价），调用 helper；异号拒绝，同号 `return nil`
- [x] 2.2 改造 `FulfillFeaturePaid`：同上（功能订单）
- [x] 2.3 改造 `FulfillXiaozhiMcpPaid`：同上（小智 MCP；注意经 Feature 分流时两处早退均已覆盖）

## 3. 验收与边界

- [ ] 3.1 人工验收：账号 A Apple 开通成功后，换账号 B 用同一 `transactionId` verify → 业务错误文案正确，B 无权益/订单不 paid
- [ ] 3.2 人工验收：账号 A 重复 verify / 模拟 ASN 后再 verify → 仍幂等成功
- [x] 3.3 确认支付宝回调路径未因本改动改变（非 `apple_iap` 仍旧幂等）
- [x] 3.4 不新增 `*_test.go`（仓库阶段约定）
