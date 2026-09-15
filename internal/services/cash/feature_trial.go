package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// IsTrialUnused 账号对该功能是否尚未成功 claim 试用（无行或 status=unused）。
func IsTrialUnused(ctx context.Context, wxID int64, featureID string) (bool, error) {
	featureID = strings.TrimSpace(featureID)
	if wxID <= 0 || featureID == "" {
		return false, nil
	}
	r, err := g.DB().Model("feature_user_trial").Ctx(ctx).
		Fields("status").
		Where("wx_id", wxID).Where("feature_id", featureID).
		One()
	if err != nil {
		return false, err
	}
	if r.IsEmpty() {
		return true, nil
	}
	return strings.TrimSpace(r["status"].String()) == TrialStatusUnused, nil
}

// ClaimTrialAfterSuccess 首次成功落库后原子 claim：trial→used + ActivateFeature(channel=trial, 24h)。
//
// 业务：仅 soft access（试用未用且无有效用户权益且非 VIP）时领取；已 used / 已开通 / VIP 则幂等跳过。
// Returns: 参数错误或授予失败。
func ClaimTrialAfterSuccess(ctx context.Context, wxID int64, featureID string) error {
	featureID = strings.TrimSpace(featureID)
	if wxID <= 0 || featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId/featureId 无效")
	}
	unused, err := IsTrialUnused(ctx, wxID, featureID)
	if err != nil {
		return err
	}
	if !unused {
		return nil
	}
	// 已有用户权益则勿消耗试用资格。
	entActive, _, eErr := HasActiveUserFeatureEntitlement(ctx, wxID, featureID)
	if eErr != nil {
		return eErr
	}
	if entActive {
		return nil
	}
	// VIP 覆盖期间保留 unused，到期后仍可 soft access。
	if st, vErr := GetVipStatus(ctx, wxID); vErr != nil {
		glog.Warningf(ctx, "[cash-trial] VIP 查询失败 wxId=%d err=%v，继续尝试 claim", wxID, vErr)
	} else if st.IsVip {
		return nil
	}

	now := time.Now().Unix()

	// 确保有一行；并发下 INSERT IGNORE 即可。
	_, _ = g.DB().Exec(ctx, `
INSERT IGNORE INTO feature_user_trial (wx_id, feature_id, status, used_at, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?)`,
		wxID, featureID, TrialStatusUnused, now, now)

	// 条件更新：仅 unused → used，避免双 claim。
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
		return nil // 已领取或竞态失败方
	}

	subjType, subjKey, sErr := ResolveActivateSubject(ctx, featureID, "", wxID)
	if sErr != nil {
		_ = revertTrialToUnused(ctx, wxID, featureID)
		return sErr
	}
	if subjType != ActivationSubjectUser {
		_ = revertTrialToUnused(ctx, wxID, featureID)
		return gerror.NewCode(gcode.CodeInvalidParameter, "试用仅支持账号维功能")
	}
	if err := ActivateFeature(ctx, ActivateFeatureRequest{
		FeatureID:     featureID,
		SubjectType:   ActivationSubjectUser,
		SubjectKey:    subjKey,
		Channel:       UnlockMethodTrial,
		ChannelRef:    "trial",
		ActorWxID:     wxID,
		GrantKind:     GrantKindEntitlement,
		GrantQty:      1,
		DurationHours: TrialDurationHours,
	}); err != nil {
		glog.Warningf(ctx, "[cash-trial] ActivateFeature 失败，回滚 trial wxId=%d feature=%s err=%v", wxID, featureID, err)
		_ = revertTrialToUnused(ctx, wxID, featureID)
		return err
	}
	glog.Infof(ctx, "[cash-trial] claim ok wxId=%d feature=%s hours=%d subject=%s",
		wxID, featureID, TrialDurationHours, subjKey)
	return nil
}

// revertTrialToUnused Activate 失败时尽力把 trial 标回 unused，避免资格被空耗。
func revertTrialToUnused(ctx context.Context, wxID int64, featureID string) error {
	now := time.Now().Unix()
	_, err := g.DB().Model("feature_user_trial").Ctx(ctx).
		Where("wx_id", wxID).Where("feature_id", featureID).
		Where("status", TrialStatusUsed).
		Data(g.Map{
			"status": TrialStatusUnused, "used_at": 0, "updated_at": now,
		}).Update()
	return err
}

// HasInviteFeatureGrantAnyCode 该账号是否已成功用任意邀请码开通过该功能（InviteOncePerUser）。
func HasInviteFeatureGrantAnyCode(ctx context.Context, wxID int64, featureID string) (bool, error) {
	featureID = strings.TrimSpace(featureID)
	if wxID <= 0 || featureID == "" {
		return false, nil
	}
	n, err := g.DB().Model("feature_invite_feature_grant").Ctx(ctx).
		Where("redeemer_wx_id", wxID).Where("feature_id", featureID).Count()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
