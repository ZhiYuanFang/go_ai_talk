package voice

import (
	"context"
	"strings"

	"hello/internal/services/aimodel"
	"hello/internal/clients/cash"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/os/glog"
)

// ModelPrivilege 选模/计次通道特权（与账号 VIP、月度额度共同构成 premium）。
type ModelPrivilege int

const (
	// PrivilegeAccount 账号通道：按 cash VIP ∪ feature 额度判定。
	PrivilegeAccount ModelPrivilege = iota
	// PrivilegeHardware 硬件通道（/voice/chat/ws、MCP/internal text chat）：强制 premium 且不计次。
	PrivilegeHardware
)

type modelPrivilegeCtxKey struct{}

// WithModelPrivilege 将选模特权写入 context（入口打标）。
func WithModelPrivilege(ctx context.Context, p ModelPrivilege) context.Context {
	return context.WithValue(ctx, modelPrivilegeCtxKey{}, p)
}

// ModelPrivilegeFromCtx 读取特权；未设置视为 Account。
func ModelPrivilegeFromCtx(ctx context.Context) ModelPrivilege {
	v, _ := ctx.Value(modelPrivilegeCtxKey{}).(ModelPrivilege)
	return v
}

// LaneEntitlement 统一权益快照（原子决策结果）。
type LaneEntitlement struct {
	Premium  bool
	VIP      bool
	Hardware bool
	Snapshot contracts.AIQuotaSnapshot
}

// PythonModelCfgPtr 将选模结果转为可 omit 的 Python 模型配置；nil 表示不传 model。
func PythonModelCfgPtr(provider, name string, maxInFlight int) *PythonModelCfg {
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	if provider == "" || name == "" {
		return nil
	}
	if maxInFlight <= 0 {
		maxInFlight = 1
	}
	return &PythonModelCfg{Provider: provider, Name: name, MaxInFlight: maxInFlight}
}

// isAccountVIP 经 cash 读 VIP；失败降级 false + Warning。
func isAccountVIP(ctx context.Context, wxID int64) bool {
	if wxID <= 0 {
		return false
	}
	vip, err := cash.RemoteIsVipByWxID(ctx, wxID)
	if err != nil {
		glog.Warningf(ctx, "[LaneEntitlement] VIP 查询降级为非 VIP wxId=%d err=%v", wxID, err)
		return false
	}
	return vip
}

// ResolveLaneEntitlement 原子判定 premium：VIP ∨ 硬件 ∨ quota.Allowed。
// VIP/硬件时 Snapshot.Allowed=true、Degraded=false；quota 读失败时按非额度处理（Premium 仍可能因 VIP/硬件为 true）。
func ResolveLaneEntitlement(ctx context.Context, wxID int64, feature contracts.AIQuotaFeature, privilege ModelPrivilege) LaneEntitlement {
	out := LaneEntitlement{}
	if privilege == PrivilegeHardware {
		out.Hardware = true
		out.Premium = true
		out.Snapshot = contracts.AIQuotaSnapshot{Allowed: true, Degraded: false}
		return out
	}
	out.VIP = isAccountVIP(ctx, wxID)
	if out.VIP {
		out.Premium = true
		// VIP：读真实 used/limit 供展示，但强制 Allowed。
		if wxID > 0 {
			if snap, err := CheckVoiceAIQuotaStore(ctx, wxID, feature); err == nil {
				snap.Allowed = true
				snap.Degraded = false
				out.Snapshot = snap
			} else {
				out.Snapshot = contracts.AIQuotaSnapshot{Allowed: true, Degraded: false}
			}
		} else {
			out.Snapshot = contracts.AIQuotaSnapshot{Allowed: true, Degraded: false}
		}
		return out
	}
	if wxID <= 0 {
		out.Snapshot = contracts.AIQuotaSnapshot{Allowed: false, Degraded: true}
		return out
	}
	snap, err := CheckVoiceAIQuotaStore(ctx, wxID, feature)
	if err != nil {
		glog.Warningf(ctx, "[LaneEntitlement] 额度查询失败按无额度 wxId=%d feature=%s err=%v", wxID, feature, err)
		out.Snapshot = contracts.AIQuotaSnapshot{Allowed: false, Degraded: true}
		return out
	}
	out.Snapshot = snap
	out.Premium = snap.Allowed
	return out
}

// ResolveLaneModel 原子选模：premium→正式模；否则 free；free 空→nil（Python omit）。
// 返回的 Profile 用于 Acquire（有模型时）；modelCfg 供 Python；nil modelCfg 表示 omit。
func ResolveLaneModel(ctx context.Context, wxID int64, lane aimodel.Lane, feature contracts.AIQuotaFeature, privilege ModelPrivilege) (ent LaneEntitlement, runtime aimodel.Profile, modelCfg *PythonModelCfg, err error) {
	ent = ResolveLaneEntitlement(ctx, wxID, feature, privilege)
	base, err := aimodel.LoadProfile(ctx, lane)
	if err != nil {
		base = aimodel.DefaultSeedProfile(lane)
		err = nil
	}
	if ent.Premium {
		runtime = base
		modelCfg = PythonModelCfgPtr(string(base.Provider), base.Model, base.MaxInFlight)
		return ent, runtime, modelCfg, nil
	}
	if base.HasFreeModel() {
		runtime = base.FreeAsRuntimeProfile()
		modelCfg = PythonModelCfgPtr(string(runtime.Provider), runtime.Model, runtime.MaxInFlight)
		return ent, runtime, modelCfg, nil
	}
	// free 空：omit model；runtime 仅占位，调用方不应 Acquire 上游池
	runtime = base
	modelCfg = nil
	return ent, runtime, modelCfg, nil
}

// ShouldConsumeOnSuccess premium 因 VIP/硬件时不计次；仅非 VIP 且非硬件且额度允许路径的成功才计次。
func (e LaneEntitlement) ShouldConsumeOnSuccess() bool {
	if e.VIP || e.Hardware {
		return false
	}
	return e.Premium && e.Snapshot.Allowed
}

// ConsumeVoiceFeatureIfNeeded 成功路径计次；VIP/硬件 no-op。
func ConsumeVoiceFeatureIfNeeded(ctx context.Context, wxID int64, feature contracts.AIQuotaFeature, ent LaneEntitlement) {
	if !ent.ShouldConsumeOnSuccess() || wxID <= 0 {
		return
	}
	if _, err := ConsumeVoiceAIQuotaStore(ctx, wxID, feature); err != nil {
		glog.Warningf(ctx, "[LaneEntitlement] 额度扣减失败 wxId=%d feature=%s err=%v", wxID, feature, err)
	}
}
