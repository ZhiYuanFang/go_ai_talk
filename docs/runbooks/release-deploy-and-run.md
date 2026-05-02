## 部署与运行指南（当前微服务形态）

适用范围：gateway / voice-service / device-service / history-service / worker-service。

### 1. 配置基线

- `gateway-service`：`manifest/config/config.yaml`（仅网关与公共项，不含数据库）
- `voice-service`：`manifest/config/config.voice-service.yaml`
- `device-service`：`manifest/config/config.device-service.yaml`
- `history-service`：`manifest/config/config.history-service.yaml`
- `worker-service`：`manifest/config/config.worker-service.yaml`

关键原则：

- 每服务一个数据库，配置仅维护 `database.default`
- 跨服务资料获取走 API，不跨库直查

### 2. 本地 Compose 启动

1) 启动依赖：
- 创建共享网络：
> docker network create go-ai-talk-net
> docker network ls
- Redis 集群：`docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d` --force-recreate
- RabbitMQ：`docker compose -f manifest/docker/docker-compose.rabbitmq.yml up -d` --force-recreate

2) 启动业务：

- `docker compose -f manifest/docker/docker-compose.microservices.yml up -d --build`

3) 健康检查：

- gateway: `http://127.0.0.1:9701/api.json`
- history-service: `http://127.0.0.1:9801/api.json`
- voice-service: `http://127.0.0.1:9802/api.json`
- device-service: `http://127.0.0.1:9803/api.json`
- worker: `http://127.0.0.1:9901/healthz`

### 3. Kubernetes 部署要点

- 使用 `manifest/deploy/kustomize/overlays/develop`
- 确认 worker deployment 的 `GF_GCFG_FILE` 指向 `manifest/config/config.worker-service.yaml`
- 确认 gateway deployment 的主配置不包含数据库字段

### 4. 发布前检查

- `go test ./cmd/... ./internal/...`
- 检查各服务 `GF_GCFG_FILE` 是否指向对应专属配置
- 检查 gateway 是否无 DB 访问路径
- 检查 worker outbox relay 的 `database.default` 是否指向目标库

### 5. 回滚步骤（按服务维度）

1) 配置回滚：将目标服务 `GF_GCFG_FILE` 回切到上一个稳定配置文件。  
2) 镜像回滚：回退对应 deployment / compose 镜像 tag。  
3) DAO 模型回滚（应急）：如发现数据库路由问题，回滚到上一个稳定版本二进制与配置。  
4) 验证恢复：健康探针、关键 API、outbox relay 状态恢复正常。

### 6. 文档治理

凡涉及运行、部署、配置边界、DAO 访问模式的需求变更，必须同步更新：

- `docs/runbooks/dao-sync-by-domain.md`
- `docs/runbooks/release-deploy-and-run.md`
