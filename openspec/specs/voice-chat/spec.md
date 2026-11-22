# voice-chat

## Purpose

语音会话 ASR/LLM/TTS、流式对话、意图落地、理解链路与 voice 服务分解的行为规格。 Realtime and text-chat voice paths owned by voice-service.

<!-- migrated-from: chat-stream-intent-land, chat-stream-reply-only, chat-streaming, feeding-chat-ws-io, intent-clarify-conversation-id, python-intent-target-action-align, voice-and-device-service-decomposition, voice-chat-dashscope-stt, voice-device-domain-http-access, voice-event-child-disambiguation, voice-internal-text-chat, voice-realtime-asr-ws, voice-route-canary-management, voice-textchat-resilience, voice-understanding-python-fixed -->

## Requirements

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

流式落地路径 MUST 与非流式核心路径共享 quota 语义：常规意图分析（非澄清续聊）成功取得可落地意图后 MUST 执行一次 `consumeVoiceAIQuotaOnSuccess`；multi-event 与普通喂养动作的 history 写库 MUST 均由 Python batch 在意图阶段完成，Go 仅透传 `content` 回复。

#### Scenario: quota 单次 consume

- **WHEN** 常规流式意图分析（非 pending confirm 恢复）成功取得可落地意图
- **THEN** voice SHALL 在成功后执行一次 `consumeVoiceAIQuotaOnSuccess`
- **AND** MUST NOT 因同轮重复意图调用导致双倍 consume

#### Scenario: multi-event 由 Python batch 落库

- **WHEN** 流式或非流式意图结果含多条 `events[]` 且 `need_confirm=false`
- **THEN** Python MUST 经 batch 按各子项 op/action 写库
- **AND** Go voice MUST 透传 Python 返回的回复文本，MUST NOT 在 Go 侧逐条 `AddHistory`

### Requirement: TTS 非流式路径不受影响

现有 TTS / 非流式对话入口 MUST 继续可走非流式意图分析，行为与本变更前对调用方可见效果一致（落库语义与流式落地共享 apply 逻辑时除外，对外仍一次意图调用）。

