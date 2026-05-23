## 1. 网关聚合

- [x] 1.1 在 `GatewayAppCtrl.DeviceLogin` 中，在解析 `data.deviceNo` 之后：若 `deviceNo == ""` 且 `strings.TrimSpace(req.DeviceNo) != ""`，则将 `deviceNo` 设为 trim 后的 `req.DeviceNo` 再进入签发与 `GatewayAppDeviceLoginRes` 赋值；若仍为空则保持现有错误返回。
- [x] 1.2 核对 `GatewayAppDeviceLoginRes` 的 `json:"deviceNo"` 与统一响应封装下客户端可见性（无字段遗漏）。

## 2. device 纯业务（复核）

- [x] 2.1 复核 `DeviceAppUserCtrl.DeviceLogin` → `WxDeviceLoginByDeviceNo` 是否在所有成功分支写入非空 `DeviceNo`；若有遗漏分支则补齐。

## 3. 验证

- [x] 3.1 手工调用网关 `POST /device/app/api/device_login` 与直连 device `POST /device/app/api/user/device_login`，确认成功体 `data.deviceNo` 非空且与预期一致。
