## MODIFIED Requirements

### Requirement: gateway-app SHALL record successful App HTTP API usage after response

`gateway-app-server` MUST 在 HTTP 响应确定后（状态码已写入客户端方向）评估是否写入使用统计。实现 MUST 覆盖经领域反代（device/ucg/history/voice）短路 `ExitAll` 的路径，不得仅依赖 `BindMiddleware("/*")` 在 `Next()` 之后记录。仅当响应状态码满足 `200 <= status < 300` 时 SHALL 计数一次。统计路径 MUST 为归一化后的 `METHOD /path`（不含 query）。下列请求 MUST NOT 写入统计：WebSocket 升级、`/device/internal/` 前缀、`/device/admin/api/` 前缀（含本变更读 API 自身）、静态资源与 HTML 壳页，以及维护型 App API（`POST /device/app/api/token/refresh`、`GET /device/app/api/version/check`、`GET /device/app/api/site/home`、`/device/app/api/version/admin/*` 前缀、**`GET /ucg/app/api/posts/{id}/comments`**）。**此外，当请求关联的 `wxId > 0` 且该 wxId 在 device 域标记为 `is_simulated=1` 时，MUST NOT 写入任何 usage 统计（全局日计数、per-wxId、交叉维度均跳过）。** 登录、注册、绑定、POST 评论与各业务 App API 对**非模拟**用户 SHALL 继续计入。写入 MUST 异步执行且 SHALL NOT 阻塞或改变业务响应。

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