#### Scenario: HandleTranscriptForStreaming 仍可非流式意图分析
- **WHEN** 调用现有 `HandleTranscriptForStreaming`（或其它走 `chatWithResult` 常规意图路径的入口）
- **THEN** voice MAY 继续使用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`
- **AND** 落库/confirm 行为矩阵 SHALL 与流式落地路径一致（共享公共 apply 逻辑）

### Requirement: 范围与服务边界

本能力约束 Go `internal/services/voice` 内流式意图落地编排；history 写库守卫在 `history-service`；voice MUST NOT 直连他域 DAO。

#### Scenario: 变更边界

- **WHEN** 实现本变更
- **THEN** voice 删码 SHALL 限于 dead 写库/pending 路径
- **AND** 设备/历史读模型访问 SHALL 继续经既有契约（如 `DeviceAdmin()`、`DeviceHistory().ListHistory`）

---

### Requirement: Chat stream pushes thinking then business reply only

文本对话流式路径（voice internal chat/stream 及经 history 委派的对外 chat/stream）MUST 在思考阶段仅推送 `event: thinking`；MUST 在意图落地（或 preamble 短路）后推送业务话术为 `event: answer`。MUST NOT 将 Python 意图分析 JSON 增量作为 `event: answer` 推送给客户端。

#### Scenario: Successful stream land with thinking

- **WHEN** 调用方发起流式文本对话且 Python `AnalyzeIntentStream` 产出 thinking 增量并成功落地得到业务 Reply
- **THEN** 服务端 SHALL 将 thinking 增量以 SSE `event: thinking` 推送
- **AND** SHALL 在落地完成后以 SSE `event: answer` 推送业务 Reply
- **AND** MUST NOT 推送意图结果 JSON 作为 answer

#### Scenario: Preamble short-circuit without Python stream

- **WHEN** preamble 短路（如 pending child）直接得到业务 Reply 且未调用 AnalyzeIntentStream
- **THEN** 服务端 MAY 不推送 thinking
- **AND** SHALL 以 SSE `event: answer` 推送业务 Reply

### Requirement: Intent stream callback has no OnAnswer

`contracts.IntentStreamCallback` MUST NOT 提供用于外泄意图 JSON 的 `OnAnswer` 回调；流式过程对外仅允许 `OnThinking`。Python 客户端内部累积意图 JSON 供解析 MUST 仍可用，但 MUST NOT 再转发到 `IntentStreamCallback`。

#### Scenario: Intent JSON not forwarded via callback

- **WHEN** `HandleTranscriptForIntentStream` 处理 AnalyzeIntentStream 的 answer（意图 JSON）事件
- **THEN** 系统 SHALL 在内部累积并解析意图结果
- **AND** MUST NOT 通过 `IntentStreamCallback` 将 JSON 增量回调给调用方

### Requirement: HandleTranscriptForIntentStream returns chat result

`HandleTranscriptForIntentStream` 成功或可降级返回时，返回结构 MUST 对齐聊天结果语义，至少包含 Ask、Reply、Exit、FinishTalk；MUST NOT 再要求调用方依赖 Thinking 或 AnswerJSON 字段消费 UI 内容（UI 内容由 thinking 回调与 Reply 承担）。

#### Scenario: Controller writes reply from return value

- **WHEN** voice internal ChatStream 调用 `HandleTranscriptForIntentStream` 结束且 Reply 非空
- **THEN** 控制器 SHALL 使用返回值中的 Reply 写入 SSE `event: answer`

#### Scenario: Stream failure surfaces error event

- **WHEN** 流式意图分析失败或业务错误返回
- **THEN** 控制器 SHALL 写入 SSE `event: error`
- **AND** 若返回结构仍含降级 Reply，MAY 同时写入 `event: answer`

---

### Requirement: Chat streaming API

Go 服务端 SHALL 暴露 `POST /device/history/api/chat/stream` 接口，以 SSE 方式流式返回文本对话的思考过程与最终回复。

#### Scenario: Flutter client requests chat via SSE
- **WHEN** Flutter 客户端向 `/device/history/api/chat/stream` 发送 POST 请求（body 包含 deviceNo、transcript）
- **THEN** Go 服务端 SHALL 立即返回 SSE 响应头（Content-Type: text/event-stream、Cache-Control: no-cache）
- **AND** 按序推送 `thinking`、`answer` 两类 SSE 事件
- **AND** 最终以 `data: [DONE]` 结束流式响应

### Requirement: Chat thinking events streamed

Go 服务端 SHALL 将 AI 思考过程增量实时推送给客户端。

#### Scenario: Chat thinking delta pushed to client
- **WHEN** Go 服务端从 voice 服务流式意图分析收到 thinking delta 回调
- **THEN** Go 服务端 SHALL 立即将该 thinking delta 以 SSE `event: thinking` 帧转发给客户端
- **AND** 不得缓冲或延迟超过 100ms

### Requirement: Chat answer events streamed

Go 服务端 SHALL 将 AI 最终回复内容增量实时推送给客户端。

#### Scenario: Chat answer delta pushed to client
- **WHEN** Go 服务端从 voice 服务流式意图分析收到 answer delta 回调
- **THEN** Go 服务端 SHALL 立即将该 answer delta 以 SSE `event: answer` 帧转发给客户端

### Requirement: Chat streaming via history delegate

history 服务 SHALL 以 HTTP SSE 客户端方式委派流式对话请求到 voice 服务。

#### Scenario: History service delegates streaming chat to voice
- **WHEN** history 服务的 `DelegateTextChatStream()` 被调用
- **THEN** history 服务 SHALL 向 voice 服务的 internal 接口 `/voice/internal/api/text/chat/stream` 发起 SSE 请求
- **AND** 将 voice 服务返回的 SSE 事件逐帧透传给调用方

### Requirement: Voice internal streaming chat API

voice 服务 SHALL 暴露 internal SSE 接口 `/voice/internal/api/text/chat/stream`，供跨进程委派使用。

#### Scenario: Voice internal chat streaming endpoint
- **WHEN** 经内部认证的请求到达 `/voice/internal/api/text/chat/stream`
- **THEN** voice 服务 SHALL 调用 `HandleTranscriptForIntentStream()` 执行流式意图分析
- **AND** 将 OnThinking/OnAnswer 回调以 SSE 帧写入 HTTP 响应
- **AND** 流式结束后推送 `data: [DONE]`

### Requirement: Chat stream preserves sync API compatibility

原同步接口 `/device/history/api/chat` SHALL 保持完全不变，继续提供一次性返回 reply 的能力。

#### Scenario: Sync chat API still works
- **WHEN** 客户端调用原同步接口 `/device/history/api/chat`
- **THEN** Go 服务端 SHALL 以与变更前完全一致的同步方式返回 `{reply: "..."}`

### Requirement: Chat API error handling

当 AI 服务不可用时，Go 服务端 SHALL 返回固定错误提示。

#### Scenario: AI unavailable during streaming chat
- **WHEN** 流式对话过程中 AI 服务调用失败
- **THEN** Go 服务端 SHALL 以 SSE `event: error` 帧推送错误消息："AI服务暂时不可用，请稍后再试"
- **AND** 紧接着推送 `data: [DONE]` 结束连接

### Requirement: Flutter sendCommand uses streaming

Flutter 端 `sendCommand()` SHALL 改为 SSE 流式调用 `/chat/stream` 接口。

#### Scenario: Flutter streaming chat display
- **WHEN** Flutter 调用 `sendCommand(text)`
- **THEN** Flutter SHALL 建立 SSE 连接到 `/device/history/api/chat/stream`
- **AND** 收到 thinking 事件时更新 `_chatThinking` 状态并在消息条展示
- **AND** 收到 answer 事件时累积 `_chatReply` 并切换展示最终回复
- **AND** 收到 `data: [DONE]` 时关闭 SSE 连接

---

### Requirement: chat WS 支持输入输出模态组合

`/voice/chat/ws` MUST 在保留现有音频流式能力的基础上，支持自然语言喂养的四种 I/O 组合：输入 `audio|text`，输出 `audio|text`。模态 MUST 由客户端在 `start` 帧中声明（字段名实现可选 `inputModality`/`outputModality` 或等价约定）；缺省 MUST 为 `audio`/`audio`，以保持横屏语音兼容。

#### Scenario: 默认音入音出兼容横屏

- **WHEN** 客户端发送既有横屏 `start`（未声明模态或显式 audio/audio）并上传 PCM
- **THEN** 服务端 MUST 按现网行为进行 ASR、意图流式、TTS，并下发 `thinking_delta`/`answer`/`audio_*`（及既有 asr 帧）

#### Scenario: 文入文出

- **WHEN** 客户端 `start` 声明 text 输入与 text 输出，随后发送 text 上行帧（非空文本）
- **THEN** 服务端 MUST NOT 依赖 PCM/ASR 完成该轮，MUST 下发 `thinking_delta` 与 `answer`，MUST NOT 下发 TTS `audio_chunk`/`audio_end`

#### Scenario: 文入音出

- **WHEN** 客户端声明 text 输入与 audio 输出并发送 text 帧
- **THEN** 服务端 MUST 跳过 ASR，MUST 在意图完成后执行 TTS 并下发音频帧

#### Scenario: 音入文出

- **WHEN** 客户端声明 audio 输入与 text 输出并完成一轮 ASR commit
- **THEN** 服务端 MUST 下发思考与 answer，MUST NOT 下发 TTS 音频帧

### Requirement: 文模式 start 仍校验音频元数据

纯 text 输入时，`start` MUST 仍要求与现网一致的音频元数据字段（含 `sampleRate`、`bits`、`channels`、`length`、`mode=stream` 等既有必填项）校验通过；服务端 MUST NOT 因文模式放宽或省略这些字段的必填校验。客户端 MAY 传入与横屏一致的占位数值。

#### Scenario: 缺 sampleRate 的文模式 start 被拒

- **WHEN** 客户端声明 text 输入但 `start` 缺少合法 `sampleRate`（或既有必填音频字段）
- **THEN** 服务端 MUST 拒绝该 `start`（错误帧或等价失败），MUST NOT 进入 `started` 可对话状态

### Requirement: 喂养对话经硬件特权与 VU 正式模

经 `/voice/chat/ws` 触发的喂养意图 LLM MUST 标记硬件特权，MUST 使用 `voiceUnderstanding` lane 的正式（非 free）模型配置，MUST 经该 lane 并发闸门 `Acquire`。该路径 MUST NOT 因 `voice_ai` 月度额度用尽而拒绝对话或改走 free 模型。

#### Scenario: 不计 voice_ai 次

- **WHEN** 一轮 WS 喂养意图 LLM 成功完成（含文模式）
- **THEN** 系统 MUST NOT 对该轮执行 `voice_ai` 成功计次（consume 为 no-op 或不调用）

---

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

当意图结果 `need_confirm=false`（或不需要确认）时，voice MUST 清除该设备本地 `conversation_id`，随后 MUST：`target_type=conversation` 或空 target 仅透传 Python 自然语言回复；喂养/多事件写库 MUST 已由 Python 在同轮意图分析中经 history `event/batch` 完成，Go `applyUnifiedIntentResult` MUST NOT 再执行 Go 侧事件 CRUD。

#### Scenario: 澄清后落地喂养

- **WHEN** 续聊轮 Python 返回 `need_confirm=false` 且 `target_type=feeding` 且 batch 已成功写库
- **THEN** voice SHALL 清除本地 `conversation_id`
- **AND** SHALL 透传 Python 返回的回复文本

#### Scenario: 澄清后闲聊回复

- **WHEN** 续聊轮 Python 返回 `need_confirm=false` 且目标为 conversation 或空 target
- **THEN** voice SHALL 清除本地 `conversation_id`
- **AND** SHALL 仅回复自然语言，不走 Go 侧喂养落库

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

---

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

---

### Requirement: Voice 与 Device 领域逻辑必须运行在独立服务中

系统 SHALL 将 voice 领域与 device 领域的业务逻辑部署到独立可部署服务，并 SHALL 定义 gateway 调用所需的明确内部服务契约。

#### Scenario: Voice 请求路由
- **WHEN** gateway 收到 voice 领域请求
- **THEN** gateway SHALL 按既定内部契约将请求路由到 `voice-service`，而不是在本地执行 voice 业务逻辑

#### Scenario: Device 请求路由
- **WHEN** gateway 收到 device 领域请求
- **THEN** gateway SHALL 按既定内部契约将请求路由到 `device-service`，而不是在本地执行 device 业务逻辑

### Requirement: 服务边界遵循领域数据归属

系统 SHALL 按当前数据库/领域归属划分服务边界，并 SHALL 通过显式服务接口处理跨领域访问。

#### Scenario: Voice 流程需要 Device 领域数据
- **WHEN** `voice-service` 需要访问 device 领域数据
- **THEN** `voice-service` SHALL 通过契约化内部 API 或事件交互获取数据，而不是直接嵌入 device 领域实现

---

### Requirement: 对话 WebSocket SHALL 使用百炼流式 STT（chat profile）

`voice-service` 在处理 `/voice/chat/ws` 且输入模态为 `audio` 时，流式 STT MUST 使用 **chat profile** 配置（`sttChat`），provider MUST 为 `dashscope`，默认模型 MUST 为 `qwen-audio-3.0-asr-flash-streaming`。MUST NOT 在该路径默认使用百度 `dev_pid=1537` 流式 ASR。

#### Scenario: 对话 WS 建连后创建 DashScope 流式 ASR

- **WHEN** 客户端对 `/voice/chat/ws` 发送合法 `start` 且 `inputModality=audio`（或省略默认为 audio）
- **AND** `sttChat.provider=dashscope` 且 `sttChat.streamEnabled=true`
- **AND** 有效 DashScope API Key 与 Workspace ID 已配置
- **THEN** 服务端 MUST 通过百炼 WebSocket 建立流式 ASR 会话
- **AND** 服务端 MUST NOT 为该会话调用百度流式 ASR 建连

#### Scenario: DashScope 凭证缺失

- **WHEN** 客户端对 `/voice/chat/ws` 开始 audio 会话
- **AND** chat profile 的 API Key 或 Workspace ID 未配置
- **THEN** 服务端 MUST 返回 `error` 帧且 `stage=stt`
- **AND** MUST NOT 静默降级为无 ASR 继续对话

### Requirement: CreateStreamASRSession SHALL 支持 profile 分流

`Voice().CreateStreamASRSession`（及 `VoiceContract` 等价接口）MUST 接受 `profile` 参数，取值 `chat` 或 `dictation`，并 MUST 根据 profile 选择 `sttChat` 或 `sttDictation` 配置块；实际引擎 MUST 由所选配置块的 `provider` 决定。

#### Scenario: chat 与 dictation 均按各自配置 provider 建连

- **WHEN** `voice_ws.go` 调用 `CreateStreamASRSession(..., profile=chat, ...)` 且 `sttChat.provider=dashscope`
- **THEN** 实现 MUST 读取 `sttChat` 并创建 DashScope 会话
- **WHEN** `voice_asr_ws.go` 调用 `CreateStreamASRSession(..., profile=dictation, ...)` 且 `sttDictation.provider=dashscope`
- **THEN** 实现 MUST 读取 `sttDictation` 并创建 DashScope 会话

#### Scenario: DashScope 降级百度使用 legacy stt

- **WHEN** 所选 profile 的 `provider=dashscope` 且建连失败
- **AND** 该配置块 `fallbackProvider=baidu`
- **THEN** 实现 MUST 使用 legacy `voiceChat.stt`（百度）配置创建百度流式会话
- **AND** MUST NOT 使用已为 dashscope 的 `sttDictation` 作为百度会话配置源

### Requirement: DashScope 流式 ASR SHALL 兼容现有 PCM 参数

百炼流式 ASR 实现 MUST 接受与现有链路一致的音频元数据：`sampleRate=16000`（或 `start` 声明值）、`bits=16`、`channels=1`、PCM 小端二进制分片。`WriteAudio` / `Commit` / `Close` 语义 MUST 与 `StreamASRSession` 契约一致，以便 `/voice/chat/ws` 无需修改客户端。

#### Scenario: 对话 WS 收到 asr_partial

- **WHEN** DashScope 返回中间识别结果且文本非空
- **THEN** `/voice/chat/ws` MUST 向客户端发送 `{"type":"asr_partial","code":0,"text":"..."}`（与现有行为一致）

#### Scenario: 对话 WS commit 后 asr_final

- **WHEN** 服务端对 DashScope 会话执行 `Commit` 且得到有效转写文本
- **THEN** `/voice/chat/ws` MUST 继续现有对话链路（LLM/TTS），且 MUST 下发 `asr_final` 或等价最终文本事件（与现实现一致）

### Requirement: 对话 STT 配置 SHALL 位于 voice-chat.shared.yaml

对话 STT 配置 MUST 位于 `manifest/config/voice-chat.shared.yaml` 的 `voiceChat.sttChat`（或经 `GF_VOICE_CHAT_FILE` 加载的等价文件）。MUST NOT 将 DashScope STT 专属字段回流到 `manifest/config/config.yaml` 主网关配置。

#### Scenario: voice-service 加载 sttChat

- **WHEN** `voice-service` 启动并加载 `voice-chat.shared.yaml`
- **THEN** `sttChat.provider`、`sttChat.model`、`sttChat.streamEnabled` MUST 可供 `CreateStreamASRSession(chat)` 读取

### Requirement: 环境变量 SHALL 支持 DashScope 凭证注入

部署 MUST 支持经环境变量注入百炼凭证：`VOICE_DASHSCOPE_API_KEY`（优先）或复用 `UCG_DASHSCOPE_API_KEY`；`DASHSCOPE_WORKSPACE_ID` 用于 WebSocket 端点。`manifest/docker/.env.example` MUST 文档化上述变量。

#### Scenario: 仅配置 UCG 密钥

- **WHEN** `VOICE_DASHSCOPE_API_KEY` 未设置且 `UCG_DASHSCOPE_API_KEY` 已设置
- **AND** `sttChat.apiKey` 配置为空
- **THEN** chat profile STT MUST 使用 `UCG_DASHSCOPE_API_KEY` 鉴权

### Requirement: 听写 WebSocket SHALL 使用百炼流式 STT（dictation profile）

`/voice/asr/ws` 流式 STT MUST 使用 **dictation profile** 配置（`sttDictation`；未配置时回退 legacy `stt` 块的既有规则仍适用），默认 provider MUST 为 `dashscope`，默认模型 MUST 为 `qwen-audio-3.0-asr-flash-streaming`（与 chat 路一致）。本变更 MUST NOT 改变听写端点的 WS 协议与事件类型（`asr_partial` / `asr_final` 由 client `commit`/`end` 触发定稿等）。

#### Scenario: 听写 WS 默认走 DashScope

- **WHEN** 客户端对 `/voice/asr/ws` 发送合法 `start`
- **AND** `sttDictation.provider=dashscope` 且 `streamEnabled=true`
- **AND** 有效 DashScope API Key 与 Workspace ID 已配置
- **THEN** 服务端 MUST 通过百炼 WebSocket 建立流式 ASR 会话（`CreateStreamASRSession` dictation profile）
- **AND** 下行 MUST 仍支持 `asr_partial`、`asr_final`（由 client commit/end 触发定稿）
- **AND** 服务端 MUST NOT 为该默认配置调用百度流式 ASR 建连

#### Scenario: 听写可配置回滚百度

- **WHEN** 运维将 `sttDictation.provider` 设为 `baidu` 且百度流式凭证可用
- **THEN** `CreateStreamASRSession(dictation)` MUST 创建百度流式 ASR 会话
- **AND** 听写 WS 协议 MUST 保持不变

### Requirement: 听写 STT 配置 SHALL 位于 voice-chat.shared.yaml 的 sttDictation

听写 STT 配置 MUST 位于 `manifest/config/voice-chat.shared.yaml` 的 `voiceChat.sttDictation`（或经 `GF_VOICE_CHAT_FILE` 加载的等价文件）。默认仓库配置 MUST 将 `sttDictation.provider` 设为 `dashscope`。MUST NOT 将 DashScope STT 专属字段回流到 `manifest/config/config.yaml` 主网关配置。

#### Scenario: voice-service 加载 sttDictation

- **WHEN** `voice-service` 启动并加载 `voice-chat.shared.yaml`
- **THEN** `sttDictation.provider`、`sttDictation.model`、`sttDictation.streamEnabled` MUST 可供 `CreateStreamASRSession(dictation)` 读取
- **AND** 默认 `sttDictation.provider` MUST 为 `dashscope`

---

### Requirement: Voice MUST 经 HTTP 访问 device 领域持久化数据

在 **voice-service** 进程内，凡涉及 `user`、`event`、`action` 等他域表的读取或写入（含语音意图、DeepSeek 实体抽取、动作词典维护等），MUST 通过 **device-service** 暴露的 HTTP 接口完成；MUST NOT 在 `voice` 包或 voice 进程内调用 `device` 包中会触发他域 `dao.User`、`dao.Event`、`dao.Action` 的实现路径。

#### Scenario: 语音链路查询事件列表

- **WHEN** voice 需要加载事件字典以匹配用户说法
- **THEN** voice MUST 向 device-service 发起 HTTP 请求获取列表，MUST NOT 使用本进程 default 数据库连接访问 `event` 表

#### Scenario: 语音链路写入动作或事件

- **WHEN** voice 需要将新动作或事件变更持久化
- **THEN** voice MUST 调用 device-service HTTP 接口完成写入，MUST NOT 在 voice 进程内执行 `dao.Action` 或 `dao.Event` 的 Insert/Update

### Requirement: 迁移期 local 路径 MUST 仍为 HTTP 到 device 入口

若配置为「本地」模式以简化联调，其语义 MUST 为调用 **本机或可解析的 device-service 基址**（如 `http://127.0.0.1:9803`）的 HTTP，MUST NOT 解释为在同一进程内直接执行他域 DAO。

