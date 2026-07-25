## Why

生产与测试长期同挤在约 2C2G ECS 上，内存与 MySQL buffer 处于 survival 档，语音/Feed 峰值易触顶。现采购全新 **4C8G** ECS，将 **生产 Docker 栈与生产 MySQL** 迁至新机并放宽资源；测试栈与测试库留守旧机，降低双栈争抢并保留旧库回滚窗口。

## What Changes

- 新增「生产换机切流」runbook：新机装 MySQL + Compose、仅迁 `ai_voice_*` 生产库、DNS/反代切流、旧机只留 test。
- 新机 MySQL **首日** `innodb_buffer_pool_size=1G`（旧机 test 维持小 buffer，不跟调）。
- 生产 Redis：**空集群冷启**（新 volume + 一次 `cluster create`）；明确冷启后果（重登、语音会话清空、AI 月额度计数归零、Feed 短暂偏冷；私信/主数据走 MySQL）。
- 更新 `memory-sizing-guide.md`：补 **4C8G / prod-only** 档位建议（含 buffer 1G 起点与 compose/Redis 放宽数值）。
- **原地更新** `docker-compose.resources.prod.yml`（及按需 `docker-compose.redis-cluster.yml` 的 maxmemory），对齐 4C8G prod-only；**不新建**平行 overlay 文件；**不改** `resources.test.yml`（test 留旧机 2G）。
- **非 BREAKING** 应用 API：仅运维拓扑与环境变量（`.env.prod` 的 `MYSQL_TCP_HOST`）变化；接口契约不变。

## Capabilities

### New Capabilities

- `prod-ecs-cutover`：生产迁至独立 4C8G ECS 的切流约定（库迁移范围、env 改址、Redis 冷启语义、新旧机职责、验收与回滚）。

### Modified Capabilities

- （无）不修改既有业务能力的 REQUIREMENTS；本变更为部署/运维规格增量。

## Impact

- **文档**：`docs/runbooks/release-deploy-and-run.md`、`docs/runbooks/memory-sizing-guide.md`；可交叉引用 `redis-disaster-recovery.md`。
- **配置**：原地改 `manifest/docker/docker-compose.resources.prod.yml`；按需原地改 `docker-compose.redis-cluster.yml`（maxmemory / mem_limit）；启动命令仍只叠现有 `resources.prod.yml`。
- **环境（机房操作，非仓内密钥）**：新机 `.env.prod` 中 `MYSQL_TCP_HOST`；旧机 `.env.test` 保持指向旧 MySQL；旧机 prod compose down。
- **运行时影响**：切流后全员需重新登录；AI 当月 Redis 额度计数重置；旧机 `ai_voice_*` 保留作短期只读回滚后可 drop。
- **不影响**：App/HTTP API 版本、跨服务契约、test 库名 `_test` 约定、`MYSQL_TCP_HOST` + `mysql-host` 占位符机制本身。
