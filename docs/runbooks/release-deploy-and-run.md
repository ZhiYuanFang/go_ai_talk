## 部署与运行指南

适用范围：gateway / gateway-app / voice-service / device-service / history-service / worker-service / ucg-service。

**三种运行方式**

| 方式 | 何时用 | 镜像来源 | 服务器是否需要完整源码 |
|------|--------|----------|------------------------|
| **A. 本地开发** | 本机改代码联调 | 本机 `docker compose --build`（`:local`） | 要（整仓） |
| **B. 测试环境** | 合并 `develop` 后在 test 域名验收 | ACR `:develop`（GitHub Actions 构建） | **不要**（仅需 compose + `.env.test`） |
| **C. 生产环境** | 发版 semver tag 后上线 | ACR `:v*`（GitHub Actions 构建） | **不要**（仅需 compose + `.env.prod`） |

镜像由 `.github/workflows/docker-acr.yml` 构建 push；服务器 **`pull` + `--no-build`**，不在服务器编译 Go 源码。

---

## A. 本地开发运行

> 工作目录：仓库根目录（如 `/www/wwwroot/go/go_ai_talk/`）。

### A.1 首次搭建（一次性）

**1. 中间件**

```bash
docker network create go-ai-talk-net

docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
  --cluster-replicas 1 --cluster-yes
# 验收：cluster_state:ok
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO

docker compose -f manifest/docker/docker-compose.rabbitmq.yml up -d --force-recreate
chmod +x hack/rabbitmq-init.sh && ./hack/rabbitmq-init.sh
```

**2. MySQL**（宿主机上）

```bash
systemctl start mysql-local   # 或你的 mysqld 服务名
```

**3. 环境变量**

复制并填写 `manifest/docker/.env.example` → `manifest/docker/.env`（或直接用 `.env.example`）：

- `HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK`、`WORKER_OUTBOX_DB_LINK`、`APP_DB_LINK`、`UCG_DB_LINK`
- MySQL 与 Docker 同机时主机用 `host.docker.internal`

**4. 静态资源目录**

```bash
sudo mkdir -p /ai_talk_images /apk/ai_talk && sudo chmod 755 /ai_talk_images /apk/ai_talk
```

### A.2 日常：改了代码要跑起来

**全量重建并启动**

```bash
docker compose --env-file manifest/docker/.env \
  -f manifest/docker/docker-compose.microservices.yml \
  up -d --build
```

**只改某一个服务**（更快）

```bash
docker compose --env-file manifest/docker/.env \
  -f manifest/docker/docker-compose.microservices.yml \
  up -d --build voice-service
# 同理：gateway / gateway-app / history-service / device-service / worker / ucg-service
```

**只改了 compose 环境变量、未改代码**

```bash
docker compose --env-file manifest/docker/.env \
  -f manifest/docker/docker-compose.microservices.yml \
  up -d --force-recreate
```

### A.3 验收

```bash
curl -s http://127.0.0.1:9701/api.json    # gateway
curl -s http://127.0.0.1:9702/api.json    # gateway-app
curl -s http://127.0.0.1:9801/api.json    # history
curl -s http://127.0.0.1:9802/api.json    # voice
curl -s http://127.0.0.1:9803/api.json    # device
curl -s http://127.0.0.1:9901/healthz     # worker
```

---

## B. 发布测试环境

对外：`https://test.pangbao.cuplay.top:9701` / `:9702`（Nginx 反代至宿主机 **19701 / 19702**）。  
镜像 tag：**`develop`**（`.env.test` 中 `IMAGE_TAG=develop`）。

### B.1 首次搭建测试栈（一次性，与生产同机且完全隔离）

