## ADDED Requirements

### Requirement: device-service 事件 logo 持久化到宿主机根目录

在 Docker Compose 部署下，device-service 容器 SHALL 通过 bind mount 将 **`/ai_talk_images`** 映射到 Linux 宿主机同路径 **`/ai_talk_images`**，使 `SaveEventLogo` 写入的文件出现在宿主机上。

#### Scenario: 上传 logo 后宿主机可见

- **WHEN** 管理员通过 API 上传事件 logo 且 device-service 使用默认 `eventImageStorageDir` `/ai_talk_images/`
- **THEN** 宿主机路径 `/ai_talk_images/` 下 SHALL 存在对应文件
- **AND** 容器内同路径 SHALL 可读取该文件

#### Scenario: 容器重建后文件保留

- **WHEN** 宿主机 `/ai_talk_images` 已存在 logo 文件且运维对 device-service 执行 `docker compose up --force-recreate`
- **THEN** 重建后容器 SHALL 仍能读取宿主机挂载目录中的同名文件

### Requirement: gateway-app APK 持久化到宿主机根目录

在 Docker Compose 部署下，gateway-app 容器 SHALL 通过 bind mount 将 **`/apk/ai_talk`** 映射到 Linux 宿主机同路径 **`/apk/ai_talk`**，使版本管理上传的 APK 出现在宿主机上。

#### Scenario: 上传 APK 后宿主机可见

- **WHEN** 管理员通过版本管理接口上传 APK 且 gateway-app 使用默认 `apkStorageDir` `/apk/ai_talk/`
- **THEN** 宿主机路径 `/apk/ai_talk/` 下 SHALL 存在对应 `.apk` 文件

#### Scenario: 容器重建后 APK 保留

- **WHEN** 宿主机 `/apk/ai_talk` 已存在 APK 且运维对 gateway-app 执行 `docker compose up --force-recreate`
- **THEN** 重建后 gateway-app SHALL 仍能通过 `GET /device/app/apk/` 提供该文件

### Requirement: 挂载路径与配置默认一致

Compose 卷挂载点 SHALL 与代码/配置默认存储目录一致：`/ai_talk_images`（device）、`/apk/ai_talk`（gateway-app）。未通过环境变量修改存储路径时，SHALL NOT 要求额外配置即可满足本需求。

#### Scenario: 默认配置下路径一致

- **WHEN** 未设置 `DEVICE_EVENT_IMAGE_STORAGE_DIR` 与 `GATEWAY_APP_APK_STORAGE_DIR`
- **THEN** 写盘路径与 bind mount 目标路径 SHALL 均为上述宿主机根下目录

### Requirement: 部署文档说明宿主机准备

项目 runbook 或等价部署文档 SHALL 说明：Linux Docker 部署前建议执行 `mkdir -p /ai_talk_images /apk/ai_talk`，并给出验证宿主机与容器内文件一致的示例命令。

#### Scenario: 运维可按文档验收

- **WHEN** 运维按文档创建目录并启动 compose 后上传 logo 与 APK
- **THEN** 文档中的 `ls` 或 `docker exec` 验收步骤 SHALL 能确认宿主机目录非空
