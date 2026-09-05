package cash

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// CareAlertAccessResult 值得留意可看合成（喂养资格 ∧（设备开通 ∨ VIP））。
type CareAlertAccessResult struct {
	Allowed              bool  `json:"allowed"`
	FeedingQualified     bool  `json:"feedingQualified"`
	FeatureActive        bool  `json:"featureActive"`
	EntitlementExpiresAt int64 `json:"entitlementExpiresAt,omitempty"`
}

// GetCareAlertAccess 合成值得留意是否可看（权威在 cash；voice 经 internal 调用）。
//
// 业务：feedingQualified 复用 care_alert_entry；featureActive = 未过期设备权益 ∨ isVip。
// VIP 查询失败当作非 VIP（仅认 entitlement）；喂养计算失败向上返回（调用方 fail-closed）。
// VIP/开通 MUST NOT 短路 feedingQualified。
func GetCareAlertAccess(ctx context.Context, deviceNo string, wxID int64) (*CareAlertAccessResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	feeding, err := GetCareAlertFeedingEligibility(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	entActive, entExp, eErr := HasActiveFeatureEntitlement(ctx, deviceNo, FeatureIDCareAlertSmartRemind)
	if eErr != nil {
		return nil, eErr
	}
	vip := false
	if st, vErr := GetVipStatus(ctx, wxID); vErr != nil {
		glog.Warningf(ctx, "[cash-care-alert-access] VIP 查询失败 wxId=%d err=%v，降级非 VIP", wxID, vErr)
	} else {
		vip = st.IsVip
	}
	featureActive := entActive || vip
	out := &CareAlertAccessResult{
		FeedingQualified:     feeding != nil && feeding.Qualified,
		FeatureActive:        featureActive,
		EntitlementExpiresAt: entExp,
	}
	out.Allowed = out.FeedingQualified && out.FeatureActive
	return out, nil
}
