## Context

现网语音意图链路：`callDeepSeekUnifiedIntent(Stream)` → `mapPythonRespToIntent` → `applyUnifiedIntentResult`。后者注释写明「落库由 Python batch 完成」，仅处理 `need_confirm`、exit 与 `content` 回复。`handleUnifiedIntentAction` 及其依赖（`resolveEventForAction`、`event_child_pending.go`、`event_history_end.go`、`event_tree.go` 等）已无调用方；`prepareChatPreamble` 对 pending child 仅 clear 不续聊。

history 写库所有路径最终调用 `AddDeviceHistory`，当前对 insert 无「同 event 未闭合行」检查；`EndLatestHistoryIfMatch` 已按 `device_no + event_id + end_time=0` 闭合，start 侧缺少对称 guard。

基线：`openspec/specs/v3.0.0/spec.md` 相关章节。

## Goals / Non-Goals

**Goals:**

- 删除 voice 侧确认无用的 Go 写库/pending/事件树代码，使实现与 Python batch 落库架构一致。
- 在 history-service 单一 choke point 拒绝同 `eventId` 的重复进行中 start（insert 或 update 改回 `end_time=0`）。
- 同步 OpenSpec delta，消除 `handleUnifiedIntentAction` 与 Go pending 子事件规格的漂移。
- 中文注释覆盖新守卫分支与删码后的 live 路径说明。

**Non-Goals:**

- 不限制不同 `eventId` 同时进行中（保留「睡眠未结束中间记喂奶」）。
- 不在本变更内改 `python-ai-talk` 源码（可选 follow-up：解析 batch `reason` 改播报）。
- 不新增 HTTP API 版本或路径。
- 不引入 Redis 读缓存或 background ticker。
- 不恢复 Go 侧 `handleUnifiedIntentAction` 作为 fallback。

## Decisions

### 1. voice 删码范围：整文件删除 + 局部精简

- **选择**：删除 `event_child_pending.go`、`event_history_end.go`、`event_tree.go`；从 `voice_chat_understanding.go` 移除 `handleUnifiedIntentAction` 簇与 `mapPythonIntentToLandPlan`；从 `voice_chat.go` 移除 `eventInfo`、`pendingChild*` 字段与初始化；保留 `parseEventIntentFromReply`（`sanitizeModelReplyText` 仍用）。
- **理由**：grep 确认无 live 调用；pending child 已在 preamble 停用。
- **备选**：仅 `@Deprecated` 标注 — 否决，持续误导评审。

### 2. start 守卫落点：`AddDeviceHistory` + `UpdateDeviceHistory`

- **选择**：insert 时若 `EndTime==0 && EventId>0`，查询是否存在同 device+event 且 `end_time=0` 的行，存在则 `return error`；update 时若将 `EndTime` 设为 0，同样检查（排除当前 `id`）。
- **理由**：App add、Python batch、Admin add 全部覆盖；与 `EndLatestHistoryIfMatch` 查询维度对称。
- **备选**：仅在 `EventBatch` controller — 否决，App 手动 add 仍漏。

### 3. 守卫触发条件：仅「进行中」insert/update

- **选择**：`EndTime == 0` 视为进行中 start；`EndTime != 0`（含 one 瞬时 `start==end`）不拦截。
- **理由**：与产品「计时事件 end 留空=进行中」及 history.html 语义一致。
- **备选**：按 `eventType=time` 查 device 主档 — 否决，增加跨服务调用且 update 路径复杂。

### 4. 错误文案

- **选择**：固定中文，如 `该事件已在进行中，请先结束后再开始`；`EventBatch` 写入 item `Reason`；HTTP `EventAdd` 返回业务错误信息。
- **理由**：与 batch 既有「没有进行中的该事件记录」风格一致。

### 5. OpenSpec：REMOVED Go pending 子事件规格

- **选择**：`voice-event-child-disambiguation` 中 Go pending/AddHistory 相关 Requirement 整段 REMOVED，Migration 指向 Python 意图图消歧。
- **理由**：代码已不实现；保留规格会强制错误实现。

## Risks / Trade-offs

- [历史脏数据：已存在多条同 event 未闭合行] → 新守卫只防新增；不自动 merge 旧数据；可选运维 SQL 另议。
- [App 手动添加进行中行被拒绝] → 产品预期行为；UI 可提示先结束（本变更不强制改 Flutter）。
- [Python 未处理 batch 失败仍播报成功] → Python 侧 follow-up；Go/history 仍返回明确 reason。
- [删码后 spec 引用 ConfirmIntent 旧句] → tasks 含 chat-stream-intent-land 等 delta 一并修正。

## Migration Plan

1. 部署 `history-service`（含 start 守卫）— 旧客户端重复 start 会收到错误而非静默成功。
2. 部署 `voice-service`（删 dead code，行为对外无新增写库路径变化）。
3. 可选：升级 `python-ai-talk` 解析 batch `ok=false` reason。
4. 回滚：分别回退两服务二进制；守卫回退后重复 start 恢复为可插入。

## Open Questions

- Python 播报 follow-up 是否纳入本 PR 或单独兄弟仓变更：默认 **本仓库 tasks 仅 Go/history**；Python 列 optional task。
- 是否对 Admin 运维页 `event/add` 增加前端预检：非必须，服务端守卫已足够。
