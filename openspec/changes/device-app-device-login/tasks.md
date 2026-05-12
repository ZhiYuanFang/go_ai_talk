## 1. API 与契约

- [x] 1.1 在 `api/v1/device_app_user_http.go` 增加 **`DeviceWxDeviceLoginReq/Res`**（path：`POST /device/app/api/user/device_login`）
- [x] 1.2 在 `api/v1/gateway_app_http.go` 增加 **`GatewayAppDeviceLoginReq/Res`**（path：`POST /device/app/api/device_login`），响应字段与聚合微信登录对齐

## 2. device-service 实现

- [x] 2.1 `DeviceAppUserCtrl.DeviceLogin`：校验 **`deviceNo`** → **`DeviceAdmin().EnsureRegistered` 或等价存在性查询** → 查 **`wx` 表 `device_no`** 得 **`wxId`**；组装响应
- [x] 2.2 `internal/services/device`：抽取小函数 **`WxByDeviceNo`**（或内联 DAO）避免 controller 直连多表散落逻辑；中文注释说明失败语义

## 3. gateway-app 实现

- [x] 3.1 `GatewayAppCtrl.DeviceLogin`：`gclient` 调 device **`/device/app/api/user/device_login`** → **`SignAccess(wxId, deviceNo)`** + **`IssueRefreshToken`**
- [x] 3.2 `gateway_app_auth_exempt.go`：为 **`POST /device/app/api/device_login`** 与 **`POST /device/app/api/user/device_login`**（若经网关路径暴露）配置白名单，与现登录类路径一致

## 4. 联调页

- [x] 4.1 `gateway-app-integration-test.html`：新增「设备号登录」按钮与日志；请求 **`POST {base}/device/app/api/device_login`**；成功时写入页面 **`accessToken`/`refreshToken`** 文本框（与微信登录区复用或并列说明）

## 5. 校验

- [x] 5.1 `openspec validate device-app-device-login`
- [x] 5.2 `go build ./...`
