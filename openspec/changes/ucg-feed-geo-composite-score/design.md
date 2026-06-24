## Context

v2.0.5 基线中 UCG 推荐 Feed 经 MySQL `ucg_post_recommend` LEFT JOIN 排序；MQ consumer 与热区 reconciler 写入 MySQL score 行。`feed.go` 已有 Redis 复合排序 TODO 注释。产品选定**方案 2**：距离不是「附近帖筛选」，而是**加性复合分**的一部分；同时要在 UI 展示精确距离（米）。

约束：Redis 访问 MUST 经 `cachekit`；键 MUST 登记于 `keys_ucg.go`；负责人已确认 Feed 读路径 Redis 缓存；仅 `POST /v2/posts` 新计入 usage 统计。

## Goals / Non-Goals

**Goals:**

- 请求时计算 `finalScore = baseScore + distanceTerm`（双方有坐标时）。
- `baseScore` 权威存储于 Redis ZSET；停止 MySQL `ucg_post_recommend` 写入。
- Feed 读路径：GEO 半径候选 + ZSET 分数 + 无 geo 帖合并 + cursor + session 防重。
- 帖子/作者 snapshot、liked SET 支撑 Feed 零 DB 热路径（miss 可回源 MySQL）。
- API 返回 `distanceMeters` 字符串；v2 创建帖；v1 兼容扩展 lat/lng。
- Backfill 历史 published 帖至 Redis。
- Flutter 端 geolocator + cursor Feed + 距离 UI。

**Non-Goals:**

- 不改关注 Feed（仍 page/total MySQL 路径，本 change 不碰）。
- 不删除 `ucg_post_recommend` 表（仅停写）。
- 不在 API 返回距离文案标签（客户端 `formatDistance` 负责 m/km）。
- 不新增常驻 background ticker（backfill 为运维脚本/一次性 job）。
- 不优化评论列表等无关读路径。

## Decisions

### D1：加性复合分公式（采用）

**决策**：

```
distanceTerm = wDist * exp(-distKm / distDecayKm)   // distKm 来自 GEODIST/haversine
finalScore   = baseScore + distanceTerm              // 仅当 viewer 有 lat/lng 且 post snapshot 含 lat/lng
             = baseScore                             // 否则
```

- `baseScore`：现有推荐算法产出（新帖权重 + 互动衰减），存于 `ucg:recommend:score` ZSET member=`postId`。
- 默认配置：`wDist=0.5`，`distDecayKm=10`（`config.ucg-service.yaml` 可覆盖）。

**理由**：距离增强而非替代互动/时效排序；无坐标优雅降级。

### D2：Redis 数据结构（采用）

| 键 | 类型 | 用途 |
|----|------|------|
| `ucg:feed:geo` | GEO | 有坐标的 published 帖；member=`postId`，coord=(lng,lat) Redis 顺序 |
| `ucg:recommend:score` | ZSET | baseScore；member=`postId` |
| `ucg:post:snapshot:{postId}` | STRING JSON | Feed 展示字段 + server-only lat/lng |
| `ucg:profile:snapshot:{wxId}` | STRING JSON | 作者公开 profile |
| `ucg:user:{wxId}:liked-posts` | SET | 用户点赞 postId |
| `ucg:feed:session:{sessionId}` | SET | 已下发 postId，TTL 30min |

snapshot JSON 字段：

- post：`id, content, media(CDN URL), authorWxId, likeCount, ipLocation, publishedAt, mediaType` + 内部 `lat,lng`（不下发客户端）
- profile：`wxId, nickname, bio, avatarUrl, avatarThumbnailUrl`

TTL：snapshot 建议 7d（可配置）；session 30min；GEO/ZSET/liked SET 随写路径维护，无自动过期（下架 ZREM/DELETE）。

### D3：Feed 候选与半径扩展（采用）

1. 解析 cursor（见 D4）；确定 frozen `lat,lng,radiusKm,geoOffset,sessionId`。
2. 从当前 `radiusKm` 用 `GEOSEARCH/GEORADIUS ... WITHDIST` 拉取 batch（`candidateBatchSize=200`），跳过 session 中已见 postId。
3. 若 geo 候选不足 `pageSize`，按阶梯扩展半径：`50 → 100 → 200 → 500 → unlimited`（`radiusKm=0` 表示不限，改从 ZSET 全量扫）。
4. 并行合并**无 geo 帖**：ZSET 中不在 GEO 索引的 postId（历史帖无坐标），同样按 baseScore 参与排序。
5. 对候选 batch 批量 ZSCORE 取 baseScore，计算 finalScore，排序 `finalScore DESC, postId DESC`，取 top `pageSize`。
6. `hasMore = len(returned) == pageSize`（无 total）；将返回 postId SADD 至 session SET。

### D4：Cursor 与 Feed session 防重（采用，CRITICAL）

**cursor**（opaque base64 JSON 或等价编码）冻结：

