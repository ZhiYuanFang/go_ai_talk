## Why

运维在 Hub「设备数据运维」页无法直接查看 App 上报到 Redis 的预测数据（临近待办与进行中 ETA），联调只能翻日志或 Redis。需要在单设备页加只读卡片，一眼核对字段值（含进行中事件的预估结束时间）。

## What Changes

- 新增 voice Admin 只读接口：按 `deviceNo` 返回 imminent pending 与 active-eta 全量字段（复用既有 `Load*`，不改写路径）。
- 在 `history.html`（设备数据运维）增加**一张卡、两段**：临近待办 + 进行中 ETA；字段完整展示；支持刷新；空列表明确提示。
- **只读**：不提供清空/改写 Redis、不发 MQ、不推送。
- **不改** App PUT、internal GET、推送调度语义。

## Capabilities

### New Capabilities

- `hub-predict-report-card`：Hub 设备数据运维页预测上报只读卡片（Admin API + 静态页两段展示）。

### Modified Capabilities

- （无）

## Impact

- **voice-service**：`api/v1` Admin GET DTO + controller；鉴权同既有 voice Admin（`X-Admin-Password`，由 gateway-app 注入）；复用 `LoadPredictImminentPending` / `LoadPredictActiveEta`。
- **gateway-app**：`/voice/admin/api/*` 反代已覆盖，通常无需新绑定；须部署更新后的 `history.html` 静态资源。
- **Hub 静态**：`resource/public/history.html`。
- **usage**：Admin 非 App 接口，不改 `maintenance_skip`。
- **部署**：voice-service + gateway-app（静态页）；无新环境变量。
