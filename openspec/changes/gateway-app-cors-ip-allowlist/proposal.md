## Why

Web 前端在局域网或公网 IP 下联调 `gateway-app-server` 时，浏览器同源策略会拦截跨域请求；当前网关未返回 CORS 响应头，导致联调无法完成。需要在 App 网关上增加受控的 CORS 行为，先覆盖两台固定主机（任意端口）的前端来源。

## What Changes

- 在 `gateway-app-server` HTTP 层增加 CORS 处理：对预检（OPTIONS）与常规响应补充 `Access-Control-*` 头。
- 允许的来源为 **精确主机** `192.168.0.131` 与 `120.55.50.105`，**任意合法端口**（即 `http://<host>:<port>` 与 `https://<host>:<port>` 在主机匹配时回显该请求的 `Origin`）。
- 允许方法：`GET`、`POST`、`OPTIONS`。
- 允许请求头至少包含：`Content-Type`、`Authorization`（及浏览器预检可能携带的常用头按需放行）。
- 不修改既有 Bearer 鉴权语义；OPTIONS 继续走现有豁免逻辑（与 CORS 中间件顺序在设计中明确）。

## Capabilities

### New Capabilities

- `gateway-app-cors`：描述 App 网关在联调场景下的 CORS 允许主机、方法、请求头及预检行为，便于后续扩展白名单或改为配置驱动。

### Modified Capabilities

- （无）现有 `openspec/specs/` 中无与本变更直接对应的网关 CORS 规格，本次以新增能力规格为主。

## Impact

- 代码：`cmd/gateway-app-server`、`internal/controller`（注册与跨切面中间件）、可选 `internal/services/gatewayapp` 若将匹配逻辑下沉。
- 行为：仅当 `Origin` 解析后的 host 属于允许列表时回写 `Access-Control-Allow-Origin`；不匹配则不添加该头（浏览器保持默认跨域失败），避免全回显。
- 运维：若未来需增删 IP，应通过配置或代码常量调整并更新规格。
