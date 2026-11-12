package cash

// feature_trial_start.go：预测能力显式点试用（7 日起算，人对人）。
//
// 业务：仅 predict_capability_basic / predict_capability_pro；Pro 须先过喂养资格。
// 成长轨迹/值得留意仍走 soft + ClaimTrialAfterSuccess（24h），不走本入口。

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// StartFeatureTrial App 点试用：预测两档 7 日；写 trial used + user entitlement。
//
// Args: wxID 账号；deviceNo 绑机（Pro 喂养资格用）；featureID 须为预测能力功能。
// Returns: 参数/门禁/写库错误。
// Side Effects: feature_user_trial、feature_user_entitlement、相关缓存。
func StartFeatureTrial(ctx context.Context, wxID int64, deviceNo, featureID string) error {
	featureID = strings.TrimSpace(featureID)
	deviceNo = strings.TrimSpace(deviceNo)
	if wxID <= 0 || featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId/featureId 无效")
	}
	if !IsPredictCapabilityFeature(featureID) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "该功能不支持点试用（请走原试用逻辑）")
	}
	if err := RequireFeatureDefActive(ctx, featureID); err != nil {
		return err
	}
	if featureID == FeatureIDPredictCapabilityPro {
		if deviceNo == "" {
			return gerror.NewCode(gcode.CodeInvalidParameter, "缺少设备号")
		}
		if err := RequirePredictProFeedingQualified(ctx, deviceNo); err != nil {
			return err
		}
	}

	unused, err := IsTrialUnused(ctx, wxID, featureID)
	if err != nil {
		return err
	}
	if !unused {
		return gerror.NewCode(gcode.CodeInvalidOperation, "试用已领取")
	}
	entActive, _, eErr := HasActiveUserFeatureEntitlement(ctx, wxID, featureID)
	if eErr != nil {
		return eErr
	}
	if entActive {
		return gerror.NewCode(gcode.CodeInvalidOperation, "已开通该功能，无需试用")
	}
	// VIP 覆盖期间不消耗试用资格。
	if st, vErr := GetVipStatus(ctx, wxID); vErr != nil {
		glog.Warningf(ctx, "[cash-trial-start] VIP 查询失败 wxId=%d err=%v，继续尝试领取", wxID, vErr)
	} else if st.IsVip {
		return gerror.NewCode(gcode.CodeInvalidOperation, "VIP 期间无需领取试用")
	}

	now := time.Now().Unix()
	_, _ = g.DB().Exec(ctx, `
INSERT IGNORE INTO feature_user_trial (wx_id, feature_id, status, used_at, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?)`,
		wxID, featureID, TrialStatusUnused, now, now)

	res, err := g.DB().Model("feature_user_trial").Ctx(ctx).
		Where("wx_id", wxID).Where("feature_id", featureID).
		Where("status", TrialStatusUnused).
		Data(g.Map{
			"status": TrialStatusUsed, "used_at": now, "updated_at": now,
		}).Update()
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return gerror.NewCode(gcode.CodeInvalidOperation, "试用已领取")
	}

	hours := PredictTrialDurationDays * 24
	if err := ActivateFeature(ctx, ActivateFeatureRequest{
		FeatureID:     featureID,
		SubjectType:   ActivationSubjectUser,
		SubjectKey:    strconv.FormatInt(wxID, 10),
		Channel:       UnlockMethodTrial,
		ChannelRef:    "trial_start",
		ActorWxID:     wxID,
		GrantKind:     GrantKindEntitlement,
		GrantQty:      1,
		DurationHours: hours,
	}); err != nil {
		glog.Warningf(ctx, "[cash-trial-start] ActivateFeature 失败，回滚 trial wxId=%d feature=%s err=%v", wxID, featureID, err)
		_ = revertTrialToUnused(ctx, wxID, featureID)
		return err
	}
	glog.Infof(ctx, "[cash-trial-start] ok wxId=%d feature=%s days=%d", wxID, featureID, PredictTrialDurationDays)
	return nil
}
