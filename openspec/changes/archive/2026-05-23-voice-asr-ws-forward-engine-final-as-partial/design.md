## Context

- `/voice/asr/ws` 已完成 `voice-asr-ws-client-only-finalize`：`onFinal` 为 no-op；`asr_final` 仅由 `commit`/`end` 产生；无服务端静音截句。
- `/voice/chat/ws` 在 `voice_ws.go` 中已将百度 `onFinal` **再发一条** `asr_partial`（不当作 `asr_final`、不关 ASR），听写线拟对齐该模式。
- 用户诉求：说话过程中预览文字能随引擎 final 纠正，**前端不改协议**（仍只消费 `asr_partial` 更新 UI）。

## Goals / Non-Goals

**Goals:**

- 引擎 `onFinal` 非空且与上次 partial 不同时，下发 `asr_partial` 携带更正文本。
- 更新 `latestTranscript` / `lastPartialText`，供后续 `commit` 回退与 FINISH 结果合并。
- **保持** `asr_final` 仅 `commit`/`end`；**保持**不因 `onFinal` 关闭 ASR。

**Non-Goals:**

- 新增 `asr_refined` 等消息类型或 `stable` 字段。
- 修改百度 STT 参数、`/voice/chat/ws`、网关路由。
- 恢复 silence / auto_commit / `onFinal` → `asr_final`。

## Decisions

### 1. onFinal → asr_partial（对齐 chat WS）

- **选择**：`onFinal` 回调内：`trim` → 非空 → 若 `text != lastPartialText` 则更新状态并 `emitAsrPartial(text)`。
- **理由**：与 `voice_ws.go` 一致；前端零改动；不破坏方案 A 截句权。
- **备选**：新消息类型 `asr_stable` — 未采用，避免联调成本。

### 2. latestTranscript 在 onFinal 上直接赋值

- **选择**：`onFinal` 时 `latestTranscript = text`（不用 `preferLongerTranscript`），允许更短但更准的纠错覆盖 partial 累积。
- **理由**：partial 路径仍用 preferLonger；引擎 final 对该短语具权威性。
- **onPartial**：保持现有 `preferLongerTranscript` + 去重逻辑。

### 3. 日志

- **选择**：`onFinal` 转发时打 Info（deviceNo、textLen），便于与「仅 commit 才 final」的联调日志区分。
- **非 Goals**：不引入新 metrics。

### 4. 与 client-only-finalize 的关系

- 本变更 **修订** client-only-finalize 中对「引擎 onFinal 完全不下发」的表述，改为「不下发 asr_final，但以 asr_partial 转发」。
- 归档顺序：本变更依赖 client-only-finalize 已落地；合并 spec 时注意 delta 叠加。

## Risks / Trade-offs

- **[Risk] 前端把某条 partial 误当「已提交」** → 仍须 UI 区分预览与 `commit` 后的 `asr_final`；文档强调。
- **[Risk] 中途 onFinal 与 commit 后 FINISH 文本仍不一致** → 接受；commit 的 `asr_final` 仍为业务定稿。
- **[Risk] onFinal 频率高于 partial** → 已有 `text == lastPartialText` 去重。

## Migration Plan

1. 部署 voice-service。
2. 前端无需改协议；可选在 UI 上对 partial 保持「识别中」样式直至 `asr_final`。
3. 回滚：将 `onFinal` 恢复为 no-op（上一版 `voice_asr_ws.go`）。

## Open Questions

- 无。
