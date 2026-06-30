## 1. ucg-service：internal 批量 profile

- [x] 1.1 在 `api/v1` 定义 `POST /ucg/internal/api/profiles/batch` 请求/响应结构
- [x] 1.2 实现 controller + 调用 `GetPublicProfilesByWxIDs`，注册到 ucg-service 路由组
- [x] 1.3 确认 internal 鉴权与现有 ucg internal 一致

## 2. device-service：sim 用户 internal 注销

- [x] 2.1 在 `api/v1` 定义 `POST /device/internal/api/sim/wx/{wxId}/deactivate`
- [x] 2.2 在 `internal/services/device` 实现 `SimWxDeactivateByID`：`is_simulated=1` 守卫 + `WxDeactivateByID` + `usage:sim_wx_ids` SREM（经 cachekit）
- [x] 2.3 实现 controller 并注册路由；映射业务错误为 4xx

## 3. sim-user-service：credential 表与 T1/T2–T6

- [x] 3.1 在 `EnsureSchema` 新增 `sim_wx_credential` 表及 CRUD helper（Insert/GetByWxIds/DeleteByWxID）
- [x] 3.2 实现随机账号/密码生成（满足 device 规则，碰撞重试），T1 替换 `NextAccountName` 固定序号逻辑
- [x] 3.3 T1 注册成功后写入 credential；实现 `simLoginByWxId`（credential 优先，历史 fallback defaultPassword）
- [x] 3.4 更新 T2–T6 与 `randomSimSession`/手动任务路径使用 per-wxId 密码
- [x] 3.5 注销编排：skip 该 wxId 的 pending/processing `sim_video_job`

## 4. sim-user-service：Admin API

- [x] 4.1 在 `api/v1/sim_admin_http.go` 定义 `GET /sim/admin/api/users` 与 `POST /sim/admin/api/users/{wxId}/deactivate`
- [x] 4.2 实现 list 编排（device list + ucg profiles batch + credential 合并，含 legacy password 字段）
- [x] 4.3 实现 deactivate 编排（device internal deactivate + 本地 credential/video job 清理）
- [x] 4.4 在 `SimAdminCtrl` 绑定新路由；确认 gateway `X-Admin-Password` 鉴权路径已覆盖

## 5. sim-admin UI

- [x] 5.1 在 `sim-admin.html` 嵌入模拟用户表格（头像、昵称、账号、wxId、注册时间、密码、注销）
- [x] 5.2 实现分页与 `confirm` 注销；成功后刷新列表与 runtime 计数
- [x] 5.3 历史用户 UI：`createdAt=0` 显示「—」；`passwordPlainLegacy` 显示「默认密码（历史）」

## 6. 验收与文档

- [x] 6.1 手工验证：列表展示 UCG 昵称/头像；注销后 wx 删除且 sim 计数减 1；T1 新用户为随机账号且 credential 可查
- [x] 6.2 确认未新增 App HTTP 路由，无需变更 `usagestats/maintenance_skip.go`
- [x] 6.3 若 runbook 需补充 sim-admin 新 API 说明，更新 `docs/runbooks/release-deploy-and-run.md`（仅运维段落，非新 DB 连接）
