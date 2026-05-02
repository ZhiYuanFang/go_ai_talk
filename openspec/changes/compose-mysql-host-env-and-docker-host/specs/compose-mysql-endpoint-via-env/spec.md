## ADDED Requirements

### Requirement: 参考微服务 Compose MUST 支持通过环境变量注入各服务 MySQL 连接串

`manifest/docker/docker-compose.microservices.yml`（或其后继官方等价文件）SHALL 允许在**不修改已提交 YAML 内口令占位**的前提下，通过环境变量为 `history-service`、`device-service`、`voice-service`、`worker` 注入数据库连接：其中 history/device/voice SHALL 分别支持 `HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK` 覆盖默认库；worker SHALL 支持通过 `GF_DATABASE_DEFAULT_LINK` 覆盖其默认库（与现有 Gf 约定一致）。

#### Scenario: device 使用注入的 link 启动

- **WHEN** 部署者在启动 Compose 前设置 `DEVICE_DB_LINK` 为合法 MySQL DSN
- **THEN** `device-service` 进程 SHALL 使用该 DSN 作为 `GF_DATABASE_DEFAULT_LINK`，而不依赖仅写在镜像内配置文件中的占位地址

### Requirement: 参考 Compose MUST 为访问宿主机 MySQL 提供 host.docker.internal 解析

当 MySQL 监听在 **运行 Docker 的宿主机** 上且业务容器使用 bridge 网络时，参考 Compose 中需访问该 MySQL 的服务 SHALL 配置 `extra_hosts`，使主机名 `host.docker.internal` 解析到宿主机（例如 `host-gateway` 语义），以便连接串中使用 `tcp(host.docker.internal:3306)` 等地址时行为可预期。

#### Scenario: Linux 上 compose up 后容器解析 host.docker.internal

- **WHEN** 在支持 `host-gateway` 的 Docker Engine 上执行 `docker compose up` 使用该参考文件
- **THEN** 业务容器内 SHALL 能将 `host.docker.internal` 解析到宿主机侧地址，从而可与宿主机上监听的 MySQL 建立 TCP 连接（在 DSN 已正确配置且 mysqld 对 Docker 网桥来源放行时）

### Requirement: 仓库 MUST 提供无密钥的 Compose 数据库环境样例

仓库 SHALL 提供一份可复制为本地 `.env` 的示例文件（例如 `manifest/docker/.env.example`），其中 SHALL 用中文或英文注释说明：**MySQL 与 Docker 同机**时推荐将主机设为 `host.docker.internal`；**MySQL 在其它机器**时将主机设为从容器网络可达的 DNS 或 IP（如 RDS、内网 IP），且 SHALL NOT 包含真实生产口令。

#### Scenario: 新成员首次接 Compose 栈

- **WHEN** 开发者复制示例为 `.env` 并按注释填写自己的 MySQL 拓扑
- **THEN** 其 SHALL 能区分同机与异机两种填法，且无需从 git 历史中寻找口令
