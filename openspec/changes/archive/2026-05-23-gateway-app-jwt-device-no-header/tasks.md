## 1. JWT 签发与解析（gatewayapp）

- [x] 1.1 扩展 `SignAccess`（或等价封装）：在 JWT 中增加 **`device_no` 私有 claim**；**`sub` 仍为 wx.id 字符串**；未绑设备时 `device_no` 行为按 design 单一策略实现
- [x] 1.2 扩展解析逻辑：验签后同时读出 **`sub`→wxId** 与 **`device_no` claim**；供中间件与 WS 使用
- [x] 1.3 签发→解析自检（遵守仓库「不新增测试文件」约束时，用手工联调或既有可运行入口验证）

## 2. 网关 Bearer 中间件与下游头

- [x] 2.1 修改或替换 Bearer 注入逻辑：设置 **`X-Internal-Wx-Id`**；在非空时设置 **`X-Internal-Device-No`**；**移除**对 `FetchWxUnionIDByID` / `internal/by-id` 的 per-request 依赖
- [x] 2.2 清理或降级 **`gw:app:wxid2union:`** 等仅服务 unionid 注入的缓存与配置项（`config.gateway-app-server.yaml` 注释同步）
- [x] 2.3 全仓检索 **`X-Internal-Wx-Union-Id`**：更新网关、device、文档与 OpenAPI 注释

## 3. 登录与刷新（gateway_app_ctrl + gatewayapp refresh）

- [x] 3.1 `POST /device/app/api/login`：device login 成功后，将 **`deviceNo`** 写入签发 access 的 **`device_no` claim**
- [x] 3.2 `POST /device/app/api/token/refresh`：按 design **D5** 在签发新 access 时 **同步 `device_no` claim**（权威来源单一实现）

## 4. device-service：用户域头与 wx 服务

- [x] 4.1 `DeviceAppUserCtrl`：`bindwx` / `auto_save` / `detail` 改为读取 **`X-Internal-Wx-Id`**；`internal/by-id` 保留为可选/运维，不得再被 gateway 每请求依赖
- [x] 4.2 `internal/services/device/wx.go`：以 **wx 主键 id** 定位行的查询/更新路径；**unionid 仅保留在登录换票写库链路**
- [x] 4.3 导出 **`HeaderInternalWxId`**（或与网关共享常量）避免拼写漂移

## 5. 历史 WebSocket

- [x] 5.1 `gateway_app_history_ws`：用 JWT **`device_no` claim** 与首帧设备号一致性校验；移除 **unionid→detail** 链

## 6. 文档、契约与校验

- [x] 6.1 更新 `api/v1` 中内部头与 dc 注释（强调对外 JSON 不变）
- [x] 6.2 视需要同步旧变更目录中的叙述或归档后全局 spec（本实现未批量改历史归档 change，以免与归档流程冲突）
- [x] 6.3 `openspec validate gateway-app-jwt-device-no-header`；`go build ./...`
