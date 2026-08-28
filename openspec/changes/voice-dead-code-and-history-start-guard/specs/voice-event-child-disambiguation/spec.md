## REMOVED Requirements

### Requirement: 命中非叶子事件时必须追问且不写 history

**Reason**: Go `voice-service` 已停用 pending 子事件内存态与 `continuePendingChildEvent`；事件消歧与落库由 `python-ai-talk` 意图图（含 `need_confirm` / `disambiguate`）完成。

**Migration**: 子事件追问语义以 Python 澄清话术与 batch 写库为准；Go 侧 `event_child_pending.go`、`event_tree.go` 已删除。

### Requirement: pending 期间仅在当前父的直接子节点中匹配

**Reason**: 同上，Go pending 子事件路径已移除。

**Migration**: 见上。

### Requirement: pending 为内存态且不跨会话恢复

**Reason**: 同上，Go 不再维护 `pendingChild` map。

**Migration**: 见上。

### Requirement: 仅叶子 event id 可写入 history

**Reason**: Go 侧不再执行喂养 `AddHistory`；叶子校验由 Python batch 提交前在兄弟仓保证。

**Migration**: 在 `python-ai-talk` 意图/batch 层确保非叶子 event 不提交 create。
