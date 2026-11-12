package cash

// feature_gate.go：功能上架与 Pro 喂养前置校验（建单/试用/邀请共用）。

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// RequireFeatureDefActive 功能定义须存在且 status=1。
func RequireFeatureDefActive(ctx context.Context, featureID string) error {
	featureID = strings.TrimSpace(featureID)
	if featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId 不能为空")
	}
	if featureID == FeatureIDPredictionUnlock {
		return gerror.NewCode(gcode.CodeInvalidOperation, "预测事项开通数量已下线")
	}
	one, err := g.DB().Model("feature_def").Ctx(ctx).
		Fields("feature_id,status").
		Where("feature_id", featureID).One()
	if err != nil {
		return err
	}
	if one.IsEmpty() || one["status"].Int() != 1 {
		return gerror.NewCode(gcode.CodeInvalidOperation, "功能不存在或已下架")
	}
	return nil
}

// RequirePredictProFeedingQualified Pro 开通前喂养资格须合格。
func RequirePredictProFeedingQualified(ctx context.Context, deviceNo string) error {
	out, err := GetPredictProFeedingEligibility(ctx, deviceNo)
	if err != nil {
		return err
	}
	if out == nil || !out.Qualified {
		msg := "尚未满足预测 Pro 喂养资格"
		if out != nil && strings.TrimSpace(out.Message) != "" {
			msg = out.Message
		}
		return gerror.NewCode(gcode.CodeInvalidOperation, msg)
	}
	return nil
}

// UnlockMethodsContainAd unlock_methods 是否含广告通道。
func UnlockMethodsContainAd(unlockMethods string) bool {
	for _, p := range strings.Split(unlockMethods, ",") {
		if strings.TrimSpace(p) == UnlockMethodAd {
			return true
		}
	}
	return false
}
