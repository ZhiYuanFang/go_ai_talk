## Context

当前语音球喂养流式入口 `HandleTranscriptForIntentStream`（`voice_chat.go`）在完成 `callDeepSeekUnifiedIntentStream`（Python `/v1/analyze/intent/stream`）后，仍调用 `chatWithResult`。而 `chatWithResult`（`voice_chat_understanding.go`）在常规路径上会再执行 `callDeepSeekUnifiedIntent` → 非流式 `AnalyzeIntent`，形成：

1. **双倍 Python 成本**：同一句 transcript 触发两次意图分析。
2. **UI / 落库不一致**：Stream 推给 UI 的 thinking/answer 与第二次 AnalyzeIntent 落库结果可能不同。
3. **confirm 语义漂移**：`NeedConfirm` / `conversation_id` / pending 可能来自第二次调用，与 UI 已展示内容脱节。

依赖前提（本变更引用、不扩 scope）：

- `confirm-ws-adaptation`：已落地 `NeedConfirm`、pending confirm、`ConfirmIntent`、`parseConfirmFeedback`。
- `fix-python-api-alignment`：已落地 `AnalyzeIntentStream`、`mapPythonRespToIntent`、`HandleTranscriptForIntentStream`；其 `streaming-intent-service` 已要求「流式与非流式同落库逻辑」，但实现仍二次调用。

范围仅 `internal/services/voice`；voice 不直连他域 DAO。

## Goals / Non-Goals

**Goals:**

