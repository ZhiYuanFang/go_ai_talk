## ADDED Requirements

### Requirement: 意图落地按 Python target_type 与 action 两轴分支

voice 统一意图落地（`applyUnifiedIntentResult` 及流式/非流式共用路径）MUST 以 Python `target_type` 为领域、`action` 为喂养动作，MUST NOT 将 `target_type` 直接传入 `ParseActionTargetType` 作为 CRUD 开关。权威枚举与兄弟仓 `IntentResponse` 一致：`target_type` 为 feeding|history|suggest|conversation|exit；喂养 `action` 为 start|end|one|multi|disambiguate。

#### Scenario: feeding + one 进入一次性记录

- **WHEN** 意图结果 `need_confirm=false` 且 `target_type=feeding` 且 `action=one`
- **AND** 事件可解析成功
- **THEN** voice SHALL 调用喂养一次性记录路径（含 `DeviceHistory().AddHistory` 或等价写库）
- **AND** SHALL NOT 仅因 `target_type=feeding` 将其当作 conversation 早退

#### Scenario: feeding + start/end 进入计时记录

- **WHEN** `need_confirm=false` 且 `target_type=feeding` 且 `action` 为 `start` 或 `end`
- **THEN** voice SHALL 分别走开始/结束记录路径

#### Scenario: feeding + multi 走多事件

- **WHEN** `need_confirm=false` 且 `target_type=feeding` 且 `action=multi` 且 `events` 非空
- **THEN** voice SHALL 走既有多事件处理路径（各子项以子 `action` 为动作）

### Requirement: 非喂养领域按 target_type 分流

当 `need_confirm=false` 时：`target_type=history` MUST 走历史查询/search 类处理；`target_type=suggest` MUST 走成长建议类处理；`target_type=exit` MUST 走退出；`target_type=conversation` 或空 MUST 仅返回自然语言而不做喂养 CRUD。领域判定 MUST 使用 `target_type`，MUST NOT 依赖将 Python `action=suggestion` 误当作 Go `suggest` 枚举的唯一依据。

#### Scenario: history 不误入 conversation 早退

- **WHEN** `target_type=history` 且 `need_confirm=false`
- **THEN** voice SHALL 进入历史/search 处理路径，而非仅因 Parse 默认而当作闲聊早退

#### Scenario: conversation 仅回复

- **WHEN** `target_type=conversation` 或空
- **THEN** voice SHALL 仅回复内容，SHALL NOT 写入喂养 history

### Requirement: disambiguate 不落库

当 `action=disambiguate`（或等价消歧动作）且未走 `need_confirm` 澄清早退时，voice MUST NOT 将其当作 start/end/one 执行喂养 CRUD，MUST 以自然语言回复（或安全降级），避免误入库。

#### Scenario: disambiguate 防御

- **WHEN** `target_type=feeding` 且 `action=disambiguate` 且 `need_confirm=false`
- **THEN** voice SHALL NOT 调用一次性/起止记录写库路径

### Requirement: 禁止错误 Parse target_type 驱动单事件动作

单事件构造传给 `handleUnifiedIntentAction` 的 `entity.Action.TargetType` MUST 来自对齐后的映射（喂养用 `action`；history→search；suggest→suggest；exit→exit），MUST NOT 使用 `ParseActionTargetType(intent.TargetType)` 的结果作为喂养 CRUD 开关。

#### Scenario: 映射源正确

- **WHEN** 审查单事件落地代码路径
- **THEN** 不存在以 `ParseActionTargetType(intent.TargetType)` 决定 feeding CRUD 的主路径
