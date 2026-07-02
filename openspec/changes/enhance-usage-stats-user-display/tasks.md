## 1. 数据库与 device 域 wx.created_at

- [x] 1.1 新增 `hack/ddl_wx_created_at.sql`：`wx` 表添加 `created_at BIGINT NOT NULL DEFAULT 0`（含中文注释）
- [x] 1.2 更新 `entity.Wx`、`dao/internal/wx.go` 的 `WxColumns` 与生成模型字段
- [x] 1.3 `WxLogin`、`Apple 首次 Insert`、`WxUsernameRegister` 插入时写入 `time.Now().Unix()`
- [x] 1.4 `ListWxPage` 增加 `WHERE is_simulated=0`，`Fields` 含 `created_at`，契约与 `DeviceAdminWxListItem` 增加 `CreatedAt`

## 2. ucg internal profiles/batch 扩展

- [x] 2.1 实现无 `ucg_profile` 行时的推导昵称（经 `Device().ValidateWx` + `defaultNickname`，不写库）
- [x] 2.2 更新 `BatchPublicProfilesForInternal`：请求中每个有效 wxId 均在 `list` 返回一条
- [x] 2.3 确认 internal controller 响应 JSON 与 `api/v1` 类型一致

## 3. gateway-app usage 编排与 API

- [x] 3.1 在 `usagestats` 包新增 ucg internal HTTP 客户端（`profiles/batch`，支持分批 wxIds）
- [x] 3.2 新增 `api/v1` 类型与 `GET /device/admin/api/usage/wx-list` Handler（合并 device wx 列表 + nickname）
- [x] 3.3 `UsageDetail` 对 `ListUsersForAPI` 结果 batch enrich `nickname` 并扩展响应结构
- [x] 3.4 在 `gateway_app_register` 绑定新 Controller 方法

## 4. 静态页

- [x] 4.1 `api-usage-stats.html`「按用户」改调 `usage/wx-list`，表头增加 UCG 昵称、注册时间（0 显示「—」）
- [x] 4.2 API 下钻用户表增加 UCG 昵称列

## 5. 文档与验收

- [x] 5.1 更新 `docs/runbooks/release-deploy-and-run.md`：DDL 步骤与部署顺序（device → ucg → gateway-app）
- [ ] 5.2 手工验收：按用户列表无模拟用户、有昵称与注册时间；API 下钻有昵称；新注册 wx `createdAt>0`

## 6. usage 统计确认（记录）

- [x] 6.1 已确认：本变更新增 `/device/admin/api/usage/wx-list` 及 `usage/detail` 字段扩展均为 admin 读 API，结构性不计入 App usage，**无需**修改 `maintenance_skip.go`
