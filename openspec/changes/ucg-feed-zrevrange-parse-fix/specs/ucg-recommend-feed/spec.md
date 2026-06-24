## MODIFIED Requirements

### Requirement: Recommend feed SHALL use mixed ranking algorithm

Recommend feed SHALL rank published posts using **composite score** at request time:

- `baseScore`：mixed score combining new-post weight and engagement decay (likes/comments age decay)，**MUST** 持久化于 Redis ZSET `ucg:recommend:score`（member=`postId`），**MUST NOT** 写入或 JOIN MySQL `ucg_post_recommend`。
- `distanceTerm`：当 **viewer 请求含有效 lat/lng** 且 **帖子 snapshot 含 lat/lng** 时，`distanceTerm = wDist * exp(-distKm / distDecayKm)`，其中 `distKm` 为 viewer 与帖坐标距离（km）；否则 `distanceTerm = 0`。
- `finalScore = baseScore + distanceTerm`。

`baseScore` MUST 由 **MQ event-driven single-post recompute** 与 **hot-zone paginated incremental reconciler** 更新至 Redis ZSET。平台 **MUST NOT** 周期性无 LIMIT 全表刷新。

Hot zone MUST 定义为 `published_at >= round_hot_cutoff`；冷区帖 MUST 依赖 MQ 更新 ZSET score。Feed 读路径 MUST 合并：(a) GEO 半径候选（有坐标帖，`radiusKm > 0`）；(b) **`radiusKm=0`（unlimited）时 ZSET 扫描** 补全剩余帖，含不在 GEO 索引的无坐标帖。排序 MUST 为 `finalScore DESC, postId DESC`。

Feed 分页 MUST 使用 opaque `cursor` + `pageSize`（默认 20，最大 50），响应 MUST 含 `{ list, hasMore, nextCursor? }`，**MUST NOT** 含 `total`。`hasMore` MUST 为 `len(list) == pageSize` 时 true。

GEO 候选 MUST 使用默认半径 50km，不足 `pageSize` 时按阶梯扩展：`50 → 100 → 200 → 500 → unlimited` km。**unlimited（`radiusKm=0`）MUST 走 ZSET 读路径，MUST NOT 在 viewer 有坐标时仍仅用 GEO。** cursor MUST 冻结首屏 `lat/lng/radiusKm/geoOffset/sessionId` 以保证 session 内一致。

#### Scenario: 有坐标时距离参与 composite 分

- **WHEN** viewer 请求 Feed 含 lat/lng，帖 A 与帖 B baseScore 相同但 A 距 viewer 更近
- **THEN** A 的 `finalScore` MUST 高于 B 且 A SHOULD 排在 B 之前

#### Scenario: 无 viewer 坐标仅 baseScore

- **WHEN** viewer 请求 Feed 不含 lat/lng
- **THEN** 排序 MUST 等价于 `baseScore DESC, postId DESC` 且响应 item MUST NOT 含 `distanceMeters`

#### Scenario: 历史帖无坐标仍参与排序

- **WHEN** published 帖无 lat/lng 不在 GEO 索引，且 viewer 请求含有效 lat/lng
- **THEN** 该帖 MUST 在 unlimited ZSET 步进入候选并按 baseScore 排序，响应 MUST NOT 含 `distanceMeters`

#### Scenario: ZSET 有索引时 Feed 非空

- **WHEN** `ucg:recommend:score` ZCARD>0、对应 `ucg:post:snapshot:<postId>` 存在，且 viewer 请求 `GET /ucg/app/api/feed/recommend?pageSize=20`（带或不带 lat/lng）
- **THEN** 响应 `data.list` MUST 非空（除非 session cursor 已耗尽全部候选，首屏请求 MUST 非空）

#### Scenario: 仅 published 入推荐

- **WHEN** 计算推荐候选集
- **THEN** 算法 SHALL 仅包含 `status=2` 帖子

#### Scenario: 禁止全表刷新

- **WHEN** 后台任务更新 recommend score
- **THEN** 系统 MUST NOT 执行无 `LIMIT` 的全 published 帖批量重算

#### Scenario: cursor 分页无 total

- **WHEN** `GET /ucg/app/api/feed/recommend?cursor=...&pageSize=20`
- **THEN** 响应 MUST 含 `list` 与 `hasMore` 且 MUST NOT 含 `total` 或 `page`

#### Scenario: 半径不足时扩展

- **WHEN** 50km GEO 候选去重后不足 pageSize
- **THEN** 系统 MUST 扩大至下一半径阶梯继续拉取直至 unlimited ZSET 步或候选满足 pageSize
