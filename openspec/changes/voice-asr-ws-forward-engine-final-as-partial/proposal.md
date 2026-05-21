## Why

`voice-asr-ws-client-only-finalize` 将百度 `onFinal` 设为 no-op，听写过程中客户端只能收到可能不准确的 `asr_partial`，往往在用户 `commit` 后才通过 `asr_final` 看到正确文本。产品希望在**不改变前端协议与截句语义**的前提下，说话过程中也能尽早展示引擎已收敛的更准文本。

## What Changes

- 听写 WS：当流式 STT 产生引擎级 `onFinal` 时，若文本非空且与上次 partial 不同，**再下发一条** `asr_partial`（与 `/voice/chat/ws` 一致），用于覆盖/纠正预览文案。
- **保持**方案 A：`onFinal` **仍不得**触发 `asr_final`、**不得**因 `onFinal` 关闭或重建 ASR 会话；`commit`/`end` 仍是业务定稿唯一路径。
- `onFinal` 文本写入 `latestTranscript` 时以引擎 final 为准（允许比 partial 更短但更准的纠错）。
- 更新 `README.MD` 听写 WS：说明中途可能收到「纠正性」`asr_partial`（来自引擎 final），与 `commit` 后的 `asr_final` 区分。
- **不修改** `/voice/chat/ws` 行为。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `voice-realtime-asr-ws`：补充引擎 `onFinal` 须以 `asr_partial` 转发、且不得作为 `asr_final` 或断句的语义（在 `voice-asr-ws-client-only-finalize` 增量之上修订）。

## Impact

- **代码**：`internal/controller/voice_asr_ws.go`（`CreateStreamASRSession` 的 `onFinal` 回调）。
- **契约**：下行仍为 `asr_partial` / `asr_final`，无新消息类型；前端若已按 partial 全量替换显示则无需改动。
- **部署**：仅 **voice-service**。
