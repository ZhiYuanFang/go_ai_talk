package cash

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// GrowthTrajectoryAccessResult 成长轨迹可看合成（账号开通 ∨ VIP ∨ 试用未用；无喂养门闸）。
type GrowthTrajectoryAccessResult struct {
	Allowed              bool  `json:"allowed"`
	FeatureActive        bool  `json:"featureActive"` // 已开通：用户权益 ∨ VIP（不含 soft trial）
	TrialAvailable       bool  `json:"trialAvailable"`
	EntitlementExpiresAt int64 `json:"entitlementExpiresAt,omitempty"`
}

// GetGrowthTrajectoryAccess 合成成长轨迹是否可看（权威在 cash；voice 经 internal 调用）。
//
// 业务：featureActive = 未过期账号维权益 ∨ isVip；Allowed = featureActive ∨ trialUnused。
// deviceNo 保留兼容调用方校验非空；开通判定不依赖设备。
// VIP 查询失败当作非 VIP（仅认 user entitlement）。
func GetGrowthTrajectoryAccess(ctx context.Context, deviceNo string, wxID int64) (*GrowthTrajectoryAccessResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	entActive, entExp, eErr := HasActiveUserFeatureEntitlement(ctx, wxID, FeatureIDGrowthTrajectoryPredict)
	if eErr != nil {
		return nil, eErr
	}
	vip := false
	if st, vErr := GetVipStatus(ctx, wxID); vErr != nil {
		glog.Warningf(ctx, "[cash-growth-trajectory-access] VIP 查询失败 wxId=%d err=%v，降级非 VIP", wxID, vErr)
	} else {
		vip = st.IsVip
	}
	trialUnused, tErr := IsTrialUnused(ctx, wxID, FeatureIDGrowthTrajectoryPredict)
	if tErr != nil {
		return nil, tErr
	}
	featureActive := entActive || vip
	out := &GrowthTrajectoryAccessResult{
		FeatureActive:        featureActive,
		TrialAvailable:       trialUnused && !featureActive,
		EntitlementExpiresAt: entExp,
	}
	out.Allowed = featureActive || trialUnused
	return out, nil
}
