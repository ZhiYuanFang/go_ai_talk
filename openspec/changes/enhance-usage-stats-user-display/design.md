## Context

使用统计读路径已在 gateway-app 本机（Redis + `GatewayAppUsageAdminCtrl`），「按用户」Tab 却反代 device `wx/list` 展示全量 wx（含模拟用户），且缺少昵称与注册时间。API 下钻 `usage/detail` 仅返回 wxId/count/lastAt。UCG 昵称推导逻辑在 `ucg/profile.go` 的 `defaultNickname`；`profiles/batch` 目前跳过无 profile 行。`wx` 表无创建时间列。

约束：跨服务须走 HTTP 契约；gateway 不得直查 device/ucg 库；admin API 不计入 usage 统计；不新增测试文件。

## Goals / Non-Goals

**Goals:**

- 为真实 wx 账号提供权威 `created_at`（创建时写入）。
- 使用统计页「按用户」与 API 下钻用户列表展示 UCG 展示昵称（含无 profile 时的推导默认昵称）。
- wx 选择列表排除 `is_simulated=1`。
- 保持「全量真实 wx + 搜索点选」交互，不改为「仅有调用的用户」列表。

**Non-Goals:**

- 不回填历史 wx 的精确注册时间（`created_at=0` 显示「—」即可）。
- 不修改 `GET /device/admin/api/usage/user`（单用户 API 明细）响应结构。
- 不将按用户列表改为 Redis 活跃用户数聚合。
- 不新增 Redis 读缓存。

## Decisions

### 1. `wx.created_at` 列与写路径

- **决策**：`ALTER TABLE wx ADD COLUMN created_at BIGINT NOT NULL DEFAULT 0`；在 `WxLogin`、`AppleLogin` 首次 Insert、`WxUsernameRegister` Insert 时设 `time.Now().Unix()`。
- **理由**：与 `database-unix-timestamp-storage` 及项目内 `ucg_profile.created_at` 一致。
- **备选**：用 `ucg_profile.created_at` 代理注册时间 — 已否决（语义是进社区而非 wx 创建）。

### 2. 模拟用户过滤在 device `ListWxPage`

- **决策**：查询默认 `WHERE is_simulated = 0`；不新增 query 开关（`wx/list` 仅 usage 页消费）。
- **理由**：最小改动；与统计写入跳过 sim 一致。
- **备选**：gateway 侧过滤 — 分页 total 会不准且仍拉取 sim 行。

### 3. 新增 `GET /device/admin/api/usage/wx-list` 而非扩 v2 `wx/list`

- **决策**：gateway 本机 Handler：反代/调用 device `wx/list`（或 HTTP 调 device admin）→ 对当前页 wxIds 调 ucg `profiles/batch` → 合并 `nickname` + 透传 `createdAt` 等。
- **理由**：昵称 enrich 属 gateway 编排；避免 device 跨域查 ucg；`usage/*` 已在 gateway 本机处理。
- **备选**：直接改 `wx/list` 响应加 nickname — device 无法直查 ucg，违反服务边界。

### 4. 昵称推导集中在 ucg `profiles/batch`

- **决策**：扩展 `BatchPublicProfilesForInternal`：对无 profile 的 wxId，调用已有 `Device().ValidateWx` 取 `babyName`，返回 `defaultNickname(babyName)`，**不**写库。
- **理由**：与 App `GetOrCreateMyProfile` 展示语义单点维护；gateway 只消费 batch 结果。
- **备选**：gateway 本地 merge device `ucg/wx/batch` + 复制 `defaultNickname` — 易漂移。

### 5. `usage/detail` 增加 `nickname`

- **决策**：`ListUsersForAPI` 不变；controller 收集 wxIds 后单次（或分批每 100）调用 ucg batch enrich，写入 `DeviceAdminUsageDetailItem.Nickname`。
- **理由**：复用同一 batch 契约；下钻列表可能较长，实现时分批 POST。

### 6. gateway 调 ucg internal HTTP

- **决策**：在 `usagestats` 包新增 `fetchProfileNicknamesHTTP`，配置项对齐 sim-user：`UCG_SERVICE_URL` 或 `gatewayApp.ucgServiceUrl`，Header `X-Device-Gateway-Internal-Secret`。
- **理由**：已有 `fetchWxIsSimulatedHTTP` 模式可复用。

## Risks / Trade-offs

- **[Risk] ucg-service 不可达时昵称列为空** → 列表/下钻仍展示 wxId 与统计数字；记录 warning；`createdAt` 来自 device 不受影响。
- **[Risk] 下钻热门 API 的 wxId 数量大** → batch 按 100 分批；admin 场景可接受。
- **[Risk] `wx/list` v1 增加 `createdAt` 字段** → 唯一消费者为 usage 页经新编排端点；JSON 增字段对 admin 工具向后兼容。
- **[Risk] DDL 未执行导致读失败** → runbook 明确先迁移后部署 device；列有 DEFAULT 0。

## Migration Plan

1. **DDL**（device 库）：执行 `hack/ddl_wx_created_at.sql`（变更内新增）添加 `created_at`。
2. **部署顺序**：device-service → ucg-service → gateway-app-server。
3. **回滚**：回滚服务二进制；**不**删除 `created_at` 列（已写入数据保留）。
4. **验收**：打开 `api-usage-stats.html`，「按用户」无 sim 行、有昵称与注册时间；API 下钻见昵称。

## Open Questions

- 无（产品侧已在 explore 阶段确认：wx 创建时间、全量 wx 列表、下钻加昵称、默认昵称推导）。
