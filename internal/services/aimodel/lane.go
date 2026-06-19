package aimodel

// Lane 标识一条 LLM 业务车道；业务代码只依赖 Lane，不硬编码 model 字符串。
type Lane string

const (
	// LaneVoiceUnderstanding 喂养语音全部 LLM（意图、闲聊流式、成长建议等）。
	LaneVoiceUnderstanding Lane = "voiceUnderstanding"
	// LaneClinic 胖宝诊疗 LLM。
	LaneClinic Lane = "clinic"
	// LanePolish UCG 润笔多模态 LLM。
	LanePolish Lane = "polish"
)

// Provider 上游大模型供应商标识。
type Provider string

const (
	ProviderZhipu     Provider = "zhipu"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderDashScope Provider = "dashscope"
)
