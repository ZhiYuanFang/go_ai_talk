# gateway-edge

## Purpose

gateway-app / gateway-service 边缘：CORS、JWT、反代、WS 委托、官网页、usage 统计与版本检查。 Traffic and policy edge without domain business workers.

<!-- migrated-from: client-feature-usage-stats, gateway-admin-jwt, gateway-app-api-usage-stats, gateway-app-cors, gateway-app-cors-reverse-proxy, gateway-app-device-login-device-no, gateway-app-jwt-device-no-header, gateway-app-official-site, gateway-app-path-only-assets, gateway-app-server, gateway-app-version-admin, gateway-app-version-admin-crud, gateway-app-version-check, gateway-no-business-workers, gateway-policy-layer-consolidation, gateway-route-middleware-domain-isolation, gateway-voice-http-proxy, gateway-ws-delegation-convergence, gateway-ws-edge-proxy, ucg-gateway-proxy, wechat-ios-universal-links, wechat-site-domain-verify -->

## Requirements

### Requirement: App client-usage report MUST accept logged-in feature events

`gateway-app-server` MUST 提供 `POST /device/app/api/client-usage/report`（或 design 锁定的等价本机 App 路径）。请求体 MUST 包含非空字符串 `featureId`（客户端自由字符串；服务端 MUST NOT 强制枚举白名单）。系统 MUST 从登录态解析 `wxId`；当 `wxId <= 0` 时 MUST 拒绝写入且 MUST NOT 记入 Redis。当 `featureId` 为空或超过实现设定的最大长度时 MUST 拒绝。成功时 MUST 使用**服务端时间**记录事件，MUST NOT 信任客户端时间戳。

#### Scenario: 登录用户上报成功

- **WHEN** 已登录用户提交合法 `featureId` 且未触发限流且非模拟用户
- **THEN** 系统 SHALL 返回成功，并 SHALL 更新聚合与该用户时间线

#### Scenario: 未登录拒绝

- **WHEN** 请求无法解析出正整数 `wxId`
- **THEN** 系统 MUST 拒绝，且 MUST NOT 写入任何 client-usage Redis 键

### Requirement: Simulated users MUST NOT be recorded

当关联 `wxId` 被判定为模拟用户（与 App API usage 同源判定）时，client-usage report MUST NOT 写入聚合或时间线。

#### Scenario: 模拟用户上报不落库

- **WHEN** 模拟用户调用 report 且请求其余字段合法
- **THEN** 系统 SHALL NOT 增加聚合计数，SHALL NOT 追加时间线（响应可为成功空结果以免客户端重试）

### Requirement: Per-user rate limit of one successful report per 3 seconds

同一 `wxId` 在任意 3 秒窗口内 MUST 最多完成一次成功的 client-usage 写入。超出限制时 MUST 拒绝本次写入（例如 HTTP 429 或明确业务错误），MUST NOT 更新聚合或时间线。

#### Scenario: 3 秒内第二次上报被拒

- **WHEN** 同一用户在成功上报后 3 秒内再次调用 report
- **THEN** 系统 MUST 拒绝第二次写入，且 Redis 聚合/时间线 MUST NOT 因第二次请求增加

### Requirement: Redis aggregation window MUST be at most 30 days

系统 MUST 以日桶等形式维护按 `featureId`、按用户、以及功能×用户交叉的聚合计数，查询窗口 MUST 支持至多 **30 天**（例如 `days=7` 与 `days=30`）。聚合数据 MUST 仅存 Redis，MUST 经 `cachekit` 访问，键 MUST 由 platform builder 登记。产品接受 Redis 丢失导致统计丢失。

#### Scenario: 按功能查询近 7 天

- **WHEN** 管理员请求功能维度列表且 `days=7`
- **THEN** 系统 SHALL 返回窗口内各 `featureId` 的合计次数（及实现约定的最近时间字段，若有）

### Requirement: Per-user timeline MUST store event name and server time only

每个真实用户 MUST 可查询上报时间线。每条时间线条目 MUST 仅包含：`featureId`（事件名）与服务端 Unix 时间。MUST NOT 要求或持久化客户端自定义 `extra` 载荷作为一期范围。

#### Scenario: 用户时间线字段

- **WHEN** 管理员查询某 `wxId` 的时间线
- **THEN** 每条记录 SHALL 含 `featureId` 与服务端时间，且 SHALL NOT 依赖客户端提交的时间戳

### Requirement: Per-user timeline MUST cap at 10000 entries and retain about 30 days

单用户时间线 MUST 最多保留 **10000** 条；超出时 MUST 丢弃最旧条目。时间线保留策略 MUST 与约 **30 天** 有效期对齐（key TTL 与/或按日过期）。条数上限与时间窗口 MUST 同时生效。

#### Scenario: 超过一万条裁剪

- **WHEN** 某用户时间线已有 10000 条且再次成功上报
- **THEN** 系统 SHALL 追加新事件并 SHALL 移除最旧事件，使条数不超过 10000

### Requirement: Report API MUST NOT count toward App API usage stats

`POST /device/app/api/client-usage/report`（精确 METHOD+path）MUST 登记于 `maintenance_skip`（或等价 denylist），MUST NOT 计入 App API 使用统计。负责人已确认本接口不统计。

#### Scenario: report 不计入 API usage

- **WHEN** 用户成功调用 client-usage report
- **THEN** App API 使用统计对应计数 MUST NOT 因该请求递增

### Requirement: Admin client-usage APIs and Hub page

`gateway-app` MUST 提供需 Admin JWT 的读接口，至少支持：按功能聚合列表、按用户列表/下钻、单用户时间线。路径 MUST 由 gateway-app 本机处理，MUST NOT 直连他域数据库。Hub MUST 增加独立入口「客户端使用统计」，交互对齐现有 App API 使用统计页（功能/用户维度；时间范围不超过 30 天）。页面 MUST 提示数据仅 Redis、可丢失，且时间线可能因 10000 条上限短于次数合计。

#### Scenario: Hub 按用户查看时间线

- **WHEN** 管理员打开「客户端使用统计」并进入某用户详情
- **THEN** 系统 SHALL 展示该用户聚合摘要与时间线（事件名 + 服务端时间）

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

---

### Requirement: gateway-app SHALL provide admin login issuing Admin JWT

`gateway-app-server` MUST 提供 `POST /device/admin/api/login`（`Content-Type: application/json`），请求体含 `username` 与 `password`。校验 MUST 对照环境变量 `GATEWAY_APP_ADMIN_USERNAME`（默认 `admin`）与 `GATEWAY_APP_ADMIN_PASSWORD`。成功时响应 MUST 为 `{ code: 0, data: { accessToken, expiresIn } }`，其中 `accessToken` 为 HS256 JWT，`aud` MUST 为 `gateway-admin`，`iss` MUST 为 `gateway-app-server`。未配置 `GATEWAY_APP_ADMIN_PASSWORD` 时 MUST 返回 503 且 SHALL NOT 签发 token。登录接口 MUST 列入 Bearer 白名单。

#### Scenario: 正确账号密码登录

- **WHEN** 客户端提交与 env 一致的 username/password
- **THEN** 系统 SHALL 返回 `code=0` 及非空 `accessToken`

#### Scenario: 密码错误

- **WHEN** 客户端提交错误 password
- **THEN** 系统 SHALL 返回 401 且 SHALL NOT 签发 token

#### Scenario: 未配置 admin 密码

- **WHEN** `GATEWAY_APP_ADMIN_PASSWORD` 为空且客户端请求 login
- **THEN** 系统 SHALL 返回 503 语义（管理未启用）

### Requirement: Admin JWT and user access JWT SHALL be mutually isolated

Bearer 中间件 MUST 区分 Admin JWT（`aud=gateway-admin`）与用户 access JWT。管理 API 路径（见下条）MUST 要求有效 Admin JWT。App/UCG 用户 API MUST 要求用户 JWT 且 MUST 拒绝仅含 Admin JWT 的请求。用户 JWT MUST NOT 访问管理 API（返回 403）。

#### Scenario: Admin JWT 访问设备管理 API

- **WHEN** 客户端对 `GET /device/admin/api/user/list` 携带有效 Admin JWT
- **THEN** 请求 SHALL 通过 gateway-app Bearer 校验并进入下游或本机 handler

#### Scenario: Admin JWT 访问用户 profile API

- **WHEN** 客户端对受保护 App API 仅携带 Admin JWT
- **THEN** gateway-app SHALL 返回 403

#### Scenario: 用户 JWT 访问管理 API

- **WHEN** 客户端对 `GET /device/admin/api/event/list` 仅携带用户 access JWT
- **THEN** gateway-app SHALL 返回 403

### Requirement: gateway-app SHALL inject downstream admin passwords server-side

校验 Admin JWT 成功后，gateway-app MUST 在转发或本机处理前注入：`/device/admin/api/*` 与 `/device/app/api/version/admin/*` 注入 `X-Admin-Password` 等于 `DEVICE_ADMIN_PASSWORD`；`/ucg/admin/api/*` 注入 `X-Admin-Password` 等于 `UCG_ADMIN_PASSWORD`。Hook 入口 MUST 删除客户端传入的 `X-Admin-Password`，防止伪造。

#### Scenario: 浏览器不传 X-Admin-Password 仍可调用 device 管理 API

- **WHEN** 管理员仅携带 Admin JWT 请求 `GET /device/admin/api/user/list`
- **THEN** gateway-app 反代至 device-service 的请求 SHALL 含服务端注入的有效 `X-Admin-Password` 且 device-service SHALL 返回业务数据

#### Scenario: 客户端伪造 X-Admin-Password 无效

- **WHEN** 客户端携带错误 `X-Admin-Password` 与有效 Admin JWT
- **THEN** 下游 SHALL 仍使用网关注入口令且 SHALL NOT 使用客户端伪造值

### Requirement: Web admin pages SHALL be served only by gateway-app

下列静态路由 MUST 仅由 `gateway-app-server` 注册：`/device/admin`、`/device/admin/qa-records`、`/device/admin/feedback-records`、`/device/admin/api-usage-stats`、`/device/admin/ucg-admin.html`、`/device/app/version-admin.html`、`/device/history/*deviceNo`（history 壳页）。`gateway-service` MUST NOT 再 `ServeFile` 上述路径。

#### Scenario: 9702 可打开设备管理页

- **WHEN** 客户端 `GET /device/admin` 访问 App 网关
- **THEN** 系统 SHALL 返回 `admin.html`

#### Scenario: 9701 不直接提供 admin 静态页

- **WHEN** 客户端 `GET /device/admin` 访问主网关且迁移已完成
- **THEN** 系统 SHALL NOT 返回 admin 静态 HTML 作为 200 正文（SHALL 302 或 runbook  documented 等价行为）

### Requirement: gateway-service SHALL redirect legacy admin URLs to App gateway

主网关 MUST 对 `/device/admin` 及子路径（静态 admin 页）、以及 `/device/history/*` 壳页（不含 `/device/history/api/*`）返回 **302**，`Location` 为 `GATEWAY_APP_PUBLIC_BASE_URL` 与请求路径拼接；env 未配置时 MAY fallback 至同 host 的 App 网关端口（如 `:9702`）。

#### Scenario: 旧 bookmark 跳转

- **WHEN** 用户访问 `https://example.com:9701/device/admin`
- **THEN** 浏览器 SHALL 被重定向至 App 网关等价路径

### Requirement: admin front-end SHALL use shared Bearer client

