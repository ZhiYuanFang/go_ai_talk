## 1. 生产 Redis 3 主 0 从

- [x] 1.1 修改 `docker-compose.redis-cluster.yml`：仅保留 `redis-node-1..3` 与 volume 1–3；更新注释与 `cluster create` 示例（`--cluster-replicas 0`）
- [x] 1.2 同步 `docs/runbooks/redis-cluster-local.md` 与 `hack/redis-cluster-init.ps1`（若仍引用 6 节点则改为 3 节点或注明本地可选 profile）
- [x] 1.3 runbook 补充生产 Redis 迁移：`down -v` → up → cluster create → 验收 `CLUSTER INFO`

## 2. 测试 Redis 单机

- [x] 2.1 新增 `manifest/docker/docker-compose.redis-standalone.test.yml`（`redis-test:6379`，`go-ai-talk-test-net`，可选宿主机 `16379:6379`）
- [x] 2.2 基线 `microservices.yml` 为 gateway/gateway-app/history/voice/device/worker/ucg 增加 `${GF_REDIS_DEFAULT_ADDRESS:-}` 环境变量
- [x] 2.3 更新 `.env.test.example` / `.env.test` 注释与 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`
- [x] 2.4 runbook B.1/B.2：测试 Redis 改 standalone；删除/归档六节点 test cluster 为默认路径；启动前清理块同步
- [x] 2.5 将 `docker-compose.redis-cluster.test.yml` 标记 deprecated 或移入 `hack/` 归档说明（runbook 不再引用为默认）

## 3. 容器资源 limits

- [x] 3.1 新增 `docker-compose.resources.prod.yml`：prod Redis×3、rabbit、7 微服务 limits（design 表默认值）
- [x] 3.2 新增 `docker-compose.resources.test.yml`：test redis-test、rabbit、7 微服务 limits（voice-test 512M）
- [x] 3.3 runbook A/B/C 启动命令追加 `-f resources.{prod,test}.yml`；附录增加 limits 表与 OOM 排错

## 4. Runbook 2G survival

- [x] 4.1 `release-deploy-and-run.md` 新增 2G 专节：MySQL buffer_pool、swap 建议、双栈内存粗算、ASR 验收约定
- [x] 4.2 更新附录生产/测试对照：Redis 拓扑（prod 3 节点 / test 单机）、端口（去掉 17001–17006 六端口描述或改为 16379）

## 5. 验收（手工）

- [x] 5.1 生产 3 节点 cluster `cluster_state:ok`；测试 `redis-cli PING` 无需 cluster create
- [x] 5.2 `docker stats` 确认 prod/test 容器存在 mem/cpu 上限
- [x] 5.3 测试域名 ASR/WS 验收通过；prod 微服务保持 Up
