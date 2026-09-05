package cash

import (
	deviceclient "hello/internal/clients/device"
	ucgclient "hello/internal/clients/ucg"
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureInviteCode 为 owner 确保一用户一码（幂等）；码存 cash，不写 wx。
// Returns: code；Side Effects: 可能 INSERT feature_invite_code。
func EnsureInviteCode(ctx context.Context, ownerWxID int64) (string, error) {
	if ownerWxID <= 0 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "ownerWxId 无效")
	}
	var row struct {
		Code string `json:"code"`
	}
	_ = g.DB().Model("feature_invite_code").Ctx(ctx).Where("owner_wx_id", ownerWxID).Scan(&row)
	if strings.TrimSpace(row.Code) != "" {
		return row.Code, nil
	}
	now := time.Now().Unix()
	for i := 0; i < 8; i++ {
		code, err := genInviteCode()
		if err != nil {
			return "", err
		}
		_, err = g.DB().Model("feature_invite_code").Ctx(ctx).Data(g.Map{
			"code": code, "owner_wx_id": ownerWxID,
			"expires_at": 0, "max_redemptions": 0, "redeemed_count": 0,
			"grant_duration_days": 0, "status": 1,
			"created_at": now, "updated_at": now,
		}).Insert()
		if err == nil {
			return code, nil
		}
		// 唯一冲突重试（码碰撞或并发双插 owner）。
		_ = g.DB().Model("feature_invite_code").Ctx(ctx).Where("owner_wx_id", ownerWxID).Scan(&row)
		if strings.TrimSpace(row.Code) != "" {
			return row.Code, nil
		}
	}
	return "", gerror.NewCode(gcode.CodeInternalError, "生成邀请码失败")
}

func genInviteCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(b)), nil
}

// InviteMine 我的邀请码与获客数（懒 Ensure）。
type InviteMine struct {
	Code          string `json:"code"`
	RedeemedCount int    `json:"redeemedCount"`
}

// GetInviteMine 读取或创建当前用户邀请码。
func GetInviteMine(ctx context.Context, ownerWxID int64) (*InviteMine, error) {
	code, err := EnsureInviteCode(ctx, ownerWxID)
	if err != nil {
		return nil, err
	}
	var row struct {
		RedeemedCount int `json:"redeemed_count"`
	}
	_ = g.DB().Model("feature_invite_code").Ctx(ctx).Where("code", code).Scan(&row)
	return &InviteMine{Code: code, RedeemedCount: row.RedeemedCount}, nil
}

// InviteeRow 成功使用我码的用户。
type InviteeRow struct {
	WxId       int64  `json:"wxId"`
	Nickname   string `json:"nickname"`
	RedeemedAt int64  `json:"redeemedAt"`
}

// ListInviteInvitees 按码主人列出成功兑换者（昵称经 ucg 批量 profile）。
func ListInviteInvitees(ctx context.Context, ownerWxID int64) ([]InviteeRow, error) {
	if ownerWxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	type redT struct {
		RedeemerWxId int64 `json:"redeemer_wx_id"`
		RedeemedAt   int64 `json:"redeemed_at"`
	}
	var reds []redT
	err := g.DB().Model("feature_invite_redemption").Ctx(ctx).
		Where("owner_wx_id", ownerWxID).
		OrderDesc("redeemed_at").
		Scan(&reds)
	if err != nil {
		return nil, err
	}
	// 同一 redeemer 多次兑不同功能时合并为最早/最近一次展示：取每条 redemption，按人去重保留首次。
	seen := map[int64]struct{}{}
	ids := make([]int64, 0)
	ordered := make([]redT, 0)
	for _, r := range reds {
		if _, ok := seen[r.RedeemerWxId]; ok {
			continue
		}
		seen[r.RedeemerWxId] = struct{}{}
		ids = append(ids, r.RedeemerWxId)
		ordered = append(ordered, r)
	}
	nicks, _ := ucgclient.FetchUcgNicknames(ctx, ids)
	out := make([]InviteeRow, 0, len(ordered))
	for _, r := range ordered {
		nick := nicks[r.RedeemerWxId]
		if nick == "" {
			nick = "用户"
		}
		out = append(out, InviteeRow{WxId: r.RedeemerWxId, Nickname: nick, RedeemedAt: r.RedeemedAt})
	}
	return out, nil
}

