## MODIFIED Requirements

### Requirement: T2 comment task SHALL post AI comments every 6 hours

每 6 小时（±10% jitter，可由 `SIM_INTERVAL_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，且 **`excludeMediaTypes` MUST 含 `2`（视频）**，以获取单条非视频已发布帖（MUST NOT 调用 `GET /ucg/app/api/feed/recommend`）。取得帖子后 MUST 经单次 `LaneSimVision` 生成评论并 `POST` 发表。T2 MUST NOT 对 `mediaType=2` 的帖子发表评论；若 sample 仍返回视频帖（例如 ucg 未升级），MUST skip 并记任务失败或 warning，且 MUST NOT POST 评论。

允许评论的帖子类型 MUST 为 `mediaType=0`（纯文字）或 `mediaType=1`（图文）。Green 审核 MUST 走正常 UCG 路径。

#### Scenario: Sample request excludes video

- **WHEN** 代码评审 `RunCommentTask` 的 sample 请求体
- **THEN** MUST 含 `"excludeMediaTypes": [2]`（或等价语义）

#### Scenario: No comment on video post

- **WHEN** sample 返回帖的 `mediaType` 为 `2`
- **THEN** T2 MUST NOT 调用 LLM 或 POST 评论

#### Scenario: Text and image posts still commented

- **WHEN** sample 返回 `mediaType=0` 或 `1` 且 LLM 可用
- **THEN** T2 MUST 按现有多模态/纯文本规则生成并发表评论

#### Scenario: No eligible post after exclude

- **WHEN** sample 因 exclude 返回空 `list`
- **THEN** T2 MUST 记录失败（如「无已发布帖」）且 MUST NOT 发表评论

#### Scenario: No recommend feed call

- **WHEN** T2 任务执行
- **THEN** MUST NOT 请求 `/ucg/app/api/feed/recommend`
