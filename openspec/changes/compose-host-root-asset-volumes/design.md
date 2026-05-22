## Context

- 事件 logo：由 **device-service** 写入 `EventImageStorageDir()`，默认 **`/ai_talk_images/`**；库字段 `event.logo` 为 path-only（如 `/ai_talk_images/event_1_xxx.png`）。
- App APK：由 **gateway-app-server** 写入 `ApkStorageDir()`，默认 **`/apk/ai_talk/`**；`version.download_url` 为 `/device/app/apk/<file>.apk`。
- 当前 Compose **无 volumes**；容器内路径虽为根下目录，但与宿主机隔离。
- 用户环境：**Linux 宿主机 + Docker**；要求宿主机上可见 **`/ai_talk_images`** 与 **`/apk/ai_talk`**。

## Goals / Non-Goals

**Goals:**

- `docker compose up` 后，上传 logo/APK 即出现在宿主机对应目录。
- 容器重建（`--force-recreate`）后文件仍保留在宿主机。
- 配置路径与现有 yaml/环境变量默认一致，无需改业务代码。

**Non-Goals:**

- 不改存储路径到项目子目录（如 `./data/...`），除非后续单独提案。
- 不为 gateway（:9701）挂 APK 卷（APK 不在该进程写盘）。
- 不迁移历史容器层内已有文件（文档说明需手工 `docker cp` 或重新上传）。

## Decisions

### 1. Bind mount 映射

```yaml
device-service:
  volumes:
    - /ai_talk_images:/ai_talk_images

gateway-app:
  volumes:
    - /apk/ai_talk:/apk/ai_talk
```

容器内路径与配置默认相同，**不**设置额外的 `DEVICE_EVENT_IMAGE_STORAGE_DIR` / `GATEWAY_APP_APK_STORAGE_DIR`，除非运维刻意覆盖（覆盖时挂载点须与变量一致）。

**备选**：挂到 `./data/ai_talk_images` — 拒绝，与用户「宿主机根目录」明确要求不符。

### 2. 服务边界

| 服务 | 卷 | 原因 |
|------|-----|------|
| device-service | `/ai_talk_images` | 唯一写 event logo 的进程 |
| gateway-app | `/apk/ai_talk` | 唯一写 APK 的进程 |
| gateway | 无 | 仅反代 `/ai_talk_images` 到 device-service |
| voice/history/worker | 无 | 不涉及这两类静态资源 |

### 3. 宿主机初始化

部署 runbook 建议（实现时写入文档）：

```bash
sudo mkdir -p /ai_talk_images /apk/ai_talk
sudo chmod 755 /ai_talk_images /apk/ai_talk
```

当前镜像未声明非 root `USER`，容器内 root 写 bind mount 一般可直接在宿主机创建文件。若日后镜像降权，需同步 `chown` 文档。

### 4. SELinux（RHEL/CentOS 等）

默认 compose 使用简单 bind。若遇 Permission denied，文档说明可尝试：

```yaml
- /ai_talk_images:/ai_talk_images:Z
```

不在首版强制加 `:Z`，避免影响无 SELinux 环境。

### 5. 与 HTTP 访问的关系

- Logo：`https://<host>:9701/ai_talk_images/...` → gateway 反代 → device-service 读**同一挂载目录**。
- APK：`https://<host>:9702/device/app/apk/...` → gateway-app 读本机 **`/apk/ai_talk`** 挂载。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 宿主机根目录文件属主为 root | 文档说明备份与 `chown`；运维知情 |
| 旧容器内文件未自动迁移 | runbook 写 `docker cp` 或重新上传 |
| 路径被误删 | 依赖宿主机备份策略；DB path 仍在 |

## Migration Plan

1. 在宿主机创建目录（见上）。
2. 更新 compose 后：`docker compose ... up -d --force-recreate device-service gateway-app`。
3. 验证上传与 `ls` 宿主机目录；已有 DB path 但无文件时需补传。

## Open Questions

- 无。
