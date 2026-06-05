## 1. Compose 基线与 prod/test overlay

- [x] 1.1 在基线 `docker-compose.microservices.yml` 顶部注释中说明：生产/测试须叠加 `*.prod.yml` / `*.test.yml` 并从 registry pull
- [x] 1.2 新增 `docker-compose.microservices.prod.yml`：`image: ${REGISTRY}/<svc>:${IMAGE_TAG}`、无 `build`、prod 网络 `go-ai-talk-prod-net`（或兼容现有 `go-ai-talk-net` 并文档化）、prod container_name/卷
- [x] 1.3 新增 `docker-compose.microservices.test.yml`：registry 镜像、test 网络 `go-ai-talk-test-net`、端口 19701/19702/19801–19804/19901、test container_name、`GF_REDIS_DEFAULT_ADDRESS`/`GATEWAY_APP_PUBLIC_BASE_URL` 等 test 环境变量、卷 `/ai_talk_images_test` 与 `/apk/ai_talk_test`

## 2. 测试中间件 Compose

- [x] 2.1 新增 `docker-compose.redis-cluster.test.yml`：6 节点、宿主机 17001–17006、仅 `go-ai-talk-test-net`、独立 volume 名
- [x] 2.2 新增 `docker-compose.rabbitmq.test.yml`：5673/15673、容器名 `go-ai-talk-rabbitmq-test`、仅 test 网络
- [x] 2.3 在 runbook 或 compose 注释中补充 test Redis `cluster create` 与 test Rabbit `rabbitmq-init.sh`（`COMPOSE_FILE` + `RABBIT_API_BASE=http://127.0.0.1:15673/api`）命令

## 3. 环境变量示例

- [x] 3.1 新增 `manifest/docker/.env.prod.example`：`REGISTRY`、`IMAGE_TAG=v1.0.0`、生产 `*_DB_LINK`、`APP_DB_LINK`、生产 secrets 占位说明
- [x] 3.2 新增 `manifest/docker/.env.test.example`：`IMAGE_TAG=develop`、`GATEWAY_APP_PUBLIC_BASE_URL=https://test.pangbao.cuplay.top:9702`、`*_test` 库 DSN、`GF_REDIS_DEFAULT_ADDRESS` test 种子、独立 `DEVICE_GATEWAY_INTERNAL_SECRET` 说明
- [x] 3.3 更新 `manifest/docker/.env.example` 注释：指向 prod/test 分环境示例文件

## 4. 脱敏种子与静态同步

- [x] 4.1 新增 `hack/mask-seed-data.sh`（或等价文档化 SQL 步骤）：mysqldump → 脱敏手机号/wx/token/设备号 → import `ai_voice_*_test`
- [x] 4.2 在脚本或 runbook 中说明：同步 `/ai_talk_images` → `/ai_talk_images_test`（按需 rsync/cp）

## 5. Runbook 文档

- [x] 5.1 在 `docs/runbooks/release-deploy-and-run.md` 新增「Compose 与镜像版本控制」：`REGISTRY`/`IMAGE_TAG`、develop vs semver、pull + `--no-build`、回滚
- [x] 5.2 新增「生产/测试双栈」：网络/端口/库名/静态目录/中间件对照表、test 启动顺序、`COMPOSE_PROJECT_NAME`
- [x] 5.3 新增「测试访问 test.pangbao.cuplay.top」：Nginx 9701/9702 → 19701/19702、TLS、`client_max_body_size`、健康检查 URL
- [x] 5.4 新增「脱敏种子刷新」与 §4 发布前 checklist 增补项（test develop 验收、prod tag 钉扎、printenv DSN 验收）
- [x] 5.5 更新 §5 回滚：明确生产回滚为改 `IMAGE_TAG` 而非 `--build`；警告勿对生产 `docker system prune -a`

## 6. 验收（手工）

- [x] 6.1 同机 prod + test 同时 up，确认端口无冲突、test `https://test.pangbao.cuplay.top:9702/api.json` 可达
- [x] 6.2 验证 test MQ 全链路：history outbox → test Rabbit → test worker；prod worker 不消费 test 消息
- [x] 6.3 验证 test 上传 APK/logo 仅落盘 `_test` 目录；test DB 写入 `_test` 库

## 7. CI（GitHub Actions → ACR）

- [x] 7.1 新增 `.github/workflows/docker-acr.yml`：预发布 tag → 测试 ACR；正式 tag `vMAJOR.MINOR.PATCH` → 生产 ACR（均附带 `:<sha>`）
- [x] 7.2 runbook §2.5.1 改为 GitHub Actions 说明，停用 ACR 控制台构建规则
- [x] 7.4 runbook 重构为 A 本地 / B 测试 / C 生产 分步发布流程，精简 ACR 配置说明

## 8. Web 管理页环境隔离

- [x] 8.1 `admin.html`：`gatewayAppBase` 按 9701↔9702 / 19701↔19702 配对，避免测试页跳转生产 App 网关
- [x] 8.2 gateway-app CORS 白名单由 `.env` 的 `GATEWAY_APP_PUBLIC_BASE_URL` hostname + 可选 `GATEWAY_APP_CORS_ALLOWED_HOSTS` 注入，移除代码内固定域名
- [x] 8.3 runbook 补充 Web 页验收（admin 版本管理链接、各页 API 同源）

## 9. CI：仅 tag 触发构建

- [x] 9.1 移除 `docker-acr.yml` 对 develop 分支 push 的触发；测试/生产均仅 tag 构建
- [x] 9.2 预发布 tag（`v1.0.0-rc.1` 等）→ 测试 ACR；正式 tag（`v1.0.0`）→ 生产 ACR
- [x] 9.3 更新 runbook B.2 与 `.env.test.example`：IMAGE_TAG 改为预发布 semver，不再使用 `:develop`

## 10. 全服务数据库 test/prod 隔离（`*_DB_LINK` 必须生效）

- [x] 10.1 新增 `internal/platform/dbcfg.ApplyGroupFromEnv`：yaml 已有 `database.*.link` 时，须 `gdb.SetConfigGroup` 才能用 Compose 注入的 DSN
- [x] 10.2 各微服务 cmd 启动最早阶段接入：`HISTORY_DB_LINK` / `DEVICE_DB_LINK` / `VOICE_DB_LINK` / `UCG_DB_LINK` / `WORKER_OUTBOX_DB_LINK` / `APP_DB_LINK`
- [x] 10.3 更新 config yaml 注释、`.env.example` 与 runbook「数据库环境隔离验收」
- [ ] 10.4 服务器 rebuild 全量微服务镜像后验收：各容器启动日志含 `database.* 已用 *_DB_LINK 覆盖，库名=ai_voice_*_test`；官网 API `appDatabase` 为 `ai_voice_app_test`
