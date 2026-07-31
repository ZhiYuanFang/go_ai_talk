package voice

import (
	"context"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/os/glog"
)

type voiceQuotaCheckedKey struct{}
type voiceQuotaDegradedKey struct{}

// MarkVoiceQuotaChecked 标记本 utterance 已完成额度预检，避免重复 check。
func MarkVoiceQuotaChecked(ctx context.Context) context.Context {
	return context.WithValue(ctx, voiceQuotaCheckedKey{}, true)
}

func isVoiceQuotaChecked(ctx context.Context) bool {
	v, _ := ctx.Value(voiceQuotaCheckedKey{}).(bool)
	return v
}

// withVoiceQuotaDegraded 将喂养额度降速标记写入 context，供意图/成长建议等共用 profile 选择。
func withVoiceQuotaDegraded(ctx context.Context, degraded bool) context.Context {
	if !degraded {
		return ctx
	}
	return context.WithValue(ctx, voiceQuotaDegradedKey{}, true)
}

// isVoiceQuotaDegraded 本轮是否处于 voice_ai 额度用尽降速路径。
func isVoiceQuotaDegraded(ctx context.Context) bool {
	v, _ := ctx.Value(voiceQuotaDegradedKey{}).(bool)
	return v
}

// loadVoiceUnderstandingProfile 额度内用 Admin lane；degraded 强制种子智谱（不计次）。
func loadVoiceUnderstandingProfile(ctx context.Context) (aimodel.Profile, error) {
	if isVoiceQuotaDegraded(ctx) {
		return aimodel.DegradedVoiceUnderstandingProfile(), nil
	}
	return aimodel.LoadProfile(ctx, aimodel.LaneVoiceUnderstanding)
}

// guardVoiceAIQuota 预检喂养 AI 额度；返回 wxId、是否降速、更新后的 ctx。
// 用尽时 degraded=true 且不报错（对齐 clinic）；仅未登录等错误返回 err。
func (s *VoiceService) guardVoiceAIQuota(ctx context.Context, deviceNo string) (wxID int64, degraded bool, newCtx context.Context, err error) {
	wxID, err = VoiceWxIDFromRequest(ctx, deviceNo)
	if err != nil {
		return 0, true, ctx, err
	}
	if isVoiceQuotaChecked(ctx) {
		return wxID, isVoiceQuotaDegraded(ctx), ctx, nil
	}
	snap, err := CheckVoiceAIQuotaSnapshot(ctx, wxID)
	if err != nil {
		return 0, false, ctx, err
	}
	// 用尽：允许继续，标记 degraded，强制后续 Python 使用种子模型
	degraded = snap.Degraded && !snap.Allowed
	newCtx = MarkVoiceQuotaChecked(ctx)
	if degraded {
		newCtx = withVoiceQuotaDegraded(newCtx, true)
		glog.Infof(ctx, "[VoiceAIQuota] 额度用尽走降速。wxId=%d used=%d limit=%d deviceNo=%s",
			wxID, snap.Used, snap.Limit, deviceNo)
	}
	return wxID, degraded, newCtx, nil
}

func (s *VoiceService) consumeVoiceAIQuotaOnSuccess(ctx context.Context, wxID int64) {
	if wxID <= 0 {
		return
	}
	if err := ConsumeVoiceAIQuota(ctx, wxID); err != nil {
		glog.Warningf(ctx, "喂养 AI 额度扣减失败 wxId=%d err=%v", wxID, err)
	}
}
