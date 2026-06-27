## MODIFIED Requirements

### Requirement: sim tasks SHALL pick random simulated user via device random API

各调度任务（T2 评论、T3 图文、T4 视频、E1 聊天、T6 关注等）当需要随机模拟用户时，sim-user-service MUST 经 device internal **`GET /device/internal/api/sim/wx/random`** 取得单条 `{wxId, account}` 并完成 gateway 登录；MUST NOT 调用 `sim/wx/list` 拉取分页列表后在内存随机；MUST NOT 对同一选取流程重复请求 list。T6 关注 MUST 选取两个不同 wxId（在仅 1 个 sim 用户时 MUST 失败「sim 用户不足」）。

#### Scenario: Comment task uses random pick once

- **WHEN** `RunCommentTask` 需要随机 sim 用户
- **THEN** MUST 仅一次 random 调用取得 account 并登录，MUST NOT 先 list 再 random index

#### Scenario: Follow picks two distinct users

- **WHEN** `RunFollowTask` 且 simulated 用户数 ≥ 2
- **THEN** MUST 经 random（或等价 device 有界选取）得到两个不同 wxId 并完成关注

#### Scenario: No sim users

- **WHEN** random 返回无用户
- **THEN** 任务 MUST 失败且错误语义与现网「无模拟用户」一致

#### Scenario: Count still uses list total

- **WHEN** T1 注册判断 `maxSimUsers` 上限
- **THEN** MAY 继续使用 list `pageSize=1` 读取 `total`，MUST NOT 为此拉取 pageSize=200 全量
