## Why

移动端需要独立入口进程（gateway-app-server）：在保持与现有 gateway 相同的领域代理与静态资源能力的前提下，统一完成基于 `wx` 表主键的鉴权、令牌签发与刷新、版本检查，以及通过 WebSocket 将历史记录的增删改实时推送到 App。微信侧绑定与登录业务落在 device-service，历史区间查询与变更通知落在 history-service，并通过 Redis 缓存与 Pub/Sub 与项目现有风格对齐。

## What Changes

- 新增 **`gateway-app-server`** 可执行入口与独立配置（含 **`ai_voice_app`** 库、`version` 表），默认行为与现有 gateway 对齐：静态资源、history/voice/device 的条件反向代理、横切中间件；在此基础上增加 **Bearer 鉴权**（**access_token 为纯 JWT**，载荷 `sub` 对应 `wx.id`）、解析后调用 device 获取 **`unionid`（`wx.union_id`）** 并为下游请求附加 **`X-Internal-Wx-Union-Id`**；鉴权白名单覆盖登录、刷新、device 微信登录等无 Bearer 路径。
- **gateway-app** 暴露 **`POST /device/app/api/login`**：内部调用 device 的 **`POST /device/app/api/user/login`**（Body **`jsCode`/`platform`**，仅业务字段），再在网关签发 **access_token（纯 JWT）/ refresh_token（不透明 + Redis）**（刷新逻辑与 Redis 会话仅存于 gateway-app）。
- **device-service**：App 用户域接口路径统一在 **`/device/app/api/user/*`**（如 get/save/bindwx/auto_save、login、detail、internal/by-id 等）；**`POST /device/app/api/user/login`** 使用服务端 **jscode2session** 换票后以 **unionid** 落库/匹配 wx 行，仅返回业务结果（如 wxId、device_no、是否新用户等），**不**签发 JWT，**不**向客户端回传 unionid/openid。
- **history-service**：新增 **`GET /device/history/api/piece`**（按 eventId、时间区间、deviceNo 返回记录）；在历史 **增删改** 成功后 **Redis PUBLISH** 通知网关侧 WS 下发（payload 含操作类型与消息体字段）。
- **gateway-app** 新增 **历史 WebSocket**：握手后首条文本帧 JSON 使用 snake_case 的 `access_token`（值为 **JWT**）；校验 **JWT 与 device_no 绑定关系** 后订阅；后台订阅 Redis 频道并向对应 `device_no` 连接广播。
- **Redis**：**id→unionid**（网关侧）、版本检查、`piece` 查询结果、refresh 旋转等按设计使用 **`internal/platform/cachekit`** 或同风格封装；Pub/Sub 订阅可使用 `g.Redis()` 独立连接（与 KV 同集群配置）。
- 部署侧新增 **gateway-app** 镜像/Deployment/Service（端口与现网关区分），并配置指向 device/history 的下游 URL 环境变量。

## Capabilities

### New Capabilities

- `gateway-app-server`：App 网关进程、鉴权中间件、**`X-Internal-Wx-Union-Id`** 注入、登录与刷新、版本检查 API、历史 WS Hub、Redis Pub/Sub 订阅与相关缓存键语义。
- `device-wx-profile-apis`：device 上 wx 绑定、画像自动保存、wx 详情、**jscode2session + unionid** 微信登录（仅业务）、按主键 id 解析 **unionid** 等契约与安全边界（内部头仅可信网络）。
- `history-piece-and-realtime-notify`：`piece` 区间查询 API、历史 CUD 后的 Redis 发布格式与网关消费语义。

### Modified Capabilities

- 无（本变更以新增能力与部署单元为主；与现有 gateway 进程并存，不修改已归档全局 spec 中的既有条款）。

## Impact

- 新增代码入口：`cmd/gateway-app-server`、网关侧登录/刷新/版本/WS/订阅等控制器或服务封装；复用 `internal/controller` 中与 gateway 共用的代理与注册模式时需避免破坏现有 `main` gateway。
- device-service：`api/v1`、控制器、`internal/services/device` 与 dao/model（wx 表）扩展；不得跨库访问 history。
- history-service：`piece` 与 CUD 钩子中的发布逻辑；保持 history 库表所有权不变。
- 配置与编排：`manifest/config` 新增 gateway-app 专用 yaml；Kustomize/compose 增加服务与 Redis/下游 URL 环境变量。
- 运维：Redis Cluster 下 Pub/Sub 客户端行为需在实现阶段验证（订阅连接、重连）。
