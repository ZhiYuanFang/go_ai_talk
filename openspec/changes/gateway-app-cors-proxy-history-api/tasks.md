## 1. 抽取与复用

- [ ] 1.1 将「白名单命中时写入的 `Access-Control-*` 头集合」抽取为可被 **中间件** 与 **ReverseProxy.ModifyResponse** 共用的函数（或置于 `internal/services/gatewayapp` 的纯写头辅助，避免循环 import），与现有 `gateway_app_cors.go` 行为对齐。

## 2. History 代理

- [ ] 2.1 在 `history_route_proxy.go` 使用的 `ReverseProxy` 上增加 `ModifyResponse`（或经评审的等价钩子）：根据原始请求 `Origin` 调用 `ReflectGatewayAppCORSOrigin`，命中则对下游 `*http.Response` 写入 CORS 头；`ModifyResponse` 错误路径不得引入无谓 502。
- [ ] 2.2 本地用 **带 `Origin` 头** 的 `curl` 或浏览器验证：`GET /device/history/api/list` + `Authorization`，确认响应含 `Access-Control-Allow-Origin`；并用非白名单 `Origin` 确认不出现误加。

## 3. Voice/Device（可选一致化）

- [ ] 3.1 若 `buildReverseProxy` 已扩展为统一注入 CORS，则核对 `voice_route_proxy.go`、`device_route_proxy.go` 是否自动受益；若未共用，则按与 history 相同语义补挂，避免行为分裂。

## 4. 文档

- [ ] 4.1 在 `docs/runbooks` 或变更说明中补充：`curl` 调试 CORS 时必须加 `-H "Origin: ..."`，否则无法模拟浏览器。
