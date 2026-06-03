## 1. 数据与领域

- [x] 1.1 DDL 与 `entity.User` / `dao.User` 列 `last_api_path`、`last_api_at`
- [x] 1.2 `TouchLastAPIAccess`、`ListUsersPage`（`internal/services/device`）
- [x] 1.3 `DeviceAdminContract` 与 `admin_http_client` 委派

## 2. HTTP API

- [x] 2.1 `device_admin_http.go`、`device_internal_full_http.go` 类型定义
- [x] 2.2 `device_admin.go`、`device_internal_handlers.go` 控制器

## 3. 网关 touch

- [x] 3.1 共享 `ExtractDeviceNo` / touch 辅助包
- [x] 3.2 主网关与 gateway-app 安装 middleware

## 4. 前端

- [x] 4.1 `admin.html` 设备记录卡片、分页、搜索、历史链接

## 5. 验收

- [x] 5.1 `go build ./...` 相关包通过
