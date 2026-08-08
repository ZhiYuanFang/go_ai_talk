package cash

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// VipStatus 账号 VIP 读模型。
type VipStatus struct {
	WxId     int64 `json:"wxId"`
	IsVip    bool  `json:"isVip"`
	ExpireAt int64 `json:"expireAt"`
}

// GetVipStatus 按 wxId 读权益；无行或过期 → isVip=false。
func GetVipStatus(ctx context.Context, wxID int64) (VipStatus, error) {
	out := VipStatus{WxId: wxID, IsVip: false, ExpireAt: 0}
	if wxID <= 0 {
		return out, nil
	}
	one, err := g.DB().Model("vip_entitlement").Ctx(ctx).
		Where("wx_id", wxID).Limit(1).One()
	if err != nil {
		return out, err
	}
	if one.IsEmpty() {
		return out, nil
	}
	expireAt := one["expire_at"].Int64()
	out.ExpireAt = expireAt
	out.IsVip = expireAt > time.Now().Unix()
	return out, nil
}

// IsVip 便捷布尔（供内部接口）。
func IsVip(ctx context.Context, wxID int64) (bool, error) {
	st, err := GetVipStatus(ctx, wxID)
	return st.IsVip, err
}

// ExtendEntitlement 续期：new_expire = max(now, current) + durationDays*86400。
func ExtendEntitlement(ctx context.Context, wxID int64, durationDays int) (int64, error) {
	if wxID <= 0 {
		return 0, nil
	}
	if durationDays <= 0 {
		durationDays = ProductDurationD
	}
	now := time.Now().Unix()
	st, err := GetVipStatus(ctx, wxID)
	if err != nil {
		return 0, err
	}
	base := now
	if st.ExpireAt > base {
		base = st.ExpireAt
	}
	newExpire := base + int64(durationDays)*86400
	_, err = g.DB().Exec(ctx, `
INSERT INTO vip_entitlement (wx_id, expire_at, updated_at) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE expire_at=VALUES(expire_at), updated_at=VALUES(updated_at)`,
		wxID, newExpire, now)
	return newExpire, err
}
