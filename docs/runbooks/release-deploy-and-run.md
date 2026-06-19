## 部署与运行指南

适用范围：gateway / gateway-app / voice-service / device-service / history-service / ucg-service / **sim-user-service**。

**Redis 容灾与恢复**（容器重启、volume 备份/还原、数据分层）：见 [redis-disaster-recovery.md](./redis-disaster-recovery.md)。

**三种运行方式**

| 方式 | 何时用 | 镜像来源 | 服务器是否需要完整源码 |
|------|--------|----------|------------------------|
| **A. 本地开发** | 本机改代码联调 | 本机 `docker compose --build`（`:local`） | 要（整仓） |
| **B. 测试环境** | 打预发布 tag 后在 test 域名验收 | ACR `:v*-rc.*` 等（GitHub Actions 构建） | **不要**（仅需 compose + `.env.test`） |
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
| 测试 Redis | `docker-compose.redis-standalone.test.yml` | `go-ai-talk-redis-test` |
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

#### ① 测试 Redis（16379 端口占用 / 旧 6 节点 cluster 残留）

```bash
# 删：旧 6 节点 test cluster（1700x）+ standalone 容器 + 误跑 project「docker」残留
docker rm -f \
  go-ai-talk-redis-test \
  docker-redis-node-1-1 docker-redis-node-2-1 docker-redis-node-3-1 \
  docker-redis-node-4-1 docker-redis-node-5-1 docker-redis-node-6-1 \
  go-ai-talk-redis-test-redis-node-1-1 go-ai-talk-redis-test-redis-node-2-1 \
  go-ai-talk-redis-test-redis-node-3-1 go-ai-talk-redis-test-redis-node-4-1 \
  go-ai-talk-redis-test-redis-node-5-1 go-ai-talk-redis-test-redis-node-6-1 \
  2>/dev/null

# 验：无输出 = 容器已空
docker ps -a --format '{{.Names}}' | grep -E 'redis-test|redis-node|go-ai-talk-redis-test' \
  || echo 'OK: 测试 Redis 容器已清空'
ss -tlnp | grep -E '16379|1700[1-6]' || echo 'OK: 16379/1700x 端口已释放'
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
  go-ai-talk-device-service-test \
  go-ai-talk-ucg-service-test \
  go-ai-talk-sim-user-service-test \
  2>/dev/null

# 验：只查 7 个微服务名（go-ai-talk-rabbitmq-test 等中间件不算残留）
docker ps -a --format '{{.Names}}' \
  | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|ucg-service|sim-user-service)-test$' \
  && echo '仍有残留，见上表' || echo 'OK: 测试微服务容器已清空'
```

#### ④ 生产微服务（**勿删 -test 容器**）

```bash
# 删：7 个生产微服务固定容器名（无 -test 后缀）
docker rm -f \
  go-ai-talk-gateway go-ai-talk-gateway-app go-ai-talk-history-service \
  go-ai-talk-voice-service go-ai-talk-device-service \
  go-ai-talk-ucg-service go-ai-talk-sim-user-service \
  2>/dev/null

# 验：无输出 = 已删（不应出现 -test 行）
docker ps -a --format '{{.Names}}' | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|ucg-service|sim-user-service)$' \
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
docker compose -f manifest/docker/docker-compose.redis-cluster.yml ps   # 3 节点均须 running

# 先验集群是否已可用（见附录「Redis Cluster 验收」）；cluster_state:ok 则跳过 cluster create
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state

# 仅当上一步非 cluster_state:ok（首次初始化或 down -v 重置卷后）才执行：
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  --cluster-replicas 0 --cluster-yes

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok

```

**2. MySQL**（宿主机上）

```bash
systemctl start mysql-local   # 或你的 mysqld 服务名
systemctl restart mysql-local
systemctl status mysql-local.service
```

**3. 环境变量**

编辑 `manifest/docker/.env.local`（本机开发专用，与 `.env.prod` / `.env.test` 隔离）：

- **`MYSQL_TCP_HOST`**：MySQL 主机 IP
- **`*_DB_LINK`**：6 条 DSN，host 写占位符 `mysql-host`
- **`GF_REDIS_DEFAULT_ADDRESS`**：本机 Redis（Cluster 三主种子或 standalone 单地址）
- **`GATEWAY_APP_JWT_SECRET`**：App JWT 签名密钥（gateway-app 签发、ucg 校验须同值）
- **`UCG_OSS_ACCESS_KEY_ID` / `UCG_OSS_ACCESS_KEY_SECRET`**：ucg OSS 直传与 Green 审核（Green 复用 OSS AK）；`config.ucg-service.yaml` 留空，见 `manifest/docker/.env.example`
- **`UCG_DASHSCOPE_API_KEY`**：ucg AI 润笔（DashScope）；yaml 中 `dashscope_api_key` 留空
- **`UCG_APNS_*` / `UCG_HMS_*` / `UCG_MIPUSH_*`**：ucg 启动器角标推送（iOS APNs、华为 HMS、小米 MiPush）；yaml `ucg.push` 留空，见 `manifest/docker/.env.example`；Flutter 客户端见 `flutter_ai_talk/app/README.md`「UCG 启动器角标推送」
- **`GLM_API_KEY`**：智谱 GLM（voice 喂养/clinic 默认种子、ucg 润笔 zhipu provider）；**生产部署前必须配置**
- **`DEEPSEEK_API_KEY`**（可选）：覆盖 `voice-chat.shared.yaml` 中 deepseek 段；Admin 切回 deepseek provider 时需配置

**LLM lane 数据表**（首次部署 voice/ucg 前在对应库执行；进程启动时 `EnsureDefaultRows` 会补种子行，但表须已存在）：

```sql
-- ai_voice_voice
CREATE TABLE IF NOT EXISTS llm_lane_config (
  lane VARCHAR(32) PRIMARY KEY,
  provider VARCHAR(32) NOT NULL,
  model VARCHAR(64) NOT NULL,
  max_in_flight INT NOT NULL DEFAULT 1,
  max_waiters INT NOT NULL DEFAULT 0,
  updated_at BIGINT NOT NULL DEFAULT 0,
  updated_by VARCHAR(64) NOT NULL DEFAULT ''
);

-- ai_voice_ucg：扩展 ucg_ai_config（若表已存在则 ALTER）
ALTER TABLE ucg_ai_config
  ADD COLUMN provider VARCHAR(32) NOT NULL DEFAULT 'zhipu' AFTER vision_model,
  ADD COLUMN max_in_flight INT NOT NULL DEFAULT 1 AFTER max_images_per_request,
  ADD COLUMN max_waiters INT NOT NULL DEFAULT 15 AFTER max_in_flight;
```

Admin 热更新：`GET/PUT /voice/admin/api/llm-lanes`（voice-admin.html「LLM 车道」Tab）、扩展后的 `GET/PUT /ucg/admin/api/ai-config`。

**验收（ALTER 后 + 新镜像）**：

```bash
# ucg 仍 running（ExitCode=0；勿 stack overflow 退出）
docker inspect go-ai-talk-ucg-service-test --format 'Status={{.State.Status}} Restarts={{.RestartCount}}'

# 直连 ucg PUT ai-config（口令见 .env.test UCG_ADMIN_PASSWORD）
curl -sS -o /dev/null -w "PUT ai-config HTTP %{http_code}\n" \
  -X PUT http://127.0.0.1:19804/ucg/admin/api/ai-config \
  -H "Content-Type: application/json" \
  -H "X-Admin-Password: YOUR_UCG_ADMIN_PASSWORD" \
  -d '{"provider":"zhipu","visionModel":"glm-4.6v-flash","maxImagesPerRequest":9,"maxInFlight":1,"maxWaiters":15,"updatedBy":"runbook"}'

docker ps | grep ucg-service-test   # 仍 Up

# voice-admin 页：Hub 登录后 /device/admin/voice-admin.html 须展示额度与 LLM 车道 Tab
# 空 outbox 时 ucg 日志无周期性 [ucg-chat-persist]/[ucg-audit-outbox] flush/relay tick failed
```

完整变量清单见 **`manifest/docker/.env.example`**。

**4. 静态资源目录**

```bash
sudo mkdir -p /ai_talk_images /apk/ai_talk && sudo chmod 755 /ai_talk_images /apk/ai_talk
```

**5. 可观测性（可选）**

