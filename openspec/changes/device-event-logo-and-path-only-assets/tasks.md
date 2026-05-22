## 1. device-service 事件 logo/color

- [x] 1.1 配置 `eventImageStorageDir`（默认 `/ai_talk_images/`）与保存/校验工具（扩展名、大小、安全文件名）
- [x] 1.2 `AddEvent`/`UpdateEvent` 支持 `color`；multipart handler 处理 add/update（含 logo 落盘与 path 写库）
- [x] 1.3 `ListEvents`、`RebuildEventCache`、`cache_repo` 投影增加 `Logo`、`Color`
- [x] 1.4 注册 `GET /ai_talk_images/*` 静态读；扩展 `DeviceAdminContract` 与 `admin_http_client`（若需）

## 2. gateway-app path-only 与反代

- [x] 2.1 APK 上传写库与响应改为 path-only；移除写库对 `publicBaseUrl` 的硬依赖
- [x] 2.2 `VersionCheck` 与缓存返回 path；实现历史绝对 URL 归一化
- [x] 2.3 注册 `/ai_talk_images/*` 反代至 device-service

## 3. API 与管理端

- [x] 3.1 调整 `api/v1/device_admin_http.go` 文档/路由说明（multipart）；`device_admin.go` 接入新 handler
- [x] 3.2 `admin.html` 事件表单 FormData（logo、color）与列表预览（9702 基址 + path）

## 4. 文档与验证

- [x] 4.1 更新 `README.MD`：事件 logo/color、path 拼接、APK path-only **BREAKING**
- [x] 4.2 `go build ./...` 通过
- [x] 4.2 对照 spec 自检列表字段与静态路径
