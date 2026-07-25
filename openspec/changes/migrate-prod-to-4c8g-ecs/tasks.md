## 1. 内存档位文档

- [x] 1.1 更新 `docs/runbooks/memory-sizing-guide.md`：新增 **4C8G / prod-only + 本机 MySQL** 档位表
- [x] 1.2 在该档位中写明新机 `innodb_buffer_pool_size` **首日 1G**，并注明旧机 test 维持约 256M、不跟调
- [x] 1.3 在档位表中写明 compose/Redis 放宽数值（voice 512–768m、gateway/ucg 256m、Redis 节点 maxmemory ~200mb），与原地更新的 prod compose 一致

## 2. 生产换机 runbook

- [x] 2.1 在 `docs/runbooks/release-deploy-and-run.md` 新增「生产迁至 4C8G ECS」专节（或附录）
- [x] 2.2 专节写明拓扑：新机 prod-only；旧机 test-only；仅迁 `ai_voice_*`；`*_test` 与 test 栈留守
- [x] 2.3 专节写明改址：只改 `.env.prod` 的 `MYSQL_TCP_HOST`；`.env.test` 不跟迁
- [x] 2.4 专节写明新机 MySQL 配置片段：`innodb_buffer_pool_size=1G`、`max_connections` 建议
- [x] 2.5 专节写明 Redis 空集群冷启步骤与后果（重登、会话清空、AI 月额度归零、Feed 偏冷、私信 MySQL fallback）
- [x] 2.6 专节写明切流顺序：准备 → dump/restore → 冒烟 → 停旧 prod → DNS → 验收 → 回滚 → 旧库保留后再 drop
- [x] 2.7 专节写明验收命令（库名/`MYSQL_TCP_HOST`/cluster_state/www vs test 串线检查）并交叉引用 `redis-disaster-recovery.md`

## 3. 生产资源 compose（原地更新）

- [x] 3.1 原地更新 `manifest/docker/docker-compose.resources.prod.yml`：按 design 放宽 mem_limit/GOMEMLIMIT（voice 512–768m、gateway/ucg 256m 等），文件头中文注释改为适用 4C8G prod-only；**不新增**平行 overlay 文件
- [x] 3.2 按需原地更新 `manifest/docker/docker-compose.redis-cluster.yml`（节点 maxmemory ~200mb / mem_limit），并在换机 runbook 说明：prod 默认即本文件；`resources.test.yml` 不改

## 4. 收尾核对

- [x] 4.1 通读专节与档位文档，确认与 `prod-ecs-cutover` spec 五条需求一致（职责分离、改址、buffer 1G、Redis 冷启、文档齐全）
- [x] 4.2 确认未改业务 API/代码路径，且未把真实主机 IP 或密钥写入 runbook 示例（示例用占位符）
