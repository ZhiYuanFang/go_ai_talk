## 部署与运行指南（当前微服务形态）

适用范围：gateway / voice-service / device-service / history-service / worker-service。

### 1. 配置基线

- `gateway-service`：`manifest/config/config.yaml`（仅网关与公共项，不含数据库）
- `voice-service`：`manifest/config/config.voice-service.yaml`
- `device-service`：`manifest/config/config.device-service.yaml`
- `history-service`：`manifest/config/config.history-service.yaml`
- `worker-service`：`manifest/config/config.worker-service.yaml`

关键原则：

- 每服务一个数据库，配置以 `database.default` 为主；`device-service` 若需向 **history 库** 的 `domain_outbox` 投递事件，可额外配置 `database.history_relay`（与 history 同实例时），未配置时进程会跳过 outbox 写入并打 Debug 日志。
- 跨服务资料获取走 API，不跨库直查；`history-service` 内对 suggest/画像/事件字典的本地实现通过 `VOICE_SERVICE_URL` / `DEVICE_SERVICE_URL` 委派到对应服务（默认值见 `internal/services/contracts/http_targets.go`）。**容器或 Pod 内** `127.0.0.1` 指向本实例自身，不得用于访问其他微服务；Compose 参考见 `manifest/docker/docker-compose.microservices.yml` 中 `history-service.environment`，K8s 参考见 `manifest/deploy/kustomize/base/history-deployment.yaml`。
- `voice-service` 对 device 域（事件/动作/画像/注册校验/最近对话等）**仅经 HTTP**（`DEVICE_SERVICE_URL` → `internal/services/device/admin_http_client.go`），不得依赖 voice 进程 default 库直连 `user`/`event`/`action`；部署时建议 `DEVICE_PROFILE_SERVICE_MODE=remote`（与 `manifest/deploy/.../voice-deployment.yaml` 一致）。

### 2. 本地 Compose 启动

1) 启动依赖：
- 创建共享网络：
> docker network create go-ai-talk-net
> <b>列出docker中的网络配置</b>
> docker network ls
> <b>列出所有容器所在的docker网络</b>
> docker ps -a --format '{{.Names}}' | xargs -I {} docker inspect {} --format '{{.Name}} => {{range $k,$v := .NetworkSettings.Networks}}{{$k}} {{end}}'
- Redis 集群：`docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d` --force-recreate
- RabbitMQ：`docker compose -f manifest/docker/docker-compose.rabbitmq.yml up -d` --force-recreate

2) 启动业务：

- `docker compose -f manifest/docker/docker-compose.microservices.yml up -d --build`

3) 健康检查（自宿主机探测各服务端口映射；**history 对 voice/device 的 HTTP 委派在容器内走服务名**，见上文环境变量）：

- gateway: `http://127.0.0.1:9701/api.json`
- history-service: `http://127.0.0.1:9801/api.json`
- voice-service: `http://127.0.0.1:9802/api.json`
- device-service: `http://127.0.0.1:9803/api.json`
- worker: `http://127.0.0.1:9901/healthz`

### 3. Kubernetes 部署要点

- 使用 `manifest/deploy/kustomize/overlays/develop`
- `history-service` Deployment 须包含 `VOICE_SERVICE_URL`、`DEVICE_SERVICE_URL`（与 `base/history-deployment.yaml` 一致或 overlay 覆盖为集群内可达基址）
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
