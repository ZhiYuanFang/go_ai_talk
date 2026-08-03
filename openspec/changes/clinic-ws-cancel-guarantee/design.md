## Context

Clinic WS（`/voice/clinic/ws`）已具备 cancel / supersede / disconnect 打断能力：

- `voice_clinic_ws.go`：读循环非阻塞；`cancel` 匹配 `activeTurnID` 时 `cancelTurn()` 并 `EmitTurnCancelled(..., "cancelled")`；新 `question` 对旧 turn 发 `reason:"superseded"`；连接 defer 取消 active ctx，**不下发** `turn_cancelled`。
- `clinic_service.go`：`turnCtx` 取消时静默结束，不写 `answer_done`、不 consume `clinic_ai`。
- `python_ai_client.go`：`ClinicStream` 使用传入 `ctx`，随 cancel 中断 HTTP/读流。

跨端：Flutter 伴侣停止 UX 依赖上行 `cancel` 与下行 `turn_cancelled{reason:cancelled}`；Python 侧依赖请求 ctx 取消（`CancelledError`），**本变更不改 Python**。

文档漂移：基线 `v2.0.24` 与 `drop-clinic-go-session-and-summary` delta 仍将 `disconnected` 列为 `turn_cancelled.reason` 合法值，但实现从不下发该 reason。

## Goals / Non-Goals

**Goals:**

- 在 OpenSpec 中锁定最小 cancel 合同，与现网 Go 实现对齐。
- 修正 `turn_cancelled.reason` 枚举：仅 `cancelled` | `superseded`；断连取消 ctx 但不发帧。
- 明确 thinking / answer 两阶段均可取消，且取消不扣额度、不写 `answer_done`。
- 任务以 verify / 文档对齐为主；仅在代码与合同不符时做最小修补。

**Non-Goals:**

- 不新增 `turn_cancelled` 字段（含 question 文本）。
- 不做 tip / intent cancel。
- 不改 Python / Flutter 仓。
- 不做 Go 侧 clinic 消息列表或 Redis session（已由 `drop-clinic-go-session-and-summary` 收敛）。
- 不为「保证取消」发明新 payload 或第二套取消协议。

## Decisions

### 决策1：以核实既有实现为主，不重做取消架构

- **方案**：合同描述现有 `activeTurnID` + `context.CancelFunc` + `EmitTurnCancelled` 路径；apply 阶段先对照代码与场景清单，缺则补、齐则只改 spec/注释。
- **理由**：用户明确 Go scope 最小；现网路径已满足 Flutter 停止 UX。
- **备选**：重写 cancel 状态机 / 引入取消队列——排除，无产品缺口。

### 决策2：`reason` 枚举去掉 `disconnected`（对齐实现）

- **方案**：规范写明 `turn_cancelled.reason ∈ {cancelled, superseded}`；WS 关闭 MUST cancel ctx、MUST NOT 尝试写 `turn_cancelled`。
- **理由**：连接已断无法可靠下发；Flutter 以本地 disconnect 清理 UI；旧文档三值枚举造成假合同。
- **备选**：断连前 best-effort 发 `reason:disconnected`——排除，竞态多且客户端不依赖。

### 决策3：取消 payload 保持 `{type, turnId, reason}`

- **方案**：不下发 question 文本；不匹配 / 无 active turn 的 `cancel` 静默忽略（保持现状）。
- **理由**：Flutter 本地已有提问文案；静默忽略避免迟到 cancel 误伤后续 turn。
- **备选**：cancel 回显 question——排除，无硬需求。

### 决策4：额度与限流仅成功 `answer_done` 结算

- **方案**：重申 cancel / supersede / disconnect 路径 MUST NOT consume `clinic_ai`，MUST NOT 递增 `voice:clinic:rate:*` 成功计数。
- **理由**：与 `clinic_rate.go` / `HandleQuestion` 现注释一致；停止控件不得「白扣」额度。

### 决策5：Python / tip / Flutter 边界

- **方案**：Go 只保证取消传入 Python 的 `ctx`；不改 tip SSE、不改 intent；Flutter supersede 不恢复输入属客户端策略，服务端仅保证 `reason:superseded`。
- **理由**：跨端 explore 已锁定 Python「无需额外 cancel hardening」。

## Risks / Trade-offs

- [文档曾承诺 `disconnected` reason，归档后客户端若误依赖] → 现网未下发；本变更明确删除枚举值；Flutter 以本地 disconnect 为准。
- [核实中发现细微竞态（cancel 与 answer_done 同时）] → 以「ctx 已取消则不得写 answer_done」为准做最小修补；不扩大范围。
- [误把本变更做成大重构] → tasks 明示 verify-first；无缺口则零应用代码改动。

## Migration Plan

1. 合并 OpenSpec 变更；apply 时核对三文件实现与场景。
2. 若代码已合规：仅保留 spec/注释对齐，无需发版专为 cancel。
3. 若有小缺口：随常规 voice-service 发版；无 Redis / API 迁移。
4. 回滚：规格回退即可；无数据迁移。

## Open Questions

- （无）产品与 explore 决策已锁定；断连不下发 `turn_cancelled` 以代码为准写入规格。
