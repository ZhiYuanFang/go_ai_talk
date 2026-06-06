## Context

- 全项目仅 `TextChat` 在 HTTP 请求路径同步 `Publish` 且失败阻断（`voice_chat.go`）。
- `HandleWithTranscript`、`HandleTranscriptForStreaming` 等 WS/语音主链路本就不发 MQ。
- history/device 的 outbox 失败已是 warn-only，本变更与之对齐。

## Goals / Non-Goals

**Goals:**

- MQ 不可达时 `TextChat` 仍返回 LLM 对话结果。
- 发布失败打可观测 Warning（含 `metric=voice_task_publish_degraded`）。

**Non-Goals:**

- 不改为 outbox 缓冲、不补发积压事件。
- 不修改 v1.0.3 其它「必需事件发布失败阻断」路径（当前仅 TextChat 需豁免）。

## Decisions

1. **warn-only** 而非 outbox：改动最小，与 exploration 建议一致；voice.task 为审计/异步任务，丢失可接受。
2. 保留 `taskProducer != nil` 判断与 publish 调用，仅放宽错误处理。

## Risks / Trade-offs

- MQ 宕机期间 text-chat 来源的 `voice.task.requested` 事件会丢失 → 可接受（容灾权衡）。
