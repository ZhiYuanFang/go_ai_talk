## ADDED Requirements

### Requirement: voice feeding LLM SHALL 经 voiceUnderstanding lane 闸门

voice-service 对 feature `voice_ai` 的额度 check 通过后、调用任何喂养 voice LLM 前 MUST 经 `LaneVoiceUnderstanding` 闸门。队列满 MUST 返回 WS code **50301**。本要求与 `ai-monthly-quota` 中喂养 AI 条款一致，强调全部 LLM 路径（含 casual 流式与成长建议）无遗漏。

#### Scenario: 闲聊流式占用 voiceUnderstanding 闸门

- **WHEN** commit 后进入 `StreamCasualReplyWithBaiduTTS` 的 LLM 段
- **THEN** MUST 使用 voiceUnderstanding profile 的 model 闸门