微服务跑通后，可按 [A.4](#a4-可观测性栈可选) 启动 Prometheus / Grafana 等，用于本地看指标与排查。

### A.2 日常：改了代码要跑起来

**清理docker内存***
```bash
docker system prune -a 清理无用数据
# 强制删除所有无用镜像
docker image prune -a -f
```

**全量重建并启动**

```bash
docker compose --env-file manifest/docker/.env.local \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.local.yml \
  up -d --build
```

**只改某一个服务**（更快）

```bash
docker compose --env-file manifest/docker/.env.local \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.local.yml \
  up -d --build voice-service
# 同理：gateway / gateway-app / history-service / device-service / ucg-service
```

**只改了 compose 环境变量、未改代码**

```bash
docker compose --env-file manifest/docker/.env.local \
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
curl -s http://127.0.0.1:9804/api.json   # ucg-service
curl -s http://127.0.0.1:9805/api.json   # sim-user-service
```

**sim-user-service（可选）**：默认 `SIM_USER_SERVICE_ENABLED=false` 仅健康检查；开启后需配置 `SIM_DB_LINK`、`GATEWAY_APP_URL`、`GLM_API_KEY`、`SIM_ADMIN_PASSWORD`（管理页 `/device/admin/sim-admin.html`）。模拟用户 API 不计入 usage 统计（`wx.is_simulated=1`）。

```bash
# 本地启用示例（.env.local）
SIM_USER_SERVICE_ENABLED=true
SIM_ADMIN_PASSWORD=与网关 Hub 口令一致或单独配置
```

### A.4 可观测性栈（可选）

> **仅本地开发调试用**；生产/测试 ECS 栈不依赖此 compose。与微服务独立 project，可随时 `up`/`down`，不影响业务容器。

**用途**：Prometheus 抓指标、Loki 收日志、Tempo 收链路、Grafana 统一查看。配置文件在 `manifest/docker/observability/`（Prometheus 已预配 gateway / history 等 scrape target）。

**前提**：须先按 [A.2](#a2-日常改了代码要跑起来) 启动微服务并暴露宿主机端口（`microservices.local.yml` 的 9701 / 9801 等），Prometheus 经 `host.docker.internal` 抓取。

**启动**

```bash
docker compose -f manifest/docker/docker-compose.observability.yml up -d
docker compose -f manifest/docker/docker-compose.observability.yml ps
```

**停止**

```bash
docker compose -f manifest/docker/docker-compose.observability.yml down
# 保留数据卷（Tempo 容器内 /tmp/tempo）无需 -v；完全重置时自行删容器即可
```

**访问**

| 组件 | 地址 | 说明 |
|------|------|------|
| Grafana | http://127.0.0.1:3000 | 默认账号 `admin` / `admin` |
| Prometheus | http://127.0.0.1:9090 | Targets 页确认 gateway / history 等为 UP |
| Loki | http://127.0.0.1:3100 | 需在 Grafana 添加数据源 `http://loki:3100` |
| Tempo | http://127.0.0.1:3200 | OTLP gRPC 宿主机 **4317**；Grafana 数据源 `http://tempo:3200` |

**Grafana 首次配置（一次性）**

1. 登录 Grafana → **Connections → Data sources**。
2. 添加 **Prometheus**，URL 填 `http://prometheus:9090`（同 compose 网内服务名）。
3. 可选：添加 **Loki** `http://loki:3100`、**Tempo** `http://tempo:3200`（Explore 查 trace）。

**链路追踪（可选）**

本地微服务默认未注入 `OTEL_EXPORTER_OTLP_ENDPOINT`。若需把 trace 写入 Tempo，可在 `docker-compose.microservices.local.yml` 或 `.env.local` 为对应服务设置：

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://host.docker.internal:4317
OTEL_EXPORTER_OTLP_INSECURE=true
```

（与 K8s develop overlay 中 `otel-collector:4317` 等价，本地改为宿主机映射的 Tempo 端口。）

**验收**

```bash
curl -s http://127.0.0.1:9090/-/healthy          # Prometheus
curl -s http://127.0.0.1:3000/api/health       # Grafana
# 浏览器打开 Prometheus → Status → Targets，确认 9701/9801 等为 UP
```

---

## B. 发布测试环境

对外：`https://test.pangbao.cuplay.top:9701` / `:9702`（Nginx 反代至宿主机 **19701 / 19702**）。  
镜像 tag：**与 git 预发布 tag 一致**（如 **`v1.0.0-rc.1`**，写在 `.env.test` 的 **`IMAGE_TAG`**）。  
**停/启全栈**（给生产腾资源）：见 [B.3](#b3-日常停启测试全栈给生产腾资源)。

### Docker 容器日志轮转（prod/test 共用 compose 策略）

长期运行的 prod/test 容器已在 compose 中配置 `json-file` 轮转，避免 `/var/lib/docker/containers/*/*-json.log` 无限增长：

| 栈 | 单文件上限 | 保留文件数 | 约上限/容器 |
|----|-----------|-----------|------------|
| 微服务六件套 + Redis | 10m | 3 | ~30MB |
| RabbitMQ | 20m | 3 | ~60MB |

RabbitMQ 另挂载 `manifest/docker/rabbitmq/rabbitmq.conf`，将 console/connection/channel 日志降到 **warning**（仍保留 alarm 与认证失败）。详见 [附录：Docker 容器日志](#docker-容器日志轮转与验收)。

**变更 compose 后须 `--force-recreate` 才生效**（仅 `up -d` 不会改已有容器的 LogConfig）。推荐顺序：RabbitMQ → Redis → 微服务。

**历史巨型 log**：轮转只限制新写入，不会自动缩小已有文件。优先 `--force-recreate` 删旧容器释放 log；或对已知路径 `truncate -s 0`（见附录）。

### B.1 首次搭建测试栈（一次性，与生产同机且完全隔离）

> 生产已运行时，测试使用独立网络、Redis、RabbitMQ、MySQL `_test` 库、静态目录 `*_test`。对照表见 [附录：生产/测试对照](#附录生产测试对照)。

```bash
# 1) 网络
docker network create go-ai-talk-test-net

# 2) MySQL 建库：ai_voice_history_test、ai_voice_device_test、…（各域 + app + ucg）

# 3) 静态目录
sudo mkdir -p /ai_talk_images_test /apk/ai_talk_test && sudo chmod 755 /ai_talk_images_test /apk/ai_talk_test

# 4) 脱敏种子（可选，从生产导入测试数据）
MYSQL_PASS='***' ./hack/mask-seed-data.sh

# 5) 测试 Redis standalone + RabbitMQ（仓库根目录；Conflict 时先跑上文 ①②）
#    .env.test 须含 GF_REDIS_DEFAULT_ADDRESS=redis-test:6379

docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml ps   # redis-test 须 running

docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test \
  redis-cli PING
# 期望 PONG（无需 cluster create）

docker compose -f manifest/docker/docker-compose.rabbitmq.test.yml up -d --force-recreate
COMPOSE_FILE=manifest/docker/docker-compose.rabbitmq.test.yml \
RABBIT_API_BASE=http://127.0.0.1:15673/api ./hack/rabbitmq-init.sh

# 6) 环境文件
cp manifest/docker/.env.test manifest/docker/.env.test.bak   # 首次可从仓库 .env.test 复制后改口令
# 填写 REGISTRY、*_DB_LINK（指向 *_test 库）、GF_REDIS_DEFAULT_ADDRESS、密钥等
```

### B.2 日常：把改动发布到测试（逐步）

**步骤 1 — 开发机：提交、打预发布 tag 并推送**

```bash
git checkout develop
git add … && git commit -m "…"
git push origin develop

# 测试镜像仅由 tag 触发 CI（develop push 不再构建）
git tag v1.0.0-rc.1    # 全量 7 服务 → 测试 ACR
git push origin v1.0.0-rc.1

# 小改动仅构建单服务（节省 GitHub Actions 分钟）：
git tag v1.0.0-rc.2+ucg    # 仅 build/push ucg-service；git tag 名含 +ucg
git push origin v1.0.0-rc.2+ucg

git tag v1.0.0-rc.3+sim    # 仅 build/push sim-user-service
git push origin v1.0.0-rc.3+sim
```

**步骤 2 — 等待 CI**

打开 GitHub → Actions → `docker-acr` 工作流：

- 无 `+` 后缀 tag（如 `v1.0.0-rc.1`）：确认 **七服务** 镜像均 push 成功。
- 带 `+` 后缀 tag（如 `v1.0.0-rc.2+ucg`）：确认 **仅列出的服务** build 成功；ACR **不会**为未构建服务创建该 tag 的镜像（**不 retag**）。

> **路由规则**：`vMAJOR.MINOR.PATCH`（如 `v1.0.0`）→ 生产 ACR；`v1.0.0-rc.1` 等预发布 tag → 测试 ACR。环境路由看 **`+` 前的 base tag**（`v1.0.0-rc.2+ucg` 仍走测试）。

**步骤 3 — 测试服务器：改 tag、拉镜像并更新**

编辑 `manifest/docker/.env.test`：

```bash
IMAGE_TAG=v1.0.0-rc.1   # 或 v1.0.0-rc.2（与 git tag 的 + 前 base 一致，不含 +ucg）
```

> **git tag 与 IMAGE_TAG**：push `v1.0.0-rc.2+ucg` 时，`.env` 中 `IMAGE_TAG` 写 **`v1.0.0-rc.2`**，不要写 `+ucg` 后缀。

**全量发版**（七服务均在新 tag 下 rebuild 过）— 与原先相同：

```bash
# 强制删除悬空镜像
docker image prune -f

# 强制删除所有无用镜像
docker image prune -a -f

cd /www/wwwroot/go/go_ai_talk/   # 仓库根目录

# ① 清理（Conflict 时必做）— 见上文「③ 测试微服务」
docker rm -f go-ai-talk-gateway-test go-ai-talk-gateway-app-test \
  go-ai-talk-history-service-test go-ai-talk-voice-service-test \
  go-ai-talk-device-service-test \
  go-ai-talk-ucg-service-test go-ai-talk-sim-user-service-test 2>/dev/null
docker ps -a --format '{{.Names}}' \
  | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|ucg-service|sim-user-service)-test$' \
  && echo '仍有残留' || echo 'OK: 测试微服务容器已清空'

# ② 拉镜像并启动
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  pull

docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  up -d --no-build
```

**只更新单个服务**（配合 CI tag `v1.0.0-rc.2+ucg` 等部分构建）

部分构建后 ACR **仅有**变更服务的 `:${IMAGE_TAG}` 镜像。须 **按服务** pull/up；对全栈执行 `compose pull` **会因缺 gateway 等镜像而失败**（表示应用了部分 tag 却做了全量拉取）。

```bash
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  pull ucg-service

docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  up -d --no-build --force-recreate ucg-service
```

**步骤 4 — 验收**

```bash
curl -sk https://test.pangbao.cuplay.top:9702/api.json
curl -s http://127.0.0.1:19701/api.json

# 环境变量：须含 _test 库名
docker exec go-ai-talk-history-service-test printenv HISTORY_DB_LINK
docker exec go-ai-talk-gateway-app-test printenv APP_DB_LINK

# 启动日志：须出现 dbcfg 配置行（未设置 *_DB_LINK 时进程无法连库）
docker logs go-ai-talk-history-service-test 2>&1 | grep "database.default 已用 HISTORY_DB_LINK"
docker logs go-ai-talk-gateway-app-test 2>&1 | grep "database.app 已用 APP_DB_LINK"

# 官网聚合：appDatabase 须为 ai_voice_app_test（非 ai_voice_app）
curl -sk https://test.pangbao.cuplay.top:9702/device/app/api/site/home | grep appDatabase

# 环境隔离一键验收（Redis/MySQL/对外入口，只读）
chmod +x hack/env-isolation-check.sh && ./hack/env-isolation-check.sh
```

> **重要**：`*_DB_LINK` 生效依赖 **含 `internal/platform/dbcfg` 的新镜像**。仅 `git pull` + `up --no-build` 不会修复 test 误连生产库；须 CI 打预发布 tag 或本地 `--build` 后 `--force-recreate` 全量微服务。

**排错：pin 某次构建** — 将 `.env.test` 中 `IMAGE_TAG` 改为 Actions 推送的 **git 完整 sha**，再 `pull` + `up --no-build --force-recreate`。

### B.3 日常：停/启测试全栈（给生产腾资源）

> **适用**：同机双栈时，暂时不用 test 域名验收、只想给 **生产** 腾内存/CPU；或发版前暂停 test 避免 ASR 与 prod 争抢。  
> **原则**：只动 **test** 三个 Compose project（`go-ai-talk-test` / `go-ai-talk-redis-test` / `go-ai-talk-rabbitmq-test`），**勿**对生产 project 执行 `down`。  
> **数据**：下列 `down` **均不加 `-v`**，Redis/RabbitMQ/MySQL `_test` 数据保留。

**工作目录**（以下命令前先执行）：

```bash
cd /www/wwwroot/go/go_ai_talk
```

#### 停止测试环境（推荐顺序：微服务 → RabbitMQ → Redis）

先停依赖中间件的 7 个微服务，再停中间件。

```bash
# 1) 测试微服务（project=go-ai-talk-test）
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  down

# 2) 测试 RabbitMQ（project=go-ai-talk-rabbitmq-test）
docker compose -f manifest/docker/docker-compose.rabbitmq.test.yml down

# 3) 测试 Redis（project=go-ai-talk-redis-test）
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml down
```

**验收（应无 test 容器在跑；生产不受影响）**：

```bash
docker ps --format 'table {{.Names}}\t{{.Status}}' | grep -E 'test|1970|1980|19901|16379|5673' \
  && echo '仍有 test 容器' || echo 'OK: 测试栈已停止'

docker ps --format 'table {{.Names}}\t{{.Status}}' | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|ucg-service|sim-user-service)$' \
  | head -8
# 应仍看到 7 个生产微服务 Up（无 -test 后缀）
```

#### 恢复测试环境（推荐顺序：网络 → Redis → RabbitMQ → 微服务）

与 [B.1](#b1-首次搭建测试栈一次性与生产同机且完全隔离) 启动顺序一致；中间件卷未删时 **无需** 再 `cluster create`，Rabbit **通常无需** 重跑 init（仅新 volume 或管理台缺队列时执行 init）。

```bash
# 0) 网络（down 不会删 external network，重复 create 无害）
docker network create go-ai-talk-test-net 2>/dev/null || true

# 1) Redis standalone
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test redis-cli PING
# 期望 PONG

# 2) RabbitMQ
docker compose -f manifest/docker/docker-compose.rabbitmq.test.yml up -d --force-recreate
# 仅首次或 down -v 后需要：
# COMPOSE_FILE=manifest/docker/docker-compose.rabbitmq.test.yml \
#   RABBIT_API_BASE=http://127.0.0.1:15673/api ./hack/rabbitmq-init.sh

# 3) 测试微服务（须 .env.test 中 IMAGE_TAG 等与上次一致）
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  pull

docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  up -d --no-build --force-recreate
```

**验收**：

```bash
docker ps --format 'table {{.Names}}\t{{.Status}}' \
  | grep -E 'go-ai-talk-.*-test|go-ai-talk-redis-test|go-ai-talk-rabbitmq-test'

curl -sk https://test.pangbao.cuplay.top:9702/api.json
curl -sk https://test.pangbao.cuplay.top:9701/api.json
```

#### 仅停/启测试微服务（中间件常开）

若只需短暂释放 voice/gateway 等占用的内存，**Redis + RabbitMQ 保持 Up**，只 down/up 微服务 project 即可（命令同上 **停止** 第 1 步与 **恢复** 第 3 步）。

#### 勿用 / 注意

| 操作 | 说明 |
|------|------|
| `down -v` | **会删** test Redis/Rabbit 数据卷；除非刻意重置 test 中间件，否则禁止 |
| `docker rm -f …-test` | 仅 [Conflict 清理](#启动前清理遇-conflict--端口占用) 时用；正常停栈用 `compose down` |
| 生产 compose | 勿对 `redis-cluster.yml` / `rabbitmq.yml` / `microservices.prod.yml` 执行 test 停栈时的 `down` |
| MySQL `_test` 库 | 停容器 **不删库**；数据仍在 ECS 上 MySQL |

---

## C. 发布生产环境

对外：`www.pangbao.cuplay.top:9701` / `:9702`。  
镜像 tag：与 git tag 一致（如 **`v1.0.0`**），写在 `manifest/docker/.env.prod` 的 **`IMAGE_TAG`**。

### C.1 首次搭建生产栈（一次性）

> 与测试栈同机时，生产使用 `go-ai-talk-net`、端口 **7001–7003**（Redis Cluster）、**5672/15672**（RabbitMQ）、**9701/980x**（微服务）；测试 Redis standalone **16379**、5673/197xx。对照表见 [附录：生产/测试对照](#附录生产测试对照)。

```bash
cd /www/wwwroot/go/go_ai_talk   # 部署目录（仓库根；勿 cd manifest/docker）

# 1) 网络（Redis / RabbitMQ / 生产微服务共用）
docker network create go-ai-talk-net

# 2) MySQL：创建各 ai_voice_* 生产库（history / device / voice / app / ucg 等）

# 3) 静态目录
sudo mkdir -p /ai_talk_images /apk/ai_talk && sudo chmod 755 /ai_talk_images /apk/ai_talk

# 4) 生产 Redis Cluster（compose project：go-ai-talk-redis；宿主机 7001–7003，3 主 0 从）
docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.yml ps   # 3 节点均须 running

# 先验集群是否已可用（见附录「Redis Cluster 验收」）；cluster_state:ok 则跳过 cluster create
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state

# 仅当上一步非 cluster_state:ok（首次初始化或 down -v 重置卷后）才执行：
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  --cluster-replicas 0 --cluster-yes

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok

# 5) 生产 RabbitMQ（compose project：go-ai-talk-rabbitmq；5672 AMQP / 15672 管理台）
docker compose -f manifest/docker/docker-compose.rabbitmq.yml up -d --force-recreate
chmod +x hack/rabbitmq-init.sh && ./hack/rabbitmq-init.sh
# 管理台 http://127.0.0.1:15672  默认 guest/guest
# 容器已起、仅补 exchange/队列拓扑：SKIP_UP=1 ./hack/rabbitmq-init.sh

# 6) 环境文件
cp manifest/docker/.env.prod manifest/docker/.env.prod.bak   # 首次可从仓库 .env.prod 复制后改口令
# 填写 REGISTRY、IMAGE_TAG、各 *_DB_LINK、生产密钥（勿提交 git）

# 7) 首次启动生产微服务 — 见 C.2 步骤 4（docker login ACR → pull → up --no-build；须叠加 resources.prod.yml）
```

**从旧 6 节点生产 Redis 迁移至 3 主 0 从**（维护窗口；**会清空 Redis 缓存**）：

```bash
cd /www/wwwroot/go/go_ai_talk
git pull   # 获取新版 docker-compose.redis-cluster.yml

docker compose -f manifest/docker/docker-compose.redis-cluster.yml down -v
docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-cluster.yml ps   # 3 节点 running

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  --cluster-replicas 0 --cluster-yes

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok；CLUSTER NODES 应仅 3 个 master
```

**从旧 6 节点测试 Redis Cluster 迁移至 standalone**：

```bash
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml down -v 2>/dev/null || true
# 执行上文 ① 清理块，释放 1700x / 旧容器

docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test redis-cli PING
# PONG

# .env.test 增加 GF_REDIS_DEFAULT_ADDRESS=redis-test:6379 后 recreate 测试微服务（含 resources.test.yml）
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
  go-ai-talk-voice-service go-ai-talk-device-service \
  go-ai-talk-ucg-service go-ai-talk-sim-user-service 2>/dev/null
docker ps -a --format '{{.Names}}' | grep -E '^go-ai-talk-(gateway|gateway-app|history-service|voice-service|device-service|ucg-service|sim-user-service)$' \
  && echo '仍有残留' || echo 'OK: 生产微服务容器已清空'

# ② 拉镜像并启动
docker compose --env-file manifest/docker/.env.prod \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.prod.yml \
  -f manifest/docker/docker-compose.resources.prod.yml \
  pull

docker compose --env-file manifest/docker/.env.prod \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.prod.yml \
  -f manifest/docker/docker-compose.resources.prod.yml \
  up -d --no-build
```

**步骤 5 — 验收**

```bash
curl -s http://127.0.0.1:9701/api.json
curl -s http://127.0.0.1:9702/api.json
curl -s http://127.0.0.1:9804/api.json
```

### C.3 生产回滚

1. 将 `.env.prod` 中 `IMAGE_TAG` 改回上一稳定版本（如 `v0.9.0`）。
2. `pull` + `up -d --no-build --force-recreate`（可按单服务）。
3. **禁止**对生产执行 `docker system prune -a`。

---

### C.4 2C2G ECS 双栈 survival（内存与 limits）

适用：**2 核 2G** 同机跑 MySQL + 生产 + 测试双栈。limits 总和可大于 2G（防单容器暴涨）；依赖 **swap** 与 MySQL 调优兜底。

**MySQL（宿主机）建议**

```ini
# my.cnf 片段（示例；按实例调整）
innodb_buffer_pool_size = 256M
max_connections = 100
```

**宿主机 swap**

```bash
# 建议额外 1G swap（若尚未配置）
sudo fallocate -l 1G /swapfile && sudo chmod 600 /swapfile
sudo mkswap /swapfile && sudo swapon /swapfile
```

**双栈内存粗算**（limits 上限，非实际占用）

| 类别 | 约计 |
|------|------|
| MySQL | ~400–512M（含 buffer_pool） |
| 生产 Redis×3 + Rabbit | ~480M |
| 测试 redis-test + Rabbit | ~288M |
| 生产 7 微服务 | ~1.5G limits 合计 |
| 测试 7 微服务（voice 512M） | ~1.7G limits 合计 |

**容器资源 limits 默认值**（见各 compose；微服务在 `resources.{prod,test}.yml`，Redis/Rabbit 在各自 compose）

| 组件 | memory | cpus |
|------|--------|------|
| prod redis ×3 | 96m | 0.1 |
| test redis-test | 96m | 0.1 |
| rabbitmq ×2 | 192m | 0.2 |
| voice-test | 512m | 0.8 |
| voice-prod | 256m | 0.3 |
| gateway / gateway-app | 192m | 0.2 |
| 其它微服务 | 128m | 0.15 |

**ASR / WebSocket 验收约定**

- 生产 **7 微服务保持 Up**；真实 ASR 压测 **仅对测试域名** 执行，避免 prod/test 同时高并发语音。
- 验收前确认：`docker stats --no-stream` 中 test voice 未 OOM；prod gateway/voice 仍为 Up。

**OOM 排错**

```bash
docker stats --no-stream | grep -E 'go-ai-talk|redis|rabbit'
dmesg | tail -20 | grep -i oom   # 宿主机是否杀进程
docker inspect go-ai-talk-voice-service-test --format '{{.HostConfig.Memory}}'
```

- 容器反复 `Restarting` 且 stats 顶满 limit → 在 `resources.*.yml` 微调（优先保证 voice-test）。
- 宿主机整体 swap 耗尽 → 降 MySQL buffer_pool、暂停非必要栈，或短期只开单栈验收。

**服务器验收清单**（部署后手工执行）

- [ ] 生产 3 节点 `cluster_state:ok`（`CLUSTER NODES` 仅 3 master）
- [ ] 测试 `redis-cli PING` → `PONG`（无需 cluster create）
- [ ] `docker stats` 显示各容器 mem/cpu 上限
- [ ] 测试域名 ASR/WS 通过；生产微服务仍 Up

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
pull access denied for crpi-xxx-vpc.../pangbao-test/gateway:v1.0.0-rc.1
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
docker pull crpi-xxx-vpc.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test/gateway:v1.0.0-rc.1

# 3) 核对 .env.test
grep -E '^REGISTRY=|^IMAGE_TAG=' manifest/docker/.env.test
```

**GitHub Actions push 侧**（凭证在 GitHub Secrets，**不**读仓库内 `.env` 文件；详见下方「ACR 与 CI 凭证」）：

| 配置项 | 要求 |
|--------|------|
| GitHub Environment `test` / `prod` | 各配置 secret `REGISTRY`（与 ECS `.env` 中值相同，可用 `-vpc`） |
| Repository secrets | `ACR_USERNAME`、`ACR_PASSWORD`（两环境通常相同） |
| `REGISTRY` | ECS pull 用，**可用** `-vpc`；workflow 自动去掉 `-vpc` 作为 push 地址 |
| 命名空间 | `REGISTRY` 中 `/pangbao-test` 等与目标环境一致 |
| ACR 控制台 | 对应命名空间下已有 7 个镜像仓库且存在预发布/正式 tag（如 `:v1.0.0-rc.1`、`:v1.0.0`） |

---

### D.3 微服务启动报 Redis `no such host`（redis-test / redis-node-*）

**现象**：

```text
lookup redis-test on 127.0.0.11:53: no such host
lookup redis-node-3 on 127.0.0.11:53: no such host
dependency check failed: redis dependency check failed
```

**原因**：业务容器在 `go-ai-talk-test-net`（或 `go-ai-talk-net`），但对应环境的 Redis 未启动，或不在同一 Docker 网络；测试栈未设 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`。

**解决**：

```bash
# 测试栈：先起 standalone Redis，再起微服务
docker network create go-ai-talk-test-net 2>/dev/null || true
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml ps

# 容器内 DNS 验收（测试 device-service）
docker exec go-ai-talk-device-service-test getent hosts redis-test
docker exec go-ai-talk-device-service-test printenv GF_REDIS_DEFAULT_ADDRESS
# 期望 redis-test:6379

# 生产栈：确认 3 节点 Redis 在 go-ai-talk-net
docker compose -f manifest/docker/docker-compose.redis-cluster.yml ps
docker exec go-ai-talk-device-service getent hosts redis-node-1
```

测试微服务须带 **`microservices.test.yml`** overlay，否则仍在 `go-ai-talk-net`，解析不到测试 Redis。

**一键只读验收**（Redis + MySQL 环境变量 + 对外 `appDatabase` + Redis run_id / 关键 key 抽样）：

```bash
chmod +x hack/env-isolation-check.sh
./hack/env-isolation-check.sh
```

出问题时带上 JWT 中的 `sub`（wxId）与 `device_no` 定点对照：

```bash
INCIDENT_WX_ID=42 INCIDENT_DEVICE_NO=ABCDEF \
MYSQL_CLI_USER=root MYSQL_CLI_PASS='***' MYSQL_HOST=120.55.50.105 \
./hack/env-isolation-check.sh
```

脚本只读、不改数据；`FAIL` 项优先处理（尤其 `GF_REDIS_DEFAULT_ADDRESS` 含 `redis-node`、库名无 `_test`、`run_id` 相同、测试 URL 返回 `appDatabase=ai_voice_app`）。

---

### D.3b 测试服偶现正式数据 / 注销成功但库无变化

**现象**：客户端连测试域名，偶现宝宝/历史像正式服；注销接口 200，但观测的 MySQL 无变化。

**常见根因**：

1. **观测库错位**：注销删的是 `ai_voice_device_test.wx`，查的是 `ai_voice_device.wx`。
2. **Redis 串线**：测试微服务 `GF_REDIS_DEFAULT_ADDRESS` 误指正式 Cluster；Redis key 无环境前缀，`dev:wx:id2dev:` / `history:record:list:` 会互踩。
3. **客户端/反代串线**：`test.pangbao` 偶发打到正式 gateway（`site/home` 的 `appDatabase` 为 `ai_voice_app` 而非 `ai_voice_app_test`）。

**排查**：先跑 `./hack/env-isolation-check.sh`；复现时记录 JWT `sub` 与 `device_no`，对照两库 `wx` 表与两侧 Redis 同名 key。

---

### D.4 测试 Redis：`redis-test` 未 running

**现象**：未 `up` 或 `up` 失败时 `exec redis-cli` 报错。

**解决**：

```bash
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml ps -a
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test redis-cli PING
# PONG — 测试 Redis 为 standalone，无需 cluster create
```

`logs` 无输出且 `ps` 为空 → 容器从未创建，检查 `up` 报错（网络、16379 端口占用，见 D.5）。

---

### D.4b 生产 Redis：`Node … is not empty`（重复 cluster create）

**现象**：

```text
[ERR] Node redis-node-1:7001 is not empty. Either the node already knows other nodes … or contains some key in database 0.
```

**原因**：数据卷里**已有集群元数据或 key**。`cluster create` 只对**空节点**执行一次；`up --force-recreate` **不会**清空 volume，日常重启后**不应**再跑 create。

**解决**：**先查集群是否已可用**（见 [附录：Redis 验收](#redis-验收与-cluster-create)）：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
```

- **`cluster_state:ok`** → **忽略上述报错，不要 cluster create**，直接继续 RabbitMQ / 微服务。
- **非 ok** 且确需重建 → 维护窗口内 `down -v` 清空卷后重新 `up` + 三节点 `cluster create --cluster-replicas 0`（见 [C.1 迁移](#c1-首次搭建生产栈一次性)）。

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
  go-ai-talk-device-service-test \
  go-ai-talk-ucg-service-test go-ai-talk-sim-user-service-test 2>/dev/null

docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  up -d --no-build --force-recreate

# 验：仅 19804，无 9804
docker ps --filter name=go-ai-talk-ucg-service-test --format '{{.Ports}}'
# 期望：0.0.0.0:19804->9804/tcp

# 启生产
docker compose --env-file manifest/docker/.env.prod \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.prod.yml \
  -f manifest/docker/docker-compose.resources.prod.yml \
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

### D.5 测试 Redis：端口 `16379` 或旧 `1700x` 占用

**现象**：

```text
Bind for 0.0.0.0:16379 failed: port is already allocated
```

或仍残留旧 6 节点 cluster 占用 **17001–17006**。

**原因**：宿主机端口已被占用。常见为旧 project **`docker`** 或已废弃的 **`redis-cluster.test.yml`** 留下的容器。

**排查**：

```bash
docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep -E '16379|1700[1-6]|redis-test|redis-node'
ss -tlnp | grep -E '16379|1700[1-6]'
```

**解决**：执行上文 **[① 测试 Redis](#-测试-redis16379-端口占用--旧-6-节点-cluster-残留)**，看到 `OK` 后再 `up`：

```bash
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test redis-cli PING
# PONG
```

**说明**：测试微服务经 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379` 连 Docker 网内 **6379**；宿主机 **16379** 仅用于本机调试。生产 Redis 用 **7001–7003**。

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

日常 **停/启全栈** 见 [B.3](#b3-日常停启测试全栈给生产腾资源)；本节适用于 Conflict 后 **首次搭建或重建**。

```bash
cd /www/wwwroot/go/go_ai_talk
docker network create go-ai-talk-test-net 2>/dev/null || true

# 清理：依次 ① Redis → ② RabbitMQ → ③ 微服务（各块见「启动前清理」）
# … 复制 ①②③ 的 docker rm + 验收命令 …

# 启动 Redis standalone → PING → RabbitMQ → init → 微服务（见 B.1 步骤 5–6 与 B.2 步骤 3；微服务须叠加 resources.test.yml）
```

完整命令见 **B.1**；Conflict 时仅在对应 `up` 前插入 **① / ② / ③** 各 2 行（删 + 验）。

---

## 附录

### 配置基线（新建实例必读）

| 服务 | 配置文件 | 要点 |
|------|----------|------|
| gateway | `manifest/config/config.yaml` | 无数据库 |
| gateway-app | `config.gateway-app-server.yaml` | `APP_DB_LINK`、`GF_REDIS_DEFAULT_ADDRESS`、`GATEWAY_APP_JWT_SECRET` |
| voice / device / history / ucg | 各 `config.*-service.yaml` | `*_DB_LINK`、`GF_REDIS_DEFAULT_ADDRESS`（ucg 另需 `GATEWAY_APP_JWT_SECRET` 与阿里云 env，见下表） |
| 跨服务 | Compose 环境变量 | 容器内勿用 `127.0.0.1` 访问他域；用服务名 |

关键原则：

- 每服务独立库；跨域走 HTTP，禁止跨库直查（见 `AGENTS.md`）。
- gateway-app：`GATEWAY_APP_PUBLIC_BASE_URL`（APK 下载绝对 URL + CORS 白名单 hostname）、APK 上传需 Nginx `client_max_body_size` ≥ 220MB。

### 环境隔离（MySQL / Redis / JWT）

各微服务 **不在** `manifest/config/*.yaml` 中配置 MySQL 与 Redis；**`.env.prod` / `.env.test` / `.env.local` 为唯一来源**。

| 变量 | 作用 |
|------|------|
| `MYSQL_TCP_HOST` | MySQL 主机；改库地址只改此行 |
| `*_DB_LINK` | 6 条 DSN，host 写 `mysql-host`；库名区分 test/prod |
| `GF_REDIS_DEFAULT_ADDRESS` | Redis 地址；生产多地址逗号分隔（Cluster），测试 `redis-test:6379` |
| `GATEWAY_APP_JWT_SECRET` | App access JWT 密钥；**prod/test 必须不同**；ucg 与 gateway-app 同值 |
| `UCG_OSS_ACCESS_KEY_ID` | ucg OSS AccessKey ID；yaml `ucg.oss.accessKeyId` 留空，Green 审核复用同一 AK |
| `UCG_OSS_ACCESS_KEY_SECRET` | ucg OSS AccessKey Secret |
| `UCG_DASHSCOPE_API_KEY` | ucg AI 润笔 DashScope API Key；yaml `ucg.ai.dashscope_api_key` 留空 |
| `UCG_APNS_KEY_ID` | iOS APNs Auth Key Key ID；yaml `ucg.push.apns.keyId` 留空 |
| `UCG_APNS_TEAM_ID` | Apple Developer Team ID |
| `UCG_APNS_BUNDLE_ID` | iOS Bundle ID（须与 APNs Key 授权一致） |
| `UCG_APNS_KEY_PATH` | 容器内 AuthKey `*.p8` 绝对路径（挂载进 ucg-service，非 env 内嵌正文） |
| `UCG_APNS_PRODUCTION` | `true`/`1` 走 production APNs；否则 sandbox |
| `UCG_HMS_APP_ID` | 华为 Push Kit 应用 ID |
| `UCG_HMS_APP_SECRET` | 华为 Push Kit Client Secret |
| `UCG_MIPUSH_APP_ID` | 小米推送 AppId（与 Flutter `push.properties` 客户端 id 一致） |
| `UCG_MIPUSH_APP_KEY` | 小米推送 AppKey |
| `UCG_MIPUSH_APP_SECRET` | 小米推送 AppSecret |
| `GLM_API_KEY` | 智谱 GLM；voice 默认 LLM lane 与 ucg zhipu 润笔 |
| `DEEPSEEK_API_KEY` | 可选；Admin 切回 deepseek provider 时 voice LLM |

占位符与注释见 **`manifest/docker/.env.example`**；真实值写入 `.env.local` / `.env.test` / `.env.prod`。

MySQL 经 `internal/platform/dbcfg`；Redis 经 `internal/platform/rediscfg`；启动日志应含 `database.* 已用` 与 `redis.default 已用`。

| 服务 | 环境变量 | gdb 分组 |
|------|----------|----------|
| history-service | `HISTORY_DB_LINK` | `default` |
| device-service | `DEVICE_DB_LINK` | `default` |
| voice-service | `VOICE_DB_LINK` | `default` |
| ucg-service | `UCG_DB_LINK` | `default` |
| gateway-app | `APP_DB_LINK` | `app` |

**验收**（测试栈示例）：

```bash
# 1) printenv 含 _test
for c in go-ai-talk-history-service-test go-ai-talk-device-service-test go-ai-talk-voice-service-test \
         go-ai-talk-ucg-service-test go-ai-talk-gateway-app-test; do
  echo "== $c =="; docker exec "$c" printenv | grep -E '_DB_LINK|APP_DB_LINK' || true
done

# 2) 启动日志：须含 MYSQL_TCP_HOST 解析后的主机与 _test 库名
docker logs go-ai-talk-history-service-test 2>&1 | grep "database.default 已用 HISTORY_DB_LINK"
docker exec go-ai-talk-history-service-test printenv MYSQL_TCP_HOST

# 3) 官网 API
curl -sk https://test.pangbao.cuplay.top:9702/device/app/api/site/home
# 期望 JSON 含 "appDatabase":"ai_voice_app_test"
```

变更 `*_DB_LINK` 或升级含 dbcfg 的代码后，须 **rebuild 并 force-recreate 对应容器**，不能只做 `docker compose up -d --no-build`。

### Web 管理页与环境隔离

静态页在 `resource/public/`，**唯一入口为 App 网关**（`:9702` / 测试 Nginx 反代 `:9702` 或宿主机 `:19702`）。主网关 `:9701` 对 `/device/admin*`、`/device/history/{deviceNo}` 返回 **302** 至 `GATEWAY_APP_PUBLIC_BASE_URL` 同路径。

**Admin JWT 环境变量**（`.env.*` → compose）：

| 变量 | 用途 |
|------|------|
| `GATEWAY_APP_ADMIN_USERNAME` | Hub 登录账号（默认 `admin`） |
| `GATEWAY_APP_ADMIN_PASSWORD` | Hub 登录密码（必填，未配置则 login 503） |
| `DEVICE_ADMIN_PASSWORD` | gateway-app 注入 device 反代 `X-Admin-Password`（默认同 Hub 密码） |
| `UCG_ADMIN_PASSWORD` | gateway-app 注入 ucg 反代 `X-Admin-Password`（默认同 Hub 密码） |
| `GATEWAY_APP_PUBLIC_BASE_URL` | APK 绝对 URL、CORS、主网关 302 目标 |

| 页面 | 路径 | API 请求方式 |
|------|------|----------------|
| 运维 Hub | `/device/admin` | `POST /device/admin/api/login` → Bearer；`/device/admin/api/*` |
| 历史记录 | `/device/history/{deviceNo}` | 同源 `/device/history/api/*`（用户 JWT 或白名单） |
| 问答库 / 反馈 / 统计 / UCG | `/device/admin/*` | 同源 + Admin JWT（须先 Hub 登录） |
| 版本管理 | `/device/app/version-admin.html` | 同源 `/device/app/api/version/admin/*` + Admin JWT |
| App 联调 | `/device/app/integration-test.html` | 默认 `window.location.origin`（可手改 baseUrl） |

**验收**（测试环境示例）：

```bash
# 1) 主网关 admin 302
curl -sI http://127.0.0.1:19701/device/admin | grep -i location
# 期望 Location 含 GATEWAY_APP_PUBLIC_BASE_URL（如 https://test.pangbao.cuplay.top/device/admin）

# 2) Hub 登录（App 网关）
curl -s -X POST http://127.0.0.1:19702/device/admin/api/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"<GATEWAY_APP_ADMIN_PASSWORD>"}'
# 期望 code=0 且 data.accessToken 非空

# 3) Admin JWT 访问设备管理 API
TOKEN=<上步 accessToken>
curl -s http://127.0.0.1:19702/device/admin/api/list -H "Authorization: Bearer $TOKEN"
# 期望 code=0

# 4) 用户 JWT 访问管理 API 应 403（若有测试用户 token）
# curl -s http://127.0.0.1:19702/device/admin/api/list -H "Authorization: Bearer <user-jwt>"

# 5) gateway-app CORS
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
| Redis 拓扑 / 宿主机端口 | Cluster 7001–7003 | standalone **16379**→6379 |
| Redis 容器内地址 | `redis-node-1:7001`（yaml 种子） | `redis-test:6379`（`GF_REDIS_DEFAULT_ADDRESS`） |
| RabbitMQ | 5672 / 15672 | 5673 / 15673 |
| gateway / gateway-app | 9701 / 9702 | 19701 / 19702 |
| history / voice / device / ucg / sim-user | 9801–9805 | 19801–19805 |
| MySQL | `ai_voice_*` | `ai_voice_*_test` |
| 静态目录 | `/ai_talk_images`、`/apk/ai_talk` | `*_test` 后缀 |
| `IMAGE_TAG` | semver（`.env.prod`，如 `v1.0.0`） | 预发布 semver（`.env.test`，如 `v1.0.0-rc.1`） |
| 域名 | www.pangbao.cuplay.top | test.pangbao.cuplay.top |

### Compose 文件说明

| 文件 | 用途 |
|------|------|
| `docker-compose.microservices.yml` | 基线拓扑（无宿主机 ports）；本地 `--build` |
| `docker-compose.microservices.local.yml` | 本地端口 9701/980x |
| `docker-compose.microservices.test.yml` | 测试 overlay + 端口 197xx |
| `docker-compose.microservices.prod.yml` | 生产 overlay + 端口 9701/980x |
| `docker-compose.resources.prod.yml` | 生产微服务 mem/cpu limits |
| `docker-compose.resources.test.yml` | 测试微服务 mem/cpu limits（voice 512M） |
| `docker-compose.redis-cluster.yml` | 生产 Redis Cluster（3 主 0 从）+ limits |
| `docker-compose.redis-standalone.test.yml` | 测试 Redis standalone |
| `docker-compose.rabbitmq.yml` / `.test.yml` | RabbitMQ + limits + **logging 20m×3** + `rabbitmq.conf` |
| `docker-compose.observability.yml` | 本地可选：Prometheus / Loki / Tempo / Grafana（见 [A.4](#a4-可观测性栈可选)） |

**日志轮转**：微服务基线 `docker-compose.microservices.yml` 定义 `x-docker-logging`（10m×3）；`microservices.prod.yml` / `.test.yml` **不含 `logging`**，与基线无冲突。验收见 [Docker 容器日志轮转与验收](#docker-容器日志轮转与验收)。

镜像引用：`${REGISTRY}/gateway:${IMAGE_TAG}`。`REGISTRY` = `<ACR域名>/<命名空间>`；仓库名单段（`gateway`、`device-service` 等，无 `go-ai-talk/` 前缀）。

### Docker 容器日志轮转与验收

compose 源文件中的策略（prod/test/local 共用基线，overlay 不覆盖 `logging`）：

| compose 文件 | logging 策略 | 备注 |
|--------------|-------------|------|
| `docker-compose.microservices.yml` | 10m × 3 | 七微服务 `logging: *docker-logging` |
| `docker-compose.redis-cluster.yml` | 10m × 3 | 三 Redis 节点 |
| `docker-compose.redis-standalone.test.yml` | 10m × 3 | 测试 standalone |
| `docker-compose.rabbitmq.yml` / `.test.yml` | 20m × 3 | 挂载 `rabbitmq/rabbitmq.conf` |

**部署后 recreate**（测试栈示例，仓库根目录）：

```bash
docker compose -f manifest/docker/docker-compose.rabbitmq.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d --force-recreate
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  -f manifest/docker/docker-compose.resources.test.yml \
  up -d --force-recreate --no-build
```

生产栈将上述 rabbit/redis 文件换为 `.yml` / `redis-cluster.yml`，微服务 overlay 换为 `microservices.prod.yml` + `resources.prod.yml`，`.env.test` 换为 `.env.prod`。

**验收 LogConfig**：

```bash
docker inspect --format='{{.HostConfig.LogConfig}}' go-ai-talk-ucg-service-test
# 期望 {json-file map[max-file:3 max-size:10m]}

docker inspect --format='{{.HostConfig.LogConfig}}' go-ai-talk-rabbitmq-test
# 期望 {json-file map[max-file:3 max-size:20m]}

docker system df -v | grep -E 'json.log|CONTAINER'
```

**清理历史巨型 log**（配置轮转前已撑满的磁盘）：

```bash
docker system df
sudo du -sh /var/lib/docker/containers/*/*-json.log 2>/dev/null | sort -h | tail

# 方案 A（推荐）：对上述 stack force-recreate，删除旧容器时 log 一并移除
# 方案 B：truncate 现有 log（一般无需停容器）
sudo truncate -s 0 /var/lib/docker/containers/*/*-json.log
```

RabbitMQ 排障需连接/channel 细节时，临时将 `manifest/docker/rabbitmq/rabbitmq.conf` 中 `*.level` 改为 `info`，recreate Rabbit 容器，排障后改回 `warning`。更多说明见 `docs/runbooks/rabbitmq-local.md`。

### Redis 验收与 cluster create（仅生产）

> **测试 Redis 为 standalone**，验收只需 `redis-cli PING` → `PONG`，**不要** `cluster create`。  
> **`cluster create` 仅对空节点执行一次。** 数据卷保留时重复执行会报 `Node … is not empty`——通常表示**集群已初始化**。

#### 1. 生产 Cluster 是否已可用

**生产**（`docker-compose.redis-cluster.yml`，3 主 0 从）：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# cluster_state:ok → 已就绪，跳过 cluster create

docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER NODES
# 期望 3 个 master
```

**测试**（`docker-compose.redis-standalone.test.yml`）：

```bash
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test \
  redis-cli PING
# PONG
```

#### 2. 仅当生产 `cluster_state` 非 ok 时：首次 cluster create

3 节点均为 **running** 且上一步**不是** `cluster_state:ok` 时（全新 volume 或重置后）才执行：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli --cluster create \
  redis-node-1:7001 redis-node-2:7002 redis-node-3:7003 \
  --cluster-replicas 0 --cluster-yes
```

create 后再验：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.yml exec -T redis-node-1 \
  redis-cli -p 7001 CLUSTER INFO | grep cluster_state
# 期望 cluster_state:ok
```

#### 3. 重置 Redis（会清空数据）

**测试 standalone**：

```bash
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml down -v
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml up -d --force-recreate
docker compose -f manifest/docker/docker-compose.redis-standalone.test.yml exec -T redis-test redis-cli PING
```

**生产 Cluster**（须评估业务影响）：

```bash
docker compose -f manifest/docker/docker-compose.redis-cluster.yml down -v
docker compose -f manifest/docker/docker-compose.redis-cluster.yml up -d --force-recreate
# 再按「2. 首次 cluster create」三节点 --cluster-replicas 0
```

相关排错：[D.4b](#d4b-生产-redisnode--is-not-empty重复-cluster-create)。

> 测试 Redis 默认使用 `docker-compose.redis-standalone.test.yml`（`redis-test:6379`），勿再使用已删除的六节点 cluster test compose。

### ACR 与 CI 凭证

个人版 ACR 镜像完整路径：`域名/命名空间/仓库名:tag`。**测试与生产使用不同命名空间**。CI 与 ECS **职责分离**：

| 用途 | 凭证存放位置 | 说明 |
|------|--------------|------|
| **GitHub Actions push** | GitHub Secrets / Environments | 不读 git 仓库内 `.env` 文件 |
| **ECS pull / compose** | 服务器本地 `manifest/docker/.env.test` / `.env.prod` | 从 `.env.example` 复制，**不上传 git** |

#### CI 触发与 Environment 路由

| 环境 | GitHub Environment | CI 触发 |
|------|-------------------|---------|
| **测试** | `test` | push tag `v*-*`（预发布，如 `v1.0.0-rc.1`）；`v1.0.0-rc.2+ucg` 等 **`+服务` 后缀** 仅 build 列出服务 |
| **生产** | `prod` | push tag `vMAJOR.MINOR.PATCH`（如 `v1.0.0`） |

**构建范围后缀**（仅 tag push；**不 retag** 未构建服务）：

| git tag 示例 | CI 行为 | ACR |
|--------------|---------|-----|
| `v1.0.0-rc.2` | 全量 7 服务 build | 七仓库均有 `:v1.0.0-rc.2` |
| `v1.0.0-rc.2+ucg` | 仅 `ucg-service` | 仅 `ucg-service:v1.0.0-rc.2`（其他仓库无此 tag） |
| `v1.0.0-rc.2+ucg,gateway` | 两项 | 对应两仓库 |
| `v1.0.0-rc.3+sim` | 仅 `sim-user-service` | 仅 `sim-user-service:v1.0.0-rc.3` |

别名：`gateway`、`gateway-app`、`history`/`history-service`、`voice`/`voice-service`、`device`/`device-service`、`ucg`/`ucg-service`、`sim`/`sim-user`/`sim-user-service`、`all`（全量）。`.env` 中 `IMAGE_TAG` 用 **`+` 前 base tag**。

手动触发：Actions → **docker-acr** → Run workflow → 选择 `target_env`、`image_tag`（base tag）；可选 `services`（如 `ucg`，留空=全量）。

#### GitHub Secrets 配置（首次必做）

1. **Settings → Environments**：创建 `test`、`prod`（prod 可按需加 required reviewers）。
2. **Settings → Secrets and variables → Actions → Repository secrets**（两环境共用）：

| Secret | 取值 |
|--------|------|
| `ACR_USERNAME` | ACR 控制台 → 访问凭证 → 登录用户名 |
| `ACR_PASSWORD` | 访问凭证 → 设置的**固定密码** |

3. **各 Environment secrets**（命名空间不同）：

| Environment | Secret | 取值 |
|-------------|--------|------|
| `test` | `REGISTRY` | `<vpc 或公网域名>/pangbao-test`（与 ECS `.env.test` 中 `REGISTRY` 一致） |
| `prod` | `REGISTRY` | `<vpc 或公网域名>/pangbao-release`（与 ECS `.env.prod` 中 `REGISTRY` 一致） |

CI push 地址由 workflow 从 `REGISTRY` **自动去掉 `-vpc`** 推导，无需单独维护公网域名。

**验证 CI**：配置完成后，Actions → docker-acr → Run workflow → `target_env=test`、`image_tag=v0.0.0-test`（或当前预发布 tag），确认七服务 build + push 成功。部分构建可填 `services=ucg` 或 `services=sim` 验证单服务 matrix。

#### ECS 本地 .env（部署用）

服务器上 `manifest/docker/.env.test` / `.env.prod` **不进 git**（见仓库 `.gitignore`）。除 CI 所需的 ACR 三字段外，还须配置 `IMAGE_TAG`、各 `*_DB_LINK`、Redis、JWT 等完整项（见 `.env.example`）。

**ECS 镜像段示例**（命名空间名以控制台为准）：

```text
# manifest/docker/.env.test（仅存在于 ECS，勿 commit）
REGISTRY=crpi-lff3xynwzvqxxxjk-vpc.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test
IMAGE_TAG=v1.0.0-rc.1
ACR_USERNAME=<ACR 登录用户名>
ACR_PASSWORD=<ACR 固定密码>

# manifest/docker/.env.prod（仅存在于 ECS，勿 commit）
REGISTRY=crpi-lff3xynwzvqxxxjk-vpc.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-release
IMAGE_TAG=v1.0.0
ACR_USERNAME=<ACR 登录用户名>
ACR_PASSWORD=<ACR 固定密码>
```

换 ACR 密码时须**同时**更新 GitHub Repository secrets 与各 ECS `.env` 中的 `ACR_PASSWORD`。

每个命名空间下须预先创建 7 个镜像仓库：`gateway`、`gateway-app`、`history-service`、`voice-service`、`device-service`、`ucg-service`、`sim-user-service`（或开启「自动创建仓库」）。

push 报 `denied` 常见原因：① GitHub 或 ECS 缺 `ACR_USERNAME` / `ACR_PASSWORD`；② `REGISTRY` 缺命名空间；③ 测试/生产命名空间混用；④ 仓库未创建；⑤ ECS 未 `docker login`。详见 [D.2](#d2-acr-拉镜像-pull-access-denied)。

### 脱敏种子（测试库刷新）

```bash
MYSQL_HOST=127.0.0.1 MYSQL_USER=root MYSQL_PASS='***' ./hack/mask-seed-data.sh
```

发版前可选执行；导入后 recreate gateway-app / history / device。

### 发布前检查

- `go test ./cmd/... ./internal/...`
- 测试栈预发布 tag 验收通过（B.2 步骤 4）
- 生产 `.env.prod` 的 `IMAGE_TAG` 与 git 正式 tag 一致（如 `v1.0.0`）
- `./hack/env-isolation-check.sh` 无 `FAIL`（或等价 `printenv`：`MYSQL_TCP_HOST GF_REDIS_DEFAULT_ADDRESS GATEWAY_APP_JWT_SECRET *_DB_LINK` 确认 test/prod 未串）
- 各微服务启动日志含 `database.* 已用 *_DB_LINK 覆盖`（见「数据库环境隔离」）

### 宝塔面板无响应

```bash
/etc/init.d/bt restart
/etc/init.d/bt status
```

### Kubernetes

见 `manifest/deploy/kustomize/overlays/develop`；Compose 发版路径以本文 A/B/C 为准。

### Apple Sign in with Apple（`wx.apple_sub` 迁移）

本变更**不新增** `*_DB_LINK`；DDL 在 **device-service 默认库**（`DEVICE_DB_LINK` / `database.default`，如 `ai_voice_device`）执行。

**部署顺序（须严格按序）：**

1. **DDL**（低峰/维护窗口）：在 `DEVICE_DB_LINK` 对应库执行 `docs/migrations/wx_apple_sub.sql`  
   - `ALTER TABLE wx ADD COLUMN apple_sub ...`  
   - `CREATE UNIQUE INDEX uk_wx_apple_sub ON wx (apple_sub)`  
   - 执行后确认既有微信/用户名行 `apple_sub` 均为 NULL，无唯一索引冲突。
2. **配置**：`manifest/config/config.device-service.yaml` 中 `apple.ios.bundleId` 须与 iOS 客户端 Bundle ID 一致（`com.fzy.pangbao`）；生产通过 overlay/Secret 覆盖时须核对。
3. **滚动 device-service**（含 `apple_auth.go`、绑定 API、profile 扩展）。
4. **滚动 gateway-app-server**（`apple_login` 聚合路由与白名单）。
5. **联调**：打开 `/device/app/integration-test.html`，用真机 `identityToken` 验证 `apple_login`、Bearer 下 `apple/bind` 与 `wx/bindwx`。

**回滚**：下线新路由即可；`apple_sub` 列可保留，已创建的 Apple 账号行不受影响。

### AI 月度额度（分域 `ai_quota_*` 表）

voice / ucg 各自在默认库维护配额配置；device-service **不再**承载 `ai_quota_*`：

1. **DDL voice**：`manifest/sql/voice_ai_quota.sql`（`ai_voice_voice`，默认 5/30）
2. **DDL ucg**：`manifest/sql/ucg_ai_quota.sql`（`ai_voice_ucg`，默认 polish=5）
3. **网关**：`VOICE_API_PROXY_URL` + `VOICE_API_ROUTE_MODE=proxy` 反代 `/voice/app/api/*`、`/voice/admin/api/*`；`VOICE_ADMIN_PASSWORD` 注入 voice admin
4. **App API**：`GET /voice/app/api/ai-quota`、`GET /ucg/app/api/ai-quota`（不计入 usage 统计）
5. **维护窗口后**：`manifest/sql/device_drop_ai_quota.sql` 于 device 库执行

### UCG 审核 MQ 卡死 / apply 失败

**背景**：资料/帖子/评论/私信审核拆为两阶段（Green 机审 → apply CAS）。**Green 单次**：对每个 `(entity_id, audit_version)` 一旦 Phase1 已发起 Green（含 API 失败），后续 MQ 消费 MUST NOT 再次调 Green；Green/API 或 persist verdict 失败进入 **机审失败终态**（profile/post/comment `status=5`，chat `audit_status=moderation_failed`）并 Ack。`moderation_verdict` 落库后 MQ 重投仅 retry apply；apply 超限 `apply_failed` 并 Ack。

**Green 图片/视频 dataId 修复**（`fix-ucg-green-media-dataid`）：旧版将完整 CDN URL 作为 `dataId`（超长且含非法字符），导致 `ImageModeration` / `baselineCheck` 参数校验失败、纯图帖落入 `status=5`。修复后 `dataId` 由 URL path 规范为 `social_YYYY_MM_xxx.ext`（≤64）。**存量 `status=5` 帖子不会自动重审**（green-once）；须作者重新提交或运维人工处理。部署后验证：

1. 新发带图帖 → 阿里云 Green 控制台应出现 `baselineCheck` 成功记录
2. ucg-service 日志不应再出现仅 `green image: code not ok`（失败时应含 `code` + `msg`）
3. 可选 SQL：`SELECT id, status, reject_reason FROM ucg_post WHERE status=5 ORDER BY id DESC LIMIT 10;`

**DDL**（`UCG_DB_LINK` 库，须先于 ucg-service 滚动）：

```bash
mysql -h "$MYSQL_HOST" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$UCG_DB_NAME" < hack/sql/ucg_audit_moderation_apply.sql
mysql -h "$MYSQL_HOST" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$UCG_DB_NAME" < hack/ddl/ucg-audit-comment-chat-moderation.sql
```

**识别**：

- RabbitMQ 队列 `ucg.profile.patch.submitted.q` / `ucg.post.created.q` 的 `messages_ready` 持续增长
- ucg-service 日志含 `[ucg-audit-mq] profile/post apply retry` 或 `apply max exceeded`
- 阿里云 Green 控制台对同一 `(job_id/post_id, audit_version)` 调用量异常

**止血**：

```bash
# 暂停审核 consumer（推荐分 consumer 不受影响）
UCG_AUDIT_MQ_CONSUMER_ENABLED=false
# 滚动 ucg-service 使 env 生效，或临时 scale 0
```

**核对 DB**（`UCG_DB_LINK`）：

```sql
-- 机审已通过但 apply 未完成（可安全重试 apply，不会重复 Green）
SELECT id, wx_id, status, audit_version, moderation_verdict, apply_attempts
FROM ucg_profile_audit_job WHERE status=1 AND moderation_verdict=1;

SELECT id, author_wx_id, status, audit_version, moderation_verdict, apply_attempts
FROM ucg_post WHERE status=1 AND moderation_verdict=1;

-- apply 超限终态（consumer 应 Ack，勿无限 requeue）
SELECT id, status, apply_attempts, apply_failed_at FROM ucg_profile_audit_job WHERE status=4;
SELECT id, status, apply_attempts, apply_failed_at FROM ucg_post WHERE status=4;

-- 机审失败待人工（资料 job status=5；作者 App 无专用展示）
SELECT id, wx_id, audit_version, reject_reason, updated_at FROM ucg_profile_audit_job WHERE status=5;
```

**资料机审失败人工处理**（UCG 管理页 `/device/admin/ucg-admin.html` →「资料机审失败」Tab，或 API）：

- `GET /ucg/admin/api/profile-audit-jobs/list?status=5` — 列表
- `POST /ucg/admin/api/profile-audit-jobs/resolve` — body `{ "jobId", "action": "approve"|"reject", "reason"? }`
- **approve**：CAS `status 5→2` 并按 job patch 更新 `ucg_profile`
- **reject**：CAS `status 5→3`，不更新已发布 profile

**清理队列**（仅当 DB 已为终态 `approved/rejected/published/apply_failed/moderation_failed` 或 `status=1 AND moderation_verdict=1` 且已部署本修复）：

- RabbitMQ Management → Queues → 对应队列 → Purge（**禁止**在 verdict 未落库时 purge 后丢弃未机审消息）
- 或对单条 delivery：确认 DB 后手动 Ack

**修复 pending job**（`moderation_verdict=1` 且 `status=1` 长期 pending）：

1. 确认 Green 已通过（`moderation_at > 0`）
2. 可选：重置 `apply_attempts=0` 后重发 outbox / 重启 consumer 让其仅 retry apply
3. 或运维确认后直接手工 approve（更新 `ucg_profile` + job `status=2`）

**修复 apply_failed**（`status=4`）：

1. 排查根因（DB 连接、事务、profile 行缺失等）
2. 修复后：用户重新提交资料/帖子，或运维重置 `apply_attempts=0`、`moderation_verdict=0`、`status=1` 并重发 MQ（须明确审批）

**部署 checklist**：

1. 执行 `hack/sql/ucg_audit_moderation_apply.sql`
2. 执行 `hack/ddl/ucg-audit-comment-chat-moderation.sql`
3. 滚动 ucg-service（含两阶段审核代码）
3. 观察队列深度下降、Green 调用量恢复正常
4. 可选 env：`UCG_AUDIT_APPLY_MAX_ATTEMPTS=5`（默认）

**手工验证**（非 CI）：模拟 apply 失败（如临时断 DB）后同一条 MQ redelivery，日志应出现 `moderation skip green` 且 Green 控制台无重复调用。

### 文档治理

运行、部署、配置边界变更须同步更新本文与 `docs/runbooks/dao-sync-by-domain.md`。
