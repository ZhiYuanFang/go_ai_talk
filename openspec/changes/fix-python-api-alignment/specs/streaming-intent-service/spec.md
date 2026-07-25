## ADDED Requirements

### Requirement: Go 侧 voice service 提供流式意图分析入口
VoiceService SHALL 提供独立的流式意图分析服务层方法，供 MCP 服务和纯文字场景调用，不得影响现有 TTS 非流式路径。

#### Scenario: 流式意图分析调用路径
- **WHEN** 调用方（MCP 或纯文字接口）调用流式意图分析方法
- **THEN** SHALL 执行设备注册校验、文本风控校验
- **AND** SHALL 调用 PythonAIClient.AnalyzeIntentStream 发起流式请求
- **AND** SHALL 累积流式结果中的意图分析 JSON（answer 事件）
- **AND** SHALL 将意图分析结果按与非流式相同的逻辑处理（喂养事件落库、对话回复等）

#### Scenario: TTS 非流式路径不受影响
- **WHEN** 调用现有 `HandleTranscriptForStreaming` 方法（TTS 场景）
- **THEN** SHALL 继续使用非流式 `/v1/analyze/intent` 接口
- **AND** 行为 SHALL 与本次变更前完全一致
