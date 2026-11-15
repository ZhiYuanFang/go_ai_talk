## REMOVED Requirements

### Requirement: cash-service MUST 种子预测能力基础版与 Pro 为人维功能

**Reason**: 本轮仅保留 Pro 开通壳；基础版取消。

**Migration**: 以 `predict-pro-unlock-shell` 为准；Ensure 停用 basic；删除 basic SKU 依赖。

### Requirement: VIP MUST 覆盖预测两档使用权且不改写 catalog unlocked

**Reason**: 推送不再消费预测能力开通；VIP 覆盖推送不再适用。

**Migration**: 推送见 `predict-push-ungate`；VIP 仍可不改写 catalog unlocked（通用规则未变）。

### Requirement: Pro 能力 MUST 包含基础版推送语义

**Reason**: 无基础版；推送不挂钩开通。

**Migration**: 无。

## ADDED Requirements

### Requirement: 原双档变更中的 Pro 开通壳 MUST 收敛保留

在回退基础版与推送门禁后，系统 MUST 仍满足 Pro 人对开通（支付/邀请/7 日试用）与喂养前置，见 `predict-pro-unlock-shell` 与 `predict-pro-feeding-eligibility`。

#### Scenario: Pro 仍可支付开通

- **WHEN** 喂养合格用户支付 Pro SKU 成功
- **THEN** 系统 MUST 授予 Pro 人权益
