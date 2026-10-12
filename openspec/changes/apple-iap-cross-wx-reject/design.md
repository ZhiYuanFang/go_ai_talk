## Context

Apple IAP（VIP / 功能 / 小智 MCP）共用 `channel=apple_iap` + `channel_txn_id=transactionId` 幂等。三处履约函数在「同渠道交易号已 paid」时一律 `return nil`，不比较目标订单的 `wx_id`。换 App 账号后复用同一收据时，verify 接口成功但新账号无权益，UX 误导。

支付宝每笔 `trade_no` 独立，不预期出现该场景；权威 ASN 按 `appAccountToken` 命中原订单，通常只幂等原 `wx`。

## Goals / Non-Goals

**Goals:**

- `channel=apple_iap` 下，已 paid 的 `channel_txn_id` 若归属其他 `wx_id`，拒绝履约并返回固定中文错误。
- 同 `wx_id` 重复 verify / ASN 双路径仍幂等成功。
- VIP、功能、小智 MCP 行为一致（公共判断，三处接入）。

**Non-Goals:**

- 不新建 Apple 交易绑定表；不做 `originalTransactionId` 全局硬绑（首版）。
- 不改支付宝文案/逻辑分支（可仅在 `apple_iap` 上启用异号拒绝）。
- 不改 Flutter；不改 ASN 通知 URL / 建单契约 / `appAccountToken` 生成。
- 不堵住「原单尚未 paid 时异号先 verify」的极短窗口（已知可接受）。

## Decisions

### 1. 判断落点：Fulfill 公共 helper，而非仅 VerifyAppleIAP

- **选择**：在 `FulfillPaid` / `FulfillFeaturePaid` / `FulfillXiaozhiMcpPaid` 的「按 channel_txn 查已 paid」分支中调用同一 helper（例如比较已履约单 `wx_id` 与当前 `orderNo` 对应订单的 `wx_id`）。
- **理由**：verify、ASN、`DispatchFulfillPaid` 都汇入 Fulfill；只改 verify 会漏其它入口。
- **备选**：仅在 `VerifyAppleIAP` 前置检查 → 否决（覆盖不全）。

### 2. 仅 `channel=apple_iap` 启用异号拒绝

- **选择**：`channel != apple_iap` 时保持现有「已 paid → return nil」。
- **理由**：产品确认支付宝不会出现该场景；避免误改支付宝幂等语义。
- **备选**：全渠道异号都拒 + Apple 文案 → 可做但非必要。

### 3. 同号 / 异号判定键

- **选择**：已 paid 订单的 `wx_id` 与**本次履约目标订单**（`orderNo`）的 `wx_id` 比较；相同或目标单即已 paid 本单 → 幂等成功；不同 → 拒绝。
- **理由**：Fulfill 入参无调用方 header wx，订单行上的 `wx_id` 即授权主体；与「订单不属于当前账号」校验一致。
- **备选**：额外传入 header `wxID` → 多余，目标订单已绑定 wx。

### 4. 错误码与文案

- **选择**：`gerror.NewCode` 使用现有业务码（建议 `CodeInvalidOperation` 或 `CodeNotAuthorized`），`message` 固定为：`一个 Apple 账号仅支持开通一个胖宝账号`。
- **理由**：App 展示 `message`；无需新错误码表。不暴露归属方 `wx_id`。
- **备选**：独立业务 code 供客户端分支 → 本仓无 Flutter 改动需求，首版不做。

### 5. MCP 与 Feature 双入口

- **选择**：helper 同时用于 `FulfillFeaturePaid` 与 `FulfillXiaozhiMcpPaid`；MCP 经 Feature 分流进入时只在 MCP 函数内判断一次即可（Feature 早退在分流前，故两处早退都要接 helper）。
- **理由**：当前两处都有独立的 channel_txn 早退。

## Risks / Trade-offs

- [Risk] 客户端仍以 StoreKit 成功为准、忽略 verify 错误 → 仍假成功。**Mitigation**：proposal 假定现网 toast `message`；验收时用换号场景确认；若否再开 Flutter 任务。
- [Risk] 原单未 paid 时异号先 verify 仍可抢授。**Mitigation**：首版接受；后续可加 `originalTransactionId` 绑定。
- [Risk] 同 wx 多笔 `created` 订单共用一笔 txn 时：第二单异 orderNo 但同 wx → 应按同 wx 幂等成功，不得授第二笔权益。**Mitigation**：helper 仅比较 wx；已 paid 早退不写第二单。
- [Risk] ASN 对已退款单等特殊状态。**Mitigation**：不改变现有 refunded skip；仅替换「已 paid → nil」中的异号分支。

## Migration Plan

1. 部署 cash-service（含公共 helper）。
2. 沙箱：账号 A 开通 → 换账号 B 用同一 txn verify → 应收到上述 `message`，B 的 `xiaozhi_mcp_entitlement` / VIP / feature 无新增。
3. 账号 A 重复 verify → 仍成功。
4. 回滚：回退该版本即可恢复「异号空成功」（无数据迁移）。

## Open Questions

- 无阻塞项。若验收发现客户端未展示 `message`，另开客户端 change。
