## Why

`history-service` 在通过 HTTP 委派到 `voice-service` / `device-service` 时使用 `internal/services/contracts/http_targets.go` 的默认基址（`http://127.0.0.1:9802`、`http://127.0.0.1:9803`）。在 Docker 同一 overlay/bridge 网段内，`127.0.0.1` 指向**当前容器自身**，无法到达其他容器，导致委派请求 connection refused。`manifest/docker/docker-compose.microservices.yml` 中 `history-service` 当前未注入 `VOICE_SERVICE_URL` / `DEVICE_SERVICE_URL`，与 `voice-service` 已显式配置下游 URL 不一致。

## What Changes

- 在 **Docker Compose（微服务栈）** 的 `history-service` 环境变量中显式设置 `VOICE_SERVICE_URL`、`DEVICE_SERVICE_URL` 为 compose 服务 DNS 名（如 `http://voice-service:9802`、`http://device-service:9803`）。
- （可选但推荐）在 **Kustomize base** 的 history Deployment 中补齐同等语义的环境变量占位或注释，避免 K8s 部署沿用错误默认。
- 在 **运行手册 / `config.history-service.yaml` 注释** 中说明：容器化部署 MUST 设置上述变量；`127.0.0.1` 默认值仅适用于「所有服务进程在同一网络命名空间（本机 go run）」场景。

## Capabilities

### New Capabilities

- `history-delegate-downstream-urls`：约束 history 在容器/K8s 网络中对 voice、device 的 HTTP 委派基址解析规则，与仓库参考部署（compose、kustomize）一致。

### Modified Capabilities

- （无）本变更以部署清单与文档为主；若未来将「默认回退 URL」行为写入全局 spec，可再开变更做 delta。

## Impact

- `manifest/docker/docker-compose.microservices.yml`（`history-service.environment`）。
- 可选：`manifest/deploy/kustomize/base/` 下 history 相关清单。
- `docs/runbooks/release-deploy-and-run.md`、`manifest/config/config.history-service.yaml` 注释。
- 无 API 契约变更；无 Go 行为变更（除非后续任务选择增加启动期校验）。
