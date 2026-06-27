## Context

- 当前 T5：`randomSimSession` → App `GET /conversations` → 对每个真人未读 `spawnEphemeralChat`（detached goroutine，循环 `ephemeralChatLoop`×`ephemeralChatWindow`）。
- `sim-gentle-polling` 已将 E1 默认降为 5min×15min，但未改变「多 goroutine 扇出」模型。
- `sim-random-wx-pick` 已引入 `GET /sim/wx/random` 解决「random 单 sim」；T5 需要 **全量 sim wxId** 供 ucg `IN` 过滤，**不能**用 `sim/wx/list?pageSize=200` 第一页（超过 200 个 sim 时后段用户永不可选）。
- 未读数据在 ucg 库 `ucg_conversation_member.unread_count`；`is_simulated` 在 device 库 `wx` 表——须 HTTP 契约组合，禁止 sim-user 直查 ucg DAO。
- 负责人已确认：仅回复真人 peer；sim↔sim 未读忽略；不要循环重试 random sim（策略 B）；要 O(1) 有界探测（策略 C + device 全量 ids）。

## Goals / Non-Goals

**Goals:**

- T5 每 tick：**1 次** device 全量 ids + **1 次** ucg sample + **1 次** LLM + **1 次** internal send；无 E1 子 goroutine。
- ucg sample：在 caller 传入的**完整** `simWxIds` 上，对 eligible 未读会话做 ID 锚点随机（2 条有界 SQL），peer 用 `NOT IN simWxIds` 排除 sim↔sim。
- sample 响应附带最近 N 条消息，T5 **无需** `usernameLogin` + App `GET messages`。
- 移除全部 ephemeral 配置与 Admin/E1 文案。

**Non-Goals:**

- 不恢复 E1 或连续多轮对话；不缩短 T5 周期（运营可自行调 `intervals.chat`）。
- 不在 ucg 库镜像 `is_simulated` 列。
- 不 Redis 缓存 sim wxIds 或未读索引。
- 不新增 `*_test.go`；不改动真人 App 聊天 API。

## Decisions

### 1. device：`GET /device/internal/api/sim/wx/ids`

**选择**：专用接口，单 SQL `SELECT id FROM wx WHERE is_simulated=1 ORDER BY id`，响应 `{ ids: int64[], total: int }`。

**上限**：`total` 超过 **10000** 时返回 400（与 Admin maxSimUsers 1000 留足余量，防误用全表传输）。

**备选**：复用 list 分页循环拉全 — 否决（用户明确要求不要多次 DB/HTTP；语义也不清晰）。

**与 list 关系**：`sim/wx/list` 保留分页列举；T5 **MUST** 使用 ids 接口。

### 2. ucg：`POST /ucg/internal/api/chat/sim-unread-sample`

**请求**：`{ "simWxIds": [101,102,...], "messageLimit": 20 }`（`messageLimit` 默认 20，最大 50）。

**eligible 集合 SQL 模型**（direct 1:1，两成员）：

```
m = ucg_conversation_member (sim 侧，未读)
peer = 同 conversation_id 且 wx_id != m.wx_id
WHERE m.unread_count > 0 AND m.deleted_at = 0
  AND m.wx_id IN (simWxIds)
  AND peer.wx_id NOT IN (simWxIds)
  AND peer.deleted_at = 0
```

**随机**：对 eligible 的 `m.id` 做 MIN/MAX + 均匀锚点 + `id >= anchor LIMIT 1` + minId 回退（对齐 `SampleRandomPublishedPost` / `PickRandomSimulatedWx`）。**MUST NOT** `ORDER BY RAND()`；**MUST NOT** 在 simWxIds 上循环重试。

**响应（命中）**：

```json
{
  "conversationId": 123,
  "simWxId": 101,
  "peerWxId": 999,
  "unreadCount": 2,
  "messages": [{ "senderWxId": 999, "content": "..." }]
}
```

无 eligible → `found: false` 或空 data，HTTP 200。

**消息**：按 `ucg_chat_message` 对该 `conversationId` 取最近 `messageLimit` 条（仅 `status` 可见/已审核通过的消息，与 App 列表语义一致）。

**备选**：sample 只返回 convId，sim 走 App API 拉消息 — 否决（多 2 次 gateway HTTP，且需 login）。

### 3. sim-user T5 编排

```
RunChatScanTask:
  ids := GET device /sim/wx/ids
  if len(ids)==0 → RecordTaskRun(false, "无模拟用户")
  sample := POST ucg /chat/sim-unread-sample { simWxIds: ids }
  if !sample.found → RecordTaskRun(false, "无未读会话")
  history := buildChatHistory(sample.messages)
  LLM chat_reply → sendInternalChat(sample.simWxId, sample.conversationId, reply)
  RecordTaskRun(true, "")
```

删除：`spawnEphemeralChat`、`ephemeralMu`、`ephemeralActive`；`RunChatScanTask` 签名可去掉 `flags RuntimeFlags` 中 ephemeral 依赖（若仅剩 chat 间隔则保留 flags 参数供 scheduler 传递）。

手动执行 `chat_scan` 走同一 `RunChatScanTask`。

### 4. 删除 E1 配置

从 `RuntimeConfigDB`、`RuntimeFlags`、env、`runtime_api`、`runtime_snapshot`、`config_admin` effects、`sim-admin.html` 移除 `ephemeralChatLoop`/`ephemeralChatWindow`。DB 既有 `runtime_json` 中残留字段 **忽略**（读路径不再映射）。

`task_llm` 中 `chat_scan` usage 由「E1 回复」改为「未读回复」。

### 5. 部署顺序

1. device-service（ids API）
2. ucg-service（sample API）
3. sim-user-service + sim-admin.html

回滚：revert sim 恢复 E1 行为需旧镜像；ucg/device 新接口对旧 sim 无影响。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| sim 数量极大导致 ids payload 大 | 上限 10000；运营 maxSimUsers 默认 100 |
| ucg `IN (1000 ids)` 查询计划 | sim 规模可控；eligible 子集更小 |
| 锚点落在无 eligible 的 id 空洞 | minId 回退；bounds 在 eligible 子集上计算 |
| 单次 tick 只回 1 条，回复密度下降 | 可调短 `intervals.chat`；符合「gentle polling」 |
| 历史 runtime_json 含 ephemeral 字段 | 读路径忽略；Admin 保存不再写入 |

## Migration Plan

1. 部署 device → ucg → sim-user。
2. 无 DB 迁移；重启 sim-user scheduler 后 T5 即新逻辑；**进行中的 E1 goroutine** 在升级后自然消失（不强制 kill，与旧 `ephemeral_may_continue` 语义一致，但新 Admin 不再提示 E1）。
3. 验收：T5 日志无 `spawnEphemeralChat`；无新 E1 goroutine；手动 T5 在无未读时 status 显示 skip 文案。

## Open Questions

- 无（探索阶段已确认：全量 ids、真人 peer、删 E1、sample 带 messages）。
