## Context

- `gateway-app-server` 在 `installGatewayAppCORSMiddleware` 中对 `/*` 写入 CORS，并在 `OPTIONS` 上全局短路 204。
- `installHistoryProxyMiddleware` 对 `/device/history/api/*` 在命中代理时调用 `httputil.ReverseProxy.ServeHTTP` 后 `r.ExitAll()`，与主网关注册说明中「ReverseProxy 可能提前 Flush、缓冲与二次写体」风险同源：**网关在 `Next` 前后写入的响应头，未必稳定出现在浏览器收到的最终响应中**，或与下游拷贝头行为叠加后被覆盖。
- Flutter Web 对 **带 `Authorization` 的 GET** 会先发 **OPTIONS 预检**，再发 GET；列表接口走 history 代理时，若 GET 响应缺少 `Access-Control-Allow-Origin` 等，浏览器仍报跨域失败。

## Goals / Non-Goals

**Goals:**

- 对 **命中 history 反向代理** 且 **Origin 通过 `ReflectGatewayAppCORSOrigin`** 的请求，浏览器收到的最终响应 **必须** 含与直连 `gateway_app_cors` 一致的 CORS 语义（回显 Origin、Methods、Headers、Max-Age 等与现实现或抽取的单一真源对齐）。
- 优先在 **ReverseProxy 边界**（如 `ModifyResponse` 或经评审的等价机制）合并头，避免依赖「外层中间件在 `ExitAll` 后是否仍执行尾逻辑」的隐式顺序。
- 若 voice/device 代理存在同类现象，**同一套合并策略**复用，避免三套漂移。

**Non-Goals:**

- 不扩大 CORS 白名单主机集合（仍由现有 `gatewayapp` 规则或后续独立配置变更负责）。
- 不修改 history-service 自身是否返回 CORS（仍以网关为浏览器边界）。
- 不在本变更中强制新增自动化测试文件（遵守仓库当前约定）。

## Decisions

1. **采用 `httputil.ReverseProxy.ModifyResponse`（或 Go 版本支持的等价钩子）在收到下游 `*http.Response` 后注入 CORS**  
   - **理由**：在「下游头已解析完毕、写回客户端前」这一稳定点合并，避免与 `ExitAll`、Flush 顺序竞态。  
   - **备选**：包装 `http.ResponseWriter` 拦截 `WriteHeader`——侵入面更大，暂缓。

2. **CORS 注入逻辑复用 `gatewayapp.ReflectGatewayAppCORSOrigin` + 与 `writeGatewayAppCORSHeaders` 相同的头集合**  
   - **理由**：单一语义来源，避免 `/device/app/api` 与 `/device/history/api` 两套常量漂移。可抽取为 `gatewayapp` 或 `controller` 内小函数供中间件与 `ModifyResponse` 共用。

3. **仅当 `ModifyResponse` 收到非 nil `*http.Response` 且 Origin 校验通过时写入；错误响应（nil body 等）按 Go 文档谨慎处理**  
   - **理由**：不破坏现有错误与状态码语义。

4. **history 优先落地；voice/device 若当前无跨域报告，仍建议共用 `buildReverseProxy` 的扩展点一次接好**  
   - **理由**：降低后续同类工单重复。

## Risks / Trade-offs

- **[Risk] `ModifyResponse` 返回错误会导致网关将 502 写给客户端**  
  - **缓解**：注入逻辑仅 `Header.Set`，不读 body；单测或本地用带 `Origin` 的 curl/浏览器验证。

- **[Risk] 下游已设置冲突的 `Access-Control-*`**  
  - **缓解**：以网关白名单策略覆盖为准（显式 Set 覆盖同名字段）。

## Migration Plan

- 发布新版本 `gateway-app-server`；无数据迁移。回滚为撤销代理层 CORS 注入。

## Open Questions

- 是否需在 **主网关**（`:9701`）同步（若未来同一前端也直连主网关的 history 代理）？本 proposal 范围仍限定 **gateway-app-server**。
