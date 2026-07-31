package voice

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"hello/internal/services/contracts"
	"hello/internal/services/device"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/frame/g"
)

// VoiceAIQuotaError 喂养/胖宝 AI 额度/登录类错误，供 WS 层映射为固定 code。
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

// ResolveVoiceWxID 优先 Header wxId，否则经 device user 域 internal API 按 deviceNo 反查。
func ResolveVoiceWxID(ctx context.Context, headerWxID int64, deviceNo string) (int64, error) {
	if headerWxID > 0 {
		return headerWxID, nil
	}
	wxID, err := device.RemoteWxIDByDeviceNo(ctx, strings.TrimSpace(deviceNo))
	if err != nil {
		return 0, err
	}
	return wxID, nil
}

// CheckVoiceAIQuota 喂养 AI 预检；wxId≤0 返回 40301；用尽且非 degraded 才 40302。
// 对话主路径请用 CheckVoiceAIQuotaSnapshot + guard 的 degraded 分支。
func CheckVoiceAIQuota(ctx context.Context, wxID int64) error {
	snap, err := CheckVoiceAIQuotaSnapshot(ctx, wxID)
	if err != nil {
		return err
	}
	if !snap.Allowed && !snap.Degraded {
		return &VoiceAIQuotaError{Code: contracts.CodeAIQuotaExhausted, Message: contracts.ErrAIQuotaExhausted.Error()}
	}
	return nil
}

// CheckVoiceAIQuotaSnapshot 喂养 AI 额度快照；用尽时 Degraded=true 而非 40302，供对话降速路径。
func CheckVoiceAIQuotaSnapshot(ctx context.Context, wxID int64) (contracts.AIQuotaSnapshot, error) {
	if wxID <= 0 {
		return contracts.AIQuotaSnapshot{}, &VoiceAIQuotaError{Code: contracts.CodeAINotLoggedIn, Message: contracts.ErrAINotLoggedIn.Error()}
	}
	snap, err := CheckVoiceAIQuotaStore(ctx, wxID, contracts.AIQuotaVoiceAI)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, mapQuotaStoreErr(err)
	}
	return snap, nil
}

// CheckClinicAIQuota 胖宝 AI 预检；wxId≤0 返回 40301，用尽返回 40302（供非 clinic 路径）。
func CheckClinicAIQuota(ctx context.Context, wxID int64) error {
	snap, err := CheckClinicAIQuotaSnapshot(ctx, wxID)
	if err != nil {
		return err
	}
	if !snap.Allowed && !snap.Degraded {
		return &VoiceAIQuotaError{Code: contracts.CodeAIQuotaExhausted, Message: contracts.ErrAIQuotaExhausted.Error()}
	}
	return nil
}

// CheckClinicAIQuotaSnapshot 胖宝诊疗额度快照；用尽时 Degraded=true 而非错误。
func CheckClinicAIQuotaSnapshot(ctx context.Context, wxID int64) (contracts.AIQuotaSnapshot, error) {
	if wxID <= 0 {
		return contracts.AIQuotaSnapshot{}, &VoiceAIQuotaError{Code: contracts.CodeAINotLoggedIn, Message: contracts.ErrAINotLoggedIn.Error()}
	}
	snap, err := CheckVoiceAIQuotaStore(ctx, wxID, contracts.AIQuotaClinicAI)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, mapQuotaStoreErr(err)
	}
	return snap, nil
}

// ConsumeClinicAIQuota 胖宝 AI 流式成功完成后扣减 clinic_ai 额度。
func ConsumeClinicAIQuota(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return nil
	}
	_, err := ConsumeVoiceAIQuotaStore(ctx, wxID, contracts.AIQuotaClinicAI)
	return mapQuotaStoreErr(err)
}

// ConsumeVoiceAIQuota 喂养 AI 成功扣减。
func ConsumeVoiceAIQuota(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return nil
	}
	_, err := ConsumeVoiceAIQuotaStore(ctx, wxID, contracts.AIQuotaVoiceAI)
	return mapQuotaStoreErr(err)
}

func mapQuotaStoreErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, contracts.ErrAINotLoggedIn) {
		return &VoiceAIQuotaError{Code: contracts.CodeAINotLoggedIn, Message: err.Error()}
	}
	if errors.Is(err, contracts.ErrAIQuotaExhausted) {
		return &VoiceAIQuotaError{Code: contracts.CodeAIQuotaExhausted, Message: err.Error()}
	}
	return err
}

// VoiceWxIDFromRequest 从 ctx + deviceNo 解析 wxId。
func VoiceWxIDFromRequest(ctx context.Context, deviceNo string) (int64, error) {
	wxID := VoiceWxIDFromCtx(ctx)
	return ResolveVoiceWxID(ctx, wxID, deviceNo)
}

// HeaderInternalWxID 与 gateway 注入头一致（供 controller 引用常量）。
const HeaderInternalWxID = gatewayapp.HeaderInternalWxId

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

// VoiceAdminPassword 读取 voice admin 口令（env 优先）。
func VoiceAdminPassword(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("VOICE_ADMIN_PASSWORD")); v != "" {
		return v
	}
	val, err := g.Cfg().Get(ctx, "voice.admin.password")
	if err != nil || val == nil || val.IsEmpty() {
		return ""
	}
	return strings.TrimSpace(val.String())
}

// VerifyVoiceAdminPassword 校验 X-Admin-Password。
func VerifyVoiceAdminPassword(ctx context.Context, password string) bool {
	expected := VoiceAdminPassword(ctx)
	if expected == "" {
		return false
	}
	return gatewayapp.ConstantTimeEqual(strings.TrimSpace(password), expected)
}
