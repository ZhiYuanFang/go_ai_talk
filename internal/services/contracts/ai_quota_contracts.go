package contracts

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

// AIQuotaFeature 额度维度。
type AIQuotaFeature string

const (
	AIQuotaPolish   AIQuotaFeature = "polish"
	AIQuotaVoiceAI  AIQuotaFeature = "voice_ai"
	AIQuotaClinicAI AIQuotaFeature = "clinic_ai"
)

// AIQuotaSnapshot 某 feature 当月 used/limit 快照。
type AIQuotaSnapshot struct {
	Used     int  `json:"used"`
	Limit    int  `json:"limit"`
	Allowed  bool `json:"allowed"`
	Degraded bool `json:"degraded"` // 额度用尽且允许降速 fallback（polish / clinic_ai / voice_ai）
}

// VoiceAIQuotaDefaultDTO voice 域全局默认（voice_ai + clinic_ai）。
type VoiceAIQuotaDefaultDTO struct {
	VoiceAiMonthlyLimit  int   `json:"voiceAiMonthlyLimit"`
	ClinicAiMonthlyLimit int   `json:"clinicAiMonthlyLimit"`
	UpdatedAt            int64 `json:"updatedAt"`
}

// VoiceAIQuotaUserOverrideDTO voice 域 per-wxId override。
type VoiceAIQuotaUserOverrideDTO struct {
	WxId                 int64 `json:"wxId"`
	VoiceAiMonthlyLimit  *int  `json:"voiceAiMonthlyLimit,omitempty"`
	ClinicAiMonthlyLimit *int  `json:"clinicAiMonthlyLimit,omitempty"`
	UpdatedAt            int64 `json:"updatedAt"`
}

// PolishAIQuotaDefaultDTO ucg 域全局润笔默认。
type PolishAIQuotaDefaultDTO struct {
	PolishMonthlyLimit int   `json:"polishMonthlyLimit"`
	UpdatedAt          int64 `json:"updatedAt"`
}

// PolishAIQuotaUserOverrideDTO ucg 域 per-wxId 润笔 override。
type PolishAIQuotaUserOverrideDTO struct {
	WxId               int64 `json:"wxId"`
	PolishMonthlyLimit *int  `json:"polishMonthlyLimit,omitempty"`
	UpdatedAt          int64 `json:"updatedAt"`
}

// VoiceAIQuotaAppStatus App 读 API：voice 域 voiceAi + clinicAi。
type VoiceAIQuotaAppStatus struct {
	VoiceAi  AIQuotaSnapshot `json:"voiceAi"`
	ClinicAi AIQuotaSnapshot `json:"clinicAi"`
}

// PolishAIQuotaAppStatus App 读 API：ucg 域 polish。
type PolishAIQuotaAppStatus struct {
	Polish AIQuotaSnapshot `json:"polish"`
}
