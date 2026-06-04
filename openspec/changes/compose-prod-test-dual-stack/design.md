## Context

当前 `manifest/docker/docker-compose.microservices.yml` 面向本机开发：`build` + `go-ai-talk/*:local`、固定 `container_name`、单一 external 网络 `go-ai-talk-net`、宿主机端口 9701–9901。生产已运行在同一 Linux 宿主机（宝塔/Nginx 反代 `www.pangbao.cuplay.top:9701/9702`），即将发布 v1.0.0。镜像仓库已配置 CI：`develop` 分支 push 自动构建并覆盖 `:develop` tag。

约束（来自 explore 与 AGENTS.md）：

- 测试与生产同机、完全隔离（含 Redis、RabbitMQ、MySQL、静态资源、Docker 网络）。
- 测试跑全 MQ（outbox relay + worker consumer）。
- 测试对外访问形态与生产一致：`test.pangbao.cuplay.top:9701/9702`（Nginx → 后端 19701/19702）。
- 测试栈默认 `IMAGE_TAG=develop`；生产钉死 semver。
- MySQL 测试库从生产脱敏种子导入。
- 不修改业务 API 行为；配置隔离依赖现有环境变量（`*_DB_LINK`、`MQ_HTTP_API_BASE`、`GF_REDIS_DEFAULT_ADDRESS`、`GATEWAY_APP_PUBLIC_BASE_URL` 等）。

## Goals / Non-Goals

**Goals:**

- 提供 prod/test 两套 Compose overlay + `.env.*.example`，实现 registry pull 部署与版本钉扎。
- 测试栈独立中间件：Redis cluster（宿主机 17001–17006）、RabbitMQ（5673/15673）、`go-ai-talk-test-net`。
- 更新 `release-deploy-and-run.md`：镜像版本控制、双栈启动顺序、Nginx、脱敏种子、发布/回滚 checklist。
- 提供脱敏种子运维剧本（脚本或 runbook 逐步命令）。

**Non-Goals:**

- 不改 K8s overlay（可后续对齐）。
- 不引入 Redis key 前缀或 RabbitMQ vhost 代码改造（物理隔离即可）。
- 不自动化 CI 部署 webhook（首版文档 + 手工 pull）。
- 不在 git 中提交真实生产口令或 `.env.prod`/`.env.test`。

## Decisions

### 1. 双 Docker 网络而非共用 `go-ai-talk-net`

**选择**：生产 `go-ai-talk-prod-net`，测试 `go-ai-talk-test-net`；各栈内仍用 `redis-node-1`、`rabbitmq`、`history-service` 等服务名。

**理由**：完全 DNS 隔离；compose 结构与 prod 镜像一致，测试 overlay 主要改端口/卷/网络/镜像源。

**备选**：共用 `go-ai-talk-net` 并 rename 所有 test 服务 → 维护成本高、易误连。

### 2. 三层 Compose 文件

| 文件 | 职责 |
|------|------|
| `docker-compose.microservices.yml` | 基线拓扑（保留 `build` + `:local` 供开发） |
| `docker-compose.microservices.prod.yml` | 去 `build`；`image: ${REGISTRY}/<svc>:${IMAGE_TAG}`；prod 网络/卷/名 |
| `docker-compose.microservices.test.yml` | 同上；test 端口 197xx；test 卷；test 网络 |

启动示例：

```bash
docker compose --env-file manifest/docker/.env.test \
  -f manifest/docker/docker-compose.microservices.yml \
  -f manifest/docker/docker-compose.microservices.test.yml \
  pull && up -d --no-build
```

### 3. 镜像 tag 双轨

| 环境 | `IMAGE_TAG` | 更新方式 |
|------|-------------|----------|
| 测试 | `develop`（默认） | CI 覆盖；运维 `pull` + `up --no-build` |
| 生产 | `v1.0.0` 等 semver | 发版人工改 `.env.prod`；禁止 `develop`/`latest` |

CI 同时 push `:<git-sha>` 供排错；发版闸门：打 release tag 前记录测试通过的 sha。

### 4. 端口与访问

