package voice

import (
	"context"
	"os"
	"strings"

	"hello/internal/clients/device"
	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/frame/g"
)

type voiceWxIDCtxKey struct{}

// WithVoiceWxID 将网关注入的 wxId 写入 context，供对话链路可观测/鉴权上下文使用。
// 业务说明：额度/选模已退役；仍保留 wxId 注入以便日志与设备反查。
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

// VoiceWxIDFromRequest 从 ctx + deviceNo 解析 wxId。
func VoiceWxIDFromRequest(ctx context.Context, deviceNo string) (int64, error) {
	wxID := VoiceWxIDFromCtx(ctx)
	return ResolveVoiceWxID(ctx, wxID, deviceNo)
}

// HeaderInternalWxID 与 gateway 注入头一致（供 controller 引用常量）。
const HeaderInternalWxID = httpmeta.HeaderInternalWxId

// ParseHeaderWxID 解析 X-Internal-Wx-Id（委托 httpmeta）。
func ParseHeaderWxID(header string) int64 {
	return httpmeta.ParseHeaderWxID(header)
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
	return httpmeta.ConstantTimeEqual(strings.TrimSpace(password), expected)
}
