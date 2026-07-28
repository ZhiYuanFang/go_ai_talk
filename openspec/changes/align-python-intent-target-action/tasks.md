## 1. 映射函数

- [x] 1.1 在 `voice_chat_understanding.go`（或邻近文件）新增 Python 两轴 → Go `entity.Action` 的映射函数，附详细中文注释（target_type 领域 / action 喂养动作）
- [x] 1.2 映射表实现：feeding+start|end|one → TargetType=action；feeding+multi 由调用方走 multi；history→search；suggest→suggest；exit→exit；conversation/空/未知→不进入 CRUD（由上层仅回复）；disambiguate→不进入 CRUD
- [x] 1.3 Action.Name 回退保持 EventName/Action/transcript；禁止 `ParseActionTargetType(intent.TargetType)` 作为喂养 CRUD 开关

## 2. 接入 applyUnifiedIntentResult

- [x] 2.1 重写 `applyUnifiedIntentResult` 在 `NeedConfirm` 早退之后的分支：按 target_type 分流，feeding 可执行动作走 `handleUnifiedIntentAction`，conversation 仅 NL
- [x] 2.2 feeding+multi 仍走 `handleMultiEventIntent`（或经 handleUnifiedIntentAction 现有 multi 入口），子项继续用子 action
- [x] 2.3 feeding+disambiguate 且未 NeedConfirm：仅 NL，不 AddHistory；未知 target_type Warning + 当 conversation

## 3. 清理与自检

- [x] 3.1 grep：`ParseActionTargetType(intent.TargetType)` 在单事件 CRUD 构造路径应为 0（或仅剩非驱动用途并注释说明）
- [x] 3.2 确认流式/非流式均经 `applyUnifiedIntentResult`，无需双改；未改 Python、未改 gateway/Redis/quota
- [x] 3.3 行为矩阵口头核对：feeding+one 入库；start/end；history/suggest/exit/conversation；need_confirm 不入库；multi；disambiguate 防御
