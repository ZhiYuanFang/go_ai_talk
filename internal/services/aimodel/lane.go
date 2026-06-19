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
	// LaneSimText 模拟用户文本生成（昵称/文案/聊天）。
	LaneSimText Lane = "simText"
	// LaneSimVision 模拟用户读帖评论多模态。
	LaneSimVision Lane = "simVision"
	// LaneSimImageGen 模拟用户 CogView 生图。
	LaneSimImageGen Lane = "simImageGen"
	// LaneSimVideoGen 模拟用户 CogVideoX 生视频。
	LaneSimVideoGen Lane = "simVideoGen"
)

// Provider 上游大模型供应商标识。
type Provider string

const (
	ProviderZhipu     Provider = "zhipu"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderDashScope Provider = "dashscope"
)
