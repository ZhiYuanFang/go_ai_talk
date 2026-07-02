## Why

App API 使用统计页「按用户」视图当前仅展示 wxId、设备号等基础字段，无法快速识别真实用户（缺少 UCG 昵称与注册时间），且 wx 选择列表包含 `is_simulated=1` 的模拟用户，与统计侧「模拟用户不计入 usage」的语义不一致。运维在按 API 下钻用户时同样只有 wxId，辨认成本高。

## What Changes

- `wx` 表新增 `created_at`（Unix 秒），所有 wx 账号创建写路径在插入时写入当前时间；历史行默认为 0，展示为「—」。
- `GET /device/admin/api/wx/list` 默认排除模拟用户（`is_simulated=0`），列表项增加 `createdAt`。
- 新增 gateway-app 编排端点 `GET /device/admin/api/usage/wx-list`：在 wx 列表基础上批量补充 UCG 展示昵称（含无 profile 时的推导默认昵称），供使用统计页「按用户」Tab 专用；**不修改**既有 `wx/list` v1 响应结构。
- `GET /device/admin/api/usage/detail` 响应每项增加 `nickname`（同上推导规则）。
- 扩展 ucg internal `POST /ucg/internal/api/profiles/batch`：对请求的每个 wxId 均返回 `nickname`；无 `ucg_profile` 行时 MUST 经 device 契约取 `babyName` 并应用与 App 一致的 `defaultNickname`（空 babyName →「家长」，否则 `{babyName}的家长`）。
- 更新 `api-usage-stats.html`：「按用户」表增加 UCG 昵称、注册时间列；API 下钻用户表增加 UCG 昵称列；改调 `usage/wx-list`。

**usage 统计策略**：本变更新增/调整的接口均为 `/device/admin/api/*` 运维读 API，结构性不计入 App usage 统计，**无需**变更 `maintenance_skip.go`。

## Capabilities

### New Capabilities

- `wx-created-at`：`wx` 表 `created_at` 列 DDL、实体/DAO、三条 Insert 写路径（微信登录、Apple 登录、用户名注册）及历史默认 0 的展示语义。

### Modified Capabilities

- `device-admin`：wx 分页列表排除模拟用户并返回 `createdAt`（device-service 实现层；对外 v1 `wx/list` 响应增加字段，唯一消费者为经 gateway 编排的 usage 页）。
- `gateway-app-api-usage-stats`：新增 `usage/wx-list` 编排读 API；`usage/detail` 增加 `nickname`；静态页列展示与接口切换。
- `ucg-internal-profile-batch`：batch 对无 profile 的 wxId 也返回推导昵称，保证 gateway 与 sim-admin 等调用方语义一致。

## Impact

- **MySQL（device 库 `wx` 表）**：DDL 增加 `created_at`；部署前须执行迁移脚本。
- **device-service**：`wx.go`、`apple_login.go`、`username_auth.go`、`wx_list_page.go`、entity/dao。
- **ucg-service**：`internal_profile_batch` / `BatchPublicProfilesForInternal` 扩展推导昵称。
- **gateway-app-server**：`usagestats` enrich 客户端、`gateway_app_usage_admin` 新端点与 detail 字段、静态页。
- **依赖**：gateway 经 internal HTTP 调用 ucg `profiles/batch`（需配置可达 `ucg-service` URL，与现有 sim-user 编排一致）。
- **文档**：`docs/runbooks/release-deploy-and-run.md` 补充 DDL 与部署顺序（device → ucg → gateway-app）。
