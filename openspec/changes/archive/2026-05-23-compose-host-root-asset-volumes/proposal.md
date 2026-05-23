## Why

在 Linux 上用 Docker Compose 部署时，事件 logo（`/ai_talk_images/`）与 App APK（`/apk/ai_talk/`）默认只写在**容器可写层**内，宿主机根目录下看不到文件，且 `recreate` 后易丢失。运维要求在**宿主机**上使用与代码约定一致的根路径持久化资源，便于备份、手工替换与排障。

## What Changes

- 在 `docker-compose.microservices.yml` 为 **device-service** 增加 bind mount：`/ai_talk_images:/ai_talk_images`（与 `device.eventImageStorageDir` 默认一致）。
- 为 **gateway-app** 增加 bind mount：`/apk/ai_talk:/apk/ai_talk`（与 `gatewayApp.apkStorageDir` 默认一致）。
- 在 `manifest/docker/.env.example` 与 `docs/runbooks/release-deploy-and-run.md` 补充：宿主机预创建目录、权限说明、验收命令（`ls` / `docker exec`）。
- **不修改** 应用写盘路径与 HTTP path 契约；不将 logo 挂到 gateway-app，不将 APK 挂到 device-service。
- 可选文档：SELinux 环境下卷后缀 `:z` 的说明（设计阶段定是否写入 compose 注释）。

## Capabilities

### New Capabilities

- `compose-host-root-asset-volumes`：Compose 将事件 logo 与 APK 存储 bind 到 Linux 宿主机 `/ai_talk_images` 与 `/apk/ai_talk`，并与现有 path-only 存库语义一致。

### Modified Capabilities

（无。`openspec/specs/` 无既有 compose 卷规格；HTTP/API 行为不变。）

## Impact

- **部署**：`manifest/docker/docker-compose.microservices.yml`、`manifest/docker/.env.example`
- **文档**：`docs/runbooks/release-deploy-and-run.md`、可选 `README.MD` 静态资源路径一句
- **运行时**：仅 **device-service**、**gateway-app** 容器需重建；gateway（:9701）反代 `/ai_talk_images` 仍指向 device-service，无需为写盘挂卷
- **运维**：首次部署前建议在宿主机 `mkdir -p /ai_talk_images /apk/ai_talk`
