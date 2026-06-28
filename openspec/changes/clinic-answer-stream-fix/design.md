## Context

智谱等 thinking 模型 SSE：先 `reasoning_content` 分片，再纯 `content` 分片。parser 仅在同分片同时含 reasoning 与 content 时将 content 标为 answer。

## Goals / Non-Goals

**Goals:** clinic 在 reasoning 阶段后正常流式 `answer_delta` 与 `answer_done.answer`。

**Non-Goals:** 改 WS 帧结构；改 Flutter；改 parser 全局「content 一律 answer」语义。

## Decisions

在 `invokeStreamHTTP` 维护 `sawReasoning`；`content` 分片在 `ThinkingEnabled && sawReasoning` 时写入 answer 通道。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| thinking 开启但模型从不发 reasoning | content 仍走 content 通道（与现网一致） |
| 闲聊 ThinkingEnabled=false | 不受影响 |
