## Context

- `ucg-sim-user-service` 已落地 `sim-user-service`，任务周期写死在 `runtime.go`，T2 经 gateway 调用 `GET /ucg/app/api/feed/recommend`，触发 ucg `ListRecommendFeed` 的 N+1 读路径（每页约 40+ SQL + device HTTP）。
- 生产/测试同机双栈，共用 MySQL 实例（`MYSQL_TCP_HOST`），`max_connections` 约 100、`innodb_buffer_pool_size` 约 256M；长期开 sim 会经共享实例间接影响生产 recommend。
- 负责人决策：**数据库不迁移**；通过档位 B「智能轮询 + 轻量读路径」将 sim 压力控制在可长期运营范围。
- `envDuration()` 已在 `runtime.go` 存在但未接入；P1 在无 pending job 时已快速返回，但 ticker 仍为 1min 唤醒。

## Goals / Non-Goals

**Goals:**

- 各 sim 任务周期支持 `SIM_INTERVAL_*` env 覆盖，默认与现网一致。
- P1 自适应：无 pending → 长间隔；有 pending → 短间隔。
- T2 改走 ucg internal 轻量帖抽样，单请求返回 `postId + content`（及评论所需最小 media 元数据），避免 recommend N+1。
- E1 降频：循环 5min、窗口 15min（均可 env 覆盖）。
- sim 对 gateway-app 的出站 HTTP 全局限速（默认 2 req/s）。
- 各任务首次 tick 随机延迟 0–30min，避免启动齐射。
- runbook 补充长期运营 env 配方。

**Non-Goals:**

- 不迁移 MySQL；不调整测试/生产库隔离策略。
- 不优化真人 `feed/recommend` 读路径（批量 `GetPublicProfilesByWxIDs` 留独立 change）。
- 不将任务周期纳入 sim-admin 管理页（首期仅 env；admin 仍管 enabled/maxSimUsers/prompt）。
- 不调 ucg `hotScanIntervalSeconds` / `relayIntervalMs`（运维 yaml 层可自行配置，本 change 不强制改代码）。
- 不新增 `*_test.go`。

## Decisions

### 1. T2 帖源：新增 ucg internal `POST /ucg/internal/api/posts/sample`

**选择**：ucg internal 只读抽样 API，internal 密钥鉴权（与现有 `chat/send`、media upload 一致）。

**请求**：`{ "limit": 20 }`（默认 20，上限 50）。

**响应**：`{ "list": [{ "postId", "content", "mediaType", "coverObjectKey?" }] }` — 仅 `status=published`，按 `published_at DESC` 或随机抽样（实现采用 `ORDER BY published_at DESC LIMIT n` 后内存随机一条给 T2；或 SQL `ORDER BY RAND()` 仅 limit≤20 可接受）。

**查询**：单条 SQL 联结 `ucg_post` + 可选首图 `ucg_post_media`（`sort_order` 最小），**不**调 device、不填 author、不走 `postsFromResult`。

**备选**：T2 继续用 `GET posts/user` — 否决（仍走 `postsFromResult` N+1）；sim 直连 DAO — 否决（违反服务边界）。

### 2. 任务周期 env 命名

| 任务 | 环境变量 | 默认值 |
|------|----------|--------|
| T1 register | `SIM_INTERVAL_REGISTER` | `24h` |
| T2 comment | `SIM_INTERVAL_COMMENT` | `6h` |
| T3 post_image | `SIM_INTERVAL_POST_IMAGE` | `3h30m` |
| T4 post_video | `SIM_INTERVAL_POST_VIDEO` | `6h30m` |
| T5 chat | `SIM_INTERVAL_CHAT` | `1h` |
| T6 follow | `SIM_INTERVAL_FOLLOW` | `7h` |
| P1 active | `SIM_INTERVAL_VIDEO_POLL_ACTIVE` | `2m` |
| P1 idle | `SIM_INTERVAL_VIDEO_POLL_IDLE` | `10m` |

解析复用现有 `envDuration()`；非法值回退默认。±10% jitter 保留。

### 3. P1 自适应调度

**选择**：`runPeriodic` 改为 `runAdaptivePeriodic`（仅 P1）：每次 tick 结束根据 `ListPendingVideoJobs` 是否为空选择 **下一等待间隔**（idle vs active），不再固定 `IntervalVideoPoll`。

