package cash

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const adDailyLimit = 5

// CompleteFeatureAd MVP：信任客户端广告完成申报并授予。
func CompleteFeatureAd(ctx context.Context, deviceNo, featureID, idemKey string, grantQty, durationDays int) error {
	deviceNo = strings.TrimSpace(deviceNo)
	featureID = strings.TrimSpace(featureID)
	if deviceNo == "" || featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo/featureId 不能为空")
	}
	if grantQty <= 0 {
		grantQty = 1
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	day := time.Now().In(loc).Format("20060102")
	c := cachekit.Default()

	if idemKey == "" {
		idemKey = featureID
	}
	idemRedisKey := cachekit.CashFeatureAdIdempotencyKey(deviceNo, featureID, idemKey)
	ok, err := c.SetNXEX(ctx, idemRedisKey, "1", 10*time.Minute)
	if err != nil {
		return err
	}
	if !ok {
		return nil // 短窗幂等成功
	}

	dailyKey := cachekit.CashFeatureAdDailyLimitKey(deviceNo, day)
	n, err := c.Incr(ctx, dailyKey)
	if err != nil {
		return err
	}
	if n == 1 {
		_ = c.Expire(ctx, dailyKey, 36*time.Hour)
	}
	if n > adDailyLimit {
		return gerror.NewCode(gcode.CodeInvalidOperation, fmt.Sprintf("今日广告开通已达上限 %d", adDailyLimit))
	}

	var def struct {
		FeatureId     string `json:"feature_id"`
		UnlockMethods string `json:"unlock_methods"`
		Status        int    `json:"status"`
	}
	_ = g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", featureID).Scan(&def)
	if def.FeatureId == "" || def.Status != 1 || !strings.Contains(def.UnlockMethods, UnlockMethodAd) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "功能不支持广告开通")
	}
	// 广告与邀请同源效果：经 ActivateFeature（预测 +1；权益型读 def.duration_days）。
	// 客户端传入 durationDays 忽略，避免与邀请分叉。
	_ = durationDays
	return ActivateFeature(ctx, ActivateFeatureRequest{
		FeatureID:   featureID,
		SubjectType: ActivationSubjectDevice,
		SubjectKey:  deviceNo,
		Channel:     UnlockMethodAd,
		ChannelRef:  "ad:" + idemKey,
		ActorWxID:   0,
		GrantQty:    grantQty,
	})
}
