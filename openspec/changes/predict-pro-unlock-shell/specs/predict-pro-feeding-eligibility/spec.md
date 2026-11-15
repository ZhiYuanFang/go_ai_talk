## ADDED Requirements

### Requirement: Pro 商业开通 MUST 仍校验 predict_pro_entry 且不授予推送特权

对 `predict_capability_pro` 的点试用、建单、邀请兑换，系统 MUST 继续要求当前绑机在场景 `predict_pro_entry` 下 `qualified=true`。App `GET /cash/app/api/predict-pro/eligibility` MUST 保持字段同构并 **计入 usage**。通过开通获得的 Pro 权益 MUST NOT 被解释为 imminent 推送许可（推送见 `predict-push-ungate`）。

#### Scenario: 未合格仍拒 Pro 试用

- **WHEN** 用户点 Pro 试用但 `predict_pro_entry` 未合格
- **THEN** 系统 MUST 拒绝且 MUST NOT 写 Pro 权益

#### Scenario: 已开通 Pro 不单独放行推送

- **WHEN** 用户仅有 Pro 权益、非 VIP，且其它推送门禁与未开通用户相同
- **THEN** 推送路径 MUST NOT 额外因 Pro 放行或拒绝（与未开通用户同一套非商业规则）
