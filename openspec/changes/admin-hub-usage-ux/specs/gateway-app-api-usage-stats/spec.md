## ADDED Requirements

### Requirement: usage wx-list MUST 透传 babyName 并返回窗口内 lastAt

`GET /device/admin/api/usage/wx-list`（gateway-app 本机、Admin JWT）在既有分页与 `q` 搜索语义不变的前提下，响应 `list[]` 每项 MUST 包含：

- `babyName`（string）：来自 device `wx/list` 已返回字段的透传；无则空串。
- `lastAt`（int64）：该 `id`（wxId）在统计时间窗口内最近一次成功 App API 调用的 Unix 秒；窗口内无记录 MUST 为 `0`。

请求 MUST 接受可选 `days`（与其它 usage 读接口语义一致：默认 7；`0` 表示实现约定的全部窗口）。计算 `lastAt` MUST 使用既有 Redis per-wx last 键（经 `cachekit`），对本页 wxId 取各 apiKey 时间戳的最大值并按 `days` 窗口过滤；MUST NOT 新增跨库直查，MUST NOT 新建与本需求无关的 Redis 键族。

#### Scenario: 透传宝宝名

- **WHEN** device `wx/list` 某行 `babyName` 为「豆豆」且经 usage wx-list 返回该行
- **THEN** 响应项 MUST 含 `babyName` 为「豆豆」

#### Scenario: 窗口内无调用

- **WHEN** 某 wxId 在所选 `days` 窗口内无成功 API 统计
- **THEN** 对应 list 项 `lastAt` MUST 为 `0`

#### Scenario: 窗口内有调用

- **WHEN** 某 wxId 在窗口内对多个 apiKey 有 last 时间戳
- **THEN** `lastAt` MUST 等于这些时间戳的最大值（且落在窗口内）
