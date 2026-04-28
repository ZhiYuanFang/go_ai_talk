## 容器化与健康探针（Task 7.1）

本阶段目标：为 `gateway`、`history-service`、`worker` 提供可独立构建镜像与基础健康探针，作为后续 Kubernetes 部署的前置。

### 新增/更新文件

- `manifest/docker/Dockerfile.gateway-service`
- `manifest/docker/Dockerfile.history-service`（新增 HEALTHCHECK）
- `manifest/docker/Dockerfile.worker-service`
- `manifest/docker/docker-compose.microservices.yml`
- `cmd/worker-service/main.go`

### 健康探针

- gateway: `GET /api.json`（端口 `9701`）
- history-service: `GET /api.json`（端口 `9801`）
- worker: `GET /healthz`（端口 `9901`）

### 本地运行

1. 启动（示例）：
   - `docker compose -f manifest/docker/docker-compose.microservices.yml up -d --build`
2. 查看健康状态：
   - `docker compose -f manifest/docker/docker-compose.microservices.yml ps`
3. 访问探针：
   - gateway: [http://127.0.0.1:9701/api.json](http://127.0.0.1:9701/api.json)
   - history: [http://127.0.0.1:9801/api.json](http://127.0.0.1:9801/api.json)
   - worker: [http://127.0.0.1:9901/healthz](http://127.0.0.1:9901/healthz)

### 说明

- `worker` 进程用于承载异步消费/relay 逻辑，默认通过环境变量开启：
  - `MQ_CONSUMER_ENABLED=true`
  - `OUTBOX_RELAY_ENABLED=true`
- `gateway` 默认关闭内部消费者，避免与 worker 重复消费：
  - `MQ_CONSUMER_ENABLED=false`