仓库 MUST 提供 `resource/public/admin-common.js`（登录、`Authorization: Bearer` fetch、logout）与 `resource/public/admin-modules.js`（模块登记）。所有管理静态页 MUST 通过 admin-common 调用 API，MUST NOT 在浏览器侧发送 `X-Admin-Password`。`admin.html` MUST 从 admin-modules 渲染导航链接，MUST NOT 使用 `gatewayAppBase()` 端口配对。

#### Scenario: Hub 登录后子页共用 token

- **WHEN** 管理员在 `/device/admin` 登录成功
- **THEN** 打开 `/device/admin/qa-records` SHALL 使用同一 Admin JWT 加载数据而无需再次输入口令

### Requirement: New admin modules SHALL be registered in admin-modules.js

新增 Web 管理模块时 MUST 在 `admin-modules.js` 增加条目（至少含 `id`、`title`、`pagePath`、`apiPrefixes`），并在 gateway-app 静态路由单点注册表中增加 `pagePath`。PR MUST NOT 新增仅主网关可见的 admin 静态路由。

#### Scenario: 登记与路由一致

- **WHEN** admin-modules 声明 `pagePath=/device/admin/foo`
- **THEN** gateway-app MUST 注册该路径对应的静态文件 handler

---

### Requirement: gateway-app SHALL record successful App HTTP API usage after response

`gateway-app-server` MUST 在 HTTP 响应确定后（状态码已写入客户端方向）评估是否写入使用统计。 实现 MUST 覆盖经领域反代（device/ucg/history/voice）短路 `ExitAll` 的路径，不得仅依赖 `BindMiddleware("/*")` 在 `Next()` 之后记录。 仅当响应状态码满足 `200 <= status < 300` 时 SHALL 计数一次。 统计路径 MUST 为归一化后的 `METHOD /path`（不含 query）。 下列请求 MUST NOT 写入统计：WebSocket 升级、`/device/internal/` 前缀、`/device/admin/api/` 前缀（含本变更读 API 自身）、静态资源与 HTML 壳页，以及维护型 App API（`POST /device/app/api/token/refresh`、`GET /device/app/api/version/check`、`GET.

#### Scenario: token 刷新不计入

- **WHEN** 经 gateway-app 的 `POST /device/app/api/token/refresh` 返回 HTTP 200
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: GET 评论列表不计入

- **WHEN** 经 gateway-app 的 `GET /ucg/app/api/posts/123/comments` 返回 HTTP 200
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: POST 评论仍计入

- **WHEN** 经 gateway-app 的 `POST /ucg/app/api/posts/123/comments` 返回 HTTP 200 且调用方 wxId 非模拟用户
- **THEN** 对应 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: Simulated user API call skipped

- **WHEN** 模拟用户 wxId=1001 经 gateway 调用 `GET /ucg/app/api/feed/recommend` 返回 HTTP 200
- **THEN** 系统 MUST NOT 增加全局、wxId=1001 或交叉维度计数

#### Scenario: 登录 API 仍计入 for real users

- **WHEN** 经 gateway-app 的 `POST /device/app/api/apple_login` 返回 HTTP 200 且为新真实用户
- **THEN** 对应 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 2xx 成功请求计入统计

- **WHEN** 真实用户经 gateway-app 的 `GET /ucg/app/api/feed/recommend` 返回 HTTP 200
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

`GET /device/admin/api/usage/list` MUST 要求有效 **Admin JWT**（`Authorization: Bearer`，`aud=gateway-admin`）。查询参数 `days` 默认 `7`；`days=0` 表示聚合 TTL 内全部日桶。响应 MUST 包含 `list` 数组，每项至少含 `apiKey`、`summary`、`count`（窗口内合计）、`lastAt`（Unix 秒）。查询参数 `sortBy` 默认 `count`；`sortBy=lastAt` 时列表 SHALL 按 `lastAt` 降序，否则 SHALL 按 `count` 降序。当 Redis 日桶 Hash 中存在 field 时，读路径 MUST 经 `cachekit.HashGetAll` 正确解析 GoFrame Redis 返回值（含 `HGETALL` 经 adapter 转为 flat `[]string` 的情形）并 SHALL NOT 因类型解析失败返回空列表。

#### Scenario: Redis 有数据时列表非空

- **WHEN** Redis 键 `gw:usage:d:{today}:g` 的 Hash 含至少一个 apiKey field，且管理员携带有效 Admin JWT 请求 `days=7`
- **THEN** 响应 `list` SHALL 包含对应 apiKey 且 `count > 0`

#### Scenario: GoFrame HGETALL flat []string 可读

- **WHEN** `cachekit.HashGetAll` 底层收到 GoFrame adapter 表示为 flat `[]string`（非 map）的 `HGETALL` 结果
- **THEN** `ListAPIs` 等读路径 SHALL 仍正确聚合 field 与计数，SHALL NOT 返回空 `list`

#### Scenario: 默认近 7 天

- **WHEN** 管理员携带有效 Admin JWT 请求列表且未传 `days`
- **THEN** 每项 `count` SHALL 为最近 7 个自然日 INCR 之和

#### Scenario: 无效或缺失 token

- **WHEN** 请求未携带有效 Admin JWT
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回统计数据

### Requirement: Admin usage detail API SHALL list wxId callers for one API

`GET /device/admin/api/usage/detail` MUST 要求有效 Admin JWT。查询参数 MUST 包含 `apiKey`；可选 `days`（默认 7）、`sortBy`（`count` 或 `lastAt`）。响应 `list` 每项 SHALL 至少包含 `wxId`、`nickname`、`count`、`lastAt`。仅 `wxId > 0` 的调用 SHALL 出现在列表中。

#### Scenario: 按 API 下钻 wxId

- **WHEN** 管理员携带 Admin JWT 查询 `apiKey=GET /ucg/app/api/feed/recommend` 且 `days=7`
- **THEN** 响应 SHALL 列出窗口内调用过该 API 的 wxId、展示昵称、次数与最近调用时间

### Requirement: Admin usage user API SHALL list APIs called by one wxId

`GET /device/admin/api/usage/user` MUST 要求有效 Admin JWT。查询参数 MUST 包含 `wxId`（正整数）；可选 `days`（默认 7）。响应 `list` 每项 SHALL 至少包含 `apiKey`、`summary`、`count`、`lastAt`。

#### Scenario: 按 wxId 查看 API 分布

- **WHEN** 管理员携带 Admin JWT 查询 `wxId=1001` 且 `days=7`
- **THEN** 响应 SHALL 列出该用户在窗口内成功调用过的 API、中文说明、次数与最近调用时间

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

html` MUST 通过 `/device/admin/api-usage-stats` 提供（**仅 App 网关**）。 页面 MUST 使用 Hub Admin JWT（`admin-common. js`）调用读 API，MUST NOT 使用 `X-Admin-Password`。 页面 SHALL 提供「按 API」「按用户」两个视图； 页脚或说明区 SHALL 注明：用户维度仅统计 `wxId>0` 的已登录账号； 页面 SHALL 提供排序选择，默认按调用次数降序，可选最近调用降序。 「按用户」视图 MUST 调用 `GET /device/admin/api/usage/wx-list`（MUST NOT 直接调用 `wx/list`）。 「按用户」wx 表格 MUST 展示列：wxId、UCG 昵称、注册时间（`createdAt=0` 显示「—」）、设备号、unionid、平台、用户名账号。 API 下钻用户表 MUST 展示列：wxId、UCG 昵称、次数、最近调用。

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

### Requirement: device admin SHALL link to API usage stats from device record card

`resource/public/admin.html` 的**设备记录**卡片头部 `card-actions` MUST 包含指向 `/device/admin/api-usage-stats` 的链接，文案为「功能使用统计」（或等价中文）。链接在管理员登录成功后 SHALL 可见（与问答库/反馈「展开更多」一致的显示时机）。

#### Scenario: 设备记录区入口

- **WHEN** 管理员登录设备管理页并进入主界面
- **THEN** 设备记录卡片 `card-actions` SHALL 显示「功能使用统计」链接

### Requirement: 新增 App HTTP 接口 MUST 经负责人确认是否计入 usage 统计

当 OpenSpec 变更或实现工作 **新增** 经 gateway-app 对外的 App HTTP 路由（`api/v1` 的 `g.Meta` 或 gateway-app `BindHandler`，不含已结构性 skip 的 internal/admin/static/WS）时，执行方（含 AI）**MUST 向产品负责人询问**该接口是否计入 App API 使用统计；负责人未明确答复前 **MUST NOT** 擅自将其加入或移出 `maintenance_skip.go` denylist。proposal 或 tasks **SHALL** 记录确认结论。

#### Scenario: 新增业务 API 且负责人要求统计

- **WHEN** 变更新增 `POST /ucg/app/api/foo` 且负责人确认需要统计
- **THEN** 实现 SHALL 在 `api/v1` 登记路由且 SHALL NOT 写入 `maintenance_skip.go`

#### Scenario: 新增维护型 API 且负责人要求不统计

- **WHEN** 变更新增维护型接口且负责人确认不统计
- **THEN** 实现 SHALL 在 `maintenance_skip.go` 增加对应排除规则并在 proposal/tasks 中说明

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

---

### Requirement: App 网关按主机白名单回显 CORS Origin

`gateway-app-server` 对浏览器跨域请求 SHALL 在响应中包含 CORS 头。当且仅当请求头 `Origin` 解析成功且其主机（不含端口比较 IP 字面量，含端口时取 hostname）等于 `192.168.0.131` 或 `120.55.50.105`，且 scheme 为 `http` 或 `https` 时，SHALL 将 `Access-Control-Allow-Origin` 设为该 `Origin` 的完整原始值（回显），从而允许该主机上任意端口的 Web 来源。

#### Scenario: 匹配内网 IP 任意端口

- **WHEN** 请求包含 `Origin: http://192.168.0.131:5173` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应包含 `Access-Control-Allow-Origin: http://192.168.0.131:5173`

#### Scenario: 匹配公网 IP 任意端口

- **WHEN** 请求包含 `Origin: https://120.55.50.105:8443` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应包含 `Access-Control-Allow-Origin: https://120.55.50.105:8443`

#### Scenario: 非白名单主机不回显

- **WHEN** 请求包含 `Origin: https://evil.example` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应 SHALL NOT 设置 `Access-Control-Allow-Origin`（或不得回显该 Origin）

### Requirement: CORS 方法与请求头

`gateway-app-server` SHALL 在 CORS 响应中声明允许方法包含 `GET`、`POST`、`OPTIONS`，并 SHALL 在 `Access-Control-Allow-Headers`（或对预检的等价响应）中允许 `Content-Type` 与 `Authorization`，以满足常见 JSON 与 Bearer 联调。

#### Scenario: 预检请求获得方法与头

- **WHEN** 浏览器发送 `OPTIONS` 预检，且 `Origin` 通过主机白名单校验，且带有 `Access-Control-Request-Method: POST` 与 `Access-Control-Request-Headers: content-type, authorization`
- **THEN** 响应状态码为成功（2xx），且包含允许上述方法与头的 CORS 响应头（具体头名大小写遵循实现，语义须满足浏览器识别）

### Requirement: 预检不破坏既有鉴权豁免

对 `OPTIONS` 请求的 Bearer 豁免行为 SHALL 保持与变更前一致：预检请求 MUST NOT 因缺少 Bearer 被拒绝（例如 401）。

#### Scenario: OPTIONS 无 Bearer 仍成功预检

- **WHEN** `OPTIONS` 请求指向需鉴权的 API 路径，且无 `Authorization` 头，但 `Origin` 通过白名单
- **THEN** 响应 SHALL NOT 仅因缺少 Bearer 而返回 401（允许返回 204 或其它 2xx 并完成 CORS 头）

