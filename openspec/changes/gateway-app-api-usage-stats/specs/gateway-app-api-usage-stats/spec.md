## ADDED Requirements

### Requirement: gateway-app SHALL record successful App HTTP API usage after response

`gateway-app-server` MUST 在 HTTP 请求处理完成后（`Middleware.Next()` 之后）评估是否写入使用统计。仅当响应状态码满足 `200 <= status < 300` 时 SHALL 计数一次。统计路径 MUST 为归一化后的 `METHOD /path`（不含 query）。下列请求 MUST NOT 写入统计：WebSocket 升级、`/device/internal/` 前缀、`/device/admin/api/` 前缀（含本变更读 API 自身）、静态资源与 HTML 壳页。写入 MUST 异步执行且 SHALL NOT 阻塞或改变业务响应。

#### Scenario: 2xx 成功请求计入统计

- **WHEN** 经 gateway-app 的 `GET /ucg/app/api/feed/recommend` 返回 HTTP 200
- **THEN** 对应归一化 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 4xx 鉴权失败不计入

- **WHEN** 经 gateway-app 的请求返回 HTTP 401
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: WebSocket 升级不计入

- **WHEN** 客户端对 `/voice/chat/ws` 发起 WebSocket 升级
- **THEN** 系统 SHALL NOT 写入 HTTP 使用统计

### Requirement: API paths SHALL be normalized and annotated with Chinese summary

系统 MUST 在启动时自 `api/v1` 路由元数据构建注册表，将动态路径段归一化为与 `g.Meta path` 一致的模板（如 `/ucg/app/api/posts/{id}`）。管理端展示的 `summary` MUST 取自注册表中文说明；未命中注册表时 SHALL 显示「未登记」并保留归一化或原始 apiKey。

#### Scenario: 动态帖子 ID 归一化

- **WHEN** 统计 `GET /ucg/app/api/posts/42` 的成功调用
- **THEN** 聚合键 SHALL 为 `GET /ucg/app/api/posts/{id}` 且 summary SHALL 为「获取单帖」或注册表中等价中文

#### Scenario: 未登记路径

- **WHEN** 某成功请求路径无法匹配任何 `api/v1` 模板
- **THEN** 列表项 summary SHALL 为「未登记」且仍 SHALL 展示 apiKey

### Requirement: Usage counters SHALL be stored in Redis with daily buckets

统计存储 MUST 使用 gateway-app 可访问的 Redis。全局 API 计数 MUST 按日分桶（`YYYYMMDD`）使用 INCR；当且仅当 `wxId > 0` 时 MUST 额外写入用户维度日计数。系统 MUST 维护全局与用户维度的 `last_at`（最近一次成功调用的 Unix 秒）。日桶 key MUST 设置 TTL（不少于 90 天）以支持近 30 天查询。

#### Scenario: 登录用户双维度计数

- **WHEN** `wxId=1001` 的用户成功调用 `POST /ucg/app/api/posts`
- **THEN** 全局 apiKey 日计数与用户 `wx:1001` 维度日计数 SHALL 各增加 1

#### Scenario: 无 wxId 仅全局计数

- **WHEN** 请求可解析 `deviceNo` 但 `wxId=0` 且返回 2xx
- **THEN** 全局 apiKey 日计数 SHALL 增加 1 且 SHALL NOT 写入用户维度键

### Requirement: Admin usage list API SHALL return API frequency for a time window

`GET /device/admin/api/usage/list` MUST 要求 Header `X-Admin-Password` 有效（与设备管理口令一致）。查询参数 `days` 默认 `7`；`days=0` 表示聚合 TTL 内全部日桶。响应 MUST 包含 `list` 数组，每项至少含 `apiKey`、`summary`、`count`（窗口内合计）、`lastAt`（Unix 秒）。列表 SHOULD 按 `count` 降序。

#### Scenario: 默认近 7 天

- **WHEN** 管理员携带正确口令请求列表且未传 `days`
- **THEN** 每项 `count` SHALL 为最近 7 个自然日 INCR 之和

#### Scenario: 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回统计数据

### Requirement: Admin usage detail API SHALL list wxId callers for one API

`GET /device/admin/api/usage/detail` MUST 要求有效 `X-Admin-Password`。查询参数 MUST 包含 `apiKey`；可选 `days`（默认 7）。响应 `list` 每项 SHALL 至少包含 `wxId`、`count`、`lastAt`。仅 `wxId > 0` 的调用 SHALL 出现在列表中。

#### Scenario: 按 API 下钻 wxId

- **WHEN** 管理员查询 `apiKey=GET /ucg/app/api/feed/recommend` 且 `days=7`
- **THEN** 响应 SHALL 列出窗口内调用过该 API 的 wxId、次数与最近调用时间

### Requirement: Admin usage user API SHALL list APIs called by one wxId

`GET /device/admin/api/usage/user` MUST 要求有效 `X-Admin-Password`。查询参数 MUST 包含 `wxId`（正整数）；可选 `days`（默认 7）。响应 `list` 每项 SHALL 至少包含 `apiKey`、`summary`、`count`、`lastAt`。

#### Scenario: 按 wxId 查看 API 分布

- **WHEN** 管理员查询 `wxId=1001` 且 `days=7`
- **THEN** 响应 SHALL 列出该用户在窗口内成功调用过的 API、中文说明、次数与最近调用时间

### Requirement: Usage read APIs SHALL be served by gateway-app and excluded from device proxy

路径 `/device/admin/api/usage/*` MUST 由 gateway-app 本机处理，MUST NOT 被 `device-service` 反代吞掉。读 API 仅读取 gateway Redis 统计，MUST NOT 直连 device/history/voice/ucg 数据库。

#### Scenario: 读 API 不经 device-service

- **WHEN** 管理员请求 `GET /device/admin/api/usage/list`
- **THEN** gateway-app SHALL 本地响应且 SHALL NOT 将请求转发至 device-service

### Requirement: api-usage-stats.html SHALL provide standalone admin UI

静态页 `resource/public/api-usage-stats.html` MUST 通过 `/device/admin/api-usage-stats` 提供（主网关与 app 网关均可访问，与 `qa-records` 一致）。页面 MUST 使用 `X-Admin-Password` 登录后调用上述读 API。页面 SHALL 提供「按 API」「按用户」两个视图；默认时间窗口为近 7 天。页脚或说明区 SHALL 注明：用户维度仅统计 `wxId>0` 的已登录账号。

#### Scenario: 独立页登录后查看 API 列表

- **WHEN** 管理员在 `api-usage-stats.html` 输入正确口令
- **THEN** 页面 SHALL 展示 API 频率表且每项含中文 summary

#### Scenario: 按用户查询

- **WHEN** 管理员在「按用户」视图输入 wxId 并查询
- **THEN** 页面 SHALL 展示该 wxId 的 API 调用列表

### Requirement: device admin SHALL link to API usage stats from device record card

`resource/public/admin.html` 的**设备记录**卡片头部 `card-actions` MUST 包含指向 `/device/admin/api-usage-stats` 的链接，文案为「功能使用统计」（或等价中文）。链接在管理员登录成功后 SHALL 可见（与问答库/反馈「展开更多」一致的显示时机）。

#### Scenario: 设备记录区入口

- **WHEN** 管理员登录设备管理页并进入主界面
- **THEN** 设备记录卡片 `card-actions` SHALL 显示「功能使用统计」链接
