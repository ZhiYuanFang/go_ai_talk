## 1. App 删除 API（本仓先做）

- [x] 1.1 `api/v1`：新增 `DeviceAppPredictProIdeaDeleteReq/Res`，`POST /device/app/api/predict-pro-ideas/delete`，body `id`
- [x] 1.2 Service：作者校验（`wx_id`）后软删（复用 `deleted_at` 语义）；非作者/已删/不存在返回业务错误
- [x] 1.3 Controller 注册 handler；确认非 auth-exempt
- [x] 1.4 核对 gateway-app 对 `/device/app/api/predict-pro-ideas/delete` 可转发；按需更新 usage/proxy 约定

## 2. 验收与交接

- [x] 2.1 用作者 JWT curl：删本人成功；删他人失败；list 不再含该 id
- [x] 2.2 部署说明：列出需重启进程（预期 `device-service`，必要时 `gateway-app-server`）
- [x] 2.3 通知姊妹仓 Flutter：`d:\work\flutter_ai_talk\openspec\changes\predict-pro-idea-author-delete` 可 `/opsx-apply`