- `sessionId`（UUID，首屏无 cursor 时生成）
- `lat, lng`（首屏请求坐标，后续页 MUST 忽略新坐标）
- `radiusKm, geoOffset`（当前半径与 GEO 扫描偏移）
- `lastFinalScore, lastPostId`（tie-break 游标，可选）

**session**：`ucg:feed:session:{sessionId}` SET 存已返回 postId；TTL 30min；每次分页 SADD。

**刷新 Feed**（下拉刷新）：客户端不传 cursor → 新 sessionId → 重新 GEO/ZSET 全排序。

**排序稳定性**：`finalScore DESC, postId DESC`。

### D5：写路径同步（采用）

`publishPostCAS` Green 成功 → published：

1. `RecomputeRecommendScore` 逻辑产出 baseScore → `ZADD ucg:recommend:score`
2. 若帖含 lat/lng → `GEOADD ucg:feed:geo`；否则 `ZREM` geo（若曾存在）
3. 写 `ucg:post:snapshot:{postId}`；刷新 `ucg:profile:snapshot:{authorWxId}`
4. **不再** UPSERT `ucg_post_recommend`

`DeletePost`/unpublished → `ZREM` score + geo + DEL snapshot；从活跃 Feed 消失。

`UpdatePost` PUT 可更新 lat/lng → 同步 GEO + snapshot。

### D6：likedByMe（采用）

- like → `SADD ucg:user:{wxId}:liked-posts {postId}`
- unlike → `SREM`
- Feed 读：对当前页 postIds 批量 `SISMEMBER`（pipeline）；**不查** `ucg_post_like` DB

Backfill 从 `ucg_post_like` 重建各用户 SET。

### D7：MQ / reconciler 写入目标（采用）

- 热区 reconciler、冷区 MQ `RecomputeRecommendScore` → **ZADD** Redis ZSET（替换 MySQL UPSERT）
- `unpublished` → ZREM score + geo + DEL snapshot
- throttle 键沿用 `UCGRecommendThrottleKey`

### D8：API 契约（采用）

| 方法 | 路径 | 变更 |
|------|------|------|
| POST | `/ucg/app/api/v2/posts` | **新增**；body 含 v1 字段 + 可选 `lat`,`lng`；**计入 usage** |
| PUT | `/ucg/app/api/posts/{id}` | 可选 `lat`,`lng` |
| GET | `/ucg/app/api/feed/recommend` | 可选 query `lat`,`lng`,`cursor`,`pageSize`；响应 `{list, nextCursor?, hasMore}` **无 total**；item 含 `distanceMeters?` |
| GET | `/ucg/app/api/posts/{id}` | 可选 query `lat`,`lng`；响应含 `distanceMeters?` |

`distanceMeters`：字符串，单位为米（如 `"1234"`、`"850"`）；viewer 或 post 无坐标时省略。

`page` 参数：**deprecated**，有 `cursor` 时忽略；无 cursor 时视为首页。

### D9：snapshot miss 回源（采用）

Feed 组装时若 snapshot 缺失 → 单帖回源 MySQL + device profile API → 回填 snapshot（best-effort）。避免 Feed 空洞。

### D10：Backfill（采用）

运维脚本/一次性 job（非常驻 ticker）：

1. 扫描 `status=2` published 帖分页
2. 对每帖：计算 baseScore → ZADD；有 lat/lng → GEOADD；写 post snapshot；写 author profile snapshot
3. 扫描 `ucg_post_like` → 重建 `ucg:user:{wxId}:liked-posts`

Runbook 记录执行顺序与验收（抽样 Feed 对比）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| Redis 与 MySQL 短暂不一致 | 写路径同步 + reconciler/MQ 收敛 ZSET；snapshot miss 回源 |
| session TTL 内刷新仍见部分旧帖 | 刷新不传 cursor 即新 session |
| 无 geo 历史帖排序偏弱 | 仍按 baseScore 参与全站排序，仅无 distanceTerm |
| cursor 坐标冻结与用户移动 | 产品接受；刷新重新定位 |
| Feed 响应无 total | 客户端改 hasMore；旧版需升级 |
| backfill 大数据量 | 分页 + 限流；运维窗口执行 |

## Migration Plan

1. 部署含 Redis 键 builder + 写路径双写 ZSET（可先 ZSET 与 MySQL 并行，本 change 直接切 Redis-only 停 MySQL 写）。
2. 执行 backfill 脚本灌入 ZSET/GEO/snapshot/liked SET。
3. 部署新 Feed 读路径 + API 变更。
4. 部署 Flutter 客户端（可灰度：无 lat/lng 时行为与现网接近，无 distance 字段）。
5. 验收：Feed 有坐标时距离排序合理；cursor 无重复；likedByMe 正确；v2 posts 出现在 usage 统计。
6. 回滚：恢复旧 Feed MySQL 读路径（需保留 MySQL 写至回滚窗口，或接受 backfill 后 MySQL stale）。

## Open Questions

（无阻塞项。usage 统计：负责人已确认 **仅 v2 POST 计入**；v1 create 若仍存在则维持原统计策略不变。）
