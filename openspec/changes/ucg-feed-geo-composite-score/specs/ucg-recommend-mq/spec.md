## MODIFIED Requirements

### Requirement: 推荐分 AMQP consumer MUST 驻留 ucg-service 且 SHALL 单帖更新

`ucg-service` MUST 经 AMQP push（`autoAck=false`）消费 `ucg.recommend.score.q`，按 routing key 分发：

- `published` / `liked` / `unliked` / `comment.published` / `comment.removed` → `RecomputeRecommendScore(postId)` **ZADD** Redis ZSET `ucg:recommend:score`（member=`postId`，score=baseScore）；**MUST NOT** UPSERT MySQL `ucg_post_recommend`
- `unpublished` → **ZREM** `ucg:recommend:score`、**ZREM** `ucg:feed:geo`、**DEL** post snapshot

MUST NOT 在 consumer 内对全部 published 帖做无 LIMIT 全表扫描。

#### Scenario: published 写入 Redis ZSET

- **WHEN** consumer 收到 `ucg.post.published` 且 `postId` 合法
- **THEN** 系统 MUST ZADD baseScore 至 `ucg:recommend:score` 且 MUST NOT INSERT `ucg_post_recommend`

#### Scenario: unpublished 清理 Redis 推荐态

- **WHEN** consumer 收到 `ucg.post.unpublished` 且 `postId` 合法
- **THEN** 系统 MUST ZREM score 与 geo 且 DEL snapshot；`RowsAffected=0` 类语义 MUST 视为成功且 MUST Ack

### Requirement: 热区 reconciler MUST 分页增量且轮首固定 hotCutoff

系统 MUST 使用 MySQL 表 `ucg_recommend_hot_scan_state` 驱动热区分页 reconciler。每 tick MUST 使用 `published_at >= round_hot_cutoff AND id > last_post_id ORDER BY id LIMIT pageSize`；对扫到的帖 MUST 执行 `RecomputeRecommendScore` 并 **ZADD** Redis ZSET；MUST NOT 一次加载全部 published 帖；MUST NOT UPSERT `ucg_post_recommend`。

#### Scenario: 热区 reconciler 写 Redis

- **WHEN** 热区 reconciler 扫描到 published 热区帖
- **THEN** 系统 MUST ZADD 该帖 baseScore 至 `ucg:recommend:score` 且 MUST NOT 写 MySQL recommend 表

#### Scenario: 无互动帖仍更新 score

- **WHEN** 热区 reconciler 扫描到 `postId` 且该帖在上一扫描周期内 like/comment 计数未变
- **THEN** 系统 MUST 仍执行 `RecomputeRecommendScore(postId)` 并 ZADD 以反映时间衰减

### Requirement: publishPostCAS 成功路径 SHALL 同步 Redis 推荐索引

帖子 Green 审核 CAS 成功变为 published 时，写路径 MUST 同步：

1. ZADD `ucg:recommend:score`（initial baseScore）
2. 若帖含 lat/lng → GEOADD `ucg:feed:geo`
3. SET post snapshot 与 author profile snapshot

MUST NOT 写入 `ucg_post_recommend`。

#### Scenario: 过审发帖同步 GEO 与 snapshot

- **WHEN** 帖 publish 成功且含 lat/lng
- **THEN** 系统 MUST GEOADD、`ZADD` 且写入 post/profile snapshot 于同一写路径事务语义内（best-effort 顺序：MySQL commit 后 Redis）
