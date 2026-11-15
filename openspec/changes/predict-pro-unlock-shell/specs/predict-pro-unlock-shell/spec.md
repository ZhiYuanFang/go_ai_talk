## ADDED Requirements

### Requirement: 系统 MUST 仅提供预测能力 Pro 商业开通壳

cash-service MUST 种子并对人维护功能 `predict_capability_pro`（`activation_subject=user`，`unlock_methods` 含 `payment,invite_code`，MUST NOT 含 `ad`）。MUST 支持支付建单、邀请兑换、显式点试用（自点击起算 7 日写入人权益）。App catalog MUST 可返回该项及真实 `unlocked`/`status`。本变更 MUST NOT 要求任何业务路径（含推送）因持有 Pro 权益而改变行为。

#### Scenario: 合格用户可点试用开通 Pro

- **WHEN** 绑机用户喂养资格合格、试用未用、非 VIP、无有效 Pro 权益，调用点试用且 featureId 为 `predict_capability_pro`
- **THEN** 系统 MUST 写入约 7 日人权益且试用标为 used

#### Scenario: 开通 Pro 不改变推送资格

- **WHEN** 用户已开通 Pro 或未开通 Pro
- **THEN** imminent 推送是否发出 MUST NOT 仅因 Pro 权益有无而不同（由既有非商业门禁决定）

### Requirement: 系统 MUST NOT 再提供预测能力基础版开通

系统 MUST NOT 再种子或启用 `predict_capability_basic` 作为可售功能（Ensure 若遇已有行 MUST 将其 `status=0` 或等价停用，MUST NOT 继续 INSERT 启用态）。点试用、建单、邀请 MUST 拒绝以基础版为目标的新开通（若仍收到该 featureId）。

#### Scenario: Ensure 后基础版不可售

- **WHEN** 执行 EnsureSchema 且库中曾有基础版定义
- **THEN** 基础版 MUST 不为上架可售态（status≠1 或不再出现在可购 products）

#### Scenario: 点试用基础版被拒

- **WHEN** 客户端对 `predict_capability_basic` 调用点试用
- **THEN** 系统 MUST 拒绝
