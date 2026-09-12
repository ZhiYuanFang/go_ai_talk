## ADDED Requirements

### Requirement: Admin 功能定义 API MUST 暴露开通主体

`GET /cash/admin/api/feature/defs` 返回的每一项 MUST 包含 `activationSubject`（`device`|`user`）。`POST /cash/admin/api/feature/defs` MUST 接受可选 `activationSubject`；若提供则 MUST 校验取值并持久化。将 `prediction_unlock`（或其它条数类）更新为 `user` MUST 被拒绝。未传该字段时 MUST 保持库中原值不变。

#### Scenario: 列表返回人对机

- **WHEN** 管理员拉取功能定义列表
- **THEN** 每条 MUST 含 `activationSubject`，成长轨迹为 `user`，值得留意为 `device`

#### Scenario: 禁止预测改为人

- **WHEN** Admin 将预测功能 `activationSubject` 设为 `user` 并保存
- **THEN** 系统 MUST 拒绝且 MUST NOT 修改该行主体

### Requirement: 开通功能管理页 MUST 标明对人或对机

`resource/public/cash-feature-admin.html`（经 gateway 运维 Hub 入口）MUST 在功能定义列表中展示每条功能的开放主体，文案 MUST 区分「对人」与「对机」（或等价清晰中文），MUST NOT 仅依赖英文枚举让运维猜测。编辑区 MUST 展示当前主体；若允许修改，MUST 提供明确控件并在保存后与列表一致。

#### Scenario: 列表一眼可辨成长轨迹对人

- **WHEN** 运维打开开通功能管理页并加载定义列表
- **THEN** `growth_trajectory_predict` 行 MUST 显示「对人」（或含「人/账号」语义的同等文案）

#### Scenario: 值得留意显示对机

- **WHEN** 列表展示 `care_alert_smart_remind`
- **THEN** 该行 MUST 显示「对机」（或含「设备」语义的同等文案）
