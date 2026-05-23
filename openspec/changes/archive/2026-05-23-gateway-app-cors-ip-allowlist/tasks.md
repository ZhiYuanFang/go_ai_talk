## 1. CORS 匹配与响应头

- [x] 1.1 在 `internal/controller` 或 `internal/services/gatewayapp` 中实现「解析 `Origin` → 校验 scheme 为 http/https → 校验 hostname 属于 `192.168.0.131` / `120.55.50.105`」的纯函数，通过则返回待回显的 Origin 字符串，否则返回空。
- [x] 1.2 在通过校验时，为响应设置 `Access-Control-Allow-Origin`（回显）、`Access-Control-Allow-Methods`（含 `GET, POST, OPTIONS`）、`Access-Control-Allow-Headers`（至少 `Content-Type, Authorization`）；按需设置 `Access-Control-Max-Age`（可选，便于减少预检）。

## 2. 中间件接入与 OPTIONS

- [x] 2.1 在 `installGatewayCrosscuttingMiddlewares`（或紧邻的全局 `/*` 中间件）中挂载 CORS：在 `r.Middleware.Next()` 前后写入头；对匹配白名单的 `OPTIONS` 按 `design.md` 决策实现短路 204 或依赖框架路由，确保预检为 2xx 且带齐 CORS 头。
- [x] 2.2 确认 `HookBeforeServe` 与中间件顺序下，OPTIONS 仍不因 Bearer 返回 401（与 `gateway_app_auth_exempt.go` 行为一致）。

## 3. 验证

- [x] 3.1 使用 `curl` 模拟预检：`OPTIONS` + 白名单 `Origin` + `Access-Control-Request-Method: POST`，检查响应头与状态码。
- [x] 3.2 使用 `curl` 带非白名单 `Origin` 检查响应中不出现 `Access-Control-Allow-Origin`（或不为该 Origin 回显）。
