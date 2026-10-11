## Why

客户端即将在预测临近 pending 中附带「下次用量/用时」结构化字段。服务端若只存 `eventId`/`nextAt`/`title`，到点推送仍只有「{昵称}要{title}了」，无法传达「预计这次约多久/多少」。需在既有推送句式上追加后缀，且兼容旧客户端不传 expect。

## What Changes

- 扩展 `PUT /device/api/predict/imminent/pending` 单条事件可选字段：`expectKind`、`expectDurationSec`、`expectAmount`、`expectUnit`、`expectLeafTitle`（均 omitempty）。
- Redis 待办与延时叫醒 MQ 载荷 MUST 透传上述字段（与 `title` 同级持久化）。
- 可见推送正文：保留基句 `{昵称}要{title}了`；当载荷含合法胜出 expect 时 MUST 追加统一「约」后缀（duration / amount 模板见 design）；无 expect 或字段不合法时 MUST 退化为仅基句。
- **不改**：整表替换、延时投递、进行中闸、历史闸、去重 TTL、空 title 跳过等既有调度语义。
- **非 BREAKING**：缺省字段时行为与现网一致；允许同路径追加可选 omitempty（不新建 v2，见 design D4）。
- 推送仅扩展 `alert` 后缀；**不**把 expect 写入推送 `data`。
- 孪生仓 Flutter change 同名：`predict-expect-amount-duration`（客户端双线推演与择一上报）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `predict-imminent-schedule`：pending 项与叫醒载荷 MAY/MUST 携带可选 expect 字段并写入 Redis / MQ。
- `predict-imminent-push-copy`：在既有昵称+title 模板上，有合法 expect 时 MUST 追加「约」后缀文案。

## Impact

- 代码：`api/v1/device_predict_imminent_http.go`、`PredictImminentPendingItem` / fire payload、`SyncPredictImminentPending`、`handlePredictImminentFire` 文案拼接；controller 映射字段。
- 不改推送厂商通道与 `bizType`。
- 须与 Flutter 同名 change 字段名对齐后联调。
