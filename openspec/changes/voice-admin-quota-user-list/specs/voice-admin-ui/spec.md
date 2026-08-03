## MODIFIED Requirements

### Requirement: voice-admin HTML SHALL configure voice_ai and clinic_ai quota

系统 MUST 提供独立 Admin 页面 **`/device/admin/voice-admin.html`**（静态文件 `resource/public/voice-admin.html`），包含「AI 额度」功能区：

1. **全局默认**表单（`voiceAiMonthlyLimit`、`clinicAiMonthlyLimit`），调用 **`GET/PUT /voice/admin/api/ai-quota/default`**。
2. **用户额度分页表**（替代原「单人 override（wxId）」表单模块）：调用 **`GET /voice/admin/api/ai-quota/users`** 分页加载；行内修改上限 MUST 调用既有 **`PUT /voice/admin/api/ai-quota/user`**。页面 MUST **移除**手输 wxId 加载/保存的单人 override 模块（含清除 checkbox UI）。

表格列顺序 MUST 为（左→右）：`deviceNo`、喂养已用、喂养上限、胖宝已用、胖宝上限、`wxId`、`account`、`babyName`。已用列只读；上限列可编辑。一行 MUST 同时展示喂养（`voice_ai`）与胖宝（`clinic_ai`）两额度。页面 MUST 提供按 **`deviceNo`** 查询并触发列表刷新。

页面 MUST NOT 包含润笔（polish）字段。页面 MUST 保留指向 **`/device/admin/ai-model-admin.html`** 的可见链接（文案含「模型与并发」或等价），MUST NOT 在本页恢复 LLM lane 编辑 Tab。页面 MUST 使用 `resource/public/admin-common.js` 的 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch`（或等价封装）初始化；Hub 登录后主内容区 MUST 可见且 MUST 加载全局默认与用户额度第一页。

页面说明区 MUST 列出额度变更影响的业务入口摘要：

- **喂养 AI**：`/voice/chat/ws`；`POST /voice/internal/api/text/chat`；`POST /voice/internal/api/text/chat/stream`；`POST /device/history/api/chat`；`POST /device/history/api/chat/stream`；`POST /voice/text/chat`。
- **胖宝 AI**：`/voice/clinic/ws`。

#### Scenario: 管理员修改喂养 AI 全局默认

- **WHEN** 运维在 voice-admin 页提交 voiceAi=5、clinicAi=30
- **THEN** 页面 SHALL 调用 PUT `/voice/admin/api/ai-quota/default` 且 voice-service 配置 SHALL 更新

#### Scenario: 页面不含润笔字段

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示 `polishMonthlyLimit` 输入控件

#### Scenario: 无单人 override 模块

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示手输 wxId 的「单人 override」加载/保存表单

#### Scenario: 额度表列顺序与双额度

- **WHEN** 列表返回至少一行
- **THEN** 表头第一列 SHALL 为 deviceNo，且同一行 SHALL 含喂养已用/上限与胖宝已用/上限

#### Scenario: 按 deviceNo 查询

- **WHEN** 运维输入 deviceNo 并查询
- **THEN** 页面 MUST 请求 `/voice/admin/api/ai-quota/users` 且携带该 deviceNo 过滤参数并刷新表格

#### Scenario: 行内修改上限

- **WHEN** 运维修改某行喂养或胖宝上限并保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/ai-quota/user`（含对应 wxId 与限额字段）

#### Scenario: Hub 登录后主面板可见

- **WHEN** 运维已在 `/device/admin` Hub 登录并打开 voice-admin.html
- **THEN** 页面 MUST 展示全局默认与用户额度表（非仅页头标题）

#### Scenario: 链至统一模型页

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 MUST 展示指向 ai-model-admin 的链接且 MUST NOT 含 LLM lane 编辑表单
