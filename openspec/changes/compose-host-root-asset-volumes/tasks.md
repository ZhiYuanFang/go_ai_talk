## 1. Compose 卷配置

- [x] 1.1 在 `docker-compose.microservices.yml` 的 `device-service` 增加 `volumes: - /ai_talk_images:/ai_talk_images`
- [x] 1.2 在 `gateway-app` 增加 `volumes: - /apk/ai_talk:/apk/ai_talk`
- [x] 1.3 在 compose 文件顶部或对应 service 旁添加中文注释：宿主机根目录持久化、与默认配置路径一致

## 2. 部署文档

- [x] 2.1 更新 `docs/runbooks/release-deploy-and-run.md`：宿主机 `mkdir`、权限、重建服务名、验收 `ls` / `docker exec`
- [x] 2.2 更新 `manifest/docker/.env.example` 简短说明（可选 `DEVICE_EVENT_IMAGE_STORAGE_DIR` / `GATEWAY_APP_APK_STORAGE_DIR` 须与挂载点一致）
- [x] 2.3 视需要更新 `README.MD` 事件 logo / APK 静态资源一句（宿主机路径 + Docker 卷）

## 3. 验收

- [x] 3.1 在 Linux Docker 环境：`up -d` 后上传事件 logo，确认宿主机 `/ai_talk_images` 有文件且 :9701 图片可访问
- [x] 3.2 上传 APK，确认宿主机 `/apk/ai_talk` 有文件且 :9702 下载可访问
- [x] 3.3 `--force-recreate device-service gateway-app` 后文件仍在宿主机且 HTTP 仍可访问