---

### Requirement: 反向代理响应在允许来源下须带齐 CORS 头

对 `gateway-app-server` 上经 `httputil.ReverseProxy` 转发至下游的 **`/device/history/api/*`** 请求：当请求头 `Origin` 经 `ReflectGatewayAppCORSOrigin` 判定为允许（`ok == true`）时，返回给客户端的最终响应 **MUST** 包含 `Access-Control-Allow-Origin`（值为该 Origin 的回显）、`Access-Control-Allow-Methods`、`Access-Control-Allow-Headers`、`Access-Control-Max-Age`，其语义 **MUST** 与同一进程内直连 `gateway_app_cors` 中间件对同源校验通过请求所写入的头一致（若实现抽取为共享函数，则以该函数为准）。

#### Scenario: 带 Authorization 的 GET 列表在代理命中时可通过浏览器 CORS

- **WHEN** 客户端为浏览器且发送 `GET /device/history/api/list?...`，带 `Origin: http://localhost:58912`（或任一当前白名单允许的 Origin），带 `Authorization: Bearer <token>`，且该请求在网关内 **命中** history 反向代理并成功自下游取得 2xx 或业务约定的 HTTP 状态
- **THEN** 最终响应 **MUST** 包含 `Access-Control-Allow-Origin` 且值等于请求 `Origin`（回显），并 **MUST** 包含与直连 API 一致的 `Access-Control-Allow-Methods` 与 `Access-Control-Allow-Headers`（至少涵盖 `GET, POST, OPTIONS` 与 `Content-Type, Authorization` 语义）

#### Scenario: 不允许的 Origin 不在代理响应中伪造 CORS

- **WHEN** 请求命中 history 反向代理且 `Origin` 未通过 `ReflectGatewayAppCORSOrigin`
- **THEN** 网关 **MUST NOT** 为通过该策略而添加 `Access-Control-Allow-Origin`（避免对非白名单来源误放行）

### Requirement: voice/device 代理与 history 行为一致（若共用构建函数）

若 `voice`、`device` 领域 HTTP 代理与 history 共用同一 `buildReverseProxy` 或同一套 CORS 注入扩展点，则对上述代理路径在相同 Origin 规则下 **MUST** 适用与 history 相同的 CORS 注入语义，除非规格或设计文档显式排除某路径。

#### Scenario: 共用构建函数时代理路径一致

- **WHEN** 实现将 CORS 注入接在共用的 ReverseProxy 构建路径上，且请求命中该代理
- **THEN** 允许来源下的 CORS 响应头行为 **MUST** 与 history 场景等效，避免出现「仅 app 直连有 CORS、代理无 CORS」的分裂

---

### Requirement: 设备号业务登录响应须含非空 deviceNo

对 **POST `/device/app/api/user/device_login`**（device-service，业务成功 `code=0`），响应 `data` **MUST** 包含 JSON 字段 **`deviceNo`**，且为 **trim 后非空字符串**，表示本次完成登录校验的设备号（与请求入参在规范化后一致，或与库内绑定到该会话的权威 `device_no` 一致）。

#### Scenario: 纯设备会话（无 wx 行绑定）

- **WHEN** 客户端提交已注册设备号且 wx 表无对应行，业务校验成功
- **THEN** 响应 `data.deviceNo` **MUST** 为非空字符串，且与本次登录所用设备号一致

### Requirement: 网关聚合设备号登录响应须含非空 deviceNo

对 **POST `/device/app/api/device_login`**（gateway-app-server，业务成功 `code=0`），响应 `data` **MUST** 包含 **`deviceNo`**，且为 **trim 后非空字符串**。若下游 device 返回的 `data.deviceNo` 为空或缺失，网关 **SHALL** 使用本次请求体中的 `deviceNo`（trim 后）作为回包与 JWT 签发所用设备号；若兜底后仍为空，**MUST** 拒绝成功语义（沿用现有参数/内部错误路径）。

#### Scenario: 下游 data 缺 deviceNo 时网关兜底

- **WHEN** device `device_login` 返回 `code=0` 但 `data` 中无 `deviceNo` 或值为空白，且请求体 `deviceNo` 经 trim 后非空
- **THEN** 网关返回的聚合响应 `data.deviceNo` **MUST** 等于该 trim 后的请求 `deviceNo`，且签发的 access/refresh 所绑定的设备号与该值一致

#### Scenario: 请求与下游均无可用设备号

- **WHEN** 请求体 `deviceNo` trim 后为空，或业务失败
- **THEN** 网关 **MUST NOT** 返回 `code=0` 且带非空 `deviceNo` 的成功形态（保持现有错误语义）

---

### Requirement: access JWT SHALL 同时携带 wx 主键与 device_no 声明

gateway-app-server 签发的 **access_token（JWT）SHALL** 使用标准 **`sub`** claim 承载 **`wx` 表主键 id**（十进制字符串，与现网 refresh 语义一致）；并 **SHALL** 包含 **`device_no` 私有声明**（与 `ai_voice_device.wx.device_no` 语义一致）。当用户尚未绑定设备时，`device_no` 声明 **MAY** 为空或省略，其实现策略 **MUST** 在实现与评审中保持唯一且文档化。

#### Scenario: 已绑定设备用户登录后拿到 access

- **WHEN** device 在 `POST /device/app/api/user/login` 返回的 `deviceNo` 非空且网关签发 access
- **THEN** JWT **MUST** 可被解析为包含非空的 **`device_no` 声明** 与与 `wxId` 一致的 **`sub`**

### Requirement: Bearer 中间件 SHALL 注入 Wx-Id 与可选 Device-No 头且不再拉取 unionid

gateway-app-server 对非白名单 HTTP 请求在校验 access JWT 成功后，**SHALL** 设置 **`X-Internal-Wx-Id`** 为 **`sub`** 所表示的整数 wx 主键（字符串形式与头规范在实现中固定）；**SHALL** 在 **`device_no` 声明非空** 时设置 **`X-Internal-Device-No`** 为该值。**MUST NOT** 为完成上述注入而调用 device-service 的 **`GET /device/app/api/user/internal/by-id`**（即 **禁止** 将「id→unionid」作为网关热路径依赖）。

#### Scenario: 受保护 HTTP 请求鉴权通过

- **WHEN** 客户端携带合法 Bearer access JWT
- **THEN** 发往 device/history/voice 等下游的代理请求 **MUST** 携带 **`X-Internal-Wx-Id`**；且当 JWT 含非空 **`device_no` 声明** 时 **MUST** 携带 **`X-Internal-Device-No`**

#### Scenario: 对外 HTTP 契约保持不变

- **WHEN** App 调用 `POST /device/app/api/login` 或 `POST /device/app/api/token/refresh`
- **THEN** 请求与响应 JSON **MUST** 保持与变更前一致的字段名与客户端可见语义（客户端 **MUST NOT** 需要解析 JWT 载荷即可集成）

### Requirement: device-service 用户域 SHALL 以 X-Internal-Wx-Id 识别 wx 行

device-service 对 **`POST /device/app/api/user/bindwx`**、**`POST /device/app/api/user/auto_save`**、**`GET /device/app/api/user/detail`** 等依赖「当前登录 wx」的接口，**SHALL** 从请求头 **`X-Internal-Wx-Id`** 读取 wx 主键并定位 `wx` 行；**MUST NOT** 将 **`X-Internal-Wx-Union-Id`** 作为网关受信任路径的必需依赖（若保留兼容，**MUST** 在部署文档中声明过渡期与移除时间）。

#### Scenario: bindwx 成功

- **WHEN** 请求携带合法 **`X-Internal-Wx-Id`** 且 body 中 `deviceNo` 合法
- **THEN** 系统 SHALL 完成绑定并返回成功语义

### Requirement: 历史 WebSocket SHALL 使用 JWT device_no 声明与首帧 device_no 校验

gateway-app-server 的历史 WebSocket 在首帧 `auth` 后，**SHALL** 校验 access JWT 的 **`device_no` 声明** 与首帧 JSON 中的 **`device_no`（或 `deviceNo`，以实现为准且单一）** 一致（在声明非空时）；**MUST NOT** 依赖「unionid → detail 拉 device_no」链完成该校验。

#### Scenario: 认证成功

- **WHEN** JWT 有效且 **`device_no` 声明** 与首帧设备号一致
- **THEN** 连接 SHALL 注册到对应 `device_no` 的推送组

#### Scenario: 认证失败

- **WHEN** JWT 有效但设备号不一致或声明缺失导致无法满足校验策略
- **THEN** 服务端 SHALL 拒绝订阅并 SHALL NOT 将连接加入推送组

### Requirement: refresh 重新签发的 access SHALL 同步 device_no 声明

gateway-app-server 在处理 **`POST /device/app/api/token/refresh`** 时，**SHALL** 在签发新 access JWT 时写入 **与当前 wx 会话权威一致的 `device_no` 声明**（以 device-service 返回或网关侧明确规则为准），以避免换绑后长期持有错误 `device_no` claim 的策略 **MUST** 在 design 的 D5 中落地为单一实现。

#### Scenario: 刷新成功

- **WHEN** refresh_token 有效且旋转策略允许签发新 access
- **THEN** 新 access JWT **MUST** 包含更新后的 **`device_no` 声明**（若设备域当前已绑定）

---

### Requirement: Gateway-app 根路径承载胖宝官网

`gateway-app-server` SHALL 在根路径 `/` 返回“胖宝”官网 HTML，而不是当前纯文本“智能语音 App 网关”。该官网路由变更 MUST 仅作用于 `gateway-app-server` 进程，MUST NOT 改变主网关或其他微服务进程的根路径行为。

#### Scenario: 访问 gateway-app 根路径
- **WHEN** 浏览器对 `gateway-app-server` 发起 `GET /`
- **THEN** 系统 SHALL 返回官网 HTML 页面，页面标题与主视觉 SHALL 展示品牌名“胖宝”

#### Scenario: 官网替换不扩散到主网关
- **WHEN** 本次变更部署完成
- **THEN** 系统 SHALL 仅修改 `gateway-app-server` 的根路径处理逻辑，主网关进程的路由与代理行为 MUST 保持不变

### Requirement: 官网展示母婴喂养定位与事件卡片

官网页面 SHALL 以玻璃拟态风格展示品牌定位文案，并 SHALL 展示从数据库权威链路读取的事件列表。每个事件项 MUST 至少包含事件名与事件 logo；若 logo 为 path-only 资源，前端或聚合接口 MUST 能将其解析为当前站点可访问的同源地址。除喂养事件卡片外，页面 SHALL 包含独立的产品能力区块（见「官网展示三大产品能力柱」），首屏定位 MUST 覆盖 0–18 月新手妈妈的喂养、轻社交与 AI 价值。

#### Scenario: 官网首屏展示品牌定位
- **WHEN** 用户打开官网首页
- **THEN** 页面 SHALL 明确表达面向 0–18 月新手妈妈（或等价阶段化表述），以及「喂养记录更便捷、同月龄轻社交、AI 智能陪伴」等核心信息，而不仅限于单一喂养工具定位

#### Scenario: 官网展示事件 logo 与事件名
- **WHEN** 官网聚合到至少一条事件数据
- **THEN** 页面 SHALL 为每条事件渲染事件 logo 与事件名，且 logo 地址 MUST 可被当前官网域名直接访问

#### Scenario: 喂养事件区说明对齐小月龄场景
- **WHEN** 用户查看 `#events` 喂养事件区块
- **THEN** 区块说明文案 SHOULD 提及小月龄高频照护场景（如吃奶、睡觉、换尿布）及语音记录等等价表述

