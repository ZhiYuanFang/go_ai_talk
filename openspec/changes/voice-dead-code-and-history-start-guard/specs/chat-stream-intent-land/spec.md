## MODIFIED Requirements

### Requirement: 流式意图结果直接落地且禁止二次非流式意图调用

VoiceService 的 `HandleTranscriptForIntentStream` 在 Python 流式意图分析成功结束后，MUST 使用本次 Stream 解析出的意图结果执行与 `chatWithResult` 相同的 confirm / 回复透传逻辑；喂养 history 写库 MUST 由 Python 在意图分析阶段经 history `event/batch` 完成，Go MUST NOT 在主路径二次写库。MUST NOT 再调用非流式 `AnalyzeIntent` 或 `callDeepSeekUnifiedIntent`。

#### Scenario: Stream 成功后透传回复

- **WHEN** 调用方调用 `HandleTranscriptForIntentStream` 且 `AnalyzeIntentStream` 成功返回可解析的意图结果
- **THEN** voice SHALL 将该结果映射为统一意图结构（与非流式 `mapPythonRespToIntent` 等价）
- **AND** SHALL 经 `applyUnifiedIntentResult` 处理 NeedConfirm / exit / 自然语言回复
- **AND** MUST NOT 对该次 transcript 再次调用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`
- **AND** MUST NOT 调用已删除的 Go 侧 `handleUnifiedIntentAction` 写库

#### Scenario: Stream 失败不得回退非流式意图分析

- **WHEN** `AnalyzeIntentStream` 失败或结果无法解析（如 Result 为空）
- **THEN** voice SHALL 返回错误或降级话术
- **AND** MUST NOT 回退调用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent` 以补救落库

### Requirement: 保留 quota 单次 consume 与统一落地矩阵

流式落地路径 MUST 与非流式核心路径共享 quota 语义：常规意图分析（非澄清续聊）成功取得可落地意图后 MUST 执行一次 `consumeVoiceAIQuotaOnSuccess`；multi-event 与普通喂养动作的 history 写库 MUST 均由 Python batch 在意图阶段完成，Go 仅透传 `content` 回复。

#### Scenario: quota 单次 consume

- **WHEN** 常规流式意图分析（非 pending confirm 恢复）成功取得可落地意图
- **THEN** voice SHALL 在成功后执行一次 `consumeVoiceAIQuotaOnSuccess`
- **AND** MUST NOT 因同轮重复意图调用导致双倍 consume

#### Scenario: multi-event 由 Python batch 落库

- **WHEN** 流式或非流式意图结果含多条 `events[]` 且 `need_confirm=false`
- **THEN** Python MUST 经 batch 按各子项 op/action 写库
- **AND** Go voice MUST 透传 Python 返回的回复文本，MUST NOT 在 Go 侧逐条 `AddHistory`

### Requirement: 范围与服务边界

本能力约束 Go `internal/services/voice` 内流式意图落地编排；history 写库守卫在 `history-service`；voice MUST NOT 直连他域 DAO。

#### Scenario: 变更边界

- **WHEN** 实现本变更
- **THEN** voice 删码 SHALL 限于 dead 写库/pending 路径
- **AND** 设备/历史读模型访问 SHALL 继续经既有契约（如 `DeviceAdmin()`、`DeviceHistory().ListHistory`）
