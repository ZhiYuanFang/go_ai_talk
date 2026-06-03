## ADDED Requirements

### Requirement: Recommend feed SHALL use mixed ranking algorithm

Recommend feed SHALL rank published posts using mixed score combining new-post weight and engagement decay (likes/comments age decay). Implementation MAY persist scores in `ucg_post_recommend` or Redis ZSET refreshed by background job.

#### Scenario: 新帖权重
- **WHEN** 两条帖子互动相同但新帖发布时间更近
- **THEN** 推荐排序 SHALL 倾向较新帖子（在衰减窗口内）

#### Scenario: 仅 published 入推荐
- **WHEN** 计算推荐候选集
- **THEN** 算法 SHALL 仅包含 `status=2` 帖子
