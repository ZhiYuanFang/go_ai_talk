## MODIFIED Requirements

### Requirement: Admin usage list API SHALL return API frequency for a time window

`GET /device/admin/api/usage/list` MUST 要求有效 **Admin JWT**（`Authorization: Bearer`，`aud=gateway-admin`）。查询参数 `days` 默认 `7`；`days=0` 表示聚合 TTL 内全部日桶。响应 MUST 包含 `list` 数组，每项至少含 `apiKey`、`summary`、`count`（窗口内合计）、`lastAt`（Unix 秒）。查询参数 `sortBy` 默认 `count`；`sortBy=lastAt` 时列表 SHALL 按 `lastAt` 降序，否则 SHALL 按 `count` 降序。当 Redis 日桶 Hash 中存在 field 时，读路径 MUST 正确解析 GoFrame Redis 返回值（含 `HGETALL` 经 adapter 转为 flat `[]string` 的情形）并 SHALL NOT 因类型解析失败返回空列表。当 apiKey 在 `api/v1` 的 `g.Meta` 中已登记（精确或模板匹配）时，`summary` MUST 返回对应中文说明，SHALL NOT 全部为「未登记」。

#### Scenario: Redis 有数据时列表非空

- **WHEN** Redis 键 `gw:usage:d:{today}:g` 的 Hash 含至少一个 apiKey field，且管理员携带有效 Admin JWT 请求 `days=7`
- **THEN** 响应 `list` SHALL 包含对应 apiKey 且 `count > 0`

#### Scenario: GoFrame HGETALL flat []string 可读

- **WHEN** `g.Redis().Do("HGETALL", key)` 返回的 `*gvar.Var` 经 adapter 内部表示为 flat `[]string`（非 map）
- **THEN** `ListAPIs` 等读路径 SHALL 仍正确聚合 field 与计数，SHALL NOT 返回空 `list`

#### Scenario: 已登记路由展示中文 summary

- **WHEN** gateway-app 在 Docker 容器内运行，Redis 含 apiKey `GET /ucg/app/api/feed/recommend`，且 `api/v1` 中该路由 `summary` 为「推荐 Feed」
- **THEN** `GET /device/admin/api/usage/list` 对应项的 `summary` SHALL 为「推荐 Feed」，SHALL NOT 为「未登记」

#### Scenario: 模板 apiKey 展示中文 summary

- **WHEN** Redis 含 apiKey `POST /ucg/app/api/posts/{id}/like`，且 `api/v1` 已登记该模板与 summary「点赞」
- **THEN** usage list 对应项 `summary` SHALL 为「点赞」

#### Scenario: 历史 raw 路径 apiKey 仍可匹配 summary

- **WHEN** Redis 含 raw apiKey `POST /ucg/app/api/posts/123/like`（数字 postId），且 `api/v1` 已登记模板 `POST /ucg/app/api/posts/{id}/like` summary「点赞」
- **THEN** usage list 对应项 `summary` SHALL 为「点赞」

#### Scenario: 默认近 7 天

- **WHEN** 管理员携带有效 Admin JWT 请求列表且未传 `days`
- **THEN** 每项 `count` SHALL 为最近 7 个自然日 INCR 之和

#### Scenario: 无效或缺失 token

- **WHEN** 请求未携带有效 Admin JWT
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回统计数据

### Requirement: api-usage-stats.html SHALL provide standalone admin UI

静态页 `resource/public/api-usage-stats.html` MUST 通过 `/device/admin/api-usage-stats` 提供（**仅 App 网关**）。页面 MUST 使用 Hub Admin JWT（`admin-common.js`）调用读 API，MUST NOT 使用 `X-Admin-Password`。页面 SHALL 提供「按 API」「按用户」两个视图；默认时间窗口为近 7 天。页脚或说明区 SHALL 注明：用户维度仅统计 `wxId>0` 的已登录账号；HTTP 统计不含 WebSocket 与维护型接口（token 刷新、版本检查等）。页面 SHALL 提供排序选择，默认按调用次数降序，可选最近调用降序。表格「中文说明」列 SHALL 展示读 API 返回的 `summary`（已登记路由不得长期全部为「未登记」）。

#### Scenario: 独立页在有效 token 下查看 API 列表

- **WHEN** 管理员已持有 Admin JWT 并打开 `api-usage-stats.html`，且 Redis 含已登记 App API 调用
- **THEN** 页面 SHALL 展示 API 频率表且「中文说明」列含非「未登记」的 summary（对已登记路由）

#### Scenario: 按用户从 wx 列表点选

- **WHEN** 管理员切换到「按用户」视图
- **THEN** 页面 SHALL 展示 wx 账号列表；点选某 wxId 后 SHALL 展示该用户的 API 调用列表
