## MODIFIED Requirements

### Requirement: Recommend feed SHALL use mixed ranking algorithm

Recommend feed SHALL rank published posts using mixed score combining new-post weight and engagement decay (likes/comments age decay). Scores MUST be persisted in `ucg_post_recommend` (or equivalent materialized store). The platform MUST update scores via **MQ event-driven single-post recompute** and **hot-zone paginated incremental reconciler** only. The platform **MUST NOT** periodically refresh scores by loading all `status=2` posts without pagination (全表刷新).

Hot zone MUST be defined as posts with `published_at >= round_hot_cutoff` where `round_hot_cutoff` is fixed at the start of each hot reconciler round (MUST NOT be recomputed with `NOW()` during pagination within the same round). Cold zone posts MUST rely on MQ events only for score updates; cold-zone paginated reconciler MUST NOT run. MQ-driven scores MAY be briefly stale; hot-zone reconciler MUST provide eventual consistency including time decay.

#### Scenario: 新帖权重

- **WHEN** 两条帖子互动相同但新帖发布时间更近
- **THEN** 推荐排序 SHALL 倾向较新帖子（在衰减窗口内）

#### Scenario: 仅 published 入推荐

- **WHEN** 计算推荐候选集
- **THEN** 算法 SHALL 仅包含 `status=2` 帖子

#### Scenario: 禁止全表刷新

- **WHEN** 后台任务更新 `ucg_post_recommend`
- **THEN** 系统 MUST NOT 执行 `SELECT` 全部 published 帖且无 `LIMIT` 的批量重算

#### Scenario: 发帖过审后入推荐排序

- **WHEN** 帖子审核通过变为 published 且 `ucg.post.published` 已处理
- **THEN** `ucg_post_recommend` MUST 存在该 `postId` 的 score 行供 Feed JOIN 排序
