package cash

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	ucgRequiredDays     = 7
	ucgMinRecordsPerDay = 10
)

// UCGEligibilityResult UCG 入场资格（仅喂养日；与 VIP/功能无关）。
type UCGEligibilityResult struct {
	Qualified     bool   `json:"qualified"`
	RequiredDays  int    `json:"requiredDays"`
	EffectiveDays int    `json:"effectiveDays"`
	RemainingDays int    `json:"remainingDays"`
	Message       string `json:"message,omitempty"`
}

// GetUCGEligibility 计算并按日缓存资格；不落 MySQL；不读 VIP。
func GetUCGEligibility(ctx context.Context, deviceNo string) (*UCGEligibilityResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少设备号")
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	dayKey := time.Now().In(loc).Format("20060102")
	cacheKey := cachekit.CashUCGEligibilityKey(deviceNo, dayKey)
	c := cachekit.Default()
	if raw, ok, gErr := c.Get(ctx, cacheKey); gErr == nil && ok && raw != "" {
		var cached UCGEligibilityResult
		if json.Unmarshal([]byte(raw), &cached) == nil {
			return &cached, nil
		}
	}

	stats, err := FetchFeedingDayStats(ctx, deviceNo, 14)
	if err != nil {
		return nil, err
	}
	// days[0]=今日，向后连续统计有效日（count>=10）。
	effective := 0
	for _, d := range stats.Days {
		if d.Count >= ucgMinRecordsPerDay {
			effective++
		} else {
			break
		}
	}
	remaining := ucgRequiredDays - effective
	if remaining < 0 {
		remaining = 0
	}
	out := &UCGEligibilityResult{
		Qualified:     effective >= ucgRequiredDays,
		RequiredDays:  ucgRequiredDays,
		EffectiveDays: effective,
		RemainingDays: remaining,
	}
	if out.Qualified {
		out.Message = "已满足连续有效喂养日"
	} else {
		out.Message = "继续保持每日有效喂养记录以解锁 UCG"
	}
	if b, mErr := json.Marshal(out); mErr == nil {
		// TTL：36h，覆盖跨日缓冲。
		_ = c.SetEX(ctx, cacheKey, string(b), 36*time.Hour)
	}
	return out, nil
}
