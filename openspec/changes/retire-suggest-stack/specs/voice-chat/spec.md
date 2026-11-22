# Spec Delta

## MODIFIED Requirements

### Requirement: 独立意图调用不得串带喂养澄清 cid

历史问答等不经过统一 chat preamble 的独立 `AnalyzeIntent` 调用 MUST NOT 自动附带喂养澄清用的本地 `conversation_id`。系统 MUST NOT 再保留成长建议类独立意图分析入口作为正式产品路径。

#### Scenario: 成长建议不携带澄清 cid

- **WHEN** 调用成长建议类独立意图分析（若遗留调用仍存在）
- **THEN** 请求 SHALL NOT 附带喂养澄清 pending 的 `conversation_id`
- **AND** 系统 MUST NOT 将此类调用作为正式产品入口保留

### Requirement: 意图落地按 Python target_type 与 action 两轴分支

voice 统一意图落地（`applyUnifiedIntentResult` 及流式/非流式共用路径）MUST 以 Python `target_type` 为领域、`action` 为喂养动作，MUST NOT 将 `target_type` 直接传入 `ParseActionTargetType` 作为 CRUD 开关。权威枚举与兄弟仓对齐时，`target_type` 可仍含历史字面 `suggest`，但 Go voice MUST NOT 将其作为成长建议写库或专用 LLM 路径的触发条件。喂养 `action` 为 start|end|one|multi|disambiguate。喂养 CRUD MUST 由 Python 经 `POST /device/history/api/event/batch` 完成；Go voice MUST NOT 调用 `DeviceHistory().AddHistory` / `UpdateHistory` / `DeleteHistory` 执行喂养写库。

#### Scenario: feeding + one 由 Python batch 落库

- **WHEN** 意图结果 `need_confirm=false` 且 `target_type=feeding` 且 `action=one`
- **AND** 事件可解析成功
- **THEN** Python MUST 经 batch 写入一次性 history 记录
- **AND** Go voice SHALL 透传 Python 回复，MUST NOT 在 Go 侧 `AddHistory`

#### Scenario: feeding + start/end 由 Python batch 落库

- **WHEN** `need_confirm=false` 且 `target_type=feeding` 且 `action` 为 `start` 或 `end`
- **THEN** Python MUST 经 batch 执行开始或结束（end 优先 `EndLatestHistoryIfMatch` 语义）
- **AND** Go voice SHALL 透传回复

#### Scenario: feeding + multi 由 Python batch 落库

- **WHEN** `need_confirm=false` 且 `target_type=feeding` 且 `action=multi` 且 `events` 非空
- **THEN** Python MUST 经 batch 按各子项 op/action 顺序写库
- **AND** Go voice MUST NOT 保留 Go 侧 multi-event 写库处理器

### Requirement: 非喂养领域按 target_type 分流

当 `need_confirm=false` 时：`target_type=history` MUST 走历史查询/search 类处理（Python 回复或读 history，不在 Go 写库）；`target_type=suggest` MUST NOT 触发成长建议生成、写库或专用 LLM 路径（可按自然语言透传或与 conversation 同等对待）；`target_type=exit` MUST 走退出；`target_type=conversation` 或空 MUST 仅返回自然语言而不做喂养 CRUD。

#### Scenario: history 不误入 conversation 早退

- **WHEN** `target_type=history` 且 `need_confirm=false`
- **THEN** voice SHALL 进入历史/search 处理路径（透传 Python 回复）

#### Scenario: conversation 仅回复

- **WHEN** `target_type=conversation` 或空
- **THEN** voice SHALL 仅回复内容，SHALL NOT 写入喂养 history

#### Scenario: suggest 字面不落成长建议库

- **WHEN** `target_type=suggest` 且 `need_confirm=false`
- **THEN** voice MUST NOT 向 `suggest` 表写入，MUST NOT 调用已删除的成长建议生成函数
