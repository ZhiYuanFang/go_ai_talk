## 1. Compose 与示例环境

- [x] 1.1 在 `manifest/docker/docker-compose.microservices.yml` 为 `history-service`、`device-service`、`voice-service`、`worker` 增加 `extra_hosts: ["host.docker.internal:host-gateway"]`（或等价 YAML 列表写法）。
- [x] 1.2 采用 `docker compose --env-file manifest/docker/.env` 插值：各服务 `environment` 中 `HISTORY_DB_LINK` / `DEVICE_DB_LINK` / `VOICE_DB_LINK` / `WORKER_DB_LINK` 使用 `${VAR:-}`；worker 在 `cmd/worker-service/main.go` 增加非空 `WORKER_DB_LINK` → `GF_DATABASE_DEFAULT_LINK`，避免空字符串覆盖 Gf 配置。
- [x] 1.3 新增 `manifest/docker/.env.example`：注释说明同机用 `host.docker.internal`、异机用远端主机名/IP；示例 DSN 使用占位口令与占位库名。

## 2. 文档与忽略规则

- [x] 2.1 更新 `docs/runbooks/release-deploy-and-run.md`：Compose 小节增加「复制 `.env.example` → `.env`」、两种 MySQL 拓扑、`extra_hosts` 用途；说明异机时 `MYSQL_HOST` 填 RDS/内网即可，无需改 compose 结构。
- [x] 2.2 若仓库尚未忽略 `manifest/docker/.env`，在根 `.gitignore` 增加该项（避免误提交密钥）。

## 3. 校验

- [x] 3.1 执行 `openspec validate compose-mysql-host-env-and-docker-host --strict`。
