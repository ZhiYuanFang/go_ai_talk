## ADDED Requirements

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
