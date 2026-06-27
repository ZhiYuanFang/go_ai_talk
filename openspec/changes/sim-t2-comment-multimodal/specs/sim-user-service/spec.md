## MODIFIED Requirements

### Requirement: T2 comment task SHALL post AI comments every 6 hours

每 6 小时（±10% jitter，可由 `SIM_INTERVAL_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`**（`mode=random`）获取单条已发布帖（MUST NOT 调用 `GET /ucg/app/api/feed/recommend`）。取得帖子后 MUST 经 **单次** `LaneSimVision` 调用生成评论并 `POST` 发表评论。

- 当 sample 返回非空 **`coverCdnUrl`** 时，MUST 以 OpenAI 兼容多模态 user message（`image_url` content part 指向 `coverCdnUrl`，`text` content part 含评论目标 prompt 与帖子正文）调用 `aimodel.Invoke`；MUST NOT 拆成两次 LLM 调用。
- 当无 `coverCdnUrl`（纯文字帖或 URL 不可用）时，MAY 以纯文本 user message 调用同一 `LaneSimVision` lane。
- 当有媒体类型但 `coverCdnUrl` 为空时，MUST 降级为纯文本调用并 SHOULD 记录 warning 日志；MAY 仍发表评论。

Green 审核 MUST 走正常 UCG 路径。

#### Scenario: Multimodal comment on image post

- **WHEN** sample 返回 `coverCdnUrl` 非空且 LLM 可用
- **THEN** T2 MUST 在单次 `LaneSimVision` 请求中包含 `image_url` 与含帖子正文的 text part，且 MUST 成功时 POST 一条评论

#### Scenario: Text-only post comment

- **WHEN** sample 无 `coverCdnUrl` 且 `mediaType=0`
- **THEN** T2 MUST 仍经 `LaneSimVision` 文本消息生成并发表评论

#### Scenario: Single LLM invocation

- **WHEN** 代码评审 `RunCommentTask` 且帖子含 `coverCdnUrl`
- **THEN** MUST NOT 对同一帖子连续两次 `aimodel.Invoke` 完成识图与写评

#### Scenario: No recommend feed call

- **WHEN** T2 任务执行
- **THEN** MUST NOT 请求 `/ucg/app/api/feed/recommend`

#### Scenario: CDN URL missing fallback

- **WHEN** `mediaType` 表示有媒体但 `coverCdnUrl` 为空
- **THEN** T2 MUST 降级纯文本 `Invoke` 且 MUST NOT 因缺少 URL  alone 而跳过评论（除非 LLM 或 POST 失败）
