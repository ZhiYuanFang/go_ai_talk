## ADDED Requirements

### Requirement: sim task intervals SHALL be overridable via environment variables

各背景任务周期 MUST 支持环境变量覆盖，未设置或非法值时 MUST 回退下列默认值：`SIM_INTERVAL_REGISTER=24h`、`SIM_INTERVAL_COMMENT=6h`、`SIM_INTERVAL_POST_IMAGE=3h30m`、`SIM_INTERVAL_POST_VIDEO=6h30m`、`SIM_INTERVAL_CHAT=1h`、`SIM_INTERVAL_FOLLOW=7h`。周期执行 MUST 保留 ±10% jitter。

#### Scenario: Default intervals preserved

- **WHEN** 未设置任何 `SIM_INTERVAL_*` 环境变量
- **THEN** T1–T6 名义周期 MUST 与变更前写死值一致（T3 为 3.5 小时）

#### Scenario: Custom comment interval

- **WHEN** `SIM_INTERVAL_COMMENT=12h` 且服务已启动
- **THEN** T2 两次成功执行间隔 MUST 约为 12h（含 jitter）

### Requirement: P1 video poll SHALL use adaptive idle and active intervals

P1 MUST 根据是否存在 `pending`/`processing` 的 `sim_video_job` 选择下一等待间隔：`SIM_INTERVAL_VIDEO_POLL_IDLE`（默认 `10m`）与 `SIM_INTERVAL_VIDEO_POLL_ACTIVE`（默认 `2m`）。无 pending job 时 MUST NOT 调用智谱轮询或 UCG 发帖。

#### Scenario: Idle backoff

- **WHEN** `sim_video_job` 无 pending/processing 行
- **THEN** 下一次 P1 唤醒间隔 MUST 使用 idle 间隔（默认 10m）

#### Scenario: Active polling

- **WHEN** 存在至少一条 pending/processing job
- **THEN** 下一次 P1 唤醒间隔 MUST 使用 active 间隔（默认 2m）且 MUST 执行智谱 `PollVideoGeneration`

### Requirement: sim outbound App HTTP SHALL be globally rate limited

经 gateway-app 的 `appGet`/`appPost`/`appPut` MUST 经全局限速器；默认 `SIM_UCG_RATE_LIMIT_RPS=2`（token bucket，burst 默认 4）。限速 MUST 在发起 HTTP 前阻塞等待，MUST NOT 静默丢弃请求。

#### Scenario: Burst within limit

- **WHEN** 2 秒内发起不超过 4 次 App API 调用且 RPS=2
- **THEN** 调用 MUST 成功发出（不因限速失败）

#### Scenario: Over limit waits

- **WHEN** 持续超过 RPS 限额发起调用
- **THEN** 超额调用 MUST 等待至许可可用后再发送

### Requirement: scheduler SHALL stagger first task ticks after startup

`SIM_USER_SERVICE_ENABLED=true` 启动 scheduler 时，各任务 goroutine 在首次执行前 MUST 额外等待 `0` 至 `SIM_STARTUP_STAGGER_MAX`（默认 `30m`）的均匀随机延迟，以避免多任务同时首次齐射。

#### Scenario: Staggered startup

- **WHEN** 服务启动且启用全部任务
- **THEN** 各任务首次 tick 时间 MUST NOT 全部相同（在 stagger 窗口内随机分布）

## MODIFIED Requirements

### Requirement: T2 comment task SHALL post AI comments every 6 hours

每 6 小时（±10% jitter，可由 `SIM_INTERVAL_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`** 获取已发布帖样本（MUST NOT 调用 `GET /ucg/app/api/feed/recommend`），随机选一条帖子，经 `simVision` 结合正文与图片生成评论，并 `POST` 发表评论。Green 审核 MUST 走正常 UCG 路径。

#### Scenario: Successful comment flow

- **WHEN** sample API 返回至少一条已发布帖子且 LLM 闸门可用
- **THEN** 系统 MUST 以 sim 用户身份发表一条评论

#### Scenario: No recommend feed call

- **WHEN** T2 任务执行
- **THEN** HTTP 日志/代码路径 MUST NOT 请求 `/ucg/app/api/feed/recommend`

### Requirement: T4 and P1 SHALL submit and poll video posts without retry

每 6.5 小时（±10% jitter，可由 `SIM_INTERVAL_POST_VIDEO` 覆盖）T4 MUST 随机 sim 用户生成文案并 `SubmitVideoGeneration`，将 `task_id` 写入 `sim_video_job`（同一 sim 同时最多 1 个 pending/processing job）。P1 MUST 按自适应间隔轮询智谱 `/paas/v4/videos/generations`（见 ADDED「P1 adaptive」）；成功 MUST 下载视频、走 UCG media 并 `submit=true` 发帖；失败 MUST 标记 job 为 skipped 且 MUST NOT 重试。

#### Scenario: Video success

- **WHEN** 轮询返回成功且视频可下载
- **THEN** 系统 MUST 创建视频动态并关闭 job

#### Scenario: Video failure

- **WHEN** 轮询返回失败状态
- **THEN** job MUST 变为 skipped 且 MUST NOT 再次提交同一 task

### Requirement: T5 and E1 SHALL handle chat with real users

每 1 小时（可由 `SIM_INTERVAL_CHAT` 覆盖）T5 MUST 随机 sim 用户检查会话列表。对未读且 peer 非 sim 的会话，MAY spawn **E1**：在 `SIM_EPHEMERAL_CHAT_WINDOW`（默认 **15m**）内每 `SIM_EPHEMERAL_CHAT_LOOP`（默认 **5m**）检查未读，有则拉历史、经 `simText` 生成回复并调用 UCG internal chat/send；窗口期满 MUST 硬停止且不续期。同一 `(simWxId, peerWxId)` 同时最多 1 个 E1 窗口。E1 MUST NOT 阻塞 T2–T6。

#### Scenario: Ephemeral window expires

- **WHEN** E1 已运行达到 `SIM_EPHEMERAL_CHAT_WINDOW`（默认 15m）
- **THEN** 系统 MUST 停止该窗口的后续回复即使仍有未读

#### Scenario: Real peer only

- **WHEN** 未读会话 peer 的 `is_simulated=1`
- **THEN** T5 MUST NOT 为该会话 spawn E1
