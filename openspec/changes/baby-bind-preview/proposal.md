## Why

Flutter 全家共享改为「扫码 → 确认宝宝名 → 再绑定」。绑定前需要只读预览昵称；现有 `GET /device/app/api/user/get` 虽可按 `deviceNo` 读画像，但字段偏全、语义是通用查档，不宜作为绑定确认契约。需新增专用预览接口，供姊妹仓 `flutter_ai_talk` 的扫码绑定流程调用。

## What Changes

- 新增 App：`POST /device/app/api/user/bind_preview`，body `{ "deviceNo": "<string>" }`，需登录会话（`X-Internal-Wx-Id` / JWT 与 `bindwx` 同级）。
- 成功 `data` 仅含 `deviceNo`（回显）与 `babyName`（可空字符串）；**不得**作为副作用写入绑定关系。
- 空 / 未注册 `deviceNo` → 业务错误；**不得**冒充成功空档。
- **不**改 `bindwx` / `auto_save` / `user/get` 既有行为。
- gateway-app 须可转发该路径（非 auth-exempt）。

## Capabilities

### New Capabilities

- `baby-bind-preview-api`：绑定前只读预览宝宝昵称的 App HTTP 契约。

### Modified Capabilities

（无）

## Impact

- 代码：`api/v1/device_app_user_http.go`、`DeviceAppUserCtrl`、`internal/services/device`（复用 `DeviceProfile().GetProfile` / `EnsureRegistered` 校验存在性，禁止写 `wx.device_no`）。
- 进程：`device-service`（必）；`gateway-app-server`（若代理/白名单需改则一并）。
- **姊妹提案（后做）**：`d:\work\flutter_ai_talk\openspec\changes\baby-share-qr-bind\`（扫码/相册、确认框、隐藏 ID UI）。
- 顺序：本变更 apply/部署完成后，再通知 Flutter `/opsx-apply`。
