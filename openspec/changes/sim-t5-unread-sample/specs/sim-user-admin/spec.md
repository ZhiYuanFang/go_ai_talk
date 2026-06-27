## MODIFIED Requirements

### Requirement: sim-admin SHALL expose editable runtime configuration aligned with DB schema

`sim-admin.html` MUST 提供可编辑运行时配置表单，字段覆盖 `taskSwitches`、`intervals`、`rateLimitRps`、`rateLimitBurst`（及既有 `enabled`、`maxSimUsers`）。**MUST NOT** 含 `ephemeralChatLoop`、`ephemeralChatWindow` 或 E1 相关文案。保存 MUST 调用 **`PUT /sim/admin/api/config`**。保存成功后 MUST 展示 API 返回的 **`effects`** 与 **`taskSchedule`**。页面 MUST 区分 `serviceEnabled`（env）与 `dbEnabled`（可在线保存）。`GET /sim/admin/api/status` 结构化任务状态 MUST 保留；`taskAiModels.chat_scan` usage MUST 为「未读回复」（非「E1 回复」）。

#### Scenario: No E1 fields in admin form

- **WHEN** 管理员打开 sim-admin 运行配置区
- **THEN** MUST NOT 展示 E1 循环/窗口输入框

#### Scenario: Save without ephemeral effects

- **WHEN** PUT 保存 task/interval 变更
- **THEN** 响应 `effects` MUST NOT 含 `ephemeral_may_continue`

#### Scenario: Chat scan task AI label

- **WHEN** GET status 或 runtime 含 `taskAiModels.chat_scan`
- **THEN** lane usage MUST 为「未读回复」

### Requirement: sim-admin runtime panel SHALL display effective task schedule

`GET /sim/admin/api/runtime` 与 status 中 `intervals` 键 MUST 含 `register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`；**MUST NOT** 含 `ephemeralChatLoop`、`ephemeralChatWindow`、`videoPollIdle`、`videoPollActive`。

#### Scenario: Runtime intervals without ephemeral

- **WHEN** GET runtime
- **THEN** `intervals` MUST NOT 含 E1 键

## REMOVED Requirements

### Requirement: Admin effects for ephemeral chat continuation

**Reason**: E1 已删除。

**Migration**: 移除 `config_admin` 中 `ephemeral_may_continue` 与 T5 关闭时的 E1 提示 effect。
