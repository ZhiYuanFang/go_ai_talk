## ADDED Requirements

### Requirement: sim-user-admin SHALL expose paginated simulated user list API

`sim-user-service` MUST 提供 `GET /sim/admin/api/users`（鉴权与现有 sim-admin 一致）。Query 参数 MUST 支持 `page`（默认 1）、`pageSize`（默认 20，最大 200）。

响应 MUST 含 `list`、`total`、`page`、`pageSize`。`list` 每项 MUST 含：

- `wxId`（int64）
- `account`（string）
- `nickname`（string，无 UCG profile 时为空字符串）
- `avatarUrl`（string，可选）
- `avatarThumbnailUrl`（string，可选）
- `createdAt`（int64 Unix 秒；无 credential 记录时为 0）
- `passwordPlain`（string；无 credential 时 MUST 为 `123456` 且 `passwordPlainLegacy` MUST 为 true）
- `passwordPlainLegacy`（bool，可选；true 表示历史用户未持久化明文）

服务 MUST 经 device internal `sim/wx/list`、ucg internal `profiles/batch` 与 sim 库 `sim_wx_credential` 合并结果，MUST NOT 直查 device/ucg 库。

#### Scenario: Admin lists sim users with profile

- **WHEN** 已鉴权 GET `/sim/admin/api/users?page=1&pageSize=20` 且存在带 UCG profile 的 sim 用户
- **THEN** 响应 `list` MUST 含对应 `nickname` 与非空 `avatarUrl` 或 `avatarThumbnailUrl`（若 profile 有头像）

#### Scenario: Legacy user password fallback

- **WHEN** 列表项 wxId 在 `sim_wx_credential` 无记录（历史 `ptest*` 用户）
- **THEN** 该项 `passwordPlain` MUST 为 `123456` 且 `passwordPlainLegacy` MUST 为 true

#### Scenario: CreatedAt from credential

- **WHEN** T1 注册后写入 `sim_wx_credential`
- **THEN** 列表该项 `createdAt` MUST 等于 credential 的 `created_at`

### Requirement: sim-user-admin SHALL expose simulated user deactivate API

`sim-user-service` MUST 提供 `POST /sim/admin/api/users/{wxId}/deactivate`（鉴权与现有 sim-admin 一致）。路径参数 `wxId` MUST 为正整数。

成功时 MUST：调用 device internal sim deactivate 删除 `wx` 行；删除 sim 库 `sim_wx_credential` 对应行；将该 wxId 的 `sim_video_job` 中 `pending`/`processing` 行标为 `skipped`。

MUST NOT 调用 ucg 或 gateway App API 删除帖子/profile。注销语义 MUST 与 App `POST /device/app/api/user/deactivate` 一致（仅删 wx）。

#### Scenario: Deactivate simulated user success

- **WHEN** 已鉴权 POST `/sim/admin/api/users/1001/deactivate` 且 wxId=1001 为 `is_simulated=1`
- **THEN** HTTP MUST 成功且该 wx 行 MUST 自 device 库删除且 credential 行 MUST 删除

#### Scenario: Reject non-sim wxId

- **WHEN** wxId 存在但 `is_simulated=0`
- **THEN** MUST 返回 4xx 业务错误且 MUST NOT 删除 wx

#### Scenario: Reject invalid wxId

- **WHEN** wxId 不存在或已注销
- **THEN** MUST 返回明确业务错误（已注销或不存在）

### Requirement: sim-admin UI SHALL display simulated user table with deactivate

`sim-admin.html` MUST 在现有页面内嵌入「模拟用户」区块（MUST NOT 要求单独 Hub 模块页）。表格 MUST 展示：头像（优先 `avatarThumbnailUrl`）、UCG 昵称、账号、wxId、注册时间、密码、注销操作。

- 注册时间：`createdAt=0` 时 MUST 显示「—」
- 密码：`passwordPlainLegacy=true` 时 MUST 标注「默认密码（历史）」
- 注销：点击前 MUST `confirm`；成功后 MUST 刷新列表并更新 runtime 区 `simUserCount`

分页 MUST 调用 `GET /sim/admin/api/users`。页面 MUST NOT 提供 LLM lane 编辑（既有约束不变）。

#### Scenario: Admin sees user row

- **WHEN** 管理员已登录 Hub 并打开 sim-admin 且存在 sim 用户
- **THEN** 表格 MUST 展示至少一行含昵称/账号/wxId/密码列

#### Scenario: Deactivate refreshes list

- **WHEN** 管理员确认注销某一 sim 用户且 API 成功
- **THEN** 该行 MUST 从表格消失且 runtime 模拟用户数 MUST 减少

## MODIFIED Requirements

### Requirement: sim admin API SHALL expose config and prompts

`sim-user-service` MUST 提供 Admin HTTP API（经 gateway 反代或直连，鉴权与 ucg-admin 对齐）：

- `GET/PUT /sim/admin/api/config` — 字段至少含 `enabled`（bool）、`maxSimUsers`（int，默认 100）
- `GET/PUT /sim/admin/api/prompts/{taskType}` — `taskType` 至少含：`register_nickname`、`register_avatar`、`comment`、`post_image_text`、`post_video_text`、`chat_reply`；每项含 `systemPrompt`、`userPromptTemplate`
- `GET /sim/admin/api/status` — 各任务上次运行时间、成功/失败计数、pending video job 数
- `GET /sim/admin/api/users` — 分页模拟用户列表（见 ADDED Requirement）
- `POST /sim/admin/api/users/{wxId}/deactivate` — 注销单个模拟用户（见 ADDED Requirement）

Prompt 变更 MUST 在下一任务 tick 生效，无需重启进程（读 DB + 短 TTL 缓存可接受）。

#### Scenario: Update max sim users

- **WHEN** 管理员 PUT `maxSimUsers=50`
- **THEN** 下一次 T1 MUST 在 sim 用户数 ≥50 时停止注册

#### Scenario: Update comment prompt

- **WHEN** 管理员修改 `comment` 的 `userPromptTemplate`
- **THEN** 下一次 T2 MUST 使用新模板渲染变量
