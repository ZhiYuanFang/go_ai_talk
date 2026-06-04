## Why

即将发布首版（v1.0.0），当前仅有一套 Compose 微服务栈（`:local` 本机构建、固定端口与单一 `go-ai-talk-net`），缺少与生产完全隔离的测试环境，runbook 也未说明镜像仓库 tag 与 Compose 的版本控制。需要在同机部署生产/测试双栈，测试栈跟踪 CI 的 `develop` 浮动 tag、跑全 MQ 链路，并通过 `test.pangbao.cuplay.top` 以与生产相同的访问形态验收后再钉死 semver 发布生产。

## What Changes

- 新增生产/测试 Compose overlay 与中间件独立栈：独立 Docker 网络、Redis Cluster、RabbitMQ、MySQL `_test` 库、静态资源目录。
- 微服务镜像改为从镜像仓库 pull（`${REGISTRY}/go-ai-talk/<service>:${IMAGE_TAG}`）；测试默认 `IMAGE_TAG=develop`，生产钉死 semver（如 `v1.0.0`）。
- 测试后端端口映射为 19701/19702/1980x/19901；对外经 Nginx 以 `test.pangbao.cuplay.top:9701/9702` 反代，与生产 URL 结构一致。
- 补充 `manifest/docker/.env.prod.example`、`.env.test.example`；Redis 经 `GF_REDIS_DEFAULT_ADDRESS` 注入测试集群种子。
- 新增脱敏种子数据运维剧本（dump → mask → import 至 `ai_voice_*_test`）。
- 在 `docs/runbooks/release-deploy-and-run.md` 补充 Compose 版本控制、双栈部署、Nginx、种子刷新与发布/回滚 checklist。
- 基线 `docker-compose.microservices.yml` 保留本机 `build + :local` 供开发；prod/test overlay **移除 build**，仅 pull registry。

## Capabilities

### New Capabilities

- `compose-prod-test-dual-stack`：同机生产/测试双栈完全隔离（网络、Redis、RabbitMQ、端口、container_name、registry 镜像 tag 策略、Nginx 访问形态）。
- `compose-mysql-test-seed-desensitization`：从生产库脱敏导出并导入测试库 `ai_voice_*_test` 的可重复运维流程与验收要求。

### Modified Capabilities

- `compose-host-root-asset-volumes`：扩展测试环境静态资源宿主机路径（`/ai_talk_images_test`、`/apk/ai_talk_test`）及 runbook 对照说明。
- `compose-mysql-endpoint-via-env`：扩展 prod/test 分环境 `.env` 示例与 `APP_DB_LINK` 等变量在测试栈的约定。
- `runtime-docs-centralization-and-governance`：要求 `release-deploy-and-run.md` 同步双栈与镜像版本控制章节。

## Impact

- **文件**：`manifest/docker/docker-compose.microservices.{prod,test}.yml`、`docker-compose.redis-cluster.test.yml`、`docker-compose.rabbitmq.test.yml`、`.env.prod.example`、`.env.test.example`；可选 `hack/mask-seed-data.sh`；`docs/runbooks/release-deploy-and-run.md`。
- **运维**：同机需创建 `go-ai-talk-test-net`、测试 Redis cluster（宿主机 17001–17006）、测试 RabbitMQ（5673/15673）、测试 MySQL 库、Nginx 站点 `test.pangbao.cuplay.top`、TLS 证书。
- **CI/Registry**：需产出 `:develop` 浮动 tag 与不可变 `:<sha>` / semver tag；生产禁止跟踪 `develop`。
- **API/业务代码**：无行为变更；配置隔离依赖现有 `*_DB_LINK`、`MQ_HTTP_API_BASE`、`GF_REDIS_DEFAULT_ADDRESS`、`GATEWAY_APP_PUBLIC_BASE_URL` 等环境变量。
