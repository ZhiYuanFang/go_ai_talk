## ADDED Requirements

### Requirement: 流式意图结果直接落地且禁止二次非流式意图调用
VoiceService 的 `HandleTranscriptForIntentStream` 在 Python 流式意图分析成功结束后，MUST 使用本次 Stream 解析出的意图结果执行与 `chatWithResult` 相同的 confirm / 落库 / 回复逻辑；MUST NOT 再调用非流式 `AnalyzeIntent` 或 `callDeepSeekUnifiedIntent`。

#### Scenario: Stream 成功后直接落库或回复
- **WHEN** 调用方调用 `HandleTranscriptForIntentStream` 且 `AnalyzeIntentStream` 成功返回可解析的意图结果
- **THEN** voice SHALL 将该结果映射为统一意图结构（与非流式 `mapPythonRespToIntent` 等价）
- **AND** SHALL 按统一行为矩阵处理（NeedConfirm / conversation / multi-event / `handleUnifiedIntentAction`）
- **AND** MUST NOT 对该次 transcript 再次调用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`

#### Scenario: Stream 失败不得回退非流式意图分析
- **WHEN** `AnalyzeIntentStream` 失败或结果无法解析（如 Result 为空）
- **THEN** voice SHALL 返回错误或降级话术
- **AND** MUST NOT 回退调用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent` 以补救落库

### Requirement: 保留 NeedConfirm 与 pending confirm 行为矩阵
流式落地路径 MUST 保留与 `confirm-ws-adaptation` / 现网 `chatWithResult` 一致的确认语义：`NeedConfirm` 置 pending、下一轮 confirm/reject 走 `ConfirmIntent`，且 pending 轮次不得无谓发起新的意图 Stream。

#### Scenario: Stream 返回 NeedConfirm
- **WHEN** 流式意图结果 `NeedConfirm=true`
- **THEN** voice SHALL 保存 pending confirm（含 `conversation_id` 等）
- **AND** SHALL 向调用方返回确认话术（优先 `ConfirmMessage`，否则 `Reply`）
- **AND** SHALL 不执行喂养事件落库

#### Scenario: 存在 pending confirm 时优先处理用户反馈
- **WHEN** 设备已有 pending confirm，且用户输入可解析为 confirm 或 reject
- **THEN** voice SHALL 调用 `ConfirmIntent` 恢复图执行并走统一落库/回复逻辑
- **AND** MUST NOT 再对该轮输入调用 `AnalyzeIntentStream` 或非流式 `AnalyzeIntent`

#### Scenario: pending 反馈无法识别则清理后进入常规流式意图
- **WHEN** 设备已有 pending confirm，但用户输入无法解析为 confirm/reject
- **THEN** voice SHALL 清理 pending
- **AND** SHALL 进入常规流式意图分析（仅一次 Stream）后落地

### Requirement: 保留 quota 单次 consume 与 multi-event / handleUnifiedIntentAction
流式落地路径 MUST 与非流式核心路径共享动作执行语义：quota 成功才 consume 一次；multi-event 与普通动作分别走既有处理器。

#### Scenario: quota 单次 consume
- **WHEN** 常规流式意图分析（非 pending confirm 恢复）成功取得可落地意图
- **THEN** voice SHALL 在成功后执行一次 `consumeVoiceAIQuotaOnSuccess`
- **AND** MUST NOT 因「曾调用 Stream + 曾误入 chatWithResult」导致同轮双倍 consume

#### Scenario: multi-event 与普通动作
- **WHEN** 流式意图 `Action` 为 `multi` 且 Events 非空
- **THEN** voice SHALL 走既有 multi-event 处理逻辑
- **WHEN** 流式意图为普通非 conversation 动作且无需确认
- **THEN** voice SHALL 调用 `handleUnifiedIntentAction` 完成落库与回复

### Requirement: TTS 非流式路径不受影响
现有 TTS / 非流式对话入口 MUST 继续可走非流式意图分析，行为与本变更前对调用方可见效果一致（落库语义与流式落地共享 apply 逻辑时除外，对外仍一次意图调用）。

#### Scenario: HandleTranscriptForStreaming 仍可非流式意图分析
- **WHEN** 调用现有 `HandleTranscriptForStreaming`（或其它走 `chatWithResult` 常规意图路径的入口）
- **THEN** voice MAY 继续使用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`
- **AND** 落库/confirm 行为矩阵 SHALL 与流式落地路径一致（共享公共 apply 逻辑）

### Requirement: 范围与服务边界
本能力仅约束 Go `internal/services/voice` 内流式意图落地编排；MUST NOT 修改 gateway、Flutter 协议、tip/feedback；voice MUST NOT 直连他域 DAO。

#### Scenario: 变更边界
- **WHEN** 实现本变更
- **THEN** 代码修改 SHALL 限于 `internal/services/voice`（主要 `voice_chat.go`、`voice_chat_understanding.go`）
- **AND** 设备读模型访问 SHALL 继续经既有契约（如 `DeviceAdmin()`），不得新增他域 DAO import
