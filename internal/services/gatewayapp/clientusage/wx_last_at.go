package clientusage

import (
	"context"
	"strconv"
	"time"

	"hello/internal/platform/cachekit"
)

// MaxLastAtForWxIDs 对本页 wxId 批量读取 client-usage per-wx last hash 取 max；
// 早于 days 窗口（NormalizeDays，≤0 为 30 天）则返回 0。
func MaxLastAtForWxIDs(ctx context.Context, days int, wxIDs []int64) map[int64]int64 {
	out := make(map[int64]int64, len(wxIDs))
	if len(wxIDs) == 0 {
		return out
	}
	effective := NormalizeDays(days)
	minTs := featWindowStartUnix(effective)
	for _, wxID := range wxIDs {
		if wxID <= 0 {
			continue
		}
		all, err := featCache.HashGetAll(ctx, cachekit.GatewayFeatUsageLastWxKey(wxID))
		if err != nil || len(all) == 0 {
			out[wxID] = 0
			continue
		}
		var maxTs int64
		for _, v := range all {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil || n <= 0 {
				continue
			}
			if n > maxTs {
				maxTs = n
			}
		}
		if maxTs < minTs {
			maxTs = 0
		}
		out[wxID] = maxTs
	}
	return out
}

// featWindowStartUnix 与 dayRange 对齐的窗口起点（含当天共 effective 个自然日）。
func featWindowStartUnix(effectiveDays int) int64 {
	if effectiveDays < 1 {
		effectiveDays = 1
	}
	now := time.Now()
	start := now.AddDate(0, 0, -(effectiveDays - 1))
	return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()).Unix()
}
