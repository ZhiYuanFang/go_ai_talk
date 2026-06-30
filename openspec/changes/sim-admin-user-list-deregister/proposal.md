## Why

运维在 sim 管理页目前只能看到模拟用户总数，无法逐条查看测试账号、UCG 昵称/头像，也无法在测试环境主动注销单个模拟用户以释放名额或清理脏数据。同时 T1 注册仍使用可预测的 `ptest{N}` + 固定默认密码，不利于测试隔离与安全；管理页若需展示密码明文，必须在注册时另行持久化（`wx.password` 为 bcrypt，不可逆）。

## What Changes

- 在现有 `sim-admin.html` 嵌入「模拟用户列表」区块：展示头像、UCG 昵称、账号、wxId、注册时间、密码（明文）、注销操作。
- 新增 sim-admin API：`GET /sim/admin/api/users`（分页列表）、`POST /sim/admin/api/users/{wxId}/deactivate`（仅删 `wx`，与 App 注销语义一致）。
- sim-user-service 新增表 `sim_wx_credential`：注册时写入 `wx_id`、`account`、`password_plain`、`created_at`；注销时删除对应行。
- T1 注册改为 **随机账号 + 随机密码**（满足 device 用户名/密码规则），停用 `sim_account_seq` / `ptest{N}` 序号分配。
- T2–T6 及手动任务登录改为按 `wxId` 从 `sim_wx_credential` 取密码；历史无 credential 行的用户 fallback 到 `simUser.defaultPassword`（默认 `123456`）。
- device-service 新增 internal：`POST /device/internal/api/sim/wx/{wxId}/deactivate`（须 `is_simulated=1` 才允许删 wx）。
- ucg-service 新增 internal：`POST /ucg/internal/api/profiles/batch`（按 wxId 批量返回 nickname、avatarUrl 等展示字段）。
- 注销时 sim-user-service 可选将该 wxId 的 pending/processing `sim_video_job` 标为 `skipped`；**不**级联删除 UCG 帖子/评论/profile。

## Capabilities

### New Capabilities

- `ucg-internal-profile-batch`：UCG internal 批量 profile 展示接口，供 sim-admin 列表合并昵称与头像。

### Modified Capabilities

- `sim-user-admin`：新增用户列表与注销 Admin API；sim-admin UI 嵌入用户表格。
- `sim-user-service`：T1 随机凭据 + credential 表；T2–T6 按 wxId 查密码登录；注销清理 credential 与 video job。
- `device-sim-user`：新增 sim 用户 internal 注销接口（仅 `is_simulated=1`）。

## Impact

- **进程**：`sim-user-service`、`device-service`、`ucg-service`、`gateway-app-server`（静态页 + 既有 sim admin 反代，无新 App API）。
- **数据库**：`SIM_DB_LINK` 新增 `sim_wx_credential`；`device` 库 `wx` 表无结构变更。
- **配置**：`simUser.defaultPassword` 保留为历史 sim 用户登录 fallback，新用户不再使用该固定值注册。
- **OpenSpec 基线**：替换 v2.0.9 中 T1「`ptest{N}` + 默认 `123456`」相关 Requirement/Scenario。
- **Usage 统计**：新增路由均为 `/sim/admin/api/*` 与 internal API，**不计入** App usage；无需变更 `maintenance_skip.go`。
- **Redis**：注销时从 `usage:sim_wx_ids` SET 移除 member（与注册时 SADD 对称）；不经业务层键字面量，使用既有 `cachekit` builder。
- **历史数据**：不 backfill `sim_wx_credential`；列表对历史 `ptest*` 用户密码列显示 `123456` 并标注「默认密码（历史）」，注册时间无记录时显示「—」。
