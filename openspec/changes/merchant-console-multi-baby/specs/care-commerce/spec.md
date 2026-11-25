# Spec Delta

## ADDED Requirements

### Requirement: 商户权益 SHALL 以单一 expire_at 为权威
系统 MUST 为每个商户维护商户权益记录，权威到期字段为 **`expire_at`**。Hub 手工授期与支付宝连续订阅履约 MUST **共用**该字段。有效判定 MUST 为 `expire_at` 晚于当前时间。`expire_at` 已过或不存在有效权益 MUST 视为欠费/过期，并驱动宝宝数量上限与 history 写闸（见 history-feeding / merchant-console）。

#### Scenario: 无权益视为过期
- **WHEN** 商户无权益行或 `expire_at` ≤ now
- **THEN** `IsMerchantEntitled`（或等价）MUST 为 false

#### Scenario: 未过期为有效
- **WHEN** `expire_at` > now
- **THEN** 商户 MUST 视为有效权益（宝宝不限量）

### Requirement: Hub SHALL 以绝对到期日授予商户权益
Hub/Admin 授期 API MUST 接受选定 `merchantId` 与 **绝对到期时间** `expireAt`（非仅 durationDays）。成功后 MUST 将该商户权益的 `expire_at` 设为所提交值（或文档化的规范化时刻），`source` MUST 可区分 `hub_grant`。浏览器 MUST NOT 发送 `X-Admin-Password`（经既有 Admin 会话）。

#### Scenario: 设置绝对到期日
- **WHEN** 管理员为商户提交未来的绝对 `expireAt`
- **THEN** 系统 MUST 将 `expire_at` 更新为该值且列表展示为未欠费

#### Scenario: 过期日不晚于现在
- **WHEN** 管理员提交 `expireAt` ≤ now
- **THEN** 系统 MUST 拒绝或写入后立即视为欠费（实现须一致且可测）；推荐拒绝非法授期

### Requirement: 商户连续订阅 SHALL 经支付宝网页支付并更新 expire_at
系统 MUST 支持商户通过 **支付宝网页** 开通/续订 **连续订阅**（周期扣款语义），MUST NOT 复用 C 端 VIP「一次性支付后仅 +durationDays」作为本能力的唯一履约模型。默认标价 MUST 为 **1000 元/月**（`price_fen=100000`），且 MUST 允许 Hub 修改该商户商品月费。扣款/签约成功履约 MUST 更新同一 `expire_at`（按计费周期延长或支付通道回调约定，须幂等）。

#### Scenario: 默认月费
- **WHEN** Hub 未改价时商户读取开通商品
- **THEN** 标价 MUST 为 100000 分/月

#### Scenario: 订阅履约延长 expire_at
- **WHEN** 支付宝连续订阅成功通知首次到达某商户订单
- **THEN** 系统 MUST 幂等将订单置已支付并延长或设置该商户 `expire_at`

#### Scenario: 与 VIP 一次性轨隔离
- **WHEN** 商户订阅回调到达
- **THEN** MUST NOT 写入 C 端 `vip_entitlement` 作为本能力履约

### Requirement: 商户权益状态 SHALL 可供 Hub 与商户台读取
系统 MUST 提供只读状态：是否开通、付费来源、`expire_at`、剩余时间或欠费标记。Hub 列表与商户台「我的订阅」MUST 基于同一权威数据。

#### Scenario: 欠费状态可读
- **WHEN** 商户 `expire_at` 已过
- **THEN** 状态 API MUST 标识欠费/过期且 MUST NOT 仍显示有效不限量
