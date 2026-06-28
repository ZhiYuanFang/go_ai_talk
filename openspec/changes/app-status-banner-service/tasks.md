# Tasks: app-status-banner-service

## 1. 后端 app-status-service

- [x] 1.1 `cmd/app-status-service/main.go`（无 Redis/MySQL/runtimecheck）
- [x] 1.2 `internal/services/appstatus/banner.go` 内存态 + RWMutex
- [x] 1.3 `api/v1/app_status_*_http.go` + controller 注册
- [x] 1.4 Admin JWT 中间件 + `POST /admin/api/login`
- [x] 1.5 `manifest/config/config.app-status-service.yaml`
- [x] 1.6 Docker：`Dockerfile.app-status-service` + compose 各 overlay + `.env.example`

## 2. 运维静态页

- [x] 2.1 `resource/public/app-status-admin.html`（登录、编辑、预览、快照、立即关闭）
- [x] 2.2 本服务静态路由 `/admin` 与 `/resource/public/*`

## 3. Hub 与文档

- [x] 3.1 `admin-modules.js` 外链 + `admin.html` 支持 `externalUrl`
- [x] 3.2 `.env.example` 增加 `APP_STATUS_*` 占位

## 4. Flutter（flutter_ai_talk）

- [x] 4.1 `AppEnv.statusBaseUrl`（`STATUS_BASE_URL`）
- [x] 4.2 `StatusBannerRepository` + `StatusBannerDismissStore`
- [x] 4.3 `status_banner_prompt.dart`（blocking 仅刷新 / dismissible+不再提示）
- [x] 4.4 `home_screen.dart` postFrame 拉取，优先于版本弹窗
- [x] 4.5 `openspec/changes/app-status-banner-service/specs/app-status-banner/spec.md`

## 5. 手工验收

- [ ] 5.1 本地/测试：`curl GET /app/api/status/banner` active true/false
- [ ] 5.2 Admin 页保存后 App 弹窗；dismissible 不再提示；同文案 re-active 仍抑制
- [x] 5.3 `go build ./...` 与 `dart analyze` 通过
