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
- Redis 集群：
> `docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate`
（各节点加入 `go-ai-talk-net` 后，业务配置里的主机名一般为 **`redis-node-1`** 等服务名；`docker ps` 显示的 `docker-redis-node-1-1` 为容器名，与 DNS 解析名不必相同。集群需先完成 `redis-cli --cluster create` 等初始化，否则会出现 `CLUSTERDOWN` / `Hash slot not served`。）
- 初始化redis: 
> `docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
>  redis-cli --cluster create \
>  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
>  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
>  --cluster-replicas 1 --cluster-yes`
- 判断redis是否初始化成功：
> `docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 redis-cli -p 7001 CLUSTER INFO`
> 里应有 cluster_state:ok。
- RabbitMQ：
> `docker compose -f manifest/docker/docker-compose.rabbitmq.yml up -d` --force-recreate

2) 准备业务库连接（**强烈建议**）：

- 在 **`manifest/docker/.env.example`** 或复制后的 **`manifest/docker/.env`** 中填写非空的 `HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK`、`WORKER_DB_LINK`（Gf DSN）。**MySQL 跑在 Docker 宿主机上**时主机名用 `host.docker.internal`；**MySQL 在其它机器**时改为容器内可达的 RDS/内网地址。  
- 四个变量**留空**时不会覆盖 yaml，进程仍使用 `config.*.yaml` 里的占位地址（如公网 `120.55.50.105:3306`）。  
- 若你已在 `.env.example` 里写好 link 却仍连旧 IP：① 确认 compose 中 DSN 插值已用**引号**（本仓库已改为 `"${DEVICE_DB_LINK:-}"` 等形式，避免 YAML 把 `mysql:...:3306` 截断）；② 对业务服务 **`docker compose ... up -d --force-recreate`**；③ 容器内 **`printenv DEVICE_DB_LINK`** 核对。  
- `manifest/docker/docker-compose.microservices.yml` 已为 `history-service`、`voice-service`、`device-service`、`worker` 配置 `extra_hosts: host.docker.internal:host-gateway`（Docker 20.10+），便于同机连库。

3) 启动业务（`--env-file` 指向你实际填写了四个 LINK 的文件即可）：

- `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build`  
  或使用已 gitignore 的 `manifest/docker/.env`：`--env-file manifest/docker/.env`
> 只改docker-compose环境配置
>  `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --force-recreate`  
> 针对特定服务build
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build device-service`  
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build voice-service`  
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build history-service`  
4) 健康检查（自宿主机探测各服务端口映射；**history 对 voice/device 的 HTTP 委派在容器内走服务名**，见上文环境变量）：

- gateway: `http://127.0.0.1:9701/api.json`
- history-service: `http://127.0.0.1:9801/api.json`
- voice-service: `http://127.0.0.1:9802/api.json`
- device-service: `http://127.0.0.1:9803/api.json`
- worker: `http://127.0.0.1:9901/healthz`

### 3. Kubernetes 部署要点

- 使用 `manifest/deploy/kustomize/overlays/develop`
- `history-service` Deployment 须包含 `VOICE_SERVICE_URL`、`DEVICE_SERVICE_URL`（与 `base/history-deployment.yaml` 一致或 overlay 覆盖为集群内可达基址）
- 各业务服务数据库连接可通过与各 `cmd/*-service/main.go` 一致的 `*_DB_LINK` 环境变量覆盖；`worker-service` 使用 `WORKER_DB_LINK`（与 Compose / `.env.example` 对齐）
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
