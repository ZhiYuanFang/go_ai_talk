package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// RedeemInviteCode 邀请码兑换单功能。
//
// 规则：不可自用；一家锁定仅成功后写入；人×功能仅一次；单次只开 featureId。
func RedeemInviteCode(ctx context.Context, redeemerWxID int64, deviceNo, code, featureID string) error {
	code = strings.TrimSpace(code)
	featureID = strings.TrimSpace(featureID)
	deviceNo = strings.TrimSpace(deviceNo)
	if redeemerWxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "缺少账号 wxId")
	}
	if deviceNo == "" || code == "" || featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo/code/featureId 不能为空")
	}

	now := time.Now().Unix()
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var codeRow struct {
			Code               string `json:"code"`
			OwnerWxId          int64  `json:"owner_wx_id"`
			ExpiresAt          int64  `json:"expires_at"`
			MaxRedemptions     int    `json:"max_redemptions"`
			RedeemedCount      int    `json:"redeemed_count"`
			GrantDurationDays  int    `json:"grant_duration_days"`
			Status             int    `json:"status"`
		}
		err := tx.Model("feature_invite_code").Ctx(ctx).Where("code", code).LockUpdate().Scan(&codeRow)
		if err != nil {
			return err
		}
		if codeRow.Code == "" || codeRow.Status != 1 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "邀请码无效或已停用")
		}
		if codeRow.ExpiresAt > 0 && codeRow.ExpiresAt < now {
			return gerror.NewCode(gcode.CodeInvalidParameter, "邀请码已过期")
		}
		if codeRow.OwnerWxId == redeemerWxID {
			return gerror.NewCode(gcode.CodeInvalidParameter, "不可使用自己的邀请码")
		}
		if codeRow.MaxRedemptions > 0 && codeRow.RedeemedCount >= codeRow.MaxRedemptions {
			return gerror.NewCode(gcode.CodeInvalidParameter, "邀请码兑换次数已满")
		}

		// 一家锁定：已绑定其他 owner 则拒绝（失败兑换不绑定，故仅查表）。
		var bind struct {
			OwnerWxId int64 `json:"owner_wx_id"`
		}
		_ = tx.Model("feature_invite_redeemer_bind").Ctx(ctx).Where("redeemer_wx_id", redeemerWxID).Scan(&bind)
		if bind.OwnerWxId > 0 && bind.OwnerWxId != codeRow.OwnerWxId {
			return gerror.NewCode(gcode.CodeInvalidParameter, "已使用其他邀请码，不能换一家")
		}

		// 人×功能去重。
		n, err := tx.Model("feature_invite_feature_grant").Ctx(ctx).
			Where("redeemer_wx_id", redeemerWxID).Where("feature_id", featureID).Count()
		if err != nil {
			return err
		}
		if n > 0 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "该功能已用邀请码开通过")
		}

		// 码是否支持该功能：子表有行则必须命中；无子表则要求 feature_def 含 invite_code。
		featN, err := tx.Model("feature_invite_code_feature").Ctx(ctx).Where("code", code).Count()
		if err != nil {
			return err
		}
		grantQty := 1
		if featN > 0 {
			var cf struct {
				FeatureId     string `json:"feature_id"`
				GrantQuantity int    `json:"grant_quantity"`
			}
			err = tx.Model("feature_invite_code_feature").Ctx(ctx).
				Where("code", code).Where("feature_id", featureID).Scan(&cf)
			if err != nil {
				return err
			}
			if cf.FeatureId == "" {
				return gerror.NewCode(gcode.CodeInvalidParameter, "该邀请码不支持此功能")
			}
			if cf.GrantQuantity > 0 {
				grantQty = cf.GrantQuantity
			}
		} else {
			var def struct {
				FeatureId     string `json:"feature_id"`
				UnlockMethods string `json:"unlock_methods"`
				Status        int    `json:"status"`
			}
			_ = tx.Model("feature_def").Ctx(ctx).Where("feature_id", featureID).Scan(&def)
			if def.FeatureId == "" || def.Status != 1 || !strings.Contains(def.UnlockMethods, UnlockMethodInviteCode) {
				return gerror.NewCode(gcode.CodeInvalidParameter, "功能不支持邀请码开通")
			}
		}

		duration := codeRow.GrantDurationDays
		grantKind := GrantKindEntitlement
		if featureID == FeatureIDPredictionUnlock {
			grantKind = GrantKindAllowedCountDelta
		}
		if err := GrantEntitlementOrCount(ctx, deviceNo, featureID, UnlockMethodInviteCode, grantKind, grantQty, duration, code); err != nil {
			return err
		}

		_, err = tx.Model("feature_invite_feature_grant").Ctx(ctx).Data(g.Map{
			"redeemer_wx_id": redeemerWxID,
			"feature_id":     featureID,
			"code":           code,
			"device_no":      deviceNo,
			"redeemed_at":    now,
		}).Insert()
		if err != nil {
			return err
		}
		_, err = tx.Model("feature_invite_redemption").Ctx(ctx).Data(g.Map{
			"code": code, "owner_wx_id": codeRow.OwnerWxId, "redeemer_wx_id": redeemerWxID,
			"device_no": deviceNo, "feature_id": featureID, "redeemed_at": now,
		}).Insert()
		if err != nil {
			return err
		}
		_, err = tx.Model("feature_invite_code").Ctx(ctx).Where("code", code).Data(g.Map{
			"redeemed_count": codeRow.RedeemedCount + 1,
			"updated_at":     now,
		}).Update()
		if err != nil {
			return err
		}
		// 仅成功开通后绑定一家。
		if bind.OwnerWxId == 0 {
			_, err = tx.Model("feature_invite_redeemer_bind").Ctx(ctx).Data(g.Map{
				"redeemer_wx_id": redeemerWxID,
				"owner_wx_id":    codeRow.OwnerWxId,
				"bound_at":       now,
			}).Insert()
			if err != nil {
				return err
			}
		}
		return nil
	})
}
