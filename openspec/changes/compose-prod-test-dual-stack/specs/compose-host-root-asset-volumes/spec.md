## ADDED Requirements

### Requirement: 测试栈事件 logo 持久化到独立宿主机目录

在 Docker Compose **测试** overlay 部署下，device-service 容器 SHALL 通过 bind mount 将容器内 **`/ai_talk_images`** 映射到 Linux 宿主机 **`/ai_talk_images_test`**，与生产目录 `/ai_talk_images` 隔离。

#### Scenario: 测试上传 logo 不写入生产目录

- **WHEN** 管理员在测试环境上传事件 logo 且 device-service 使用默认 `eventImageStorageDir` `/ai_talk_images/`
- **THEN** 文件 SHALL 出现在宿主机 `/ai_talk_images_test/` 下
- **AND** 宿主机 `/ai_talk_images/`（生产）SHALL NOT 因该上传而新增或修改同名文件

### Requirement: 测试栈 APK 持久化到独立宿主机目录

在 Docker Compose **测试** overlay 部署下，gateway-app 容器 SHALL 通过 bind mount 将容器内 **`/apk/ai_talk`** 映射到 Linux 宿主机 **`/apk/ai_talk_test`**，与生产目录 `/apk/ai_talk` 隔离。

#### Scenario: 测试上传 APK 不写入生产目录

- **WHEN** 管理员在测试环境版本管理页上传 APK
- **THEN** 文件 SHALL 出现在宿主机 `/apk/ai_talk_test/` 下
- **AND** 宿主机 `/apk/ai_talk/`（生产）SHALL NOT 因该上传而新增或修改同名文件

## MODIFIED Requirements

### Requirement: 部署文档说明宿主机准备

项目 runbook 或等价部署文档 SHALL 说明：Linux Docker **生产**部署前建议执行 `mkdir -p /ai_talk_images /apk/ai_talk`；**测试**部署前建议执行 `mkdir -p /ai_talk_images_test /apk/ai_talk_test`；并给出验证宿主机与容器内文件一致的示例命令。

#### Scenario: 运维可按文档验收生产静态目录

- **WHEN** 运维按文档创建生产目录并启动 prod compose 后上传 logo 与 APK
- **THEN** 文档中的 `ls` 或 `docker exec` 验收步骤 SHALL 能确认宿主机 `/ai_talk_images` 与 `/apk/ai_talk` 非空

#### Scenario: 运维可按文档验收测试静态目录

- **WHEN** 运维按文档创建测试目录并启动 test compose 后上传 logo 与 APK
- **THEN** 文档中的验收步骤 SHALL 能确认宿主机 `/ai_talk_images_test` 与 `/apk/ai_talk_test` 非空
