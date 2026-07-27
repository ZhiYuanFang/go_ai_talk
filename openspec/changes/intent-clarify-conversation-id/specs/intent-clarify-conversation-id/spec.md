## ADDED Requirements

### Requirement: 意图请求可携带 conversation_id 续聊

voice 调用 Python `/v1/analyze/intent` 与 `/v1/analyze/intent/stream` 时，若该设备存在本地保存的澄清 `conversation_id`，MUST 在请求体中附带该字段与用户原文；若不存在则 MUST NOT 伪造。流式与非流式 MUST 使用同一请求字段契约。

#### Scenario: 有 pending cid 时附带续聊

- **WHEN** 设备本地存有上一轮 `need_confirm` 返回的 `conversation_id`
- **AND** 用户再次提交文本进入统一意图路径（流式或非流式）
- **THEN** voice SHALL 调用对应 intent 接口且请求包含该 `conversation_id` 与用户文本
- **AND** SHALL NOT 调用 `/v1/analyze/intent/confirm`

#### Scenario: 无 pending cid 时冷启动

- **WHEN** 设备本地无 `conversation_id`
- **THEN** voice SHALL 发起不含 `conversation_id`（或为空省略）的意图分析请求

### Requirement: need_confirm 仅保存 cid 并透传自然语言

当意图结果 `need_confirm=true` 时，voice MUST 按设备保存返回的 `conversation_id`，MUST 将 Python 返回的自然语言（优先 `confirm_message`，否则 `content`）作为对用户回复，MUST NOT 执行喂养事件 CRUD，MUST NOT 在 Go 侧将用户话术解析为 `confirm|reject`。

#### Scenario: 澄清问句直接回传

- **WHEN** Python 意图结果 `need_confirm=true` 且带有澄清话术与 `conversation_id`
- **THEN** voice SHALL 保存该 `conversation_id`
- **AND** SHALL 以该自然语言作为 Reply 返回给调用方
- **AND** SHALL NOT 写入喂养 history/事件落库路径

### Requirement: 非确认结果清理 cid 后走统一落库或闲聊

当意图结果 `need_confirm=false`（或不需要确认）时，voice MUST 清除该设备本地 `conversation_id`，随后 MUST 按既有行为：conversation/空 target 仅回复；其余经 `handleUnifiedIntentAction`（含 multi-event）做事件相关 CRUD/动作。

#### Scenario: 澄清后落地喂养

- **WHEN** 续聊轮 Python 返回 `need_confirm=false` 且 `target_type` 为可执行喂养/动作意图
- **THEN** voice SHALL 清除本地 `conversation_id`
- **AND** SHALL 走统一动作处理完成事件 CRUD 或等价落库逻辑

#### Scenario: 澄清后闲聊回复

- **WHEN** 续聊轮 Python 返回 `need_confirm=false` 且目标为 conversation 或空 target
- **THEN** voice SHALL 清除本地 `conversation_id`
- **AND** SHALL 仅回复自然语言，不走喂养落库

### Requirement: 删除 ConfirmIntent 与 parseConfirmFeedback 主路径

voice MUST NOT 再提供或调用 `ConfirmIntent`（含对 `/v1/analyze/intent/confirm` 的 HTTP 封装），MUST NOT 使用 `parseConfirmFeedback` 将用户输入映射为 `confirm|reject` 以驱动澄清。`prepareChatPreamble` MUST NOT 因 pending 澄清而短路并绕过统一 intent 调用（`pending child` 本域逻辑除外）。

#### Scenario: 代码路径无 confirm 接口

- **WHEN** 审查 voice 意图澄清相关实现
- **THEN** 不存在对 `/v1/analyze/intent/confirm` 的调用
- **AND** 不存在将用户文本解析为仅 `confirm|reject` 后调用专用恢复接口的主路径

### Requirement: 流式与非流式行为矩阵一致

`chatWithResult`（非流式）与 `HandleTranscriptForIntentStream`（流式落地）MUST 共享同一澄清语义：同一 cid 读写规则、同一 `need_confirm` 透传与清 cid、同一 CRUD 边界；澄清轮 MUST 发起对应的 AnalyzeIntent / AnalyzeIntentStream（而非 ConfirmIntent）。

#### Scenario: 非流式澄清续聊

- **WHEN** 非流式路径上一轮 `need_confirm=true` 后用户再次输入
- **THEN** voice SHALL 经非流式 `AnalyzeIntent` 附带 `conversation_id` 续聊并按结果落地

#### Scenario: 流式澄清续聊

- **WHEN** 流式落地路径上一轮 `need_confirm=true` 后用户再次输入
- **THEN** voice SHALL 经 `AnalyzeIntentStream` 附带 `conversation_id` 续聊并按结果落地
- **AND** SHALL NOT 为澄清轮回退到 ConfirmIntent

### Requirement: 独立意图调用不得串带喂养澄清 cid

成长建议、历史问答等不经过统一 chat preamble 的独立 `AnalyzeIntent` 调用 MUST NOT 自动附带喂养澄清用的本地 `conversation_id`。

#### Scenario: 成长建议不携带澄清 cid

- **WHEN** 调用成长建议类独立意图分析
- **THEN** 请求 SHALL NOT 附带喂养澄清 pending 的 `conversation_id`

### Requirement: 澄清续聊轮免计 AI quota

当统一意图路径（流式或非流式）在发起 `AnalyzeIntent` / `AnalyzeIntentStream` **之前** 本地已存在有效澄清 `conversation_id` 时，voice MUST NOT 对该轮执行喂养 AI 额度预检扣减门槛失败拦截所依赖的计次语义，MUST NOT 在意图成功后对该轮执行额度 consume。冷启动（发起前无有效 cid）MUST 仍走常规 guard 与成功 consume（含结果为 `need_confirm=true` 的提问轮）。流式与非流式 MUST 使用同一免计判定。

#### Scenario: 带 cid 续聊免计次

- **WHEN** 设备本地存有有效澄清 `conversation_id`
- **AND** 用户再次进入统一意图路径并成功完成意图调用
- **THEN** voice SHALL NOT 对该轮执行 `consumeVoiceAIQuotaOnSuccess`（或等价扣减）
- **AND** SHALL 仍发起带 `conversation_id` 的意图分析并按结果落地

#### Scenario: 冷启动含首次澄清提问仍计次

- **WHEN** 设备本地无有效澄清 `conversation_id`
- **AND** 意图结果为 `need_confirm=true`
- **THEN** voice SHALL 按冷启动规则对该轮执行成功后的额度 consume（与普通意图一致）

#### Scenario: 额度用尽仍可澄清续聊

- **WHEN** 用户 AI 额度已用尽
- **AND** 设备本地仍有有效澄清 `conversation_id`
- **THEN** voice SHALL 允许该续聊轮继续调用意图接口（不得因额度预检阻断澄清完成）
