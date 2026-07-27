## 1. Python 客户端契约

- [x] 1.1 在 `AnalyzeIntentRequest` 增加可选 `conversation_id`（`json:"conversation_id,omitempty"`），补充中文注释说明续聊语义
- [x] 1.2 删除 `ConfirmIntentRequest` / `ConfirmIntentResponse` / `ConfirmIntent` 方法及对 `/v1/analyze/intent/confirm` 的引用
- [x] 1.3 更新 `AnalyzeIntentResponse` 中 `ConversationID` / `NeedConfirm` 相关注释（改为同一 intent 续聊，不再指向 confirm 接口）

## 2. 本地 cid 状态

- [x] 2.1 瘦身 `voice_confirm_pending.go`：保留按 `deviceNo` 的 cid 存取与超时懒清理；删除 `parseConfirmFeedback`
- [x] 2.2 调整条目字段与中文注释（核心为 `ConversationID`；可观测字段按需保留）

## 3. 统一调用与落地

- [x] 3.1 从 `prepareChatPreamble` 移除整个 pending confirm / ConfirmIntent 分支；保留文本校验与 pending child
- [x] 3.2 `callDeepSeekUnifiedIntent`：若本地有 cid 则写入请求；失败不 clear cid
- [x] 3.3 `callDeepSeekUnifiedIntentStream`：同样附带 cid（与非流式一致）
- [x] 3.4 `applyUnifiedIntentResult`：`NeedConfirm=true` → set cid + 透传 `ConfirmMessage`/`Reply`，不 CRUD；否则 clear cid 后走 conversation / `handleUnifiedIntentAction`
- [x] 3.5 确认成长建议/历史问答等独立 `AnalyzeIntent` 调用不读取喂养澄清 cid

## 4. 注释与可观测

- [x] 4.1 更新 `chatWithResult` / `HandleTranscriptForIntentStream` / preamble 相关中文注释与日志（去掉 ConfirmIntent 表述；可保留 `need_confirm` / `conversation_id` 日志字段）
- [x] 4.2 全仓 grep：`ConfirmIntent`、`parseConfirmFeedback`、`/intent/confirm`、`user_feedback` 在 voice 意图路径应为 0

## 5. 澄清续聊免计 AI quota

- [x] 5.1 在 `chatWithResult`：若 `pendingConversationID(deviceNo) != ""`，跳过 `guardVoiceAIQuota` 与成功后的 `consumeVoiceAIQuotaOnSuccess`；无 cid 时保持现网计次（含首次 need_confirm）
- [x] 5.2 在 `HandleTranscriptForIntentStream` 使用与非流式同一免计判定；更新相关中文注释/日志（标明澄清续聊免计）
- [x] 5.3 核对：额度用尽 + 有 cid 时可续聊；cid 超时后恢复计次；离题当新意图的本轮仍免计

## 6. 自检

- [x] 6.1 按 spec 行为矩阵口头/笔记核对：冷启动、NeedConfirm 透传、带 cid 续聊落库、清 cid、流式与非流式一致、独立调用不串 cid、澄清续聊免计
- [x] 6.2 确认未新增测试文件；未改 gateway-app 路由 / usage skip / Redis 键