### Requirement: 官网提供匿名只读聚合数据接口

系统 SHALL 提供一个适用于官网匿名访问的只读聚合接口，由 `gateway-app-server` 统一返回官网所需的事件展示数据、Android 下载信息与 iOS 下载说明。该接口 MUST 通过服务契约或本进程已有能力获取数据，MUST NOT 让前端直接调用受保护业务接口或跨服务直连数据库。接口返回的 `heroTitle`、`heroSubtitle`、`serviceSummary` MUST 与官网 0–18 月新手妈妈及「喂养 + 轻社交 + AI」三角定位一致。

#### Scenario: 匿名读取官网数据
- **WHEN** 未登录用户请求官网聚合接口
- **THEN** 系统 SHALL 返回成功响应，其中包含事件列表、Android 下载展示信息、iOS 下载说明，以及已更新语义的 Hero 文案字段

#### Scenario: 官网数据来源遵守服务边界
- **WHEN** `gateway-app-server` 组装官网响应
- **THEN** 系统 MUST 通过现有服务契约读取事件权威数据，并复用本进程版本信息读取能力，MUST NOT 新增跨服务直连他域库表行为

### Requirement: 官网展示 Android 下载二维码与 iOS 指引

官网 SHALL 提供独立的应用下载区块。Android 下载区 MUST 基于数据库中的最新下载链接生成二维码并展示可点击下载入口：当最新行为 OSS objectKey 时，聚合接口 MUST 返回 CDN 绝对 URL；当最新行为历史 path-only 时，MUST 仍拼出 gateway 可访问的绝对地址。iOS 下载区 MUST 提示用户前往 App Store 搜索“胖宝”下载。

#### Scenario: Android 存在可下载版本（OSS）

- **WHEN** 版本表存在最新 Android 下载记录且 `download_url` 为 `apk/` objectKey
- **THEN** 官网聚合接口 SHALL 返回 CDN 绝对 Android 下载地址，页面 SHALL 生成对应二维码并展示下载入口

#### Scenario: Android 存在可下载版本（历史 path）

- **WHEN** 版本表存在最新 Android 下载记录且 `download_url` 为 `/device/app/apk/` 路径且本地文件可访问
- **THEN** 官网聚合接口 SHALL 返回官网可直接使用的绝对下载地址（基于 `publicBaseUrl` 或请求 Host），页面 SHALL 生成对应二维码并展示下载入口

#### Scenario: Android 暂无可下载版本

- **WHEN** 版本表没有可用的 Android 下载记录
- **THEN** 页面 SHALL 不展示失效二维码，并 SHALL 展示明确的“Android 下载暂未开放”或等价提示

#### Scenario: iOS 下载说明固定展示

- **WHEN** 用户查看官网下载区
- **THEN** 页面 SHALL 展示“前往 App Store 搜索‘胖宝’下载”的文案，而不要求数据库提供 iOS 下载链接

### Requirement: 官网展示三大产品能力柱

官网页面 SHALL 在 Hero 与喂养事件区块之间展示独立的「产品能力」区块（锚点 `#capabilities`），以玻璃拟态卡片呈现三大能力柱：**智能喂养记录**、**同月龄轻社交**、**AI 智能陪伴**。每柱 MUST 至少包含标题与 2 条场景化说明；移动端 MUST 可正常阅读（单列或等价布局）。

#### Scenario: 用户浏览产品能力区块
- **WHEN** 用户打开官网首页并滚动至 `#capabilities`
- **THEN** 页面 SHALL 展示上述三柱，且文案 MUST 涵盖语音记事件、同阶段妈妈社区（发帖/关注/私信）、AI 语音助手与智能喂养问答等等价能力描述

#### Scenario: 导航可跳转至产品能力
- **WHEN** 用户点击导航中的「产品能力」或等价链接
- **THEN** 页面 SHALL 滚动至 `#capabilities` 锚点

### Requirement: 官网面向 0-18 月新手妈妈定位

官网首屏（Hero）SHALL 明确面向 **0–18 月新手妈妈** 或等价表述（如 kicker 含「0–18 月」）。Hero MUST 同时传达「喂养记录」「同月龄连接」「AI 问答」三类价值，而非仅喂养工具定位。

#### Scenario: 首屏展示阶段化定位
- **WHEN** 用户打开官网首页
- **THEN** Hero 区域 SHALL 包含 0–18 月或新手妈妈阶段化定位文案，且副标题或 metric 卡片 MUST 提及社交与 AI 能力

#### Scenario: SEO meta 覆盖核心关键词
- **WHEN** 爬虫或浏览器读取页面 `<head>`
- **THEN** `<title>` 与 `<meta name="description">` MUST 包含喂养记录相关表述，并 SHOULD 包含 0–18 月、AI、同月龄或社区等关键词中的至少两项

### Requirement: 官网 AI 能力合规表述

官网展示 AI 能力时 SHALL 使用「智能喂养问答」或等价非医疗措辞，MUST NOT 单独使用「诊疗」作为官网主宣传语。页面 MUST 包含简短免责声明，说明 AI 回答仅供参考、不能替代医生诊断。

#### Scenario: AI 能力柱不含诊疗主宣传
- **WHEN** 用户阅读 `#capabilities` 中 AI 相关文案
- **THEN** 页面 SHALL NOT 以「胖宝诊疗」作为唯一或主标题对外展示，SHALL 使用智能问答/AI 助手等表述

#### Scenario: 展示 AI 免责声明
- **WHEN** 用户浏览含 AI 描述的区块或页脚
- **THEN** 页面 SHALL 展示等价于「AI 回答仅供参考，不能替代医生诊断」的免责声明

### Requirement: 官网 Footer 法律页链接

官网页脚 SHALL 提供指向隐私政策与用户协议的可见链接，分别对应 `/privacy-policy.html` 与 `/user-agreement.html`（或 gateway-app 注册的等价路由）。

#### Scenario: 页脚可访问法律文档
- **WHEN** 用户查看官网页脚
- **THEN** 页面 SHALL 展示可点击的「隐私政策」与「用户协议」链接，且链接 MUST 在同源 gateway-app 域名下可访问

---

### Requirement: APK download_url SHALL 仅存应用内路径

gateway-app-server 在 **历史兼容** 场景下仍可识别以 `/` 开头的 path-only `download_url`（格式 `/device/app/apk/<filename>.apk`）。**新上传写库**时，`app_version.download_url` MUST 存 OSS objectKey（前缀 `apk/`，无 scheme/域名）；MUST NOT 将 `publicBaseUrl` 或 CDN 域名写入该列（新写入）。上传接口与对外读接口的 JSON `downloadUrl` 对 OSS 行 MUST 为 CDN 绝对 URL，对 path-only 历史行 MUST 保持可下载语义（check 可为 path；site 拼绝对 URL）。

#### Scenario: 上传 APK 写库为 objectKey

- **WHEN** 管理员成功上传 APK 并完成数据库 Insert
- **THEN** `download_url` MUST 为以 `apk/` 开头的 objectKey（无 `http://`/`https://`）
- **AND** 上传接口 JSON 响应中的 `downloadUrl` MUST 为 CDN 绝对 URL

#### Scenario: 上传不再因缺少 publicBaseUrl 拒绝写库

- **WHEN** `gatewayApp.publicBaseUrl` 未配置
- **AND** 上传文件与其它表单字段合法且 OSS 上传成功
- **THEN** 服务端 MUST 仍能完成写库（objectKey 存库）

### Requirement: 版本检查接口 SHALL 返回 path 型 downloadUrl

`GET /device/app/api/version/check` 响应中的 `downloadUrl`：当最新行为 **OSS objectKey**（或库内已是 CDN 绝对 URL）时，MUST 返回 **CDN 绝对 URL**（以 `https://` 开头），客户端 MUST 可直接使用而不再拼接 gateway 基址；当最新行为 **历史 path-only**（`/device/app/apk/...`）时，MUST 仍返回应用内路径（供客户端拼接网关基址），与变更前兼容。MUST NOT 对 CDN/OSS 值使用「剥 host 只留 path」后再当作应用内 APK 路径返回。

#### Scenario: 版本检查对 OSS 行返回 CDN

- **WHEN** 版本表存在最新行且 `download_url` 为 `apk/` objectKey
- **THEN** `downloadUrl` MUST 以 `https://` 开头且指向 CDN
- **AND** MUST NOT 形如仅 `/device/app/apk/xxx.apk` 的 path-only

#### Scenario: 版本检查对历史 path 行仍返回路径

- **WHEN** 版本表最新行 `download_url` 为 `/device/app/apk/xxx.apk`
- **THEN** `downloadUrl` MUST 形如 `/device/app/apk/xxx.apk`
- **AND** MUST NOT 以 `http://` 或 `https://` 开头

### Requirement: gateway-app SHALL 代理事件 logo 静态路径

gateway-app-server MUST 注册 `GET /ai_talk_images/*`，将请求转发至 device-service 同路径（或等价静态源），使客户端可通过 gateway-app 端口（如 `:9702`）访问 `https://<host>:9702/ai_talk_images/...` 而无需直连 device 端口。

#### Scenario: 经 gateway-app 访问事件 logo

- **WHEN** 客户端请求 gateway-app 的 `GET /ai_talk_images/event_1.png`
- **AND** device-service 上对应文件存在
- **THEN** gateway-app MUST 返回成功图片响应（经反代或共享存储）

### Requirement: APK 下载路径契约保持不变

既有 `GET /device/app/apk/*filename` 下载处理器 MUST 继续从 `apkStorageDir` 提供文件，供 **历史 path-only** `download_url` 与邀请群二维码等本地资源使用。新上传 APK 的权威下载 MUST 走 CDN，不依赖本路由作为唯一分发通道。

#### Scenario: path 与下载路由一致

- **WHEN** `download_url` 为 `/device/app/apk/foo.apk`
- **THEN** 对 gateway-app 发起 `GET /device/app/apk/foo.apk` MUST 可下载该文件

---

### Requirement: App 网关进程独立运行

系统 SHALL 提供名为 gateway-app-server 的独立 HTTP 服务进程，具备与现有 gateway 相当的静态资源与领域反向代理能力，并额外承载 App 鉴权、令牌、版本检查、历史 WebSocket，以及 **UCG HTTP 反向代理**（`/ucg/app/api/*` → ucg-service）、**UCG 聊天 WebSocket 升级代理**（`/ucg/app/ws/chat` → ucg-service `/ws/chat`），以及 **voice HTTP 反向代理**（`/voice/app/api/*`、`/voice/admin/api/*` → voice-service）。App 对外 UCG 与 voice App/Admin 流量 MUST 经本进程暴露，与现有 App API 同域。

#### Scenario: 进程启动与配置隔离

- **WHEN** 使用 gateway-app-server 专用配置文件启动进程
- **THEN** 服务 SHALL 仅加载该进程所需的数据库分组（含 ai_voice_app）与下游 URL 配置（含 `UCG_SERVICE_BASE_URL`、`UCG_WS_PROXY_URL`、**`VOICE_API_PROXY_URL`**），且 SHALL NOT 将 voiceChat 等业务配置错误合并到错误进程的权威配置源中（遵循仓库既有配置边界约定）

#### Scenario: UCG HTTP 代理可用

- **WHEN** 配置 `UCG_SERVICE_BASE_URL` 且 ucg-service 健康
- **THEN** 对 `/ucg/app/api/*` 的请求 SHALL 经 Bearer 鉴权与头注入后转发至 ucg-service

#### Scenario: voice HTTP 代理可用

