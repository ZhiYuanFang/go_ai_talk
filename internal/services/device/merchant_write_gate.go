package device

import (
	"context"
	"strings"

	cashclient "hello/internal/clients/cash"
)

// MerchantBlockHistoryCreate 若 device 属商户且权益无效则禁 create。
// 探活失败：对已归属商户的宝宝 fail-closed（禁 create）；无商户归属放行。
func MerchantBlockHistoryCreate(ctx context.Context, deviceNo string) (block bool, merchantID int64, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return false, 0, nil
	}
	merchantID, err = MerchantOwnerOfDevice(ctx, deviceNo)
	if err != nil {
		return true, 0, err
	}
	if merchantID <= 0 {
		return false, 0, nil
	}
	entitled, e2 := cashclient.RemoteIsMerchantEntitled(ctx, merchantID)
	if e2 != nil {
		return true, merchantID, e2
	}
	return !entitled, merchantID, nil
}
