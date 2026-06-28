## Context

`app-status-banner-service` 变更已实现独立通知微服务，命名暂用 `app-status-service`。CI `docker-acr.yml` 的 `ALL_SERVICES` 未包含该服务，无法 `+notify` 或全量 push。

## Goals / Non-Goals

**Goals:**

- 统一部署与 ACR 仓库名为 **`notify-service`**
- `docker-acr.yml` 支持全量 8 服务与 `v*+notify` 单服务构建
- 保持 App/Admin HTTP 契约不变

**Non-Goals:**

- 不改 banner API 路径或 JSON 字段
- 不重命名 `appstatus` Go 包或 `app_status_*_http.go`（除非 import 路径冲突）
- 不在 ACR 保留 `app-status-service` 镜像双推送

## Decisions

### D1：仅重命名运行时边界

进程二进制 `notify-service`，GoFrame server name `notify-service`，Compose service key `notify-service`，ACR repo `notify-service`。

### D2：环境变量

`NOTIFY_SERVICE_ADDR` 替代 `APP_STATUS_SERVICE_ADDR`；Compose/.env.example 仅保留新名（无长期双源）。

### D3：CI 别名

`notify`、`notify-service` → matrix id `notify-service` → `manifest/docker/Dockerfile.notify-service`。

### D4：文件迁移

| 旧 | 新 |
|----|-----|
| `cmd/app-status-service/` | `cmd/notify-service/` |
| `config.app-status-service.yaml` | `config.notify-service.yaml` |
| `Dockerfile.app-status-service` | `Dockerfile.notify-service` |
| `RegisterAppStatusServiceHTTP` | `RegisterNotifyServiceHTTP` |

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 测试/生产已部署 `app-status-service` 镜像 | runbook 说明首次需 pull `notify-service` 并改 compose |
| ACR 无 `notify-service` 仓库 | 首次 push 前在控制台创建或与现有命名空间策略一致 |

## Migration Plan

1. 合并代码并重命名文件
2. CI 推送 `notify-service:${IMAGE_TAG}`
3. 服务器更新 compose 服务名与 image 行，`docker compose pull notify-service && up -d notify-service`