#### Scenario: 开发单机多端口

- **WHEN** 开发者在同一主机分别启动 voice 与 device 监听不同端口
- **THEN** voice 的 local 配置 MUST 仍指向 device HTTP 基址，而非共享同一 ORM 连接访问 device 库

---

### Requirement: 事件匹配时子节点优先于父节点

voice-service 在本地文本匹配事件时，SHALL 对候选事件按**深度降序、名称长度降序**排序后再匹配，使叶子或深层子节点优先于浅层父节点命中。

#### Scenario: 同时提及父名与子名时命中叶子

- **WHEN** 事件树含「换尿布(父)」与「大便(子)」
- **AND** 用户说「换尿布，拉了大便」
- **THEN** 匹配结果 SHALL 为「大便」事件 id
- **AND** SHALL NOT 进入父节点追问流程

---

### Requirement: voice-service 提供 internal 文本对话 API

voice-service SHALL 提供 `POST /voice/internal/api/text/chat`，接受 JSON body `deviceNo`、`transcript`； 调用方 MUST 在 Header 携带 `X-Device-Gateway-Internal-Secret`（与 `DEVICE_GATEWAY_INTERNAL_SECRET` 一致）。 成功时响应 data MUST 含 `reply` 字符串。 失败时 MUST 使用与 App 网关一致的 business code（含 40301 未登录、40302 额度用尽）。 对话能力 MUST 统一走母婴喂养模式，MUST NOT 存在闲聊模式切换。 AI 推理 MUST 统一调用 Python 微服务接口（`/v1/analyze/intent`），MUST NOT 在 Go 侧保留 DeepSeek/LLM 直连兜底。 Python 服务不可用时，MUST 返回固定降级提示语"AI 服务暂时不可用，请稍后再试"。

