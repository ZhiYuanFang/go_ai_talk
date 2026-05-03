## Context

各服务 `cmd/*/main.go` 已支持用环境覆盖默认库：`HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK` 会写入 `GF_DATABASE_DEFAULT_LINK`。`worker-service` 未单独封装 `WORKER_DB_LINK`，可直接依赖 Gf 对 `GF_DATABASE_DEFAULT_LINK` 的环境读取，或在 compose 中仅设置该变量。

## Goals / Non-Goals

**Goals:**

- Compose 参考栈在 **MySQL 跑在 Docker 宿主机** 时，默认推荐 **`host.docker.internal:3306`**（配合 `extra_hosts: host-gateway`），避免容器访问宿主机公网 IP 的发夹失败。
- **MySQL 在独立主机**（云 RDS、另一台 VM、同 VPC 数据库）时，同一机制仍可用：将连接串中的 host 设为 **从容器内 `ping`/`nc` 可达** 的 DNS 或 IP（通常是内网地址或 RDS endpoint），**不需要** `host.docker.internal`。
- 敏感信息不进 git：仅提交 `.env.example`，实际 `.env` 由部署方生成并加入 `.gitignore`（若尚未忽略则任务中说明）。

**Non-Goals:**

- 不在本变更中删除或强制改写各 `config.*.yaml` 内已有占位 link（避免破坏未使用 compose 的本地启动）。
- 不为 `database.history_relay` 引入新的专用环境变量（若需覆盖可继续用挂载整份 yaml 或后续独立变更）。

## Decisions

1. **`extra_hosts`**  
   - **决策**：在 `history-service`、`device-service`、`voice-service`、`worker` 上增加 `host.docker.internal:host-gateway`。  
   - **理由**：Linux 与 Docker Desktop 均支持 `host-gateway` 特解；无该字段时 Linux 上 `host.docker.internal` 可能未定义。  
   - **异机 MySQL**：`extra_hosts` 无害；连接串 host 指向远端即可，不经过 `host.docker.internal`。

2. **环境变量命名**  
   - **决策**：复用已有 `HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK`；worker 使用 `GF_DATABASE_DEFAULT_LINK`。  
   - **理由**：与现有 `main.go` 一致，无需改二进制。

3. **`.env.example` 形态**  
   - **决策**：提供 `MYSQL_HOST` / `MYSQL_PORT` 与完整 `*_DB_LINK` 示例各一行注释说明两种拓扑；Compose 使用 `env_file: [.env]` 或 `environment` + `variable substitution`，以团队习惯选一种并在 runbook 写明。  
   - **备选**：仅用完整 link 四个变量，不拆 host/port——更简单，任务实现时取更简方案。

## Risks / Trade-offs

- **[Risk]** 开发者未复制 `.env` 时仍走 yaml 内公网 IP → 行为与变更前相同；缓解：runbook 强调 Compose 启动前检查 env。  
- **[Risk]** 极老 Docker 无 `host-gateway` → 需升级 Docker 或改用手写宿主机 bridge IP；在 runbook 脚注说明。

## Migration Plan

1. 部署方创建 `.env`（自 `.env.example`）。  
2. `docker compose --env-file .env up`。  
3. 回滚：移除 env_file / 环境块中的 link，恢复仅 yaml。

## Open Questions

- 是否在仓库根 `.gitignore` 增加 `manifest/docker/.env`：若已全局忽略 `.env` 则不必。
