## ADDED Requirements

### Requirement: client-usage wx-list MUST 透传 babyName 并返回窗口内 lastAt

`GET /device/admin/api/client-usage/wx-list`（gateway-app 本机、Admin JWT）在既有分页与 `q` 搜索语义不变的前提下，响应 `list[]` 每项 MUST 包含：

- `babyName`（string）：来自 device `wx/list` 透传；无则空串。
- `lastAt`（int64）：该 wxId 在统计时间窗口内最近一次成功客户端上报的 Unix 秒；窗口内无记录 MUST 为 `0`。

请求 MUST 接受可选 `days`（与其它 client-usage 读接口一致，窗口上限 30 天）。`lastAt` MUST 经 `cachekit` 读取既有 client-usage per-wx last 键并取 max，再按 `days` 过滤；MUST NOT 直连他域库表。

#### Scenario: 透传宝宝名

- **WHEN** device 侧该 wx 有非空 `babyName` 且出现在 client-usage wx-list 本页
- **THEN** 响应项 MUST 含相同 `babyName`

#### Scenario: 无上报时 lastAt 为 0

- **WHEN** 某 wxId 在所选窗口内无 client-usage 记录
- **THEN** 对应项 `lastAt` MUST 为 `0`
