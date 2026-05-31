## Why

当前 device 用户域提供登录、绑定与画像保存能力，但缺少账号注销能力，导致用户无法在设备域完成账号生命周期闭环。现在补齐该能力，可以满足合规与用户自助注销诉求，并减少人工介入。

## What Changes

- 新增 device-service 接口：`POST /device/app/api/user/deactivate`。
- 接口从请求头读取 `X-Internal-Wx-Id`，定位 `wx` 表记录。
- 执行注销语义：删除 `wx` 表中对应 `id` 的单条记录。
- 返回统一成功响应；当记录不存在时返回明确业务错误语义。
- 保持服务边界不变：仅在 device 域内处理 `wx` 表，不新增跨服务数据库访问。

## Capabilities

### New Capabilities
- （无）

### Modified Capabilities
- `device-wx-profile-apis`：补充“账号注销”能力要求，定义 `/device/app/api/user/deactivate` 的入参来源、删除语义、错误语义与幂等行为。

## Impact

- 受影响 API：device app 用户域新增 `POST /device/app/api/user/deactivate`。
- 受影响代码：`api/v1/device_app_user_http.go`、`internal/controller/device_app_user.go`、`internal/services/device/wx.go`。
- 受影响数据：`ai_voice_device.wx` 表按主键删除单行。
- 兼容性：新增接口，非破坏性；仅在调用新接口时产生删除行为。
