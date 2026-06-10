package voice

import (
	"context"

	"github.com/gogf/gf/v2/os/glog"
)

type voiceQuotaCheckedKey struct{}

// MarkVoiceQuotaChecked 标记本 utterance 已完成额度预检，避免 casual 流式重复 check。
func MarkVoiceQuotaChecked(ctx context.Context) context.Context {
	return context.WithValue(ctx, voiceQuotaCheckedKey{}, true)
}

func isVoiceQuotaChecked(ctx context.Context) bool {
	v, _ := ctx.Value(voiceQuotaCheckedKey{}).(bool)
	return v
}

// guardVoiceAIQuota 预检喂养 AI 额度；已 check 则仅返回 wxId。
func (s *VoiceService) guardVoiceAIQuota(ctx context.Context, deviceNo string) (int64, context.Context, error) {
	wxID, err := VoiceWxIDFromRequest(ctx, deviceNo)
	if err != nil {
		return 0, ctx, err
	}
	if isVoiceQuotaChecked(ctx) {
		return wxID, ctx, nil
	}
	if err := CheckVoiceAIQuota(ctx, wxID); err != nil {
		return 0, ctx, err
	}
	return wxID, MarkVoiceQuotaChecked(ctx), nil
}

func (s *VoiceService) consumeVoiceAIQuotaOnSuccess(ctx context.Context, wxID int64) {
	if wxID <= 0 {
		return
	}
	if err := ConsumeVoiceAIQuota(ctx, wxID); err != nil {
		glog.Warningf(ctx, "喂养 AI 额度扣减失败 wxId=%d err=%v", wxID, err)
	}
}
