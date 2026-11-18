## MODIFIED Requirements

### Requirement: api-usage-stats.html SHALL provide standalone admin UI

静态页 `resource/public/api-usage-stats.html` MUST 通过 `/device/admin/api-usage-stats` 提供（**仅 App 网关**）。页面 MUST 使用 Hub Admin JWT（`admin-common.js`）调用读 API，MUST NOT 使用 `X-Admin-Password`。页面 SHALL 提供「按 API」「按用户」两个视图；默认时间窗口为近 7 天。页脚或说明区 SHALL 注明：用户维度仅统计 `wxId>0` 的已登录账号；HTTP 统计不含 WebSocket 与维护型接口（token 刷新、版本检查等）。页面 SHALL 提供排序选择，默认按调用次数降序，可选最近调用降序。「按用户」视图 MUST 调用 `GET /device/admin/api/usage/wx-list`（MUST NOT 直接调用 `wx/list`）。「按用户」wx 表格 MUST 展示列：wxId、UCG 昵称、注册时间（`createdAt=0` 显示「—」）、设备号、unionid、平台、**型号（`deviceModel`，空显示「—」）**、用户名账号。API 下钻用户表 MUST 展示列：wxId、UCG 昵称、次数、最近调用。

#### Scenario: 独立页在有效 token 下查看 API 列表

- **WHEN** 管理员已持有 Admin JWT 并打开 `api-usage-stats.html`
- **THEN** 页面 SHALL 展示 API 频率表且每项含中文 summary

#### Scenario: 按用户表展示型号列

- **WHEN** 管理员打开按用户视图且某行 `deviceModel` 为 `xiaomi`
- **THEN** 该行「型号」列 MUST 显示 `xiaomi`

#### Scenario: 型号为空

- **WHEN** 某行 `deviceModel` 为空
- **THEN** 「型号」列 MUST 显示「—」（或等价占位）

### Requirement: Admin usage wx-list API SHALL return enriched wx rows for usage stats UI

`GET /device/admin/api/usage/wx-list` MUST 要求有效 Admin JWT。查询参数 MUST 与 `GET /device/admin/api/wx/list` 对齐：`page`、`pageSize`、`q`（均可选，默认分页语义一致）。gateway-app MUST 经 device 契约取得非模拟 wx 分页列表（含 `createdAt`、**`deviceModel`**），并对当前页 wxIds 经 ucg internal `profiles/batch` 补充 `nickname`。响应 MUST 含 `list`、`total`、`page`、`pageSize`；`list` 每项 SHALL 至少含 `id`（wxId）、`deviceNo`、`unionid`、`platform`、**`deviceModel`**、`account`、`createdAt`、`nickname`。

#### Scenario: 透传型号

- **WHEN** device `wx/list` 某行 `deviceModel` 为 `iPhone11` 且经 usage wx-list 返回该行
- **THEN** 响应项 MUST 含相同 `deviceModel`
