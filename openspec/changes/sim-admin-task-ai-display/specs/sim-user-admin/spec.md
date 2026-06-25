## ADDED Requirements

### Requirement: sim admin API SHALL expose per-task AI model catalog

`GET /sim/admin/api/status` 与 `GET /sim/admin/api/runtime` 响应 MUST 含 `taskAiModels`（object）：键为调度任务名（至少含 `register`、`comment`、`post_image`、`post_video_submit`、`chat_scan`、`follow`、`video_poll`），值为数组。数组元素 MUST 含：

- `laneKey`（string）— sim lane 标识，如 `simText`、`simImageGen`
- `usage`（string，可选）— 该 lane 在本任务中的用途说明，如「昵称」「E1 回复」
- `provider`（string）— 当前生效上游，来自 `SimLLMLaneStore` / env 解析
- `model`（string）— 当前生效模型名

catalog MUST 由服务端根据 `tasks.go` 一致的 lane 映射生成；MUST NOT 含 API Key。不使用 LLM 的任务（`follow`）MUST 对应空数组或省略键（UI 显示为无 AI）。

#### Scenario: Status returns AI for register task

- **WHEN** 已鉴权管理员 GET `/sim/admin/api/status` 且 simText、simImageGen 已配置
- **THEN** `taskAiModels.register` MUST 含两条记录，分别对应 simText 与 simImageGen 的 provider/model

#### Scenario: Follow task has no AI

- **WHEN** 管理员 GET `/sim/admin/api/status`
- **THEN** `taskAiModels.follow` MUST 为空数组或不存在，且 MUST NOT 含 lane 条目

#### Scenario: Runtime includes same catalog

- **WHEN** 管理员 GET `/sim/admin/api/runtime`
- **THEN** 响应 `taskAiModels` MUST 与 status 接口解析逻辑一致（同一 lane profile 源）

### Requirement: sim admin UI SHALL display AI models per scheduled task

`sim-admin.html` 任务状态表 MUST 增加「AI 模型」列。每行 MUST 根据 `taskAiModels[taskName]` 展示：

- 用途说明（若有 `usage`）
- lane 键（如 `simText`）
- `provider/model`；未配置时 MUST 可读提示（如「未配置」）

无 LLM 的任务 MUST 显示「—」。页面 MUST 说明修改模型须至 ai-model-admin，本页 MUST NOT 提供 lane 编辑。

#### Scenario: Task table shows AI column

- **WHEN** 管理员打开 sim-admin 并加载状态
- **THEN** 表格 MUST 含「AI 模型」列且 T1 行展示 simText 与 simImageGen 信息

#### Scenario: Refresh reflects lane change

- **WHEN** 管理员在 ai-model-admin 修改 simText model 后点击 sim-admin「刷新状态」
- **THEN** 任务表中依赖 simText 的行 MUST 展示新 model

#### Scenario: No lane editor on status table

- **WHEN** 管理员查看任务状态表
- **THEN** UI MUST NOT 提供修改 provider/model 的输入控件
