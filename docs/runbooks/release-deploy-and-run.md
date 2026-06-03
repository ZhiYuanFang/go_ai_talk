## 部署与运行指南（当前微服务形态）

适用范围：gateway / **gateway-app** / voice-service / device-service / history-service / worker-service。

### 1. 配置基线

- `gateway-service`：`manifest/config/config.yaml`（仅网关与公共项，不含数据库）
- **`gateway-app-server`**：`manifest/config/config.gateway-app-server.yaml`（**含** `database.app`，用于版本检查读 `ai_voice_app`；**新建实例**须在 yaml 中把 `database.app.link` 配为真实 DSN，或通过环境变量 **`APP_DB_LINK`** 覆盖——与 `HISTORY_DB_LINK` 等同理，Compose 已传 `${APP_DB_LINK:-}`，勿漏配导致版本接口连库失败）
- **gateway-app-server（版本管理 / APK）**：`version.download_url` 存 **path-only**（如 `/device/app/apk/<file>.apk`），客户端与 `version/check` 按同源或网关基址拼接访问。管理员口令仅通过 **`GATEWAY_APP_VERSION_ADMIN_PASSWORD`**（或本地私有 yaml 中的 `gatewayApp.versionAdmin.password`，**勿将生产口令提交到 git**）注入；未配置时管理 API 返回 503。APK 默认落盘 **`/apk/ai_talk/`**（可用 `GATEWAY_APP_APK_STORAGE_DIR` 覆盖，**须与 Compose 挂载点一致**）。**Docker Compose** 已将 **`/apk/ai_talk`** bind 到宿主机同路径。管理页：`GET /device/app/version-admin.html`（登录后可列表/编辑/删除/上传）。管理 API（均需先 `POST /device/app/api/version/admin/login`，Cookie 鉴权）：`GET .../admin/list`、`GET .../admin/get?id=`、`POST .../admin/update`、`POST .../admin/delete`、`POST .../admin/upload`。匿名下载：`GET /device/app/apk/{文件名}`。`GET /device/app/api/version/check` 取表中 **最大 `id`** 行作为最新发版。**APK 上传体积极限**：`config.gateway-app-server.yaml` 中 `server.clientMaxBodySize`（当前 220MB）须 **≥** `gatewayApp.apkMaxBytes`（默认 200MB）；GoFrame 未配置时默认仅 **8MB**，超过会在框架层断连，管理页表现为 **`failed to fetch`** 而非业务 JSON「文件过大」。反代 Nginx 时另需 `client_max_body_size` 与超时与之一致。
- **事件 logo 落盘（device-service）**：默认 **`/ai_talk_images/`**（`device.eventImageStorageDir`；可用 **`DEVICE_EVENT_IMAGE_STORAGE_DIR`** 覆盖，须与挂载点一致）。**Docker Compose** 已将 **`/ai_talk_images`** bind 到宿主机根目录。
- `voice-service`：`manifest/config/config.voice-service.yaml`
- `device-service`：`manifest/config/config.device-service.yaml`
- `history-service`：`manifest/config/config.history-service.yaml`
- `worker-service`：`manifest/config/config.worker-service.yaml`
- **对话链路（`voiceChat`）**：`manifest/config/voice-chat.shared.yaml`（`voice-service` 与 `history-service` 共用，避免在两份服务 YAML 中重复维护）；可选环境变量 `GF_VOICE_CHAT_FILE` 指向其它路径（仍按 `gfile.Search` 规则解析）。若共享文件缺失且 `GF_GCFG` 中未声明 `voiceChat`，文本对话会报「DeepSeek endpoint 未配置」。

关键原则：

- 每服务一个数据库，配置以 `database.default` 为主；`device-service` 若需向 **history 库** 的 `domain_outbox` 投递事件，可额外配置 `database.history_relay`（与 history 同实例时），未配置时进程会跳过 outbox 写入并打 Debug 日志。
- 跨服务资料获取走 API，不跨库直查；`history-service` 内对 suggest/画像/事件字典的本地实现通过 `VOICE_SERVICE_URL` / `DEVICE_SERVICE_URL` 委派到对应服务（默认值见 `internal/services/contracts/http_targets.go`）。**容器或 Pod 内** `127.0.0.1` 指向本实例自身，不得用于访问其他微服务；Compose 参考见 `manifest/docker/docker-compose.microservices.yml` 中 `history-service.environment`，K8s 参考见 `manifest/deploy/kustomize/base/history-deployment.yaml`。
- `voice-service` 对 device 域（事件/动作/画像/注册校验/最近对话等）**仅经 HTTP**（`DEVICE_SERVICE_URL` → `internal/services/device/admin_http_client.go`），不得依赖 voice 进程 default 库直连 `user`/`event`/`action`；部署时建议 `DEVICE_PROFILE_SERVICE_MODE=remote`（与 `manifest/deploy/.../voice-deployment.yaml` 一致）。
- **设备管理页事件 logo（:9701）**：`GET /device/admin` 加载 `admin.html`；事件 logo 预览为 **主网关同源** `origin + /ai_talk_images/...`。主网关须反代 `/ai_talk_images/*` 至 device-service（与 `DEVICE_API_PROXY_URL` 一致），并部署含 Logo/色调列与行内点击编辑的静态页。App 网关 `:9702` 仍保留 `/ai_talk_images` 反代供客户端使用。

### 解决宝塔面板无响应
1. 重启宝塔面板服务
> /etc/init.d/bt restart

2. 等待几秒后，检查服务状态（看是否显示 running）
> /etc/init.d/bt status

