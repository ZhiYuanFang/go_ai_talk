package history

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// FeedingDayCount 某一上海日历日的 history 条数。
type FeedingDayCount struct {
	Date  string `json:"date"`  // yyyy-MM-dd（上海）
	Count int    `json:"count"` // 落入该日的记录数（按 start_time）
}

// FeedingDayStats 近 N 个上海日每日计数（含今日，从旧到新或从新到旧由调用方约定：本实现从「今日往前」顺序返回，index0=今日）。
type FeedingDayStats struct {
	DeviceNo string            `json:"deviceNo"`
	Days     []FeedingDayCount `json:"days"`
}

// GetFeedingDayStats 按 device_no 统计近 days 个上海日历日每日 history 行数。
// 业务：UCG 入场资格有效日判定（≥10 条/日）；口径为该设备全部 history，按 start_time 落入上海日。
// Args: deviceNo 设备号；days 天数（建议 7~14，上限 31）。
// Returns: 每日 date+count；Side Effects: 仅读 history 表，一次范围扫描后按日聚合。
func GetFeedingDayStats(ctx context.Context, deviceNo string, days int) (*FeedingDayStats, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if days <= 0 {
		days = 7
	}
	if days > 31 {
		days = 31
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	// 今日 00:00 起向前 days 天窗口。
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	windowStart := todayStart.AddDate(0, 0, -(days - 1))
	startUnix := windowStart.Unix()
	endUnix := todayStart.AddDate(0, 0, 1).Unix() // 明日 0 点，半开区间

	type row struct {
		StartTime int64 `json:"start_time"`
	}
	var rows []row
	// 最少扫描：仅拉时间戳，按设备+时间索引过滤。
	err = g.DB().Model("history").Ctx(ctx).
		Fields("start_time").
		Where("device_no", deviceNo).
		Where("start_time >= ?", startUnix).
		Where("start_time < ?", endUnix).
		Scan(&rows)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, days)
	for i := 0; i < days; i++ {
		d := todayStart.AddDate(0, 0, -i)
		counts[d.Format("2006-01-02")] = 0
	}
	for _, r := range rows {
		if r.StartTime <= 0 {
			continue
		}
		key := time.Unix(r.StartTime, 0).In(loc).Format("2006-01-02")
		if _, ok := counts[key]; ok {
			counts[key]++
		}
	}

	out := &FeedingDayStats{DeviceNo: deviceNo, Days: make([]FeedingDayCount, 0, days)}
	// 今日起向过去排列，便于资格算法从前向后扫连续日。
	for i := 0; i < days; i++ {
		d := todayStart.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")
		out.Days = append(out.Days, FeedingDayCount{Date: key, Count: counts[key]})
	}
	return out, nil
}