- **WHEN** 配置 `VOICE_API_PROXY_URL` 且 voice-service 健康
- **THEN** 对 `/voice/app/api/*` 的请求 SHALL 经 Bearer 鉴权与头注入后转发至 voice-service
- **AND** 对 `/voice/admin/api/*` 的请求 SHALL 经 Admin JWT 校验与 `VOICE_ADMIN_PASSWORD` 注入后转发至 voice-service

#### Scenario: UCG 聊天 WS 升级代理可用

- **WHEN** 配置 `UCG_WS_ROUTE_MODE=proxy` 且 `UCG_WS_PROXY_URL` 指向可达的 ucg-service `/ws/chat`
- **AND** 客户端对 `/ucg/app/ws/chat` 发起 WebSocket Upgrade
- **THEN** gateway-app SHALL 将握手与后续双向帧透传至 ucg-service，行为与 `ws_route_proxy.go` voice WS 透传一致

### Requirement: Bearer 鉴权与内部头注入

系统 SHALL 对除白名单外的受保护 HTTP 路径校验 `Authorization: Bearer <access_token>`，其中 `access_token` MUST 为合法 JWT。系统 SHALL 在校验签名与过期时间后，从 `sub` 解析 `wx.id`（允许 `0` 表示纯设备会话）并向下游注入 **`X-Internal-Wx-Id`**；当 access 含 `device_no` 声明时，系统 SHALL 同步注入 **`X-Internal-Device-No`**。

#### Scenario: 鉴权通过并注入头
- **WHEN** Bearer 为合法未过期 JWT，且 `sub` 与 `device_no` 组合满足会话规则
- **THEN** 网关 SHALL 设置 `X-Internal-Wx-Id`，并在有值时设置 `X-Internal-Device-No`

#### Scenario: 鉴权失败
- **WHEN** Bearer 缺失、签名错误、已过期或会话字段非法
- **THEN** 网关 SHALL 返回未授权错误，且 SHALL NOT 注入内部头

### Requirement: 登录与令牌仅由 gateway-app 签发

系统 SHALL 在 gateway-app-server 提供并维护两类聚合登录：
1. `POST /device/app/api/login`（微信聚合登录，转发 device 微信业务登录）
2. 用户名聚合登录接口（路径位于 `/device/app/api/` 前缀下，转发 device 用户名登录业务接口）

两类聚合登录在成功后 SHALL 统一由 gateway 签发 access/refresh；access MUST 为 JWT，`sub` MUST 等于目标 `wx.id`；refresh SHALL 为不透明随机串并绑定 Redis 会话。

#### Scenario: 用户名聚合登录成功
- **WHEN** 客户端调用用户名聚合登录且 device 返回有效 `wxId`
- **THEN** 网关 SHALL 返回 accessToken 与 refreshToken，且 access `sub` SHALL 等于该 `wxId`

#### Scenario: 微信聚合登录保持兼容
- **WHEN** 客户端调用既有微信聚合登录
- **THEN** 网关 SHALL 按现有语义返回 token 与业务字段，不因新增用户名能力破坏兼容性

### Requirement: 刷新令牌接口

系统 SHALL 在 gateway-app-server 提供刷新 access 的 HTTP 接口（路径位于 `/device/app/api/` 前缀下），使用 Redis 校验 refresh 后签发新的 **JWT** 形态 access_token（`sub`/`iat`/`exp` 规则与登录接口一致），并可按产品策略旋转 refresh_token。

#### Scenario: 刷新成功

- **WHEN** 客户端提交有效 refresh_token
- **THEN** 系统 SHALL 返回新的 access_token 且该 token SHALL 为合法 JWT，且旧 refresh 的处理策略（保留至过期或立即失效）SHALL 与设计文档一致并在实现中单一实现

### Requirement: 版本检查 API

系统 SHALL 在 gateway-app-server 提供 `GET /device/app/api/version/check`，从查询参数读取 `currentVersion`，读取 ai_voice_app.version 表（或经缓存的等价数据）并返回 needUpdate、latestVersion、releaseNotes、downloadUrl、forceUpdate。

#### Scenario: 返回版本信息

- **WHEN** 客户端携带合法 currentVersion 调用版本检查接口
- **THEN** 响应 SHALL 包含布尔 needUpdate 及 latestVersion、releaseNotes、downloadUrl、forceUpdate 字段，且 MAY 使用 Redis 缓存版本行以降低数据库压力

### Requirement: 历史 WebSocket 与首帧认证

系统 SHALL 在 gateway-app-server 提供 WebSocket 端点；连接建立后首条文本帧 MUST 为 JSON，包含 `type` 为 `auth`、`access_token`（snake_case 键名，值为 **JWT 字符串**）与 `device_no`；服务端 MUST 按与 HTTP Bearer 相同的规则校验 JWT 后，再校验 `sub` 对应 wx 身份与该 device_no 的绑定关系，通过后才将连接注册到按 device_no 分组的推送集合。

#### Scenario: 认证成功并订阅

- **WHEN** 客户端发送合法 auth 帧且 access_token 为有效 JWT、device_no 与该 token 身份匹配
- **THEN** 连接 SHALL 保持打开并能够接收后续由 Redis 通知触发的历史变更消息

#### Scenario: 认证失败

- **WHEN** auth 帧缺失、字段不合法或 device_no 与身份不匹配
- **THEN** 服务端 SHALL 拒绝订阅（关闭连接或发送错误文本帧）且 SHALL NOT 将该连接加入任何 device_no 推送组

### Requirement: Redis Pub/Sub 消费与下行

系统 SHALL 在 gateway-app-server 进程内维护对约定 Redis channel 的订阅；当收到 history-service 发布的消息时，SHALL 向所有已认证且匹配 `device_no` 的 WebSocket 连接推送 JSON 业务消息。

#### Scenario: 收到发布并推送

- **WHEN** Redis 收到一条包含已知 device_no 与历史载荷的合法通知
- **THEN** 网关 SHALL 向该 device_no 下已注册且仍存活的连接广播该消息体

### Requirement: gateway-app-server SHALL expose history SSE endpoint

`gateway-app-server` MUST register `GET /device/app/sse/history` (or the path locked with Flutter) that responds with `Content-Type: text/event-stream` after successful JWT Bearer authentication. The handler MUST NOT perform WebSocket Upgrade on this path. `gateway-app-server` MUST 注册历史 SSE 端点；鉴权成功后 MUST 返回 `text/event-stream`；MUST NOT 在该路径 Upgrade WebSocket。

#### Scenario: 合法 Bearer 打开 SSE

- **WHEN** 客户端携带有效 `Authorization: Bearer` 请求 `GET /device/app/sse/history` 且 `Accept` 含 `text/event-stream`
- **THEN** 响应 MUST 为 HTTP 200 与 `text/event-stream`
- **AND** 连接 MUST 保持打开直至客户端断开或服务端结束流

#### Scenario: 无/无效 Bearer

- **WHEN** 请求缺少有效 Bearer
- **THEN** 响应 MUST 为 HTTP 401（或与 App HTTP 一致的未授权语义）
- **AND** MUST NOT 建立 SSE 流

### Requirement: History SSE SHALL bind device_no like history WebSocket

After JWT validation, the history SSE subscription MUST associate the connection with the device identity: query/header `deviceNo` (lowerCamelCase) MUST match the access JWT `device_no` claim when the claim is non-empty. Mismatch MUST reject the subscription. 历史 SSE MUST 按与历史 WS 等价规则绑定 `device_no`；声明非空时与请求 deviceNo 不一致 MUST 拒绝。

#### Scenario: deviceNo 与 JWT 一致

- **WHEN** JWT `device_no` 为 D 且请求 `deviceNo=D`
- **THEN** 服务端 MUST 接受订阅并将该连接加入 D 的推送集合

#### Scenario: deviceNo 与 JWT 不一致

- **WHEN** JWT `device_no` 为 D 且请求 `deviceNo=E`（E≠D）
- **THEN** 服务端 MUST 拒绝订阅（4xx）
- **AND** MUST NOT 向该连接推送 D 的历史事件

### Requirement: History SSE SHALL fan out the same Redis-sourced events as history WS

When history-service publishes a history mutation to the agreed Redis channel, gateway-app-server MUST push an SSE event to all authenticated history SSE subscribers whose bound `device_no` matches, using JSON data compatible with existing history WebSocket application frames (`action` + `payload`). history-service Redis 推送 MUST 扇出到匹配 device_no 的历史 SSE 订阅者；data JSON MUST 与现网 history WS 业务帧兼容。

#### Scenario: create 推送

- **WHEN** Redis 收到某 `device_no` 的 create 消息且存在该 device 的 SSE 订阅
- **THEN** 网关 MUST 向这些订阅写入至少一条含 create 语义的 SSE data 帧

### Requirement: History SSE MUST NOT count as WebSocket usage and MUST avoid proxy buffering

History SSE requests MUST follow gateway usage exclusions appropriate for long-lived streams (MUST NOT be mis-counted as successful short HTTP if existing rules exclude WS upgrades—document whether SSE is counted). Reverse-proxy responses for this path MUST disable response buffering where applicable so events flush promptly. 历史 SSE 长连接的统计与反代缓冲策略 MUST 明确；适用处 MUST 禁用缓冲以保证及时刷出。

#### Scenario: 事件及时到达

- **WHEN** 网关向 SSE 订阅写入一帧业务 data
- **THEN** 在未缓冲配置下客户端 MUST 能在短时间内收到该帧（联调可观察）

### Requirement: Adding history SSE MUST NOT break existing history WebSocket clients

在提供历史 SSE 的同时，`gateway-app-server` MUST 继续提供 `/device/app/ws/history`，且 MUST 保持现网握手与推送契约（GET Bearer 豁免、首帧 auth/`auth_ok`、ping/pong、Redis 原文下行含 `device_no`+`action`+`payload`）。MUST NOT 从豁免表移除该路径；MUST NOT 为 SSE 包装改写 WS 下行。Adding history SSE MUST NOT change existing history WebSocket client-visible behavior.

#### Scenario: 未升级客户端仍可握手并鉴权

- **WHEN** 未升级 App 对 `/device/app/ws/history` 发起 WebSocket Upgrade 且未携带 HTTP `Authorization`
- **THEN** 握手 MUST 成功（路径仍在 Bearer 豁免）
- **AND** 首帧合法 `auth` 后 MUST 收到 `auth_ok`（或与现网等价的成功语义）

#### Scenario: 未升级客户端仍可收到增量

- **WHEN** 某 `device_no` 的 history WS 已认证订阅，且 Redis 发布该设备的 create/update/delete
- **THEN** 网关 MUST 仍向该 WS 连接写入与现网同形的业务 JSON（含顶层 `device_no`、`action`、`payload`）

#### Scenario: 豁免表不误伤 WS

- **WHEN** 检查 `gatewayAppAuthExemptExactGET`（或等价豁免清单）
- **THEN** `/device/app/ws/history` MUST 仍在清单中
- **AND** `/device/app/sse/history` MUST NOT 在 Bearer 豁免清单中

### Requirement: SSE proxy buffering disable MUST be path-scoped

禁用响应缓冲的反代/网关配置（若有）MUST 限定于历史 SSE 路径（或等价精确 location），MUST NOT 以「全局关闭 buffering」作为历史 SSE 的默认部署方式，以免改变其它 HTTP/WS 入口特性。SSE proxy buffering disable MUST be scoped to the history SSE path.

#### Scenario: 仅 SSE location 禁缓冲

