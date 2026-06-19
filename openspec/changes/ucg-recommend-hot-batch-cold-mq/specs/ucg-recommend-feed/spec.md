## MODIFIED Requirements

### Requirement: Recommend feed SHALL rank published posts using mixed score combining new-post weight and engagement decay

Recommend feed SHALL rank published posts using mixed score combining new-post weight and engagement decay (likes/comments age decay). Implementation MAY persist scores in `ucg_post_recommend` refreshed by热区 reconciler（默认 1h tick）与冷区 MQ 互动重算。

对 **尚无** `ucg_post_recommend` 行的 published 帖，Feed MUST 在排序中优先于已有 score 的帖（置顶区），置顶区内 MUST 按 `published_at` 降序；已有 score 的帖 MUST 按 `score` 降序（同分则 `published_at`、`id` 降序）。该置顶 MUST 持续直至 reconciler 或冷区 MQ 路径首次写入 recommend 行。

#### Scenario: Unscored new post appears before scored posts

- **WHEN** 用户请求 `GET /ucg/app/api/feed/recommend` 且存在 published 帖尚无 `ucg_post_recommend` 行
- **THEN** 响应 `list` 中该类帖 MUST 排在已有 score 帖之前（同置顶区内按 `published_at` 新者优先）

#### Scenario: After reconciler scores post leaves pin tier

- **WHEN** reconciler 已为帖 UPSERT `ucg_post_recommend.score`
- **THEN** 该帖 MUST 按 score 与全站其他已算分帖一起排序，不再仅因「无 recommend 行」置顶
