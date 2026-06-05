## 部署与运行指南

适用范围：gateway / gateway-app / voice-service / device-service / history-service / worker-service / ucg-service。

**三种运行方式**

| 方式 | 何时用 | 镜像来源 | 服务器是否需要完整源码 |
|------|--------|----------|------------------------|
| **A. 本地开发** | 本机改代码联调 | 本机 `docker compose --build`（`:local`） | 要（整仓） |
| **B. 测试环境** | 合并 `develop` 后在 test 域名验收 | ACR `:develop`（GitHub Actions 构建） | **不要**（仅需 compose + `.env.test`） |
| **C. 生产环境** | 发版 semver tag 后上线 | ACR `:v*`（GitHub Actions 构建） | **不要**（仅需 compose + `.env.prod`） |

镜像由 `.github/workflows/docker-acr.yml` 构建 push；服务器 **`pull` + `--no-build`**，不在服务器编译 Go 源码。

### 同机生产 + 测试：Compose 项目名隔离（必读）

生产与测试 **必须** 使用 **不同的 Compose project**，否则后执行的 `up` 会接管同一套 service 名（`gateway`、`device-service` 等），**把另一环境容器停掉**。

overlay 文件已内置 `name`（无需再手填 `COMPOSE_PROJECT_NAME`，除非你要覆盖）：

| 栈 | Compose 文件 | Project 名 |
|----|--------------|------------|
| 生产微服务 | `microservices.yml` + `microservices.prod.yml` | `go-ai-talk-prod` |
| 测试微服务 | `microservices.yml` + `microservices.test.yml` | `go-ai-talk-test` |
| 生产 Redis | `docker-compose.redis-cluster.yml` | `go-ai-talk-redis` |
| 测试 Redis | `docker-compose.redis-cluster.test.yml` | `go-ai-talk-redis-test` |
| 生产 RabbitMQ | `docker-compose.rabbitmq.yml` | `go-ai-talk-rabbitmq` |
| 测试 RabbitMQ | `docker-compose.rabbitmq.test.yml` | `go-ai-talk-rabbitmq-test` |

验收双栈并存：

```bash
docker ps --format 'table {{.Names}}\t{{.Status}}' | grep -E 'go-ai-talk-(gateway|device|redis|rabbitmq)'
# 应同时看到生产（9701/9803…）与测试（-test 后缀 / 197xx）容器均为 Up
```

