## Context

`CreateStreamASRSession` 已按 `chat` / `dictation` profile 读取 `sttChat` / `sttDictation`，并按 `provider` 分发到 `newDashScopeStreamASRSession` 或 `newBaiduStreamASRSession`。听写控制器 `voice_asr_ws.go` 只依赖 `StreamASRSession` 契约；`Commit` 在百炼侧对应 `finish-task`，与「前端 commit/end 定稿」兼容。

基线 `voice-chat-dashscope-stt` 曾要求听写保持百度；本变更显式推翻该默认。

## Goals / Non-Goals

**Goals:**

- 默认配置下 `/voice/asr/ws` 使用 DashScope 流式 ASR。
- 听写 WS 协议、事件类型、无服务端静音截句行为保持不变。
- DashScope 失败且配置了 `fallbackProvider=baidu` 时，降级会话使用 legacy `stt`（百度）配置。

**Non-Goals:**

- 不合并 `sttChat` 与 `sttDictation` 为同一配置块。
- 不改变 `/voice/chat/ws` 实时翻译 / LLM / TTS 行为。
- 不引入新 Redis、后台 ticker、新 HTTP 版本接口。
- 不强制开启 `fallbackProvider`（仓库默认可不设）。

## Decisions

1. **路径 A：配置切换 + 最小代码修正**  
   - 理由：分发逻辑已存在；改 `sttDictation.provider` 即可切换引擎。  
   - 备选 B（听写直接读 `sttChat`）会耦合两路运维参数，否决。

2. **听写 `sttDictation` 字段形态对齐 `sttChat`**  
   - `provider/model/streamEnabled/streamEndpoint/workspaceId/apiKey/speechNoiseThreshold/format/timeout/maxConcurrency`。  
   - 凭证仍走既有 `resolveDashScopeAPIKey` / `resolveDashScopeWorkspaceID`（env 兜底）。

3. **降级配置源改为 legacy `stt`**  
   - 当前代码在 DashScope 失败时用 `s.cfg.STTDictation` 建百度会话；听写改为 dashscope 后该路径会错误。  
   - 改为 `s.cfg.STT`（仍为百度兼容块）。

4. **规格同步**  
   - 修改 `voice-chat-dashscope-stt` 与 `voice-realtime-asr-ws` 中「听写=百度」的 Requirement/Scenario。

## Risks / Trade-offs

- [识别效果变化] → 听写近场场景可能与百度 `dev_pid=1537` 不同；可用配置回滚到 baidu。  
- [凭证缺失] → Workspace/API Key 未配时听写建连失败并返回 `stage=stt`（与对话一致，不静默假成功）。  
- [降级未配置] → 默认无 fallback；运维若需降级须显式设 `fallbackProvider: baidu` 并保证 legacy `stt` 百度凭证可用。

## Migration Plan

1. 部署前确认环境已有对话用的 DashScope 凭证（听写复用）。  
2. 发布含新 `voice-chat.shared.yaml` 的 voice-service。  
3. 回滚：`sttDictation.provider=baidu` 并恢复百度 stream 字段，重启。

## Open Questions

- （无）路径 A 与「不默认开 fallback」已确认。
