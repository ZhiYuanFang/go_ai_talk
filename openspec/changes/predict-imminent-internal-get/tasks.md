## 1. API 与 Load

- [x] 1.1 `api/v1`：`VoiceInternalPredictImminentPendingGetReq/Res`（`events`；字段复用 `PredictImminentPendingEventItem`）
- [x] 1.2 `LoadPredictImminentPending`：缺键/空/非法 JSON → `[]`；私有 load 委托；不改推送逻辑

## 2. 路由

- [x] 2.1 `VoicePredictImminentInternalCtrl.Get` + `register_voice_service` 挂入 `InternalSecretMiddleware` 组

## 3. 验收

- [ ] 3.1 手工：PUT 有 events 后 internal GET 字段对齐；空键 → `events: []`
