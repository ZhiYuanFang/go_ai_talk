package cash

// predict_capability_access.go：预测能力使用态合成（基础版∨Pro∨VIP）。
//
// 业务：供 voice imminent 写入门禁；权威在 cash，经 internal HTTP 出站。

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// PredictCapabilityAccessResult 预测能力可推送合成。
type PredictCapabilityAccessResult struct {
	Allowed       bool `json:"allowed"`       // basic∨pro∨VIP
	BasicActive   bool `json:"basicActive"`   // 基础版人权益未过期
	ProActive     bool `json:"proActive"`     // Pro 人权益未过期
	FeatureActive bool `json:"featureActive"` // basic∨pro∨VIP（同 Allowed；不含 soft）
	IsVip         bool `json:"isVip"`
}

// GetPredictCapabilityAccess 合成某账号是否具备预测 imminent 写入资格。
//
// 业务：Allowed = 基础版权益 ∨ Pro 权益 ∨ isVip。deviceNo 仅校验非空（绑机上下文）。
func GetPredictCapabilityAccess(ctx context.Context, deviceNo string, wxID int64) (*PredictCapabilityAccessResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	basic, _, err := HasActiveUserFeatureEntitlement(ctx, wxID, FeatureIDPredictCapabilityBasic)
	if err != nil {
		return nil, err
	}
	pro, _, err := HasActiveUserFeatureEntitlement(ctx, wxID, FeatureIDPredictCapabilityPro)
	if err != nil {
		return nil, err
	}
	vip := false
	if st, vErr := GetVipStatus(ctx, wxID); vErr != nil {
		glog.Warningf(ctx, "[cash-predict-capability-access] VIP 查询失败 wxId=%d err=%v，降级非 VIP", wxID, vErr)
	} else {
		vip = st.IsVip
	}
	allowed := basic || pro || vip
	return &PredictCapabilityAccessResult{
		Allowed:       allowed,
		BasicActive:   basic,
		ProActive:     pro,
		FeatureActive: allowed,
		IsVip:         vip,
	}, nil
}
