## Why

语音喂养落库已迁移至 Python 经 `POST /device/history/api/event/batch` 写库，Go `voice-service` 的 `applyUnifiedIntentResult` 仅透传话术；但 `handleUnifiedIntentAction` 及关联 pending/写库代码仍留在仓库中，与 OpenSpec 描述不一致，增加维护与误改风险。与此同时，`AddDeviceHistory` 对「进行中」（`end_time=0`）记录无同 `eventId` 重复 start 校验，导致语音或 App 可在同一事件未结束时再次插入新的进行中行，与 `EndLatestHistoryIfMatch` 的闭合语义不对称。

## What Changes

### A：删除 voice Go 侧 dead code 并收敛规格

- 删除未被调用的 `handleUnifiedIntentAction`、`handleMultiEventIntent`、`mapPythonIntentToLandPlan` 及仅服务于上述路径的写库/事件树/pending 代码（含 `event_child_pending.go`、`event_history_end.go`、`event_tree.go` 等）。
- 精简 `prepareChatPreamble`（移除已停用的 pending child 清理逻辑）与 `VoiceService` 的 `pendingChild*` 字段。
- 保留仍 live 的符号（如 `parseEventIntentFromReply`、`DeviceHistory().ListHistory` 读路径）。
- 更新 OpenSpec：喂养 CRUD 权威落库为 Python → `event/batch`；Go 经 `applyUnifiedIntentResult` 仅处理 confirm/exit/回复透传。

### B：history-service 禁止同 event 重复 start

- 在 `AddDeviceHistory`（及 `UpdateHistory` 将行改回 `end_time=0` 时）增加守卫：同一 `deviceNo + eventId` 已存在 `end_time=0` 行则拒绝新建/改回进行中。
- `EventBatch` 的 `op=create` 与 App `event/add` 均经 `AddHistory`，自动受益。
- 错误语义与 batch 单条失败风格一致（可读中文 reason）；**不**限制不同 `eventId` 并行进行中（与「中间夹其它事件再结束睡眠」设计兼容）。
- 瞬时记录（`end_time != 0` 或 `start_time == end_time`）**不**触发本守卫。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `end-open-history-by-event`：新增同 event 重复 start 拒绝 Requirement；更新语音 end 落库描述为 Python batch 权威路径。
- `chat-stream-intent-land`：移除对 `handleUnifiedIntentAction` 的 MUST；明确 Go 透传 + Python batch 落库矩阵。
- `python-intent-target-action-align`：喂养写库 SHALL 由 Python 经 history batch 完成，Go 不再 `AddHistory`。
- `intent-clarify-conversation-id`：非确认结果落地描述与上述一致。
- `voice-event-child-disambiguation`：**REMOVED** Go 侧 pending 子事件追问与写库 Requirement（消歧已迁移 Python 意图图）。

## Impact

- **进程**：`voice-service`（删 dead code）；`history-service`（start 守卫）。
- **API**：`event/add`、`event/batch`、`event/update` 在重复 start 时返回失败（非 BREAKING：此前无客户端依赖「允许重复 start」的契约保证）。
- **兄弟仓**：`python-ai-talk` 收到 batch 拒绝时应调整播报（本变更 Go 仓库内不强制改 Python；design 记为可选 follow-up）。
- **无**新 Redis 键、无新背景 ticker、无新 DB 表结构。
- **usage**：不改 `maintenance_skip`（除非负责人另行要求）。
