## Context

- `gateway-app-server` 使用 GoFrame `ghttp`，路由在 `internal/controller/gateway_app_register.go` 注册；跨切面逻辑在 `gateway_crosscutting.go`。
- Bearer 注入通过 `HookBeforeServe` 实现；`gatewayAppPathAuthExempt` 已对 **OPTIONS** 整方法豁免，避免预检被 401。
- 仓库内尚无 CORS 实现；需在网关进程内集中处理，避免各反代路由重复配置。

## Goals / Non-Goals

**Goals:**

- 对 `Origin` 的 host 属于 `{192.168.0.131, 120.55.50.105}` 的请求，将 **完整 `Origin` 值** 写入 `Access-Control-Allow-Origin`（从而支持任意端口、http/https）。
- 预检与正常响应均返回一致的 CORS 头集合；允许方法 `GET, POST, OPTIONS`；允许头包含 `Content-Type` 与 `Authorization`。
- 不匹配白名单的请求 **不** 设置 `Access-Control-Allow-Origin`（非全回显）。

**Non-Goals:**

- 不在本变更中实现可运维动态配置（环境变量/配置文件）；IP 列表以代码常量或单一模块内常量维护即可，后续可再抽配置。
- 不扩大至其他网关进程（如主 gateway）；仅 `gateway-app-server`。
- 不引入新的跨服务契约或修改 API 路径语义。

## Decisions

1. **匹配规则**  
   - 使用标准库解析 `Origin`（如 `url.Parse`），要求 scheme 为 `http` 或 `https`，host **不含端口时** 与 IP 字面量相等；含端口时 **hostname 部分**（`Host`）等于允许 IP 即通过。  
   - 理由：避免手写字符串前缀误判；与「同一 IP 任意端口」一致。

2. **放置位置**  
   - 在 `installGatewayCrosscuttingMiddlewares` 同一链路增加 CORS 中间件（或合并到该 `/*` 中间件内靠前位置），保证反代与自有 API 均覆盖。  
   - 理由：与现有「全网关追踪 ID」一致，单点维护。

3. **OPTIONS 与正文**  
   - 中间件内若 `OPTIONS` 且为 CORS 预检（可检查 `Access-Control-Request-Method` 存在，或一律对 OPTIONS 在白名单 Origin 下补头），在写完 CORS 头后 **`r.Middleware.Next()` 仍调用** 或由框架处理；若业务层无匹配路由导致 404，则改为在 CORS 匹配时对 OPTIONS **短路返回 204**（无 body），避免 404 影响预检。  
   - 理由：浏览器预检只要求 2xx 与 CORS 头；需在实现阶段验证 GoFrame 默认 OPTIONS 行为并二选一写清。

4. **`Access-Control-Allow-Credentials`**  
   - 默认 **不** 主动设为 `true`，除非现有 Web 联调明确使用 Cookie 且 `fetch` 使用 `credentials: 'include'`。Bearer-only 通常不需要。  
   - 若后续需要，在常量或配置中单独打开，并与「不可与 `*` 共用 Allow-Origin」的规范保持一致（本设计为回显 Origin，已满足）。

5. **与 Bearer Hook 顺序**  
   - CORS 中间件在 `BindMiddleware` 阶段执行；`HookBeforeServe` 在 ServeHTTP 中先于 Middleware。预检 OPTIONS 已豁免 Bearer，故 **无冲突**；实现时保持 OPTIONS 豁免不变。

## Risks / Trade-offs

- **[Risk] 固定 IP 白名单仍可能被仿冒 Origin 的非浏览器客户端调用**  
  - **缓解**：CORS 仅约束浏览器；服务端仍依赖 Bearer 与其它安全控制。

- **[Risk] 公网 IP `120.55.50.105` 若暴露给不可信前端，白名单范围大于纯内网**  
  - **缓解**：此为联调阶段性取舍；后续可改为配置或缩窄来源。

- **[Risk] OPTIONS 短路与普通路由重复**  
  - **缓解**：单一路径写单元测试或手工预检验证（当前仓库阶段可不新增测试文件，以手工联调为准）。

## Migration Plan

- 部署新版本 `gateway-app-server` 即可；无数据迁移。回滚为撤销该中间件。

## Open Questions

- Web 是否使用 `credentials: 'include'`（Cookie）？若是，实现时显式开启 `Access-Control-Allow-Credentials: true` 并在规格中固化。
