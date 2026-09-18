## 1. Schema：token 全局唯一

- [x] 1.1 `EnsureSchema`：存量按 token 去重（保留 `updated_at` 最大，并列保留更大 `id`），再添加 `UNIQUE KEY uk_token (token)`；保留既有 `uk_wx_device_channel`
- [x] 1.2 索引名常量与幂等探测（information_schema），中文注释说明迁移语义

## 2. 注册后来顶上

- [x] 2.1 改造 `RegisterPushDevice`：写入前删除库中相同 `token` 的全部行，再按 `(wx_id, device_key, channel)` upsert
- [x] 2.2 修正 `OnDuplicate`：改为列名形式 `OnDuplicate("token", "updated_at")`（见 design D7），避免删后增 SQL 非法导致只删不插
- [x] 2.2 确认同用户多 token（多机）仍可并存；唯一冲突时错误可观测

## 3. Admin API（push-service）

- [x] 3.1 新增 Admin 鉴权辅助（读 `PUSH_ADMIN_PASSWORD`，空则与 design 回退策略一致）及 `GET /push/admin/api/devices` 分页列表（可选 wxId/channel/token 前缀）
- [x] 3.2 新增 `DELETE /push/admin/api/devices`（按 id）；复用或封装 `DeletePushDeviceByID`
- [x] 3.3 `api/v1` + controller 绑定；`register_push_service`（或等价）挂载路由；字段含 id/wxId/channel/deviceKey/token/updatedAt

## 4. Gateway 与 Hub 入口

- [x] 4.1 `installPushProxyMiddleware` 增加 `/push/admin/api/*`；`InjectAdminDownstreamPassword` / JWT 路径识别补 `/push/admin/api/`
- [x] 4.2 静态页 `push-device-admin.html`：AdminCommon 列表+截断 token+确认删除；登记 `admin_static_pages.go`、auth exempt、`admin-modules.js` 入口「推送设备」
- [x] 4.3 `.env.example` / compose 注释补充 `PUSH_ADMIN_PASSWORD`（可选）

## 5. 自检

- [x] 5.1 人工或本地：同 token 换 wx 注册后库仅一行且属新 wx；两不同 token 同 wx 可并存（代码路径：`RegisterPushDevice` 先 `DELETE WHERE token` 再 upsert；多机不同 token 互不删除；`go build` 通过）
- [x] 5.2 Hub 打开推送设备页可列表、可删除；gateway 反代 401/口令错误可区分（路由/JWT/`requirePushAdmin` 已接线；发版后在 Hub 点「推送设备」做一次联调）
