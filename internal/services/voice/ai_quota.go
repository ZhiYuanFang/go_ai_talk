package voice

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"hello/internal/services/device"
	"hello/internal/services/gatewayapp"
)

// VoiceAIQuotaError 喂养 AI 额度/登录类错误，供 WS 层映射为固定 code。
type VoiceAIQuotaError struct {
	Code    int
	Message string
}

func (e *VoiceAIQuotaError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

type voiceWxIDCtxKey struct{}

// WithVoiceWxID 将网关注入的 wxId 写入 context，供对话链路额度校验。
func WithVoiceWxID(ctx context.Context, wxID int64) context.Context {
	if wxID <= 0 {
		return ctx
	}
	return context.WithValue(ctx, voiceWxIDCtxKey{}, wxID)
}

// VoiceWxIDFromCtx 读取 context 中的 wxId；未设置返回 0。
func VoiceWxIDFromCtx(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	v, _ := ctx.Value(voiceWxIDCtxKey{}).(int64)
	return v
}

// ResolveVoiceWxID 优先 Header wxId，否则 deviceNo 反查。
func ResolveVoiceWxID(ctx context.Context, headerWxID int64, deviceNo string) (int64, error) {
	if headerWxID > 0 {
		return headerWxID, nil
	}
	wxID, err := device.AIQuotaHTTP().RemoteWxIDByDeviceNo(ctx, strings.TrimSpace(deviceNo))
	if err != nil {
		return 0, err
	}
	return wxID, nil
}

// ParseHeaderWxID 解析 X-Internal-Wx-Id。
func ParseHeaderWxID(header string) int64 {
	s := strings.TrimSpace(header)
	if s == "" {
		return 0
	}
	wxID, err := strconv.ParseInt(s, 10, 64)
	if err != nil || wxID <= 0 {
		return 0
	}
	return wxID
}

// CheckVoiceAIQuota 喂养 AI 预检。
func CheckVoiceAIQuota(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return &VoiceAIQuotaError{Code: device.CodeAINotLoggedIn, Message: device.ErrAINotLoggedIn.Error()}
	}
	snap, err := device.AIQuotaHTTP().RemoteCheck(ctx, wxID, device.AIQuotaVoiceAI)
	if err != nil {
		if errors.Is(err, device.ErrAINotLoggedIn) {
			return &VoiceAIQuotaError{Code: device.CodeAINotLoggedIn, Message: err.Error()}
		}
		if errors.Is(err, device.ErrAIQuotaExhausted) {
			return &VoiceAIQuotaError{Code: device.CodeAIQuotaExhausted, Message: err.Error()}
		}
		return err
	}
	if !snap.Allowed {
		return &VoiceAIQuotaError{Code: device.CodeAIQuotaExhausted, Message: device.ErrAIQuotaExhausted.Error()}
	}
	return nil
}

// ConsumeVoiceAIQuota 喂养 AI 成功扣减。
func ConsumeVoiceAIQuota(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return nil
	}
	_, err := device.AIQuotaHTTP().RemoteConsume(ctx, wxID, device.AIQuotaVoiceAI)
	if err != nil {
		if errors.Is(err, device.ErrAIQuotaExhausted) {
			return &VoiceAIQuotaError{Code: device.CodeAIQuotaExhausted, Message: err.Error()}
		}
		return err
	}
	return nil
}

// VoiceWxIDFromRequest 从 ctx + deviceNo 解析 wxId。
func VoiceWxIDFromRequest(ctx context.Context, deviceNo string) (int64, error) {
	wxID := VoiceWxIDFromCtx(ctx)
	return ResolveVoiceWxID(ctx, wxID, deviceNo)
}

// HeaderInternalWxID 与 gateway 注入头一致（供 controller 引用常量）。
const HeaderInternalWxID = gatewayapp.HeaderInternalWxId
