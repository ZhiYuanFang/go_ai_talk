## ADDED Requirements

### Requirement: sim admin web SHALL be served from gateway-app

系统 MUST 在 `gateway-app-server` 注册静态页 `/device/admin/sim-admin.html`（`resource/public/sim-admin.html`）。页面 MUST 要求管理员已在运维 Hub 登录（与 ucg-admin 一致的鉴权模式）。

#### Scenario: Admin page reachable

- **WHEN** 管理员已登录 Hub 并访问 `/device/admin/sim-admin.html`
- **THEN** 浏览器 MUST 加载模拟管理界面

### Requirement: sim admin API SHALL expose config and prompts

`sim-user-service` MUST 提供 Admin HTTP API（经 gateway 反代或直连，鉴权与 ucg-admin 对齐）：

- `GET/PUT /sim/admin/api/config` — 字段至少含 `enabled`（bool）、`maxSimUsers`（int，默认 100）
- `GET/PUT /sim/admin/api/prompts/{taskType}` — `taskType` 至少含：`register_nickname`、`register_avatar`、`comment`、`post_image_text`、`post_video_text`、`chat_reply`；每项含 `systemPrompt`、`userPromptTemplate`
- `GET /sim/admin/api/status` — 各任务上次运行时间、成功/失败计数、pending video job 数

Prompt 变更 MUST 在下一任务 tick 生效，无需重启进程（读 DB + 短 TTL 缓存可接受）。

#### Scenario: Update max sim users

- **WHEN** 管理员 PUT `maxSimUsers=50`
- **THEN** 下一次 T1 MUST 在 sim 用户数 ≥50 时停止注册

#### Scenario: Update comment prompt

- **WHEN** 管理员修改 `comment` 的 `userPromptTemplate`
- **THEN** 下一次 T2 MUST 使用新模板渲染变量

### Requirement: Task intervals SHALL NOT be editable from admin UI in v1

任务周期（24h/6h/3.5h/6.5h/1h/7h/1min）MUST 由代码或环境变量固定；模拟管理页 MUST NOT 提供周期编辑以免误配导致 LLM 风暴。

#### Scenario: No interval editor

- **WHEN** 管理员打开 sim-admin 配置页
- **THEN** UI MUST NOT 展示任务周期间隔输入框
