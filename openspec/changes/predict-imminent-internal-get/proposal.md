## Why

Python 意图双通道（`intent-predict-dual-channel`）的 `next` 通道需要只读 imminent 待办；现有仅有 App PUT + 推送消费路径，无私有对外 GET。须对齐已落地的 active-eta internal GET。

## What Changes

- 新增 `GET /voice/internal/api/predict/imminent/pending?deviceNo=`（voice-service，`InternalSecretMiddleware`）。
- 响应 `{ "events": [ ... ] }`，字段对齐 App PUT `PredictImminentPendingEventItem`。
- 缺键/空/非法 JSON → `events: []`（不当事故）。
- 复用/导出 `LoadPredictImminentPending`；**不改**推送、延时 MQ、同步 PUT 逻辑。

## Capabilities

### New Capabilities

- `predict-imminent-internal-get`：imminent 待办内部只读契约。

### Modified Capabilities

（无）

## Impact

- 代码：`api/v1`、voice internal controller、`predict_imminent.go` Load 导出、`register_voice_service.go`。
- 消费方：`python_ai_talk` `intent-predict-dual-channel` 任务 2.1/2.2。