- **WHEN** 运维按部署清单配置反代
- **THEN** 禁缓冲规则 MUST 仅匹配历史 SSE 路径（或文档明示的等价精确集合）
- **AND** MUST NOT 要求全局 `proxy_buffering off` 作为唯一合规方式

### Requirement: 鉴权白名单

系统 SHALL 将无需 Bearer 的入口纳入白名单，至少包含：微信聚合登录、用户名聚合登录、refresh、公开版本检查（若启用）及 WebSocket 握手路径。

#### Scenario: 无令牌访问用户名登录
- **WHEN** 客户端无 Authorization 头调用用户名聚合登录接口
- **THEN** 请求 SHALL 进入对应处理器且 SHALL NOT 被 Bearer 中间件拦截

---

### Requirement: 版本管理页访问控制

gateway-app-server SHALL 提供「版本管理」相关 UI 与 API；在未通过 **Admin JWT**（`aud=gateway-admin`，由 Hub `POST /device/admin/api/login` 签发）鉴权前，SHALL NOT 暴露 APK 上传与写库能力。系统 MUST NOT 再使用独立版本管理口令、`POST /device/app/api/version/admin/login` 或 `gw_ver_admin` Cookie 会话作为鉴权手段（B1）。

#### Scenario: 无 Admin JWT 拒绝管理操作

- **WHEN** 客户端在未携带有效 Admin JWT 的情况下调用上传或写库接口
- **THEN** 系统 SHALL 返回未授权响应且 SHALL NOT 写入磁盘或数据库

#### Scenario: Admin JWT 校验通过允许操作

- **WHEN** 客户端携带由 Hub 登录签发的有效 Admin JWT
- **THEN** 系统 SHALL 允许后续受保护操作（上传、写库、列表等）

### Requirement: Android APK 落盘路径

系统对 **新上传** 的 Android APK MUST 经 ucg 写入 OSS（objectKey 前缀 `apk/`），MUST NOT 再要求将新包保存至 `/apk/ai_talk/` 作为权威存储。`apkStorageDir`（默认 `/apk/ai_talk/`）MUST 仍可用于服务历史本地 APK 与同目录其它本地资源（如邀请群二维码）；若目录不存在且仍需写入本地资源，MAY 继续 `MkdirAll`。

#### Scenario: 新上传权威存储为 OSS

- **WHEN** 管理员已鉴权且上传合法 APK 且 OSS 可用
- **THEN** 系统 MUST 在 OSS 中存在对应 `apk/` 对象
- **AND** MUST NOT 将「仅本地落盘成功」视为上传成功的充分条件

### Requirement: 下载 URL 与数据库一致性

上传成功后，系统 MUST 将 OSS **objectKey** 写入 `ai_voice_app` 版本表（或 `app_version` 等价表）对应新记录的 **`download_url`** 字段；对外返回的下载链接 MUST 为指向该对象的 **CDN 绝对 URL**。客户端使用该 CDN URL 发起 GET 时 SHALL 能获取 APK 内容（由 CDN/OSS 提供），路径穿越与本地目录外访问防护仍适用于本地 `GET /device/app/apk/*` 处理器。

#### Scenario: 客户端可经 CDN 下载

- **WHEN** 任意客户端使用版本检查或上传响应中的 CDN `downloadUrl` 发起 GET 请求
- **THEN** 响应 SHALL 返回对应 APK 内容（`application/vnd.android.package-archive` 或等价二进制流）

### Requirement: 上传与文件名校验

系统 SHALL 仅接受扩展名为 **`.apk`**（大小写不敏感可统一规范）的上传；SHALL 拒绝路径分隔符、空文件名等非法文件名；SHALL 可配置单文件大小上限并在超限时拒绝。

#### Scenario: 非 APK 扩展名拒绝

- **WHEN** 管理员上传文件扩展名不是 `.apk`
- **THEN** 系统 SHALL 拒绝保存且不更新 `download_url`

### Requirement: 与现有版本检查行为兼容

新增写库记录后，`GET /device/app/api/version/check` 所依据的「最新版本行」语义 SHALL 与现有实现一致（按主键或约定排序取最新一条）。当该行为 OSS 行时，返回的 **`downloadUrl`** MUST 为 CDN 绝对 URL（由 objectKey 映射），MUST NOT 要求与库内 objectKey 字符串字面相等。

#### Scenario: 新插入 OSS 行成为最新发版

- **WHEN** 管理员通过本功能插入一条包含 `latest_version` 与 OSS `download_url`（objectKey）的新版本行且其排序上为最新
- **THEN** 客户端调用版本检查接口时 SHALL 收到该行的 `latestVersion` 与 CDN 形式的 `downloadUrl`

---

### Requirement: 版本管理历史列表

gateway-app-server SHALL 向已通过 **Admin JWT** 鉴权的客户端提供历史发版列表接口，返回 `ai_voice_app.version` 表中的记录，默认按主键 `id` 降序排列。

#### Scenario: 已登录管理员获取列表

- **WHEN** 客户端携带有效 Admin JWT 并请求列表接口且 Hub admin 登录功能已启用
- **THEN** 系统 SHALL 返回 `code=0` 及包含版本记录的列表（含 `id`、`latestVersion`、`releaseDate`、`releaseNotes`、`downloadUrl`、`forceUpdate`、`minVersion`）

#### Scenario: 未登录拒绝列表

- **WHEN** 客户端未携带有效 Admin JWT 即请求列表接口
- **THEN** 系统 SHALL 返回未授权响应且 SHALL NOT 返回版本行数据

#### Scenario: 分页参数生效

- **WHEN** 客户端传入合法的 `limit` 与 `offset`
- **THEN** 系统 SHALL 按 `id` 降序返回不超过 `limit` 条记录（`limit` 不得超过约定上限）

### Requirement: 按 id 查询单条版本

系统 SHALL 支持已携带有效 Admin JWT 的管理员按主键 `id` 查询单条 `version` 记录。

#### Scenario: 存在记录时返回详情

- **WHEN** 客户端携带有效 Admin JWT 请求存在的 `id`
- **THEN** 系统 SHALL 返回该行的完整发版字段

#### Scenario: 不存在记录

- **WHEN** 客户端携带有效 Admin JWT 请求不存在的 `id`
- **THEN** 系统 SHALL 返回未找到响应且 SHALL NOT 返回伪造数据

### Requirement: 更新版本元数据

系统 SHALL 允许已鉴权管理员更新已有 `version` 行的元数据字段，且 SHALL NOT 通过本接口修改 `download_url`。

#### Scenario: 成功更新可编辑字段

- **WHEN** 已鉴权管理员提交有效 `id` 及至少一个允许字段（如 `latestVersion`、`releaseNotes`、`forceUpdate`、`minVersion`、`releaseDate`）
- **THEN** 系统 SHALL 持久化更新并返回成功响应

#### Scenario: 更新后版本检查缓存失效

- **WHEN** 更新操作成功提交
- **THEN** 系统 SHALL 删除或失效用于 `GET /device/app/api/version/check` 的最新行 Redis 缓存键 `gw:app:version:latest`

#### Scenario: 未鉴权拒绝更新

- **WHEN** 未持有有效管理会话的客户端调用更新接口
- **THEN** 系统 SHALL 拒绝写库

### Requirement: 删除版本记录

系统 SHALL 允许已鉴权管理员按 `id` 删除版本表记录，并在安全条件下清理关联 APK：**path-only** 行尽力删除 `apkStorageDir` 内文件；**`apk/` objectKey** 行尽力经 ucg 删除 OSS 对象。清理失败 MUST NOT 阻止删行成功，MUST 记录可观测警告。

#### Scenario: 成功删除数据库行

- **WHEN** 已鉴权管理员删除存在的 `id`
- **THEN** 系统 SHALL 从版本表移除该行并返回成功响应

#### Scenario: 删除后失效最新版本缓存

- **WHEN** 删除操作成功
- **THEN** 系统 SHALL 失效用于 `version/check` 的最新行缓存，使检查按剩余行中最大 `id` 重新加载

#### Scenario: 尽力删除磁盘 APK

- **WHEN** 被删行的 `download_url` 为约定 path-only 形式（前缀 `/device/app/apk/`）且文件名为安全 APK 名且在存储目录内存在对应文件
- **THEN** 系统 SHALL 尝试删除该文件；若删除失败 SHALL 仍完成删行并记录可观测警告日志

#### Scenario: 尽力删除 OSS APK

- **WHEN** 被删行的 `download_url` 为 `apk/` 前缀 objectKey
- **THEN** 系统 SHALL 尝试经 ucg 删除该 OSS 对象；若删除失败 SHALL 仍完成删行并记录可观测警告日志

#### Scenario: 非法路径不删盘外文件

- **WHEN** `download_url` 不含约定 path 前缀且不是安全的 `apk/` objectKey，或文件名未通过安全校验
- **THEN** 系统 SHALL 仅删除数据库行且 SHALL NOT 删除存储目录外或路径穿越目标

### Requirement: 新增发版与现有上传兼容

「新增（Create）」SHALL 继续通过现有 `POST /device/app/api/version/admin/upload` multipart 接口完成（APK 校验、经 ucg 上传 OSS、插入新行、写入 objectKey `download_url`）；调用 MUST 携带有效 Admin JWT。

#### Scenario: 上传成功仍插入新行

- **WHEN** 客户端携带有效 Admin JWT 上传合法 APK 及 `latestVersion`
- **THEN** 系统 SHALL 插入新的版本行并失效最新版本缓存
- **AND** `download_url` SHALL 为 `apk/` objectKey
- **AND** 响应 `downloadUrl` SHALL 为 CDN 绝对 URL

### Requirement: 与版本检查接口语义一致

增删改及上传写库后，`GET /device/app/api/version/check` 所依据的「最新版本行」SHALL 仍为版本表中 **`id` 最大** 且 `latest_version` 非空的一行；返回的 `downloadUrl` SHALL 按该行存储形态映射（OSS→CDN 绝对 URL；历史 path→path-only）。

#### Scenario: 删除当前最大 id 后检查回落

- **WHEN** 表中仍存在其它发版行且管理员删除了当前 `id` 最大的行
- **THEN** 客户端调用 `version/check` 时 SHALL 使用剩余行中 `id` 最大者作为最新发版

### Requirement: 管理页展示与操作

`gateway-app-version-admin.html`（或等价路由页面）SHALL 在持有有效 Admin JWT 时展示历史版本列表，并 SHALL 提供触发列表刷新、编辑元数据、删除记录及上传新版本 APK 的交互；所有写操作请求 SHALL 携带 `Authorization: Bearer` Admin JWT，MUST NOT 依赖版本管理独立 Cookie。

#### Scenario: 登录后可见历史表格

- **WHEN** 管理员已在 Hub 取得 Admin JWT 并打开版本管理页
- **THEN** 页面 SHALL 请求列表接口并展示历史版本行

#### Scenario: 操作后刷新列表

- **WHEN** 上传、更新或删除任一操作成功
- **THEN** 页面 SHALL 刷新列表以反映当前表状态

### Requirement: 版本管理未启用时的错误语义

当网关进程未配置 `GATEWAY_APP_ADMIN_PASSWORD`（Hub 登录不可用）时，受保护的管理接口（含列表、查询、更新、删除、上传）SHALL 与 Hub login 一致返回服务不可用语义，且 SHALL NOT 暴露写库能力。

#### Scenario: 未配置 admin 密码拒绝受保护接口

- **WHEN** `GATEWAY_APP_ADMIN_PASSWORD` 为空且客户端调用受保护版本管理接口
- **THEN** 系统 SHALL 返回与「管理未启用」一致的不可用响应

---

