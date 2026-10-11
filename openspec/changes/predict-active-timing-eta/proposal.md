## Why

客户端将上报「进行中计时预计结束」；意图智能体需要按宝宝查询该 ETA。须与预测临近 pending（下次发生推送）分离存储，并提供内部读接口给 Python。

## What Changes

- 新增 App `PUT /device/api/predict/active-eta/pending`：全量替换进行中 ETA 列表（登录+绑机校验）；空 `items` 清空 Redis。
- Redis 键与同步短锁登记于 `cachekit`（沿用 `PredictImminentPending`：权威在 Redis、无 TTL、整表替换）；**不**发延时 MQ、**不**触发推送。
- 新增内部 `GET /voice/internal/api/predict/active-eta?deviceNo=`（voice 宿主，供 Python）：返回当前 ETA 列表；缺键返回空列表。
- 单条字段与 Flutter 对齐：`rootEventId`、`leafEventId`、可选 `historyId`、`title`、`startAt`、`expectEndAt`（unix 秒）。
- **不改** `predict/imminent/pending` 调度与推送文案。
- **Usage**：负责人确认 `PUT /device/api/predict/active-eta/pending` **不计入** App API 使用统计（写入 `maintenance_skip.go`，对标 imminent pending）。
- 孪生：`flutter_ai_talk` / `python_ai_talk` 同名 change。

## Capabilities

### New Capabilities

- `predict-active-timing-eta`：进行中 ETA 的写入、Redis 权威、内部读取契约。

### Modified Capabilities

（无）

## Impact

- 代码：`api/v1`、voice controller/service、`cachekit` 键与同步锁、`gatewayapp/usagestats/maintenance_skip.go`。
- 宿主：**voice-service**（经 gateway-app `/device/api/predict/*` 反代；无需新增 proxy 前缀；须登录，不进 auth exempt）。
- Python 经 internal GET 消费；无 MySQL 权威表。
