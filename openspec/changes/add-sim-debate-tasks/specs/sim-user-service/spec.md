## MODIFIED Requirements

### Requirement: T2 comment task SHALL post AI comments every 6 hours

每 6 小时（±10% jitter，可由 `SIM_INTERVAL_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，且 **`excludeMediaTypes` MUST 含 `2`（视频）**，**`excludeDebate` MUST 为 `true`**，以获取单条非视频、**非辩论**已发布帖（MUST NOT 调用 `GET /ucg/app/api/feed/recommend`）。取得帖子后 MUST 经单次 `LaneSimVision` 生成评论并 `POST` 发表。T2 MUST NOT 对 `mediaType=2` 的帖子发表评论；T2 MUST NOT 对辩论帖发表评论；若 sample 仍返回视频帖或辩论帖（例如 ucg 未升级），MUST skip 并记任务失败或 warning，且 MUST NOT POST 评论。

允许评论的帖子类型 MUST 为 `mediaType=0`（纯文字）或 `mediaType=1`（图文），且 MUST NOT 为辩论帖。Green 审核 MUST 走正常 UCG 路径。

#### Scenario: Sample request excludes video and debate

- **WHEN** 代码评审 `RunCommentTask` 的 sample 请求体
- **THEN** MUST 含 `"excludeMediaTypes": [2]` 且 MUST 含 `"excludeDebate": true`

#### Scenario: No comment on video post

- **WHEN** sample 返回帖的 `mediaType` 为 `2`
- **THEN** T2 MUST NOT 调用 LLM 或 POST 评论

#### Scenario: No comment on debate post

- **WHEN** sample 返回帖含非空 `debateLeft` 与 `debateRight`（或 ucg 判定为辩论帖）
- **THEN** T2 MUST NOT 调用 LLM 或 POST 评论

#### Scenario: Text and image posts still commented

- **WHEN** sample 返回非辩论且 `mediaType=0` 或 `1` 且 LLM 可用
- **THEN** T2 MUST 按现有多模态/纯文本规则生成并发表评论

#### Scenario: No eligible post after exclude

- **WHEN** sample 因 exclude 返回空 `list`
- **THEN** T2 MUST 记录失败（如「无已发布帖」）且 MUST NOT 发表评论

#### Scenario: No recommend feed call

- **WHEN** T2 任务执行
- **THEN** MUST NOT 请求 `/ucg/app/api/feed/recommend`

## ADDED Requirements

### Requirement: T7 post_debate task SHALL publish debate posts every 12 hours

每 12 小时（±10% jitter，可由 `SIM_INTERVAL_POST_DEBATE` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，经 `LaneSimText`（prompt `post_debate_text`）生成 JSON：`content`（话题正文）、`debateLeft`、`debateRight`（各 ≤5 字，与 UCG `validateDebateLabel` 一致），并 `POST /ucg/app/api/posts`，body MUST 含 `content`、`debateLeft`、`debateRight`、`mediaType=0`、`submit=true`，且 MUST NOT 含 `media`。标签校验失败或 LLM 失败 MUST 记 `post_debate` task 失败。

#### Scenario: Debate post submitted for audit

- **WHEN** LLM 输出合法且 POST 成功
- **THEN** 帖子 MUST 为 `type=debate`（或等价双填标签）且 MUST 进入 `pending_audit`

#### Scenario: Invalid debate labels fail task

- **WHEN** LLM 输出任一方立场标签超过 5 字或为空
- **THEN** MUST 记 `post_debate` 失败且 MUST NOT POST

#### Scenario: Default interval twelve hours

- **WHEN** 未设置 `SIM_INTERVAL_POST_DEBATE` 且 DB 无覆盖
- **THEN** T7 名义周期 MUST 为 12h（含 jitter）

### Requirement: T8 debate_comment task SHALL vote and post argument every 1 hour

每 1 小时（±10% jitter，可由 `SIM_INTERVAL_DEBATE_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，**`onlyDebate` MUST 为 `true`**，取得单条辩论帖（含 `debateLeft`、`debateRight`）。若 `authorWxId` 等于当前 sim 用户 wxId，MUST 记失败（如「不可评自己的帖」）且 MUST NOT 投票或评论。否则 MUST 经 `LaneSimText`（prompt `debate_comment`）生成 JSON：`side`（`left` 或 `right`）、`argument`（utf8 长度 ≤10），MUST 先 `POST /ucg/app/api/posts/{id}/vote` 再 `POST /ucg/app/api/posts/{id}/comments`。argument 超长 MUST 截断至 10 字或记失败（实现 MUST 在 POST 前保证 ≤10 字）。无 eligible 辩论帖时 MUST 记失败（如「无已发布辩论帖」）。

#### Scenario: Vote before comment

- **WHEN** T8 tick 且 sample 命中他人辩论帖
- **THEN** MUST 先 POST vote 成功再 POST comment

#### Scenario: Skip own debate post

- **WHEN** sample 返回帖 `authorWxId` 等于当前 sim wxId
- **THEN** MUST NOT POST vote 或 comment

#### Scenario: Argument length bound

- **WHEN** LLM 输出 argument 超过 10 字
- **THEN** 实现 MUST 截断至 10 字或记失败，且 POST 的 comment MUST NOT 超过 10 字

#### Scenario: Default interval one hour

- **WHEN** 未设置 `SIM_INTERVAL_DEBATE_COMMENT` 且 DB 无覆盖
- **THEN** T8 名义周期 MUST 为 1h（含 jitter）

#### Scenario: No debate posts available

- **WHEN** `onlyDebate=true` 且 sample 返回空 `list`
- **THEN** MUST 记 `debate_comment` 失败且 MUST NOT 调用 LLM

### Requirement: sim task intervals SHALL include T7 and T8 defaults

`SIM_INTERVAL_POST_DEBATE` 默认 MUST 为 `12h`；`SIM_INTERVAL_DEBATE_COMMENT` 默认 MUST 为 `1h`。对应 env 开关 `SIM_TASK_POST_DEBATE_ENABLED`、`SIM_TASK_DEBATE_COMMENT_ENABLED` 迁移期默认 SHOULD 为 true（与 DB runtime_json 合并后生效）。

#### Scenario: T7 T8 env defaults

- **WHEN** 未设置 T7/T8 interval env
- **THEN** LoadRuntimeFlags MUST 回退 12h 与 1h