**曾用默认 project 名 `go_ai_talk` 启动过生产？** 更新 compose 后首次 `up` 会新建 `go-ai-talk-prod` 项目；确认新栈健康后，再清理旧项目（见 [D.1](#d1-生产--测试微服务互相把对方停掉)）。

### 启动前清理（遇 Conflict / 端口占用）

> **原则**：compose 固定了 `container_name`，用 **`docker rm -f <容器名>`** 一条命令即可删掉冲突容器（不管当初挂在哪个 project 下）。**不加 `-v`**，数据卷保留。  
> 生产清理**只删无 `-test` 后缀的 7 个容器**，不会动测试栈。

**工作目录**（以下所有命令前先执行）：

```bash
cd /www/wwwroot/go/go_ai_talk
```

#### ① 测试 Redis（1700x 端口占用 / redis-node Created）

```bash
# 删：旧 project「docker」+ 正式 project 的测试 Redis 共 12 个可能容器名
docker rm -f \
  docker-redis-node-1-1 docker-redis-node-2-1 docker-redis-node-3-1 \
  docker-redis-node-4-1 docker-redis-node-5-1 docker-redis-node-6-1 \
  go-ai-talk-redis-test-redis-node-1-1 go-ai-talk-redis-test-redis-node-2-1 \
  go-ai-talk-redis-test-redis-node-3-1 go-ai-talk-redis-test-redis-node-4-1 \
  go-ai-talk-redis-test-redis-node-5-1 go-ai-talk-redis-test-redis-node-6-1 \
  2>/dev/null

# 验：无输出 = 容器已空；无输出且下一行 OK = 端口已释放
docker ps -a --format '{{.Names}}' | grep -E 'redis-node|go-ai-talk-redis-test' \
  || echo 'OK: 测试 Redis 容器已清空'
ss -tlnp | grep -E '1700[1-6]' || echo 'OK: 17001-17006 端口已释放'
```

#### ② 测试 RabbitMQ（go-ai-talk-rabbitmq-test 冲突）

```bash
# 删：固定容器名
docker rm -f go-ai-talk-rabbitmq-test 2>/dev/null

# 验：无输出 = 已删
docker ps -a --filter name=go-ai-talk-rabbitmq-test --format '{{.Names}}' \
  | grep . && echo '仍有残留' || echo 'OK: 测试 RabbitMQ 已删除'
```

#### ③ 测试微服务（go-ai-talk-*-test 冲突；**不影响生产**）

```bash
# 删：7 个测试微服务固定容器名
docker rm -f \
  go-ai-talk-gateway-test go-ai-talk-gateway-app-test \
  go-ai-talk-history-service-test go-ai-talk-voice-service-test \
  go-ai-talk-device-service-test go-ai-talk-worker-test \
  go-ai-talk-ucg-service-test \
  2>/dev/null

# 验：只查 7 个微服务名（go-ai-talk-rabbitmq-test 等中间件不算残留）
docker ps -a --format '{{.Names}}' \
  | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|worker|ucg-service)-test$' \
  && echo '仍有残留，见上表' || echo 'OK: 测试微服务容器已清空'
```

#### ④ 生产微服务（**勿删 -test 容器**）

```bash
# 删：7 个生产微服务固定容器名（无 -test 后缀）
docker rm -f \
  go-ai-talk-gateway go-ai-talk-gateway-app go-ai-talk-history-service \
  go-ai-talk-voice-service go-ai-talk-device-service go-ai-talk-worker \
  go-ai-talk-ucg-service \
  2>/dev/null

# 验：无输出 = 已删（不应出现 -test 行）
docker ps -a --format '{{.Names}}' | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|worker|ucg-service)$' \
  && echo '仍有残留' || echo 'OK: 生产微服务容器已清空'
```

---

## A. 本地开发运行

> 工作目录：仓库根目录（如 `/www/wwwroot/go/go_ai_talk/`）。

### A.1 首次搭建（一次性）

**1. 中间件**

```bash
docker network create go-ai-talk-net

docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.yml ps   # 6 节点均须 running

# 先验收集群是否已可用（见附录「Redis Cluster 验收」）；cluster_state:ok 则跳过 cluster create
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state

# 仅当上一步非 cluster_state:ok（首次初始化或 down -v 重置卷后）才执行：
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
  --cluster-replicas 1 --cluster-yes

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok

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
  -f manifest/docker/docker-compose.microservices.local.yml \
  up -d --build
```

**只改某一个服务**（更快）

```bash
docker compose --env-file manifest/docker/.env \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.local.yml \
  up -d --build voice-service
# 同理：gateway / gateway-app / history-service / device-service / worker / ucg-service
```

**只改了 compose 环境变量、未改代码**

```bash
docker compose --env-file manifest/docker/.env \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.local.yml \
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

# 5) 测试 Redis + RabbitMQ（仓库根目录；Conflict 时先跑上文 ①②）

docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml ps   # 6 节点均须 running

# 先验收集群是否已可用（见附录「Redis Cluster 验收」）；cluster_state:ok 则跳过 cluster create
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state

# 仅当上一步非 cluster_state:ok（首次初始化或 down -v 重置卷后）才执行：
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
  --cluster-replicas 1 --cluster-yes

docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok

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
cd /path/to/deploy   # 仓库根目录

# ① 清理（Conflict 时必做）— 见上文「③ 测试微服务」
docker rm -f go-ai-talk-gateway-test go-ai-talk-gateway-app-test \
  go-ai-talk-history-service-test go-ai-talk-voice-service-test \
  go-ai-talk-device-service-test go-ai-talk-worker-test \
  go-ai-talk-ucg-service-test 2>/dev/null
docker ps -a --format '{{.Names}}' \
  | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|worker|ucg-service)-test$' \
  && echo '仍有残留' || echo 'OK: 测试微服务容器已清空'

# ② 拉镜像并启动
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  pull

docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  up -d --no-build
```

**只更新单个服务**

```bash
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  pull ucg-service

docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  up -d --no-build --force-recreate ucg-service
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

> 与测试栈同机时，生产使用 `go-ai-talk-net`、端口 **7001–7006**（Redis）、**5672/15672**（RabbitMQ）、**9701/980x**（微服务）；勿与测试 1700x/5673/197xx 混淆。对照表见 [附录：生产/测试对照](#附录生产测试对照)。

```bash
cd /www/wwwroot/go/go_ai_talk   # 部署目录（仓库根；勿 cd manifest/docker）

# 1) 网络（Redis / RabbitMQ / 生产微服务共用）
docker network create go-ai-talk-net

# 2) MySQL：创建各 ai_voice_* 生产库（history / device / voice / worker / app / ucg 等）

# 3) 静态目录
sudo mkdir -p /ai_talk_images /apk/ai_talk && sudo chmod 755 /ai_talk_images /apk/ai_talk

# 4) 生产 Redis Cluster（compose project：go-ai-talk-redis；宿主机 7001–7006）
docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.yml ps   # 6 节点均须 running

# 先验收集群是否已可用（见附录「Redis Cluster 验收」）；cluster_state:ok 则跳过 cluster create
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state

# 仅当上一步非 cluster_state:ok（首次初始化或 down -v 重置卷后）才执行：
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
  --cluster-replicas 1 --cluster-yes

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok

# 5) 生产 RabbitMQ（compose project：go-ai-talk-rabbitmq；5672 AMQP / 15672 管理台）
docker compose -f manifest/docker/docker-compose.rabbitmq.yml up -d --force-recreate
chmod +x hack/rabbitmq-init.sh && ./hack/rabbitmq-init.sh
# 管理台 http://127.0.0.1:15672  默认 guest/guest
# 容器已起、仅补 exchange/队列拓扑：SKIP_UP=1 ./hack/rabbitmq-init.sh

# 6) 环境文件
cp manifest/docker/.env.prod.example manifest/docker/.env.prod
# 填写 REGISTRY、IMAGE_TAG、各 *_DB_LINK、生产密钥（勿提交 git）

# 7) 首次启动生产微服务 — 见 C.2 步骤 4（docker login ACR → pull → up --no-build）
```

**日常：仅重启生产中间件**（不删 volume；Redis **不要**再 cluster create）

```bash
cd /www/wwwroot/go/go_ai_talk

docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# cluster_state:ok 即可

docker compose -f manifest/docker/docker-compose.rabbitmq.yml up -d --force-recreate
# rabbitmq-init 仅在新数据卷或管理台缺 voice.events 时执行
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
cd /path/to/deploy   # 仓库根目录

# ① 清理（Conflict 时必做）— 见上文「④ 生产微服务」
docker rm -f go-ai-talk-gateway go-ai-talk-gateway-app go-ai-talk-history-service \
  go-ai-talk-voice-service go-ai-talk-device-service go-ai-talk-worker \
  go-ai-talk-ucg-service 2>/dev/null
docker ps -a --format '{{.Names}}' | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|worker|ucg-service)$' \
  && echo '仍有残留' || echo 'OK: 生产微服务容器已清空'

