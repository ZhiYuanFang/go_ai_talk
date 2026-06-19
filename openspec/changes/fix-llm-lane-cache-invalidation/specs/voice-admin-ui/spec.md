## MODIFIED Requirements

### Requirement: voice-admin HTML SHALL configure voice_ai and clinic_ai quota

系统 MUST 提供独立 Admin 页面 **`/device/admin/voice-admin.html`**（静态文件 `resource/public/voice-admin.html`），包含「AI 额度」功能区：全局默认表单（`voiceAiMonthlyLimit`、`clinicAiMonthlyLimit`）与 per-wxId override 表单（query/body 含 `wxId`）。页面 MUST 调用 **`/voice/admin/api/ai-quota/default`** 与 **`/voice/admin/api/ai-quota/user`**（经 gateway-app 反代至 voice-service）。页面 MUST NOT 包含润笔（polish）字段。同一页面 MUST 提供 **「LLM 车道」** Tab，配置 `voiceUnderstanding` 与 `clinic` 的 `provider`、`model`、`maxInFlight`、`maxWaiters`，并调用 **`/voice/admin/api/llm-lanes`**。页面 MUST 使用 `resource/public/admin-common.js` 的 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch`（或等价封装）初始化；Hub 登录后主内容区 MUST 可见且 MUST 加载 AI 额度默认配置。

#### Scenario: 管理员修改喂养 AI 全局默认

- **WHEN** 运维在 voice-admin 页提交 voiceAi=5、clinicAi=30
- **THEN** 页面 SHALL 调用 PUT `/voice/admin/api/ai-quota/default` 且 voice-service 配置 SHALL 更新

#### Scenario: 页面不含润笔字段

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示 `polishMonthlyLimit` 输入控件

#### Scenario: LLM 车道 Tab 可保存

- **WHEN** 运维在「LLM 车道」Tab 修改 clinic 的 maxInFlight 并保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes`

#### Scenario: Hub 登录后主面板可见

- **WHEN** 运维已在 `/device/admin` Hub 登录并打开 voice-admin.html
- **THEN** 页面 MUST 展示 AI 额度与 LLM 车道 Tab 表单（非仅页头标题）
