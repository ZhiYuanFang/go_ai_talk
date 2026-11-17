## 1. bind_preview API（本仓先做）

- [x] 1.1 `api/v1`：新增 `DeviceProfileBindPreviewReq/Res`，`POST /device/app/api/user/bind_preview`，body `deviceNo`；成功 data 含 `deviceNo`、`babyName`
- [x] 1.2 Service：校验登录 wx + trim `deviceNo`；未注册 → 业务错误；已注册则读画像返回昵称；**禁止**调用 `WxBindDevice` 或写 `wx.device_no`
- [x] 1.3 `DeviceAppUserCtrl` 注册 handler；确认非 auth-exempt；中文注释说明「绑定前只读预览」
- [x] 1.4 核对 gateway-app 对 `/device/app/api/user/bind_preview` 可转发；按需更新 usage/proxy 约定

## 2. 验收与交接

- [x] 2.1 用已登录 JWT curl：已注册 dn 返回昵称且 wx 未改绑；空 dn / 未注册 dn 失败；无鉴权失败
- [x] 2.2 部署说明：列出需重启进程（预期 `device-service`，必要时 `gateway-app-server`）
- [x] 2.3 通知姊妹仓 Flutter：`d:\work\flutter_ai_talk\openspec\changes\baby-share-qr-bind` 可 `/opsx-apply`

### 实现摘记

- **代码**：`api/v1/device_app_user_http.go`（Req/Res）、`device.WxBindPreview`、`DeviceAppUserCtrl.BindPreview`（映射空 dn / 未注册 / 无效 wx 为参数或业务错误）。
- **Gateway**：`/device/app/api/user/*` 反代已覆盖；`bind_preview` **不在** auth-exempt；usage **不**加入 maintenance skip（与 bindwx 同属用户意图操作，计入统计）。
- **部署**：重启 **`device-service`**；若边缘仅靠反代且无本地 DeviceAppUserCtrl，通常 **不必**改 gateway 代码，但滚动发布后需确认边缘已指向含本变更的 device-service。若 gateway-app 与 device 同发版窗口，可一并重启 **`gateway-app-server`**。
- **curl 配方**（部署后手工）：
  ```bash
  # 成功（替换 BASE / TOKEN / DEVICE_NO）
  curl -sS -X POST "$BASE/device/app/api/user/bind_preview" \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"deviceNo\":\"$DEVICE_NO\"}"
  # 空 dn / 未注册 / 无 Authorization → 非成功信封
  ```
- **Flutter**：Go 部署完成后可对 `baby-share-qr-bind` 执行 `/opsx-apply`。