#### Scenario: 合法 internal 请求成功

- **WHEN** secret 正确且 `deviceNo`、`transcript` 合法且额度允许
- **THEN** 接口 MUST 返回 `code=0` 且 `reply` 为非空或允许的空串回复

#### Scenario: 未登录 wxId

- **WHEN** 未携带有效 `X-Internal-Wx-Id` 且链路需要额度预检
- **THEN** 接口 MUST 返回 business code 40301

#### Scenario: 额度用尽

- **WHEN** wxId 有效且当月 voice_ai 额度已用尽
- **THEN** 接口 MUST 返回 business code 40302

#### Scenario: secret 无效

- **WHEN** `X-Device-Gateway-Internal-Secret` 缺失或错误
- **THEN** 接口 MUST 拒绝请求且 MUST NOT 执行 TextChat

#### Scenario: Python 微服务不可用

- **WHEN** Python 微服务（`/v1/analyze/intent`）调用失败（超时、连接失败、5xx 等）
- **THEN** 接口 MUST 返回固定降级提示语"AI 服务暂时不可用，请稍后再试"
- **AND** MUST NOT 回退到 DeepSeek/LLM 直连

#### Scenario: 闲聊模式切换命令

- **WHEN** 用户输入包含"切换到闲聊"等闲聊模式切换语义
- **THEN** 系统 MUST 按母婴喂养模式处理
- **AND** MUST NOT 切换到闲聊模式
- **AND** MUST NOT 返回闲聊相关回复

