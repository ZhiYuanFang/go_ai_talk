## ADDED Requirements

### Requirement: 运维 MUST 能分页查看并删除 push_device

系统 MUST 在 push-service 提供经 Admin 鉴权的管理 API，用于分页列出 `push_device` 行（至少含 `id`、`wxId`、`channel`、`deviceKey`、`token`、`updatedAt`），并支持按行 `id` 删除。列表 MUST 支持合理分页；MAY 支持按 `wxId`、`channel` 或 token 前缀筛选。删除成功后该行 MUST 不再参与后续按 wx 扇出推送。API MUST 经 gateway-app 反代暴露（建议前缀 `/push/admin/api/`），MUST 使用与现网一致的 Admin JWT + 下游 `X-Admin-Password` 注入模式；浏览器 MUST NOT 直接携带运维口令头。

#### Scenario: 列表可见

- **WHEN** 已登录运维人员请求设备列表接口且鉴权通过
- **THEN** 响应 MUST 返回当前 `push_device` 分页数据（字段至少含 id 与 wxId、channel、token、updatedAt）

#### Scenario: 手动删除

- **WHEN** 运维人员对某 `id` 调用删除且鉴权通过
- **THEN** 该行 MUST 从 `push_device` 移除，后续对该原 `wxId` 的推送若无其它 token 则 MUST 跳过该设备投递

### Requirement: 运维 Hub MUST 提供推送设备管理入口与静态页

系统 MUST 在 gateway-app 注册推送设备管理静态页（建议 `/device/admin/push-device-admin.html`），并在运维 Hub 模块列表（`admin-modules`）中提供可点击入口。页面 MUST 使用 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch` 访问上述 Admin API，支持查看列表与确认后删除。页面展示 token 时 MUST 截断或脱敏，MUST NOT 以明文完整 token 为默认展示（MAY 提供展开，但默认截断）。

#### Scenario: Hub 可进入

- **WHEN** 运维人员打开运维 Hub 并已登录
- **THEN** MUST 能通过登记入口打开推送设备管理页并加载列表（在 API 可用前提下）
