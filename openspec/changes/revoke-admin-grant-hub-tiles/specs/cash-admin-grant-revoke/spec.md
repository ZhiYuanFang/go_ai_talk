## ADDED Requirements

### Requirement: 系统仅在最近一笔为手工授且仍有效时允许撤销 VIP

cash-service MUST 提供 `POST /cash/admin/api/vip/entitlements/revoke`，鉴权 MUST 校验 `X-Admin-Password`。请求体 MUST 含 `wxId`（>0）。系统 MUST 仅在同时满足以下条件时撤销：

1. 该 `wxId` 的 `vip_entitlement.expire_at` 大于当前时间；
2. 该 `wxId` 最近一笔 `vip_order.status=paid` 的 `channel` 为 `admin`。

否则 MUST 拒绝，MUST NOT 修改权益或订单。

#### Scenario: 最近一笔为手工授且未过期

- **WHEN** 管理员对仍有效、且最近 paid 订单 `channel=admin` 的 `wxId` 提交撤销
- **THEN** 系统 MUST 将该 `expire_at` 设为当前时间，且该笔订单 `status` MUST 变为 `revoked`

#### Scenario: 最近一笔为付费渠道

- **WHEN** 最近 paid 订单 `channel` 为 `alipay` 或 `apple_iap`
- **THEN** 系统 MUST 拒绝撤销，MUST NOT 修改 `expire_at`

#### Scenario: 口令错误

- **WHEN** 请求未携带或携带错误的 `X-Admin-Password`
- **THEN** 系统 MUST 拒绝且 MUST NOT 写库

### Requirement: 撤销 VIP MUST 直接过期且 MUST NOT 按天数回退

撤销成功时系统 MUST 将 `vip_entitlement.expire_at` 设为不大于当前时间，使 `IsVip` 立即为 false。系统 MUST NOT 使用 `ShrinkEntitlement` 或按 `durationDays` 扣减。系统 MUST 使用订单状态 `revoked`，MUST NOT 把该订单标为 `refunded`。

#### Scenario: 付费剩余被一并清掉

- **WHEN** 账号先有付费剩余、随后手工续期，最近 paid 为 `admin`，管理员确认撤销
- **THEN** 系统 MUST 将到期设为当前时间，MUST NOT 只扣手工授的天数

#### Scenario: 重复撤销

- **WHEN** 目标 admin 订单已是 `revoked`，或权益已过期
- **THEN** 系统 MUST 返回错误，MUST NOT 再改写其它 paid 订单

### Requirement: 功能手工授撤销 MUST 按主体校验最近 admin 订单

cash-service MUST 提供 `POST /cash/admin/api/feature/grants/revoke`。请求 MUST 含 `featureId`，并按 `activation_subject` 要求 `wxId` 或 `deviceNo`。系统 MUST 找到该功能该主体最近一笔 `feature_order.status=paid`；仅当其 `channel=admin` 且对应权益仍有效时，MUST 将权益到期设为当前时间，并将该订单标为 `revoked`。`prediction_unlock` 与数量累加类 MUST NOT 提供此撤销。

#### Scenario: 撤销 user 主体功能

- **WHEN** 管理员对仍有效、最近 paid 为 `admin` 的账号维功能提交 `wxId` 与 `featureId`
- **THEN** 系统 MUST 使该账号该功能立即过期，并标记该 admin 订单 `revoked`

#### Scenario: 主体标识缺失

- **WHEN** 功能为 `device` 主体且请求没有 `deviceNo`
- **THEN** 系统 MUST 拒绝，MUST NOT 改权益
