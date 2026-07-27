## Why

兄弟仓 `python_ai_talk` 已移除 `/v1/analyze/intent/confirm`，改为在同一 `/v1/analyze/intent` 与 `/v1/analyze/intent/stream` 上通过可选 `conversation_id` + 自由文本完成消歧/澄清续聊。本仓仍调用已删除的 confirm 接口，并用 `parseConfirmFeedback` 把用户话术收成 `confirm|reject`，与对端契约断裂，父事件消歧（子名/序号等）也无法工作。需立即对齐，否则澄清轮次全部失败。

## What Changes

- **BREAKING（对内 Python 契约）**：删除 `ConfirmIntent` / `ConfirmIntentRequest` / `ConfirmIntentResponse` 及对 `POST /v1/analyze/intent/confirm` 的调用。
- 删除 `parseConfirmFeedback` 与 preamble 中的 confirm/reject 短路分支。
- `AnalyzeIntentRequest`（流式与非流式共用）新增可选 `conversation_id`；有本地 pending cid 时随请求带回。
- Go 仅按 `deviceNo` 保留/清理 `conversation_id`：`need_confirm=true` 时保存并直接回传 Python 自然语言（`confirm_message`/`content`），不做额外解析；`need_confirm=false` 时清理 cid，再走既有 conversation 回复或事件 CRUD（`handleUnifiedIntentAction`）。
- 流式 `HandleTranscriptForIntentStream` 与非流式 `chatWithResult` **统一**上述逻辑（同一 preamble + 同一 `applyUnifiedIntentResult` 行为矩阵）。
- 不接 `confirm_type` / `options`；不新增澄清专用业务分支。
- 澄清续聊轮（发起意图前本地已有有效 `conversation_id`）走常规 `AnalyzeIntent`/`AnalyzeIntentStream`，但 **MUST 免计** AI quota（对齐旧 `ConfirmIntent` 免计次；冷启动含首次 `need_confirm` 提问仍计次）。

## Capabilities

### New Capabilities

- `intent-clarify-conversation-id`：voice 侧意图澄清续聊契约——按设备保留 `conversation_id`、经统一 intent 接口续聊、透传自然语言、仅在非确认结果上执行事件 CRUD，并删除旧 confirm 枚举通道。

### Modified Capabilities

- （无）主规格树 `openspec/specs/v*` 无独立 confirm 能力条目；本变更以 change-local 新能力规格为准。历史变更 `confirm-ws-adaptation` / `fix-chat-stream-intent-land` 中的 ConfirmIntent 行为由本能力取代。

## Impact

- **代码**：`internal/services/voice/python_ai_client.go`、`voice_confirm_pending.go`、`voice_chat_understanding.go`、`voice_chat.go`（注释/路径说明）；成长建议/历史问答等独立 `AnalyzeIntent` 调用点不得串带喂养澄清 cid。
- **依赖**：要求已部署的 `python_ai_talk` 含 `parent-event-disambiguation` 后的 intent 路由（无 `/intent/confirm`，请求支持 `conversation_id`）。
- **行为**：澄清轮改为自由文本续聊；本地不再做肯定/否定词表；`pending child`（Go 事件树子节点选择）保留，与 Python cid 澄清并存。
- **非范围**：gateway-app 新路由、Redis、后台 ticker、跨进程 pending 持久化、App 选项 UI。
