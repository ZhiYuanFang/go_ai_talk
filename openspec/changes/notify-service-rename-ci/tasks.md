## 1. 重命名运行时（go_ai_talk）

- [x] 1.1 `cmd/app-status-service` → `cmd/notify-service`；二进制与 `g.Server("notify-service")`；`NOTIFY_SERVICE_ADDR`
- [x] 1.2 `config.app-status-service.yaml` → `config.notify-service.yaml`
- [x] 1.3 `RegisterAppStatusServiceHTTP` → `RegisterNotifyServiceHTTP`（`register_notify_service.go`）
- [x] 1.4 `Dockerfile.app-status-service` → `Dockerfile.notify-service`

## 2. Compose 与 env

- [x] 2.1 各 `docker-compose.microservices*.yml` / `resources*.yml`：`app-status-service` → `notify-service`，image `${REGISTRY}/notify-service`
- [x] 2.2 `.env.example`：`NOTIFY_SERVICE_ADDR`、注释更新

## 3. CI

- [x] 3.1 `docker-acr.yml`：`notify`/`notify-service` 别名、`dockerfile_for`、`ALL_SERVICES` 8 服务、注释「八微服务」

## 4. 验证

- [x] 4.1 `go build ./...` 通过
- [x] 4.2 `openspec validate notify-service-rename-ci --strict` 通过
