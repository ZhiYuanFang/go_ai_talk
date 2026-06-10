## 1. ucg-service 内部上传

- [x] 1.1 `UploadMediaObjectWithPrefix` 支持 `event/` 前缀与图片校验
- [x] 1.2 注册 `POST /ucg/internal/api/media/upload` + 内部密钥中间件

## 2. device-service OSS 集成

- [x] 2.1 HTTP 客户端调 ucg internal upload（`UCG_SERVICE_BASE_URL`）
- [x] 2.2 `saveEventLogoFromRequest` 改 OSS 上传；删除本地落盘逻辑
- [x] 2.3 `EventLogoCdnURL` + HTTP 层事件列表 logo 映射为 CDN URL
- [x] 2.4 移除 `deviceEventImageServe` 与 `/ai_talk_images` 路由注册

## 3. gateway 与 site

- [x] 3.1 移除 `eventImageProxy` 与 `/ai_talk_images/` 鉴权白名单
- [x] 3.2 `gateway_app_site` logoUrl 改用 CDN 映射

## 4. 管理页与配置

- [x] 4.1 更新 `admin.html` logo 预览文案与 `eventLogoUrl`
- [x] 4.2 移除 `device.eventImageStorageDir` 配置与 Docker volume 挂载
- [x] 4.3 更新 README 相关段落

## 5. 迁移工具与文档

- [x] 5.1 新增 `hack/migrate-event-logos-to-oss/main.go`
- [x] 5.2 新增 `docs/runbooks/event-logo-oss-migration.md`

## 6. flutter_ai_talk

- [x] 6.1 更新 `event_catalog_paths` / README（CDN 为 logo 唯一来源）
- [x] 6.2 确认 `EventLogo` 与冷启动下载 CDN 路径可用

## 7. 验证

- [x] 7.1 `go build ./...` 通过
