## Why

语音球喂养流式路径 `HandleTranscriptForIntentStream` 当前先调 `AnalyzeIntentStream`（供 UI thinking），再经 `chatWithResult` **二次**调用非流式 `AnalyzeIntent` 落库。这造成双倍 Python 成本、UI 展示意图与落库意图可能不一致，以及 `NeedConfirm` / pending confirm 语义漂移风险。`fix-python-api-alignment` 的 `streaming-intent-service` 已要求「流式与非流式同落库逻辑」，但实现仍走二次调用，需本变更将流式结果直接落地。

## What Changes

- 重构 `HandleTranscriptForIntentStream`：Stream 结束后用已解析的意图结果走与 `chatWithResult` 相同的 confirm / 落库 / 回复逻辑，**禁止**再调非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`。
- 抽取可复用的「已有意图 → confirm/落库/回复」处理路径（供流式入口与 `chatWithResult` 共用），保留 `NeedConfirm` / pending confirm / quota 单次 consume / multi-event / `handleUnifiedIntentAction` 行为矩阵。
- 补充可观测日志：单次 Python intent 调用可区分流式落地路径。
- **不改** gateway、Flutter 协议、tip/feedback；TTS 非流式 `HandleTranscriptForStreaming` / `chatWithResult` 常规路径保持可用（仍可走非流式 AnalyzeIntent）。

## Capabilities

### New Capabilities

- `chat-stream-intent-land`: 流式意图分析结束后，用 stream 结果直接执行 confirm/落库/回复，且全路径仅一次 Python 意图调用（禁止二次非流式 AnalyzeIntent）

### Modified Capabilities

- （无）主库 `openspec/specs/` 无对应能力；本变更依赖并落实 `fix-python-api-alignment` 中 `streaming-intent-service` 的「同落库逻辑」要求，但不修改该已归档/进行中变更的 scope。

## Impact

- **代码**：仅 `internal/services/voice`（主要 `voice_chat.go` 的 `HandleTranscriptForIntentStream`、`voice_chat_understanding.go` 的 `chatWithResult` / 意图映射与落库编排）；可能微调 `mapPythonRespToIntent` 复用方式。
- **依赖**：引用已完成的 `confirm-ws-adaptation`（NeedConfirm / pending / ConfirmIntent）与 `fix-python-api-alignment`（AnalyzeIntentStream / 流式入口）；不扩其 scope。
- **API / 协议**：无对外 HTTP/WS 协议变更；gateway、Flutter、tip/feedback 不变。
- **服务边界**：voice 不直连他域 DAO；设备事件等仍经 `DeviceAdmin()` 等既有契约。
- **测试**：当前阶段不新增测试文件。
