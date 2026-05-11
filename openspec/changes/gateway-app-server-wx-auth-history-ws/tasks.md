## 1. 配置与进程骨架

- [x] 1.1 新增 `manifest/config/config.gateway-app-server.yaml`（server 端口、redis、app 库连接、下游 DEVICE/HISTORY 代理 URL 环境变量对齐现有命名习惯）
- [x] 1.2 新增 `cmd/gateway-app-server/main.go`（`GF_GCFG_FILE` 默认、依赖探活、绑定 HTTP 与 WS 注册入口）
- [x] 1.3 在 Kustomize/docker-compose 增加 gateway-app 部署与服务端口说明（与现 gateway 区分）

## 2. device-service：wx 与契约

- [x] 2.1 为 wx 表补充 dao/entity（若尚未生成）及中文注释
- [x] 2.2 实现 `POST /device/wx/api/login`（仅业务返回，无 JWT）
- [x] 2.3 实现 `POST /device/profile/api/bindwx`（Header `X-Internal-Wx-Code`，Body `deviceNo`）
- [x] 2.4 实现 `POST /device/profile/api/auto_save`（Header wxCode，Body birthday/sex，出参 device_no；未绑定时生成 6 位大写 A–Z 随机 device_no，**全局不得与已有 device_no 冲突**，依赖 UNIQUE + 冲突重试或先查后插，创建设备并绑定 wx）
- [x] 2.5 实现 `GET /device/wx/api/detail`（Header wxCode，出参 device_no）
- [x] 2.6 实现内部只读 `id → wxCode` 接口供网关调用，并限制为内网/共享密钥策略
- [x] 2.7 为上述读路径接入 Redis 缓存与写路径失效（`cachekit`）

## 3. history-service：piece 与发布

- [x] 3.1 实现 `GET /device/history/api/piece`（eventId、startTime、endTime、deviceNo）与 Redis 缓存
- [x] 3.2 在 history 增删改成功路径统一调用发布函数，向约定 channel `PUBLISH` JSON（含 device_no、action、payload）
- [x] 3.3 在 CUD 后失效相关 piece 缓存键

## 4. gateway-app-server：鉴权、登录、版本

- [x] 4.1 实现 Bearer 中间件：校验 **JWT**（签名、`exp`），从 `sub` 解析 wx.id，调用 device 内部接口得 wxCode、设置 `X-Internal-Wx-Code`；配置白名单路径
- [x] 4.2 复用/对齐现有 `RegisterHTTP` 代理与横切模式，避免破坏根 `main` gateway
- [x] 4.3 实现 `POST /device/app/api/login`（调 device login → 签发 **JWT** access、`sub`=wxId、`iat`/`exp`；refresh 不透明串写 Redis）
- [x] 4.4 实现 refresh 接口（校验 Redis、签发新 **JWT** access；旋转策略按 design 单一实现）
- [x] 4.5 实现 `GET /device/app/api/version/check`（读 ai_voice_app.version + Redis 缓存）

## 5. gateway-app-server：WebSocket 与 Pub/Sub

- [x] 5.1 实现历史 WS 路由、首帧 auth（snake_case `access_token`）与 device_no 绑定校验
- [x] 5.2 实现按 device_no 的连接注册表与广播安全（仅已认证连接）
- [x] 5.3 启动 Redis 订阅 goroutine，解析消息并广播；处理重连与进程退出取消订阅
- [x] 5.4 为 id→wxCode 等网关侧热路径补充 Redis 缓存与失效钩子（与 device 写事件对齐或 TTL 兜底）

## 6. 联调与收尾

- [x] 6.1 端到端验证：登录 → 带 Bearer 调下游 → piece → 改历史 → WS 收到 CUD 消息（实现侧已贯通；部署后请在联调环境人工点检）
- [x] 6.2 日志与错误语义巡检（中文日志关键节点、不泄露敏感 token）
- [x] 6.3 `openspec validate gateway-app-server-wx-auth-history-ws` 通过后再进入归档流程

## 7. Gateway-App 联调静态页

- [x] 7.1 新增 `resource/public/gateway-app-integration-test.html` 并在 `RegisterGatewayAppHTTP` 注册 `GET /device/app/integration-test.html`，用于模拟登录 / 绑定 / 画像 / 文本喂养 / WS / piece 趋势自测