| 角色 | 生产（容器/宿主机） | 测试（容器/宿主机） | 对外 URL |
|------|---------------------|---------------------|----------|
| gateway | 9701 | 19701 | `*.pangbao.cuplay.top:9701` |
| gateway-app | 9702 | 19702 | `*.pangbao.cuplay.top:9702` |
| history | 9801 | 19801 | 经 gateway 反代 |
| voice | 9802 | 19802 | 经 gateway 反代 |
| device | 9803 | 19803 | 经 gateway 反代 |
| ucg | 9804 | 19804 | 经 gateway 反代 |
| worker | 9901 | 19901 | 内网 healthz |

测试 `GATEWAY_APP_PUBLIC_BASE_URL=https://test.pangbao.cuplay.top:9702`。

### 5. 中间件隔离

- **Redis test**：`docker-compose.redis-cluster.test.yml`，宿主机端口 17001–17006，仅加入 test 网络；cluster create 与 prod 相同流程。
- **RabbitMQ test**：`docker-compose.rabbitmq.test.yml`，5673/15673；`hack/rabbitmq-init.sh` 通过 `COMPOSE_FILE` + `RABBIT_API_BASE=http://127.0.0.1:15673/api` 初始化。
- **MySQL**：同 mysqld 实例，库名 `ai_voice_*_test`；`.env.test` 各 `*_DB_LINK` 指向 `_test`。

### 6. 静态资源

- 生产：`/ai_talk_images`、`/apk/ai_talk`
- 测试：`/ai_talk_images_test`、`/apk/ai_talk_test`（test overlay 卷 + 可选 env 覆盖）

### 7. 脱敏种子

- 流程：`mysqldump` 各 prod 库 → `hack/mask-seed-data.sh`（或 runbook 逐步 SQL）→ import 至 `_test` 库。
- 脱敏：手机号、wx openid/unionid、token、设备号前缀等；同步 logo 文件至 `/ai_talk_images_test`。
- 刷新策略：发版前必刷（首版写死 runbook）。

### 8. 现有 prod 栈迁移

- 不强制一次性 rename 网络：prod 可继续用 `go-ai-talk-net` 直至维护窗口，或 prod overlay 显式命名为 `go-ai-talk-prod-net` 并文档化一次性迁移步骤。
- test 栈新建，不影响 prod `docker compose down`。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| `develop` 浮动 tag 导致「测 A 发 B」 | 发版前记录 sha；release tag 基于同一 commit；可选 prod 前 pin `v1.0.0` smoke |
| 同机双 Redis cluster 资源占用 | 6 节点 test cluster 内存可控；runbook 注明最低宿主机规格 |
| 脱敏不彻底泄露 PII | mask 脚本 + import 前人工 spot check；测试库账号权限隔离 |
| prod 误用 test DSN | `.env` 分离 + 部署后 `printenv *_DB_LINK` 验收项 |
| voiceChat 外部 API 仍打生产额度 | runbook 注明测试可选独立 key（配置层，非本变更代码） |
| Nginx/TLS 未配 test 域名 | runbook 给宝塔模板；证书 `test.pangbao.cuplay.top` |

## Migration Plan

1. 新增 compose/env/runbook 文件（本变更 apply）。
2. 运维：创建 test 网络、MySQL `_test` 库、静态目录、TLS、Nginx 站点。
3. 启动 test Redis → cluster create → test Rabbit → `rabbitmq-init.sh`。
4. 执行脱敏种子 import + 静态文件同步。
5. `pull develop` 启动 test 栈，全链路验收。
6. CI 打 `v1.0.0`，prod `.env.prod` 改 tag，`pull` + `up --no-build`。
7. **回滚**：prod 改回上一 semver tag + `force-recreate`；勿 `docker system prune -a` 于生产。

## Open Questions

- 镜像仓库 `REGISTRY` 主机名（apply 时在 `.env.*.example` 用占位符 `registry.example.com`，运维替换）。
- prod 网络是否从 `go-ai-talk-net` 一次性迁为 `go-ai-talk-prod-net`（可在 apply 时 prod overlay 保持兼容 alias）。
