## Why

统一列表态 Live Activity（`unify-live-activity-list`）当前把 Redis imminent pending **全量**投影进锁屏，远点预测也会占位，且无总条数上限；同时 `push_live_activity_token` 仍保留已无业务意义的 `attributes_kind` / `activity_id` 空列。锁屏应只展示「进行中 + 已进 lead 窗、即将/应推的预测」，并硬顶 3 条，避免噪声与挤占。

## What Changes

- **预测过滤**：Refresh 组装时，预测段只保留 `nextAt - leadSeconds ≤ now` 的 pending 项（`leadSeconds` 与普通推送同一 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS`，默认 300）；远点 pending 仍可留在 Redis / MQ 调度，但 **MUST NOT** 进入 LA items。
- **总条数上限**：`items` 合计最多 **3** 条；截断优先 **进行中（A）**：先按既有排序取 open timing，再用进窗预测填满剩余名额。
- **展示顺序不变**：仍为预测段（进窗、`nextAt` 升序）在前，进行中段（`startTime` 倒序）在后；截断只决定入选集合，不改为时间混排。
- **Token 表清理**：物理去掉（或迁移删除）`push_live_activity_token.attributes_kind`、`activity_id` 列及代码中的空串写入；register / APNs 单一 `attributes-type` / ContentState item `kind` 判别保持不变。
- **非目标**：不改 leadSeconds 数值语义、普通 `predict_imminent` 闸门、pending PUT 全量替换、delayed MQ 调度；不改 Flutter Extension UI。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `unified-live-activity-push`：Refresh 预测段由 pending 全量改为 lead 进窗过滤；items 总条数 ≤ 3 且进行中优先截断；token 表不再保留 `attributes_kind` / `activity_id` 兼容列。

## Impact

- `internal/services/history/live_activity_refresh.go`：`buildLiveActivityItems` 过滤 + 截断。
- history 侧读取 `leadSeconds`：与 voice 同 env（`VOICE_PREDICT_IMMINENT_LEAD_SECONDS`）或抽到 `platform` 共享读法，**禁止** history import `services/voice`。
- `internal/services/push/push_schema.go`、`push_live_activity_token.go`：DROP/停用列，去掉恒写空串。
- 兄弟仓 Flutter：列表契约字段不变，仅语义变为「更短、更近」；表结构迁移需与部署同拍。
- 无新 App HTTP 接口；usage 统计不改。
