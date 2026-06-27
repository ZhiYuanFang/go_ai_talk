## MODIFIED Requirements

### Requirement: sim admin API SHALL expose per-task AI model catalog

`GET /sim/admin/api/status` 与 `GET /sim/admin/api/runtime` 响应 MUST 含 `taskAiModels`（object）：键为调度任务名（至少含 `register`、`comment`、`post_image`、`post_video_submit`、`chat_scan`、`follow`），值为数组。**MUST NOT** 含 `video_poll` 键。

数组元素 MUST 含 `laneKey`、`usage`（可选）、`provider`、`model`。`post_video_submit` MUST 含 simText 与 simVideoGen 相关条目。

#### Scenario: Status returns AI for post video task

- **WHEN** 已鉴权 GET `/sim/admin/api/status`
- **THEN** `taskAiModels.post_video_submit` MUST 含 simText 与 simVideoGen 信息且 MUST NOT 存在 `video_poll` 键

### Requirement: sim admin runtime intervals SHALL include T4 inline poll parameters

`GET /sim/admin/api/config` 与 `GET /sim/admin/api/runtime` 的 `intervals` MUST 含：

- `postVideoPollInterval` — T4 提交后 async-result 轮询间隔（字符串 duration）
- `postVideoPollMaxWait` — T4 视频发布最大等待（字符串 duration）

MUST NOT 再返回 `videoPollIdle`、`videoPollActive`。`taskSwitches` MUST NOT 含 `videoPoll`。

`PUT /sim/admin/api/config` MUST 接受并持久化上述字段。

#### Scenario: Config get includes poll intervals

- **WHEN** 管理员 GET `/sim/admin/api/config`
- **THEN** `intervals.postVideoPollInterval` 与 `intervals.postVideoPollMaxWait` MUST 为可读 duration 字符串

### Requirement: sim admin UI SHALL reflect T4 inline video flow

`sim-admin.html` MUST：

- 移除 P1 任务开关、P1 周期输入、P1 手动执行行及 runtime 只读区中的 P1 项
- 在 T4 相关区域提供可编辑 `postVideoPollInterval`、`postVideoPollMaxWait`（保存走 `PUT /sim/admin/api/config`）
- T4 手动「执行」按钮：自点击至 `sim_task_run.post_video_submit.lastRunAt` 更新前 MUST 显示「执行中…」且 disabled；**MUST NOT** 依赖新增 status 字段
- status 轮询对 `post_video_submit` MUST 足够长以覆盖 `postVideoPollMaxWait`（避免过早恢复按钮）

#### Scenario: T4 button stays busy until task run updates

- **WHEN** 管理员手动执行 T4 且视频轮询需 5 分钟完成
- **THEN** 5 分钟内按钮 MUST 保持「执行中…」直至 status 中 lastRunAt 变化

#### Scenario: P1 removed from admin form

- **WHEN** 管理员打开 sim-admin 配置区
- **THEN** MUST NOT 显示 P1 开关或 P1 idle/active 输入

## REMOVED Requirements

### Requirement: sim admin manual run for video_poll

**Reason**: P1 删除；轮询并入 T4。

**Migration**: `POST /sim/admin/api/tasks/video_poll/run` 不再可用；运维通过 T4 手动执行观察整条流水线。
