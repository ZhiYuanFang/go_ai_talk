## MODIFIED Requirements

### Requirement: T6 follow task SHALL have sim follow real post authors every 7 hours

每 7 小时（±10% jitter）MUST 随机选取一个 sim 用户 A，经 gateway 登录；调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，body **`excludeAuthorWxIds`** MUST 为 device internal 拉取的全量 sim wxId 列表；从返回帖取得 **`authorWxId`** 作为关注目标 B，并对 B 调用 `POST /ucg/app/api/follow/{wxId}`。`authorWxId` MUST NOT 等于 A 的 wxId（不等则重试抽样，有界次数后仍失败则记 task 失败）。MUST NOT 执行 sim→sim 关注；MUST NOT 使用 `pickTwoDistinctSimWx` 或两次 `sim/wx/random` 互关。已关注 MUST 幂等跳过。无 eligible 真人 author 时 MUST 记 task 失败（如「无真人作者」），MUST NOT 假 success。

#### Scenario: Sim follows real author

- **WHEN** 存在 published 帖且 author 非 sim，A 未关注 B
- **THEN** `POST /ucg/app/api/follow/{B}` MUST 成功或幂等

#### Scenario: No sim to sim follow

- **WHEN** T6 tick 执行
- **THEN** MUST NOT 选取两个 sim 用户互相关注

#### Scenario: No real author available

- **WHEN** 所有 published 帖作者均在 `excludeAuthorWxIds` 中
- **THEN** MUST 记 follow task 失败且 MUST NOT POST follow