---

### Requirement: Voice-service SHALL 提供实时听写 WebSocket 端点

`voice-service` MUST 在路径 `/voice/asr/ws` 提供 WebSocket 服务，将客户端上行的 PCM 音频流送入 **dictation profile** 所配置的流式 STT（默认 `sttDictation.provider=dashscope`），并将识别出的中文文本实时返回给客户端。

#### Scenario: 握手成功并开始会话

- **WHEN** 客户端对 `/voice/asr/ws` 发起 WebSocket Upgrade 且握手成功
- **AND** 客户端发送合法 `start` 文本帧（含非空 `deviceNo` 与有效 `sampleRate`/`bits`/`channels`）
- **THEN** 服务端 MUST 回复 `{"type":"started","code":0,"mode":"stream"}`（或等价字段集）
- **AND** 服务端 MUST NOT 调用对话 LLM、TTS 或设备最近对话落库接口

#### Scenario: 流式 STT 未启用或不可用

- **WHEN** `start` 成功但流式 STT 配置不可用（如 `sttDictation.streamEnabled=false`、provider 不支持，或 DashScope 凭证缺失导致建连失败）
- **THEN** 服务端 MUST 发送 `{"type":"error","code":1,"stage":"stt",...}` 且 MUST NOT 假装听写成功

