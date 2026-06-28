## MODIFIED Requirements

### Requirement: notify-service SHALL run as dedicated microservice

平台 SHALL 提供 **`notify-service`** 进程（原 app-status 维护通知能力），监听 `NOTIFY_SERVICE_ADDR` 默认 `:9806`，加载 `manifest/config/config.notify-service.yaml`。进程 **MUST NOT** 依赖 MySQL、Redis 或 RabbitMQ 启动探活。Docker 镜像名与 ACR 仓库名 **MUST** 为 `notify-service`。

#### Scenario: 启动与配置隔离

- **WHEN** 启动 notify-service 且未设置 `GF_GCFG_FILE`
- **THEN** 进程 SHALL 加载 `config.notify-service.yaml`，且配置 **MUST NOT** 含 `database` 段

#### Scenario: HTTP 契约保持不变

- **WHEN** 客户端 `GET /app/api/status/banner`
- **THEN** 路径与响应语义 SHALL 与 rename 前一致

### Requirement: docker-acr workflow SHALL build notify-service image

`.github/workflows/docker-acr.yml` MUST 将 `notify-service` 纳入全量构建矩阵（与 gateway、ucg 等并列，共 8 服务）。构建别名 **MUST** 接受 `notify` 与 `notify-service`，映射 Dockerfile `manifest/docker/Dockerfile.notify-service`。Tag `vX.Y.Z+notify` **MUST** 仅构建并 push `notify-service` 镜像。

#### Scenario: 全量 tag 含 notify-service

- **WHEN** push tag `v1.0.0-rc.1`（无 `+` 后缀）
- **THEN** CI matrix **MUST** 包含 `notify-service` 且 push `${REGISTRY}/notify-service:${tag}`

#### Scenario: 单服务 notify 构建

- **WHEN** push tag `v1.0.0-rc.2+notify` 或 workflow_dispatch `services=notify`
- **THEN** CI **MUST** 仅 build/push `notify-service` 镜像
