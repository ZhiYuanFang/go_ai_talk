## ADDED Requirements

### Requirement: 仓库 SHALL 提供生产与测试双栈 Compose overlay

仓库 SHALL 在 `manifest/docker/` 提供 `docker-compose.microservices.prod.yml` 与 `docker-compose.microservices.test.yml`，与基线 `docker-compose.microservices.yml` 组合使用。prod/test overlay SHALL 使用 `${REGISTRY}/go-ai-talk/<service>:${IMAGE_TAG}` 引用镜像仓库，且 SHALL NOT 包含 `build` 段。基线文件 MAY 保留 `build` 与 `:local` 供本机开发。

#### Scenario: 测试栈从 registry pull 启动

- **WHEN** 运维设置 `REGISTRY`、 `IMAGE_TAG=develop` 并执行 `docker compose -f ...microservices.yml -f ...microservices.test.yml pull && up -d --no-build`
- **THEN** 各业务容器 SHALL 使用 registry 中 `:develop` 镜像启动，且 SHALL NOT 在宿主机执行源码 build

#### Scenario: 生产栈使用 semver tag

- **WHEN** 运维在 `.env.prod` 设置 `IMAGE_TAG=v1.0.0` 并 pull + up
- **THEN** 生产容器 SHALL 使用 `:v1.0.0` 镜像，且 SHALL NOT 使用 `:develop` 或 `:local`

### Requirement: 生产与测试 SHALL 使用独立 Docker 网络完全隔离

生产栈与测试栈 SHALL 分别仅加入独立的 external Docker 网络（约定名 `go-ai-talk-prod-net` 与 `go-ai-talk-test-net`）。同一宿主机上 prod 与 test 的中间件与微服务 SHALL NOT 共用同一 bridge 网络的 DNS 解析。

#### Scenario: test 网络内 rabbitmq 不可被 prod 容器解析

- **WHEN** prod 与 test 栈同时运行且各自 RabbitMQ 仅加入对应网络
- **THEN** prod 容器内 SHALL NOT 通过服务名 `rabbitmq` 解析到 test 的 RabbitMQ 实例

### Requirement: 测试栈 SHALL 独立 Redis Cluster 与 RabbitMQ

仓库 SHALL 提供 `docker-compose.redis-cluster.test.yml` 与 `docker-compose.rabbitmq.test.yml`。测试 Redis cluster 宿主机映射端口 SHALL 使用 17001–17006（或与 prod 7001–7006 不冲突的 documented 端口段）。测试 RabbitMQ 宿主机映射 SHALL 使用 5673/15673（或与 prod 5672/15672 不冲突的 documented 端口段）。测试微服务 SHALL 通过环境变量 `GF_REDIS_DEFAULT_ADDRESS` 与 `MQ_HTTP_API_BASE` 指向 test 网络内中间件。

#### Scenario: 测试 history 与 worker 使用 test RabbitMQ

- **WHEN** 测试栈 `history-service` 与 `worker` 已启动且 `OUTBOX_RELAY_ENABLED`/`MQ_CONSUMER_ENABLED` 为 true
- **THEN** 二者 SHALL 仅与 test 网络内 RabbitMQ 通信，且 prod worker SHALL NOT 消费 test 队列中的消息

### Requirement: 测试栈后端端口 SHALL 与生产错开

测试栈微服务宿主机端口映射 SHALL 为：gateway 19701、gateway-app 19702、history 19801、voice 19802、device 19803、ucg 19804、worker 19901（或与 runbook  documented 表一致且不与 prod 9701–9901 冲突）。测试栈 container_name SHALL 与 prod 不同（例如带 `-test` 后缀或使用 `COMPOSE_PROJECT_NAME` 前缀）。

#### Scenario: 同机 prod 与 test 同时监听

- **WHEN** prod 与 test 栈同时 up
- **THEN** 宿主机 SHALL 可同时访问 `127.0.0.1:9701`（prod gateway）与 `127.0.0.1:19701`（test gateway）且无端口绑定冲突

### Requirement: 测试对外访问形态 SHALL 与生产一致

测试环境对外入口 SHALL 为 `https://test.pangbao.cuplay.top:9701`（主网关）与 `https://test.pangbao.cuplay.top:9702`（App 网关），由 Nginx（或等价反代）转发至测试后端 19701/19702。测试栈 SHALL 设置 `GATEWAY_APP_PUBLIC_BASE_URL` 为 `https://test.pangbao.cuplay.top:9702`（或 runbook documented 等价 HTTPS 基址）。

#### Scenario: 客户端仅换域名访问测试 App 网关

- **WHEN** 客户端将生产基址 `www.pangbao.cuplay.top:9702` 换为 `test.pangbao.cuplay.top:9702` 且路径不变（如 `/device/app/api/version/check`）
- **THEN** 请求 SHALL 到达测试 gateway-app 且 API 路径语义与生产一致

### Requirement: 镜像 tag 策略 SHALL 区分测试浮动与生产钉死

测试默认 `IMAGE_TAG=develop`（CI 覆盖的浮动 tag）。生产 MUST 使用 semver release tag（如 `v1.0.0`），MUST NOT 在生产 `.env.prod` 中使用 `develop` 或 `latest`。CI SHOULD 同时 push 不可变 `:<git-sha>` tag 供排错。

#### Scenario: 生产 env 拒绝 develop

- **WHEN** 运维检查生产部署配置
- **THEN** `.env.prod` 中 `IMAGE_TAG` SHALL 为 semver 形式且 SHALL NOT 等于 `develop`