### Requirement: 上行协议 SHALL 限定为听写所需子集

听写 WebSocket MUST 接受以下上行消息类型：

- Text JSON：`type` 为 `start`（开始会话）
- Binary：16-bit 小端 PCM 音频分片，参数与 `start` 声明一致
- Text JSON：`type` 为 `commit`（**一句听写结束**，触发 finalize 并下发 `asr_final`）
- Text JSON：`type` 为 `end`（结束当前听写 WebSocket 会话）
- Text：心跳 `ping`（服务端回复 `pong`）

#### Scenario: 客户端发送 commit

- **WHEN** 客户端在已 `start` 的会话中发送 `{"type":"commit"}`
- **AND** 当前已有流式 ASR 会话或已接收过非空音频缓冲
- **THEN** 服务端 MUST 对当前句执行 STT finalize 并下发 `{"type":"asr_final",...,"source":"client"}`
- **AND** 服务端 MUST NOT 因此进入对话/TTS 链路

#### Scenario: commit 前无音频

- **WHEN** 客户端发送 `{"type":"commit"}` 且当前无 ASR 会话且无已缓冲音频
- **THEN** 服务端 MUST 返回 `error`（如 `stage=validate`）

#### Scenario: 客户端发送 end

- **WHEN** 客户端在已 `start` 的会话中发送 `{"type":"end"}`
- **THEN** 服务端 MUST 关闭当前流式 ASR 会话并回复 `{"type":"ended","code":0}`
- **AND** 若关闭前仍有未 finalize 的音频，服务端 MAY 先执行 finalize 并下发 `asr_final`（`source` 为 `end`）

