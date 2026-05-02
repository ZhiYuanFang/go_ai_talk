## Why

业务容器在 `docker-compose.microservices.yml` 中通过仓库内 YAML 写死的公网 IP（如 `120.55.50.105:3306`）连接 MySQL 时，若 MySQL 实际跑在**同一台宿主机**上，常出现 **发夹 NAT / 路由回环** 问题：外网其它机器可连该公网 IP，但容器内访问「本机公网 IP」却 `connection refused`。同时，MySQL 与 Docker **分属不同机器**时，应使用对**容器出网路径**可达的 DNS/内网 IP，与「宿主机回环」方案不同，但应通过**同一套环境变量注入**覆盖，避免再改代码。

## What Changes

- 在参考 Compose 中为需连库的服务增加 **`extra_hosts: host.docker.internal:host-gateway`**（Docker 20.10+ 通用写法），使容器内可使用固定主机名指向宿主机。
- 通过 **`HISTORY_DB_LINK` / `DEVICE_DB_LINK` / `VOICE_DB_LINK`**（及 worker 的 **`GF_DATABASE_DEFAULT_LINK`**，因 worker 入口未封装专用变量）在 Compose 中从 **`.env` 注入**，覆盖各服务 `GF_DATABASE_DEFAULT_LINK`；提供 **`manifest/docker/.env.example`**（无真实口令）说明「同机 MySQL」与「异机 MySQL」填法。
- 更新 **`docs/runbooks/release-deploy-and-run.md`**：说明两种拓扑下 `MYSQL_HOST` 的取值差异；仓库内 `manifest/config/config.*.yaml` 中的占位 link 仍可保留作本地 go run 默认，Compose 以 env 为准。

## Capabilities

### New Capabilities

- `compose-mysql-endpoint-via-env`：约束参考 Compose 下 MySQL 连接串通过环境注入，并区分「库在宿主机」与「库在远端」的解析方式。

### Modified Capabilities

- （无）

## Impact

- `manifest/docker/docker-compose.microservices.yml`
- 新增 `manifest/docker/.env.example`（或等价路径）
- `docs/runbooks/release-deploy-and-run.md`
- 可选：各 `config.*.yaml` 顶部一行注释指向 `.env.example`
