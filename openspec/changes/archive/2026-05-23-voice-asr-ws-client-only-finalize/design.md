## Context

- `/voice/asr/ws` 已在 `voice-realtime-asr-ws` 变更中实现，当前仍包含：
  - `wsInterruptCommitGap` / `wsInitialNoASRGap` → `runAsrFinalize("silence")`
  - `tryAutoCommitWhenNoASRCallback` → `runAsrFinalize("auto_commit")`
  - `CreateStreamASRSession` 的 `onFinal` → `emitAsrFinal` + `resetStreamASRUntilNextValid`
- 用户确认采用 **方案 A**：百度引擎的 final 不作为业务句界，**仅前端 `commit`/`end`** 触发 `asr_final`。

## Goals / Non-Goals

**Goals:**

- 听写 WS **MUST NOT** 因服务端计时或静音启发式调用 `runAsrFinalize`。
- `asr_partial` 仍实时下发（百度 partial 回调）。
- `asr_final` **仅**在客户端 `commit` 或 `end`（且选择带 finalize）时下发。
- `onFinal` 回调：忽略（不 WS 下行、不因 final 关闭 ASR 会话）。

**Non-Goals:**

- 修改 `/voice/chat/ws` 静音 interrupt 逻辑。
- 修改百度 STT SDK/连接参数（`devPid` 等）。
- 新增配置开关（本变更听写线固定 client-only，避免与 chat 共用的复杂度）。

## Decisions

### 1. 方案 A：忽略 `onFinal`

- **选择**：`onFinal` 注册为 no-op（或 Debug 日志），**不** `emitAsrFinal`，**不** `resetStreamASRUntilNextValid`。
- **理由**：与「仅前端结束一句」一致；避免百度 VAD 与产品截句冲突。
- **备选 B**（onFinal 当 partial）：未采用，避免前端误把引擎 final 当可提交句。

### 2. 删除服务端主动截句

- 移除二进制处理路径中的 `noFirstSTTTimeout` / `sttTimeout` → `silence`。
- 移除 `tryAutoCommitWhenNoASRCallback` 及调用。
- 可删除仅听写使用的 `utteranceStartAt` / `lastASRAt` 等静音判断字段（若不再引用）。

### 3. `commit` 与 `end` 语义

| 消息 | 行为 |
|------|------|
| `commit` | 若存在流式 ASR 或缓冲音频，执行 `runAsrFinalize("client")`；否则 `validate` 错误 |
| `end` | 若仍有未 finalize 音频，先 `runAsrFinalize("end")`；然后 `started=false`、关闭 ASR、回复 `ended` |

- **不**在 `end` 时强制要求此前发过 `commit`（允许用户只 end 清理连接）。

### 4. `asr_final` 的 `source` 枚举收敛

- 听写线仅保留：`client`、`end`（及 commit 失败时仍可能 `asr_no_result`）。
- 文档与实现 **不再产生** `silence`、`auto_commit`、`asr_callback`。

### 5. 与 chat WS 代码关系

- **不**抽取共享包强制统一；听写 handler 保持独立文件，避免 chat 行为被误改。
- chat WS 继续保留静音 commit 逻辑。

## Risks / Trade-offs

- **[Risk] 用户不说话也不 commit/end** → 连接与百度 WS 常驻；由前端页面卸载发 `end` 或业务超时。
- **[Risk] 仅 partial 无 final** → UI 须明确「完成」按钮发 `commit`；README 与联调说明需强调。
- **[Risk] 长句中间停顿** → partial 可能停更久直至用户 commit；可接受（听写/翻译场景由用户提交）。

## Migration Plan

1. 部署新版 voice-service。
2. 通知前端：听写页 **必须** 在结束时发 `commit`（或 `end` 前 commit）；勿等待自动 `asr_final`。
3. 回滚：恢复带 silence/onFinal 的旧 `voice_asr_ws.go` 版本。

## Open Questions

- 无。方案 A 已由产品确认。