### Requirement: 下行协议 SHALL 仅包含听写相关事件

服务端下行 Text JSON MUST 以听写为主，至少包含：

- `asr_partial`：非空中间识别文本；**亦包含**流式 STT 引擎级 final 回调转发（与中间 partial 同为该类型，供客户端覆盖预览）
- `asr_final`：一句听写定稿文本（**仅**由客户端 `commit` 或 `end` 触发的 finalize 产生）
- `asr_no_result`：finalize 后无有效文本时
- `error`、`started`、`ended`

服务端 MUST NOT 在该端点下发 `audio_chunk`、`chat_delta`、`exit` 或 TTS 相关字段。

#### Scenario: 收到 ASR 中间结果

- **WHEN** 流式 STT 产生新的中间文本且与上次 partial 不同
- **THEN** 服务端 MUST 发送 `{"type":"asr_partial","code":0,"text":"<识别文本>"}`

#### Scenario: 收到 ASR 最终结果

- **WHEN** 客户端发送 `commit` 或 `end` 导致服务端执行 finalize 且得到有效转写文本
- **THEN** 服务端 MUST 发送 `{"type":"asr_final","code":0,"text":"<识别文本>"}` 且 `source` MUST 为 `client` 或 `end`

### Requirement: 听写连接 SHALL 与对话连接隔离

`voice-service` 在处理 `/voice/asr/ws` 时 MUST NOT 将连接注册到用于 `/voice/chat/ws` 的「单设备单连接」替换管理器（`VoiceWSManager`），以避免听写页与对话页互相踢连接。

#### Scenario: 同一 deviceNo 同时存在 chat 与 asr 连接

- **WHEN** 设备 `device-001` 已建立 `/voice/chat/ws` 连接且另建 `/voice/asr/ws` 连接
- **THEN** 两条连接 MUST 均可保持，直至各自关闭

### Requirement: Voice 域边界 SHALL 保持不变

听写实现 MUST 仅使用 voice 域已有 STT 能力与配置（`voice-chat.shared.yaml` / `Voice().CreateStreamASRSession`），且 MUST NOT 在 voice-service 内直接访问 device/history/user 等他域数据库表。

#### Scenario: 听写会话不查 device 库

- **WHEN** 客户端在 `start` 中提供 `deviceNo`
- **THEN** 服务端 MAY 将其用于日志与限流键，且 MUST NOT 为听写路径新增对 device 表 DAO 的依赖

### Requirement: 听写 WS MUST NOT 服务端主动截句

`/voice/asr/ws` 实现 MUST NOT 基于服务端静音计时、STT 回调间隔或无回调时长自动调用 STT finalize 或向客户端发送 `asr_final`。

#### Scenario: 长时间静音但未发送 commit

- **WHEN** 客户端已 `start` 并持续发送二进制音频或保持连接
- **AND** 客户端在超过 2 秒内未发送 `commit` 或 `end`
- **AND** 流式 STT 未产生新的 partial 或产生引擎 final 回调
- **THEN** 服务端 MUST NOT 因静音或超时自动下发 `asr_final`
- **AND** 服务端 MUST NOT 因上述原因自动关闭并重建 ASR 会话

