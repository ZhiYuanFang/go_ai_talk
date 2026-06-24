## Why

UCG 推荐 Feed 当前依赖 MySQL `ucg_post_recommend` JOIN 排序，无法在请求时融合用户位置信息，且读路径 N+1 严重。产品需要在**不改变「全站推荐排序」语义**的前提下，将距离作为**复合推荐分的一项**（方案 2：additive composite），并在客户端展示精确距离；同时以 Redis GEO/ZSET + 快照缓存支撑高性能 Feed 读路径。负责人已确认引入 Redis 读缓存（帖子快照、profile 快照、liked SET、Feed session）。

## What Changes

- **复合推荐分**：请求时 `finalScore = baseScore + distanceTerm`（viewer 与 post 均有坐标时）；`baseScore` 仅存 Redis ZSET `ucg:recommend:score`，**停止写入** MySQL `ucg_post_recommend`（表可保留不用）。
- **距离展示**：API 返回 `distanceMeters` 字符串（米，精度数值），无文案标签；无坐标时省略字段。
- **Feed 读路径**：Redis GEO `ucg:feed:geo` + ZSET 候选 + 半径阶梯扩展 `[50,100,200,500,unlimited]` km；cursor 分页（无 `total`）；服务端 Feed session 防重复（`ucg:feed:session:{sessionId}` SET，TTL 30min）。
- **快照与社交**：帖子/作者 profile JSON 快照键；`likedByMe` 经 Redis SET `ucg:user:{wxId}:liked-posts` 批量 SISMEMBER，Feed 读路径不查 DB。
- **写路径同步**：`publishPostCAS` 成功时 ZADD baseScore、GEOADD（有坐标）、写 post/profile snapshot；like/unlike 维护 liked SET 与 ZSET 分数更新。
- **API**：
  - **新增** `POST /ucg/app/api/v2/posts`（可选 lat/lng）— **计入** usage 统计。
  - **兼容修改** v1：`PUT /ucg/app/api/posts/{id}` 可选 lat/lng；`GET /ucg/app/api/feed/recommend` 可选 lat/lng + cursor（无 total，含 nextCursor、distanceMeters）；`GET /ucg/app/api/posts/{id}` 可选 lat/lng query 返回 distanceMeters。
- **迁移**：backfill 已发布帖 → ZSET + GEO + snapshots + liked sets；runbook 步骤。
- **Flutter**（`flutter_ai_talk`）：geolocator、v2 发帖、cursor Feed、distance 展示 — 见 tasks 第 5 阶段。

## Capabilities

### New Capabilities

- `ucg-feed-redis-store`：Feed 专用 Redis 键（GEO/ZSET/snapshot/session/liked SET）、TTL 与 cachekit builder 登记、读写语义。

### Modified Capabilities

- `ucg-recommend-feed`：排序源从 MySQL JOIN 改为 Redis 复合分；半径扩展、无 geo 帖合并、cursor/session 防重、hasMore 语义。
- `ucg-app-http-api`：Feed cursor 契约、distanceMeters 字段、坐标可选入参、v2 创建帖 endpoint。
- `ucg-recommend-mq`：MQ/reconciler 重算结果写入 Redis ZSET（及 GEO 坐标变更），不再 UPSERT `ucg_post_recommend`。

## Impact

- **go_ai_talk**
  - `internal/platform/cachekit/keys_ucg.go` — 新增 Feed 相关键 builder
  - `internal/platform/cachekit/` — GEO/ZSET 批量读封装（若尚无）
  - `internal/services/ucg/` — feed、post、recommend、like、snapshot、session
  - `api/v1/ucg_app_http.go`、`api/v2/`（或等价）— 路由与 DTO
  - `internal/controller/ucg_app_api.go` — handler
  - `manifest/config/config.ucg-service.yaml` — wDist、distDecayKm、radiusSteps、candidateBatchSize、session TTL
  - `internal/services/gatewayapp/usagestats/maintenance_skip.go` — **不**排除 v2 posts（计入统计）
  - backfill 脚本或 runbook 步骤
- **flutter_ai_talk**（`d:\work\flutter_ai_talk`）— geolocator、API 客户端、Feed/详情 UI 距离展示
- **Redis 读缓存**：负责人已确认；实现 MUST 经 cachekit + WithObserver
- **无新增测试文件**；**无**新增 background ticker（backfill 为一次性/运维脚本，非常驻 reconciler 以外的新循环任务）
- **破坏性**：Feed recommend 响应分页字段由 `{total,page,pageSize}` 改为 cursor 模式（`nextCursor`、`hasMore`）；旧客户端若依赖 `total` 需升级（v1 path 兼容保留 page 参数 deprecated 或忽略）
