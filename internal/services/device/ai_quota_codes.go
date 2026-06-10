package device

import (
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
)

const (
	// CodeAINotLoggedIn wxId=0 或未绑定账号时禁止 AI 功能。
	CodeAINotLoggedIn = 40301
	// CodeAIQuotaExhausted 当月额度已用尽。
	CodeAIQuotaExhausted = 40302
)

var (
	// ErrAINotLoggedIn 未登录账号不可使用 AI。
	ErrAINotLoggedIn = errors.New("请先登录账号")
	// ErrAIQuotaExhausted 本月 AI 额度已用尽。
	ErrAIQuotaExhausted = errors.New("本月额度已用完")
)

// GCodeAINotLoggedIn 供 controller 返回稳定业务码。
func GCodeAINotLoggedIn() gcode.Code {
	return gcode.New(CodeAINotLoggedIn, ErrAINotLoggedIn.Error(), nil)
}

// GCodeAIQuotaExhausted 供 controller 返回稳定业务码。
func GCodeAIQuotaExhausted() gcode.Code {
	return gcode.New(CodeAIQuotaExhausted, ErrAIQuotaExhausted.Error(), nil)
}

// AIQuotaFeature 额度维度：润笔 / 喂养 AI。
type AIQuotaFeature string

const (
	AIQuotaPolish  AIQuotaFeature = "polish"
	AIQuotaVoiceAI AIQuotaFeature = "voice_ai"
)

// AIQuotaSnapshot 某 feature 当月 used/limit 快照。
type AIQuotaSnapshot struct {
	Used    int  `json:"used"`
	Limit   int  `json:"limit"`
	Allowed bool `json:"allowed"`
}

// AIQuotaDefaultDTO 全局默认配置。
type AIQuotaDefaultDTO struct {
	PolishMonthlyLimit  int   `json:"polishMonthlyLimit"`
	VoiceAiMonthlyLimit int   `json:"voiceAiMonthlyLimit"`
	UpdatedAt           int64 `json:"updatedAt"`
}

// AIQuotaUserOverrideDTO per-wxId override。
type AIQuotaUserOverrideDTO struct {
	WxId                int64 `json:"wxId"`
	PolishMonthlyLimit  *int  `json:"polishMonthlyLimit,omitempty"`
	VoiceAiMonthlyLimit *int  `json:"voiceAiMonthlyLimit,omitempty"`
	UpdatedAt           int64 `json:"updatedAt"`
}

// AIQuotaAppStatus App 读 API 响应片段。
type AIQuotaAppStatus struct {
	Polish  AIQuotaSnapshot `json:"polish"`
	VoiceAi AIQuotaSnapshot `json:"voiceAi"`
}
