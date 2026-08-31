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

// FeedingDayStats 近 N 个上海「已闭合」日历日每日计数（不含今日）。
// 顺序：从昨天往过去，index0=昨天，供 cash 连续有效日从昨天起扫描。
type FeedingDayStats struct {
	DeviceNo string            `json:"deviceNo"`
	Days     []FeedingDayCount `json:"days"`
}

// shanghaiDayExpr 将 unix start_time 转为 Asia/Shanghai 日历日（yyyy-MM-dd）。
//
// 不用 CONVERT_TZ：未加载 mysql 时区表时 CONVERT_TZ 返回 NULL，GROUP BY 对不上 Go 预填日期，count 全为 0。
// 口径：epoch 秒 + 28800 后按 86400 整除得到东八自然日序号，再加到 1970-01-01（上海无夏令时，偏移固定）。
const shanghaiDayExpr = "DATE_ADD('1970-01-01', INTERVAL ((`start_time` + 28800) DIV 86400) DAY)"

// GetFeedingDayStats 按 device_no 统计近 days 个上海已闭合日历日每日 history 行数。
//
// 业务：供 cash 喂养资格（UCG / 值得留意）取数；口径为该设备全部 history，按 start_time 落入上海日。
// 窗口：自上海昨日起往前共 days 天，半开区间 [windowStart, todayStart)，今日不计入。
// 实现：DB 按上海日 GROUP BY + COUNT（东八 epoch 算术，不依赖 mysql 时区表），零条日由本函数补 0。
//
// Args: deviceNo 设备号；days 天数（建议等于场景 requiredDays，上限 31）。
// Returns: Days[0]=昨天 … 向过去；Side Effects: 仅读 history 表聚合，不拉全量行。
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
	// 今日 00:00：窗口右开端；昨日 00:00：days[0] 锚点。
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	windowStart := yesterdayStart.AddDate(0, 0, -(days - 1))
	startUnix := windowStart.Unix()
	endUnix := todayStart.Unix() // 不含今日

	type aggRow struct {
		D   string `json:"d"`
		Cnt int    `json:"cnt"`
	}
	var rows []aggRow
	// 按设备+时间窗过滤后按上海日聚合，避免拉回全量 start_time。
	err = g.DB().Model("history").Ctx(ctx).
		Fields(shanghaiDayExpr+" AS d, COUNT(*) AS cnt").
		Where("device_no", deviceNo).
		Where("start_time >= ?", startUnix).
		Where("start_time < ?", endUnix).
		Group(shanghaiDayExpr).
		Scan(&rows)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, days)
	for i := 0; i < days; i++ {
		d := yesterdayStart.AddDate(0, 0, -i)
		counts[d.Format("2006-01-02")] = 0
	}
	for _, r := range rows {
		key := strings.TrimSpace(r.D)
		// MySQL DATE 可能以 time.Time 字符串化；统一取前 10 位 yyyy-MM-dd。
		if len(key) >= 10 {
			key = key[:10]
		}
		if _, ok := counts[key]; ok {
			counts[key] = r.Cnt
		}
	}

	out := &FeedingDayStats{DeviceNo: deviceNo, Days: make([]FeedingDayCount, 0, days)}
	// 昨天起向过去排列，便于资格算法从前向后扫连续已闭合日。
	for i := 0; i < days; i++ {
		d := yesterdayStart.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")
		out.Days = append(out.Days, FeedingDayCount{Date: key, Count: counts[key]})
	}
	return out, nil
}
