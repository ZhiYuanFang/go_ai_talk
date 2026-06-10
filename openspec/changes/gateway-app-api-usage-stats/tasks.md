## 1. API 注册表与路径归一化

- [x] 1.1 新增 `internal/services/gatewayapp/apiregistry`（或等价包）：启动时扫描 `api/v1` 的 `g.Meta`，构建 `method+path模板 → summary` 映射
- [x] 1.2 实现请求路径归一化：动态段匹配模板（如数字 id、wxId），未命中时回退原始 path

## 2. Redis 统计写入

- [x] 2.1 新增 `internal/services/gatewayapp/usagestats`：日桶 INCR、`last_at` 更新、异步写入（失败静默 + warning 日志）
- [x] 2.2 实现跳过规则（WebSocket、internal、admin、静态）与 2xx 门控

## 3. 采集中间件

- [x] 3.1 新增 `installGatewayAppAPIUsageStatsMiddleware`：在 `Next()` 后提取 wxId、归一化 apiKey、触发异步 Redis 写入
- [x] 3.2 在 `RegisterGatewayAppHTTP` 注册中间件（Bearer 之后、反代链可见响应状态的位置）

## 4. 管理读 API

- [x] 4.1 在 `api/v1` 增加 `device_admin_usage_http.go`（或 gateway 专用）：list / detail / user 请求响应类型
- [x] 4.2 实现 `GatewayAppUsageAdminCtrl`（或 handler）：`X-Admin-Password` 校验 + Redis 聚合查询（days 默认 7）
- [x] 4.3 在 `device_route_proxy.go` 排除 `/device/admin/api/usage/*`，确保 gateway 本机处理

## 5. 静态页与入口

- [x] 5.1 新增 `resource/public/api-usage-stats.html`：登录、按 API 列表、按 API 下钻 wxId、按 wxId 查询、7 天默认、返回设备管理链接
- [x] 5.2 在 `gateway_app_register.go` 与 `register.go` 注册 `/device/admin/api-usage-stats` 静态路由
- [x] 5.3 在 `admin.html` 设备记录卡片 `card-actions` 增加「功能使用统计」链接（登录后显示）

## 6. 验证

- [x] 6.1 本地验证：2xx 计入、401 不计入、动态路径归一化、summary 中文展示
- [x] 6.2 验证 list/detail/user 三 API 口令鉴权、默认 7 天聚合、wxId 下钻口径
- [x] 6.3 验证 `/device/admin/api/usage/*` 不经 device-service 反代；`api-usage-stats.html` 与 admin 入口可访问
