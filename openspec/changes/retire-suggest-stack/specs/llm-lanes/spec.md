# Spec Delta

## MODIFIED Requirements

### Requirement: voiceUnderstanding lane SHALL 覆盖喂养语音全部 LLM 路径

下列路径 MUST 经 `LaneVoiceUnderstanding` 调用上游：统一意图解析、实体/动作抽取、闲聊直答、闲聊流式 LLM 段、历史问答，以及 `event_child_pending` 中的实体抽取。纯 ASR、纯 TTS、规则回复与模式切换（不触发 LLM）MUST NOT 经过该 lane。系统 MUST NOT 再将「成长建议」作为须经该 lane 的正式产品 LLM 路径。

#### Scenario: 成长建议走 voiceUnderstanding

- **WHEN** 用户触发成长建议且将调用 LLM
- **THEN** 系统 MUST NOT 再将成长建议作为正式产品 LLM 路径；MUST NOT 仅为该已退役功能发起独立的 `voiceUnderstanding` 调用

#### Scenario: 纯 ASR 不占用 LLM 闸门

- **WHEN** 用户仅进行语音转写且无 LLM 调用
- **THEN** 系统 MUST NOT 调用 `aimodel.Acquire`
