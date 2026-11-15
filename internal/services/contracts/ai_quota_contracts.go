package contracts

import (
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
)

// 历史兼容业务码：额度/clinic/care-alert 能力已退役，但仍可能出现在旧客户端或
// history 本地反代对 voice 响应的透传映射中，故保留稳定数字码。

const (
	// CodeAINotLoggedIn wxId=0 或未绑定账号时禁止 AI 功能。
	CodeAINotLoggedIn = 40301
	// CodeAIQuotaExhausted 当月额度已用尽（能力已退役；保留码值兼容）。
	CodeAIQuotaExhausted = 40302
)

var (
	// ErrAINotLoggedIn 未登录账号不可使用 AI。
	ErrAINotLoggedIn = errors.New("请先登录账号")
	// ErrAIQuotaExhausted 本月 AI 额度已用尽（能力已退役；保留文案兼容）。
	ErrAIQuotaExhausted = errors.New("本月额度已用完")
)

// GCodeAINotLoggedIn 供 controller / history 反代返回稳定业务码。
func GCodeAINotLoggedIn() gcode.Code {
	return gcode.New(CodeAINotLoggedIn, ErrAINotLoggedIn.Error(), nil)
}

// GCodeAIQuotaExhausted 供 history 反代透传旧额度耗尽码。
func GCodeAIQuotaExhausted() gcode.Code {
	return gcode.New(CodeAIQuotaExhausted, ErrAIQuotaExhausted.Error(), nil)
}
