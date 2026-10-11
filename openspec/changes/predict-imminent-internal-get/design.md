## Context

App 已 `PUT /device/api/predict/imminent/pending` 写 Redis；推送经延时 MQ 消费时调用私有 `loadPredictImminentPending`。active-eta 已有对称 internal GET。Python next 通道需同样只读。

## Goals / Non-Goals

**Goals:**

- Internal GET + 密钥中间件；`events` 与 PUT 条目字段对齐。
- 空键安全；导出公开 Load 供 controller 调用。

**Non-Goals:**

- 不改推送闸、dedup、MQ、PUT 清洗；不重算 nextAt；不写测试。

## Decisions

### D1：路径与鉴权

- 路径：`GET /voice/internal/api/predict/imminent/pending?deviceNo=`。
- 挂在 voice-service 与 active-eta 同一 `InternalSecretMiddleware` 组（`X-Device-Gateway-Internal-Secret`）。

### D2：响应与空语义

- GoFrame `data` 内：`{ "events": [ ... ] }`。
- 条目：`eventId`、`nextAt`、`title`、可选 `expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit` / `expectLeafTitle`。
- 缺键、空值、非法 JSON → `events: []`；仅系统级 Redis 错误向上返回。

### D3：Load 导出

- 公开 `LoadPredictImminentPending`；私有 `loadPredictImminentPending` 委托之，推送路径行为保持只读校验。