- Stream 结束后，用 `streamRes.Result`（经 `mapPythonRespToIntent`）走与 `chatWithResult` **相同**的 confirm / 落库 / 回复行为矩阵。
- 流式落地全路径 **禁止**再调非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`；单次 transcript 仅一次 Python intent 调用（pending confirm 走 `ConfirmIntent` 的情况除外，见行为矩阵）。
- 保留：`NeedConfirm` / pending confirm / quota 单次 consume / multi-event / `handleUnifiedIntentAction`。
- 日志可观测：能区分「流式落地」且可核对仅一次 Python intent。

**Non-Goals:**

- 不改 gateway、Flutter 协议、tip/feedback。
- 不改 TTS 非流式 `HandleTranscriptForStreaming` 的 AnalyzeIntent 路径。
- 不修改 Python 侧接口契约。
- 不新增测试文件。
- 不扩展 `confirm-ws-adaptation` / `fix-python-api-alignment` 的变更目录内容。

## Decisions

### 1. 抽取「已有意图 → 落库/回复」公共路径

**决策**：从 `chatWithResult` 中抽出「拿到 `deepSeekUnifiedIntent` 之后」的处理为可复用函数（命名实现期自定，下文称 `applyUnifiedIntentResult`），入参为已映射的 intent + 规范化 transcript + 事件列表等；流式入口与非流式 `chatWithResult` 均调用它。

**理由**：避免在 `HandleTranscriptForIntentStream` 复制 NeedConfirm / conversation / action 分支；保证行为矩阵单一真源。

**备选**：仅让 `chatWithResult` 接受 optional 预解析 intent 参数 → 耦合入口形态，不如显式抽取清晰。

### 2. streamRes → 统一意图结构的映射

**决策**：Stream 结束后使用已有 `mapPythonRespToIntent(streamRes.Result)`，与非流式 `callDeepSeekUnifiedIntent` 成功路径一致；`streamRes.Result == nil` 视为意图解析失败，返回降级话术，**不得**回退调用非流式 AnalyzeIntent。

**映射字段**（与现网一致）：`TargetType` / `Action` / `EventName` / `EventId` / `EventType` / `EventUnit` / `Quantity` / `IsNewEvent` / `Content→Reply` / `NeedConfirm` / `ConfirmMessage` / `ConversationID` / `Events`。

### 3. 流式入口编排顺序（与 chatWithResult 对齐的 preamble）

**决策**：`HandleTranscriptForIntentStream` 在发起 Stream 之前复用与 `chatWithResult` 相同的前置分支（逻辑可抽取为共享 preamble）：

1. 规范化文本校验  
2. **pending confirm**：命中 confirm/reject → `ConfirmIntent` → `handleUnifiedIntentAction`（或 conversation 直回）；**不**发起 AnalyzeIntentStream  
3. **pending child event**：继续子事件匹配；不发起 Stream  
4. 否则：`guardVoiceAIQuota` → **一次** `AnalyzeIntentStream`（带 UI callback）→ `mapPythonRespToIntent` → `consumeVoiceAIQuotaOnSuccess`（成功时）→ `applyUnifiedIntentResult`

**理由**：若先 Stream 再进 `chatWithResult`，pending confirm 轮次会白白多一次 Python stream，且可能污染 confirm 语义。

### 4. 行为矩阵（必须保留）

| 分支 | 条件 | 行为 | Python 调用 |
|------|------|------|-------------|
| pending confirm 命中 | 设备有 pending 且反馈可解析为 confirm/reject | `ConfirmIntent` → 映射 → conversation 或 `handleUnifiedIntentAction` | 仅 ConfirmIntent |
| pending confirm 未识别 | 有 pending 但反馈为空 | clear pending，落入常规流式意图 | 随后一次 Stream |
| pending child | 有待选子事件 | `continuePendingChildEvent` | 无 intent |
| NeedConfirm | stream/intent 返回 `NeedConfirm=true` | set pending + 返回 ConfirmMessage/Reply | 已发生的一次 Stream（或非流式） |
| conversation / 空 target | TargetType 为 conversation 或空 | insertQa + 返回 Reply | 已发生的一次 |
| multi-event | `Action=="multi"` 且 Events 非空 | `handleMultiEventIntent` | 已发生的一次 |
| 普通动作 | 其他 | 构造 Action → `handleUnifiedIntentAction` | 已发生的一次 |
| Stream/解析失败 | stream 错误或 Result nil | 降级话术；**禁止**二次 AnalyzeIntent | 0 或失败的一次 Stream |
| quota | guard 失败 | 直接错误返回，不调 Python | 无 |

Quota：**单次** consume——仅在本次成功拿到可落地意图（Stream 或非流式 AnalyzeIntent 成功）后 `consumeVoiceAIQuotaOnSuccess`；pending confirm 恢复路径不重复 consume intent quota（与现网 confirm 分支一致：confirm 分支在 quota 之前）。

### 5. chatWithResult 常规路径保持非流式 AnalyzeIntent

**决策**：TTS / 非流式入口仍走 `chatWithResult` → `callDeepSeekUnifiedIntent`；仅流式入口禁止二次调用。抽取公共落库函数后，两边共享 apply 逻辑。

### 6. 可观测性

**决策**：流式落地成功路径打结构化日志，至少包含：`deviceNo`、路径标记（如 `intent_path=stream_land`）、`target_type`、`action`、`need_confirm`；警告级记录「未再调用非流式 AnalyzeIntent」不是必须逐次打印，但实现评审应以代码路径保证零二次调用。失败路径保留现有 Warning。

### 7. 服务边界

设备事件列表等继续经 `DeviceAdmin().ListEvents`；禁止 voice import 他域 DAO。

## Risks / Trade-offs

- **[Risk] preamble 抽取遗漏导致流式与非流式行为漂移** → Mitigation：任务清单强制对照行为矩阵逐项勾选；优先抽取共用函数而非复制粘贴。
- **[Risk] stream Result 缺字段导致落库失败率上升** → Mitigation：依赖 `fix-python-api-alignment` 已对齐流式/非流式响应结构；Result nil 明确失败语义，不 silent fallback 到非流式。
- **[Risk] 先改入口顺序影响 pending confirm 体验** → Mitigation：Stream 前处理 pending，与现 `chatWithResult` 优先级一致。
- **[Trade-off] 流式路径不再“借” chatWithResult 整函数** → 短期多一个编排函数，长期消除二次调用与语义漂移，可接受。

## Migration Plan

- 纯 voice 进程内行为修复；无 DB / API 版本迁移。
- 部署：随 voice-service 发布即可；回滚为恢复「Stream + chatWithResult」旧编排（会恢复双调用，仅作紧急回退）。
- 无需配置开关（本变更默认修正错误路径）。

## Open Questions

- 无阻塞问题。若 apply 阶段发现 pending child 与流式入口调用方（MCP / text chat stream）交互频率极低，仍按矩阵完整保留，不做裁剪。