**备选**：双 goroutine 切换 — 否决（复杂度高）；完全去掉 P1 ticker、仅 T4 后手动 poll — 否决（视频完成延迟不可控）。

### 4. E1 参数 env 化

| 变量 | 默认 | 说明 |
|------|------|------|
| `SIM_EPHEMERAL_CHAT_LOOP` | `5m` | E1 循环间隔 |
| `SIM_EPHEMERAL_CHAT_WINDOW` | `15m` | E1 窗口总时长 |

`spawnEphemeralChat` 读取 `LoadRuntimeFlags` 或包级配置；硬停语义不变。

### 5. HTTP 全局限速

**选择**：`simuser` 包内 `rate.Limiter`（`golang.org/x/time/rate`），在 `appGet`/`appPost`/`appPut` 入口 `Wait(ctx)`；限额 `SIM_UCG_RATE_LIMIT_RPS`（默认 `2`，浮点），burst 默认 `4`。

仅限制经 **gateway-app** 的 App API 调用（含 T2 sample 若走 internal 则不经 limiter — internal 走 `ucgInternalPost` 单独通道，可选同限 2rps 到 ucg base）。

**备选**：每任务独立 limit — 否决（多 goroutine 仍可叠加超标）。

### 6. 首次 tick 错峰

`StartScheduler` 各 `runPeriodic`/`runAdaptivePeriodic` 在首次 `Jittered(interval)` 之前再加 `randomStartupDelay()`：`0` 至 `SIM_STARTUP_STAGGER_MAX`（默认 `30m`）均匀随机。

### 7. sim T2 调用链变更

```
T2: usernameLogin → ucgInternalPost("/ucg/internal/api/posts/sample") 
    → aimodel simVision → POST comments
```

不再调用 `feed/recommend`。

## Risks / Trade-offs

- **[Risk] T2 抽样非推荐算法排序，评论可能集中在最新帖** → 可接受；广场运营目标是互动密度，非复刻 recommend 分布。
- **[Risk] `ORDER BY RAND()` 在帖量大时偶发慢查询** → Mitigation：limit≤20 + `status=2` 索引 + 优先 `published_at DESC` 取前 20 再随机。
- **[Risk] 全局限速 2rps 拉长单次任务 wall time** → Mitigation：任务本身为小时级周期，可接受；env 可调高。
- **[Risk] 与 `ucg-sim-user-service` 归档前 spec 文本并行演进** → 本 change delta 明确 MODIFIED 条款；归档时合并。
- **[Risk] 共享 MySQL 下仍可能有残余争抢** → Mitigation：配合 `maxSimUsers` 与 runbook 运营配方；真人读优化另开 change。

## Migration Plan

1. 部署 ucg-service（含 internal sample API）→ 部署 sim-user-service（新轮询逻辑）→ 无需 DB 迁移。
2. 生产 `.env.prod` 追加变量（可先仅用默认值，等价现网行为除 E1/P1/T2 路径变更）。
3. **推荐生产配方**（长期运营）：
   ```bash
   SIM_USER_SERVICE_ENABLED=true
   SIM_INTERVAL_COMMENT=12h
   SIM_INTERVAL_POST_IMAGE=6h
   SIM_INTERVAL_VIDEO_POLL_IDLE=10m
   SIM_INTERVAL_VIDEO_POLL_ACTIVE=2m
   SIM_EPHEMERAL_CHAT_LOOP=5m
   SIM_EPHEMERAL_CHAT_WINDOW=15m
   SIM_UCG_RATE_LIMIT_RPS=2
   SIM_TASK_CHAT_ENABLED=false   # 初期关 E1，按需开
   ```
4. 回滚：回退镜像 tag；或仅 `SIM_USER_SERVICE_ENABLED=false` 并重启 sim-user-service。
5. 验收：`GET /sim/admin/api/status` 正常；T2 日志不再出现 `feed/recommend`；MySQL `Threads_running` 在开 sim 时低于此前峰值。

## Open Questions

- T2 抽样是否需排除「仅 sim 作者」帖以促真人帖互动 — 首期不做，实现简单优先；若运营需要再加 `excludeSimAuthors` 查询参数。
- internal sample 是否需计入 gateway usage — **否**，走 ucg internal，不经 gateway-app。
