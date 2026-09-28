package usagestats

import (
	"context"
	"strconv"
	"time"

	"hello/internal/platform/cachekit"
)

// MaxLastAtForWxIDs 对本页 wxId 批量读取 per-wx last hash，取各 apiKey 时间戳最大值；
// 若结果早于 days 窗口起点则视为 0。days<=0 时窗口按 90 天（与 dayRange 全部语义一致）。
// 业务逻辑：运维用户维度列表只需「最近使用时间」一列，不必扫日桶交叉键。
func MaxLastAtForWxIDs(ctx context.Context, days int, wxIDs []int64) map[int64]int64 {
	out := make(map[int64]int64, len(wxIDs))
	if len(wxIDs) == 0 {
		return out
	}
	minTs := windowStartUnix(days, 90)
	for _, wxID := range wxIDs {
		if wxID <= 0 {
			continue
		}
		all, err := usageCache.HashGetAll(ctx, cachekit.GatewayUsageLastWxKey(wxID))
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
		// 窗口外的最近时间不计入，避免「近 7 天」展示更早的活跃。
		if maxTs < minTs {
			maxTs = 0
		}
		out[wxID] = maxTs
	}
	return out
}

// windowStartUnix 计算与 dayRange 对齐的窗口起点 Unix 秒（含当天共 effectiveDays 个自然日）。
func windowStartUnix(days, fallbackWhenZero int) int64 {
	effective := days
	if effective <= 0 {
		effective = fallbackWhenZero
	}
	if effective < 1 {
		effective = 1
	}
	now := time.Now()
	start := now.AddDate(0, 0, -(effective - 1))
	return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()).Unix()
}
