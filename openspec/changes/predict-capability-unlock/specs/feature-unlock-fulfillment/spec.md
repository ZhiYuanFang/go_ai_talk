## REMOVED Requirements

### Requirement: 广告完成开通 MVP MUST 信任客户端申报

**Reason**: 功能开通域广告通道已下线；客户端申报不可再授予权益。

**Migration**: App 停止调用 `POST /cash/app/api/feature/ad/complete`；若仍调用，服务端返回明确错误且不授权益。改用 payment / invite_code / 点试用（预测）/ VIP。

## MODIFIED Requirements

### Requirement: Admin MUST 可配置功能的邀请/广告授予天数

开通功能管理中的功能定义编辑 MUST 提供「邀请授予天数」表单字段并持久化到功能定义的邀请天数字段。天数 MUST ≥1（不支持永久邀请授予）。该字段 MUST NOT 被解释为支付套餐时长（支付仍以 SKU 为准）。Admin MUST NOT 再将广告授予天数作为可配置开通通道参数；保存时 MUST NOT 接受或持久化将 `ad` 纳入 `unlock_methods` 的配置。

#### Scenario: 运维修改邀请天数后兑码

- **WHEN** 管理员将某权益型功能的邀请天数从 7 改为 3 并保存，之后用户首次邀请开通该功能
- **THEN** 新授予 MUST 按 3 天计算到期

#### Scenario: 支付套餐独立

- **WHEN** 功能定义邀请天数为 7 且某支付 SKU `duration_days=30`
- **THEN** 支付履约 MUST 仍按 SKU 30 天授予，MUST NOT 被邀请天数覆盖

## ADDED Requirements

### Requirement: ActivateFeature MUST 拒绝广告通道

统一开通入口在 `channel=ad` 时 MUST 拒绝并 MUST NOT 写入权益。允许的通道 MUST 包含 `payment`、`invite_code`、`trial`、`admin`（以实现为准）。

#### Scenario: ad 通道被拒

- **WHEN** 调用方以 Channel=`ad` 请求 ActivateFeature
- **THEN** 系统 MUST 返回错误且 MUST NOT 变更权益表
