package voice

import (
	"context"

	"hello/internal/services/aimodel"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/os/glog"
)

type voiceQuotaCheckedKey struct{}
type voiceLaneEntitlementKey struct{}

// MarkVoiceQuotaChecked 标记本 utterance 已完成额度/权益预检，避免重复 check。
func MarkVoiceQuotaChecked(ctx context.Context) context.Context {
	return context.WithValue(ctx, voiceQuotaCheckedKey{}, true)
}

func isVoiceQuotaChecked(ctx context.Context) bool {
	v, _ := ctx.Value(voiceQuotaCheckedKey{}).(bool)
	return v
}

func withLaneEntitlement(ctx context.Context, ent LaneEntitlement) context.Context {
	return context.WithValue(ctx, voiceLaneEntitlementKey{}, ent)
}

// LaneEntitlementFromCtx 读取本轮喂养权益；未设置返回零值。
func LaneEntitlementFromCtx(ctx context.Context) (LaneEntitlement, bool) {
	v, ok := ctx.Value(voiceLaneEntitlementKey{}).(LaneEntitlement)
	return v, ok
}

// resolveVoiceUnderstandingModel 经公共出口为喂养意图选模（含硬件特权）。
func resolveVoiceUnderstandingModel(ctx context.Context, wxID int64) (ent LaneEntitlement, runtime aimodel.Profile, modelCfg *PythonModelCfg, err error) {
	priv := ModelPrivilegeFromCtx(ctx)
	return ResolveLaneModel(ctx, wxID, aimodel.LaneVoiceUnderstanding, contracts.AIQuotaVoiceAI, priv)
}

// guardVoiceAIQuota 预检喂养 AI 权益；返回 wxId、是否应计次、更新后的 ctx。
// 未登录/反查失败按 wxId=0 非 premium 放行；硬件特权强制 premium 不计次。不再因额度返回 40302。
func (s *VoiceService) guardVoiceAIQuota(ctx context.Context, deviceNo string) (wxID int64, shouldConsume bool, newCtx context.Context, err error) {
	wxID, resolveErr := VoiceWxIDFromRequest(ctx, deviceNo)
	if resolveErr != nil {
		wxID = VoiceWxIDFromCtx(ctx)
		glog.Warningf(ctx, "[VoiceAIQuota] wxId 解析降级为 %d deviceNo=%s err=%v", wxID, deviceNo, resolveErr)
	}
	if isVoiceQuotaChecked(ctx) {
		ent, ok := LaneEntitlementFromCtx(ctx)
		if !ok {
			ent = ResolveLaneEntitlement(ctx, wxID, contracts.AIQuotaVoiceAI, ModelPrivilegeFromCtx(ctx))
		}
		return wxID, ent.ShouldConsumeOnSuccess(), ctx, nil
	}
	priv := ModelPrivilegeFromCtx(ctx)
	ent := ResolveLaneEntitlement(ctx, wxID, contracts.AIQuotaVoiceAI, priv)
	newCtx = withLaneEntitlement(MarkVoiceQuotaChecked(ctx), ent)
	if ent.Hardware {
		glog.Infof(ctx, "[VoiceAIQuota] 硬件特权 premium 不计次 deviceNo=%s", deviceNo)
	} else if ent.VIP {
		glog.Infof(ctx, "[VoiceAIQuota] VIP premium 不计次 wxId=%d deviceNo=%s", wxID, deviceNo)
	} else if !ent.Premium {
		glog.Infof(ctx, "[VoiceAIQuota] 非 premium 走 free/omit wxId=%d used=%d limit=%d deviceNo=%s",
			wxID, ent.Snapshot.Used, ent.Snapshot.Limit, deviceNo)
	}
	return wxID, ent.ShouldConsumeOnSuccess(), newCtx, nil
}

func (s *VoiceService) consumeVoiceAIQuotaOnSuccess(ctx context.Context, wxID int64) {
	ent, ok := LaneEntitlementFromCtx(ctx)
	if ok {
		ConsumeVoiceFeatureIfNeeded(ctx, wxID, contracts.AIQuotaVoiceAI, ent)
		return
	}
	if wxID <= 0 || ModelPrivilegeFromCtx(ctx) == PrivilegeHardware {
		return
	}
	if isAccountVIP(ctx, wxID) {
		return
	}
	if err := ConsumeVoiceAIQuota(ctx, wxID); err != nil {
		glog.Warningf(ctx, "喂养 AI 额度扣减失败 wxId=%d err=%v", wxID, err)
	}
}