### Requirement: 版本表无数据时版本检查须成功且无需更新

gateway-app-server 对 **`GET /device/app/api/version/check`** SHALL 在版本配置表（如 `app_version`）中**不存在任何可用版本行**时，仍返回 **`code=0`**。响应 **`data.needUpdate`** SHALL 为 **`false`**。响应 SHALL NOT 因「结果集无行」或等价空表语义返回非 0 业务码。

#### Scenario: 表无任何记录

- **WHEN** 版本表为空或查询不到任何版本行
- **THEN** HTTP 业务包装 SHALL 为成功（`code=0`）且 **`needUpdate` 为 false**，且 SHALL NOT 将空结果集作为错误返回给客户端

#### Scenario: 有版本记录时行为不变

- **WHEN** 存在至少一条版本记录且 `latestVersion` 可解析
- **THEN** 系统 SHALL 继续按现有规则比较 `currentVersion` 与 `latestVersion` 并设置 **`needUpdate`**

### Requirement: 区分空表与数据库基础设施故障

当版本表查询因**连接、权限、语法等**失败时，系统 MAY 返回非 0 业务码或错误信息以便运维定位。系统 SHALL NOT 将**仅无匹配行**与上述基础设施错误等同为「统一失败」而掩盖空表成功语义。

#### Scenario: 真实读库错误

- **WHEN** 数据库返回非「无行」类错误（如连接失败）
- **THEN** 系统 MAY 返回错误响应且 SHOULD NOT 冒充「无需更新」的成功语义

### Requirement: 最新行为 OSS 时 downloadUrl SHALL 为 CDN 绝对 URL

当 `GET /device/app/api/version/check` 所选取的最新版本行的 `download_url` 为 OSS objectKey（前缀 `apk/`）或已是 `http(s)` CDN URL 时，响应中的 `downloadUrl` MUST 为客户端可直接使用的 CDN 绝对 URL；MUST NOT 要求客户端再拼接 gateway-app 基址。

#### Scenario: OSS 最新行返回可直用 CDN

- **WHEN** 客户端请求 `version/check` 且最新行 `download_url` 为 `apk/` objectKey
- **THEN** 响应 `downloadUrl` MUST 以 `https://` 开头
- **AND** 客户端无需拼接 gateway 基址即可发起下载

---

### Requirement: Gateway MUST 保持无业务后台任务职责

`gateway-service` MUST 仅承担请求入口、路由转发与横切能力，MUST NOT 在进程启动阶段承担消息消费、事件中继、domain outbox relay 等业务后台任务职责。

#### Scenario: 网关处理请求

- **WHEN** gateway 接收 HTTP/WS 请求
- **THEN** gateway MUST 仅执行入口与代理逻辑，不应存在后台任务消费副作用

#### Scenario: 部署配置审查

- **WHEN** 审查 gateway 部署配置与启动流程
- **THEN** 必须能够确认 gateway 未启动 ticker 扫表或 MQ 消费；history/device 缓存由各自服务同步 patch；UCG 后台任务仅在 ucg-service

---

### Requirement: Gateway 必须收敛为流量与策略层

`gateway-service` SHALL 仅提供边缘层能力，包括鉴权、路由、策略执行、元数据透传和流量控制。

#### Scenario: 请求进入 gateway
- **WHEN** 客户端请求到达 `gateway-service`
- **THEN** gateway SHALL 执行边缘策略并转发至对应领域服务，不得执行领域业务规则

### Requirement: Gateway 在委派领域执行时必须保持外部契约稳定

`gateway-service` SHALL 在服务拆分过程中保持对外 API 契约稳定，并 SHALL 将领域业务执行委派给下游领域服务。

#### Scenario: 拆分后的既有外部 API 调用
- **WHEN** 客户端调用既有公开 API 端点
- **THEN** gateway SHALL 在调用下游领域服务处理业务的同时返回契约兼容响应

---

### Requirement: Gateway 路由中间件 MUST 按领域拆分管理

gateway MUST 将 voice 与 device 路由代理逻辑拆分为独立中间件与配置读取路径，不得在同一中间件实现中混合管理两个领域。

#### Scenario: 修改 voice 路由逻辑
- **WHEN** 开发者调整 voice 路由代理策略
- **THEN** 变更 MUST 限定在 voice 独立中间件实现内，且不应直接影响 device 路由行为

#### Scenario: 修改 device 路由逻辑
- **WHEN** 开发者调整 device 路由代理策略
- **THEN** 变更 MUST 限定在 device 独立中间件实现内，且不应直接影响 voice 路由行为

### Requirement: 领域路由配置 MUST 互相隔离

voice 与 device 的路由模式、目标地址、canary 百分比配置 MUST 分别独立，禁止共享配置键。

#### Scenario: 仅调整 voice canary 百分比
- **WHEN** 运维仅修改 voice 的 canary 百分比配置
- **THEN** device 路由行为 MUST 保持不变

---

### Requirement: gateway-app-server SHALL proxy voice App HTTP APIs

`gateway-app-server` MUST 将 **`/voice/app/api/*`** 注册为 HTTP 反向代理至 voice-service（环境变量 `VOICE_API_PROXY_URL`、`VOICE_API_ROUTE_MODE`，模式对齐 `device_route_proxy.go` / `ucg_route_proxy.go`）。对受保护路径 MUST 经 Bearer 鉴权并注入 `X-Internal-Wx-Id`（及有值时的 `X-Internal-Device-No`）后转发。gateway-app MUST NOT 在本地实现 voice App 业务逻辑。

#### Scenario: App 查询 voice 域额度经 gateway 反代

- **WHEN** Flutter 携带有效 Bearer 请求 `GET /voice/app/api/ai-quota`
- **THEN** gateway-app MUST 注入内部头并转发至 voice-service 同路径
- **AND** gateway-app MUST NOT 本地聚合 polish 或 clinic 数据

#### Scenario: 反代目标不可达

- **WHEN** `VOICE_API_ROUTE_MODE=proxy` 且 voice-service 不可达
- **THEN** gateway-app MUST 返回可诊断的代理错误

### Requirement: gateway-app-server SHALL proxy voice Admin HTTP APIs with password injection

`gateway-app-server` MUST 将 **`/voice/admin/api/*`** 反代至 voice-service。Admin JWT 校验通过后，`InjectAdminDownstreamPassword` MUST 对 `/voice/admin/api/` 前缀注入 `X-Admin-Password`（值来自 **`VOICE_ADMIN_PASSWORD`** env / `voice.admin.password` 配置）。

#### Scenario: voice admin API 口令注入

- **WHEN** 已登录 Admin Hub 的用户请求 PUT `/voice/admin/api/ai-quota/default`
- **THEN** gateway-app MUST 注入 `X-Admin-Password` 并转发至 voice-service

#### Scenario: 未登录 Admin 拒绝

- **WHEN** 请求 `/voice/admin/api/*` 且无有效 Admin JWT
- **THEN** gateway-app SHALL 返回未授权且 SHALL NOT 转发

### Requirement: gateway-app-server SHALL remove device ai-quota App proxy

`gateway-app-server` MUST **移除** `device_route_proxy.go` 中对 **`/device/app/api/ai-quota`** 的反代登记。该路径 MUST NOT 再可达 device-service ai-quota 读 API。

#### Scenario: 旧 App 读路径不可用

- **WHEN** 客户端请求 `GET /device/app/api/ai-quota`
- **THEN** gateway-app SHALL NOT 反代至 device ai-quota（返回 404 或由网关本机明确拒绝）

### Requirement: gateway-service SHALL sync voice HTTP proxy paths

`gateway-service` MUST 同步注册 `/voice/app/api/*` 与 `/voice/admin/api/*` 反代，行为与 gateway-app 对齐，以便管理/通用网关与 App 网关路径一致。

#### Scenario: gateway-service 反代 voice App API

- **WHEN** 客户端经 gateway-service 请求 `/voice/app/api/ai-quota` 且 proxy 配置正确
- **THEN** gateway-service MUST 转发至 voice-service

---

### Requirement: Gateway MUST 移除本地 voice WS 领域执行

`gateway-service` MUST 不再在 `/voice/chat/ws` 执行本地语音对话业务逻辑，领域处理必须由 `voice-service` 承担。

#### Scenario: Gateway 收到 voice WS 请求
- **WHEN** 客户端连接 `/voice/chat/ws`
- **THEN** gateway MUST 仅执行边缘层职责（路由、策略、元数据透传），并将领域执行委派给 `voice-service`

### Requirement: Gateway SHALL 保持对外 WS 入口契约稳定

迁移到委派模式时，gateway MUST 保持外部 WS 路径与接入方式稳定，避免要求前端同步改地址。

#### Scenario: 前端继续使用既有 WS 地址
- **WHEN** 前端仍连接 gateway 既有 `/voice/chat/ws` 地址
- **THEN** 系统 MUST 可完成握手与消息收发，且业务执行由下游 `voice-service` 完成

---

### Requirement: Gateway SHALL 支持 voice WebSocket 边缘透传

`gateway-service` MUST 在 `/voice/chat/ws` 提供可配置透传能力，将 WebSocket 连接转发到 `voice-service` 目标地址。

#### Scenario: WS 透传启用且目标可达
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为可达地址
- **THEN** gateway MUST 将 `/voice/chat/ws` 的握手与后续双向消息透传至目标服务

#### Scenario: WS 透传启用但目标不可达
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **THEN** gateway MUST 返回明确的握手/代理失败错误，且 MUST NOT 回退本地业务执行

### Requirement: Gateway MUST 提供 WS 透传配置约束

gateway MUST 通过环境变量控制 WS 路由行为，并对非法配置执行可预测回退。

#### Scenario: 路由模式非法
- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **THEN** gateway MUST 将 WS 路由模式视为 `local`

#### Scenario: 代理地址为空或非法
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为空或非法
- **THEN** gateway MUST 视为未启用可用代理目标并返回可诊断错误，不得出现静默成功

### Requirement: Gateway SHALL 支持听写 WebSocket 边缘透传

`gateway-service` 与 `gateway-app-server` MUST 在 `/voice/asr/ws` 提供与 `/voice/chat/ws` 相同的可配置 WebSocket 透传能力，将连接转发至 `voice-service`（与 `VOICE_WS_PROXY_URL` 同一目标基址）。

#### Scenario: WS 透传启用且目标可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为可达地址
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 将该路径的握手与后续双向消息透传至 voice-service，行为与 `/voice/chat/ws` 一致

#### Scenario: WS 透传启用但目标不可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 返回明确的握手/代理失败错误，且 MUST NOT 在 gateway 本地执行听写业务逻辑

#### Scenario: 路由模式非 proxy

- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 返回可诊断的配置错误（与 chat WS 一致），且 MUST NOT 静默成功

### Requirement: App 网关 SHALL 将听写 WS 纳入 Bearer 白名单

`gateway-app-server` MUST 将 `GET /voice/asr/ws`（WebSocket Upgrade）列入 Bearer 鉴权豁免路径，与 `/voice/chat/ws` 策略一致。

#### Scenario: 无 Bearer 的 Upgrade 请求

- **WHEN** App 客户端对 `/voice/asr/ws` 发起 WebSocket Upgrade 且未携带 `Authorization`
- **THEN** gateway-app MUST 允许进入透传或 voice-service 处理链，不得仅因缺少 Bearer 拒绝 Upgrade

### Requirement: gateway-app-server SHALL 注册胖宝 clinic WebSocket 为 App 对外入口

