## Context

- `/voice/chat/ws`（`voice_ws.go`）已实现 `mode=stream` + `CreateStreamASRSession` + `asr_partial`/`asr_final`，但 `commit` 后进入 `HandleTranscriptForStreaming`、TTS 与 `DeviceAdmin.UpdateLastTalk`；`triggerRealtimeTranslate` 为空实现，无机器翻译。
- 网关通过 `installVoiceWSProxyMiddleware` 仅绑定 `/voice/chat/ws`，`VOICE_WS_PROXY_URL` 指向 voice-service（:9802）。
- 仓库约定：voice 域不跨库直查 device/history 表；听写 WS 不需要设备画像时可弱化 `deviceNo`，但为与现有 STT 日志/限额习惯一致，仍建议在 `start` 中携带 `deviceNo`（可选会话标识，**不**注册 `VoiceWSManager`）。

## Goals / Non-Goals

**Goals:**

- 提供 **`GET /voice/asr/ws`（WebSocket Upgrade）** 专用入口，上行 PCM、下行实时中文听写文本。
- 复用 `ws_protocol` 的 `start` 解析与 PCM 校验、`Voice().CreateStreamASRSession`、百度流式 ASR 实现。
- 网关与 App 网关对前端暴露与 chat WS 相同的根路径前缀策略（经 gateway 访问，不强制直连 voice-service）。
- 控制消息与 chat WS 对齐子集：`start` / `end` / `ping`，**不要求** `commit`（无对话轮次）。

**Non-Goals:**

- 机器翻译、DeepSeek、TTS、`audio_chunk`、`exit`、事件记录、成长建议。
- 修改 `/voice/chat/ws` 协议或实现。
- 听写结果落库（若未来需要历史，应经 history-service HTTP 契约，不在本变更实现）。
- 新增测试文件（遵循仓库当前阶段约定）。

## Decisions

### 1. 路径命名：`/voice/asr/ws`

- **选择**：`/voice/asr/ws`（听写语义明确）。
- **备选**：`/voice/translate/ws` — 易与语种翻译混淆，弃用。

### 2. 独立 handler 文件，不扩展 `voice_ws.go`

- **选择**：`voice_asr_ws.go` + `registerVoiceAsrWS`，逻辑控制在 ~200 行量级（读循环、ASR 回调、错误写回）。
- **理由**：避免 chat WS 状态机继续膨胀；前端契约稳定、可独立演进。

### 3. 不复用 `VoiceWSManager`

- **选择**：听写连接不与「单设备单对话连接」互踢。
- **理由**：用户可能同时打开听写页与对话页；manager 仅服务 chat 打断/替换场景。

### 4. 协议子集（与 chat stream 对齐音频，精简控制）

**上行：**

- Text `start`：`type`、`deviceNo`（非空，作会话/日志标识）、`sampleRate`、`bits`、`channels`；`mode` 固定为 `stream` 或省略由服务端默认 stream。
- Binary：s16le PCM 分片（与 chat 相同）。
- Text `end`：关闭 ASR 会话并回复 `ended`。
- Text `ping` → `pong`。

**下行：**

- `started`、`asr_partial`、`asr_final`（含服务端静音触发的 finalize，可带 `source` 字段区分 `client`/`silence`/`asr_callback`）、`ended`、`error`。
- **不发送**：`audio_chunk`、`chat_delta`、`interrupt_commit`（听写场景可用 `asr_final` + 可选 `utterance_end` 若需显式句界；首版与 chat 一致仅 `asr_*` 即可）。

**静音与句界：**

- 复用 chat WS 同类启发式（无 STT 回调超时、STT 静默后 finalize）**或** 简化为仅依赖百度 `onFinal` + 客户端 `end`；首版建议 **复制必要静音 commit 逻辑** 以保证「说完即出 final」，避免只听 partial 不收 final。实现时在 design 落地为：从 `voice_ws.go` 抽取可共享的静音检测常量到包内私有 helper，或 ASR handler 内保留最小副本（避免大范围重构 chat WS）。

### 5. 网关透传：同 env，第二路径绑定

- **选择**：`installVoiceWSProxyMiddleware` 泛化为对多个路径注册同一 proxy（`/voice/chat/ws` + `/voice/asr/ws`），仍使用 `VOICE_WS_ROUTE_MODE` / `VOICE_WS_PROXY_URL`。
- **备选**：单独 `VOICE_ASR_WS_PROXY_URL` — 配置重复，弃用。

### 6. App 网关鉴权

- **选择**：`/voice/asr/ws` 加入 `gatewayAppAuthExemptExactGET`（与 chat WS 一致，Upgrade 不要求 Bearer）。
- **理由**：设备侧听写页若需登录，由产品后续在 `start` 扩展 token；本变更保持与 chat WS 相同边缘策略。

## Risks / Trade-offs

- **[Risk] 与 chat WS 逻辑重复** → 首版允许适度重复读循环/ASR 回调；若重复超阈值再抽 `internal/controller/ws_asr_stream.go` 共享。
- **[Risk] 前端误连 chat WS** → README 与规格明确两条 URL；听写页只用 `/voice/asr/ws`。
- **[Risk] 网关未配 proxy 时 503** → 与 chat WS 相同，依赖部署文档；compose 已 `VOICE_WS_ROUTE_MODE=proxy`。
- **[Risk] 百度流式 ASR 未启用** → `start` 后返回 `error` stage=`stt`，与 chat 一致。

## Migration Plan

1. 先部署 **voice-service**（新路由），再部署 gateway / gateway-app（透传与白名单）。
2. 前端将听写场景切到 `/voice/asr/ws`；chat 场景保持 `/voice/chat/ws`。
3. 回滚：下线 voice 新 handler 或 gateway 去掉第二路径绑定；无数据迁移。

## Open Questions

- 是否在 `asr_final` 中增加 `finish_utterance: true` 字段以便 UI 分句 — 首版可沿用 chat 的 `asr_final` 形状，减少前端分支。
- `deviceNo` 是否强制校验已在 device 表注册 — 首版 **不强制**（仅非空校验），避免听写页依赖 device 注册流程；若产品要求可后续加 HTTP 校验契约。
