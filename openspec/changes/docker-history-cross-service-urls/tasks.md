## 1. Compose 与 K8s 部署面

- [x] 1.1 在 `manifest/docker/docker-compose.microservices.yml` 的 `history-service.environment` 增加 `VOICE_SERVICE_URL: "http://voice-service:9802"` 与 `DEVICE_SERVICE_URL: "http://device-service:9803"`（与同文件内服务名、端口一致）。
- [x] 1.2 在 `manifest/deploy/kustomize/base/history-deployment.yaml` 的容器 `env` 中增加上述两个变量，基址使用 ClusterIP Service 名（`http://voice-service:9802`、`http://device-service:9803`）。

## 2. 配置与文档

- [x] 2.1 在 `manifest/config/config.history-service.yaml` 顶部或 `server` 段附近补充中文注释：容器/K8s 部署 MUST 设置 `VOICE_SERVICE_URL`、`DEVICE_SERVICE_URL`；`127.0.0.1` 默认仅适用于本机同命名空间多进程。
- [x] 2.2 更新 `docs/runbooks/release-deploy-and-run.md`：在 history 相关小节说明跨容器访问不得使用 `127.0.0.1` 指向他服务，并指向 compose/kustomize 示例。

## 3. 校验

- [x] 3.1 本地执行 `openspec validate docker-history-cross-service-urls --strict`（若 CLI 可用）确保变更目录通过校验。
- [ ] 3.2 （可选）`docker compose -f manifest/docker/docker-compose.microservices.yml` 起栈后，从 history 容器内或经网关验证委派链路无 connection refused（人工或现有 smoke 脚本）。
