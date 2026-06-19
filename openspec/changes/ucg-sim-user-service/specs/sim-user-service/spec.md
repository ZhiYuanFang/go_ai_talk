## ADDED Requirements

### Requirement: sim-user-service SHALL run as an independent process with a master enable switch

系统 MUST 提供独立进程 `sim-user-service`（`cmd/sim-user-service`）。环境变量 `SIM_USER_SERVICE_ENABLED` 为 `false` 时，进程 MAY 启动健康检查但 MUST NOT 启动任何周期 ticker 或视频轮询。为 `true` 时 MUST 注册全部已声明背景任务。

#### Scenario: Service disabled

- **WHEN** `SIM_USER_SERVICE_ENABLED=false`
- **THEN** 进程 MUST NOT 执行 T1–T6 与 P1 任务

#### Scenario: Service enabled

- **WHEN** `SIM_USER_SERVICE_ENABLED=true` 且依赖服务可达
- **THEN** 进程 MUST 按各任务独立开关启动对应 ticker

### Requirement: T1 register task SHALL create simulated users every 24 hours

每 24 小时（±10% jitter）MUST 执行注册任务（`SIM_TASK_REGISTER_ENABLED`）。当当前 sim 用户数 `< maxSimUsers` 时，MUST 分配下一序号账号 `ptest{N}`（N 递增）、密码默认 `123456`，调用 device internal sim 注册，经 `simText` 生成昵称、`simImageGen` 生成头像，完成 UCG profile 与 media 上传，并更新 nickname/avatarKey。当已达 `maxSimUsers` 时 MUST 跳过本次执行。

#### Scenario: Register under cap

- **WHEN** sim 用户数为 99 且 `maxSimUsers=100`
- **THEN** 系统 MUST 注册 `ptest100`（或下一可用序号）并标记 `is_simulated=1`

#### Scenario: Skip at cap

- **WHEN** sim 用户数已达 `maxSimUsers`
- **THEN** T1 MUST NOT 调用注册接口

### Requirement: T2 comment task SHALL post AI comments every 6 hours

每 6 小时（±10% jitter）MUST 随机选取一个 sim 用户，经 gateway 登录，拉取推荐 Feed，随机选一条已发布动态，经 `simVision` 结合正文与图片生成评论，并 `POST` 发表评论。Green 审核 MUST 走正常 UCG 路径。

#### Scenario: Successful comment flow

- **WHEN** Feed 存在已发布帖子且 LLM 闸门可用
- **THEN** 系统 MUST 以 sim 用户身份发表一条评论

### Requirement: T3 image post task SHALL publish image posts every 3.5 hours

每 3.5 小时（±10% jitter）MUST 随机 sim 用户经 `simText` 生成母婴文案、`simImageGen` 生成配图，完成 OSS media 链路后以 `submit=true` 创建图文动态。

#### Scenario: Image post submitted

- **WHEN** 生图与上传成功
- **THEN** 帖子 MUST 进入 `pending_audit` 并触发正常审核 MQ

### Requirement: T4 and P1 SHALL submit and poll video posts without retry

每 6.5 小时（±10% jitter）T4 MUST 随机 sim 用户生成文案并 `SubmitVideoGeneration`，将 `task_id` 写入 `sim_video_job`（同一 sim 同时最多 1 个 pending/processing job）。P1 每 1 分钟 MUST 轮询智谱 `/paas/v4/videos/generations`；成功 MUST 下载视频、走 UCG media 并 `submit=true` 发帖；失败 MUST 标记 job 为 skipped 且 MUST NOT 重试。

#### Scenario: Video success

- **WHEN** 轮询返回成功且视频可下载
- **THEN** 系统 MUST 创建视频动态并关闭 job

#### Scenario: Video failure

- **WHEN** 轮询返回失败状态
- **THEN** job MUST 变为 skipped 且 MUST NOT 再次提交同一 task

### Requirement: T5 and E1 SHALL handle chat with real users

每 1 小时 T5 MUST 随机 sim 用户检查会话列表。对未读且 peer 非 sim 的会话，MAY spawn **E1**：30 分钟内每 1 分钟检查未读，有则拉历史、经 `simText` 生成回复并调用 UCG internal chat/send；30 分钟 MUST 硬停止且不续期。同一 `(simWxId, peerWxId)` 同时最多 1 个 E1 窗口。E1 MUST NOT 阻塞 T2–T6。

#### Scenario: Ephemeral window expires

- **WHEN** E1 已运行 30 分钟
- **THEN** 系统 MUST 停止该窗口的后续回复即使仍有未读

#### Scenario: Real peer only

- **WHEN** 未读会话 peer 的 `is_simulated=1`
- **THEN** T5 MUST NOT 为该会话 spawn E1

### Requirement: T6 follow task SHALL follow sim to sim every 7 hours

每 7 小时（±10% jitter）MUST 随机选取两个不同 sim 用户 A、B，A 经 gateway 登录并对 B 调用关注 API；已关注 MUST 幂等跳过。

#### Scenario: Sim follows sim

- **WHEN** A 与 B 均为 sim 且 A 未关注 B
- **THEN** `POST /ucg/app/api/follow/{wxId}` MUST 成功或幂等

### Requirement: sim-user-service MUST use HTTP contracts only

sim-user-service MUST NOT import 或查询 device/ucg 域 DAO。跨服务数据 MUST 经 gateway App API、device internal API、ucg internal API 与 `aimodel` 完成。

#### Scenario: No cross-domain DAO

- **WHEN** 代码评审检索 `internal/dao` import
- **THEN** sim-user-service 包路径下 MUST NOT 出现 device/ucg 业务表 DAO

### Requirement: AI queue full SHALL abort current task tick

当 `aimodel.Acquire` 返回队列满或调用超时时，当前任务 tick MUST 提前结束并记录日志，MUST NOT 阻塞其他任务 goroutine。

#### Scenario: Queue full

- **WHEN** `ErrQueueFull` 在评论任务中发生
- **THEN** 该次 T2 执行 MUST 结束且不重试至下一周期
