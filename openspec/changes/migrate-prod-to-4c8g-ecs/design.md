## Context

当前生产与测试同机（约 2C2G）：宿主机 MySQL 同时承载 `ai_voice_*` 与 `ai_voice_*_test`，Compose 双栈 + Redis（prod cluster / test standalone）争内存；`resources.prod.yml` 与 MySQL `innodb_buffer_pool` 均为 survival 档（约 192–256M）。

应用侧已约定：各服务 `*_DB_LINK` 使用占位符 `mysql-host`，真实主机仅由 `MYSQL_TCP_HOST` 覆盖（`internal/platform/dbcfg`）。改库地址不改六条 DSN 主体。

本设计落实 explore 已拍板方案：全新 4C8G ECS 专跑生产；旧机专跑测试；生产 Redis 空集群冷启；新机 MySQL buffer 首日 1G。

## Goals / Non-Goals

**Goals:**

- 文档化可执行的生产换机切流步骤与验收项。
- 明确新旧机职责、库迁移范围、env 改址点、Redis 冷启语义与用户可感知后果。
- 更新内存档位说明：4C8G prod-only + buffer 1G 起点；原地放宽 prod compose / Redis limits。
- 切流后旧机仅 test，旧生产库短期只读回滚。

**Non-Goals:**

- 不改造业务代码或 API 版本。
- 不迁测试栈、不迁 `*_test` 库、不改 `resources.test.yml`。
- 不把生产 MySQL 拆到 RDS（仍为本机 MySQL）。
- 不迁生产 Redis AOF/volume（明确冷启）。
- 不在本 change 内修改仓内真实 `.env.prod` 密钥或写入新机公网 IP（由运维在维护窗口填写）。

## Decisions

### D1 — 拓扑：新机 prod-only / 旧机 test-only

- **选择**：Docker + MySQL 随生产整迁新机；test Compose 与 `ai_voice_*_test` 留守旧机。
- **理由**：一次拆开双栈争抢；test 数据无需 dump；旧 MySQL 自然成为 test + 回滚副本。
- **备选**：同机垂直升配（IP/盘可能不变，但 test 仍与 prod 争资源）——已否决。

### D2 — 改库地址只改 `MYSQL_TCP_HOST`

- **选择**：新机 `.env.prod` 将 `MYSQL_TCP_HOST` 设为容器可达的新机地址（同机常用宿主机内网 IP，禁止容器内误用「他机」或错误的 `127.0.0.1` 语义时需按现网惯例）；`*_DB_LINK` 保持 `mysql-host` + `ai_voice_*`。旧机 `.env.test` 保持指向旧 MySQL + `_test` 库名。
- **理由**：与现有 env 隔离设计一致（runbook：「改库地址只改此行」）。
- **备选**：改写六条 DSN host——易漏、已否决。

### D3 — 新机 MySQL `innodb_buffer_pool_size=1G`（首日）

- **选择**：4C8G 同机 prod-only 首日固定 **1G**；`max_connections` 维持 100–150，不猛加。旧机 MySQL 维持约 256M，不跟调。
- **理由**：8G 余量下 1G 明显优于 survival 256M，又为 OS/Docker/Redis 留余量；观察 `available` 与慢查询后再考虑升到 2G。
- **备选**：首日 2G——余量更紧，延后；已否决为首日方案。

### D4 — 生产 Redis 空集群冷启

- **选择**：新机新建 Redis Cluster volume，`up` 后执行一次 `cluster create`；不从旧机拷 AOF。
- **理由**：避免跨机 cluster 元数据/槽位风险；与 `redis-disaster-recovery.md`「刻意 `down -v`」后果一致且可预期。
- **冷启后果（须写入 runbook）**：refresh token 失效需重登；语音/诊所 session 清空；AI 月额度 Redis 计数归零；Feed 索引/快照需 lazy warm；私信与主数据依赖 MySQL，不丢正文。
- **备选**：迁 volume——复杂度高，已否决。

### D5 — 原地改 prod compose（不新建 overlay）

- **选择**：MUST 更新 `release-deploy-and-run.md` 换机专节与 `memory-sizing-guide.md` 4C8G 表。MUST **原地**更新 `docker-compose.resources.prod.yml`（voice 512–768m、gateway/ucg 256m 等 + 文件头注明适用 4C8G prod-only）；按需 **原地**更新 `docker-compose.redis-cluster.yml`（节点 maxmemory ~200mb / mem_limit）。MUST NOT 新增 `resources.prod.4c8g.yml` 等平行文件。`resources.test.yml` 不改。
- **理由**：换机后 prod 默认环境即 4C8G；runbook 启动命令只认一个 `resources.prod.yml`，双文件易漂移。旧 2G survival 数字仅作 git 历史，不再作为生产目标态。

### D6 — 静态资源与域名

- **选择**：`/ai_talk_images`、`/apk/ai_talk` 随 prod 拷至新机；`*_test` 目录留旧机。公网 `www`（或等价生产域名）切到新机；`test` 域名仍指旧机。
- **理由**：与现有 prod/test 静态路径对照表一致。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| DNS/反代串线，写回旧机 `ai_voice_*` | 切流后旧机 **prod compose 必须 down**；验收 www 的 `appDatabase=ai_voice_app` 且日志主机为新 IP；test 侧 `ai_voice_app_test` + 旧 IP |
| Docker 无法连本机 MySQL（bind-address / 安全组） | 新机提前用容器网段探测 3306；优先内网；3306 不对公网全开 |
| dump/restore 窗口内写入丢失 | 停写（down prod 或摘反代）后做最终增量或二次 dump |
| Redis 冷启导致全员重登与 AI 额度重置 | 维护窗口公告；接受一次性「换机」副作用；额度敏感可择日靠近月初 |
| 新机 buffer 1G 仍不足或过大 | 巡检 `free` / 慢查询后再调 1.5–2G 或回调 |
| 旧机 drop 生产库过早 | 约定保留 N 天只读回滚后再 drop |

## Migration Plan

1. **新机准备**：安装 MySQL（buffer=1G）、Docker、目录、安全组；拉取镜像；写入 `.env.prod`（`MYSQL_TCP_HOST`=新机可达地址）。
2. **建库导入**：创建 `ai_voice_*`；从旧机 dump 生产库并导入；校验表行数抽样。
3. **Redis**：空集群 `up` + `cluster create` → `cluster_state:ok`。
4. **冒烟**：新机起 prod 栈，临时 hosts/IP 验收（不改公网 DNS）。
5. **停写**：旧机 down prod 微服务（保留 MySQL 与 test 栈）。
6. **最终同步**（若需要）后切 DNS/Nginx 至新机。
7. **验收**：生产日志主机、库名、登录/私信/Feed；test 域名仍旧机。
8. **回滚**：DNS 切回旧机并临时 up 旧 prod（指向旧 `ai_voice_*`）；新机保留现场。
9. **收尾**：稳定后停旧 prod Redis；N 天后 drop 旧 `ai_voice_*`（保留 `_test`）。

## Open Questions

- 旧生产库只读保留天数（建议 7 天，运维最终确认）。
