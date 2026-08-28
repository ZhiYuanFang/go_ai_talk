## MODIFIED Requirements

### Requirement: 非确认结果清理 cid 后走统一落库或闲聊

当意图结果 `need_confirm=false`（或不需要确认）时，voice MUST 清除该设备本地 `conversation_id`，随后 MUST：`target_type=conversation` 或空 target 仅透传 Python 自然语言回复；喂养/多事件写库 MUST 已由 Python 在同轮意图分析中经 history `event/batch` 完成，Go `applyUnifiedIntentResult` MUST NOT 再执行 Go 侧事件 CRUD。

#### Scenario: 澄清后落地喂养

- **WHEN** 续聊轮 Python 返回 `need_confirm=false` 且 `target_type=feeding` 且 batch 已成功写库
- **THEN** voice SHALL 清除本地 `conversation_id`
- **AND** SHALL 透传 Python 返回的回复文本

#### Scenario: 澄清后闲聊回复

- **WHEN** 续聊轮 Python 返回 `need_confirm=false` 且目标为 conversation 或空 target
- **THEN** voice SHALL 清除本地 `conversation_id`
- **AND** SHALL 仅回复自然语言，不走 Go 侧喂养落库
