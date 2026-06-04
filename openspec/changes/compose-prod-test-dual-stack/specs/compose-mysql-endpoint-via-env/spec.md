## ADDED Requirements

### Requirement: 仓库 MUST 提供分环境的 Compose 数据库环境样例

除现有 `manifest/docker/.env.example` 外，仓库 SHALL 提供 `manifest/docker/.env.prod.example` 与 `manifest/docker/.env.test.example`。prod 示例 SHALL 说明各 `*_DB_LINK` 指向生产库名（无 `_test` 后缀）及 `IMAGE_TAG` 为 semver。test 示例 SHALL 说明各 `*_DB_LINK` 指向 `ai_voice_*_test` 库、`IMAGE_TAG=develop`、`GATEWAY_APP_PUBLIC_BASE_URL=https://test.pangbao.cuplay.top:9702`，且 SHALL NOT 包含真实生产口令。

#### Scenario: 新运维区分 prod 与 test env 文件

- **WHEN** 运维复制 `.env.test.example` 为 `.env.test` 并按注释填写
- **THEN** 其 SHALL 能将 `HISTORY_DB_LINK` 指向 `ai_voice_history_test` 且将 `IMAGE_TAG` 设为 `develop`，且 SHALL NOT 误用生产库 DSN

### Requirement: 测试栈 Compose MUST 支持 Redis 地址环境注入

测试 overlay 或 `.env.test.example` SHALL 文档化并通过 compose `environment` 注入 `GF_REDIS_DEFAULT_ADDRESS`，指向测试 Redis cluster 三主种子（测试网络内服务名与端口，与 prod 物理隔离）。

#### Scenario: 测试 gateway-app 使用 test Redis

- **WHEN** 测试栈 gateway-app 启动且 `GF_REDIS_DEFAULT_ADDRESS` 已按 `.env.test.example` 配置为 test cluster 种子
- **THEN** 版本检查等 Redis 缓存 SHALL 读写 test cluster，SHALL NOT 依赖 prod cluster 的节点地址

## MODIFIED Requirements

### Requirement: 参考微服务 Compose MUST 支持通过环境变量注入各服务 MySQL 连接串

`manifest/docker/docker-compose.microservices.yml`（或其后继官方等价文件）SHALL 允许在**不修改已提交 YAML 内口令占位**的前提下，通过环境变量为 `history-service`、`device-service`、`voice-service`、`worker`、`ucg-service` 及 `gateway-app`（`APP_DB_LINK`）注入数据库连接：其中 history/device/voice/ucg SHALL 分别支持 `HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK`、`UCG_DB_LINK` 覆盖默认库；worker SHALL 支持 `WORKER_OUTBOX_DB_LINK`（由 cmd 写入 `GF_DATABASE_OUTBOX_LINK`）；gateway-app SHALL 支持 `APP_DB_LINK` 覆盖 `database.app`。prod/test 分环境 `.env` 文件 SHALL 分别注入对应库名。

#### Scenario: device 使用注入的 link 启动

- **WHEN** 部署者在启动 Compose 前设置 `DEVICE_DB_LINK` 为合法 MySQL DSN
- **THEN** `device-service` 进程 SHALL 使用该 DSN 作为 `GF_DATABASE_DEFAULT_LINK`，而不依赖仅写在镜像内配置文件中的占位地址

#### Scenario: 测试 worker 使用 test outbox 库

- **WHEN** 测试栈设置 `WORKER_OUTBOX_DB_LINK` 指向 `ai_voice_worker_test`
- **THEN** worker-service SHALL 使用该 DSN 作为 outbox 库连接，SHALL NOT 写入生产 `ai_voice_worker`
