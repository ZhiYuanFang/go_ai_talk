## MODIFIED Requirements

### Requirement: T2 comment task SHALL post AI comments every 6 hours

每 6 小时（±10% jitter，可由 `SIM_INTERVAL_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`** 且 body **`mode=random`** 获取 **单条** 已发布帖（MUST NOT 调用 `GET /ucg/app/api/feed/recommend`）。选帖随机 MUST 由 ucg random 模式完成；sim-user-service MUST NOT 先拉取 latest N 条再在内存中 `rand.Intn` 选帖。取得帖子后 MUST 经 `simVision` 结合正文与图片生成评论，并 `POST` 发表评论。Green 审核 MUST 走正常 UCG 路径。

#### Scenario: Successful comment flow

- **WHEN** sample API random 模式返回 1 条已发布帖子且 LLM 闸门可用
- **THEN** 系统 MUST 以 sim 用户身份对该帖发表一条评论

#### Scenario: No recommend feed call

- **WHEN** T2 任务执行
- **THEN** HTTP 日志/代码路径 MUST NOT 请求 `/ucg/app/api/feed/recommend`

#### Scenario: No client-side post selection random

- **WHEN** 代码评审 `RunCommentTask`
- **THEN** MUST NOT 对 sample `list` 调用 `rand.Intn` 选帖；MUST 使用 `mode=random` 请求且直接使用返回的唯一帖子

#### Scenario: Empty sample

- **WHEN** sample random 模式返回空 `list`
- **THEN** T2 MUST 记录失败（如「无已发布帖」）且 MUST NOT 发表评论