> 生产已运行时，测试使用独立网络、Redis、RabbitMQ、MySQL `_test` 库、静态目录 `*_test`。对照表见 [附录：生产/测试对照](#附录生产测试对照)。

```bash
# 1) 网络
docker network create go-ai-talk-test-net

# 2) MySQL 建库：ai_voice_history_test、ai_voice_device_test、…（各域 + worker + app + ucg）

# 3) 静态目录
sudo mkdir -p /ai_talk_images_test /apk/ai_talk_test && sudo chmod 755 /ai_talk_images_test /apk/ai_talk_test

# 4) 脱敏种子（可选，从生产导入测试数据）
MYSQL_PASS='***' ./hack/mask-seed-data.sh

# 5) 测试 Redis + RabbitMQ
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
  --cluster-replicas 1 --cluster-yes

docker compose -f manifest/docker/docker-compose.rabbitmq.test.yml up -d --force-recreate
COMPOSE_FILE=manifest/docker/docker-compose.rabbitmq.test.yml \
RABBIT_API_BASE=http://127.0.0.1:15673/api ./hack/rabbitmq-init.sh

# 6) 环境文件
cp manifest/docker/.env.test.example manifest/docker/.env.test
# 填写 REGISTRY、*_DB_LINK（指向 *_test 库）、密钥等
```

### B.2 日常：把改动发布到测试（逐步）

**步骤 1 — 开发机：提交并推送**

```bash
git checkout develop
git add … && git commit -m "…"
git push origin develop
```

**步骤 2 — 等待 CI**

打开 GitHub → Actions → `docker-acr` 工作流，确认 **develop** 推送的构建全部成功（七服务镜像 tag `:develop`）。

**步骤 3 — 测试服务器：拉镜像并更新**

> 服务器只需保留 `manifest/docker/` 下 compose 与 `.env.test`，**无需**整仓源码。

```bash
cd /path/to/deploy   # 含 manifest/docker/ 的目录即可

COMPOSE_PROJECT_NAME=go-ai-talk-test \
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  pull

COMPOSE_PROJECT_NAME=go-ai-talk-test \
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  up -d --no-build
```

**只更新单个服务**

```bash
COMPOSE_PROJECT_NAME=go-ai-talk-test \
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  pull voice-service

COMPOSE_PROJECT_NAME=go-ai-talk-test \
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  up -d --no-build --force-recreate voice-service
```

**步骤 4 — 验收**

```bash
curl -sk https://test.pangbao.cuplay.top:9702/api.json
curl -s http://127.0.0.1:19701/api.json
docker exec go-ai-talk-history-service-test printenv HISTORY_DB_LINK
# 须含 _test 库名
```

**排错：pin 某次构建** — 将 `.env.test` 中 `IMAGE_TAG` 改为 Actions 推送的 **git 完整 sha**，再 `pull` + `up --no-build --force-recreate`。

---

## C. 发布生产环境

对外：`www.pangbao.cuplay.top:9701` / `:9702`。  
镜像 tag：与 git tag 一致（如 **`v1.0.0`**），写在 `manifest/docker/.env.prod` 的 **`IMAGE_TAG`**。

### C.1 首次搭建生产栈（一次性）

```bash
docker network create go-ai-talk-net

# Redis / RabbitMQ（生产端口 7001–7006、5672/15672）— 见 A.1 同类命令，用非 .test 的 compose 文件

sudo mkdir -p /ai_talk_images /apk/ai_talk && sudo chmod 755 /ai_talk_images /apk/ai_talk

cp manifest/docker/.env.prod.example manifest/docker/.env.prod
# 填写 REGISTRY、IMAGE_TAG、各 *_DB_LINK、生产密钥（勿提交 git）
```

### C.2 日常：把改动发布到生产（逐步）

**步骤 1 — 测试环境验收通过**（完成 B.2 全链路，含 MQ / 关键业务）

**步骤 2 — 开发机：打 tag 并推送**

```bash
git tag v1.0.0
git push origin v1.0.0
```

**步骤 3 — 等待 CI**

GitHub Actions `docker-acr` 对 **tag v1.0.0** 构建成功（七服务镜像 `:v1.0.0`）。

**步骤 4 — 生产服务器：改 tag、拉镜像、更新**

编辑 `manifest/docker/.env.prod`：

```bash
IMAGE_TAG=v1.0.0
```

```bash
cd /path/to/deploy

docker compose --env-file manifest/docker/.env.prod \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.prod.yml \
  pull

docker compose --env-file manifest/docker/.env.prod \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.prod.yml \
  up -d --no-build
```

**步骤 5 — 验收**

```bash
curl -s http://127.0.0.1:9701/api.json
curl -s http://127.0.0.1:9702/api.json
curl -s http://127.0.0.1:9901/healthz
```

### C.3 生产回滚

1. 将 `.env.prod` 中 `IMAGE_TAG` 改回上一稳定版本（如 `v0.9.0`）。
2. `pull` + `up -d --no-build --force-recreate`（可按单服务）。
3. **禁止**对生产执行 `docker system prune -a`。

---

## 附录

### 配置基线（新建实例必读）

| 服务 | 配置文件 | 要点 |
|------|----------|------|
| gateway | `manifest/config/config.yaml` | 无数据库 |
| gateway-app | `config.gateway-app-server.yaml` | 须 `APP_DB_LINK` 或 `database.app`；APK `/apk/ai_talk`；版本管理页 |
| voice / device / history / worker / ucg | 各 `config.*-service.yaml` | `database.default`；history+voice 共用 `voice-chat.shared.yaml` |
| 跨服务 | Compose 环境变量 | 容器内勿用 `127.0.0.1` 访问他域；用服务名 |

关键原则：

- 每服务独立库；跨域走 HTTP，禁止跨库直查（见 `AGENTS.md`）。
- gateway-app：`GATEWAY_APP_PUBLIC_BASE_URL`、APK 上传需 Nginx `client_max_body_size` ≥ 220MB。
- device logo：`/ai_talk_images`；测试栈用 `/ai_talk_images_test`。

### 生产/测试对照

| 项 | 生产 | 测试 |
|----|------|------|
| Compose project | 默认 | `go-ai-talk-test` |
| Docker 网络 | `go-ai-talk-net` | `go-ai-talk-test-net` |
| Redis 宿主机端口 | 7001–7006 | 17001–17006 |
| RabbitMQ | 5672 / 15672 | 5673 / 15673 |
| gateway / gateway-app | 9701 / 9702 | 19701 / 19702 |
| MySQL | `ai_voice_*` | `ai_voice_*_test` |
| 静态目录 | `/ai_talk_images`、`/apk/ai_talk` | `*_test` 后缀 |
| `IMAGE_TAG` | semver（`.env.prod`） | `develop`（`.env.test`） |
| 域名 | www.pangbao.cuplay.top | test.pangbao.cuplay.top |

### Compose 文件说明

| 文件 | 用途 |
|------|------|
| `docker-compose.microservices.yml` | 基线拓扑；本地 `--build` |
| `docker-compose.microservices.test.yml` | 测试 overlay + registry pull |
| `docker-compose.microservices.prod.yml` | 生产 overlay + registry pull |

镜像引用：`${REGISTRY}/gateway:${IMAGE_TAG}`。`REGISTRY` = `<ACR域名>/<命名空间>`；仓库名单段（`gateway`、`device-service` 等，无 `go-ai-talk/` 前缀）。

### ACR 与 GitHub Secrets

个人版 ACR 镜像完整路径：`域名/命名空间/仓库名:tag`。**测试与生产使用不同命名空间**，须分别配置。

| 环境 | 服务器 `.env` 的 `REGISTRY` | GitHub Secret | CI 触发 |
|------|------------------------------|---------------|---------|
| **测试** | `crpi-xxx-vpc.../<测试命名空间>` | `ACR_REGISTRY_TEST` = `crpi-xxx.../<测试命名空间>`（公网，无 `-vpc`） | push `develop` |
| **生产** | `crpi-xxx-vpc.../<生产命名空间>` | `ACR_REGISTRY_PROD` = `crpi-xxx.../<生产命名空间>`（公网，无 `-vpc`） | push tag `v*` |

共用 Secrets（同一 ACR 实例）：

| Secret | 取值 |
|--------|------|
| `ACR_USERNAME` | ACR 控制台 → 访问凭证 → 登录用户名 |
| `ACR_PASSWORD` | 访问凭证 → 设置的**固定密码** |

**填写示例**（命名空间名以控制台为准）：

```text
# GitHub Secrets（push 用公网域名）
ACR_REGISTRY_TEST=crpi-lff3xynwzvqxxxjk.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test
ACR_REGISTRY_PROD=crpi-lff3xynwzvqxxxjk.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-prod

# 服务器 .env.test（pull 可用 -vpc）
REGISTRY=crpi-lff3xynwzvqxxxjk-vpc.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test

# 服务器 .env.prod
REGISTRY=crpi-lff3xynwzvqxxxjk-vpc.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-prod
```

每个命名空间下须预先创建 7 个镜像仓库：`gateway`、`gateway-app`、`history-service`、`voice-service`、`device-service`、`worker`、`ucg-service`（或开启「自动创建仓库」）。

push 报 `denied` 常见原因：① GitHub 用了 `-vpc` 域名；② 缺命名空间；③ 测试/生产 Secret 与 `.env` 命名空间不一致；④ 仓库未创建。

### 脱敏种子（测试库刷新）

```bash
MYSQL_HOST=127.0.0.1 MYSQL_USER=root MYSQL_PASS='***' ./hack/mask-seed-data.sh
```

发版前可选执行；导入后 recreate gateway-app / history / device / worker。

### 发布前检查

- `go test ./cmd/... ./internal/...`
- 测试栈 `develop` 验收通过（B.2 步骤 4）
- 生产 `.env.prod` 的 `IMAGE_TAG` 与 git tag 一致，且不为 `develop`
- `printenv *_DB_LINK` 确认 test/prod 库未串

### 宝塔面板无响应

```bash
/etc/init.d/bt restart
/etc/init.d/bt status
```

### Kubernetes

见 `manifest/deploy/kustomize/overlays/develop`；Compose 发版路径以本文 A/B/C 为准。

### 文档治理

运行、部署、配置边界变更须同步更新本文与 `docs/runbooks/dao-sync-by-domain.md`。