# ② 拉镜像并启动
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

## D. 常见问题与排错（同机双栈部署）

> **工作目录**：所有 `docker compose` 命令须在**仓库根目录**执行（如 `/www/wwwroot/go/go_ai_talk`）。  
> 若 `cd manifest/docker` 后再跑 compose，project 名会变成 **`docker`**，留下 `docker-redis-node-*` 等孤儿容器，与正式 project（`go-ai-talk-redis-test` 等）冲突。

### D.1 生产 / 测试微服务互相把对方停掉

**现象**：执行 prod 或 test 的 `up -d --no-build` 后，另一环境容器消失或重启。

**原因**：两环境共用默认 Compose project 名（目录名 `go_ai_talk`），service 名相同（`gateway`、`device-service` …），后执行的 `up` 会接管同一 project。

**解决**：

- 生产必须叠加 `microservices.prod.yml`（project **`go-ai-talk-prod`**）。
- 测试必须叠加 `microservices.test.yml`（project **`go-ai-talk-test`**）。
- 验收：`docker ps | grep go-ai-talk` 应同时存在 `go-ai-talk-gateway` 与 `go-ai-talk-gateway-test` 等。

**清理旧默认 project**（确认新栈健康后）：

```bash
docker compose -p go_ai_talk \
  -f manifest/docker/docker-compose.microservices.yml down
# 勿加 -v，除非明确要删卷
```

---

### D.2 ACR 拉镜像 `pull access denied`

**现象**：

```text
pull access denied for crpi-xxx-vpc.../pangbao-test/gateway:develop
repository does not exist or may require 'docker login'
```

**原因（按优先级）**：

1. ECS 未 `docker login`
2. GitHub Actions 尚未成功 push 到对应命名空间
3. `.env` 中 `REGISTRY` / `IMAGE_TAG` 与 ACR 控制台不一致
4. 测试/生产命名空间混用

