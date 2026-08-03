## Why

Flutter 伴侣页停止控件依赖 Clinic WS 的 `cancel` → `turn_cancelled{reason:cancelled}` 契约；Go 侧实现已基本到位，但基线/既有 delta 仍把 `disconnected` 列为 `turn_cancelled.reason` 取值，与代码「断连只 cancel ctx、不下发帧」不一致。需要在 OpenSpec 中锁定最小 cancel 合同，便于跨端对齐与回归验证，避免再发明新 payload 或误改 Python。

## What Changes

- **锁定既有 cancel 合同**（无新帧类型、无新字段）：上行 `cancel{turnId}` 匹配 active turn → 取消 LLM/Python 请求 context；下行 `turn_cancelled{turnId, reason:"cancelled"}`。
- **明确适用相位**：thinking 与 answer 流式阶段均可中断；取消路径 **MUST NOT** 下发 `answer_done`，**MUST NOT** consume `clinic_ai`（亦不得递增 clinic 限流成功计数）。
- **保留 supersede**：新 `question` 打断旧 turn 时仍下发 `turn_cancelled{reason:"superseded"}`（Flutter 不因 supersede 恢复输入框；本仓只保证服务端 reason）。
- **对齐断连行为（spec 修正）**：**BREAKING（相对旧文档枚举，非相对现网实现）**：`turn_cancelled.reason` 取值收窄为 **`cancelled` | `superseded`**；WS 断开 MUST 取消 active turn context，但 **MUST NOT** 下发 `turn_cancelled`（连接已不可写）。
- **非目标**：不向 `turn_cancelled` 增加 question 文本；不做 tip/intent cancel；不改 Python；不做 Go 侧 clinic 消息列表 / Redis session；不实现应用层新功能（以核实/文档/对齐为主）。

## Capabilities

### New Capabilities

- （无）行为收敛到既有 `pangbao-ai-clinic` 的 cancel 合同澄清。

### Modified Capabilities

- `pangbao-ai-clinic`：锁定 Clinic WS cancel/supersede/disconnect 语义；将 `turn_cancelled.reason` 从含 `disconnected` 的三值枚举收窄为 `cancelled`/`superseded`；明确取消不写 `answer_done`、不扣 `clinic_ai`，且 thinking/answer 两阶段均生效。

## Impact

- **后端（本仓）**：主要对照 `internal/controller/voice_clinic_ws.go`、`internal/services/voice/clinic_service.go`、`python_ai_client.go`（ClinicStream ctx）；预期以 verify/document 为主，仅在发现与合同不符的小缺口时做最小修补。
- **OpenSpec**：相对 `drop-clinic-go-session-and-summary` / 基线 `v2.0.24` 中仍写 `reason=disconnected` 的帧协议条文做 delta 修正。
- **Flutter**：伴侣停止 UX 依赖本合同；本变更不改 Flutter 仓。Supersede 不恢复输入由 Flutter 自行处理。
- **Python**：无改动（ASGI `CancelledError` 已足够）。
- **风险**：极低——现网 Go 已按「断连不下发 turn_cancelled」实现；变更主要消除文档漂移。
