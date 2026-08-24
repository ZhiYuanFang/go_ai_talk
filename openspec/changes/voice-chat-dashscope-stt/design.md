## Context

- **现状**：`voice-service` 全部流式 STT 经 `CreateStreamASRSession` → `stt_stream_baidu.go`（百度 `wss://vop.baidu.com/realtime_asr`，`dev_pid=1537`）。`/voice/chat/ws` 与 `/voice/asr/ws` 共用同一配置与实现。
- **问题**：对话场景为硬件**远场**（老人/成人），1537 近场模型字错率高；听写场景为 App **近场**，百度仍有免费额度且效果可接受。
- **约束**：客户端 WS 协议不变；TTS 继续百度；不新增 Redis；不新增测试文件；配置仍集中在 `voice-chat.shared.yaml`；跨服务边界不变。

## Goals / Non-Goals

**Goals:**

- `/voice/chat/ws` 流式 STT 主路径切换为百炼 `qwen-audio-3.0-asr-flash-streaming`。
- 配置与代码支持 **profile 分流**：`chat` → DashScope，`dictation` → 百度（沿用现有 `stt` 块或 `sttDictation`）。
- 实现 DashScope WebSocket 协议（run-task → 二进制 PCM → finish-task → result-generated）。
- 远场可调参数：`speech_noise_threshold`、`language_hints: ["zh"]`；预留热词扩展点（配置或后续迭代）。
- 日志可观测：provider、model、profile、音频 avgAbs 等与现有百度日志对齐。

**Non-Goals:**

- 不切换 `/voice/asr/ws` 听写 STT（仍百度）。
- 不切换 TTS。
- 不在本变更内做 A/B 自动化或 badcase 评测工具。
- 不删除 `stt_stream_baidu.go`（保留听写与可选 fallback）。
- 不引入非流式 HTTP ASR 作为主路径（除非流式 commit 失败时的明确 fallback，可选二期）。

## Decisions

### 1. Profile 分流而非双 WS 端点

**决策**：扩展 `CreateStreamASRSession(ctx, profile, meta, onPartial, onFinal)`，`profile` 为 `chat` | `dictation`。

**理由**：两条 WS 已分离（`voice_ws.go` / `voice_asr_ws.go`），仅需在调用侧传入 profile，避免重复端点或客户端协议变更。

**备选**：两套独立 Service 方法 — 拒绝，重复逻辑多。

### 2. 配置结构：`sttChat` + `sttDictation`

**决策**：在 `voice-chat.shared.yaml` 的 `voiceChat` 下新增：

```yaml
sttChat:
  provider: dashscope
  model: qwen-audio-3.0-asr-flash-streaming
  streamEnabled: true
  streamEndpoint: ""   # 空则按 workspaceId 拼 wss://{id}.cn-beijing.maas.aliyuncs.com/api-ws/v1/inference
  workspaceId: ""
  apiKey: ""           # 空则从 VOICE_DASHSCOPE_API_KEY / UCG_DASHSCOPE_API_KEY 读取
  speechNoiseThreshold: -0.2   # 远场小声，可运维调参
  format: pcm
  timeoutSeconds: 20
  maxConcurrency: 50

sttDictation:
  # 与现有 stt 块等价，provider: baidu，devPid: 1537/80001
```

现有顶层 `stt` 块在迁移期可 **mirror 到 sttDictation 默认值**，避免听写行为突变。

**理由**：语义清晰；对话与听写独立演进。

### 3. DashScope 实现：`stt_stream_dashscope.go`

**决策**：新建 `dashscopeStreamASRSession`，实现 `StreamASRSession` 三方法：

| 方法 | 行为 |
|------|------|
| `WriteAudio` | 发送 BinaryMessage（16-bit LE PCM，与百度一致） |
| `Commit` | 发送 `finish-task` JSON，等待 `task-finished` 或最终 `result-generated` |
| `Close` | 关闭 WS |

**协议要点**（参考百炼 WebSocket API）：

1. 握手 Header：`Authorization: Bearer <api_key>`
2. 发送 `run-task`：`model`、`parameters.format=pcm`、`parameters.sample_rate=16000`、`input.task_group=audio`、`input.task=asr`、`input.function=recognition`
3. 等待 `task-started`
4. 流式 `result-generated` → 解析 partial/final 文本 → 回调 `onPartial` / `onFinal`
5. `finish-task` → `task-finished`

**理由**：Go 侧无官方 SDK，与现有 `stt_stream_baidu.go` 结构对称，便于维护。

### 4. 模型选型：对话固定 `qwen-audio-3.0-asr-flash-streaming`

**决策**：对话 profile 默认该模型，不用 `fun-asr-realtime`。

**理由**：用户已选定；支持即时热词与更强噪声鲁棒；与听写百度形成能力差异，后续听写再统一百炼时可换 `fun-asr-realtime`。

### 5. 密钥与端点

**决策**：

- API Key 加载顺序：`sttChat.apiKey` → `VOICE_DASHSCOPE_API_KEY` → `UCG_DASHSCOPE_API_KEY`。
- Workspace ID：`sttChat.workspaceId` → `DASHSCOPE_WORKSPACE_ID` env；必填，缺失则 chat STT 启动失败并打明确日志。

**理由**：复用现有 UCG 密钥降低运维成本；独立 env 便于 voice 计费隔离。

### 6. Fallback

**决策**：默认 **不** 自动 fallback 到百度（避免远场问题复现）。配置项 `sttChat.fallbackProvider: baidu` 可选，仅百炼建连/首包失败时启用；本变更 tasks 中标记为可选任务。

### 7. 整段 transcribe（非流式路径）

**决策**：`HandleWithDialogue` 等非 WS 路径若走 STT，chat 场景同样按 `sttChat` provider 分发；若 DashScope 仅实现流式，可 buffer 后走单次 WS session commit，或暂保留百度 fallback 并记录 TODO — **实现阶段优先保证 `/voice/chat/ws` 流式路径**。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 百炼 WS 协议变更 | 封装独立文件；日志 raw message 截断 |
| Workspace ID 配置遗漏 | 启动/建连时 fail-fast + 明确 error stage=stt |
| 远场参数不适配 | `speechNoiseThreshold` 可配置；保留 avgAbs 告警 |
| 单厂商依赖（对话 STT） | 保留百度代码；可选 fallback 配置 |
| 密钥与 UCG 共用 | 支持 `VOICE_DASHSCOPE_API_KEY` 拆分 |
| 计费 | ~1.19 元/小时；对话量低于听写，可接受 |

## Migration Plan

1. **开发/测试**：test 环境配置 `sttChat` + DashScope 密钥；`/voice/chat/ws` 远场样例验收。
2. **灰度**：prod 配置切换 `sttChat.provider=dashscope`；听写 `sttDictation` 仍 baidu。
3. **回滚**：`sttChat.provider=baidu` 或移除 profile 分流回退全局 `stt`（保留代码路径）。
4. **文档**：`.env.example` 补充 `VOICE_DASHSCOPE_API_KEY`、`DASHSCOPE_WORKSPACE_ID`；runbook 注明对话/听写 STT 分离。

## Open Questions

- 生产百炼 **Workspace ID** 是否已与 UCG 共用同一业务空间？（实现前需确认）
- 产品词热词表是否首期接入，还是二期？（建议二期，design 预留配置字段 `hotwords`）
- `speechNoiseThreshold` 初值 `-0.2` 是否需真机远场标定？（上线后根据日志调整）
