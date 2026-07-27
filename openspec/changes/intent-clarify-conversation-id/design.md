## Context

`python_ai_talk` 的 `parent-event-disambiguation` 已删除 `/v1/analyze/intent/confirm`，澄清改为同一 `/intent`（及 `/intent/stream`）+ 可选 `conversation_id` + 自由文本；pending 解析（子名/序号/拒绝/叶子肯定词/离题当新意图）在 Python 侧完成。

本仓 `confirm-ws-adaptation` 仍保留：

- `ConfirmIntent` → 已死接口
- `parseConfirmFeedback` → 仅 `confirm|reject`
- `prepareChatPreamble` 在 pending 时短路、不发 intent
- 流式落地规格仍要求 pending 轮走 `ConfirmIntent`

目标：Go 降为「按设备记 cid + 透传自然语言 + 原有事件 CRUD」，流式与非流式同一行为矩阵。

约束：服务边界不变（voice → Python HTTP）；不新增 Redis/后台任务；不改 gateway-app 对外路由；最少额外逻辑。

## Goals / Non-Goals

**Goals:**

- 请求契约：`AnalyzeIntent` / `AnalyzeIntentStream` 支持可选 `conversation_id`
- 响应：`need_confirm=true` 时仅保存 cid 并回传 Python 自然语言，不落库
- 续聊：下一轮统一 intent 调用附带 cid + 用户原文；删除 confirm 枚举通道
- `need_confirm=false` 时清 cid，再走 conversation 回复或 `handleUnifiedIntentAction` CRUD
- `chatWithResult` 与 `HandleTranscriptForIntentStream` 共用同一套逻辑
- 澄清续聊轮（发起前有有效 cid）免计 AI quota，与旧 ConfirmIntent 对齐

**Non-Goals:**

- 不实现自由文本消歧解析、不维护肯定/否定词表
- 不接 `confirm_type` / `options`（本阶段不改 App UI）
- 不改 Go `pending child`（`event_child_pending`）语义
- 不做跨进程 cid 持久化；不扩 Redis
- 成长建议/历史问答等独立 `AnalyzeIntent` 不串带喂养澄清 cid
- 不以「响应 `need_confirm=true`」作为免计条件（首次提问仍计次）

## Decisions

### 1. Go 只保管 conversation_id，解析全交给 Python

- **选择**：本地 state 按 `deviceNo` 存 `conversation_id`（可保留超时懒清理）；不解析用户话术语义。
- **理由**：与 Python 产品契约一致；避免双端词表漂移；父消歧自由文本只能由 Python 消化。
- **替代**：扩 ConfirmIntent 收自由文本 —— 否决（对端已删通道）。

### 2. 删除 ConfirmIntent / parseConfirmFeedback

- **选择**：删除方法、类型与 preamble 分支；澄清轮不再短路。
- **理由**：死接口必须移除；短路会阻止带 cid 的 intent 调用。
- **替代**：留空壳兼容 —— 否决（无调用方需要）。

### 3. 流式与非流式统一挂载 cid

- **选择**：在 `callDeepSeekUnifiedIntent` 与 `callDeepSeekUnifiedIntentStream` 读取本地 cid 写入请求；落地统一经 `applyUnifiedIntentResult`。
- **理由**：单一真源，避免 stream_land 与 chatWithResult 行为分叉。
- **替代**：仅改 Stream —— 否决（用户明确要求非流式一并改）。

### 4. cid 生命周期

```
need_confirm=true  → set(deviceNo, conversation_id)
need_confirm=false → clear(deviceNo)
调用失败           → 不动（保留以便重试续聊）
```

- **理由**：简单；过期 cid 由 Python 当冷启动，Go 收到非 confirm 结果后再清。
- **替代**：固定 TTL 强制清且不调 Python —— 否决（与「无额外逻辑」冲突，且易丢续聊）。

### 5. NeedConfirm 话术透传

- **选择**：优先 `confirm_message`，否则 `content`（`Reply`）；仍可保留极简空串兜底以免 TTS 空白。
- **理由**：用户要求「自然语言直接返回」；不在 Go 侧改写澄清文案。
- **替代**：拼接 options —— 否决（本阶段不接 options）。

### 6. 澄清续聊轮免计 AI quota

- **选择**：若发起统一意图前本地已有有效 `conversation_id`（`pendingConversationID(deviceNo) != ""`），流式与非流式均 **跳过** `guardVoiceAIQuota` 预检扣减门槛与成功后的 `consumeVoiceAIQuotaOnSuccess`；冷启动（无 cid）仍走常规 guard → 成功 consume（含返回 `need_confirm=true` 的提问轮）。
- **理由**：对齐旧 `ConfirmIntent` 免计次，避免用户被系统追问后额外扣额度；额度用尽时仍可完成澄清，避免半途卡死。
- **替代**：澄清也计次 —— 否决（产品要求免计）；以响应 `need_confirm` 免计 —— 否决（会把首次提问也免掉，与旧语义不符）。
- **边界**：带 cid 但 Python 判离题当新意图的本轮仍免计；cid 超时后下一轮按冷启动计次。

### 7. 与 pending child 的优先级

- **选择**：`prepareChatPreamble` 删除 confirm 分支后，顺序为：文本校验 → **pending child** → Continue（再带 cid 调 intent）。
- **理由**：child pending 是 Go 本域事件树逻辑，与 Python cid 澄清正交；父事件消歧已主要由 Python 接管，child 触达变少但行为保持。
- **替代**：有 cid 时跳过 child —— 暂不采用（需产品确认；默认不增逻辑）。

### 8. pending 存储形态

- **选择**：瘦身现有 `pendingConfirmState`（可改名）为 cid + CreatedAt；删除 EventName/Action 等仅日志字段亦可保留作可观测。
- **理由**：改动面小；超时懒清理可保留（与现网一致）。

## Risks / Trade-offs

- [过期/重启丢失 Go cid 或 Python pending] → 用户再说一遍冷启动；可接受（与旧 MemorySaver 同级）
- [澄清免计被滥用] → 受 60s 超时与 `need_confirm=false` 清 cid 约束；成本主要在 Python 侧
- [action=disambiguate 且 need_confirm 漏标] → 可能误入 CRUD；依赖 Python 契约；评审时核对 NeedConfirm 早退仍在 `applyUnifiedIntentResult` 最前
- [独立 AnalyzeIntent 误带 cid] → 成长建议/历史问答调用点禁止读喂养 pending cid
- [部署顺序] → 必须先/同时上线已删 confirm 的 Python；否则旧 ConfirmIntent 已无意义，本变更上线后澄清才恢复

## Migration Plan

1. 确认运行中的 `python_ai_talk` 已无 `/intent/confirm`，且 intent 请求接受 `conversation_id`。
2. 合入本变更并部署 `voice-service`（及仍内嵌 voice 意图路径的进程）。
3. 回归：叶子直接落库；父消歧 → 回问 NL（计 1 次）→ 自由文本续聊落叶子（免计）；拒绝/离题当新意图；流式与非流式各走一遍；额度用尽时澄清续聊仍可完成。
4. 回滚：回退 voice 二进制即可；不涉及 DB migration。回滚后若 Python 仍无 confirm，澄清仍不可用——回滚不能恢复旧 confirm 契约。

## Open Questions

- 无（澄清续聊免计 quota、不接 options、非流式一并改，已拍板）。
