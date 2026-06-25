## MODIFIED Requirements

### Requirement: voice-admin 与 ucg-admin SHALL 提供 LLM lane 配置 UI

LLM lane 的模型与并发 Admin UI MUST 集中在 **`/device/admin/ai-model-admin.html`**（见 `ai-model-admin-ui`）。`voice-admin.html` 与 `ucg-admin.html` MUST NOT 再提供 LLM lane 编辑 Tab 或表单；MUST 仅链至统一页。后端 API（`GET/PUT /voice/admin/api/llm-lanes`、`GET/PUT /ucg/admin/api/ai-config` 之 polish 字段）MUST 不变，由统一页调用。

#### Scenario: voice LLM 经统一页保存

- **WHEN** 运维在 ai-model-admin 修改 voiceUnderstanding 并保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes` 且 voice-admin MUST NOT 含重复编辑 UI

#### Scenario: ucg polish 经统一页保存

- **WHEN** 运维在 ai-model-admin 修改 polish maxWaiters 并保存
- **THEN** 页面 MUST 调用 PUT `/ucg/admin/api/ai-config` 且 ucg-admin MUST NOT 含重复编辑 UI
