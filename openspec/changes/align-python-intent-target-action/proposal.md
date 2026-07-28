## Why

Python `python_ai_talk` 意图响应用两轴语义：`target_type`（领域：feeding/history/suggest/conversation/exit）与 `action`（动作：start/end/one/search/suggestion/reply/exit/multi/disambiguate）。Go 侧 `applyUnifiedIntentResult` 误将 `target_type` 传入 `ParseActionTargetType`，导致 `feeding` 被默认成 `conversation`，喂养意图成功却不走 `AddHistory`。澄清续聊（`intent-clarify-conversation-id`）落地后该错位更易暴露，必须以 Python 契约为准对齐 Go 映射。

## What Changes

- 以 Python `IntentResponse` 为准，在 Go 统一意图落地路径按 **先 `target_type` 再 `action`** 分支，不再把 `target_type` 当作 `ActionTargetType`。
- `feeding` + `start|end|one` → 经 `handleUnifiedIntentAction` 做事件 CRUD；`feeding` + `multi` → 既有多事件路径；`feeding` + `disambiguate`（无 need_confirm 时）→ 仅自然语言，不落库。
- `history` → 历史/search 路径；`suggest` → 成长建议路径；`conversation`/空 → 仅回复；`exit` → 退出。
- 修正单事件构造 `entity.Action.TargetType` 时使用 `action`（喂养）或由 `target_type` 映射到 Go 动作枚举（history→search、suggest→suggest、exit→exit）。
- 不改 Python 服务；不改 `handleUnifiedIntentAction` 内部 CRUD 语义（仅修正入口映射）。

## Capabilities

### New Capabilities

- `python-intent-target-action-align`：Go voice 统一意图落地与 Python `target_type`/`action` 两轴契约对齐，保证喂养等可执行意图正确进入 CRUD。

### Modified Capabilities

- （无）主规格树无独立条目；相关历史变更 `python-intent-crud-ready` / `intent-clarify-conversation-id` 假定 Python 字段可直接驱动 CRUD，本能力补齐 Go 侧映射缺口。

## Impact

- **代码**：主要 `internal/services/voice/voice_chat_understanding.go`（`applyUnifiedIntentResult` 及同类单事件 Action 构造）；必要时小幅注释/`definitions` 说明两轴差异。
- **行为**：`target_type=feeding` + `action=one|start|end` 将恢复 history 入库；`history`/`suggest` 按领域正确分流，不再因 Parse 默认掉进 conversation。
- **依赖**：依赖已对齐的 Python intent 响应（含澄清续聊）；与 `intent-clarify-conversation-id` 正交但建议同批验证。
- **非范围**：gateway-app、Redis、quota、confirm 通道、`confirm_type`/`options` UI。
