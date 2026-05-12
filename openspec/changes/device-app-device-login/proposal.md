## Why

联调与部分场景需要 **仅凭已注册 `device_no` 完成 App 会话**（不经过微信 `jsCode` 换票），以便快速验证 token、Bearer 注入与下游接口。当前仅支持 **微信登录链路**，缺少「设备号即凭证」的入口。

## What Changes

- **device-service**：新增 **`POST /device/app/api/user/device_login`**，请求体仅 **`deviceNo`**；校验设备在设备域注册表中存在，且 **`wx` 表已绑定该 `device_no`**（存在对应 wx 行）；成功则返回与微信业务登录对齐的字段 **`wxId`、`deviceNo`、`isNewUser`**（**不**签发 JWT；`isNewUser` 恒为 `false` 或省略由实现固定）。若设备不存在或未绑定 wx，返回明确业务错误。
- **gateway-app-server**：新增 **`POST /device/app/api/device_login`**（与现有 **`/device/app/api/login`** 区分），内部 HTTP 调用 device 的 **`/device/app/api/user/device_login`** 获取业务字段后，按现网规则签发 **`access_token`（JWT，含 `sub`/`device_no` claim）** 与 **`refresh_token`**；将该路径加入 **Bearer 白名单**。
- **联调页**：在 **`resource/public/gateway-app-integration-test.html`** 增加「设备号登录」区块，调用网关 **`POST /device/app/api/device_login`**，展示返回的 token 与业务字段，便于回测。

**非 BREAKING**：现有微信登录接口与字段保持不变；新增路径为增量能力。

## Capabilities

### New Capabilities

- `device-app-device-login`：device 设备号登录契约、网关聚合登录、联调页调试入口的规格说明。

### Modified Capabilities

- （无全局 `openspec/specs/` 必改项；归档时可按需合并到设备/App 网关相关能力。）

## Impact

- `api/v1/device_app_user_http.go`、`internal/controller/device_app_user.go`、`internal/services/device`（设备存在性 + wx 绑定查询）。
- `api/v1/gateway_app_http.go`（若单独定义聚合请求）、`internal/controller/gateway_app_ctrl.go`、`internal/controller/gateway_app_auth_exempt.go`（白名单）。
- `resource/public/gateway-app-integration-test.html`。
