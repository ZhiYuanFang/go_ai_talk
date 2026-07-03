## Context

sim-user-service 经 ucg internal `POST /ucg/internal/api/posts/sample` 为 T2/T6 抽帖。辩论帖判定与 UCG `isDebatePost` 一致：左右立场标签均非空，或 `type=debate`。辩论评论门禁要求先投票。现有 sample 无 type/立场字段，无法在 sim 侧可靠过滤。

约束：sim MUST NOT 直查 ucg DAO；新增周期任务须 OpenSpec 批准；LLM MUST 经 `aimodel.Invoke`；Redis 无新增读缓存。

## Goals / Non-Goals

**Goals:**

- T2 仅评论非辩论帖（moment/图文等）。
- T7 每 12h 自动产一条可投票的辩论帖（pending_audit）。
- T8 每 1h 对随机辩论帖投票并发表 ≤10 字论点。
- Admin 可配开关/周期/prompt，与 T1–T6 一致。

**Non-Goals:**

- UCG 侧强制 10 字论点校验（sim prompt + 截断兜底即可）。
- T8 去重「同一 sim 已评论同一帖」（允许多 sim 各评一条）。
- 修改 App 对外 API 结构（v1 不变）。

## Decisions

### 1. Sample API 在 ucg SQL 层过滤辩论帖

- **决策**：请求 body 增加互斥布尔 `excludeDebate`、`onlyDebate`；WHERE 与 `isDebatePost` 对齐。
- **理由**：T2/T8 单次 random 只返 1 条，客户端 retry 浪费；与 `excludeMediaTypes` 同模式。
- **响应**：辩论相关字段 `debateLeft`、`debateRight` 供 T8 prompt 使用。

### 2. T7 LLM 输出 JSON 三字段

- **决策**：prompt `post_debate_text` 要求 JSON `{content, debateLeft, debateRight}`；sim 校验标签 ≤5 字后发帖。
- **理由**：与 `validateDebateLabel` 一致；失败记 task fail。

### 3. T8 流程 vote → comment

- **决策**：LLM 一次输出 `{side, argument}`；`side` 为 `left|right`；`argument` utf8 ≤10；先 `POST .../vote` 再 `POST .../comments`。
- **理由**：满足 UCG 门禁；论点与立场一致。
- **作者排除**：若 `authorWxId == simWxId`，记失败「不可评自己的帖」且 MUST NOT 投票/评论。

### 4. 默认周期

- T7 `post_debate`：12h（`SIM_INTERVAL_POST_DEBATE`）
- T8 `debate_comment`：1h（`SIM_INTERVAL_DEBATE_COMMENT`）

### 5. LLM lane

- T7/T8 均用 `LaneSimText`（辩论帖无配图）；T2 仍用 `LaneSimVision`。

## Risks / Trade-offs

- **[Risk] 无 published 辩论帖时 T8 持续失败** → 依赖 T7 或人工发帖；记明确错误信息。
- **[Risk] T8 1h + T7 12h 增加 App HTTP 与 LLM 负载** → 已有全局限速；admin 可关任务。
- **[Risk] sample 过滤与 `isDebatePost` 漂移** → SQL 注释引用同一判定函数语义；评审对照 `vote.go`。

## Migration Plan

1. 部署 **ucg-service**（sample API 扩展）。
2. 部署 **sim-user-service**（T2/T7/T8 + schema seed prompt）。
3. sim-admin 确认 T7/T8 开关与周期；可选手动执行验证。

## Open Questions

- 无（explore 阶段已确认：T7 12h + T2 排除辩论 + T8 1h 论点 ≤10 字）。