// RedeemInviteCode 邀请码兑换单功能。
//
// 规则：不可自用；不可使用同一宝宝（同 device_no）下其他账号的码；人×码×功能仅一次；
// 多好友码可兑（不同设备）；预测永久 +1；非预测经 ActivateFeature（邀请天数读 feature_def）；
// 值得留意等 InviteOncePerDevice：device×feature 邀请仅一次；原力仍记码主人用户。
// 码级有效期/功能子表/一家锁定不再校验；开通能力仅看 feature_def.unlock_methods。
// 主人设备号经 device 契约查询：失败 fail-closed；主人未绑机（空 device_no）不因同设备规则拒绝。
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

	// 事务前窥视码主人，便于同设备校验时不持行锁打 HTTP。
	var peek struct {
		Code      string `json:"code"`
		OwnerWxId int64  `json:"owner_wx_id"`
		Status    int    `json:"status"`
	}
	_ = g.DB().Model("feature_invite_code").Ctx(ctx).Where("code", code).Scan(&peek)
	if peek.Code == "" || peek.Status != 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "邀请码无效或已停用")
	}
	if peek.OwnerWxId == redeemerWxID {
		return gerror.NewCode(gcode.CodeInvalidParameter, "不可使用自己的邀请码")
	}
	ownerDeviceNo, dErr := deviceclient.FetchDeviceNoByWxID(ctx, peek.OwnerWxId)
	if dErr != nil {
		glog.Warningf(ctx, "[cash-invite] owner device lookup failed owner=%d err=%v", peek.OwnerWxId, dErr)
		return gerror.WrapCode(gcode.CodeInternalError, dErr, "暂时无法校验邀请码，请稍后重试")
	}
	if ownerDeviceNo != "" && ownerDeviceNo == deviceNo {
		return gerror.NewCode(gcode.CodeInvalidParameter, "不可使用同一宝宝下其他账号的邀请码")
	}

	now := time.Now().Unix()
	var ownerWxID int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var codeRow struct {
			Code          string `json:"code"`
			OwnerWxId     int64  `json:"owner_wx_id"`
			RedeemedCount int    `json:"redeemed_count"`
			Status        int    `json:"status"`
		}
		err := tx.Model("feature_invite_code").Ctx(ctx).Where("code", code).LockUpdate().Scan(&codeRow)
		if err != nil {
			return err
		}
		if codeRow.Code == "" || codeRow.Status != 1 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "邀请码无效或已停用")
		}
		if codeRow.OwnerWxId == redeemerWxID {
			return gerror.NewCode(gcode.CodeInvalidParameter, "不可使用自己的邀请码")
		}
		ownerWxID = codeRow.OwnerWxId

		// 人×码×功能去重。
		n, err := tx.Model("feature_invite_feature_grant").Ctx(ctx).
			Where("redeemer_wx_id", redeemerWxID).
			Where("code", code).
			Where("feature_id", featureID).Count()
		if err != nil {
			return err
		}
		if n > 0 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "该邀请码已用于开通此功能")
		}

		var def struct {
			FeatureId     string `json:"feature_id"`
			UnlockMethods string `json:"unlock_methods"`
			Status        int    `json:"status"`
		}
		_ = tx.Model("feature_def").Ctx(ctx).Where("feature_id", featureID).Scan(&def)
		if def.FeatureId == "" || def.Status != 1 || !strings.Contains(def.UnlockMethods, UnlockMethodInviteCode) {
			return gerror.NewCode(gcode.CodeInvalidParameter, "功能不支持邀请码开通")
		}

		// 值得留意等：同一 device_no 对本功能仅能邀请开通一次（全家共享防刷）。
		if InviteOncePerDevice(featureID) {
			dn, err := tx.Model("feature_invite_device_grant").Ctx(ctx).
				Where("device_no", deviceNo).Where("feature_id", featureID).Count()
			if err != nil {
				return err
			}
			if dn > 0 {
				return gerror.NewCode(gcode.CodeInvalidParameter, "该设备已使用过邀请码开通此功能")
			}
		}

		// 经原子入口授予：预测 +1；其它读 feature_def.duration_days（邀请/广告同源）。
		if err := ActivateFeature(ctx, ActivateFeatureRequest{
			FeatureID:   featureID,
			SubjectType: ActivationSubjectDevice,
			SubjectKey:  deviceNo,
			Channel:     UnlockMethodInviteCode,
			ChannelRef:  code,
			ActorWxID:   redeemerWxID,
		}); err != nil {
			return err
		}

		if InviteOncePerDevice(featureID) {
			_, err = tx.Model("feature_invite_device_grant").Ctx(ctx).Data(g.Map{
				"device_no": deviceNo, "feature_id": featureID, "code": code,
				"redeemer_wx_id": redeemerWxID, "redeemed_at": now,
			}).Insert()
			if err != nil {
				return err
			}
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
		return err
	})
	if err != nil {
		return err
	}
	// 获客原力：兑码成功后尽力加分，失败不回滚开通。
	if ownerWxID > 0 {
		if ferr := ucgclient.NotifyUcgInviteAcquisition(ctx, ownerWxID, code); ferr != nil {
			glog.Warningf(ctx, "[cash-invite] ucg force acquire failed owner=%d code=%s err=%v", ownerWxID, code, ferr)
		}
	}
	return nil
}
