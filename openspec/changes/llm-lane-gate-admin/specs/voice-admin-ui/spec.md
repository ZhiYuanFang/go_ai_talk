## MODIFIED Requirements

### Requirement: voice-admin HTML SHALL configure voice_ai and clinic_ai quota

系统 MUST 提供独立 Admin 页面 **`/device/admin/voice-admin.html`**（静态文件 `resource/public/voice-admin.html`），包含「AI 额度」功能区：全局默认表单（`voiceAiMonthlyLimit`、`clinicAiMonthlyLimit`）与 per-wxId override 表单（query/body 含 `wxId`）。页面 MUST 调用 **`/voice/admin/api/ai-quota/default`** 与 **`/voice/admin/api/ai-quota/user`**（经 gateway-app 反代至 voice-service）。页面 MUST NOT 包含润笔（polish）字段。同一页面 MUST 提供 **「LLM 车道」** Tab，配置 `voiceUnderstanding` 与 `clinic` 的 `provider`、`model`、`maxInFlight`、`maxWaiters`，并调用 **`/voice/admin/api/llm-lanes`**。

#### Scenario: 管理员修改喂养 AI 全局默认

- **WHEN** 运维在 voice-admin 页提交 voiceAi=5、clinicAi=30
- **THEN** 页面 SHALL 调用 PUT `/voice/admin/api/ai-quota/default` 且 voice-service 配置 SHALL 更新

#### Scenario: 页面不含润笔字段

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示 `polishMonthlyLimit` 输入控件

#### Scenario: LLM 车道 Tab 可保存

- **WHEN** 运维在「LLM 车道」Tab 修改 clinic 的 maxInFlight 并保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes`

### Requirement: ucg-admin SHALL remove voice and clinic quota fields

`resource/public/ucg-admin.html`「AI 配置」Tab MUST **移除** `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit` 相关表单与 API 字段，**仅保留** `polishMonthlyLimit` 全局默认与 per-wxId override。同一 Tab MUST 扩展 **润笔 lane** 的 `provider`、`maxInFlight`、`maxWaiters` 配置（`visionModel` 作为 polish 的 model 选择），并调用扩展后的 **`/ucg/admin/api/ai-config`**。

#### Scenario: ucg-admin 仅润笔配置

- **WHEN** 运维打开 ucg-admin「AI 配置」Tab
- **THEN** 页面 SHALL 展示 polish 相关字段（含模型、并发、缓冲池）且 SHALL NOT 调用 voice/clinic 配额 API