### 2. 本地 Compose 启动
> cd /www/wwwroot/go/go_ai_talk/
> mysql可能没有启动
> systemctl start mysql-local
> systemctl status mysql-local
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
> 开始的时候需要初始化交换机:
> `chmod +x hack/rabbitmq-init.sh && ./hack/rabbitmq-init.sh`

2) 准备业务库连接（**强烈建议**）：

- 在 **`manifest/docker/.env.example`** 或复制后的 **`manifest/docker/.env`** 中填写非空的 `HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK`、`WORKER_OUTBOX_DB_LINK`、**`APP_DB_LINK`**（Gf DSN；**含 gateway-app** 连 `ai_voice_app`，与其它服务一样避免新实例沿用 yaml 占位地址）。**MySQL 跑在 Docker 宿主机上**时主机名用 `host.docker.internal`；**MySQL 在其它机器**时改为容器内可达的 RDS/内网地址。  
- 上述变量**留空**时不会覆盖 yaml，进程仍使用 `config.*.yaml` 里的占位地址（如公网 `120.55.50.105:3306`）；**gateway-app** 另看 `database.app.link`，未设 `APP_DB_LINK` 时同样易踩占位库。  
- 若你已在 `.env.example` 里写好 link 却仍连旧 IP：① 确认 compose 中 DSN 插值已用**引号**（本仓库已改为 `"${DEVICE_DB_LINK:-}"` 等形式，避免 YAML 把 `mysql:...:3306` 截断）；② 对业务服务 **`docker compose ... up -d --force-recreate`**；③ 容器内 **`printenv DEVICE_DB_LINK`** 核对。  
- `manifest/docker/docker-compose.microservices.yml` 已为 `history-service`、`voice-service`、`device-service`、`worker` 配置 `extra_hosts: host.docker.internal:host-gateway`（Docker 20.10+），便于同机连库。

**宿主机静态资源目录（Linux + Compose，首次部署建议）**

Compose 将容器内写盘路径映射到宿主机根目录，便于 `ls`、备份与 `recreate` 后文件仍在：

| 宿主机路径 | 服务 | 用途 |
|------------|------|------|
| `/ai_talk_images` | device-service | 事件 logo 文件 |
| `/apk/ai_talk` | gateway-app | App 版本 APK |

```bash
sudo mkdir -p /ai_talk_images /apk/ai_talk
sudo chmod 755 /ai_talk_images /apk/ai_talk
```

更新 compose 卷配置后，至少重建写盘服务：

```bash
docker compose -f manifest/docker/docker-compose.microservices.yml up -d --force-recreate device-service gateway-app
```

验收示例（上传 logo / APK 之后）：

```bash
ls -la /ai_talk_images/
ls -la /apk/ai_talk/
docker exec go-ai-talk-device-service ls -la /ai_talk_images/
docker exec go-ai-talk-gateway-app ls -la /apk/ai_talk/
```

RHEL/CentOS 等若遇容器无法写入，可在 compose 卷行尝试后缀 `:z`（SELinux）。若曾把文件只写在旧容器层内，需 `docker cp` 到宿主机目录或重新上传。

3) 启动业务（`--env-file` 指向你实际填写了各 `*_DB_LINK` / `APP_DB_LINK` 的文件即可）：
> 停止并删除 compose 项目
> docker compose -f manifest/docker/docker-compose.microservices.yml down
> 清理 Docker 系统
> docker system prune -a -f
> 清理构建缓存
> docker builder prune -f
- `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build`  
  或使用已 gitignore 的 `manifest/docker/.env`：`--env-file manifest/docker/.env`
> 只改docker-compose环境配置
>  `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --force-recreate`  
> 针对特定服务build
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build device-service`  
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build voice-service`  
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build history-service`  
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build ucg-service`  
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build gateway-app`  
> `docker compose --env-file manifest/docker/.env.example -f manifest/docker/docker-compose.microservices.yml up -d --build gateway`  
4) 健康检查（自宿主机探测各服务端口映射；**history 对 voice/device 的 HTTP 委派在容器内走服务名**，见上文环境变量）：

- gateway: `http://127.0.0.1:9701/api.json`
- gateway-app: `http://127.0.0.1:9702/api.json`
- history-service: `http://127.0.0.1:9801/api.json`
- voice-service: `http://127.0.0.1:9802/api.json`
- device-service: `http://127.0.0.1:9803/api.json`
- worker: `http://127.0.0.1:9901/healthz`

### 3. Kubernetes 部署要点

- 使用 `manifest/deploy/kustomize/overlays/develop`
- `history-service` Deployment 须包含 `VOICE_SERVICE_URL`、`DEVICE_SERVICE_URL`（与 `base/history-deployment.yaml` 一致或 overlay 覆盖为集群内可达基址）
- 各业务服务数据库连接可通过与各 `cmd/*-service/main.go` 一致的 `*_DB_LINK` 环境变量覆盖；`worker-service` 使用 `WORKER_OUTBOX_DB_LINK`（与 Compose / `.env.example` 对齐）；**`gateway-app`** 使用 **`APP_DB_LINK`** → `GF_DATABASE_APP_LINK`（覆盖 `database.app`，新建 Pod/实例时勿漏）
- 确认 worker deployment 的 `GF_GCFG_FILE` 指向 `manifest/config/config.worker-service.yaml`
- 确认 **主网关** `gateway` deployment 的主配置不包含数据库字段；**`gateway-app`** 则必须在 yaml 或 `APP_DB_LINK` 中配置 `ai_voice_app`

### 4. 发布前检查

- `go test ./cmd/... ./internal/...`
- 检查各服务 `GF_GCFG_FILE` 是否指向对应专属配置
- 检查主网关 `gateway` 是否无 DB 访问路径；检查 **`gateway-app`** 是否已配置 **`APP_DB_LINK`** 或可连通的 `database.app.link`
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
