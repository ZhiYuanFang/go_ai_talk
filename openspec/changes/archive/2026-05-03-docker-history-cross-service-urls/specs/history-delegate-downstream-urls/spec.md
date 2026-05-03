## ADDED Requirements

### Requirement: history-service 在隔离网络运行时 MUST 使用可路由的下游 HTTP 基址

当 `history-service` 进程运行在与其他领域服务**不同的网络命名空间**（例如 Docker 独立容器、Kubernetes Pod）且需要通过 HTTP 委派访问 `voice-service` 或 `device-service` 时，其 MUST 通过环境变量 `VOICE_SERVICE_URL` 与 `DEVICE_SERVICE_URL` 配置为可在该命名空间内解析并路由到目标服务的基址（例如同一编排系统中的服务 DNS 名），MUST NOT 依赖指向本容器 loopback 的默认基址（如 `http://127.0.0.1:9802`）作为跨容器访问手段。

#### Scenario: Docker Compose 微服务栈中 history 访问 device

- **WHEN** `history-service` 与 `device-service` 作为不同容器加入同一用户定义 bridge/overlay 网络，且 history 需调用 device 的 HTTP API
- **THEN** 部署配置 SHALL 为 history 设置 `DEVICE_SERVICE_URL`（例如 `http://device-service:9803`），使得 TCP 连接目标为 device 容器而非 history 容器自身

#### Scenario: Docker Compose 微服务栈中 history 访问 voice

- **WHEN** `history-service` 与 `voice-service` 作为不同容器在同一编排网络内，且 history 需调用 voice 的 HTTP API
- **THEN** 部署配置 SHALL 为 history 设置 `VOICE_SERVICE_URL`（例如 `http://voice-service:9802`），使得请求到达 voice 服务监听端口

### Requirement: 仓库参考 Compose 中 history 段落 SHALL 与 voice 下游配置语义一致

`manifest/docker/docker-compose.microservices.yml`（或其后继等价的官方微服务 Compose 参考）中，若同时定义 `history-service` 与 `voice-service`、`device-service`，则 `history-service` 的环境变量段落 SHALL 包含与同文件内其他服务一致的、基于服务名的 `VOICE_SERVICE_URL` 与 `DEVICE_SERVICE_URL`，以避免开箱即用栈中出现 history 误连 `127.0.0.1` 的失败。

#### Scenario: 审查者对比 voice 与 history 环境块

- **WHEN** 审查者检查微服务 Compose 文件中 voice 已配置 `DEVICE_SERVICE_URL` 指向 `device-service`
- **THEN** 其 SHALL 能在 history 段落找到对 voice、device 的显式 URL 配置，且主机部分为 compose 服务名而非 `127.0.0.1`
