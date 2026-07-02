## ADDED Requirements

### Requirement: Admin usage wx-list API SHALL return enriched wx rows for usage stats UI

`GET /device/admin/api/usage/wx-list` MUST 要求有效 Admin JWT。查询参数 MUST 与 `GET /device/admin/api/wx/list` 对齐：`page`、`pageSize`、`q`（均可选，默认分页语义一致）。gateway-app MUST 经 device 契约取得非模拟 wx 分页列表（含 `createdAt`），并对当前页 wxIds 经 ucg internal `profiles/batch` 补充 `nickname`。响应 MUST 含 `list`、`total`、`page`、`pageSize`；`list` 每项 SHALL 至少含 `id`（wxId）、`deviceNo`、`unionid`、`platform`、`account`、`createdAt`、`nickname`。

#### Scenario: Usage wx-list returns nickname and createdAt

- **WHEN** 管理员携带 Admin JWT 请求 `GET /device/admin/api/usage/wx-list?page=1&pageSize=20`
- **THEN** 响应 `list` 每项 MUST 含 `nickname` 与 `createdAt`
- **AND** MUST NOT 含 `is_simulated=1` 的 wxId

#### Scenario: Usage wx-list excludes simulated from total

- **WHEN** 库中存在模拟用户且存在真实用户
- **THEN** `total` MUST 仅统计 `is_simulated=0` 的行数

### Requirement: Admin usage detail list items SHALL include display nickname

`GET /device/admin/api/usage/detail` 响应 `list` 每项除 `wxId`、`count`、`lastAt` 外 MUST 含 `nickname`（string）。`nickname` MUST 经 ucg internal `profiles/batch` 取得，语义与 batch 变更一致（含无 profile 时的推导默认昵称）。仅 `wxId > 0` 的调用 SHALL 出现在列表中。

#### Scenario: Detail includes nickname per wxId

- **WHEN** 管理员携带 Admin JWT 查询某 `apiKey` 且下钻列表含 wxId=1001
- **THEN** 对应项 MUST 含非空 `nickname` 或推导默认昵称（如「家长」）

## MODIFIED Requirements

### Requirement: Admin usage detail API SHALL list wxId callers for one API

`GET /device/admin/api/usage/detail` MUST 要求有效 Admin JWT。查询参数 MUST 包含 `apiKey`；可选 `days`（默认 7）、`sortBy`（`count` 或 `lastAt`）。响应 `list` 每项 SHALL 至少包含 `wxId`、`nickname`、`count`、`lastAt`。仅 `wxId > 0` 的调用 SHALL 出现在列表中。

#### Scenario: 按 API 下钻 wxId

- **WHEN** 管理员携带 Admin JWT 查询 `apiKey=GET /ucg/app/api/feed/recommend` 且 `days=7`
- **THEN** 响应 SHALL 列出窗口内调用过该 API 的 wxId、展示昵称、次数与最近调用时间

### Requirement: Usage read APIs SHALL be served by gateway-app and excluded from device proxy

路径 `/device/admin/api/usage/*` MUST 由 gateway-app 本机处理，MUST NOT 被 `device-service` 反代吞掉。统计计数读路径（`usage/list`、`usage/detail`、`usage/user`）MUST 从 gateway Redis 读取，MUST NOT 直连 device/history/voice/ucg **数据库**。展示字段 enrich（`usage/wx-list` 的 wx 列表与昵称、`usage/detail` 的 nickname）MAY 经 **HTTP 内部契约**调用 device-service 与 ucg-service，MUST NOT 直连他域 DAO 或库表。

#### Scenario: 读 API 不经 device-service 反代

- **WHEN** 管理员请求 `GET /device/admin/api/usage/list`
- **THEN** gateway-app SHALL 本地响应且 SHALL NOT 将请求转发至 device-service

#### Scenario: Wx-list orchestrates via HTTP contracts

- **WHEN** 管理员请求 `GET /device/admin/api/usage/wx-list`
- **THEN** gateway-app SHALL 本地处理该请求
- **AND** MUST 经 HTTP 契约获取 wx 列表与 nickname，MUST NOT 直连 MySQL

### Requirement: api-usage-stats.html SHALL provide standalone admin UI

静态页 `resource/public/api-usage-stats.html` MUST 通过 `/device/admin/api-usage-stats` 提供（**仅 App 网关**）。页面 MUST 使用 Hub Admin JWT（`admin-common.js`）调用读 API，MUST NOT 使用 `X-Admin-Password`。页面 SHALL 提供「按 API」「按用户」两个视图；默认时间窗口为近 7 天。页脚或说明区 SHALL 注明：用户维度仅统计 `wxId>0` 的已登录账号；HTTP 统计不含 WebSocket 与维护型接口（token 刷新、版本检查等）。页面 SHALL 提供排序选择，默认按调用次数降序，可选最近调用降序。「按用户」视图 MUST 调用 `GET /device/admin/api/usage/wx-list`（MUST NOT 直接调用 `wx/list`）。「按用户」wx 表格 MUST 展示列：wxId、UCG 昵称、注册时间（`createdAt=0` 显示「—」）、设备号、unionid、平台、用户名账号。API 下钻用户表 MUST 展示列：wxId、UCG 昵称、次数、最近调用。

#### Scenario: 独立页在有效 token 下查看 API 列表

- **WHEN** 管理员已持有 Admin JWT 并打开 `api-usage-stats.html`
- **THEN** 页面 SHALL 展示 API 频率表且每项含中文 summary

#### Scenario: 按用户从 wx 列表点选

- **WHEN** 管理员切换到「按用户」视图
- **THEN** 页面 SHALL 展示非模拟 wx 账号列表且含 UCG 昵称与注册时间
- **AND** 点选某 wxId 后 SHALL 展示该用户的 API 调用列表

#### Scenario: API 下钻展示昵称

- **WHEN** 管理员在「按 API」视图对某 API 点击下钻
- **THEN** 下钻用户表 MUST 展示每行 wxId 的 UCG 昵称
