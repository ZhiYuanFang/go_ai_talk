## MODIFIED Requirements

### Requirement: client-usage wx-list MUST 透传 babyName 并返回窗口内 lastAt

`GET /device/admin/api/client-usage/wx-list`（gateway-app 本机、Admin JWT）在既有分页与 `q` 搜索语义不变的前提下，响应 `list[]` 每项 MUST 包含：

- `babyName`（string）：来自 device `wx/list` 透传；无则空串。
- `lastAt`（int64）：该 wxId 在统计时间窗口内最近一次成功客户端上报的 Unix 秒；窗口内无记录 MUST 为 `0`。
- **`deviceModel`（string）**：来自 device `wx/list` 透传的设备型号展示文案；无则空串。

请求 MUST 接受可选 `days`（与其它 client-usage 读接口一致，窗口上限 30 天）。`lastAt` MUST 经 `cachekit` 读取既有 client-usage per-wx last 键并取 max，再按 `days` 过滤；MUST NOT 直连他域库表。

#### Scenario: 透传宝宝名

- **WHEN** device 侧该 wx 有非空 `babyName` 且出现在 client-usage wx-list 本页
- **THEN** 响应项 MUST 含相同 `babyName`

#### Scenario: 透传型号

- **WHEN** device 侧该 wx 有非空 `deviceModel` 且出现在 client-usage wx-list 本页
- **THEN** 响应项 MUST 含相同 `deviceModel`

#### Scenario: 无上报时 lastAt 为 0

- **WHEN** 某 wxId 在所选窗口内无 client-usage 记录
- **THEN** 对应项 `lastAt` MUST 为 `0`

---

## ADDED Requirements

### Requirement: client-usage-stats.html SHALL show device model column

`resource/public/client-usage-stats.html` 的「按用户」wx 表格 MUST 在「平台」列旁展示「型号」列，值为 `deviceModel`；空值 MUST 显示「—」（或等价占位）。MUST NOT 要求运维从枚举下拉选择型号。

#### Scenario: 有型号时展示原文

- **WHEN** 管理员打开客户端使用统计按用户表且某行 `deviceModel` 为 `huawei`
- **THEN** 该行型号列 MUST 显示 `huawei`

## MODIFIED Requirements

### Requirement: api-usage-stats.html SHALL provide standalone admin UI

`api-usage-stats.html` MUST 经 `/device/admin/api-usage-stats`（仅 App 网关）提供，使用 Hub Admin JWT，MUST NOT 使用 `X-Admin-Password`。页面 SHALL 提供「按 API」「按用户」视图与排序；「按用户」MUST 调 `GET /device/admin/api/usage/wx-list`。按用户表 MUST 含型号列（`deviceModel`，空显示「—」）。 The system SHALL keep nickname/createdAt drill-down behavior.

#### Scenario: 独立页在有效 token 下查看 API 列表

- **WHEN** 管理员已持有 Admin JWT 并打开 `api-usage-stats.html`
- **THEN** 页面 SHALL 展示 API 频率表且每项含中文 summary

#### Scenario: 按用户从 wx 列表点选

- **WHEN** 管理员切换到「按用户」视图
- **THEN** 页面 SHALL 展示非模拟 wx 账号列表并含 UCG 昵称与注册时间
- **AND** 点选某 wxId 后 SHALL 展示该用户的 API 频率列表

#### Scenario: API 下钻展示昵称

- **WHEN** 管理员在「按 API」视图点某 API 进入下钻
- **THEN** 下钻用户表 MUST 展示每个 wxId 的 UCG 昵称

#### Scenario: 按用户表展示型号列

- **WHEN** 管理员打开按用户视图且某行 `deviceModel` 为 `xiaomi`
- **THEN** 该行「型号」列 MUST 显示 `xiaomi`

#### Scenario: 型号为空

- **WHEN** 某行 `deviceModel` 为空
- **THEN** 「型号」列 MUST 显示「—」（或等价占位）

### Requirement: Admin usage wx-list API SHALL return enriched wx rows for usage stats UI

`GET /device/admin/api/usage/wx-list` MUST 要求 Admin JWT；分页参数与 `wx/list` 对齐。gateway-app MUST 经 device 取得非模拟 wx 列表（含 `createdAt`、`deviceModel`），并对当前页经 ucg `profiles/batch` 补 `nickname`。响应 `list` 每项 SHALL 含 `id`、`deviceNo`、`unionid`、`platform`、`deviceModel`、`account`、`createdAt`、`nickname`。

#### Scenario: Usage wx-list returns nickname and createdAt

- **WHEN** 管理员携带 Admin JWT 请求 `GET /device/admin/api/usage/wx-list?page=1&pageSize=20`
- **THEN** 响应 `list` 每项 MUST 含 `nickname` 与 `createdAt`
- **AND** MUST NOT 含 `is_simulated=1` 的 wxId

#### Scenario: Usage wx-list excludes simulated from total

- **WHEN** 库中存在模拟用户与大量真实用户
- **THEN** `total` MUST 仅统计 `is_simulated=0` 的行数

#### Scenario: 透传型号

- **WHEN** device `wx/list` 某行 `deviceModel` 为 `iPhone11` 且经 usage wx-list 返回该行
- **THEN** 响应项 MUST 含相同 `deviceModel`
