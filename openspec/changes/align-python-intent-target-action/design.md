## Context

Python `IntentResponse` 明确两轴：

- `target_type`：feeding | history | suggest | conversation | exit
- `action`：start | end | one | search | suggestion | reply | exit | multi |（澄清）disambiguate

Go `ActionTargetType` 是单轴动作枚举：start | end | one | exit | suggest | search | conversation。  
`applyUnifiedIntentResult` 当前用 `ParseActionTargetType(intent.TargetType)`：`feeding` 落入 default → conversation，喂养永不进 `handleUnifiedIntentAction` / `AddHistory`。多事件路径已用 `eventItem.Action`，单事件错位。

权威来源：兄弟仓 `app/feeding/schemas/intent.py`、`leaf_intent_result`、`pending_to_response_fields`。Go 必须跟 Python，不要求 Python 改字段名迁就 Go。

## Goals / Non-Goals

**Goals:**

- 统一落地路径按 Python 两轴正确分支
- `feeding` + `one|start|end` 进入既有 CRUD
- `history` / `suggest` / `conversation` / `exit` 按领域分流
- 流式与非流式共用同一映射（均经 `applyUnifiedIntentResult`）

**Non-Goals:**

- 不修改 Python intent API 或字段命名
- 不重写 `handleUnifiedIntentAction` / `AddHistory` 业务细节
- 不扩展 `options`/`confirm_type` UI
- 不改澄清 cid / quota（属 `intent-clarify-conversation-id`）

## Decisions

### 1. 以 target_type 为领域开关，action 为喂养动作

```
need_confirm=true → 既有澄清分支（不在此变更改）

need_confirm=false:
  switch target_type:
    feeding:
      action=multi           → handleMultiEventIntent
      action=start|end|one   → Action.TargetType=action → handleUnifiedIntentAction
      action=disambiguate    → 仅 NL（防御；正常应带 need_confirm）
      其他 action            → 仅 NL 或降级提示
    history                  → Action.TargetType=search → handleUnifiedIntentAction
    suggest                  → Action.TargetType=suggest → handleUnifiedIntentAction
    exit                     → Action.TargetType=exit
    conversation / 空 / 未知 → 仅回复 content
```

- **理由**：与 Python schema / 分类 prompt 一致；`suggestion`（action）≠ `suggest`（target_type），领域以 target_type 为准。
- **替代**：把 feeding 加入 `ParseActionTargetType` —— 否决（污染动作枚举，无法表达 feeding+one）。

### 2. 抽取显式映射函数

- **选择**：新增如 `mapPythonIntentToAction(intent) (entity.Action, kind)` 或在 `applyUnifiedIntentResult` 内集中 switch，禁止再 `ParseActionTargetType(intent.TargetType)` 驱动喂养。
- **理由**：单点真源，流式/非流式自动一致；便于中文注释标明两轴。
- **替代**：散落 if —— 易再错。

### 3. Action.Name 回退

- **选择**：`Action.Name` 仍可用 EventName / Action / transcript 回退（现网）；`TargetType` 必须来自上表映射，不得用 TargetType 字符串直接 Parse。
- **理由**：日志/展示名与 CRUD 开关分离。

### 4. 与澄清变更正交

- **选择**：本变更只修映射；`NeedConfirm` 早退、cid、免计额度保持澄清变更行为。
- **理由**：职责分离；联调时先澄清再喂养落库一条龙。

## Risks / Trade-offs

- [history/suggest 旧误路径曾被当 conversation] → 对齐后行为「变正确」；若客户端依赖错误闲聊需回归
- [未知 target_type] → 当 conversation，打 Warning 日志
- [action=suggestion 但 target_type 错标] → 以 target_type 为准；不解析 action=suggestion 为领域
- [disambiguate 且 need_confirm 漏标] → 仅 NL，避免误入库

## Migration Plan

1. 合入 Go voice 映射修复并部署。
2. 回归：feeding+one 入库；feeding+start/end；history/suggest/exit/conversation；need_confirm 澄清仍不入库；multi 不变。
3. 回滚：回退 voice 二进制即可。

## Open Questions

- 无。
