## Why

当前母婴喂养 AI 能力已全部迁移至 Python 微服务（LangGraph 编排），Go 侧仍保留了闲聊模式业务代码和 DeepSeek/LLM 直连兜底代码。这些冗余代码增加了维护成本、混淆了调用链路，且与 Python 侧的能力存在重复。清理后可简化架构、减少维护负担，并使 AI 能力调用路径统一由 Python 服务承载。

## What Changes

- **删除闲聊模式业务代码**：移除母婴/闲聊双模式切换逻辑，所有对话统一走母婴喂养模式
- **删除 AI 能力兜底代码**：移除 Go 侧 DeepSeek/LLM 直连兜底，所有 AI 推理统一走 Python 微服务接口
- **Python 不可用时返回固定提示语降级**：Python 服务不可用时，返回"AI 服务暂时不可用，请稍后再试"，不再回退到 DeepSeek 直连
- **删除死代码**：一并移除 `callDeepSeekActionExtract` 等未被调用的死代码
- **BREAKING**：`HandleTranscriptForStreaming` 接口签名简化，移除 `mode` 和 `needCasualStream` 返回值（本次特例，不做版本升级）

## Capabilities

### New Capabilities

（无新增能力，本次为代码清理与架构简化）

### Modified Capabilities

- `voice-chat`：语音对话能力移除闲聊模式，AI 推理统一依赖 Python 微服务，移除 DeepSeek 兜底

## Impact

- **影响代码范围**：`internal/services/voice/` 下多个文件（chat_enums.go、voice_chat_understanding.go、voice_chat_deepseek.go、voice_chat.go、clinic_llm.go），`internal/controller/voice_ws.go`，`internal/services/contracts/runtime_contracts.go`
- **接口变更**：`HandleTranscriptForStreaming`、`HandleTranscriptChatOnly` 签名变更；`StreamCasualReplyWithBaiduTTS` 方法删除；`VoiceContract` 接口变更
- **依赖变更**：不再直接依赖 aimodel 的 DeepSeek/LLM 直连（entity extract 除外，业务兜底保留）
- **运行时影响**：Python 服务不可用时，喂养 AI 直接返回固定降级提示语，而非走 DeepSeek 兜底