**解决**：

```bash
# 1) 登录（VPC 域名，与 .env REGISTRY 主机一致）
docker login crpi-xxx-vpc.cn-hangzhou.personal.cr.aliyuncs.com \
  -u '<ACR_USERNAME>' -p '<ACR_PASSWORD>'

# 2) 单镜像试拉
docker pull crpi-xxx-vpc.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test/gateway:develop

# 3) 核对 .env.test
grep -E '^REGISTRY=|^IMAGE_TAG=' manifest/docker/.env.test
```

**GitHub Actions push 侧**（与 pull 分离）：

| 配置项 | 要求 |
|--------|------|
| `ACR_REGISTRY_TEST` / `ACR_REGISTRY_PROD` | **公网域名**，**禁止** `-vpc` |
| 命名空间 | 与 `.env.test` / `.env.prod` 的 `REGISTRY` 中 `/pangbao-test` 等一致 |
| ACR 控制台 | 对应命名空间下已有 7 个仓库且存在 `:develop` 或 `:v*` tag |

---

### D.3 微服务启动报 Redis `no such host`（redis-node-1 / redis-node-3）

**现象**：

```text
lookup redis-node-3 on 127.0.0.11:53: no such host
dependency check failed: redis dependency check failed
```

**原因**：业务容器在 `go-ai-talk-test-net`（或 `go-ai-talk-net`），但对应环境的 Redis 未启动，或不在同一 Docker 网络。

**解决**：

```bash
# 测试栈：先起 Redis，再起微服务
docker network create go-ai-talk-test-net 2>/dev/null || true
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml up -d
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml ps

# 容器内 DNS 验收（测试 device-service）
docker exec go-ai-talk-device-service-test getent hosts redis-node-1
```

测试微服务须带 **`microservices.test.yml`** overlay，否则仍在 `go-ai-talk-net`，解析不到测试 Redis。

---

### D.4 测试 Redis：`service "redis-node-1" is not running`

**现象**：未 `up` 或 `up` 失败时直接 `exec … cluster create`。

**解决**：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml ps -a
# 6 行均为 running 后再验收集群（见附录）；非 ok 才 cluster create
```

`logs` 无输出且 `ps` 为空 → 容器从未创建，检查 `up` 报错（网络、端口）。

---

### D.4b 测试/生产 Redis：`Node … is not empty`（重复 cluster create）

**现象**：

```text
[ERR] Node redis-node-1:7001 is not empty. Either the node already knows other nodes … or contains some key in database 0.
```

**原因**：数据卷里**已有集群元数据或 key**。`cluster create` 只对**空节点**执行一次；`up --force-recreate` **不会**清空 volume，日常重启后**不应**再跑 create。

**解决**：**先查集群是否已可用**（见 [附录：Redis Cluster 验收](#redis-cluster-验收与-cluster-create)）：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
```

- **`cluster_state:ok`** → **忽略上述报错，不要 cluster create**，直接继续 RabbitMQ / 微服务。
- **非 ok** 且确需重建 → 测试环境可 `down -v` 清空卷后重新 `up` + `cluster create`（见附录）。

---

### D.5b 生产 `9804` / `9701` 等端口被测试栈占用

**现象**：

```text
Bind for 0.0.0.0:9804 failed: port is already allocated
```

或 test ucg 显示 **同时** `9804` 与 `19804` 映射。

**原因（已修复）**：旧版基线含 `9804:9804`，test overlay 再写 `19804:9804` 会**叠加**。现改为：**基线无 ports**，prod/test 各自 overlay 单独定义端口。

**排查**：

```bash
docker ps --filter 'publish=9804' --format '{{.Names}} {{.Ports}}'
```

- `0.0.0.0:9804->9804` 且容器名含 **-test** → 仍用旧 compose，需 `git pull` 后重建 test 栈  
- 仅 `9804/tcp`（无 `0.0.0.0:`）→ **未占宿主机 9804**，可直接启生产；但 test 需重建以绑定 **19804**

**解决**（拉最新 compose 后）：

```bash
git pull

# 删测试微服务（③ 清理块）后重建
docker rm -f go-ai-talk-gateway-test go-ai-talk-gateway-app-test \
  go-ai-talk-history-service-test go-ai-talk-voice-service-test \
  go-ai-talk-device-service-test go-ai-talk-worker-test \
  go-ai-talk-ucg-service-test 2>/dev/null

docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  up -d --no-build --force-recreate

# 验：仅 19804，无 9804
docker ps --filter name=go-ai-talk-ucg-service-test --format '{{.Ports}}'
# 期望：0.0.0.0:19804->9804/tcp

# 启生产
docker compose --env-file manifest/docker/.env.prod \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.prod.yml \
  up -d --no-build
```

