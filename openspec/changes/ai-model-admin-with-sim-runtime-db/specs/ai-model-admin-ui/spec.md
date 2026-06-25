## ADDED Requirements

### Requirement: gateway-app SHALL provide unified AI model and concurrency admin page

系统 MUST 提供独立 Admin 页面 **`/device/admin/ai-model-admin.html`**（静态文件 `resource/public/ai-model-admin.html`），经 gateway-app 托管。页面 MUST 在一屏内展示并编辑 **7 条 LLM lane** 的 `provider`、`model`、`maxInFlight`、`maxWaiters`：

| 分组 | lane 标识 | 后端 API |
|------|-----------|----------|
| Voice | `voiceUnderstanding`、`clinic` | `GET/PUT /voice/admin/api/llm-lanes` |
| UCG 润笔 | polish（`visionModel` 作 model） | `GET/PUT /ucg/admin/api/ai-config` |
| Sim | `simText`、`simVision`、`simImageGen`、`simVideoGen` | `GET/PUT /sim/admin/api/llm-lanes` |

页面 MUST 并行加载三域配置；保存 MUST 分别调用对应 PUT（可 `Promise.all`），并在部分失败时分域展示错误。页面 MUST 展示 allowlist 驱动的 provider→model 下拉联动。页面 MUST 含简短说明：同 `provider+model` 的多 lane 共享 Redis 闸门池。

#### Scenario: 统一页加载七 lane

- **WHEN** 已鉴权管理员打开 ai-model-admin.html
- **THEN** 页面 MUST 展示 voice×2、ucg polish、sim×4 共七组模型与并发字段且值来自对应 GET API

#### Scenario: 统一页保存 voice lane

- **WHEN** 管理员修改 `clinic.maxInFlight` 并点击保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes` 且 MUST NOT 调用 sim scheduler reload

#### Scenario: 部分域保存失败

- **WHEN** voice PUT 成功而 sim PUT 返回 400
- **THEN** 页面 MUST 明确提示 sim 失败原因且 voice 成功状态 MUST 可见

### Requirement: Admin Hub SHALL link ai-model-admin

`resource/public/admin-modules.js` MUST 增加 `id: ai-model-admin` 模块入口，导航至 `/device/admin/ai-model-admin.html`，`showInNav: true`。Hub 登录后 MUST 可点击进入。

#### Scenario: Hub 导航可见 AI 模型与并发

- **WHEN** 管理员登录 Admin Hub
- **THEN** 模块列表 MUST 包含 ai-model-admin 入口

### Requirement: voice-admin ucg-admin sim-admin SHALL link to unified page instead of editing LLM

`voice-admin.html`、`ucg-admin.html`、`sim-admin.html` MUST **移除** LLM 模型/并发编辑 Tab 或表单控件。各页 MUST 保留指向 `/device/admin/ai-model-admin.html` 的可见链接（文案含「模型与并发」或等价）。各页原有非 LLM 职责（额度、Prompt、任务状态等）MUST 不变。

#### Scenario: voice-admin 无 LLM 编辑 Tab

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 MUST NOT 含「LLM 车道」编辑 Tab 或 maxInFlight 输入框，且 MUST 含跳转统一页链接

#### Scenario: ucg-admin 无 polish 并发编辑

- **WHEN** 运维打开 ucg-admin「AI 配置」Tab
- **THEN** 页面 MUST NOT 含 provider/maxInFlight/maxWaiters/visionModel 编辑控件，且 MUST 含跳转统一页链接

#### Scenario: sim-admin 无 LLM lane 编辑

- **WHEN** 运维打开 sim-admin.html
- **THEN** 页面 MUST NOT 含 sim LLM lane 编辑表单，且 MUST 含跳转统一页链接

### Requirement: ai-model-admin static page MUST NOT count toward App usage stats

`/device/admin/ai-model-admin.html` 及本页触发的 voice/ucg/sim Admin PUT 为运维型接口，MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: 统一页保存不计入 usage

- **WHEN** 管理员从 ai-model-admin 保存 llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增 App API 计数
