## Why

`app-status-service` 已实现维护/公告能力，但进程名、镜像名与 CI 全量构建列表不一致：GitHub `docker-acr.yml` 仍为 **7 服务**，该镜像无法随 tag 构建推送。产品命名统一为 **`notify-service`**，并纳入 ACR 流水线（全量 8 服务、`+notify` 单服务构建）。

## What Changes

- **重命名**：`cmd/app-status-service` → `cmd/notify-service`；配置/Dockerfile/Compose 服务名与镜像名 `notify-service`；环境变量 `NOTIFY_SERVICE_ADDR`（默认 `:9806`）。
- **不变**：HTTP 路径 `/app/api/status/banner`、`/admin/*`；静态页 `app-status-admin.html`；`internal/services/appstatus` 包名。Flutter 客户端已迁移至 `NOTIFY_BASE_URL`（`AppEnv.notifyBaseUrl`，legacy 回退 `STATUS_BASE_URL`）；HTTP 路径与 JSON 契约不变。
- **CI**：`.github/workflows/docker-acr.yml` 增加 `notify`/`notify-service` 别名、`Dockerfile.notify-service`，`ALL_SERVICES` 扩为 8 服务；注释与 runbook 同步。
- **非 BREAKING**（对外）：API 与 App 契约不变；部署需 pull 新镜像名 `notify-service` 并更新 compose 服务名。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `notify-service-runtime`：进程/镜像/Compose/CI 命名与构建范围。

## Impact

- `cmd/`、`manifest/config/`、`manifest/docker/`、`.github/workflows/docker-acr.yml`
- `internal/controller/register_*` 注册函数名
- OpenSpec `app-status-banner-service` 文档中的服务名可在归档时合并修正
