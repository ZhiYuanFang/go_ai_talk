## Why

Flutter Web 从本机页（如 `http://localhost:*`）跨域访问 `gateway-app-server` 上 **`/device/app/api/*`** 时，现有全局 CORS 中间件已能覆盖；但访问经 **`httputil.ReverseProxy`** 短路的 **`/device/history/api/*`**（如列表 `GET .../device/history/api/list`）时仍出现浏览器 CORS 失败。根因是代理链与响应写回方式使「仅在中间件 `Next` 前后写头」不足以稳定合并到最终对浏览器的响应，需在代理边界显式合并 CORS 策略。

## What Changes

- 在 **history（及必要时 voice/device）反向代理** 路径上，保证在 **Origin 命中现有 App 网关 CORS 白名单** 时，最终 HTTP 响应 **必须** 包含与直连 API 一致的 `Access-Control-*` 头（至少 `Allow-Origin` 回显、`Allow-Methods`、`Allow-Headers`；与现有 `gateway_app_cors` 语义对齐）。
- 不改变既有 Bearer 注入、分流键与代理目标 URL 配置语义；**不**放宽白名单（仍由 `gatewayapp.ReflectGatewayAppCORSOrigin` 或后续统一配置定义）。
- 联调说明：`curl` **若不携带 `Origin` 头** 无法复现浏览器 CORS 行为，文档或 runbook 中补充示例命令。

## Capabilities

### New Capabilities

- `gateway-app-cors-reverse-proxy`：规定经领域反向代理返回的响应在允许来源下与直连接口一致的 CORS 暴露行为。

### Modified Capabilities

- （无）`openspec/specs/` 下尚无独立「网关 CORS」基线规格；本变更在 change 内新增能力规格即可。

## Impact

- 代码：`internal/controller/history_route_proxy.go`、`domain_route_proxy.go`（若抽取共用 `ModifyResponse` 或包装器），必要时同步 `voice_route_proxy.go`、`device_route_proxy.go`。
- 行为：仅影响 **命中反向代理** 且 **浏览器带允许 Origin** 的响应头；非浏览器或不在白名单的来源不变。
- 运维：无新环境变量硬性要求；若采用 `ModifyResponse`，需关注 Go 版本与 `httputil.ReverseProxy` 行为一致性。
