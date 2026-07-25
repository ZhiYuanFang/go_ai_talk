## Context

当前项目已完成 Python AI 流式接口（`/v1/tip/stream`、`/v1/analyze/intent/stream`）与 Go 侧核心流式逻辑封装（`PythonAIClient.TipStream()`、`HandleTranscriptForIntentStream()`、`IntentStreamCallback`）。但这些能力只存在于 voice 进程内部，缺少对外 HTTP 接口暴露给 Flutter 客户端。

现有架构为多进程部署：
- gateway-app：对外 App 接口入口，Bearer 鉴权，反向代理到各领域服务
- history-service：设备历史/建议/生日/文本对话接口
- voice-service：语音 ASR/TTS/AI 对话能力
- mcp-service：MCP 桥接，与小智平台对接

跨进程通信走 HTTP 契约接口（internal header 鉴权），禁止跨库直连。

## Goals / Non-Goals

**Goals:**
- 暴露小贴士流式生成的对外 SSE 接口，Flutter 客户端可直接调用
- 暴露纯文本对话流式接口，Flutter 语音球与 MCP 可复用
- 复用已有的流式核心逻辑（PythonAIClient、HandleTranscriptForIntentStream），零重复开发
- 保持原同步接口完全兼容，不做任何破坏性变更

**Non-Goals:**
- 不改造 MCP notifications/progress（本次 scope 排除）
- 不改造 `/voice/chat/ws` 语音球 TTS 音频链路
- 不新增任何 Redis 缓存
- 不修改数据库表结构

## Decisions

### Decision 1: SSE 作为流式传输协议（而非 WebSocket）

**方案选择**：使用 HTTP Server-Sent Events (SSE)

**理由**：
- Flutter 端小贴士 UI 已按 SSE 模式实现（`TipRepository.streamTip()`），对齐最小改动
- SSE 是单向服务端推送，正好匹配"客户端发请求→服务端流式推结果"的模式
- 相比 WebSocket，SSE 无需握手协议、自动重连、HTTP 缓存友好
- GoFame 框架对 SSE 的支持简单（写 ResponseWriter + flush）

**备选方案**：
- WebSocket：需要额外握手、心跳维护，对单向推送场景过重
- gRPC Stream：项目未引入 gRPC 技术栈，引入成本高

### Decision 2: chat/stream 经由 history → voice 两级委派（而非直连 voice）

**方案选择**：
```
Flutter → /device/history/api/chat/stream
  → history DelegateTextChatStream() (HTTP SSE 客户端)
    → /voice/internal/api/text/chat/stream
      → voice HandleTranscriptForIntentStream()
```

**理由**：
- 与现有同步接口 `/device/history/api/chat` → `DelegateTextChat()` → `/voice/internal/api/text/chat` 的链路完全一致，保持架构一致性
- history 进程是设备域的统一入口，所有与设备相关的接口必须经过 history
- 符合服务边界约定：跨服务数据访问必须走服务接口契约

**备选方案**：
- Flutter 直连 voice：违反服务边界，设备号维度的鉴权、配额、风控逻辑重复实现

### Decision 3: tip/generate 直接在 device 控制器调 voice.TipStream()（而非走 history 委派）

**方案选择**：小贴士接口直接由网关绑定到 voice 服务控制器，不经过 history

**理由**：
- 小贴士是 voice 专属能力，与 history 领域（历史记录/建议/生日）无关
- 直接调用减少一次 HTTP 跳转，延迟更低
- Flutter 端已有 TipRepository 调用 `/device/tip/generate`，路径与 `/device/history/api/*` 天然区隔

### Decision 4: Flutter sendCommand() 返回 Stream（而非回调式 API）

**方案选择**：`sendCommand(String text) → Stream<String>`，按事件顺序产出 thinking/answer

**理由**：
- Dart Stream 是 Flutter 原生流式抽象，与 Riverpod 状态管理天然契合
- 与现有小贴士 `streamTip()` 返回 Stream 的模式一致，客户端代码风格统一
- 调用方可直接用 `await for` 或 `listen()` 消费，比回调更符合 Dart 习惯

### Decision 5: 同步接口保持不变，新旧接口并存

**方案选择**：`/device/history/api/chat`（同步）与 `/device/history/api/chat/stream`（流式）同时存在

**理由**：
- 向后兼容：可能有存量调用方依赖同步接口
- 兜底方案：流式接口异常时可回退到同步接口
- 零破坏性变更：符合 AGENTS.md 中"已有接口版本永远不可修改结构"的约定

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| SSE 在某些反向代理/CDN 下可能被缓冲 | 确保设置 `X-Accel-Buffering: no` 响应头；网关 Nginx 需配置 `proxy_buffering off` |
| Flutter 端 SSE 断连后无自动重连 | 流式对话为单次请求-响应模式（一次提问→流式返回），断连即可认为本次失败，无需重连；可由用户手动重试 |
| 流式接口无超时，客户端异常断开可能导致 Go 侧 goroutine 泄漏 | 使用 `ctx` 传递客户端连接状态，客户端断开时 `ctx.Done()` 触发，Python HTTP 请求和 SSE 写回均应检查 ctx |
| 两次 HTTP 跳转（history→voice）增加延迟 | 均为内网 HTTP，延迟 <10ms，相对 AI 推理（秒级）可忽略 |

## Migration Plan

1. **Phase 1 - Go 侧接口上线**：先部署 Go 侧 `/device/tip/generate` 和 `/device/history/api/chat/stream` 接口，同步接口保留
2. **Phase 2 - Flutter 小贴士接入**：Flutter 端确保 tip 相关代码调用 `/device/tip/generate`（已实现，验证即可）
3. **Phase 3 - Flutter 语音球接入**：Flutter 端 `sendCommand()` 切换到 `/chat/stream` 流式接口，加 thinking 展示
4. **Rollback**：任何阶段出现问题，Flutter 端回退到原同步接口即可，Go 侧接口保留不影响存量
