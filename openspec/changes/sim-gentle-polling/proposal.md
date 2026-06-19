## Why

生产环境需长期开启 `sim-user-service` 维持 UCG 广场活跃，但测试与生产共用同一 MySQL 实例（`120.55.50.105`）。实测 `SIM_USER_SERVICE_ENABLED=true` 时，sim 任务扇出整条 UCG 业务链（发帖审核、推荐重算、聊天持久化），叠加 recommend Feed 的 N+1 查询，导致测试/生产 recommend 接口超时或明显变慢。数据库短期不迁移，须通过**智能轮询与轻量读路径**将 sim 对共享 MySQL 的压力控制在可长期运营的范围。

## What Changes

- **sim 任务周期 env 可配置**：各任务间隔支持 `SIM_INTERVAL_*` 环境变量覆盖，默认值与现网写死周期一致，向后兼容。
- **P1 视频轮询自适应**：无 pending job 时延长休眠（默认 10min）；存在 pending 时缩短为活跃轮询（默认 2min）；仍受 `SIM_VIDEO_POLL_ENABLED` 控制。
- **T2 评论改走轻量帖源**：不再调用 `GET /ucg/app/api/feed/recommend`；改为 ucg internal 轻量「已发布帖抽样」契约（单请求、批量返回、无 author/media 富化）。
- **E1 聊天窗口降频**：循环间隔默认 1min → **5min**；窗口时长默认 30min → **15min**；行为语义不变（到期硬停、同 peer 去重）。
- **sim→gateway HTTP 全局限速**：token bucket 限制对 gateway-app/ucg 方向请求（默认 **2 req/s**），防止多任务 goroutine 叠加打满 MySQL 连接。
- **任务首次 tick 错峰**：各周期任务启动后增加 0–30min 随机初始延迟，避免多任务同时齐射。
- **运营文档**：runbook 补充「长期开 sim + 共享 MySQL」推荐 env 配方与观测指标；`.env.example` 增加新变量说明。
- **不迁移数据库**；不改动 recommend Feed 真人读路径（批量查询优化留独立 change）。

## Capabilities

### New Capabilities

- `ucg-sim-feed-sample`：ucg-service internal 轻量已发布帖抽样 API，供 sim T2 选帖，避免 recommend N+1。

### Modified Capabilities

- `sim-user-service`：任务周期 env 化、P1 自适应轮询、E1 降频参数、HTTP 全局限速、首次 tick 错峰；T2 改用 `ucg-sim-feed-sample`。

## Impact

- **代码**：`internal/services/simuser/`（runtime、scheduler、tasks、clients）、`internal/services/ucg/`（新 internal handler）、`internal/controller/`（注册路由）、`api/v1/`（契约）。
- **配置**：`manifest/docker/.env.example`、runbook；可选生产 overlay 示例 env（`SIM_INTERVAL_*`、`SIM_UCG_RATE_LIMIT_RPS`）。
- **进程**：仅 `sim-user-service` 与 `ucg-service`；不新增微服务。
- **数据库**：无新表；T2 新接口只读 `ucg_post`（+ 可选 `ucg_post_recommend` 排序），禁止跨库。
- **OpenSpec 强制项**：背景任务周期/开关变更须在 design 中声明；不新增 `*_test.go`；不默认引入 Redis 读缓存。
