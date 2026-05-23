## Why

`/voice/asr/ws` 当前沿用了对话 WS 的静音截句（`silence`、`auto_commit`）与百度 `onFinal` 自动下发 `asr_final`，导致：① 与产品「仅听写、由前端决定何时结束」不符；② 易触发空 commit、ASR 反复建连；③ 前端难以区分「预览 partial」与「可提交的 final」。听写线应收敛为 **方案 A：仅前端 `commit`/`end` 产生 `asr_final`**。

## What Changes

- 移除听写 WS 内所有服务端主动截句：`silence`（STT 静默计时）、`noFirstSTT` 超时、`auto_commit` 无回调兜底。
- **方案 A**：忽略流式 STT 的 `onFinal` 回调——不向客户端发送 `asr_final`，不因此关闭当前 ASR 会话；仅通过 `asr_partial` 推送中间结果。
- 明确 **`commit` 为一句听写结束的唯一业务截句**：触发百度 FINISH/Commit 后下发 `asr_final`（`source: client`）。
- 明确 **`end` 为整段听写会话结束**：可选在有关闭前执行一次 commit（有音频时），并回复 `ended`。
- 更新 `README.MD` 听写 WS 说明：前端必须 `commit` 才能得到 `asr_final`；**不**再文档化 `source: silence` / `auto_commit`。
- **不修改** `/voice/chat/ws` 行为。

## Capabilities

### New Capabilities

（无独立新能力名；本变更修订听写 WS 截句语义。）

### Modified Capabilities

- `voice-realtime-asr-ws`：修订 `/voice/asr/ws` 的截句权、下行 `asr_final` 触发条件与 `commit` 必选语义（相对 `openspec/changes/voice-realtime-asr-ws` 中既有要求）。

## Impact

- **代码**：`internal/controller/voice_asr_ws.go`（删除静音/自动 commit 分支；`onFinal` 空实现或仅日志）。
- **契约**：前端听写页须在松手/完成时发 `commit`；不得依赖服务端自动 `asr_final`。
- **部署**：仅 **voice-service**（及已部署的 gateway 透传，无变更）。
