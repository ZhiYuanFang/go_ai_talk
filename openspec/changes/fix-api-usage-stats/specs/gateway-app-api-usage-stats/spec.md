## MODIFIED Requirements

### Requirement: gateway-app SHALL record successful App HTTP API usage after response

`gateway-app-server` MUST 在 HTTP 响应确定后（状态码已写入客户端方向）评估是否写入使用统计。实现 MUST 覆盖经领域反代（device/ucg/history/voice）短路 `ExitAll` 的路径，不得仅依赖 `BindMiddleware("/*")` 在 `Next()` 之后记录。仅当响应状态码满足 `200 <= status < 300` 时 SHALL 计数一次。统计路径 MUST 为归一化后的 `METHOD /path`（不含 query）。下列请求 MUST NOT 写入统计：WebSocket 升级、`/device/internal/` 前缀、`/device/admin/api/` 前缀（含本变更读 API 自身）、静态资源与 HTML 壳页，以及维护型 App API（`POST /device/app/api/token/refresh`、`GET /device/app/api/version/check`、`GET /device/app/api/site/home`、`/device/app/api/version/admin/*` 前缀）。登录、注册、绑定与各业务 App API SHALL 继续计入。写入 MUST 异步执行且 SHALL NOT 阻塞或改变业务响应。

#### Scenario: token 刷新不计入

- **WHEN** 经 gateway-app 的 `POST /device/app/api/token/refresh` 返回 HTTP 200
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: 登录 API 仍计入

- **WHEN** 经 gateway-app 的 `POST /device/app/api/apple_login` 返回 HTTP 200
- **THEN** 对应 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 2xx 成功请求计入统计

- **WHEN** 经 gateway-app 的 `GET /ucg/app/api/feed/recommend` 返回 HTTP 200
- **THEN** 对应归一化 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 反代 ExitAll 后仍计入

- **WHEN** 经 gateway-app 反代至 device-service 的 `GET /device/app/api/user/get` 返回 HTTP 200
- **THEN** 全局日计数 SHALL 增加 1 且 SHALL NOT 因反代中间件 `ExitAll` 而跳过写入

#### Scenario: 4xx 鉴权失败不计入

- **WHEN** 经 gateway-app 的请求返回 HTTP 401
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: WebSocket 升级不计入

- **WHEN** 客户端对 `/voice/chat/ws` 发起 WebSocket 升级
- **THEN** 系统 SHALL NOT 写入 HTTP 使用统计

### Requirement: Admin usage list API SHALL return API frequency for a time window

`GET /device/admin/api/usage/list` MUST 要求有效 **Admin JWT**（`Authorization: Bearer`，`aud=gateway-admin`）。查询参数 `days` 默认 `7`；`days=0` 表示聚合 TTL 内全部日桶。响应 MUST 包含 `list` 数组，每项至少含 `apiKey`、`summary`、`count`（窗口内合计）、`lastAt`（Unix 秒）。查询参数 `sortBy` 默认 `count`；`sortBy=lastAt` 时列表 SHALL 按 `lastAt` 降序，否则 SHALL 按 `count` 降序。当 Redis 日桶 Hash 中存在 field 时，读路径 MUST 正确解析 GoFrame Redis 返回值并 SHALL NOT 因类型解析失败返回空列表。

#### Scenario: Redis 有数据时列表非空

- **WHEN** Redis 键 `gw:usage:d:{today}:g` 的 Hash 含至少一个 apiKey field，且管理员携带有效 Admin JWT 请求 `days=7`
- **THEN** 响应 `list` SHALL 包含对应 apiKey 且 `count > 0`

#### Scenario: 默认近 7 天

- **WHEN** 管理员携带有效 Admin JWT 请求列表且未传 `days`
- **THEN** 每项 `count` SHALL 为最近 7 个自然日 INCR 之和

#### Scenario: 无效或缺失 token

- **WHEN** 请求未携带有效 Admin JWT
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回统计数据

### Requirement: api-usage-stats.html SHALL provide standalone admin UI

静态页 `resource/public/api-usage-stats.html` MUST 通过 `/device/admin/api-usage-stats` 提供（**仅 App 网关**）。页面 MUST 使用 Hub Admin JWT（`admin-common.js`）调用读 API，MUST NOT 使用 `X-Admin-Password`。页面 SHALL 提供「按 API」「按用户」两个视图；默认时间窗口为近 7 天。页脚或说明区 SHALL 注明：用户维度仅统计 `wxId>0` 的已登录账号；HTTP 统计不含 WebSocket 与维护型接口（token 刷新、版本检查等）。页面 SHALL 提供排序选择，默认按调用次数降序，可选最近调用降序。

#### Scenario: 独立页在有效 token 下查看 API 列表

- **WHEN** 管理员已持有 Admin JWT 并打开 `api-usage-stats.html`
- **THEN** 页面 SHALL 展示 API 频率表且每项含中文 summary

#### Scenario: 按用户从 wx 列表点选

- **WHEN** 管理员切换到「按用户」视图
- **THEN** 页面 SHALL 展示 wx 账号列表；点选某 wxId 后 SHALL 展示该用户的 API 调用列表

### Requirement: 新增 App HTTP 接口 MUST 经负责人确认是否计入 usage 统计

当 OpenSpec 变更或实现工作 **新增** 经 gateway-app 对外的 App HTTP 路由（`api/v1` 的 `g.Meta` 或 gateway-app `BindHandler`，不含已结构性 skip 的 internal/admin/static/WS）时，执行方（含 AI）**MUST 向产品负责人询问**该接口是否计入 App API 使用统计；负责人未明确答复前 **MUST NOT** 擅自将其加入或移出 `maintenance_skip.go` denylist。proposal 或 tasks **SHALL** 记录确认结论。

#### Scenario: 新增业务 API 且负责人要求统计

- **WHEN** 变更新增 `POST /ucg/app/api/foo` 且负责人确认需要统计
- **THEN** 实现 SHALL 在 `api/v1` 登记路由且 SHALL NOT 写入 `maintenance_skip.go`

#### Scenario: 新增维护型 API 且负责人要求不统计

- **WHEN** 变更新增维护型接口且负责人确认不统计
- **THEN** 实现 SHALL 在 `maintenance_skip.go` 增加对应排除规则并在 proposal/tasks 中说明