#### Scenario: 引擎 onFinal 回调

- **WHEN** 流式 STT 提供商推送引擎级 final 结果（如百度 `FIN_TEXT`）且文本非空
- **THEN** 服务端 MUST NOT 将该结果作为 `asr_final` 下发给客户端
- **AND** 服务端 MUST NOT 仅因该回调而关闭当前 ASR 会话
- **AND** 若该文本与上次已下发的 `asr_partial` 的 `text` 不同，服务端 MUST 再发送一条 `{"type":"asr_partial","code":0,"text":"<识别文本>"}` 以供客户端更新预览

#### Scenario: 禁止的 asr_final source

- **WHEN** 服务端在听写 WS 上产生 `asr_final`
- **THEN** `source` MUST NOT 为 `silence`、`auto_commit` 或 `asr_callback`

---

### Requirement: Gateway MUST 为 voice 路由提供独立可配置代理能力

gateway MUST 以独立中间件管理 `/voice/text/*` 路由，并支持 `local|proxy|canary` 三态。

#### Scenario: voice 路由进入 local 模式
- **WHEN** `VOICE_API_ROUTE_MODE=local`
- **THEN** gateway MUST 执行本地处理链路，且 MUST NOT 将请求转发到 voice-service

#### Scenario: voice 路由进入 proxy 模式
- **WHEN** `VOICE_API_ROUTE_MODE=proxy` 且 `VOICE_API_PROXY_URL` 可用
- **THEN** gateway MUST 将 `/voice/text/*` 请求全量转发到 voice-service

#### Scenario: voice 路由进入 canary 模式
- **WHEN** `VOICE_API_ROUTE_MODE=canary` 且配置了 `VOICE_API_PROXY_CANARY_PERCENT`
- **THEN** gateway MUST 按稳定分流键执行百分比转发，其余请求保持本地处理

### Requirement: voice canary 分流 MUST 保持同键稳定

gateway MUST 采用稳定分流键（如 deviceNo）对 canary 流量做无状态一致性计算。

#### Scenario: 同一分流键连续请求
- **WHEN** 同一设备在 canary 模式下发起多次 `/voice/text/*` 请求
- **THEN** 请求 MUST 稳定命中同一流量路径（proxy 或 local）

---

### Requirement: TextChat SHALL 不依赖 voice.task 事件链路

TextChat 对话完成路径 MUST NOT 发布或依赖 `voice.task.requested` 事件；worker-service consumer 已删除，对话持久化与业务逻辑 MUST 在 voice-service 请求内完成。

#### Scenario: 对话完成无 task 发布

- **WHEN** TextChat 会话正常结束
- **THEN** 系统 MUST NOT 调用 `publishTaskRequested` 或等价 MQ 发布

---

### Requirement: 喂养语音理解 MUST 直调 Python 且不传选模

voice-service 在喂养语音理解（统一意图 / AnalyzeIntent 及 stream）路径上 MUST 调用 Python `POST /v1/analyze/intent` 或 `/v1/analyze/intent/stream`，请求体 MUST NOT 包含 `model`（或等价 provider/name 选模对象）。Go MUST NOT 再经 `llm_lane_config` / `ResolveLaneModel` / VIP / AI 额度决定理解所用模型。

#### Scenario: intent 请求无 model 字段

- **WHEN** 已登录用户触发喂养理解并调用 Python intent
- **THEN** 发往 Python 的 JSON MUST 省略 `model`，且 Python 侧 MUST 仍能完成意图分析（定死 LLM）

### Requirement: 理解路径 MUST NOT 依赖 AI 额度门禁

理解路径 MUST NOT 因 `voice_ai` 月度额度用尽而返回 40302，MUST NOT 在调用前对理解做 premium/free omit 分轨。登录/设备鉴权失败等既有非额度错误语义 MAY 保留。

#### Scenario: 无额度表时理解仍可用

- **WHEN** voice AI 额度表与检查逻辑已删除且用户合法进入聊天
- **THEN** 理解调用 MUST 仍可到达 Python intent，MUST NOT 因缺失额度行而失败

### Requirement: 理解编排与 ASR/TTS 契约 MUST 保留

系统 MUST 保留聊天 WS/MCP 入口、理解编排（澄清/确认/reply 透传）及理解所依赖的 ASR/TTS/历史上下文路径。删除选模 MUST NOT 删除 `PythonAIClient` 的 intent 调用能力。

#### Scenario: 聊天主路径仍返回 reply

- **WHEN** 用户通过既有喂养聊天通道发送可理解文本
- **THEN** 系统 MUST 仍能返回基于 Python intent 的 reply（或既有规则/确认流），MUST NOT 因删除选模而整段聊天不可用

---
