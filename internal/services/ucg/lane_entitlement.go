package ucg

import (
	"context"
	"strings"

	"hello/internal/services/aimodel"
	"hello/internal/clients/cash"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/os/glog"
)

// PolishEntitlement polish 路径 VIP∪额度原子结果。
type PolishEntitlement struct {
	Premium bool
	VIP     bool
	Snap    contracts.AIQuotaSnapshot
}

func isAccountVIP(ctx context.Context, wxID int64) bool {
	if wxID <= 0 {
		return false
	}
	vip, err := cash.RemoteIsVipByWxID(ctx, wxID)
	if err != nil {
		glog.Warningf(ctx, "[UcgEntitlement] VIP 查询降级 wxId=%d err=%v", wxID, err)
		return false
	}
	return vip
}

// ResolvePolishEntitlement VIP 或 polish 额度允许即为 premium；VIP 不计次。
func ResolvePolishEntitlement(ctx context.Context, wxID int64) PolishEntitlement {
	out := PolishEntitlement{}
	out.VIP = isAccountVIP(ctx, wxID)
	if out.VIP {
		out.Premium = true
		if snap, err := CheckPolishAIQuota(ctx, wxID); err == nil {
			snap.Allowed = true
			snap.Degraded = false
			out.Snap = snap
		} else {
			out.Snap = contracts.AIQuotaSnapshot{Allowed: true}
		}
		return out
	}
	if wxID <= 0 {
		out.Snap = contracts.AIQuotaSnapshot{Allowed: false, Degraded: true}
		return out
	}
	snap, err := CheckPolishAIQuota(ctx, wxID)
	if err != nil {
		out.Snap = contracts.AIQuotaSnapshot{Allowed: false, Degraded: true}
		return out
	}
	out.Snap = snap
	out.Premium = snap.Allowed
	return out
}

// ResolvePolishRuntimeProfile premium→正式；否则 free；free 空则 ok=false（调用方可跳过覆盖用默认上游语义）。
func ResolvePolishRuntimeProfile(ctx context.Context, wxID int64) (ent PolishEntitlement, profile aimodel.Profile, useProfile bool, err error) {
	ent = ResolvePolishEntitlement(ctx, wxID)
	base, err := aimodel.LoadProfile(ctx, aimodel.LanePolish)
	if err != nil {
		return ent, aimodel.Profile{}, false, err
	}
	if ent.Premium {
		return ent, base, true, nil
	}
	if base.HasFreeModel() {
		return ent, base.FreeAsRuntimeProfile(), true, nil
	}
	// free 空：不覆盖为硬编码 Degraded*；用正式 profile 的闸门参数但标记 useProfile 仍 true 走上游？
	// 按 design：Go 直调路径不以 Degraded 为真相源；此处仍 Acquire 正式 lane 闸门但 model 用种子？
	// 简化：free 空时使用 DefaultSeedProfile 作为免费模兜底（运维应配置 free；空则种子）。
	seed := aimodel.DefaultSeedProfile(aimodel.LanePolish)
	seed.MaxInFlight = base.MaxInFlight
	seed.MaxWaiters = base.MaxWaiters
	return ent, seed, true, nil
}

// ShouldConsumePolish 非 VIP 且额度允许时的成功才计次。
func (e PolishEntitlement) ShouldConsumePolish() bool {
	return !e.VIP && e.Premium
}

// applyVIPAllowed 供 App ai-quota 展示。
func applyVIPAllowed(ctx context.Context, wxID int64, snap contracts.AIQuotaSnapshot) contracts.AIQuotaSnapshot {
	if isAccountVIP(ctx, wxID) {
		snap.Allowed = true
		snap.Degraded = false
	}
	return snap
}

func trimFreePair(provider, model string) (string, string) {
	return strings.TrimSpace(provider), strings.TrimSpace(model)
}
