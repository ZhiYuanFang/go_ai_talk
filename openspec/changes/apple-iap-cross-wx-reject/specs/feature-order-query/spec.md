## MODIFIED Requirements

### Requirement: 功能订单按交易号查空 MUST 不阻断履约

`FulfillFeaturePaid`（及小智 MCP 的 `FulfillXiaozhiMcpPaid`）在带有渠道交易号时 MUST 先查是否已有同渠道同交易号的已支付单。查无行时 MUST 视为尚未履约，MUST 继续按 `order_no` 加载订单并履约。MUST NOT 因该次空查询向调用方返回 `sql.ErrNoRows` 或导致支付宝回调失败。

当已存在同 `channel` 与 `channel_txn_id` 的 `paid` 功能/MCP 订单时：

- 若 `channel=apple_iap` 且已履约单的 `wx_id` 与本次目标订单 `wx_id` **不同**：MUST 拒绝履约，错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」，MUST NOT 将目标单标为 `paid`，MUST NOT 向目标 `wx` 授功能或 MCP 能力。
- 否则（同 `wx_id`，或非 `apple_iap`）：MUST 幂等成功且 MUST NOT 重复开通。

#### Scenario: 首次支付宝回调功能单

- **WHEN** 功能订单状态为 `created`，支付宝以新的 `trade_no` 回调，且库中尚无该 `channel_txn_id`
- **THEN** 系统 MUST 将该单置为 `paid` 并完成功能开通，MUST NOT 因「按交易号查无」失败

#### Scenario: 同交易号重复回调（同 wx 或非 Apple）

- **WHEN** 同一 `channel` 与 `channel_txn_id` 已对应一笔 `paid` 功能订单，且（`channel` 非 `apple_iap`，或目标订单 `wx_id` 与已履约单相同）
- **THEN** 系统 MUST 幂等成功且 MUST NOT 重复开通

#### Scenario: Apple 同交易号异 wx

- **WHEN** `channel=apple_iap` 且同一 `channel_txn_id` 已对应账号 A 的 `paid` 功能或小智 MCP 订单，账号 B 的订单对该交易号发起履约
- **THEN** 系统 MUST 拒绝，错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」，MUST NOT 为 B 开通
