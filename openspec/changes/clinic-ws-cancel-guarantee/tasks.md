## 1. Verify existing cancel contract

- [x] 1.1 对照 `voice_clinic_ws.go`：匹配 active turn 的上行 `cancel{turnId}` 调用 `cancelTurn` 并下发 `turn_cancelled{turnId, reason:"cancelled"}`；不匹配/无 active 时静默忽略
- [x] 1.2 对照 supersede 路径：新 `question` 取消旧 turn 并下发 `turn_cancelled{reason:"superseded"}`，且随后处理新 turn
- [x] 1.3 对照 disconnect：连接关闭/读循环退出时取消 active ctx，且 **不下发** `turn_cancelled`（无 `reason:"disconnected"`）
- [x] 1.4 对照 `clinic_service.go` / `python_ai_client.go`：turn ctx 取消在 thinking 与 answer 阶段均可中断流；取消后不写 `answer_done`、不 consume `clinic_ai`、不递增限流成功计数

## 2. Align specs and comments (minimal code)

- [x] 2.1 确认本变更 `specs/pangbao-ai-clinic/spec.md` 与实现一致：`reason` 仅 `cancelled`|`superseded`；断连场景含 MUST NOT 下发 `turn_cancelled`
- [x] 2.2 若注释仍暗示会下发 `reason:disconnected` 或与合同不符，做最小注释修正（不改协议字段）
- [x] 2.3 仅当核实发现真实行为缺口时做最小代码修补；无缺口则本变更零应用逻辑改动

## 3. Regression checklist (manual / spot)

- [x] 3.1 手工或日志 spot-check：thinking 阶段 `cancel` → `turn_cancelled` cancelled，无 `answer_done`
- [x] 3.2 手工或日志 spot-check：answer 阶段 `cancel` → 同上
- [x] 3.3 手工或日志 spot-check：supersede → 旧 turn `reason:superseded`；新 turn 正常流式
- [x] 3.4 确认未引入 tip/intent/Python/session 相关改动