`gateway-app-server` MUST 将 `GET /voice/clinic/ws` 注册为 App 客户端的**唯一对外 WebSocket 入口**（与 `/voice/chat/ws`、`/voice/asr/ws` 同 `apiBaseUrl` 主机、同 `installVoiceWSProxyMiddleware` 透传链）。实现 MUST 将 `/voice/clinic/ws` 加入 `internal/controller/ws_route_proxy.go` 的 `voiceWSProxyPaths`，由 `RegisterGatewayAppHTTP` 已挂载的 `installVoiceWSProxyMiddleware` 将握手与双向消息透传至 `voice-service`（`VOICE_WS_PROXY_URL` 同一目标基址）。App 客户端 MUST NOT 被要求或配置为直连 `voice-service` 内网地址。

#### Scenario: App 经 gateway-app 连接 clinic WS

- **WHEN** Flutter 使用 `wss://{apiBaseUrl host}/voice/clinic/ws` 发起 WebSocket Upgrade
- **AND** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 可达
- **THEN** gateway-app-server MUST 透传握手与后续双向帧至 voice-service `/voice/clinic/ws`
- **AND** gateway-app MUST NOT 在本地执行 clinic 业务逻辑

#### Scenario: WS 透传启用但目标不可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **AND** 客户端连接 gateway-app `/voice/clinic/ws`
- **THEN** gateway-app MUST 返回明确的握手/代理失败错误，且 MUST NOT 在 gateway 本地执行 clinic 业务逻辑

#### Scenario: 路由模式非 proxy

- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **AND** 客户端连接 gateway-app `/voice/clinic/ws`
- **THEN** gateway-app MUST 返回可诊断的配置错误，且 MUST NOT 静默成功

### Requirement: gateway-service MUST 同步 clinic WebSocket 透传路径

`gateway-service` MUST 将 `/voice/clinic/ws` 加入同一 `voiceWSProxyPaths` 列表，行为与 `/voice/asr/ws` 一致，以便管理/通用网关与 App 网关路径对齐；**App 主入口仍为 gateway-app-server**。

#### Scenario: gateway-service 透传 clinic WS

- **WHEN** 客户端连接 gateway-service `/voice/clinic/ws` 且 proxy 模式配置正确
- **THEN** gateway-service MUST 透传至 voice-service，行为与 chat/ASR WS 一致

### Requirement: gateway-app-server SHALL 将 clinic WS 纳入 Bearer 白名单

`gateway-app-server` MUST 将 `GET /voice/clinic/ws`（WebSocket Upgrade）列入 `gateway_app_auth_exempt.go` 的 `gatewayAppAuthExemptExactGET`，与 `/voice/asr/ws` 策略一致：Upgrade 不要求 HTTP 层 Bearer；若客户端仍携带 Bearer，`HookBeforeServe` MAY 注入 `X-Internal-Wx-Id`，但 clinic 身份校验 MUST 由 voice-service 首帧 `auth` 完成。

#### Scenario: 无 Bearer 的 Upgrade 请求

- **WHEN** App 客户端对 gateway-app `/voice/clinic/ws` 发起 WebSocket Upgrade 且未携带 `Authorization`
- **THEN** gateway-app SHALL 允许请求进入 WS 透传链（由 voice-service 首帧 `auth` 校验 wxId；**非** deviceNo 反查）

#### Scenario: 可选 Bearer 仍注入内部头

- **WHEN** App 对 `/voice/clinic/ws` Upgrade 且携带有效 App access Bearer
- **THEN** gateway-app MAY 经 `InjectAccessHeadersFromBearer` 注入 `X-Internal-Wx-Id`
- **AND** voice-service clinic handler MUST NOT 仅以该头作为鉴权依据（首帧 JWT 为准）

---

### Requirement: gateway-app SHALL HTTP-proxy /ucg/app/api to ucg-service

gateway-app-server SHALL register reverse proxy for path prefix `/ucg/app/api/*` to configured `UCG_SERVICE_BASE_URL`, applying existing Bearer JWT validation and injecting `X-Internal-Wx-Id` from JWT `sub` before forwarding. CORS behavior SHALL match other domain proxies.

#### Scenario: 鉴权后转发
- **WHEN** App 带合法 Bearer 请求 `/ucg/app/api/profile/me`
- **THEN** gateway SHALL 转发至 ucg-service 且 SHALL 设置 `X-Internal-Wx-Id`

#### Scenario: 推荐接口匿名可读
- **WHEN** 产品配置推荐 Feed 为匿名可读且请求在白名单内
- **THEN** gateway SHALL 允许无 Bearer 转发 `/ucg/app/api/feed/recommend`（若实现匿名策略）

### Requirement: gateway-app SHALL WebSocket-proxy /ucg/app/ws/chat to ucg-service

gateway-app-server SHALL register WebSocket upgrade reverse proxy for exact path `/ucg/app/ws/chat` to ucg-service internal endpoint `/ws/chat`, using the same `httputil.ReverseProxy` pattern as `ws_route_proxy.go` / voice WS edge proxy. Configuration SHALL use `UCG_WS_ROUTE_MODE` and `UCG_WS_PROXY_URL`. App clients MUST NOT connect directly to ucg-service for chat.

#### Scenario: WS 经网关同域
- **WHEN** 客户端连接 `wss://{apiBaseUrl host}/ucg/app/ws/chat`
- **THEN** gateway SHALL 透传至 ucg-service `/ws/chat`，且 SHALL NOT 要求 App 配置独立 ucg-service 公网 WS 域名

#### Scenario: WS 代理目标不可达
- **WHEN** `UCG_WS_ROUTE_MODE=proxy` 且 ucg-service WS 不可达或握手失败
- **THEN** gateway SHALL 返回可诊断的 `ws_proxy` 阶段错误，且 SHALL NOT 静默成功

---

### Requirement: Apple AASA 文件 SHALL 在 `www.pangbao.cuplay.top` 主机根路径可访问

系统 SHALL 为胖宝 iOS 应用提供 Apple `apple-app-site-association` 文件，并且 MUST 同时支持 `GET /apple-app-site-association` 与 `GET /.well-known/apple-app-site-association`。两条路径返回的内容 MUST 等价、响应状态 MUST 为 `200`、传输协议 MUST 为 HTTPS，且响应不得被改写到任何其它业务路径后才可获取。

#### Scenario: Apple 从根路径获取 AASA 文件
- **WHEN** 运维或校验程序请求 `https://www.pangbao.cuplay.top/apple-app-site-association`
- **THEN** 系统返回 `200` 和可解析的 AASA JSON 内容，且不发生 301/302 到其它路径

#### Scenario: Apple 从 well-known 路径获取 AASA 文件
- **WHEN** 运维或校验程序请求 `https://www.pangbao.cuplay.top/.well-known/apple-app-site-association`
- **THEN** 系统返回与根路径等价的 AASA JSON 内容，并使用适合 JSON/AASA 的响应头

### Requirement: AASA 内容 SHALL 与微信 Universal Links 前缀保持一致

AASA 内容 MUST 使用 `appIDs = ["<TEAM_ID>.com.fzy.pangbao"]` 的结构，其中 `com.fzy.pangbao` 为固定 Bundle ID，`<TEAM_ID>` 为部署时注入的真实 Apple Team ID。AASA `components` 或等价路径约束 MUST 放行 `https://www.pangbao.cuplay.top/wx/ulink/` 对应的 `/wx/ulink/*` 路径，使微信开放平台填写值、iOS 客户端 `universalLink` 和服务端声明保持一致。

#### Scenario: Team ID 已配置时生成正式 AASA 内容
- **WHEN** 部署环境已经提供真实 Team ID
- **THEN** AASA 中的 `appIDs` 使用 `<真实TeamID>.com.fzy.pangbao`，且放行路径覆盖 `/wx/ulink/*`

#### Scenario: 微信后台使用推荐的 Universal Links 前缀
- **WHEN** 接入人员在微信开放平台填写 Universal Links
- **THEN** 文档与服务端约束均指向 `https://www.pangbao.cuplay.top/wx/ulink/`

### Requirement: Team ID 缺失时系统 SHALL 提供显式待配置语义

在 Team ID 尚未提供的阶段，仓库 MUST 保留明确的 AASA 模板或配置占位说明；正式对外端点在未配置 Team ID 时 MUST 返回显式不可验证语义或运维可识别的失败提示，而不是伪造一个看似可用的生产 `appIDs`。

#### Scenario: 生产配置缺少 Team ID
- **WHEN** AASA 端点所在环境未设置 Team ID
- **THEN** 系统返回显式错误或不可用提示，并在日志/文档中指向需要补充的配置项

#### Scenario: 仓库中保留待补位模板
- **WHEN** 开发人员阅读仓库内 Universal Links 相关资源
- **THEN** 可以看到 Team ID 待补位规则，以及 `com.fzy.pangbao` 已固定、仅 Team ID 需要在部署前补齐

### Requirement: 仓库 SHALL 提供 GitHub 打包上架的 Universal Links 操作文档

仓库 MUST 提供面向 GitHub 打包链路的 runbook，明确 iOS 工程需要开启 `Associated Domains`、加入 `applinks:www.pangbao.cuplay.top`、保证 Provisioning Profile 启用该能力，并在微信 SDK 注册配置中使用与 AASA 一致的 `https://www.pangbao.cuplay.top/wx/ulink/`。文档 MUST 说明该流程适用于 GitHub Actions / CI 打包，不要求人工在本地 Xcode 界面逐步操作才能理解。文档 MUST 同时明确 `http://www.pangbao.cuplay.top/` 不能作为 Universal Links 或 AASA 校验地址。

#### Scenario: GitHub Actions 打包配置指引可读
- **WHEN** 维护者按照 runbook 配置 GitHub 打包环境
- **THEN** 可以明确知道需要准备哪些证书/描述文件/Secrets、如何确认 entitlements 被正确签入产物

#### Scenario: 发布后可执行 Universal Links 验证
- **WHEN** 维护者完成部署与打包
- **THEN** runbook 提供 `curl`、Apple/微信侧检查项或真机验证步骤，以确认 Universal Links 已生效

---

### Requirement: gateway-app MUST 在根路径提供微信域名校验 TXT

`gateway-app-server` MUST 对 `GET` 与 `HEAD` 请求路径 `/90fafbe9bf8308ecbd2063d7b479a309.txt` 返回 HTTP 200，响应体为纯文本校验串 `41a8fe99b666f6646724f29ea87962e34c6df2b6`（不得包裹 HTML、JSON 或网关统一业务 envelope）。静态文件 MUST 存放于 `resource/public/90fafbe9bf8308ecbd2063d7b479a309.txt`，并由显式路由提供（与官网根路径静态页同一托管进程）。该路径 MUST 允许匿名访问（Bearer 豁免），且 MUST NOT 计入 App API usage 统计。

#### Scenario: 微信爬虫校验域名归属

- **WHEN** 未登录客户端请求 `GET /90fafbe9bf8308ecbd2063d7b479a309.txt`（公网形态为 `https://www.pangbao.cuplay.top/90fafbe9bf8308ecbd2063d7b479a309.txt`）
- **THEN** 系统 SHALL 返回 200，且响应体 SHALL 为 `41a8fe99b666f6646724f29ea87962e34c6df2b6`

#### Scenario: 校验路径不要求登录

- **WHEN** 请求未携带 App Bearer Token
- **THEN** 系统 SHALL 仍返回校验 TXT，MUST NOT 因鉴权失败而 401 或跳转登录页

#### Scenario: 校验路径不计入 usage

- **WHEN** 客户端访问 `/90fafbe9bf8308ecbd2063d7b479a309.txt`
- **THEN** 系统 SHALL NOT 将该请求记入 App API 使用统计

---
