## Why

预测临近同步对每条待办按 `nextAt` 计算 `x-delay` 毫秒并投递 RabbitMQ delayed exchange。该插件延时上限为 Erlang timer 的 `2^32−1` 毫秒（约 49.7 天）；超过时插件会 **立刻投递**（nodelay），消费侧又只对账 Redis、不校验「是否已到临近窗口」，导致疫苗等 **2 年后** 的约定可能被误推。预约 `nextAt` 与 Redis pending 的 `int64`/`BIGINT` 已能表达远点，缺口在调度介质类型上限。

## What Changes

- 同步写 Redis 成功后，对单条待办：若计算得到的 `delayMs` 超过安全上限（对齐插件 `ERL_MAX_T`，实现可留余量），**跳过** `PublishDelayed`，仍视为同步成功（该条保留在 Redis）。
- 近端（delay 在上限内）行为不变：照常发延时 MQ。
- 远预约的准时叫醒依赖客户端在接近 `nextAt` 时再次 `PUT pending` 补发近端延时；服务端本变更 **不** 做分段续约、**不** 在消费路径二次 Publish。
- 增加可观测日志（跳过原因 + eventId / delayMs 量级），便于排查「为何暂无 MQ」。
- **非 BREAKING** API：路径、请求体、响应字段不变；语义上远点从「误立刻推」变为「只落 Redis、暂不调度」。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `predict-imminent-schedule`：修正「同步成功后对每一条有效事件 MUST 投递延时消息」——改为仅当 `x-delay` 不超过插件可表达上限时 MUST 投递；超上限 MUST NOT 投递且 MUST NOT 因此使整表同步失败。

## Impact

- 代码：`internal/services/voice/predict_imminent.go`（`SyncPredictImminentPending`）；可选常量放同文件或 `eventkit`。
- 规格：`predict-imminent-schedule` 延时投递 Requirement / Scenario。
- 不改：appointment_next 表与 API、消费路径闸门、推送契约、gateway 路由、Redis 键命名。
- 客户端：远预约须在进入可调度窗口后再次 sync（产品/Flutter 约定；本仓可在 design/runbook 备忘，不强制本仓改 Flutter）。
