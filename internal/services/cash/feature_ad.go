package cash

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// CompleteFeatureAd 广告开通已下线；保留入口给旧客户端明确错误。
func CompleteFeatureAd(ctx context.Context, deviceNo, featureID, idemKey string, grantQty, durationDays int, wxID int64) error {
	_ = ctx
	_ = deviceNo
	_ = featureID
	_ = idemKey
	_ = grantQty
	_ = durationDays
	_ = wxID
	return gerror.NewCode(gcode.CodeInvalidOperation, "广告开通已下线")
}
