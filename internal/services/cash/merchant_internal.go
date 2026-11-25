package cash

import (
	"context"
)

// BlockHistoryCreateForMerchant 若商户欠费则 true（供 device 聚合写闸）。
func BlockHistoryCreateForMerchant(ctx context.Context, merchantID int64) (bool, error) {
	if merchantID <= 0 {
		return false, nil
	}
	ok, _, err := IsMerchantEntitled(ctx, merchantID)
	if err != nil {
		return true, err // 探活失败：对已知商户宝宝 fail-closed 禁 create
	}
	return !ok, nil
}