**合并结果自检**（可选）：

```bash
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  config | grep -A2 'ucg-service:' | grep published
# 期望只有 published: 19804
```

---

### D.5 测试 Redis：端口 `1700x already allocated`

**现象**：

```text
Bind for 0.0.0.0:17004 failed: port is already allocated
```

**原因**：宿主机 **17001–17006** 已被占用。常见为旧 project **`docker`** 留下的容器（在 `manifest/docker/` 目录误跑 compose 时产生）：

```text
docker-redis-node-1-1   0.0.0.0:17001->7001/tcp
docker-redis-node-4-1   0.0.0.0:17004->7004/tcp
…
```

与新 project `go-ai-talk-redis-test-redis-node-*`（状态 `Created` 无法启动）并存。

**排查**：

```bash
docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep -E '1700[1-6]|redis-node'
ss -tlnp | grep -E '1700[1-6]'
```

**解决**：执行上文 **[① 测试 Redis](#-测试-redis1700x-端口占用--redis-node-created)**，看到 `OK` 后再 `up`：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml up -d --force-recreate
```

**说明**：测试微服务走 Docker 网内 `redis-node-1:7001`，不依赖宿主机 1700x；1700x 仅用于宿主机调试。生产 Redis 用 **7001–7006**，与 1700x 无关。

节点 **running** 后，**先验收集群**（[附录：Redis Cluster 验收](#redis-cluster-验收与-cluster-create)）。**若 `cluster_state:ok`，不要执行 cluster create**；仅首次或 `down -v` 重置卷后才 create：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# cluster_state:ok → 跳过 create

docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
  --cluster-replicas 1 --cluster-yes
```

---

### D.6 测试 RabbitMQ / 微服务：容器名冲突

**现象**：

```text
The container name "/go-ai-talk-rabbitmq-test" is already in use
The container name "/go-ai-talk-history-service-test" is already in use
```

**原因**：旧 compose project（`docker`、`go_ai_talk`）留下的同名容器。

**解决**：RabbitMQ 用 **[②](#-测试-rabbitmqgo-ai-talk-rabbitmq-test-冲突)**，微服务用 **[③](#-测试微服务go-ai-talk--test-冲突不影响生产)**，验收到 `OK` 后再 `up`。

---

### D.7 反复试错后一键清空测试栈容器

按顺序执行 **① → ② → ③**（见 [启动前清理](#启动前清理遇-conflict--端口占用)），再按 B.1 / D.8 启动。

**总览**（仍存活时排查）：

```bash
docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' \
  | grep -E 'go-ai-talk|docker-redis-node'
```

---

### D.8 推荐启动顺序（测试栈首次 / 重建）

```bash
cd /www/wwwroot/go/go_ai_talk
docker network create go-ai-talk-test-net 2>/dev/null || true

# 清理：依次 ① Redis → ② RabbitMQ → ③ 微服务（各块见「启动前清理」）
# … 复制 ①②③ 的 docker rm + 验收命令 …

# 启动 Redis → 验收集群（ok 则跳过 create）→ RabbitMQ → init → 微服务（见 B.1 步骤 5–6 与 B.2 步骤 3）
```

完整命令见 **B.1**；Conflict 时仅在对应 `up` 前插入 **① / ② / ③** 各 2 行（删 + 验）。

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
- gateway-app：`GATEWAY_APP_PUBLIC_BASE_URL`（APK 下载绝对 URL + CORS 白名单 hostname）、APK 上传需 Nginx `client_max_body_size` ≥ 220MB。

### Web 管理页与环境隔离

静态页在 `resource/public/`，经主网关（`:9701` / 测试 `:19701`）或 App 网关提供。**业务 API 须与当前页面同源或配对端口**，域名由宝塔 Nginx + 服务器 `.env` 配置，**不写入 compose/代码**。

| 页面 | 路径 | API 请求方式 |
|------|------|----------------|
| 设备管理 | `/device/admin` | 同源相对路径 `/device/admin/api/*` |
| 历史记录 | `/device/history/{deviceNo}` | 同源 `/device/history/api/*` |
| 问答库 | `/device/admin/qa-records` | 同源 `/device/admin/api/qa/*` |
| 版本管理 | `/device/app/version-admin.html` | 同源 `/device/app/api/version/admin/*` |
| App 联调 | `/device/app/integration-test.html` | 默认 `window.location.origin`（可手改 baseUrl） |

**易错点**：设备管理页跳转「App 版本管理」时，须打开**同环境** App 网关（9701→9702、19701→19702），已在 `admin.html` 按当前端口配对。

**验收**（测试环境示例）：

```bash
# 1) 浏览器打开测试设备管理（Nginx 或直连 19701）
# 2) 开发者工具 → 网络：/device/admin/api/* 的 Host 须为测试域名或 :19701，不得出现 :9702（生产直连端口）
# 3) 点击「App 版本管理」：地址栏应为 :19702 或测试域名 :9702（Nginx 反代），不得为生产宿主机 9702
# 4) gateway-app CORS：.env.test 中 GATEWAY_APP_PUBLIC_BASE_URL 填测试对外 HTTPS 基址
docker exec go-ai-talk-gateway-app-test printenv GATEWAY_APP_PUBLIC_BASE_URL
```
- device logo：`/ai_talk_images`；测试栈用 `/ai_talk_images_test`。

### 生产/测试对照

| 项 | 生产 | 测试 |
|----|------|------|
| Compose project（微服务） | `go-ai-talk-prod` | `go-ai-talk-test` |
| Compose project（Redis） | `go-ai-talk-redis` | `go-ai-talk-redis-test` |
| Compose project（RabbitMQ） | `go-ai-talk-rabbitmq` | `go-ai-talk-rabbitmq-test` |
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
| `docker-compose.microservices.yml` | 基线拓扑（无宿主机 ports）；本地 `--build` |
| `docker-compose.microservices.local.yml` | 本地端口 9701/980x |
| `docker-compose.microservices.test.yml` | 测试 overlay + 端口 197xx |
| `docker-compose.microservices.prod.yml` | 生产 overlay + 端口 9701/980x |

镜像引用：`${REGISTRY}/gateway:${IMAGE_TAG}`。`REGISTRY` = `<ACR域名>/<命名空间>`；仓库名单段（`gateway`、`device-service` 等，无 `go-ai-talk/` 前缀）。

### Redis Cluster 验收与 cluster create

> **`cluster create` 仅对空节点执行一次。** 数据卷保留时，重复执行会报 `Node … is not empty` 或 `already knows other nodes`——这通常表示**集群已初始化**，不是故障。

**MUST 先检查**；若已可用则 **不要** 再跑 `cluster create`，直接继续 RabbitMQ / 微服务。日常 `up --force-recreate` 只重建容器，**不**清空 volume，**无需**重新 create。

#### 1. 查看集群是否已可用

**生产**（`docker-compose.redis-cluster.yml`）：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# cluster_state:ok → 已就绪，跳过 cluster create

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER NODES
# 期望 6 个节点（3 主 3 从）
```

**测试**（`docker-compose.redis-cluster.test.yml`）：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# cluster_state:ok → 已就绪，跳过 cluster create

docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER NODES
# 期望 6 个节点
```

#### 2. 仅当 `cluster_state` 非 ok 时：首次 cluster create

6 节点均为 **running** 且上一步**不是** `cluster_state:ok` 时（全新 volume 或重置后）才执行：

```bash
# 生产
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
  --cluster-replicas 1 --cluster-yes

# 测试（compose 文件换 .test.yml）
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  redis-node-4:7004 redis-node-5:7005 redis-node-6:7006 \
  --cluster-replicas 1 --cluster-yes
```

create 后再验：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok
```

#### 3. 重置集群（会清空 Redis 数据，仅测试环境或明确维护窗口）

```bash
# 测试示例
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml down -v
docker compose -f manifest/docker/docker-compose.redis-cluster.test.yml up -d --force-recreate
# 再按「2. 首次 cluster create」执行
```

生产重置同理，换 `docker-compose.redis-cluster.yml`；**须评估业务影响**。

相关排错：[D.4b](#d4b-测试生产-redisnode--is-not-empty重复-cluster-create)。

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

push 报 `denied` 常见原因：① GitHub 用了 `-vpc` 域名；② 缺命名空间；③ 测试/生产 Secret 与 `.env` 命名空间不一致；④ 仓库未创建；⑤ ECS 未 `docker login`。详见 [D.2](#d2-acr-拉镜像-pull-access-denied)。

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
