## MODIFIED Requirements

### Requirement: 意图落地按 Python target_type 与 action 两轴分支

voice 统一意图落地（`applyUnifiedIntentResult` 及流式/非流式共用路径）MUST 以 Python `target_type` 为领域、`action` 为喂养动作，MUST NOT 将 `target_type` 直接传入 `ParseActionTargetType` 作为 CRUD 开关。权威枚举与兄弟仓 `IntentResponse` 一致：`target_type` 为 feeding|history|suggest|conversation|exit；喂养 `action` 为 start|end|one|multi|disambiguate。喂养 CRUD MUST 由 Python 经 `POST /device/history/api/event/batch` 完成；Go voice MUST NOT 调用 `DeviceHistory().AddHistory` / `UpdateHistory` / `DeleteHistory` 执行喂养写库。

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

当 `need_confirm=false` 时：`target_type=history` MUST 走历史查询/search 类处理（Python 回复或读 history，不在 Go 写库）；`target_type=suggest` MUST 走成长建议类处理；`target_type=exit` MUST 走退出；`target_type=conversation` 或空 MUST 仅返回自然语言而不做喂养 CRUD。

#### Scenario: history 不误入 conversation 早退

- **WHEN** `target_type=history` 且 `need_confirm=false`
- **THEN** voice SHALL 进入历史/search 处理路径（透传 Python 回复）

#### Scenario: conversation 仅回复

- **WHEN** `target_type=conversation` 或空
- **THEN** voice SHALL 仅回复内容，SHALL NOT 写入喂养 history

### Requirement: disambiguate 不落库

当 `action=disambiguate` 且未走 `need_confirm` 澄清早退时，Python MUST NOT 执行喂养 batch 写库；Go voice MUST 透传自然语言回复。

#### Scenario: disambiguate 防御

- **WHEN** `target_type=feeding` 且 `action=disambiguate` 且 `need_confirm=false`
- **THEN** Python MUST NOT 提交喂养 batch create/start
- **AND** Go voice MUST NOT 调用 history 写库

## REMOVED Requirements

### Requirement: 禁止错误 Parse target_type 驱动单事件动作

**Reason**: Go 侧 `handleUnifiedIntentAction` 与 `entity.Action.TargetType` 映射已删除；喂养 CRUD 由 Python batch 驱动，不再经 Go 构造 Action 写库。

**Migration**: 审查 Python 意图图与 batch items 的 op/action 映射；Go 仅保留 `applyUnifiedIntentResult` 透传路径。
