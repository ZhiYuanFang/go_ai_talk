// Admin 撤销手工授：仅最近一笔 paid 且 channel=admin、权益仍有效时可撤。
//
// 业务：权益表只有一根到期时钟。撤销把 expire/expires_at 写成 now（直接过期），
// 并把该笔 admin 订单标为 revoked。不按天数回退，不走 ShrinkEntitlement，不用 refunded。
package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// AdminRevokeVip 撤销该 wx 最近一笔手工授 VIP，使权益立即过期。
//
// Args: wxID>0。
// Returns: 错误（已过期、最近一笔非 admin、口令由 controller 校验）。
// Side Effects: vip_entitlement.expire_at=now；对应 vip_order.status=revoked。
func AdminRevokeVip(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	st, err := GetVipStatus(ctx, wxID)
	if err != nil {
		return err
	}
	if !st.IsVip {
		return gerror.NewCode(gcode.CodeInvalidOperation, "权益已过期，无法撤销")
	}
	one, err := g.DB().Model("vip_order").Ctx(ctx).
		Fields("order_no,channel").
		Where("wx_id", wxID).Where("status", OrderPaid).
		OrderDesc("paid_at").OrderDesc("id").Limit(1).One()
	if err != nil {
		return err
	}
	if one.IsEmpty() || one["channel"].String() != ChannelAdmin {
		return gerror.NewCode(gcode.CodeInvalidOperation, "仅最近一笔手工授可撤销")
	}
	orderNo := one["order_no"].String()
	now := time.Now().Unix()
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, e := tx.Model("vip_entitlement").Ctx(ctx).Where("wx_id", wxID).Data(g.Map{
			"expire_at":  now,
			"updated_at": now,
		}).Update(); e != nil {
			return e
		}
		res, e := tx.Model("vip_order").Ctx(ctx).
			Where("order_no", orderNo).Where("status", OrderPaid).
			Data(g.Map{"status": OrderRevoked}).Update()
		if e != nil {
			return e
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return gerror.NewCode(gcode.CodeInvalidOperation, "订单已撤销或状态已变")
		}
		return nil
	})
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[cash-admin-revoke] VIP orderNo=%s wxId=%d", orderNo, wxID)
	return nil
}

// AdminRevokeFeature 撤销该功能该主体最近一笔手工授，使权益立即过期。
//
// Args: featureID；user 主体须 wxID，device 主体须 deviceNo。
// Side Effects: 对应权益 expires_at=now；feature_order.status=revoked；失效相关缓存。
func AdminRevokeFeature(ctx context.Context, featureID string, wxID int64, deviceNo string) error {
	featureID = strings.TrimSpace(featureID)
	deviceNo = strings.TrimSpace(deviceNo)
	if featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId 必填")
	}
	if featureID == FeatureIDPredictionUnlock {
		glog.Warningf(ctx, "[cash] prediction_unlock 已下线，拒绝撤销 featureId=%s", featureID)
		return gerror.NewCode(gcode.CodeInvalidOperation, "预测事项开通数量已下线")
	}
	subj, err := GetFeatureActivationSubject(ctx, featureID)
	if err != nil {
		return err
	}
	if subj == ActivationSubjectUser {
		if wxID <= 0 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "该功能须填写 wxId")
		}
		deviceNo = ""
	} else if deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "该功能须填写 deviceNo")
	}

	active, err := featureEntitlementStillActive(ctx, subj, featureID, wxID, deviceNo)
	if err != nil {
		return err
	}
	if !active {
		return gerror.NewCode(gcode.CodeInvalidOperation, "权益已过期，无法撤销")
	}

	orderNo, channel, err := latestPaidFeatureOrder(ctx, featureID, subj, wxID, deviceNo)
	if err != nil {
		return err
	}
	if orderNo == "" || channel != ChannelAdmin {
		return gerror.NewCode(gcode.CodeInvalidOperation, "仅最近一笔手工授可撤销")
	}

	now := time.Now().Unix()
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var res sqlResult
		var e error
		if subj == ActivationSubjectUser {
			res, e = tx.Model("feature_user_entitlement").Ctx(ctx).
				Where("wx_id", wxID).Where("feature_id", featureID).
				Data(g.Map{"expires_at": now, "updated_at": now}).Update()
		} else {
			res, e = tx.Model("feature_entitlement").Ctx(ctx).
				Where("device_no", deviceNo).Where("feature_id", featureID).
				Data(g.Map{"expires_at": now, "updated_at": now}).Update()
		}
		if e != nil {
			return e
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return gerror.NewCode(gcode.CodeInvalidOperation, "权益不存在")
		}
		res, e = tx.Model("feature_order").Ctx(ctx).
			Where("order_no", orderNo).Where("status", OrderPaid).
			Data(g.Map{"status": OrderRevoked}).Update()
		if e != nil {
			return e
		}
		n, _ = res.RowsAffected()
		if n == 0 {
			return gerror.NewCode(gcode.CodeInvalidOperation, "订单已撤销或状态已变")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if subj == ActivationSubjectUser {
		invalidateFeatureDefCache(ctx)
	} else {
		invalidateDeviceFeatureCaches(ctx, deviceNo)
	}
	glog.Infof(ctx, "[cash-admin-revoke] feature orderNo=%s featureId=%s wxId=%d deviceNo=%s", orderNo, featureID, wxID, deviceNo)
	return nil
}

// sqlResult 抽象 RowsAffected，避免在签名里泄漏 gdb.Result 细节。
type sqlResult interface {
	RowsAffected() (int64, error)
}

func featureEntitlementStillActive(ctx context.Context, subj, featureID string, wxID int64, deviceNo string) (bool, error) {
	var exp int64
	var empty bool
	if subj == ActivationSubjectUser {
		r, err := g.DB().Model("feature_user_entitlement").Ctx(ctx).
			Fields("expires_at").Where("wx_id", wxID).Where("feature_id", featureID).One()
		if err != nil {
			return false, err
		}
		empty = r.IsEmpty()
		if !empty {
			exp = r["expires_at"].Int64()
		}
	} else {
		r, err := g.DB().Model("feature_entitlement").Ctx(ctx).
			Fields("expires_at").Where("device_no", deviceNo).Where("feature_id", featureID).One()
		if err != nil {
			return false, err
		}
		empty = r.IsEmpty()
		if !empty {
			exp = r["expires_at"].Int64()
		}
	}
	if empty {
		return false, nil
	}
	// expires_at=0 表示永久，仍视为有效。
	if exp == 0 {
		return true, nil
	}
	return exp > time.Now().Unix(), nil
}

func latestPaidFeatureOrder(ctx context.Context, featureID, subj string, wxID int64, deviceNo string) (orderNo, channel string, err error) {
	placeholder := "admin_" + featureID
	sql := `
SELECT o.order_no AS order_no, o.channel AS channel
FROM feature_order o
WHERE o.status = ?
  AND (
    o.product_code IN (SELECT product_code FROM feature_product WHERE feature_id = ?)
    OR o.product_code = ?
  )`
	args := []interface{}{OrderPaid, featureID, placeholder}
	if subj == ActivationSubjectUser {
		sql += ` AND o.wx_id = ?`
		args = append(args, wxID)
	} else {
		sql += ` AND o.device_no = ?`
		args = append(args, deviceNo)
	}
	sql += ` ORDER BY o.paid_at DESC, o.id DESC LIMIT 1`
	one, err := g.DB().GetOne(ctx, sql, args...)
	if err != nil {
		return "", "", err
	}
	if one.IsEmpty() {
		return "", "", nil
	}
	return one["order_no"].String(), one["channel"].String(), nil
}
